package app

import (
	"sync"
)

/*
監視・アラート発報
*/

type WatchAlertThread struct {
	to   chan ChannelMessage
	from chan ChannelMessage
}

type alertSourceData struct {
	overAllCommonTime               int64
	baseInformation                 BaseInformation
	monitorConditionses             []MonitorConditions
	environmentSensores             []EnvironmentSensor
	energyAlertStatusElectricPowers []EnergyAlertStatusElectricPower
	environmentSensorStatuses       []EnvironmentSensorStatus
	inputContactAlertStatuses       []InputContactAlertStatus
	outputContactAlertStatus        []OutputContactAlertStatus
	upsAlertStatus                  UpsAlertStatus
}

// NewWatchAlert .
func NewWatchAlert() *WatchAlertThread {
	return &WatchAlertThread{
		to:   make(chan ChannelMessage, WatchAlertChannelBufferSize), // 監視(main -> watch)
		from: make(chan ChannelMessage, WatchAlertChannelBufferSize), // 監視(main <- watch)
	}
}

// Run .
func (w *WatchAlertThread) Run(wg *sync.WaitGroup, sendchCloudMsg, sendchMailMsg chan ChannelMessage) {
	Logger.Writef(LOG_LEVEL_DEBUG, "Start WatchAlertThread")
	wg.Add(1) // goroutine起動前にカウントを増やす
	go func() {
		w.watchAlertThreadRun(sendchCloudMsg, sendchMailMsg)
		Logger.Writef(LOG_LEVEL_DEBUG, "Stop WatchAlertThread")
		wg.Done()
	}()
}

func (w *WatchAlertThread) watchAlertThreadRun(sendchCloudMsg, sendchMailMsg chan ChannelMessage) {
	for {
		chmsg := <-w.to

		switch chmsg.messageType {
		case End:
			return

		case DataGet:
			w.watchAlertDataGet(chmsg, sendchCloudMsg, sendchMailMsg)

			// childUnitスレ終了
			msg := ChannelMessage{End, "", chmsg.time}
			SendChannelMessageSafely(w.from, msg, false)
		}
	}
}

func readAlertSourceData(t int64) *alertSourceData {
	asd := new(alertSourceData)
	asd.overAllCommonTime = t

	// 親機情報
	if err := database.SelectOne(&asd.baseInformation); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select base information: %s", err.Error())
	}

	// 監視情報
	if err := database.SelectAll(&asd.monitorConditionses); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select monitor conditions: %s", err.Error())
	}

	// 環境センサ情報
	if err := database.SelectAll(&asd.environmentSensores); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select environment sensors: %s", err.Error())
	}

	// エネルギーセンサステータス
	if err := database.SelectAll(&asd.energyAlertStatusElectricPowers); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select energy alert status electric powers: %s", err.Error())
	}

	// 環境センサステータス
	if err := database.SelectAll(&asd.environmentSensorStatuses); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select environment sensor statuses: %s", err.Error())
	}

	// 入力IOステータス
	if err := database.SelectAll(&asd.inputContactAlertStatuses); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select input contact alert statuses: %s", err.Error())
	}

	// 出力IOステータス
	if err := database.SelectAll(&asd.outputContactAlertStatus); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select output contact alert status: %s", err.Error())
	}

	// アラートステータス
	if err := database.SelectOne(&asd.upsAlertStatus); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select UPS alert status: %s", err.Error())
	}

	return asd
}

func alertEnergyStatus(asd *alertSourceData, mc MonitorConditions) int {
	count := 0

	// 機器用のエネルギーセンサーステータス取得
	energyStatusElectricPower, _ := SearchEnergyStatusElectricPower(mc.EnergySensorID, asd.overAllCommonTime, asd.energyAlertStatusElectricPowers)
	Logger.Writef(LOG_LEVEL_DEBUG, "energyStatusElectricPower:%+v", energyStatusElectricPower)

	var energysensor EnergySensor
	err := database.GormDB.Where("id = ?", mc.EnergySensorID).First(&energysensor).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "energysensor Cloud Read DB:%s", err.Error())
	}

	// センサ、子機無効処理
	if IsEnableChildUnit(energysensor.ChildDeviceID) && i82b(energysensor.Enable) && energyStatusElectricPower.Updated > 0 {
		//Logger.Writef(LOG_LEVEL_DEBUG, "energysensorJudge:%s", energysensor.ID)
		// デバイスインフォメーションのidを取得
		var deviceInformation DeviceInformation
		err := database.GormDB.Where("energy_sensor_id = ?", energysensor.ID).First(&deviceInformation).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "deviceInformation Cloud Read DB:id=%s, %s", energysensor.ID, err.Error())
		} else {
			//Logger.Writef(LOG_LEVEL_DEBUG, "EnergyJudge:%s", energysensor.ID)
			deviceid := deviceInformation.ID
			// 電力監視
			isAlert := EnergyJudge(mc, energyStatusElectricPower, energysensor, deviceid, asd.overAllCommonTime)
			if isAlert {
				count++
			}
		}
	}

	return count
}

