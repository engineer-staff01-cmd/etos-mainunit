package app

import (
	"net/http"
	"sync"
	"time"
)

// クラウド通信関連の定数定義
const (
	// CloudEnergyDataSendInterval クラウドへのエネルギーセンサーデータ送信間隔（ミリ秒）
	// エネルギーセンサーデータの送信間隔
	CloudEnergyDataSendInterval = 50 * time.Millisecond

	// CloudDataSendInterval クラウドへのデータ送信間隔（ミリ秒）
	// 制御ステータスやその他のデータの送信間隔
	CloudDataSendInterval = 100 * time.Millisecond
)

var (
	ParentID = ""
	baseID   = ""
)

type CloudThread struct {
	to           chan ChannelMessage
	from         chan ChannelMessage
	serialNumber string
}

func outputContactStatusRegister() {
	// 出力接点データの登録
	var outputContactStatus []OutputContactStatus
	if err := database.SelectAll(&outputContactStatus); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select output contact status: %s", err.Error())
		return
	}

	// クラウドに送信
	for _, o := range outputContactStatus {
		// r, err := RequestOutputContact(baseID, o)
		// if err != nil {
		// 	// 送信失敗
		// 	Logger.Writef(LOG_LEVEL_ERR, "[ERROR] Post OutputContactStatus:%s", err.Error())

		// 	//クラウド通信異常のアラート
		// 	cloudComState := CloudCommonState{COM_NG}
		// 	database.CloudCommonStateSaveDB(cloudComState)
		// 	return
		// }

		// // 正常に終了
		// Logger.Writef(LOG_LEVEL_DEBUG, "RequestOutputContact response:%+v", r)
		err := database.GormDB.Where("time = ? AND sensor_id = ?", o.Time, o.SensorID).Delete(&o).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "outputContactStatus DB Delete:%s", err.Error())
		} else {
			//Logger.Writef(LOG_LEVEL_DEBUG, "outputContactStatus DB Delete:%+v", o)
		}
	}
}

func inputContactStatusRegister() {
	// 入力接点データの登録
	var inputContactStatus []InputContactStatus
	if err := database.SelectAll(&inputContactStatus); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select input contact status: %s", err.Error())
		return
	}

	// クラウドに送信
	for _, i := range inputContactStatus {
		// r, err := RequestInputContact(baseID, i)
		// if err != nil {
		// 	// 送信失敗
		// 	Logger.Writef(LOG_LEVEL_ERR, "[ERROR] Post inputContactStatus:%s", err.Error())

		// 	//クラウド通信異常のアラート
		// 	cloudComState := CloudCommonState{COM_NG}
		// 	database.CloudCommonStateSaveDB(cloudComState)
		// 	return
		// }

		// // 正常に終了
		// Logger.Writef(LOG_LEVEL_DEBUG, "RequestInputContacts response:%+v", r)
		err := database.GormDB.Where("time = ? AND sensor_id = ?", i.Time, i.SensorID).Delete(&i).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "inputContactStatus DB Delete:%s", err.Error())
		} else {
			//Logger.Writef(LOG_LEVEL_DEBUG, "inputContactStatus DB Delete:%+v", i)
		}
	}
}

func energyStatusElectricPoweregister(env Env) {

	tick := time.NewTicker(CloudEnergyDataSendInterval)

	var energyStatusElectricPower []EnergyStatusElectricPower
	if err := database.SelectAll(&energyStatusElectricPower); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select energy status electric power: %s", err.Error())
		return
	}
	//	Logger.Writef(LOG_LEVEL_DEBUG, "energyStatusElectricPower len : %d", len(energyStatusElectricPower))

	// クラウドに送信
	// Logger.Writef(LOG_LEVEL_DEBUG, "拠点 ID : エネルギーセンサー : %s", baseID)
