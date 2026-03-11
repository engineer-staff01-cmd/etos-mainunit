package app

import (
	"math"
	"strings"
	"time"
)

/* アラート復旧の待ち時間 */
var alertCancel int64 = 0

func FloatRound(val float64) (ret float64) {
	ret = math.Round(val*1000.0) / 1000.0
	return
}

/*
デマンド監視
引数：現在電力、調整電力、限界電力、契約電力
戻り値：発生時間、発生フラグ
警戒：調整電力 ≧ 0
限界：現在電力 ≧ 限界電力
超過：現在電力 ≧ 契約電力
*/
func DemandJudge(demandPulseID string, demandData ReturnValueCurrentPower, overAllCommonTime int64) {
	var alert EnergyStatusDemand
	//var dc DemandControl

	//database.GormDB.First(&dc)

	if demandData.currentPower >= demandData.ContactElectricPower {
		// 超過（現在電力 ≧ 契約電力）
		alert.AlarmStatus = StrExcess
	} else if demandData.currentPower >= demandData.LimitElectricPower {
		// 限界（現在電力 ≧ 限界電力）
		alert.AlarmStatus = Strlimit
	} else if DemandEnableVal == 1 {
		//} else if dc.DemandControl == 1 {
		if demandData.currentPower-demandData.targetCurrentPower < demandData.CancellationElectricPower {
			alert.AlarmStatus = StrRecover
		} else {
			alert.AlarmStatus = StrBeVigilant
		}
	} else if demandData.currentPower >= demandData.InitialElectricPower && demandData.adjustedPower >= 0 {
		// 逼迫・警戒（調整電力 ≧ 0） → デマンド制御開始
		alert.AlarmStatus = StrBeVigilant
	} else {
		alert.AlarmStatus = StrNormal
	}
	alert.Time = overAllCommonTime
	alert.CurrentElectricPower = demandData.currentPower
	alert.PredictedElectricPower = demandData.predictedPower
	alert.AdjustedElectricPower = demandData.adjustedPower
	//alert.CurrentElectricPower = FloatRound(demandData.currentPower)
	//alert.PredictedElectricPower = FloatRound(demandData.predictedPower)
	//alert.AdjustedElectricPower = FloatRound(demandData.adjustedPower)
	alert.TimeLeft = int64(demandData.leftTime)

	if err := database.Save(&alert); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to save demand alert: %s", err.Error())
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "Demand judge:%+v ", alert)

	var prevStatus DemandAlertStatus
	database.GormDB.Where("id = ?", demandPulseID).First(&prevStatus)
	status := DemandAlertStatus{
		ID:     demandPulseID,
		Time:   overAllCommonTime,
		Status: alert.AlarmStatus,
	}
	if err := database.Save(&status); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to save demand alert status: %s", err.Error())
	}

	// デマンドアラート用（通常以外で発報）
	if alert.AlarmStatus != prevStatus.Status && alert.AlarmStatus != StrNormal {
		var a EnergyAlertStatusDemand
		a.AlarmStatus = alert.AlarmStatus
		a.Time = alert.Time
		a.CurrentElectricPower = alert.CurrentElectricPower
		a.PredictedElectricPower = alert.PredictedElectricPower
		a.AdjustedElectricPower = alert.AdjustedElectricPower
		//a.CurrentElectricPower = demandData.currentPower
		//a.PredictedElectricPower = demandData.predictedPower
		//a.AdjustedElectricPower = demandData.adjustedPower
		a.TimeLeft = alert.TimeLeft
		a.ContactElectricPower = demandData.ContactElectricPower
		a.TargetElectricPower = demandData.TargetElectricPower
		if err := database.Save(&a); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save energy alert status demand: %s", err.Error())
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "Demand Alert judge:%+v ", a)

		mailAlert := DemandStatusMailAlert{
			Time:                   alert.Time,
			CurrentElectricPower:   alert.CurrentElectricPower,
			PredictedElectricPower: alert.PredictedElectricPower,
			AdjustedElectricPower:  alert.AdjustedElectricPower,
			//CurrentElectricPower:   demandData.currentPower,
			//PredictedElectricPower: demandData.predictedPower,
			//AdjustedElectricPower:  demandData.adjustedPower,
			TargetElectricPower:   demandData.TargetElectricPower,
			LimitElectricPower:    demandData.LimitElectricPower,
			ContractElectricPower: demandData.ContactElectricPower,
			AlarmStatus:           alert.AlarmStatus,
		}
		if err := database.Save(&mailAlert); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save demand status mail alert: %s", err.Error())
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "Demand Mail Alert judge:%+v ", a)

		if alert.AlarmStatus != StrRecover {
			DemandEnableVal = 1
		} else {
			DemandEnableVal = 0
		}
	} else if alert.AlarmStatus == StrNormal || alert.AlarmStatus == StrRecover {
		DemandEnableVal = 0
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "DemandEnableVal : %d", DemandEnableVal)
}