func alertEnvStatus(asd *alertSourceData, mc MonitorConditions) int {
	count := 0

	// 機器用の環境センサ情報の取得
	environmentSensor, _ := SearchEnvironmentSensor(mc.EnvironmentSensorID, asd.environmentSensores)
	// 機器用の環境センサステータス取得
	environmentSensorStatus, _ := SearchEnvironmentSensorStatus(mc.EnvironmentSensorID, asd.overAllCommonTime, asd.environmentSensorStatuses)

	if IsEnableChildUnit(environmentSensor.ChildDeviceID) && i82b(environmentSensor.Enable) && environmentSensorStatus.Time > 0 {
		// センサー監視
		//Logger.Writef(LOG_LEVEL_DEBUG, "EnvironmentJudge")

		// デバイスインフォメーションのidを取得
		var deviceEnvironmentList DeviceEnvironmentInformation
		err := database.GormDB.Where("sensor_id = ?", environmentSensor.ID).First(&deviceEnvironmentList).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "deviceEnvironmentList Cloud Read DB:%s", err.Error())
		} else {
			deviceid := deviceEnvironmentList.DeviceID
			// 環境センサ
			isAlert := EnvironmentJudge(mc, environmentSensorStatus, environmentSensor, deviceid, asd.overAllCommonTime)
			if isAlert {
				count++
			}
		}
	}
	return count
}

func alertInputContact(asd *alertSourceData, mc MonitorConditions) int {
	count := 0

	// 入力接点アラート判定
	inputContactAlertStatus, _ := SearchInputContactStatus(mc.InputContactID, asd.overAllCommonTime, asd.inputContactAlertStatuses)
	Logger.Writef(LOG_LEVEL_DEBUG, "inputContactAlertStatus:%+v", inputContactAlertStatus)
	if inputContactAlertStatus.ID == "" {
		return 0
	}

	var inputContact InputContact
	err := database.GormDB.Where("id = ?", mc.InputContactID).First(&inputContact).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "energysensor Cloud Read DB:%s", err.Error())
	}

	if IsEnableChildUnit(inputContact.ChildDeviceID) && i82b(inputContact.Enable) {

		// 入力接点監視 ※発生時間リストを作成
		isAlert := InputContactJudge(mc, inputContactAlertStatus, inputContact, asd.overAllCommonTime)
		if isAlert {
			count++
		}
	}

	return count
}

func alertOutputContact(asd *alertSourceData, mc MonitorConditions) int {
	count := 0

	// 出力接点アラート判定
	var outputContact OutputContact
	err := database.GormDB.Where("id = ?", mc.OutputContactID).First(&outputContact).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "energysensor Cloud Read DB:%s", err.Error())
	}

	outputContactAlertStatus, err := SearchOutputContactAlertStatus(mc.OutputContactID, asd.overAllCommonTime, asd.outputContactAlertStatus)
	if err != nil {
		return 0
	}
	if IsEnableChildUnit(outputContact.ChildDeviceID) && i82b(outputContact.Enable) {
		// 出力接点監視 ※発生時間リストを作成
		isAlert := OutputContactJudge(mc, outputContactAlertStatus, outputContact, asd.overAllCommonTime)
		if isAlert {
			count++
		}
	}

	return count
}

