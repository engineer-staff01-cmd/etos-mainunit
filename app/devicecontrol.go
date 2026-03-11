package app

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
)

var EnableOnlyTemperatureJudge bool = true
var PrevPowerState [64]bool = [64]bool{false}
var PowerStartFlag [64]bool = [64]bool{false}
var DemandEnableVal int = 0 // デマンド制御条件成立

//デマンド逼迫判断・制御を行うスレッド

type DeviceControlThread struct {
	to   chan ChannelMessage
	from chan ChannelMessage
	out  chan ChannelControl
}

func NewDeviceControl() (d *DeviceControlThread) {
	d = new(DeviceControlThread)
	d.to = make(chan ChannelMessage, DeviceControlChannelBufferSize)   // 制御(main -> control)
	d.from = make(chan ChannelMessage, DeviceControlChannelBufferSize) // 制御(main <- control)
	d.out = make(chan ChannelControl, DeviceControlOutBufferSize)      // 制御(control -> child)
	return
}

func (d *DeviceControlThread) Run(wg *sync.WaitGroup, sendchCloudMsg chan ChannelMessage) {
	Logger.Writef(LOG_LEVEL_DEBUG, "Start DeviceControlThread")
	wg.Add(1) // goroutine起動前にカウントを増やす
	go func() {
		d.ControlRun(sendchCloudMsg)
		Logger.Writef(LOG_LEVEL_DEBUG, "Stop DeviceControlThread")
		wg.Done()
	}()
}

/*
* 並列で常時起動、
* 一定周期でデータを取得
 */

func (d *DeviceControlThread) ControlRun(sendchCloudMsg chan ChannelMessage) {

	//debug start
	for {
		chmsg := <-d.to
		switch chmsg.messageType {
		case End:
			return

		case DataGet:
			var messageType uint32
			DeviceControlFlag = false
			systemError := d.controlDataGet(chmsg, sendchCloudMsg)
			DeviceControlFlag = true
			if systemError {
				messageType = ErrorEnd
				Logger.Writef(LOG_LEVEL_ERR, "==================== Control End(System Error) ====================")

			} else {
				messageType = End
				Logger.Writef(LOG_LEVEL_DEBUG, "==================== Control End ====================")
			}

			// cloudスレに送る
			SendDeviceControlFlag = true
			//			msg := ChannelMessage{Alert, strdevicecontrol, chmsg.time}
			//			SendChannelMessageSafely(sendchCloudMsg, msg, false)
			//			Logger.Writef(LOG_LEVEL_DEBUG, "==================== Send Device Control msg : %+v", msg)

			// childUnitスレ終了
			msg := ChannelMessage{messageType, "", chmsg.time}
			SendChannelMessageSafely(d.from, msg, false)
		}
	}
}