/*
入力接点監視
*/
func InputContactJudge(m MonitorConditions, status InputContactAlertStatus, inputContact InputContact, overAllCommonTime int64) bool {
	Logger.Writef(LOG_LEVEL_DEBUG, "InputContactJudge")

	var alertStatus SensorMonitorAlertStatus
	result := database.GormDB.Where("id = ?", m.ID).First(&alertStatus)
	if result.Error != nil {
		Logger.Writef(LOG_LEVEL_ERR, "SensorMonitorAlertStatus Read input DB:%s", result.Error.Error())
	} else {
		Logger.Writef(LOG_LEVEL_DEBUG, "SensorMonitorAlertStatus Read input DB:%+v", alertStatus)
	}

	if alertStatus.Category != m.MonitorCategory {
		alertStatus.ID = m.ID
		alertStatus.Category = m.MonitorCategory
		alertStatus.StartTime = 0
		alertStatus.LastStatus = strRestoration
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "SensorMonitorAlertStatus first status:%+v", alertStatus)

	m.Threshold = 0.5
	ret := sensorValueJudge(status.ID, m, &alertStatus, float64(status.Status), overAllCommonTime, strInputContact)
	return ret
}

/*
出力接点監視
*/
func OutputContactJudge(m MonitorConditions, status OutputContactAlertStatus, outputContact OutputContact, overAllCommonTime int64) bool {
	//Logger.Writef(LOG_LEVEL_DEBUG, "OutputContactJudge")

	var alertStatus SensorMonitorAlertStatus
	result := database.GormDB.Where("id = ?", m.ID).First(&alertStatus)
	if result.Error != nil {
		Logger.Writef(LOG_LEVEL_ERR, "SensorMonitorAlertStatus Read output DB:%s", result.Error.Error())
	} else {
		Logger.Writef(LOG_LEVEL_DEBUG, "SensorMonitorAlertStatus Read output DB:%+v", alertStatus)
	}

	if alertStatus.Category != m.MonitorCategory {
		alertStatus.ID = m.ID
		alertStatus.Category = m.MonitorCategory
		alertStatus.StartTime = 0
		alertStatus.LastStatus = strRestoration
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "SensorMonitorAlertStatus first status:%+v", alertStatus)

	m.Threshold = 0.5
	ret := sensorValueJudge(status.ID, m, &alertStatus, float64(status.Status), overAllCommonTime, strOutputContact)
	return ret
}

/*
社屋停電監視
引数：
戻り値：発生時間、発生フラグ
*/
var commerr_cnt int = 0
var commerr_flag bool = false
var battery_flag bool = false