Loop:
	for _, e := range energyStatusElectricPower {
		select {
		case <-tick.C:
			//		r, err := RequestEnergySensor(env, baseID, e)
			_, err := RequestEnergySensor(env, baseID, e)
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "Post EnergyStatusElectricPower:%s", err.Error())
				//クラウド通信異常のアラート
				cloudComState := CloudCommonState{COM_NG}
				database.CloudCommonStateSaveDB(cloudComState)
				break Loop
				//				return
			}
			//Logger.Writef(LOG_LEVEL_DEBUG, "Post EnergyStatusElectricPower Success:%+v", e)

			// 正常に終了
			//		Logger.Writef(LOG_LEVEL_DEBUG, "energyStatusElectricPower response:%+v", r)
			err = database.GormDB.Where("time = ? AND energy_sensor_id = ?", e.Time, e.EnergySensorID).Delete(&e).Error
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "energyStatusElectricPower DB Delete:%s", err.Error())
			} else {
				//Logger.Writef(LOG_LEVEL_DEBUG, "energyStatusElectricPower DB Delete:%+v", e)
			}
		}
	}
	tick.Stop()
}

func environmentDataRegister(env Env) {
	var environmentSensorStatus []EnvironmentSensorStatus
	if err := database.SelectAll(&environmentSensorStatus); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select environment sensor status: %s", err.Error())
		return
	}

	// クラウドに送信
	// Logger.Writef(LOG_LEVEL_DEBUG, "拠点 ID : 環境センサー : %s", baseID)
	for _, e := range environmentSensorStatus {
		//		r, err := RequestEnvironmentalSensor(env, baseID, e)
		_, err := RequestEnvironmentalSensor(env, baseID, e)
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Post EnvironmentSensorStatus:%s", err.Error())

			//クラウド通信異常のアラート
			cloudComState := CloudCommonState{COM_NG}
			database.CloudCommonStateSaveDB(cloudComState)
			return
		}
		//		Logger.Writef(LOG_LEVEL_DEBUG, "Post EnvironmentSensorStatus Success:%+v", e)

		// 正常に終了
		//		Logger.Writef(LOG_LEVEL_DEBUG, "RequestEnvironmentalSensors response:%+v", r)
		err = database.GormDB.Where("time = ? AND sensor_id = ?", e.Time, e.SensorID).Delete(&e).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "environmentSensorStatus DB Delete:%s", err.Error())
		} else {
			//Logger.Writef(LOG_LEVEL_DEBUG, "environmentSensorStatus DB Delete:%+v", e)
		}
	}
}

func sensorMonitorAlertRegister(env Env) {
	var sensorMonitorAlert []SensorMonitorAlert
	if err := database.SelectAll(&sensorMonitorAlert); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select sensor monitor alert: %s", err.Error())
		return
	}
	// Logger.Writef(LOG_LEVEL_DEBUG, "拠点 ID : センサー監視アラート : %s", baseID)
	for _, s := range sensorMonitorAlert {
		r, err := RequestAlertSensor(env, baseID, s)
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Post SensorMonitorAlert:%s", err.Error())
			//クラウド通信異常のアラート
			cloudComState := CloudCommonState{COM_NG}
			database.CloudCommonStateSaveDB(cloudComState)
			return
		}

		Logger.Writef(LOG_LEVEL_DEBUG, "RequestAlertSensor response:%+v", r)
		result := database.GormDB.Where("time = ? AND sensor_id = ?", s.Time, s.SensorID).Delete(&s) //TODO:デバッグ用にコメントアウト
		if result.Error != nil {
			Logger.Writef(LOG_LEVEL_ERR, "sensorMonitorAlert DB Delete:%s", err.Error())
		} else {
			//Logger.Writef(LOG_LEVEL_DEBUG, "sensorMonitorAlert DB Delete:%+v", s)
		}
	}
}