func (d *DeviceControlThread) controlDataGet(chmsg ChannelMessage, sendchCloudMsg chan ChannelMessage) bool {
	overAllCommonTime := chmsg.time.Unix()
	fmt.Printf("==================== Control Start ==================== time=%v\n", chmsg.time)

	/*** 設定及び命令の取得 ***/
	/* 機器情報 */
	var deviceInformation []DeviceInformation
	database.SelectAll(&deviceInformation)

	// 出力制御接点情報
	var outputContact []OutputContact
	database.SelectAll(&outputContact)

	// 制御情報
	var controlConditionses []ControlConditions
	database.SelectAll(&controlConditionses)

	// クラウド通信 0:通信可能 1:通信不可
	var cloudComState CloudCommonState
	result := database.GormDB.First(&cloudComState)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "record not found") {
			cloudComState.CommunicationError = 1
		} else {
			Logger.Writef(LOG_LEVEL_ERR, "CloudCommonState Cloud Read DB :%s", result.Error.Error())
			return true
		}
	}

	// エネルギーセンサ情報
	var energySensores []EnergySensor
	database.SelectAll(&energySensores)

	// 環境センサ情報
	var environmentSensores []EnvironmentSensor
	database.SelectAll(&environmentSensores)

	// デマンド制御命令
	var demandControl DemandControl
	//database.SelectOne(&demandControl)
	database.GormDB.Where("id = ?", "DEMAND_CONTROL").First(&demandControl)
	// Logger.Writef(LOG_LEVEL_DEBUG, "demandControl DemandControl DB:%+v", demandControl)

	// 遠隔制御命令
	var remoteControles []RemoteControl
	database.SelectAll(&remoteControles)

	// 遠隔デフロスト制御命令
	var remoteDefrostCommands []RemoteDefrostCommand
	database.SelectAll(&remoteDefrostCommands)

	/*** ステータス取得 ***/

	// 制御状態
	var deviceStatuses []DeviceStatus
	result = database.GormDB.Find(&deviceStatuses)
	if result.Error != nil {
		Logger.Writef(LOG_LEVEL_ERR, "deviceStatuses Cloud Read DB :%s", result.Error.Error())
		return true
	}

	// 機器停止判定時
	var controlStatuses []ControlStatus
	result = database.GormDB.Find(&controlStatuses)
	if result.Error != nil {
		Logger.Writef(LOG_LEVEL_ERR, "controlStatuses Cloud Read DB :%s", result.Error.Error())
		return true
	}

	// 現在値：センサタイプ/温度/湿度/不快指数
	var values []PresentValue
	result = database.GormDB.Find(&values)
	if result.Error != nil {
		Logger.Writef(LOG_LEVEL_ERR, "PresentValue Cloud Read DB :%s", result.Error.Error())
		return true
	}

	// デバイス環境情報を一括取得（N+1クエリ問題の解決）
	var allDeviceEnvironmentList []DeviceEnvironmentInformation
	if err := database.GormDB.Find(&allDeviceEnvironmentList).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		Logger.Writef(LOG_LEVEL_ERR, "DeviceEnvironmentInformation Read DB:%s", err.Error())
		return true
	}
	// デバイスIDをキーとしたマップを作成して高速検索を可能にする
	// 複数のレコードが存在する可能性があるため、最初の1件のみを使用
	deviceEnvironmentMap := make(map[string]DeviceEnvironmentInformation)
	for _, de := range allDeviceEnvironmentList {
		if _, exists := deviceEnvironmentMap[de.DeviceID]; !exists {
			deviceEnvironmentMap[de.DeviceID] = de
		}
	}

	// 遠隔制御から自動制御に切り替わったことを示すフラグ
	shouldNotifyRemoteControlWasReleased := false
	// 機器別制御
	// for _, deviceInfo := range deviceInformation {
	for index, deviceInfo := range deviceInformation {
		var previousControlStatus ControlStatus
		// 内部判定保持用変数
		// エネルギーセンサ情報
		energySensor, _ := SearchenergySensor(deviceInfo, energySensores)

		// 環境センサ情報
		// 環境センサーのidを取得（マップから取得）
		var environmentSensor EnvironmentSensor
		deviceEnvironmentList, exists := deviceEnvironmentMap[deviceInfo.ID]
		var err error
		if !exists {
			err = gorm.ErrRecordNotFound
		}
		if err != nil {
			// Logger.Writef(LOG_LEVEL_DEBUG, "deviceEnvironmentList device ID Read DB:%s, device id : %s", err.Error(), deviceInfo.ID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				count := len(environmentSensores)
				if count == 0 {
					// スライスが空の場合の処理
					Logger.Writef(LOG_LEVEL_DEBUG, "Environment Sensor List is empty")
					// continue
				} else {
					// Logger.Writef(LOG_LEVEL_DEBUG, "deviceEnvironmentList empty. dummy device added.")
					var t NullFloat64
					t.Float64 = 0.0
					t.Valid = true
					environmentSensor = EnvironmentSensor{
						ID:              "01234567-890a-bcde-f012-345678abcdef", // 環境センサー識別子
						Name:            "dummy_temp",                           // 環境センサー名
						ChildDeviceID:   environmentSensores[0].ChildDeviceID,   // 接続子機ID
						DeviceName:      "",                                     // デバイス名
						Category:        strTemperature,                         // 機器カテゴリ
						CorrectionValue: t,                                      // 補正量[℃]
						CorrectionRatio: t,                                      // 補正倍率
						ControlID:       0,                                      // 接続先ID
						ControlChannel:  0,                                      // 接続先CH
						Enable:          1,                                      // 有効
					}
				}
			} else {
				Logger.Writef(LOG_LEVEL_ERR, "deviceEnvironmentList Database error: %s, device id : %s", err.Error(), deviceInfo.ID)
				continue
			}
		} else {
			environmentSensor, _ = SearchEnvironmentSensor(deviceEnvironmentList.SensorID, environmentSensores)
			//Logger.Writef(LOG_LEVEL_DEBUG, "環境センサー情報:%+v", environmentSensor)
		}

		// 制御情報
		controlConditions, _ := SearchControlConditions(deviceInfo, controlConditionses)
		// 遠隔制御
		remoteControl, _ := SearchRemoteControl(deviceInfo, remoteControles)
		// 遠隔デフロスト制御
		remoteDefrostCommand, _ := SearchRemoteDefrostCommand(deviceInfo, remoteDefrostCommands)
		// デフロスト接点
		defrostContact, _ := SearchDefrostContact(deviceInfo, outputContact)
		// 制御状態
		deviceStatus := SearchDeviceStatus(deviceInfo, deviceStatuses)
		// 機器停止判定時
		controlStatus := SearchControlStatus(deviceInfo, controlStatuses)
		// 現在値
		value, _ := SearchPresentValue(deviceInfo, values)

		// 電力(W)
		deviceStatus.ElectricPower = value.CurrentPower

		// 電力量(Wh)
		deviceStatus.ElectricEnergy = value.CurrentPowerIntegration

		// 定格消費電力
		RatedPowerConsumption := SearchRatedPowerConsumption(deviceInfo, energySensor)

		// 稼働率
		deviceStatus.OccupancyRate = CalcOccupancyRate(value.CurrentPower, RatedPowerConsumption)

		// 稼働状況
		if strings.Contains(deviceInfo.Name, "太陽光") {
			Logger.Writef(LOG_LEVEL_DEBUG, "太陽光発電: %f", value.CurrentPower)
			deviceStatus.Operating = value.CurrentPower > 0 // 太陽光発電
		} else {
			deviceStatus.Operating = value.CurrentPower > deviceInfo.StopElectricPower
		}

		// 前回制御ステータス退避
		previousControlStatus = controlStatus

		if strings.Contains(deviceInfo.Name, "太陽光") { // 太陽光の場合
			controlStatus.Control = 0
			controlStatus.Status = StrControlRelease
		} else { // 太陽光以外の制御
			var mode string
			if environmentSensor.Category == strTemperature {
				mode = "temp"
				EnableOnlyTemperatureJudge = true
			} else if environmentSensor.Category == strTemperatureAndHumidity {
				if strings.Contains(environmentSensor.DeviceName, "humi") {
					mode = "humi"
					EnableOnlyTemperatureJudge = false
				} else if strings.Contains(environmentSensor.DeviceName, "temp") {
					mode = "temp"
					EnableOnlyTemperatureJudge = true
				}
			}

			if environmentSensor.DeviceName == "" {
				environmentSensor.DeviceName = "abcdefgh"
			}
			// Logger.Writef(LOG_LEVEL_DEBUG, "environmentSensor.DeviceName : %s, mode : %s, EnableOnlyTemperatureJudge : %t", environmentSensor.DeviceName, mode, EnableOnlyTemperatureJudge)

			// 制御ステータス判定
			//Logger.Writef(LOG_LEVEL_DEBUG, "Demand Control Condition %s : controlConditions.DemandEnable : %d, DemandEnableVal : %d", deviceInfo.Name, controlConditions.DemandEnable, DemandEnableVal)
			if remoteControl != nil {
				/* 遠隔制御 */
				if remoteControl.ControlCommand == 0 {
					controlStatus.Control = 0
				} else if remoteControl.ControlCommand == 1 {
					controlStatus.Control = 1
				}
				controlStatus.Status = StrControlRemote
			} else if remoteDefrostCommand.StopCommand == 1 && defrostContact.Enable == 1 {
				/* 遠隔デフロスト制御 */
				controlStatus.Control = 0
				controlStatus.Status = StrControlRemoteDefrost
			} else if isDefrostControl(deviceInfo) {
				/* デフロスト入力あり（制御中判定？） */
				controlStatus.Control = 0
				controlStatus.Status = StrControlDefrost
			} else if controlConditions.ControlEnable == 0 && controlConditions.DemandEnable == 0 {
				controlStatus.Control = 0
				controlStatus.Status = StrControlRelease
			} else if controlConditions.DemandEnable == 1 && DemandEnableVal == 1 { // クラウドからのデマンド命令を無視
				//} else if cloudComState.CommunicationError == 0 && demandControl.DemandControl == 1 && controlConditions.DemandEnable == 1 {
				/* デマンド制御 */
				if previousControlStatus.Status != StrControlDemand {
					// 前回がデマンド制御以外の場合
					controlStatus.Control = 0
				}
				controlStatus.Status = StrControlDemand
				demandControlJudge(deviceInfo, controlConditions, value, &controlStatus, environmentSensor, mode, deviceStatus)
				Logger.Writef(LOG_LEVEL_DEBUG, "Demand Control Judge : Status : %s, Control : %d", controlStatus.Status, controlStatus.Control)
				//		} else if controlConditions.ControlEnable == 0 && controlConditions.DemandEnable == 0 {
				//			controlStatus.Control = 0
				//			controlStatus.Status = StrControlRelease
			} else {
				/* 自動制御 */
				if previousControlStatus.Status != StrControlAuto {
					controlStatus.Control = 0
				}
				controlStatus.Status = StrControlAuto
				autoControlJudge(deviceInfo, controlConditions, value, &controlStatus, overAllCommonTime, environmentSensor, mode, deviceStatus, index)
				Logger.Writef(LOG_LEVEL_DEBUG, "Auto Control Judge : %s, Control : %d", controlStatus.Status, controlStatus.Control)
			}
		} // 太陽光以外の制御ここまで

		// 前回制御ステータスと比較して変化した場合
		if previousControlStatus != controlStatus {
			// ステータス変化可能判定
			var isPossibleChangeStatus bool
			var nowControlStatus string = controlStatus.Status
			switch controlStatus.Status {
			case StrControlRemote:
				// 遠隔制御
				isPossibleChangeStatus = true
			case StrControlRemoteDefrost:
				// 遠隔デフロスト制御
				isPossibleChangeStatus = true
			default:
				// 稼働必須時間経過判定
				if isPastRequiredControllingTime(deviceInfo, previousControlStatus) {
					isPossibleChangeStatus = true
				}
			}

			// 変化可能
			if isPossibleChangeStatus {
				if controlStatus.Control == 0 {
					// 制御終了
					if previousControlStatus.Control == 1 {
						// 制御中 -> 制御終了の１回のみ実行
						controlStatus.ControlEndTime = overAllCommonTime // 制御終了時間
					}
				} else {
					// 制御開始
					controlStatus.ControlStartTime = overAllCommonTime // 制御開始時間
				}
				// 内部保持の機器別ステータスの保存
				setControlStatusToDeviceStatus(controlStatus, &deviceStatus, overAllCommonTime)
				// 制御ステータス更新
				result = database.GormDB.Where("id = ?", controlStatus.ID).Save(&controlStatus)
				if result.Error != nil {
					Logger.Writef(LOG_LEVEL_ERR, "controlStatuses Save DB :%s", result.Error.Error())
				} else {
					Logger.Writef(LOG_LEVEL_INFO, "controlStatuses Save DB:%+v count:%d", controlStatus, result.RowsAffected)
				}
			} else {
				setControlStatusToDeviceStatus(previousControlStatus, &deviceStatus, overAllCommonTime)
				deviceStatus.Status = nowControlStatus
				// Logger.Writef(LOG_LEVEL_DEBUG, "Previous DeviceStatus Set. controlStatus : %+v", previousControlStatus)
			}
		} else {
			setControlStatusToDeviceStatus(previousControlStatus, &deviceStatus, overAllCommonTime)
			if PowerStartFlag[index] {
				// 制御ステータス更新
				result = database.GormDB.Where("id = ?", controlStatus.ID).Save(&controlStatus)
				if result.Error != nil {
					Logger.Writef(LOG_LEVEL_ERR, "Previous controlStatuses Save DB :%s", result.Error.Error())
				} else {
					Logger.Writef(LOG_LEVEL_INFO, "Previous controlStatuses Save DB:%+v count:%d", controlStatus, result.RowsAffected)
				}
			}
		}
		// デバイスステータス更新
		//		database.Delete(&DeviceStatus{}, "device_id = ?", deviceStatus.DeviceID)
		if err := database.Save(&deviceStatus); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save device status: %s", err.Error())
		}
		//Logger.Writef(LOG_LEVEL_DEBUG, "DeviceStatus Update Success:%+v", &deviceStatus)

		// チャイルドユニットにチャンネルで値渡し
		for {
			if IoUnitFlag == false {
				msg := ChannelControl{deviceStatus.Status, deviceInfo.ID, deviceStatus.Control, overAllCommonTime}
				SendChannelControlSafely(d.out, msg, false)
				break
			}
		}

		if previousControlStatus.Status == StrControlRemote && controlStatus.Status != StrControlRemote {
			shouldNotifyRemoteControlWasReleased = true
			if err := database.Save(&RemoteControlReleaseStatus{controlStatus.ID}); err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "Failed to save remote control release status: %s", err.Error())
			}
		}
	}

	// 緊急制御命令を受信、処理後は完了通知をクラウドに送信
	if shouldNotifyRemoteControlWasReleased {
		msg := ChannelMessage{Notify, strManualControl, chmsg.time}
		SendChannelMessageSafely(sendchCloudMsg, msg, false)
	}

	return false
}