func PowerOutageJudge(oldStatus, output string, overAllCommonTime int64) bool {
	nowStatus := ReadPowerOutageStatus(output)

	// UPS異常検知用のステータスを保存
	upsAlertStatus := UpsAlertStatus{
		ID:         UPS_ALERT_STATUS_ID,
		Time:       time.Now().Unix(),
		Message:    output,
		LastStatus: nowStatus,
	}
	if err := database.Save(&upsAlertStatus); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to save UPS alert status: %s", err.Error())
	}

	Logger.Writef(LOG_LEVEL_DEBUG, "UPS OLD Status:%+v", oldStatus)
	Logger.Writef(LOG_LEVEL_DEBUG, "UPS NOW Status:%+v", nowStatus)
	// ステータス変化がない場合アラート発報しない
	if oldStatus == nowStatus {
		if !(strings.Contains(nowStatus, UPS_STATUS_BATTERY)) &&
			!(strings.Contains(nowStatus, UPS_STATUS_COMMERR)) {
			return false
		}
	}

	// 以前のステータスが存在しない、かつ現在ONLINEの場合は発報しない
	//	if oldStatus == "" && strings.Contains(nowStatus, UPS_STATUS_ONLINE) {
	//		return false
	//	}
	// ステータス変更があればメールアラートを送信
	//	upsMailAlert := UpsMailAlert{
	//		Time:   overAllCommonTime,
	//		Status: nowStatus,
	//	}
	//	database.Save(&upsMailAlert)

	// 電源供給されている場合、発報しない
	//	if strings.Contains(nowStatus, UPS_STATUS_ONLINE) ||
	//		strings.Contains(nowStatus, UPS_STATUS_TRIM) ||
	//		strings.Contains(nowStatus, UPS_STATUS_BOOST) {
	//		return false
	//	}

	if strings.Contains(nowStatus, UPS_STATUS_BATTERY) {
		// メールアラートを送信
		commerr_cnt = 0
		commerr_flag = false
		if oldStatus == nowStatus && !battery_flag {
			battery_flag = true
			upsMailAlert := UpsMailAlert{
				Time:   overAllCommonTime,
				Status: nowStatus,
			}
			if err := database.Save(&upsMailAlert); err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "Failed to save UPS mail alert: %s", err.Error())
			}
			// バッテリーに切り替わったとき
			upsAlert := UpsAlert{
				Time:    overAllCommonTime,
				Message: "UPS alert : battery powered",
			}
			if err := database.Save(&upsAlert); err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "Failed to save UPS alert: %s", err.Error())
			}
		} else {
			return false
		}
	} else if strings.Contains(nowStatus, UPS_STATUS_OVERLOAD) ||
		strings.Contains(nowStatus, UPS_STATUS_LOWBATT) ||
		strings.Contains(nowStatus, UPS_STATUS_REPLACEBATT) ||
		strings.Contains(nowStatus, UPS_STATUS_NOBATT) {
		commerr_cnt = 0
		commerr_flag = false
		battery_flag = false
		sysAlert := GenerateSystemAlert(overAllCommonTime, ERCD_UPS_ERROR, strings.Replace(nowStatus, "ONLINE ", "", 1))
		if err := database.Save(&sysAlert); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save system alert: %s", err.Error())
		}
		//	} else if strings.Contains(nowStatus, UPS_STATUS_COMMLOST) ||
	} else if strings.Contains(nowStatus, UPS_STATUS_COMMLOST) {
		//		strings.Contains(nowStatus, UPS_STATUS_COMMERR) {
		commerr_cnt = 0
		commerr_flag = false
		battery_flag = true
		sysAlert := GenerateSystemAlert(overAllCommonTime, ERCD_UPS_ERROR, nowStatus)
		if err := database.Save(&sysAlert); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save system alert: %s", err.Error())
		}
	} else if strings.Contains(nowStatus, UPS_STATUS_ONLINE) ||
		strings.Contains(nowStatus, UPS_STATUS_TRIM) ||
		strings.Contains(nowStatus, UPS_STATUS_BOOST) {
		commerr_cnt = 0
		commerr_flag = false
		battery_flag = false
		return false
	} else if strings.Contains(nowStatus, UPS_STATUS_COMMERR) {
		battery_flag = false
		Logger.Writef(LOG_LEVEL_DEBUG, "UPS Status : COMMERR count : %d", commerr_cnt)
		if commerr_cnt > (60*24) && !commerr_flag {
			commerr_flag = true
			commerr_cnt = 0
			sysAlert := GenerateSystemAlert(overAllCommonTime, ERCD_UPS_ERROR, nowStatus)
			if err := database.Save(&sysAlert); err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "Failed to save system alert: %s", err.Error())
			}
		} else {
			commerr_cnt++
			return false
		}
	} else if nowStatus == "" {
		Logger.Writef(LOG_LEVEL_DEBUG, "UPS Status : null")
		commerr_cnt = 0
		commerr_flag = false
		battery_flag = false
		return false
	} else {
		commerr_cnt = 0
		commerr_flag = false
		battery_flag = false
		sysAlert := GenerateSystemAlert(overAllCommonTime, ERCD_UPS_ERROR, nowStatus)
		if err := database.Save(&sysAlert); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save system alert: %s", err.Error())
		}
	}

	return true
}