func upsAlertRegister(env Env) {
	var upsAlert []UpsAlert
	if err := database.SelectAll(&upsAlert); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select UPS alert: %s", err.Error())
		return
	}

	// Logger.Writef(LOG_LEVEL_DEBUG, "拠点 ID : 親機電源アラート : %s", baseID)
	for _, u := range upsAlert {
		r, err := RequestAlertUps(env, baseID, ParentID, u)
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Post UpsAlert:%s", err.Error())
			//クラウド通信異常のアラート
			cloudComState := CloudCommonState{COM_NG}
			database.CloudCommonStateSaveDB(cloudComState)
			return
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "RequestAlertUps response:%+v", r)
		result := database.GormDB.Where("time = ?", u.Time).Delete(&u) //TODO:デバッグ用にコメントアウト
		if result.Error != nil {
			Logger.Writef(LOG_LEVEL_ERR, "upsAlert DB Delete:%s", err.Error())
		} else {
			//Logger.Writef(LOG_LEVEL_DEBUG, "upsAlert DB Delete:%+v", u)
		}
	}
}

func watchalertThreadMessage(env Env) {
	// 入力接点データの登録
	inputContactStatusRegister()

	// エネルギーセンサーの登録
	energyStatusElectricPoweregister(env)

	// 環境センサーデータの登録
	environmentDataRegister(env)

	// センサーアラートの登録
	sensorMonitorAlertRegister(env)

	// UPSアラートの登録
	upsAlertRegister(env)
}

// 制御ステータスの登録
func deviceControlThreadMessage(env Env) {

	SendDeviceControlFlag = false

	tick := time.NewTicker(CloudDataSendInterval)

	// 制御ステータスの取得
	var deviceStatus []DeviceStatus
	if err := database.SelectAll(&deviceStatus); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select device status: %s", err.Error())
		return
	}

	// クラウドに送信
	// Logger.Writef(LOG_LEVEL_DEBUG, "拠点 ID : 機器別ステート : %s", baseID)
Loop:
	for _, d := range deviceStatus {
		select {
		case <-tick.C:
			//			r, err := RequestEquipmentState(env, baseID, d)
			_, err := RequestEquipmentState(env, baseID, d)
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "Post deviceStatus:%s", err.Error())
				//クラウド通信異常のアラート
				cloudComState := CloudCommonState{COM_NG}
				database.CloudCommonStateSaveDB(cloudComState)
				break Loop
				//				return
			}
			Logger.Writef(LOG_LEVEL_DEBUG, "Post deviceStatus Success:%+v", d) // 2023.5.12

			// 正常に終了
			//		Logger.Writef(LOG_LEVEL_DEBUG, "RequestEquipmentStates response:%+v", r)
			err = database.GormDB.Where("time = ? AND device_id = ?", d.Time, d.DeviceID).Delete(&d).Error
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "deviceStatus DB Delete:%s", err.Error())
			} else {
				//Logger.Writef(LOG_LEVEL_DEBUG, "deviceStatus DB Delete:%+v", d)
			}
		}
	}
	tick.Stop()
}

func energyStatusDemandRegister(env Env) {
	// デマンドデータ取得
	var energyStatusDemand []EnergyStatusDemand
	if err := database.SelectAll(&energyStatusDemand); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select energy status demand: %s", err.Error())
		return
	}

	// デマンドのIDの取得
	var demandPulseUnit DemandPulseUnit
	result := database.GormDB.Find(&demandPulseUnit)
	if result.Error != nil {
		Logger.Writef(LOG_LEVEL_ERR, "demandPulseUnit Cloud Read DB:%s", result.Error.Error())
		return
	}

	// デマンドデータ登録
	// Logger.Writef(LOG_LEVEL_DEBUG, "拠点 ID : デマンドデータ : %s", baseID)
	for _, e := range energyStatusDemand {
		Logger.Writef(LOG_LEVEL_DEBUG, "データ登録 EnergyStatusDemand:%+v", e)
		// クラウドに送信
		r, err := RequestDemand(env, baseID, demandPulseUnit.ID, e)
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Post EnergyStatusDemand : %s, Data : %+v", err.Error(), r)

			//クラウド通信異常のアラート
			cloudComState := CloudCommonState{COM_NG}
			database.CloudCommonStateSaveDB(cloudComState)
			return
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "RequestDemand:%+v", r)
		err = database.GormDB.Where("time = ?", e.Time).Delete(&e).Error //TODO:デバッグ用にコメントアウト
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "energyStatusDemand Delete DB:%s", err.Error())
		} else {
			//Logger.Writef(LOG_LEVEL_DEBUG, "energyStatusDemand Delete DB:%+v", e)
		}
	}
}