func setControlStatusToDeviceStatus(controlStatus ControlStatus, deviceStatus *DeviceStatus, statusTime int64) {
	if (controlStatus.Status == StrControlDemand || controlStatus.Status == StrControlAuto) && controlStatus.Control == 0 {
		deviceStatus.Status = StrControlRelease
	} else {
		deviceStatus.Status = controlStatus.Status
	}
	deviceStatus.Status = controlStatus.Status
	deviceStatus.Control = controlStatus.Control
	deviceStatus.Time = statusTime

}

func isDefrostControl(devInfo DeviceInformation) (ret bool) {
	overAllCommonTime := time.Now().Unix()
	// InputContact read DB
	var inputContacts []InputContact
	database.SelectAll(&inputContacts)
	for _, inputContact := range inputContacts {
		if !IsEnableChildUnit(inputContact.ChildDeviceID) {
			continue
		}
		if inputContact.ID == devInfo.DefrostInputID {
			inputIo := NewIoUnit(inputContact.DeviceName, byte(inputContact.ControlID))
			if inputContact.Enable == 1 {
				value, err := inputIo.GetInputValue(uint16(inputContact.ControlChannel))
				if err == nil {
					if value == 1 {
						ret = true
						break
					}
				} else {
					Logger.Writef(LOG_LEVEL_ERR, "ControlChannel ID:%d CH:%d :%s", inputContact.ControlID, inputContact.ControlChannel, err.Error())
					// 通信異常判定
					if strings.Contains(err.Error(), "timeout") {

						SaveDBChildDeviceCommunicationAlert(inputContact.ChildDeviceID, strInputContact, COM_NG)
						UpdateCommunicationMailAlertStatus(inputContact.ID, strInputContact, overAllCommonTime, COM_NG)
						SaveDBCommunicationAlert(inputContact.ID, inputContact.Name, time.Now().Unix(), strInputContact)

						Logger.Writef(LOG_LEVEL_ERR, "Alert ControlChannel ID:%d CH:%d :%s", inputContact.ControlID, inputContact.ControlChannel, err.Error())
					} else {
						SaveDBChildDeviceCommunicationAlert(inputContact.ChildDeviceID, strInputContact, COM_OK)
						UpdateCommunicationMailAlertStatus(inputContact.ID, strInputContact, overAllCommonTime, COM_OK)
					}
				}
			}
		}
	}
	return ret
}

/*
デマンド制御 開始判定
引数：機器情報、制御情報、機器状態、現在値
戻り値：
*/
func demandControlJudge(deviceInfo DeviceInformation,
	controlConditions ControlConditions,
	value PresentValue,
	controlStatus *ControlStatus,
	environmentSensor EnvironmentSensor, mode string, deviceStatus DeviceStatus) {

	//if controlConditions.ControlEnable == 0 || controlConditions.DemandEnable == 0 {
	if controlConditions.DemandEnable == 0 {
		controlStatus.Control = 0
		return
	}

	// 稼働していない場合は制御しない
	if !deviceStatus.Operating && controlStatus.Control == 0 {
		controlStatus.Control = 0
		return
	}

	//if environmentSensor.DeviceName == "" {
	//	environmentSensor.DeviceName = "abcdefgh"
	//}

	/* デマンド制御判定 */
	if controlStatus.Control == 0 {
		//データベースのデータ次第
		controlLeftTime := int64(time.Since(time.Unix(controlStatus.ControlEndTime, 0)).Seconds())
		Logger.Writef(LOG_LEVEL_DEBUG, "demandControl start Judge deviceInfo.RequiredControllingTime :%d", deviceInfo.RequiredControllingTime)
		Logger.Writef(LOG_LEVEL_DEBUG, "demandControl start Judge controlLeftTime :%d, ControlEndTime :%d", controlLeftTime, controlStatus.ControlEndTime)
		Logger.Writef(LOG_LEVEL_DEBUG, "demandControl start Judge value.CurrentPower :%f", value.CurrentPower)
		Logger.Writef(LOG_LEVEL_DEBUG, "demandControl start Judge deviceInfo.StopElectricPower :%f", deviceInfo.StopElectricPower)
		if (deviceInfo.RequiredControllingTime < controlLeftTime) && value.CurrentPower > deviceInfo.StopElectricPower {
			controlStatus.Control = demandControlDeviceStopJudge(controlConditions, value, mode, environmentSensor.Category) // 機器停止判定
		}
	} else {
		/* 親機情報 */
		var baseInformation BaseInformation
		result := database.GormDB.First(&baseInformation)
		if result.Error != nil {
			Logger.Writef(LOG_LEVEL_ERR, "baseInformation Cloud Read DB :%s", result.Error.Error())
			controlStatus.Control = 0
			return
		}
		controlLeftTime := int64(time.Since(time.Unix(controlStatus.ControlStartTime, 0)).Seconds())
		Logger.Writef(LOG_LEVEL_DEBUG, "demandControl end Judge deviceInfo.RequiredControllingTime :%d", deviceInfo.RequiredControllingTime)
		Logger.Writef(LOG_LEVEL_DEBUG, "demandControl end Judge controlLeftTime :%d, ControlStartTime :%d", controlLeftTime, controlStatus.ControlStartTime)
		Logger.Writef(LOG_LEVEL_DEBUG, "demandControl end Judge dbaseInformation.CancellationElectricPower :%f", baseInformation.CancellationElectricPower)
		if deviceInfo.RequiredControllingTime < controlLeftTime {
			deviceRelease := demandControlDeviceReleaseJudge(controlConditions, value, mode, environmentSensor.Category) // 機器停止判定
			if deviceRelease == 1 {
				controlStatus.Control = 0
			} else {
				controlStatus.Control = 1
			}
		} else {
			controlStatus.Control = 1
		}
	}
}