/*
電池残量監視
*/
func BatteryLevelAlertJudge(demandPulseUnit DemandPulseUnit, demandStatus DemandStatus, overAllCommonTime int64) {
	//Logger.Writef(LOG_LEVEL_DEBUG, "demandPulseUnit Threshold:%f , Voltage:%f", demandPulseUnit.Threshold, demandStatus.Voltage)
	var alertStatus SensorMonitorAlertStatus
	err := database.GormDB.Where("id = ? AND category = ?", demandPulseUnit.ID, strDemandPulseUnit).First(&alertStatus).Error
	if err != nil {
		// DB上に無い場合もあるためログしない
		// Logger.Writef(LOG_LEVEL_ERR, "SensorMonitorAlertStatus Read environment DB:%s", err.Error())
	} else {
		//Logger.Writef(LOG_LEVEL_DEBUG, "SensorMonitorAlertStatus Read DB:%+v", alertStatus)
	}

	if alertStatus.Category != strDemandPulseUnit {
		alertStatus.ID = demandPulseUnit.ID
		alertStatus.Category = strDemandPulseUnit
		alertStatus.StartTime = 0
		alertStatus.LastStatus = strRestoration
	}
	// 最大電圧[V] - (最大電圧[V] - 最小電圧[V]) * (しきい値[%])
	threshold := 3.6 - (3.6-2.7)*(1.0-float64(demandPulseUnit.Threshold)/100)
	//Logger.Writef(LOG_LEVEL_DEBUG, "threshold = %f %f", demandPulseUnit.Threshold, threshold)
	isAlert := underSensorValueJudge(demandStatus.Voltage, threshold, &alertStatus, 1)

	if isAlert {
		var alert BatteryLevelAlert
		alert.SensorID = demandPulseUnit.ID
		alert.Time = alertStatus.OccurredAt
		alert.kind = strDemandPulseUnit
		if err := database.Save(&alert); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save battery level alert: %s", err.Error())
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "BatteryLevelAlert Save battery DB:%+v", alert)
	}

	if demandStatus.Voltage < threshold {
		UpdateBatteryLevelMailAlertStatus(demandPulseUnit.ID, overAllCommonTime, VOLTAGE_NG)
		mailAlert := BatteryLevelMailAlert{
			Time:                  overAllCommonTime,
			SensorID:              demandPulseUnit.ID,
			kind:                  strDemandPulseUnit,
			JudgementAbnormalTime: demandPulseUnit.JudgementAbnormalTime,
		}
		database.BatteryLevelMailAlertSaveDB(mailAlert)
		Logger.Writef(LOG_LEVEL_DEBUG, "BatteryLevelMailAlert Save battery DB:%+v", mailAlert)
	} else {
		UpdateBatteryLevelMailAlertStatus(demandPulseUnit.ID, overAllCommonTime, VOLTAGE_OK)
	}
}

/*
通信接続監視
*/
func SaveDBCommunicationAlert(id, sensorName string, overAllCommonTime int64, kind string) {
	if id != "" && overAllCommonTime != 0 {
		var alert CommunicationAlert
		alert.SensorID = id
		alert.Time = overAllCommonTime
		alert.Kind = kind
		database.CommunicationAlertCloudSaveDB(overAllCommonTime, alert)

		mailAlert := CommunicationMailAlert{
			Time:       overAllCommonTime,
			SensorID:   id,
			SensorName: sensorName,
			Kind:       kind,
		}
		database.CommunicationMailAlertSaveDB(mailAlert)
	}
}