func energyAlertStatusDemandRegister(env Env) {
	// デマンドデータ取得
	var energyAlertStatusDemand []EnergyAlertStatusDemand
	if err := database.SelectAll(&energyAlertStatusDemand); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select energy alert status demand: %s", err.Error())
		return
	}

	// デマンドデータ登録
	// Logger.Writef(LOG_LEVEL_DEBUG, "拠点 ID : デマンドアラート : %s", baseID)
	for _, e := range energyAlertStatusDemand {
		// デマンドアラートの登録
		r, err := RequestPostDemandAlert(env, baseID, e)
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Post DemandAlert:%s", err.Error())

			//クラウド通信異常のアラート
			cloudComState := CloudCommonState{COM_NG}
			database.CloudCommonStateSaveDB(cloudComState)
			return
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "RequestPostDemandAlert:%+v", r)
		err = database.GormDB.Where("time = ?", e.Time).Delete(&e).Error //TODO:デバッグ用にコメントアウト
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "energyAlertStatusDemand Delete DB:%s", err.Error())
		} else {
			//Logger.Writef(LOG_LEVEL_DEBUG, "energyAlertStatusDemand Delete DB:%+v", e)
		}
	}
}

func batteryLevelAlertRegister(env Env) {
	// 電池残量取得
	var batteryLevelAlert []BatteryLevelAlert
	if err := database.SelectAll(&batteryLevelAlert); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select battery level alert: %s", err.Error())
		return
	}

	// 電池残量アラートの登録
	// Logger.Writef(LOG_LEVEL_DEBUG, "拠点 ID : 電池残量アラート : %s", baseID)
	for _, b := range batteryLevelAlert {
		r, err := RequestAlertBattery(env, baseID, b)
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Post BatteryLevelAlert:%s", err.Error())
			//クラウド通信異常のアラート
			cloudComState := CloudCommonState{COM_NG}
			database.CloudCommonStateSaveDB(cloudComState)
			return
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "RequestAlertBattery:%+v", r)
		err = database.GormDB.Where("time = ? AND sensor_id = ?", b.Time, b.SensorID).Delete(&b).Error //TODO:デバッグ用にコメントアウト
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "batteryLevelAlert Delete DB:%s", err.Error())
		} else {
			//Logger.Writef(LOG_LEVEL_DEBUG, "batteryLevelAlert Delete DB:%+v", b)
		}
	}
}

// `デマンドパルスユニットの登録`
func demandPulseThreadMessage(env Env) {
	// デマンドデータの登録
	energyStatusDemandRegister(env)

	// デマンドアラートの登録
	energyAlertStatusDemandRegister(env)

	// 電池残量アラートの登録
	batteryLevelAlertRegister(env)
}