/*
func calculateDemandPower(baseInfo BaseInformation) (ret ReturnValueCurrentPower) {

	// database data status get
	var demandTightData DemandTightData
	result := database.GormDB.First(&demandTightData)
	if result.Error != nil {
		Logger.Writef(LOG_LEVEL_ERR, "demandTightData Cloud Read DB :%s", result.Error.Error())
		return
	}

	ret = CalcDemandPower(&demandTightData, baseInfo, time.Now().Unix())

	return
}
*/

/*
	自動制御
	引数：機器情報、制御情報、現在値、機器状態、
	戻り値：
*/

func autoControlJudge(deviceInfo DeviceInformation,
	controlConditions ControlConditions,
	value PresentValue,
	controlStatus *ControlStatus,
	overAllCommonTime int64,
	environmentSensor EnvironmentSensor, mode string, deviceStatus DeviceStatus, index int) {
	var isIntermittentControl bool // 間欠制御フラグ
	var powerStartTime int64
	// var powerStartFlag bool = false

	// 制御対象出ない場合は制御しない
	if controlConditions.ControlEnable != 1 {
		controlStatus.Control = 0
		return
	}

	// 稼働していない場合は制御しない
	if !deviceStatus.Operating && controlStatus.Control == 0 {
		controlStatus.Control = 0
		// Logger.Writef(LOG_LEVEL_DEBUG, "Operation stop, deviceStatus.Operating : %t, controlStatus.Control : %d", deviceStatus.Operating, controlStatus.Control)
		PrevPowerState[index] = true
		PowerStartFlag[index] = false
		return
	} else if PrevPowerState[index] && deviceStatus.Operating && controlStatus.Control == 0 {
		PrevPowerState[index] = false
		PowerStartFlag[index] = true
		powerStartTime = overAllCommonTime // 制御開始時間
		if controlStatus.ControlEndTime < powerStartTime {
			Logger.Writef(LOG_LEVEL_DEBUG, "Operation start, controlStatus.ControlEndTime : %d, powerStartTime : %d", controlStatus.ControlEndTime, powerStartTime)
			controlStatus.ControlEndTime = powerStartTime
		}
	}

	// 温度未設定の場合は間欠制御
	//if environmentSensor.Category == strTemperature || strings.Contains(environmentSensor.DeviceName, "temp") {
	if environmentSensor.Category == strTemperature || mode == "temp" {
		if !controlConditions.ControlStartTemperature.Valid {
			isIntermittentControl = true
		}
		//} else if environmentSensor.Category == strTemperatureAndHumidity && strings.Contains(environmentSensor.DeviceName, "humi") {
	} else if environmentSensor.Category == strTemperatureAndHumidity && mode == "humi" {
		if !controlConditions.ControlStartHumidity.Valid {
			isIntermittentControl = true
		}
	}

	//if environmentSensor.DeviceName == "" {
	//	environmentSensor.DeviceName = "abcdefgh"
	//}

	// 温度が範囲外の場合は間欠制御
	if controlConditions.ControlStartTemperature.Float64 > maxLimitTemperature || controlConditions.ControlStartTemperature.Float64 < minLimitTemperature {
		isIntermittentControl = true
	}
	// Logger.Writef(LOG_LEVEL_DEBUG, "Auto Control Judge Start, DeviceID : %s, isIntermittentControl : %t", deviceInfo.ID, isIntermittentControl)

	if controlStatus.Control == 0 {
		// 制御開始判定
		releaseControlTime := controlConditions.ReleaseControlTime // 制御解除時間
		if PowerStartFlag[index] {
			// 比較時間を稼働必須時間に変更
			releaseControlTime = deviceInfo.RequiredControllingTime
		}
		// Logger.Writef(LOG_LEVEL_DEBUG, "Auto Control Judge Start, release Control Time : %v", releaseControlTime)

		// 必須稼働時間と制御解除時間を比較
		if !isOnlyTemperatureControl(controlConditions) && releaseControlTime < deviceInfo.RequiredControllingTime {
			// 稼働必須時間より制御解除時間が短い場合は稼働必須時間分稼働させる
			// システムアラートを発報
			sysAlert := GenerateSystemAlert(overAllCommonTime, ERCD_CONTROL_CONDITION_ERROR)
			if err := database.Save(&sysAlert); err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "Failed to save system alert: %s", err.Error())
			}

			// 比較時間を稼働必須時間に変更
			releaseControlTime = deviceInfo.RequiredControllingTime
		}

		/* 自動制御 開始判断 */
		//データベースのデータ次第
		controlLeftTime := int64(time.Since(time.Unix(controlStatus.ControlEndTime, 0)).Seconds())
		// Logger.Writef(LOG_LEVEL_DEBUG, "Auto Control Judge Start, controlLeftTime : %d, ControlEndTime : %d", controlLeftTime, controlStatus.ControlEndTime)
		// ※稼働必須時間入っているか確認する

		if isIntermittentControl {
			//間欠制御
			if releaseControlTime < controlLeftTime {
				// 前回制御終了時間から稼働必須時間経過していれば制御開始
				controlStatus.Control = 1
			}
		} else if isOnlyTemperatureControl(controlConditions) ||
			(releaseControlTime < controlLeftTime && value.CurrentPower > deviceInfo.StopElectricPower) {
			// 温度判定が必要な制御
			// 温帯制御 = 制御時間：0 かつ 制御解除時間：0 = (controlConditions.ControlTime == 0 && releaseControlTime == 0)
			// 稼働必須時間経過 = 稼働必須時間 < 経過時間 かつ 稼働中（電力 > 稼働判断電力）
			// 温帯制御 または 制御時間 を超えているとき
			//Logger.Writef(LOG_LEVEL_DEBUG, "Only Temperature Control : %v, PresentValue : %v, EnvironmentSensor : %v", controlConditions, value, environmentSensor)
			controlStatus.Control = autoControlDeviceStopJudge(controlConditions, value, mode, environmentSensor.Category)
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "Auto Control Judge Start, controlStatus.Control : %d", controlStatus.Control)
	} else {
		PowerStartFlag[index] = false

		controlTime := controlConditions.ControlTime // 制御時間
		// Logger.Writef(LOG_LEVEL_DEBUG, "Auto Control Judge Start, controlTime : %d", controlTime)

		// 必須稼働時間と制御解除時間を比較
		if !isOnlyTemperatureControl(controlConditions) && controlTime < deviceInfo.RequiredControllingTime {
			// 稼働必須時間より制御時間が短い場合は稼働必須時間分稼働させる
			// システムアラートを発報
			sysAlert := GenerateSystemAlert(overAllCommonTime, ERCD_CONTROL_CONDITION_ERROR)
			if err := database.Save(&sysAlert); err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "Failed to save system alert: %s", err.Error())
			}

			// 比較時間を稼働必須時間に変更
			controlTime = deviceInfo.RequiredControllingTime
			// Logger.Writef(LOG_LEVEL_DEBUG, "Auto Control Judge Start, RequiredControllingTime : %d", controlTime)
		}

		controlLeftTime := int64(time.Since(time.Unix(controlStatus.ControlStartTime, 0)).Seconds())
		Logger.Writef(LOG_LEVEL_DEBUG, "Auto Control Judge Start, controlLeftTime : %d, controlTime : %d", controlLeftTime, controlTime)

		if isIntermittentControl {
			//間欠制御
			if controlTime < controlLeftTime {
				// 間欠制御の場合は稼働時間を経過したら制御終了
				controlStatus.Control = 0
			} else {
				// 制御継続
				controlStatus.Control = 1
			}
			// Logger.Writef(LOG_LEVEL_DEBUG, "Auto Control Judge Start, controlTime : %d, controlLeftTime : %d, controlStatus.Control : %d", controlTime, controlLeftTime, controlStatus.Control)
			// 温度判定が必要な制御
			// 温帯制御 = 制御時間：0 かつ 制御解除時間：0 = (controlConditions.ControlTime == 0 && releaseControlTime == 0)
			// 稼働必須時間経過 = 稼働必須時間 < 経過時間
			// 温帯制御 または 制御時間経過（温帯制御かつ間欠制御）
		} else if isOnlyTemperatureControl(controlConditions) {
			deviceRelease := autoControlDeviceReleaseJudge(controlConditions, value, mode, environmentSensor.Category)
			if deviceRelease == 1 {
				controlStatus.Control = 0
			} else {
				controlStatus.Control = 1
			}
			Logger.Writef(LOG_LEVEL_DEBUG, "Auto Control Judge Start, deviceRelease : %d, controlStatus.Control : %d", deviceRelease, controlStatus.Control)
		} else {
			deviceRelease := autoControlDeviceReleaseJudge(controlConditions, value, mode, environmentSensor.Category)
			if deviceRelease == 1 || controlTime < controlLeftTime {
				controlStatus.Control = 0
			} else {
				controlStatus.Control = 1
			}
			Logger.Writef(LOG_LEVEL_DEBUG, "Auto Control Judge Start, deviceRelease : %d, controlStatus.Control : %d", deviceRelease, controlStatus.Control)
		}
	}
}