func SaveDBChildDeviceCommunicationAlert(childDeviceId string, kind string, status int) {
	// 子機別センサー通信状況の取得
	var childDeviceCommonState ChildDeviceCommonState
	database.SelectByQuery(&childDeviceCommonState, "id = ?", childDeviceId)
	// DB上にない場合があるのでIDは設定する
	childDeviceCommonState.ID = childDeviceId

	switch kind {
	case strEnergySensor:
		childDeviceCommonState.EnergySensor = status
	case strEnvironmentalSensor:
		childDeviceCommonState.EnvironmentSensor = status
	case strInputContact:
		childDeviceCommonState.InputContact = status
	case strOutputContact:
		childDeviceCommonState.OutputContact = status
	default:
		return
	}
	if err := database.Save(&childDeviceCommonState); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to save child device common state: %s", err.Error())
	}
}

// 子機通信異常判定
func isChildUnitCommunicationError(c ChildDeviceCommonState) bool {
	if c.InputContact == 0 {
		return false
	}
	if c.OutputContact == 0 {
		return false
	}
	if c.EnergySensor == 0 {
		return false
	}
	if c.EnvironmentSensor == 0 {
		return false
	}

	// 異常
	return true
}

// 子機通信異常監視
func ChildDeviceCommunicationAlert() {
	// 子機通信異常（IO,電力、環境すべてが通信異常の場合に発報）

	// 子機IDを取得
	var childUnit []ChildUnit
	if err := database.SelectAll(&childUnit); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select child unit: %s", err.Error())
		return
	}

	// 子機IDに紐づく通信
	for _, child := range childUnit {
		if child.Enable == 0 {
			continue
		}
		var childUnitCommonState ChildDeviceCommonState
		result := database.GormDB.Where("id = ?", child.ID).First(&childUnitCommonState)
		if result.Error != nil {
			Logger.Writef(LOG_LEVEL_ERR, "childUnitCommonState Read DB:%s", result.Error.Error())
			continue
		}
		// 通信異常判定
		overAllCommonTime := time.Now().Unix()
		if isChildUnitCommunicationError(childUnitCommonState) {
			UpdateCommunicationMailAlertStatus(child.ID, strChildUnit, overAllCommonTime, COM_NG)
			SaveDBCommunicationAlert(child.ID, child.Name, overAllCommonTime, strChildUnit)
		} else {
			UpdateCommunicationMailAlertStatus(child.ID, strChildUnit, overAllCommonTime, COM_OK)
		}
	}
}

/*
センサー監視のアラートデータ作成
*/
func sensorAlertCreate(id string, alertStatus *SensorMonitorAlertStatus, value float64, overAllCommonTime int64, monitorid string, kind string) (alert SensorMonitorAlert) {
	alert.SensorID = id
	alert.Time = overAllCommonTime
	alert.OccurredAt = alertStatus.OccurredAt
	alert.Status = alertStatus.LastStatus
	alert.Value = value
	alert.MonitorConditionsID = monitorid
	alert.Kind = kind
	return alert
}

/*
センサ監視の電力判定(電流 / 電圧 / 電力 / 力率)
*/
func EnergyJudge(m MonitorConditions, status EnergyAlertStatusElectricPower, energysensor EnergySensor, deviceid string, overAllCommonTime int64) (ret bool) {
	//Logger.Writef(LOG_LEVEL_DEBUG, "EnergyJudge")

	var alertStatus SensorMonitorAlertStatus
	id := status.ID
	err := database.GormDB.Where("id = ? AND category = ?", m.ID, m.MonitorCategory).First(&alertStatus).Error
	if err != nil {
		// DB上に無い場合もあるためログしない
		// Logger.Writef(LOG_LEVEL_ERR, "SensorMonitorAlertStatus Read environment DB:%s", err.Error())
	} else {
		//Logger.Writef(LOG_LEVEL_DEBUG, "SensorMonitorAlertStatus Read DB:%+v", alertStatus)
	}

	if alertStatus.Category != m.MonitorCategory {
		alertStatus.ID = m.ID
		alertStatus.Category = m.MonitorCategory
		alertStatus.StartTime = 0
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "SensorMonitorAlertStatus write DB:%+v", alertStatus)
	Logger.Writef(LOG_LEVEL_DEBUG, "MonitorCategory %s", m.MonitorCategory)

	switch m.MonitorCategory {
	case strCurrent1:
		ret = sensorValueJudge(id, m, &alertStatus, status.Current1, overAllCommonTime, strEnergySensor)
	case strCurrent2:
		ret = sensorValueJudge(id, m, &alertStatus, status.Current2, overAllCommonTime, strEnergySensor)
	case strCurrent3:
		ret = sensorValueJudge(id, m, &alertStatus, status.Current3, overAllCommonTime, strEnergySensor)
	case strVoltage1:
		ret = sensorValueJudge(id, m, &alertStatus, status.Voltage1, overAllCommonTime, strEnergySensor)
	case strVoltage2:
		ret = sensorValueJudge(id, m, &alertStatus, status.Voltage2, overAllCommonTime, strEnergySensor)
	case strVoltage3:
		ret = sensorValueJudge(id, m, &alertStatus, status.Voltage3, overAllCommonTime, strEnergySensor)
	case strPowerFactor:
		ret = sensorValueJudge(id, m, &alertStatus, status.PowerFactor, overAllCommonTime, strEnergySensor)
	case strElectricPower:
		ret = sensorValueJudge(id, m, &alertStatus, status.EffectivePower, overAllCommonTime, strEnergySensor)
	}

	return ret
}