// manualControlNotifyMessage 手動制御後にクラウドに対して通知を行う処理
func manualControlNotifyMessage(env Env, chmsg ChannelMessage) {
	releaseStatuses := []RemoteControlReleaseStatus{}
	if err := database.SelectAll(&releaseStatuses); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select release statuses: %s", err.Error())
		return
	}
	for _, status := range releaseStatuses {
		res, err := RequestEquipmentsControllerReleaseManualControlOrder(env, status.ID)
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to RequestEquipmentsControllerReleaseManualControlOrder:%s", err.Error())
			return
		}
		if res.StatusCode() >= http.StatusBadRequest {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to RequestEquipmentsControllerReleaseManualControlOrder code:%d body:%s", res.StatusCode(), string(res.Body))
			return
		}
		database.Delete(&status, "id = ?", status.ID)
		Logger.Writef(LOG_LEVEL_INFO, "Notify remote control mode is released. deviceID:%s", status.ID)
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "Remote control notification is done.")
}

// rebootNotifyMessage 再起動後にクラウドに対して通知を行う処理
// RequestParentDevicesControllerCompleteParentDeviceRestartが成功するまでリトライする
func rebootNotifyMessage(env Env) {
	for {
		res, err := RequestParentDevicesControllerCompleteParentDeviceRestart(env, ParentID)
		if err != nil {
			Logger.Writef(LOG_LEVEL_WARNING, "Failed to RequestParentDevicesControllerCompleteParentDeviceRestart. retrying in 5s: %s", err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		if res.StatusCode() >= http.StatusBadRequest {
			Logger.Writef(LOG_LEVEL_WARNING, "Failed to RequestParentDevicesControllerCompleteParentDeviceRestart code:%d body:%s. retrying in 5s", res.StatusCode(), string(res.Body))
			time.Sleep(5 * time.Second)
			continue
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "Reboot notification is done.")
		return
	}
}

// 通信異常の登録
func communicationAlertThread(env Env) {
	var communicationAlerts []CommunicationAlert
	if err := database.SelectAll(&communicationAlerts); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select communication alerts: %s", err.Error())
		return
	}

	// クラウドに送信
	// Logger.Writef(LOG_LEVEL_DEBUG, "拠点 ID : 通信アラート : %s", baseID)
	if len(communicationAlerts) > 0 {
		r, err := RequestAlertConnection(env, baseID, ParentID, communicationAlerts)
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Post communicationAlerts:%s", err.Error())
			//クラウド通信異常のアラート
			cloudComState := CloudCommonState{COM_NG}
			database.CloudCommonStateSaveDB(cloudComState)
		} else {
			Logger.Writef(LOG_LEVEL_DEBUG, "RequestAlertConnections Response:%+v", r)
			err = database.GormDB.Delete(&communicationAlerts).Error
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "RequestAlertConnections DB Delete:%s", err.Error())
			} else {
				//Logger.Writef(LOG_LEVEL_DEBUG, "RequestAlertConnections DB Delete:%+v", c)
			}
		}
	}
}

// システムアラートの登録
func masterAlertThread(env Env) {
	Logger.Writef(LOG_LEVEL_DEBUG, "==================== Cloud post masterAlertThread ====================")

	var sysAlerts []SystemAlert
	if err := database.SelectAll(&sysAlerts); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select system alerts: %s", err.Error())
		return
	}
	// Logger.Writef(LOG_LEVEL_DEBUG, "拠点 ID : 親機システムアラート : %s", baseID)
	for _, i := range sysAlerts {
		if i.Time != 0 {
			r, err := RequestPostAlertSystem(env, baseID, ParentID, i)
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "Post BatteryLevelAlert:%s", err.Error())
				//クラウド通信異常のアラート
				cloudComState := CloudCommonState{COM_NG}
				database.CloudCommonStateSaveDB(cloudComState)
			} else {
				database.Delete(&i, "time = ?", i.Time)
				Logger.Writef(LOG_LEVEL_DEBUG, "masterAlertThread:%+v", r)
			}
		}
	}
}