func isOnlyTemperatureControl(controlConditions ControlConditions) bool {
	if controlConditions.ControlTime == 0 && controlConditions.ReleaseControlTime == 0 {
		return true
	} else {
		return false
	}
}

// 稼働停止（制御開始）判定
/*
func controlDeviceStopJudge(t, h, d NullFloat64, value PresentValue) int8 {
	if t.Float64 > maxLimitTemperature || t.Float64 < minLimitTemperature {
		return 1
	}

	// 開始温度比較
	if t.Valid && value.CurrentTemperature > t.Float64 {
		//Logger.Writef(LOG_LEVEL_DEBUG, "機器稼働継続")
		return 0
	}

	// 温湿度有効
	if !EnableOnlyTemperatureJudge {
		// 開始湿度比較
		if h.Valid && value.CurrentHumidity > h.Float64 {
			return 0
		}
		// 開始不快指数比較
		if d.Valid && value.CurrentDiscomfortIndex > d.Float64 {
			return 0
		}
	}

	return 1
}

// 稼働（制御停止）判定
func controlDeviceReleaseJudge(t, h, d NullFloat64, value PresentValue) int8 {
	if t.Float64 > maxLimitTemperature || t.Float64 < minLimitTemperature {
		return 1
	}

	// 開始温度比較
	if t.Valid && value.CurrentTemperature <= t.Float64 {
		return 0
	}

	// 温湿度有効
	if !EnableOnlyTemperatureJudge {
		// 開始湿度比較
		if h.Valid && value.CurrentHumidity <= h.Float64 {
			return 0
		}
		// 開始不快指数比較
		if d.Valid && value.CurrentDiscomfortIndex <= d.Float64 {
			return 0
		}
	}

	return 1
}
*/

// 稼働停止（制御開始）判定
func temphumicontrolDeviceStopJudge(t, h, d NullFloat64, value PresentValue, demand bool, mode string, category string) int8 {
	var ret int8

	//if category == strTemperature {
	//	Logger.Writef(LOG_LEVEL_DEBUG, "tempcontrolDeviceStopJudge:%s", category)
	//	ret = tempcontrolDeviceStopJudge(t, h, d, value, demand, mode)
	//} else if category == strTemperatureAndHumidity {
	Logger.Writef(LOG_LEVEL_DEBUG, "humicontrolDeviceStopJudge:%s", category)
	ret = humicontrolDeviceStopJudge(t, h, d, value, demand, mode)
	//}

	return ret
}

/*
func tempcontrolDeviceStopJudge(t, h, d NullFloat64, value PresentValue, demand bool, mode string) int8 {
	var tempANDCondition TempANDCondition
	err := database.GormDB.Where("id = ?", value.ID).First(&tempANDCondition).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_DEBUG, "not AND Condition tempANDConditions DB:%s", err.Error())
		if t.Float64 > maxLimitTemperature || t.Float64 < minLimitTemperature {
			return 1
		}

		if EnableOnlyTemperatureJudge || mode == "temp" {
			// 開始温度比較
			if t.Valid && value.CurrentTemperature > t.Float64 {
				//Logger.Writef(LOG_LEVEL_DEBUG, "機器稼働継続")
				Logger.Writef(LOG_LEVEL_DEBUG, "value.CurrentTemperature(%f) > t.Float64(%f) : return 0", value.CurrentTemperature, t.Float64)
				return 0
			}
		} else if !EnableOnlyTemperatureJudge && mode == "humi" { // 温湿度有効
			// 開始湿度比較
			if h.Valid && value.CurrentHumidity > h.Float64 {
				Logger.Writef(LOG_LEVEL_DEBUG, "value.CurrentHumidity(%f) > h.Float64(%f) : return 0", value.CurrentHumidity, h.Float64)
				return 0
			}
			// 開始不快指数比較
			if d.Valid && value.CurrentDiscomfortIndex > d.Float64 {
				return 0
			}
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "tempcontrolDeviceStopJudge : return 1")
		return 1
		//} else {
		//	Logger.Writef(LOG_LEVEL_DEBUG, "AND condition DB select : %v", tempANDCondition)
	}

	// AND条件判定
	Logger.Writef(LOG_LEVEL_DEBUG, "AND Condition Judge Start(temp)")
	var tempVal = [6]float64{tempANDCondition.Temp1_Val, tempANDCondition.Temp2_Val, tempANDCondition.Temp3_Val, tempANDCondition.Temp4_Val, tempANDCondition.Temp5_Val, tempANDCondition.Temp6_Val}
	var tempCnt = [6]string{tempANDCondition.Temp1_Cnt, tempANDCondition.Temp2_Cnt, tempANDCondition.Temp3_Cnt, tempANDCondition.Temp4_Cnt, tempANDCondition.Temp5_Cnt, tempANDCondition.Temp6_Cnt}
	var ret = [6]int8{1, 1, 1, 1, 1, 1}
	var num int = 0
	for num < tempANDCondition.Number {
		var controlData ControlConditions
		err := database.GormDB.Where("target_devices_id = ?", tempCnt[num]).Find(&controlData).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "controlData Read DB:%s, Device ID:%s", err.Error(), tempCnt[num])
		} else {
			if demand {
				t = controlData.DemandStartTemperature
				h = controlData.DemandStartHumidity
				d = controlData.DemandStartDiscomfortIndex
			} else {
				t = controlData.ControlStartTemperature
				h = controlData.ControlStartHumidity
				d = controlData.ControlStartDiscomfortIndex
			}
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "AND ControlStartTemperature : %v, Temperature data : %v, Demand : %v", t, tempVal[num], demand)
		if t.Float64 > maxLimitTemperature || t.Float64 < minLimitTemperature {
			ret[num] = 1
		} else if t.Valid && (tempVal[num] > t.Float64) { // 開始温度比較
			//Logger.Writef(LOG_LEVEL_DEBUG, "機器稼働継続")
			ret[num] = 0
		} else if !EnableOnlyTemperatureJudge { // 温湿度有効
			if h.Valid && value.CurrentHumidity > h.Float64 { // 開始湿度比較
				ret[num] = 0
			} else if d.Valid && value.CurrentDiscomfortIndex > d.Float64 { // 開始不快指数比較
				ret[num] = 0
			}
		}
		num++
	}

	var ret_and int8 = ret[0] & ret[1] & ret[2] & ret[3] & ret[4] & ret[5]
	Logger.Writef(LOG_LEVEL_DEBUG, "AND condition ret data : %v, AND data : %d", ret, ret_and)

	return ret_and
}
*/