/*
センサ監視の環境判定(温度 / 湿度 / 不快指数)
*/
func EnvironmentJudge(m MonitorConditions, status EnvironmentSensorStatus, sensor EnvironmentSensor, deviceid string, overAllCommonTime int64) (ret bool) {
	//Logger.Writef(LOG_LEVEL_DEBUG, "EnvironmentJudge status:%+v", status)

	var alertStatus SensorMonitorAlertStatus
	id := status.SensorID
	err := database.GormDB.Where("id = ? AND category = ?", m.ID, strTemperature).First(&alertStatus).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "SensorMonitorAlertStatus Read environment DB:%s", err.Error())
	} else {
		Logger.Writef(LOG_LEVEL_DEBUG, "SensorMonitorAlertStatus Read environment DB:%+v", alertStatus)
	}

	if alertStatus.Category != m.MonitorCategory {
		alertStatus.ID = m.ID
		alertStatus.Category = m.MonitorCategory
		alertStatus.StartTime = 0
	}
	//Logger.Writef(LOG_LEVEL_DEBUG, "SensorMonitorAlertStatus first status:%+v", alertStatus)

	switch sensor.Category {
	case strTemperatureAndHumidity:
		// 温湿度センサの場合
		switch m.MonitorCategory {
		case strTemperature:
			// 温度
			ret = sensorValueJudge(id, m, &alertStatus, status.Temperature, overAllCommonTime, strEnvironmentalSensor)
		case strTemperatureAndHumidity:
			// 湿度
			ret = sensorValueJudge(id, m, &alertStatus, status.Humidity, overAllCommonTime, strEnvironmentalSensor)
		case strDiscomfortIndex:
			// 不快指数
			ret = sensorValueJudge(id, m, &alertStatus, status.DiscomfortIndex, overAllCommonTime, strEnvironmentalSensor)
		default:
			return false
		}

	case strTemperature:
		if m.MonitorCategory == strTemperature {
			// 温度センサの場合
			ret = sensorValueJudge(id, m, &alertStatus, status.Temperature, overAllCommonTime, strEnvironmentalSensor)
		} else {
			return false
		}

	default:
		return false
	}

	return ret
}