func alertMessage(env Env, chmsg ChannelMessage) {
	// スレッドごとに処理する
	switch chmsg.message {
	case strchildIn:
		//childInThreadMessage()

	case strchildOut:
		// 出力データの登録
		outputContactStatusRegister()

	case strwatchalert:
		// アラートの登録
		watchalertThreadMessage(env)

	case strdevicecontrol:
		// 制御ステータスの登録
		deviceControlThreadMessage(env)
		SendDeviceControlFlag = false

		// 子機/デマンドの通信異常のアラート
		communicationAlertThread(env)
		SendDeviceControlFlag = false

		// 親機のシステムアラート
		masterAlertThread(env)
		SendDeviceControlFlag = false

	case strdemandPulse:
		// デマンドパルスユニットの登録
		demandPulseThreadMessage(env)
	}

}
func notifyMessage(env Env, chmsg ChannelMessage) {
	switch chmsg.message {
	case strManualControl:
		manualControlNotifyMessage(env, chmsg)
	case strReboot:
		rebootNotifyMessage(env)

	default:
		Logger.Writef(LOG_LEVEL_ERR, "recieved unknown message at cloud thread. message:%v", chmsg.message)
	}
}

// dataGetMaster 親機に紐づくデータを取得
func dataGetMaster(env Env) error {
	parentDevicesControllerGetBase, err := RequestParentDevicesControllerGetBase(env, ParentID)
	if err != nil || parentDevicesControllerGetBase == nil {
		Logger.Writef(LOG_LEVEL_ERR, "ParentDevicesControllerGetBase:%s", err.Error())
		time.Sleep(1 * time.Second)
		return err
	}
	//Logger.Writef(LOG_LEVEL_DEBUG, "ParentDevicesControllerGetBase:%v", parentDevicesControllerGetBase)
	CloudtoDatabase(parentDevicesControllerGetBase)

	// 親機に紐づくデータから拠点識別を取得
	baseID = parentDevicesControllerGetBase.Base.ID
	return nil
}

// dataGetManualControl 機器の手動制御命令の取得
func dataGetManualControl(env Env) error {
	basesControllerGetManualControlOrders, err := RequestBasesControllerGetManualControlOrders(env, baseID)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "BasesControllerGetManualControlOrders:%s", err.Error())
		return err
	}

	remoteControls := make([]RemoteControl, len(basesControllerGetManualControlOrders))
	for i, v := range basesControllerGetManualControlOrders {
		remoteControls[i] = RemoteControl{v.EquipmentId, b2i8(v.Enabled)}
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "RemoteControl：%+v", remoteControls)
	database.RemoteControlDeleteAndSaveDB(remoteControls)

	return nil
}

// dataGetDefrost 出力接点の手動デフロスト命令の取得
func dataGetDefrost(env Env) error {
	basesControllerGetManualDefrostOrders, err := RequestBasesControllerGetManualDefrostOrders(env, baseID)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "BasesControllerGetManualDefrostOrders:%s", err.Error())
		return err
	}

	tx := database.GormDB.Begin()
	err = tx.Delete(&RemoteDefrostCommand{}).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "RemoteDefrostCommand delete DB: %+v", err.Error())
	}

	for _, v := range basesControllerGetManualDefrostOrders {
		remoteDefrost := RemoteDefrostCommand{v.OutputContactId, b2i8(v.Enabled)}
		err := tx.Save(&remoteDefrost).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "remoteDefrost Save DB: %+v", err.Error())
		}
		Logger.Writef(LOG_LEVEL_INFO, "RemoteDefrostCommand Cloud Save DB：%+v", remoteDefrost)
	}
	tx.Commit()

	return nil
}

// dataGetReboot 親機の再起動命令を取得
func dataGetReboot(env Env) error {
	parentDevicesControllerGetParentDeviceRestartOrder, err := RequestParentDevicesControllerGetParentDeviceRestartOrder(env, ParentID)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "ParentDevicesControllerGetParentDeviceRestartOrder:%s", err.Error())
		return err
	}
	rebootCommand := RebootCommand{
		ID:     REBOOT_COMMAND_ID,
		Enable: b2i8(parentDevicesControllerGetParentDeviceRestartOrder.Enabled),
	}
	if err := database.Save(&rebootCommand); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to save reboot command: %s", err.Error())
	}
	return nil
}