func humicontrolDeviceStopJudge(t, h, d NullFloat64, value PresentValue, demand bool, mode string) int8 {
	var humiANDCondition HumiANDCondition
	err := database.GormDB.Where("id = ?", value.ID).First(&humiANDCondition).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "not AND Condition humiANDConditions DB:%s", err.Error())
		if t.Float64 > maxLimitTemperature || t.Float64 < minLimitTemperature {
			return 1
		}

		if EnableOnlyTemperatureJudge || mode == "temp" {
			// 開始温度比較
			if t.Valid && value.CurrentTemperature > t.Float64 {
				//Logger.Writef(LOG_LEVEL_DEBUG, "機器稼働継続")
				Logger.Writef(LOG_LEVEL_DEBUG, "tempcontrolDeviceStopJudge : ID(%s), value.CurrentTemperature(%f) > t.Float64(%f) : return 0", value.ID, value.CurrentTemperature, t.Float64)
				return 0
			}
			Logger.Writef(LOG_LEVEL_DEBUG, "tempcontrolDeviceStopJudge : ID(%s), t.Valid(%t), value.CurrentTemperature(%f), t.Float64(%f),  return 1", value.ID, t.Valid, value.CurrentTemperature, t.Float64)
		} else if !EnableOnlyTemperatureJudge && mode == "humi" { // 温湿度有効
			// 開始湿度比較
			if h.Valid && value.CurrentHumidity > h.Float64 {
				Logger.Writef(LOG_LEVEL_DEBUG, "humicontrolDeviceStopJudge : ID(%s), value.CurrentHumidity(%f) > h.Float64(%f) : return 0", value.ID, value.CurrentHumidity, h.Float64)
				return 0
			}
			Logger.Writef(LOG_LEVEL_DEBUG, "humicontrolDeviceStopJudge : ID(%s), h.Valid(%t), value.CurrentHumidity(%f), h.Float64(%f),  return 1", value.ID, h.Valid, value.CurrentHumidity, h.Float64)
			// 開始不快指数比較
			if d.Valid && value.CurrentDiscomfortIndex > d.Float64 {
				return 0
			}
		}
		// Logger.Writef(LOG_LEVEL_DEBUG, "tempcontrolDeviceReleaseJudge : EnableOnlyTemperatureJudge(%t), mode(%s)", EnableOnlyTemperatureJudge, mode)
		return 1
		//} else {
		//	Logger.Writef(LOG_LEVEL_DEBUG, "AND condition DB select : %v", tempANDCondition)
	}

	// AND条件判定
	// Logger.Writef(LOG_LEVEL_DEBUG, "AND Condition Judge Start(humi)")
	var humiVal = [8]float64{humiANDCondition.Humi1_Val, humiANDCondition.Humi2_Val, humiANDCondition.Humi3_Val, humiANDCondition.Humi4_Val, humiANDCondition.Humi5_Val, humiANDCondition.Humi6_Val, humiANDCondition.Humi7_Val, humiANDCondition.Humi8_Val}
	var humiCnt = [8]string{humiANDCondition.Humi1_Cnt, humiANDCondition.Humi2_Cnt, humiANDCondition.Humi3_Cnt, humiANDCondition.Humi4_Cnt, humiANDCondition.Humi5_Cnt, humiANDCondition.Humi6_Cnt, humiANDCondition.Humi7_Cnt, humiANDCondition.Humi8_Cnt}
	var humiMode = [8]string{humiANDCondition.DeviceName1, humiANDCondition.DeviceName2, humiANDCondition.DeviceName3, humiANDCondition.DeviceName4, humiANDCondition.DeviceName5, humiANDCondition.DeviceName6, humiANDCondition.DeviceName7, humiANDCondition.DeviceName8}
	
	// 必要なcontrolDataを一括取得（N+1クエリ問題の解決）
	controlDataMap := make(map[string]ControlConditions)
	if humiANDCondition.Number > 0 {
		// 重複を除去したIDリストを作成
		deviceIDs := make([]string, 0, humiANDCondition.Number)
		seen := make(map[string]bool)
		for num := 0; num < humiANDCondition.Number; num++ {
			if humiCnt[num] != "" && !seen[humiCnt[num]] {
				deviceIDs = append(deviceIDs, humiCnt[num])
				seen[humiCnt[num]] = true
			}
		}
		// 一括取得
		if len(deviceIDs) > 0 {
			var controlDataList []ControlConditions
			if err := database.GormDB.Where("target_devices_id IN (?)", deviceIDs).Find(&controlDataList).Error; err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "controlData batch Read DB:%s", err.Error())
			} else {
				for _, cd := range controlDataList {
					controlDataMap[cd.TargetDevicesID] = cd
				}
			}
		}
	}
	
	var ret = [8]int8{1, 1, 1, 1, 1, 1, 1, 1}
	var num int = 0
	for num < humiANDCondition.Number {
		controlData, exists := controlDataMap[humiCnt[num]]
		if !exists {
			Logger.Writef(LOG_LEVEL_ERR, "controlData not found, Device ID:%s", humiCnt[num])
			num++
			continue
		}
		
		if demand {
			t = controlData.DemandStartTemperature
			h = controlData.DemandStartHumidity
			d = controlData.DemandStartDiscomfortIndex
		} else {
			t = controlData.ControlStartTemperature
			h = controlData.ControlStartHumidity
			d = controlData.ControlStartDiscomfortIndex
		}

		if t.Float64 > maxLimitTemperature || t.Float64 < minLimitTemperature {
			ret[num] = 1
		} else if humiMode[num][4:8] == "temp" {
			Logger.Writef(LOG_LEVEL_DEBUG, "AND ControlStartTemperature : %v, Temperature data : %v, Demand : %v", t, humiVal[num], demand)
			if t.Valid && (humiVal[num] > t.Float64) { // 開始温度比較
				//Logger.Writef(LOG_LEVEL_DEBUG, "機器稼働継続")
				ret[num] = 0
			}
		} else if humiMode[num][4:8] == "humi" { // 湿度
			Logger.Writef(LOG_LEVEL_DEBUG, "AND ControlStartHumidity : %v, Humidity data : %v, Demand : %v", h, humiVal[num], demand)
			//if h.Valid && value.CurrentHumidity > h.Float64 { // 開始湿度比較
			if h.Valid && humiVal[num] > h.Float64 { // 開始湿度比較
				ret[num] = 0
			} //else if d.Valid && value.CurrentDiscomfortIndex > d.Float64 { // 開始不快指数比較
			//ret[num] = 0
			//}
		}
		num++
	}

	var ret_and int8 = ret[0] & ret[1] & ret[2] & ret[3] & ret[4] & ret[5] & ret[6] & ret[7]
	// Logger.Writef(LOG_LEVEL_DEBUG, "AND condition ret data : %v, AND data : %d", ret, ret_and)

	return ret_and
}