// 下回り判定
func underSensorValueJudge(value, threshold float64, alertStatus *SensorMonitorAlertStatus, JudgementAbnormalTime int64) bool {
	if value < threshold {
		// 発生判定
		if alertStatus.StartTime == 0 && JudgementAbnormalTime > 0 {
			alertStatus.StartTime = time.Now().Unix()
			database.Save(alertStatus)
		} else {
			leftTime := int64(time.Since(time.Unix(alertStatus.StartTime, 0)).Seconds())
			if JudgementAbnormalTime <= leftTime {
				// 発生
				alertStatus.LastStatus = strOccurrence
				if alertStatus.OccurredAt == 0 {
					alertStatus.OccurredAt = time.Now().Unix()
				}
				alertStatus.StartTime = 0
				database.Save(alertStatus)
				return true
			}
		}
	} else if value >= threshold && alertStatus.LastStatus == strOccurrence {
		// 復旧判定
		var alertCancelTime int64 = alertCancel * 60 * 60
		if alertStatus.StartTime == 0 {
			alertStatus.StartTime = time.Now().Unix()
			database.Save(alertStatus)
		} else {
			leftTime := int64(time.Since(time.Unix(alertStatus.StartTime, 0)).Seconds())
			if alertCancelTime <= leftTime {
				// 復旧
				alertStatus.LastStatus = strRestoration
				alertStatus.OccurredAt = 0
				alertStatus.StartTime = 0
				database.Save(alertStatus)
				// 復旧アラートは不要
				// return true
			}
		}
	} else {
		alertStatus.StartTime = 0
		database.Save(alertStatus)
	}

	return false
}

// 上回り判定
func overSensorValueJudge(value, threshold float64, alertStatus *SensorMonitorAlertStatus, JudgementAbnormalTime int64) bool {
	if value > threshold {
		// 発生判定
		if alertStatus.StartTime == 0 && JudgementAbnormalTime > 0 {
			alertStatus.StartTime = time.Now().Unix()
			database.Save(alertStatus)
		} else {
			leftTime := int64(time.Since(time.Unix(alertStatus.StartTime, 0)).Seconds())
			if JudgementAbnormalTime <= leftTime {
				// 発生
				alertStatus.LastStatus = strOccurrence
				if alertStatus.OccurredAt == 0 {
					alertStatus.OccurredAt = time.Now().Unix()
				}
				alertStatus.StartTime = 0
				database.Save(alertStatus)
				return true
			}
		}
	} else if value <= threshold && alertStatus.LastStatus == strOccurrence {
		// 復旧判定
		var alertCancelTime int64 = alertCancel * 60 * 60
		if alertStatus.StartTime == 0 {
			alertStatus.StartTime = time.Now().Unix()
			database.Save(alertStatus)
		} else {
			leftTime := int64(time.Since(time.Unix(alertStatus.StartTime, 0)).Seconds())
			if alertCancelTime <= leftTime {
				// 復旧
				alertStatus.LastStatus = strRestoration
				alertStatus.OccurredAt = 0
				alertStatus.StartTime = 0
				database.Save(alertStatus)
				// 復旧アラートは不要
				// return true
			}
		}
	} else {
		alertStatus.StartTime = 0
		database.Save(alertStatus)
	}

	return false
}

/*
センサ監視の値判定
環境 / 電力 /
*/
func sensorValueJudge(id string, m MonitorConditions, alertStatus *SensorMonitorAlertStatus, value float64, overAllCommonTime int64, kind string) bool {
	isAlert := false

	//Logger.Writef(LOG_LEVEL_DEBUG, "senorValueJudge:%s", kind)
	//上回るか下回るか判定
	switch m.JudgementMethod {
	case strExceed:
		isAlert = overSensorValueJudge(value, m.Threshold, alertStatus, m.JudgementAbnormalTime)
	case strOn:
		isAlert = overSensorValueJudge(value, m.Threshold, alertStatus, m.JudgementAbnormalTime)

	case strBelow:
		isAlert = underSensorValueJudge(value, m.Threshold, alertStatus, m.JudgementAbnormalTime)
	case strOff:
		isAlert = underSensorValueJudge(value, m.Threshold, alertStatus, m.JudgementAbnormalTime)

	default:
		Logger.Writef(LOG_LEVEL_ERR, "SensorMonitorAlert Save %s Judgement Method:%s", kind, m.JudgementMethod)
		return false
	}

	if isAlert {
		// センサーアラート(クラウド送信用)を保存
		alerts := sensorAlertCreate(id, alertStatus, value, overAllCommonTime, m.ID, kind)
		//Logger.Writef(LOG_LEVEL_DEBUG, "Sensor Alert Save : %+v", alerts)
		if err := database.Save(&alerts); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save sensor alerts: %s", err.Error())
		}

		// センサーアラート(メール送信用)を保存
		mailAlert := SensorMonitorMailAlert{
			Time:                overAllCommonTime,
			MonitorConditionsID: m.ID,
			OccurredAt:          alertStatus.OccurredAt,
			SensorID:            id,
			Status:              alertStatus.LastStatus,
			Value:               value,
			Kind:                kind,
			AlertMessage:        m.Message,
		}
		if err := database.Save(&mailAlert); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to save mail alert: %s", err.Error())
	}
		return true
	}

	return false
}