// 2022-09-16
// 対応方針としてスケジュールで無効になった(APIのレスポンスに含まれなくなった)
// センサー監視ステータスをリセットするようにする
//
// 変更前
// センサー値が異常を検出した時点から異常継続時間経過した際にアラートが発報
// 変更後
// スケジュールで有効になった時点から異常継続時間経過した際にアラートが発報
func removeDisabledSensorMonitorAlert(monitorConditions []MonitorConditions, sensorMonitorAlertStatuses []SensorMonitorAlertStatus) {
	unusedStatusIDs := []string{}

	for _, status := range sensorMonitorAlertStatuses {
		found := false
		for _, mc := range monitorConditions {
			if status.ID == mc.ID {
				found = true
				break
			}
		}

		// デマンドパルスの監視は削除しない
		if status.Category == strDemandPulseUnit {
			continue
		}

		if !found {
			unusedStatusIDs = append(unusedStatusIDs, status.ID)
		}
	}

	for _, unusedStatusID := range unusedStatusIDs {
		database.Delete(&SensorMonitorAlertStatus{}, "id = ?", unusedStatusID)
		Logger.Writef(LOG_LEVEL_DEBUG, "Remove disabled SensorMonitorAlertStatus id:%s", unusedStatusID)
	}
}

func (w *WatchAlertThread) watchAlertDataGet(chmsg ChannelMessage, sendchCloudMsg, sendchMailMsg chan ChannelMessage) {
	//---------------------
	// DB取得
	//---------------------
	asd := readAlertSourceData(chmsg.time.Unix())

	// スケジュールで無効に設定されたセンサー監視ステータスを削除
	var sensorMonitorAlertStatuses []SensorMonitorAlertStatus
	if err := database.SelectAll(&sensorMonitorAlertStatuses); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select sensor monitor alert statuses: %s", err.Error())
		return
	}
	removeDisabledSensorMonitorAlert(asd.monitorConditionses, sensorMonitorAlertStatuses)

	//---------------------
	// 判定
	//---------------------
	count := 0

	// 社屋停電監視 UPS確認
	output, err := GetUPSStatus()
	//	Logger.Writef(LOG_LEVEL_DEBUG, "UPS STATUS :%+v", output)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "failed to read UPS status:%s", err.Error())
	}
	isAlert := PowerOutageJudge(asd.upsAlertStatus.LastStatus, output, asd.overAllCommonTime)
	if isAlert {
		count++
	}

	// 機器別で開始
	for _, mc := range asd.monitorConditionses {
		// 監視無効
		if mc.Enable == 0 {
			continue
		}

		Logger.Writef(LOG_LEVEL_DEBUG, "Sensor Alert Judge Start :%+v", mc)

		// 電力監視
		if mc.EnergySensorID != "" {
			//Logger.Writef(LOG_LEVEL_DEBUG, "============== Sensor Alert Judge Start EnergySensor ID:%s ==============", mc.EnergySensorID)
			//Logger.Writef(LOG_LEVEL_DEBUG, "overAllCommonTime:%d", asd.overAllCommonTime)
			count += alertEnergyStatus(asd, mc)
		}

		// 環境監視
		if mc.EnvironmentSensorID != "" {
			//Logger.Writef(LOG_LEVEL_DEBUG, "============== Sensor Alert Judge Start EnvironmentSensor ID:%s ==============", mc.EnvironmentSensorID)
			//Logger.Writef(LOG_LEVEL_DEBUG, "overAllCommonTime:%d", asd.overAllCommonTime)
			count += alertEnvStatus(asd, mc)
		}

		// 入力監視
		if mc.InputContactID != "" {
			//Logger.Writef(LOG_LEVEL_DEBUG, "============== Sensor Alert Judge Start InputContact ID:%s ==============", mc.InputContactID)
			//Logger.Writef(LOG_LEVEL_DEBUG, "overAllCommonTime:%d", asd.overAllCommonTime)
			count += alertInputContact(asd, mc)
		}

		// 出力監視
		if mc.OutputContactID != "" {
			//Logger.Writef(LOG_LEVEL_DEBUG, "============== Sensor Alert Judge Start OutputContact ID:%s ==============", mc.OutputContactID)
			//Logger.Writef(LOG_LEVEL_DEBUG, "overAllCommonTime:%d", asd.overAllCommonTime)
			count += alertOutputContact(asd, mc)
		}
	}

	// 子機通信異常監視
	ChildDeviceCommunicationAlert()

	Logger.Writef(LOG_LEVEL_DEBUG, "Alert count:%v", count)

	// cloudスレに送る
	{
		msg := ChannelMessage{Alert, strwatchalert, chmsg.time}
		SendChannelMessageSafely(sendchCloudMsg, msg, false)
	}
	// メール送信スレッドに送る
	{
		msg := ChannelMessage{Alert, "true", chmsg.time}
		SendChannelMessageSafely(sendchMailMsg, msg, false)
	}
}