/*
// dataGetDemand デマンド制御命令の取得
func dataGetDemand(env Env) error {
	basesControllerGetDemandControlOrder, err := RequestBasesControllerGetDemandControlOrder(env, baseID)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "BasesControllerGetDemandControlOrder:%s", err.Error())
		return err
	}
	demandControl := DemandControl{
		ID:            DEMAND_CONTROL_ID,
		DemandControl: b2i8(basesControllerGetDemandControlOrder.Enabled),
	}
	if err := database.Save(&demandControl); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to save demand control: %s", err.Error())
	}

	return nil
}		/*  */

// dataGetMessage クラウドからデータ取得
func dataGetMessage(env Env, chmsg ChannelMessage) error {
	switch chmsg.message {
	case strMaster:
		return dataGetMaster(env)
	case strManualControl:
		return dataGetManualControl(env)
	case strDefrost:
		return dataGetDefrost(env)
	case strReboot:
		return dataGetReboot(env)
	case strDemand:
		//return dataGetDemand(env)		// クラウドからのデマンド命令を取得
		return nil
	}
	return nil
}

// NewCloud クラウド通信用のスレッド生成
func NewCloud(serialNumber string) *CloudThread {
	return &CloudThread{
		to:           make(chan ChannelMessage, CloudChannelBufferSizeTo),   // クラウドchanel(main -> cloud)
		from:         make(chan ChannelMessage, CloudChannelBufferSizeFrom), // クラウドchanel(main <- cloud)
		serialNumber: serialNumber,
	}
}

// CloudThreadRun クラウド通信スレッドのメインループ処理
func (c *CloudThread) CloudThreadRun(env Env) {
	ParentID = c.serialNumber
	notFirst := true
	for {
		chmsg := <-c.to
		//		Logger.Writef(LOG_LEVEL_DEBUG, "Cloud thread channel : len(%d), chmsg:%+v", len(c.to), chmsg)

		switch chmsg.messageType {
		case End:
			return
		case Alert:
			alertMessage(env, chmsg)
		case DataGet:
			err := dataGetMessage(env, chmsg)

			//クラウド通信状態の更新
			cloudComState := CloudCommonState{COM_OK}
			if err != nil {
				cloudComState.CommunicationError = COM_NG
				Logger.Writef(LOG_LEVEL_ERR, "cloudCommonState Save DB: %+v", err.Error())
			}
			database.CloudCommonStateSaveDB(cloudComState)
			if notFirst {
				if err == nil {
					// 初回のデータ取得成功
					sysAlert := GenerateSystemAlert(chmsg.time.Unix(), ERCD_SUCCESS, VERSION)
					RequestPostAlertSystem(env, baseID, ParentID, sysAlert)
					msg := ChannelMessage{FirstCommon, "true", chmsg.time}
					SendChannelMessageSafely(c.from, msg, false)
					notFirst = false
				} else {
					// 初回のデータ取得失敗
					msg := ChannelMessage{End, "false", chmsg.time}
					SendChannelMessageSafely(c.from, msg, false)
				}
			}
		case Notify:
			notifyMessage(env, chmsg)
		default:
			Logger.Writef(LOG_LEVEL_ERR, "recieved unknown message type at cloud thread. channel_message:%v", chmsg)
		}
	}
}

// Run クラウド通信スレッドのエントリーポイント
func (c *CloudThread) Run(env Env, wg *sync.WaitGroup) {
	Logger.Writef(LOG_LEVEL_DEBUG, "Start CloudThread")
	wg.Add(1) // goroutine起動前にカウントを増やす
	go func() {
		c.CloudThreadRun(env)
		Logger.Writef(LOG_LEVEL_DEBUG, "Stop CloudThread")
		wg.Done()
	}()
}