// 引数で指定した通信デバイス識別子の通信状態をローカルDB上に保存する
func UpdateCommunicationMailAlertStatus(communicationDeviceID, deviceType string, now int64, status int) {

	communicationMailAlertStatus := CommunicationMailAlertStatus{
		ID:         communicationDeviceID,
		DeviceType: deviceType,
		Time:       now,
		LastStatus: status,
		OccurredAt: now,
		SentMailAt: 0,
	}

	latestStatus := database.CommunicationMailAlertStatusReadDB(communicationDeviceID)
	if latestStatus != nil &&
		latestStatus.LastStatus == COM_NG &&
		status == COM_NG {
		communicationMailAlertStatus.OccurredAt = latestStatus.OccurredAt
		communicationMailAlertStatus.SentMailAt = latestStatus.SentMailAt
	}
	if err := database.Save(&communicationMailAlertStatus); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to save communication mail alert status: %s", err.Error())
	}
}

// 引数で指定した通信デバイス識別子の通信状態をローカルDB上に保存する
func UpdateBatteryLevelMailAlertStatus(demandPulseID string, now int64, status int) {

	batteryLevelMailAlertStatus := BatteryLevelMailAlertStatus{
		ID:         demandPulseID,
		Time:       now,
		LastStatus: status,
		OccurredAt: now,
		SentMailAt: 0,
	}

	latestStatus := database.BatteryLevelMailAlertStatusReadDB(demandPulseID)
	if latestStatus != nil &&
		latestStatus.LastStatus == VOLTAGE_NG &&
		status == VOLTAGE_NG {
		batteryLevelMailAlertStatus.OccurredAt = latestStatus.OccurredAt
		batteryLevelMailAlertStatus.SentMailAt = latestStatus.SentMailAt
	}
	if err := database.Save(&batteryLevelMailAlertStatus); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to save battery level mail alert status: %s", err.Error())
	}
}

// 通信異常に関するアラートメールを送信するべきか判断する
// 通信異常検知時刻から5分後、それ以降は24時間経過毎に送信する
func ShouldSendCommunicationMailAlert(now, occuredAt, sentMailAt int64) bool {
	const fiveMinutes = 60 * 5
	const oneDay = 60 * 60 * 24
	return ShouldSendMailByTime(now, occuredAt, sentMailAt, fiveMinutes, oneDay)
}

// 電圧以上に関するアラートメールを送信するべきか判断する
// 通信異常検知時刻から5分後、それ以降は24時間経過毎に送信する
func ShouldSendBatteryLevelMailAlert(now, occuredAt, sentMailAt, judgementAbnormalTime int64) bool {
	const oneDay = 60 * 60 * 24
	return ShouldSendMailByTime(now, occuredAt, sentMailAt, judgementAbnormalTime, oneDay)
}

// 現在時刻、異常発生時刻、最終メール送信時刻、初回メール送信間隔、2回目以降のメール送信間隔から
// アラートメールを送信するべきか判断する
// 初回の場合は firstInterval (秒) 経過後、2回目以降は afterInterval (秒) 経過後かどうかを判断する
// sentMailAt が 0の場合、メールは未送信と判断する
func ShouldSendMailByTime(now, occuredAt, sentMailAt, firstInterval, afterInterval int64) bool {
	// メールを1度も送信していない場合
	if sentMailAt == 0 {
		elapsedSeconds := now - occuredAt
		if elapsedSeconds >= firstInterval {
			return true
		}
	} else {
		// 初回メールは送信済の場合
		elapsedSecondsByLastMail := now - sentMailAt
		if elapsedSecondsByLastMail >= afterInterval {
			return true
		}
	}
	return false
}