// 稼働（制御停止）判定
func temphumicontrolDeviceReleaseJudge(t, h, d NullFloat64, value PresentValue, demand bool, mode string, category string) int8 {
	var ret int8

	//if category == strTemperature {
	//	ret = tempcontrolDeviceReleaseJudge(t, h, d, value, demand, mode)
	//} else if category == strTemperatureAndHumidity {
	ret = humicontrolDeviceReleaseJudge(t, h, d, value, demand, mode)
	//}

	return ret
}

/*
func tempcontrolDeviceReleaseJudge(t, h, d NullFloat64, value PresentValue, demand bool, mode string) int8 {
	var tempANDCondition TempANDCondition
	err := database.GormDB.Where("id = ?", value.ID).First(&tempANDCondition).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_DEBUG, "not AND Condition tempANDConditions DB:%s", err.Error())
		if t.Float64 > maxLimitTemperature || t.Float64 < minLimitTemperature {
			return 1
		}

		if EnableOnlyTemperatureJudge || mode == "temp" {
			// 開始温度比較
			if t.Valid && value.CurrentTemperature <= t.Float64 {
				Logger.Writef(LOG_LEVEL_DEBUG, "value.CurrentTemperature(%f) <= t.Float64(%f) : return 0", value.CurrentTemperature, t.Float64)
				return 0
			}
		} else if !EnableOnlyTemperatureJudge && mode == "humi" { // 温湿度有効
			// 開始湿度比較
			if h.Valid && value.CurrentHumidity <= h.Float64 {
				Logger.Writef(LOG_LEVEL_DEBUG, "value.CurrentHumidity(%f) <= h.Float64(%f) : return 0", value.CurrentHumidity, h.Float64)
				return 0
			}
			// 開始不快指数比較
			if d.Valid && value.CurrentDiscomfortIndex <= d.Float64 {
				return 0
			}
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "tempcontrolDeviceReleaseJudge : return 1")
		return 1
		//} else {
		//	Logger.Writef(LOG_LEVEL_DEBUG, "AND condition DB select : %v", tempANDCondition)
	}

	// AND条件判定
	Logger.Writef(LOG_LEVEL_DEBUG, "AND Condition Judge Start(temp)")
	var tempVal = [6]float64{tempANDCondition.Temp1_Val, tempANDCondition.Temp2_Val, tempANDCondition.Temp3_Val, tempANDCondition.Temp4_Val, tempANDCondition.Temp5_Val, tempANDCondition.Temp6_Val}
	var tempCnt = [6]string{tempANDCondition.Temp1_Cnt, tempANDCondition.Temp2_Cnt, tempANDCondition.Temp3_Cnt, tempANDCondition.Temp4_Cnt, tempANDCondition.Temp5_Cnt, tempANDCondition.Temp6_Cnt}
	var ret = [6]int8{1, 1, 1, 1, 1, 1}
	var num int = 0
	for num < tempANDCondition.Number {
		var controlData ControlConditions
		err := database.GormDB.Where("target_devices_id = ?", tempCnt[num]).Find(&controlData).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "controlData Read DB:%s, Device ID:%s", err.Error(), tempCnt[num])
		} else {
			if demand {
				t = controlData.DemandStartTemperature
				h = controlData.DemandStartHumidity
				d = controlData.DemandStartDiscomfortIndex
			} else {
				t = controlData.ControlStartTemperature
				h = controlData.ControlStartHumidity
				d = controlData.ControlStartDiscomfortIndex
			}
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "AND ControlStartTemperature : %v, Temperature data : %v, Demand : %v", t, tempVal[num], demand)
		if t.Float64 > maxLimitTemperature || t.Float64 < minLimitTemperature {
			ret[num] = 1
		} else if t.Valid && (tempVal[num] <= t.Float64) { // 開始温度比較
			ret[num] = 0
		} else if !EnableOnlyTemperatureJudge { // 温湿度有効
			if h.Valid && value.CurrentHumidity <= h.Float64 { // 開始湿度比較
				ret[num] = 0
			} else if d.Valid && value.CurrentDiscomfortIndex <= d.Float64 { // 開始不快指数比較
				ret[num] = 0
			}
		}
		num++
	}

	var ret_and int8 = ret[0] & ret[1] & ret[2] & ret[3] & ret[4] & ret[5]
	Logger.Writef(LOG_LEVEL_DEBUG, "AND condition ret data : %v, AND data : %d", ret, ret_and)

	return ret_and
}
*/

func humicontrolDeviceReleaseJudge(t, h, d NullFloat64, value PresentValue, demand bool, mode string) int8 {
	var humiANDCondition HumiANDCondition
	err := database.GormDB.Where("id = ?", value.ID).First(&humiANDCondition).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "not AND Condition humiANDConditions DB:%s", err.Error())
		if t.Float64 > maxLimitTemperature || t.Float64 < minLimitTemperature {
			return 1
		}

		if EnableOnlyTemperatureJudge || mode == "temp" {
			// 開始温度比較
			if t.Valid && value.CurrentTemperature <= t.Float64 {
				Logger.Writef(LOG_LEVEL_DEBUG, "tempcontrolDeviceReleaseJudge : ID(%s), value.CurrentTemperature(%f) <= t.Float64(%f) : return 0", value.ID, value.CurrentTemperature, t.Float64)
				return 0
			}
			Logger.Writef(LOG_LEVEL_DEBUG, "tempcontrolDeviceReleaseJudge : ID(%s), t.Valid(%t), value.CurrentTemperature(%f), t.Float64(%f),  return 1", value.ID, t.Valid, value.CurrentTemperature, t.Float64)
		} else if !EnableOnlyTemperatureJudge && mode == "humi" { // 温湿度有効
			// 開始湿度比較
			if h.Valid && value.CurrentHumidity <= h.Float64 {
				Logger.Writef(LOG_LEVEL_DEBUG, "humicontrolDeviceReleaseJudge : ID(%s), value.CurrentHumidity(%f) <= h.Float64(%f) : return 0", value.ID, value.CurrentHumidity, h.Float64)
				return 0
			}
			Logger.Writef(LOG_LEVEL_DEBUG, "humicontrolDeviceReleaseJudge : ID(%s), h.Valid(%t), value.CurrentHumidity(%f), h.Float64(%f),  return 1", value.ID, h.Valid, value.CurrentHumidity, h.Float64)
			// 開始不快指数比較
			if d.Valid && value.CurrentDiscomfortIndex <= d.Float64 {
				return 0
			}
		}
		// Logger.Writef(LOG_LEVEL_DEBUG, "tempcontrolDeviceReleaseJudge : EnableOnlyTemperatureJudge(%t), mode(%s)", EnableOnlyTemperatureJudge, mode)
		return 1
		//} else {
		//	Logger.Writef(LOG_LEVEL_DEBUG, "AND condition DB select : %v", tempANDCondition)
	}

	// AND条件判定
	// Logger.Writef(LOG_LEVEL_DEBUG, "AND Condition Judge Start(humi)")
	var humiVal = [8]float64{humiANDCondition.Humi1_Val, humiANDCondition.Humi2_Val, humiANDCondition.Humi3_Val, humiANDCondition.Humi4_Val, humiANDCondition.Humi5_Val, humiANDCondition.Humi6_Val, humiANDCondition.Humi7_Val, humiANDCondition.Humi8_Val}
	var humiCnt = [8]string{humiANDCondition.Humi1_Cnt, humiANDCondition.Humi2_Cnt, humiANDCondition.Humi3_Cnt, humiANDCondition.Humi4_Cnt, humiANDCondition.Humi5_Cnt, humiANDCondition.Humi6_Cnt, humiANDCondition.Humi7_Cnt, humiANDCondition.Humi8_Cnt}
	var humiMode = [8]string{humiANDCondition.DeviceName1, humiANDCondition.DeviceName2, humiANDCondition.DeviceName3, humiANDCondition.DeviceName4, humiANDCondition.DeviceName5, humiANDCondition.DeviceName6, humiANDCondition.DeviceName7, humiANDCondition.DeviceName8}
	
	// 必要なcontrolDataを一括取得（N+1クエリ問題の解決）
	controlDataMap := make(map[string]ControlConditions)
	if humiANDCondition.Number > 0 {
		// 重複を除去したIDリストを作成
		deviceIDs := make([]string, 0, humiANDCondition.Number)
		seen := make(map[string]bool)
		for num := 0; num < humiANDCondition.Number; num++ {
			if humiCnt[num] != "" && !seen[humiCnt[num]] {
				deviceIDs = append(deviceIDs, humiCnt[num])
				seen[humiCnt[num]] = true
			}
		}
		// 一括取得
		if len(deviceIDs) > 0 {
			var controlDataList []ControlConditions
			if err := database.GormDB.Where("target_devices_id IN (?)", deviceIDs).Find(&controlDataList).Error; err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "controlData batch Read DB:%s", err.Error())
			} else {
				for _, cd := range controlDataList {
					controlDataMap[cd.TargetDevicesID] = cd
				}
			}
		}
	}
	
	var ret = [8]int8{1, 1, 1, 1, 1, 1, 1, 1}
	var num int = 0
	for num < humiANDCondition.Number {
		controlData, exists := controlDataMap[humiCnt[num]]
		if !exists {
			Logger.Writef(LOG_LEVEL_ERR, "controlData not found, Device ID:%s", humiCnt[num])
			num++
			continue
		}
		
		if demand {
			t = controlData.DemandStartTemperature
			h = controlData.DemandStartHumidity
			d = controlData.DemandStartDiscomfortIndex
		} else {
			t = controlData.ControlStartTemperature
			h = controlData.ControlStartHumidity
			d = controlData.ControlStartDiscomfortIndex
		}

		if t.Float64 > maxLimitTemperature || t.Float64 < minLimitTemperature {
			ret[num] = 1
		} else if humiMode[num][4:8] == "temp" {
			Logger.Writef(LOG_LEVEL_DEBUG, "AND ControlStartTemperature : %v, Temperature data : %v, Demand : %v", t, humiVal[num], demand)
			if t.Valid && (humiVal[num] <= t.Float64) { // 開始温度比較
				ret[num] = 0
			}
		} else if humiMode[num][4:8] == "humi" { // 湿度
			Logger.Writef(LOG_LEVEL_DEBUG, "AND ControlStartHumidity : %v, Humidity data : %v, Demand : %v", h, humiVal[num], demand)
			//if h.Valid && value.CurrentHumidity <= h.Float64 { // 開始湿度比較
			if h.Valid && humiVal[num] <= h.Float64 { // 開始湿度比較
				ret[num] = 0
			} //else if d.Valid && value.CurrentDiscomfortIndex <= d.Float64 { // 開始不快指数比較
			//ret[num] = 0
			//}
		}
		num++
	}

	var ret_and int8 = ret[0] & ret[1] & ret[2] & ret[3] & ret[4] & ret[5] & ret[6] & ret[7]
	// Logger.Writef(LOG_LEVEL_DEBUG, "AND condition ret data : %v, AND data : %d", ret, ret_and)

	return ret_and
}

func demandControlDeviceStopJudge(controlData ControlConditions, value PresentValue, mode string, category string) int8 {
	var t, h, d NullFloat64
	if controlData.DemandStartTemperature.Valid || controlData.DemandStartHumidity.Valid || controlData.DemandStartDiscomfortIndex.Valid {
		t = controlData.DemandStartTemperature
		h = controlData.DemandStartHumidity
		d = controlData.DemandStartDiscomfortIndex
	} else {
		return 1
	}
	//return controlDeviceStopJudge(t, h, d, value)
	return temphumicontrolDeviceStopJudge(t, h, d, value, true, mode, category)
}

func demandControlDeviceReleaseJudge(controlData ControlConditions, value PresentValue, mode string, category string) int8 {
	var t, h, d NullFloat64
	if controlData.DemandStopTemperature.Valid || controlData.DemandStopHumidity.Valid || controlData.DemandStopDiscomfortIndex.Valid {
		t = controlData.DemandStopTemperature
		h = controlData.DemandStopHumidity
		d = controlData.DemandStopDiscomfortIndex
	} else {
		return 0
	}
	//return controlDeviceReleaseJudge(t, h, d, value)
	return temphumicontrolDeviceReleaseJudge(t, h, d, value, true, mode, category)
}

func autoControlDeviceStopJudge(controlData ControlConditions, value PresentValue, mode string, category string) int8 {
	t := controlData.ControlStartTemperature
	h := controlData.ControlStartHumidity
	d := controlData.ControlStartDiscomfortIndex
	//return controlDeviceStopJudge(t, h, d, value)
	// Logger.Writef(LOG_LEVEL_DEBUG, "autoControlDeviceStopJudge start")
	return temphumicontrolDeviceStopJudge(t, h, d, value, false, mode, category)
}

func autoControlDeviceReleaseJudge(controlData ControlConditions, value PresentValue, mode string, category string) int8 {
	t := controlData.ControlStopTemperature
	h := controlData.ControlStopHumidity
	d := controlData.ControlStopDiscomfortIndex
	//return controlDeviceReleaseJudge(t, h, d, value)
	return temphumicontrolDeviceReleaseJudge(t, h, d, value, false, mode, category)
}

// 稼働必須時間経過判定
func isPastRequiredControllingTime(deviceInfo DeviceInformation, previousControlStatus ControlStatus) bool {
	// TODO: autoControlJudge の中で判定しているので不要だと思われる
	// 必要なら稼働必須時間と制御時間、制御解除時間を比較して長い方を使用するように変更する
	var controlTime int64
	if previousControlStatus.Control == 0 {
		controlTime = int64(time.Since(time.Unix(previousControlStatus.ControlEndTime, 0)).Seconds())
	} else {
		controlTime = int64(time.Since(time.Unix(previousControlStatus.ControlStartTime, 0)).Seconds())
	}

	// Logger.Writef(LOG_LEVEL_DEBUG, "RequiredControllingTime : %d, controlTime : %d", deviceInfo.RequiredControllingTime, controlTime)
	if deviceInfo.RequiredControllingTime < controlTime {
		return true
	} else {
		return false
	}
}
