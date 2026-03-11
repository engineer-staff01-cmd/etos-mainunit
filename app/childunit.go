package app

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

// 子機通信

type ChildThread struct {
	to   chan ChannelMessage
	from chan ChannelMessage
}

type UnitStatus struct {
	id  int16
	err bool
}

var IoUnitFlag bool = false

func readChildUnit(chmsg ChannelMessage) bool {
	switch chmsg.messageType {
	case End:
		return true

	case DataGet:
		overAllCommonTime := chmsg.time.Unix()

		// IO入力 取得
		IOunitInputProcess(overAllCommonTime)
		// 電力モニタ 計測
		EnergyUnitProcess(overAllCommonTime)

		// 環境センサ 計測
		EnvironmentUnitProcess(overAllCommonTime)
	}

	return false
}

// initializeOutputContacts クラウドで使用している出力接点の未使用のチャンネルをOFFに初期化する
func initializeOutputContacts() {
	outputContacts := []OutputContact{}
	if err := database.SelectAll(&outputContacts); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select output contacts: %s", err.Error())
	}
	type Contact struct {
		deviceName string
		channels   []bool
	}

	// 使用しているControlIDとChannelを取得してまとめる
	controlIDList := map[int16]Contact{}
	for _, outputContact := range outputContacts {
		// 子機または出力接点が無効なら未使用と判定する
		if !IsEnableChildUnit(outputContact.ChildDeviceID) || !i82b(outputContact.Enable) {
			continue
		}

		io := NewIoUnit(outputContact.Name, byte(outputContact.ControlID))
		channelID := outputContact.ControlChannel
		if channelID <= 0 || io.Channels() < channelID {
			continue
		}
		contact, ok := controlIDList[outputContact.ControlID]
		if ok {
			contact.channels[channelID-1] = true
			continue
		}

		newContact := Contact{
			deviceName: outputContact.Name,
			channels:   make([]bool, io.Channels()),
		}
		newContact.channels[channelID-1] = true
		controlIDList[outputContact.ControlID] = newContact
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "OutputContact Control ID List:%v", controlIDList)

	// まとめたデータから使用していないチャネルに対してOFFを設定する
	for controlID, contact := range controlIDList {
		io := NewIoUnit(contact.deviceName, byte(controlID))
		if err := io.SetWatchdog(); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to set Watchdog:%02d Error:%s", controlID, err.Error())
		} else {
			Logger.Writef(LOG_LEVEL_DEBUG, "Set to Watchdog:%02d Time:%d sec", controlID, IOUnitWatchdogTime)
		}
		for i, isChannelUsed := range contact.channels {
			if !isChannelUsed {
				controlChannel := i + 1
				if err := io.SetOutputStatus(uint16(controlChannel), 0); err != nil {
					Logger.Writef(LOG_LEVEL_ERR, "Failed to set status to OutputContact Connection:%02d-%02d Error:%s", controlID, controlChannel, err.Error())
				} else {
					Logger.Writef(LOG_LEVEL_DEBUG, "Set status to OutputContact Connection:%02d-%02d Value:%d", controlID, controlChannel, 0)
				}
			}
		}
	}
}

func outputControlRequest(chOutput ChannelControl) {
	var outputContact []OutputContact
	result := database.GormDB.Find(&outputContact)
	if result.Error != nil {
		Logger.Writef(LOG_LEVEL_ERR, "OutputContact Cloud Read DB:%s", result.Error.Error())
		return
	}
	//	Logger.Writef(LOG_LEVEL_DEBUG, "OutputContact Control Request:%v", outputContact)

	var deviceInformation []DeviceInformation
	result = database.GormDB.Find(&deviceInformation)
	if result.Error != nil {
		Logger.Writef(LOG_LEVEL_ERR, "DeviceInformation Cloud Read DB:%s", result.Error.Error())
		return
	}
	//	Logger.Writef(LOG_LEVEL_DEBUG, "deviceInformation Control Request:%v", deviceInformation)

	// 出力
	overAllCommonTime := chOutput.time
	IOunitOutputProcess(outputContact, overAllCommonTime, chOutput, deviceInformation)
}

func NewChild() (c *ChildThread) {
	c = new(ChildThread)
	c.to = make(chan ChannelMessage, ChildUnitChannelBufferSize)   // 子機(main -> child)
	c.from = make(chan ChannelMessage, ChildUnitChannelBufferSize) // 子機(main <- child)
	return
}

func (c *ChildThread) childThreadRun(receivechOutput <-chan ChannelControl, sendchCloudMsg chan ChannelMessage) {

	var environmentSensores []EnvironmentSensor
	if err := database.SelectAll(&environmentSensores); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select environment sensors: %s", err.Error())
	}

	var tu TemperatureUnit
	var prevID int16

	for _, v := range environmentSensores {
		if v.Category == strTemperatureAndHumidity {
			// 環境センサ 湿度
			// TODO: 湿度が取れる環境センサーは別途実装
		} else {
			// 環境センサ センサータイプ設定（0x29：-200～200℃）
			if v.ControlID != prevID {
				prevID = v.ControlID
				tu = NewTemperatureUnit(v.DeviceName, byte(v.ControlID))
				tu.InitSensorType()
			}
		}
	}

	for {
		select {
		case chmsg := <-c.to:
			//			Logger.Writef(LOG_LEVEL_DEBUG, "ChildThread from main to child")
			endFlag := readChildUnit(chmsg)
			if endFlag {
				return
			}
			// cloudスレに送る
			msg := ChannelMessage{Alert, strchildIn, chmsg.time}
			SendChannelMessageSafely(sendchCloudMsg, msg, false)
			// childUnitスレ終了
			//			Logger.Writef(LOG_LEVEL_DEBUG, "ChildThread End")
			msg = ChannelMessage{End, "", chmsg.time}
			SendChannelMessageSafely(c.from, msg, false)

		case chOutput := <-receivechOutput:
			IoUnitFlag = true
			//Logger.Writef(LOG_LEVEL_DEBUG, "Output Contact Request Count:%d", cnt)
			outputControlRequest(chOutput)
			// cloudスレに送る
			msg := ChannelMessage{Alert, strchildOut, time.Unix(chOutput.time, 0)}
			SendChannelMessageSafely(sendchCloudMsg, msg, false)
			IoUnitFlag = false
		}
	}
}

/*
並列で常時起動、
一定周期でデータを取得
*/
func (c *ChildThread) Run(wg *sync.WaitGroup, receivechOutput <-chan ChannelControl, sendchCloudMsg chan ChannelMessage) {
	initializeOutputContacts()
	Logger.Writef(LOG_LEVEL_DEBUG, "Start ChildThread")
	wg.Add(1) // goroutine起動前にカウントを増やす
	go func() {
		c.childThreadRun(receivechOutput, sendchCloudMsg)
		Logger.Writef(LOG_LEVEL_DEBUG, "Stop ChildThread")
		wg.Done()
	}()
}

func IsEnableChildUnit(childDeviceId string) bool {
	childUnit := ChildUnit{}
	database.SelectByQuery(&childUnit, "id = ?", childDeviceId)
	if childUnit.Enable != 1 {
		return false
	}
	return true
}

// 入力処理
func IOunitInputProcess(overAllCommonTime int64) {
	// InputContact read DB
	var inputContacts []InputContact
	var unitStatus []UnitStatus
	if err := database.SelectAll(&inputContacts); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select input contacts: %s", err.Error())
	}

	for _, v := range inputContacts {
		if !IsEnableChildUnit(v.ChildDeviceID) {
			continue
		}

		// IOユニット 入力状態取得
		//controlID 1つにつき1回updateDataを呼ぶ
		io := NewIoUnit(v.DeviceName, byte(v.ControlID))
		if CheckNeedUpdate(&unitStatus, v.ControlID) {
			err := io.UpdateData()
			if err != nil {
				notifyIOUnitInputError(v.ID, v.Name, v.ChildDeviceID, overAllCommonTime)
				Logger.Writef(LOG_LEVEL_ERR, "Failed to get status from InputContact Connection:%02d-%02d Error:%s", v.ControlID, v.ControlChannel, err.Error())
				SetErrorToStatus(&unitStatus, v.ControlID)
				continue
			}
		} else if IsErrorStatus(&unitStatus, v.ControlID) {
			notifyIOUnitInputError(v.ID, v.Name, v.ChildDeviceID, overAllCommonTime)
			continue
		}

		// UpdateDataでエラー判定しているため、ここでエラー判定は不要
		val, _ := io.GetInputValue(uint16(v.ControlChannel))
		// DBへ値の保存
		inputContactStatus := InputContactStatus{Time: overAllCommonTime, SensorID: v.ID, Status: int8(val)}
		inputContactAlertStatus := InputContactAlertStatus{ID: v.ID, Status: int8(val), Updated: time.Now().Unix()}
		if err := database.Save(&inputContactStatus); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save input contact status: %s", err.Error())
		}
		if err := database.Save(&inputContactAlertStatus); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save input contact alert status: %s", err.Error())
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "Input ID:%d CH:%d value:%d", v.ControlID, v.ControlChannel, inputContactStatus.Status)
		SaveDBChildDeviceCommunicationAlert(v.ChildDeviceID, strInputContact, COM_OK)
		UpdateCommunicationMailAlertStatus(v.ID, strInputContact, overAllCommonTime, COM_OK)
		//		Logger.Writef(LOG_LEVEL_DEBUG,
		//			"Get status from InputContact Connection:%02d-%02d ID:%s Name:%s Category:%s Value:%d",
		//			v.ControlID,
		//			v.ControlChannel,
		//			v.ID,
		//			v.Name,
		//			v.Category,
		//			int8(val),
		//		)
	}
}

func notifyIOUnitInputError(id, sensorName, ChildDeviceID string, time int64) {
	SaveDBChildDeviceCommunicationAlert(ChildDeviceID, strInputContact, COM_NG)
	UpdateCommunicationMailAlertStatus(id, strInputContact, time, COM_NG)
	SaveDBCommunicationAlert(id, sensorName, time, strInputContact)
}

/*
電力センサ処理
*/
func EnergyUnitProcess(overAllCommonTime int64) {
	// 電力モニタ // DBからデータ取得したらidを取得できるように改造
	var energySensores []EnergySensor
	var unitStatus []UnitStatus
	if err := database.SelectAll(&energySensores); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select energy sensors: %s", err.Error())
	}

	for _, e := range energySensores {
		if !IsEnableChildUnit(e.ChildDeviceID) || e.Enable == 0 {
			continue
		}

		// 電力取得
		eu := NewEnergyUnit(e.DeviceName, byte(e.ControlID))

		//Logger.Writef(LOG_LEVEL_DEBUG, "Energy Unit ID [ %d ], Device Name: [ %s ]", e.ControlID, e.DeviceName)
		if strings.Contains(e.DeviceName, CT_225A) {
			eu.setCTtype(CT_Type_225A[:])
			// Logger.Writef(LOG_LEVEL_DEBUG, "Energy Unit ID [ %d ], CT Type Set 225A: [ %s ]", e.ControlID, e.DeviceName)
		} else if strings.Contains(e.DeviceName, CT_100A) {
			eu.setCTtype(CT_Type_100A[:])
			// Logger.Writef(LOG_LEVEL_DEBUG, "Energy Unit ID [ %d ], CT Type Set 100A: [ %s ]", e.ControlID, e.DeviceName)
			//		} else {
			//			eu.setCTtype(CT_Type_100A[:])
		}

		//controlID 1つにつき1回updateDataを呼ぶ
		if CheckNeedUpdate(&unitStatus, e.ControlID) == true {
			err := eu.UpdateData()
			if err != nil {
				notifyEnergyUnitError(e.ID, e.Name, e.ChildDeviceID, overAllCommonTime)
				SetErrorToStatus(&unitStatus, e.ControlID)
				Logger.Writef(LOG_LEVEL_ERR, "Failed to get status from EnergySensor Connection:%02d-%02d Error:%s", e.ControlID, e.ControlChannel, err.Error())
				continue
			}
		} else if IsErrorStatus(&unitStatus, e.ControlID) {
			notifyEnergyUnitError(e.ID, e.Name, e.ChildDeviceID, overAllCommonTime)
			continue
		}

		// UpdateDataでエラー判定しているため、ここでエラー判定は不要
		energyStatus := getEnergyValue(eu, e.ControlChannel)
		// DB保存
		energyStatus.EnergySensorID = e.ID
		energyStatus.Time = overAllCommonTime
		energyAlertStatusElectricPower := EnergyAlertStatusElectricPower{
			ID:                        energyStatus.EnergySensorID,
			EffectivePower:            energyStatus.EffectivePower,
			EffectivePowerConsumption: energyStatus.EffectivePowerConsumption,
			Current1:                  energyStatus.Current1,
			Current2:                  energyStatus.Current2,
			Current3:                  energyStatus.Current3,
			Voltage1:                  energyStatus.Voltage1,
			Voltage2:                  energyStatus.Voltage2,
			Voltage3:                  energyStatus.Voltage3,
			Frequency:                 energyStatus.Frequency,
			PowerFactor:               energyStatus.PowerFactor,
			Updated:                   overAllCommonTime,
		}
		database.EnergyStatusElectricPowerSaveDB(energyStatus)
		database.EnergyAlertStatusElectricPowerSaveDB(energyAlertStatusElectricPower)
		database.PresentValueEnergyCloudSaveDB(e, energyAlertStatusElectricPower)
		SaveDBChildDeviceCommunicationAlert(e.ChildDeviceID, strEnergySensor, COM_OK)
		UpdateCommunicationMailAlertStatus(e.ID, strEnergySensor, overAllCommonTime, COM_OK)
		Logger.Writef(LOG_LEVEL_DEBUG,
			"Get status from EnergySensor Channel:%02d-%02d ID:%s Name:%s EneryStatus:%+v",
			e.ControlID,
			e.ControlChannel,
			e.ID,
			e.DeviceName,
			energyStatus,
		)
	}
}

func notifyEnergyUnitError(id, sensorName, ChildDeviceID string, time int64) {
	SaveDBCommunicationAlert(id, sensorName, time, strEnergySensor)
	SaveDBChildDeviceCommunicationAlert(ChildDeviceID, strEnergySensor, COM_NG)
	UpdateCommunicationMailAlertStatus(id, strEnergySensor, time, COM_NG)
}

func CheckNeedUpdate(u *[]UnitStatus, e int16) bool {
	for _, a := range *u {
		if a.id == e {
			return false
		}
	}
	var unitStatus UnitStatus

	unitStatus.id = e
	unitStatus.err = false
	*u = append(*u, unitStatus)
	return true
}

func IsErrorStatus(u *[]UnitStatus, e int16) bool {
	for _, a := range *u {
		if a.id == e {
			return a.err
		}
	}
	return false
}

func SetErrorToStatus(u *[]UnitStatus, e int16) {
	for i := 0; i < len(*u); i++ {
		if (*u)[i].id == e {
			(*u)[i].err = true
			break
		}
	}
}

// 環境センサ処理
func EnvironmentUnitProcess(overAllCommonTime int64) {
	// 環境センサ // DBからデータ取得したらidを取得できるように改造
	var environmentSensores []EnvironmentSensor
	var unitStatus []UnitStatus
	if err := database.SelectAll(&environmentSensores); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select environment sensors: %s", err.Error())
	}

	//var tempANDConditions []TempANDCondition
	//database.SelectAll(&tempANDConditions)

	for _, v := range environmentSensores {
		if !IsEnableChildUnit(v.ChildDeviceID) || v.Enable == 0 {
			continue
		}

		// 環境センサ 温度・湿度取得
		var hu HumidityUnit
		var tu TemperatureUnit
		if v.Category == strTemperatureAndHumidity {
			// 環境センサ 湿度
			// TODO: 湿度が取れる環境センサーは別途実装
			hu = NewHumidityUnit(v.DeviceName, byte(v.ControlID))
			//EnableOnlyTemperatureJudge = false
		} else {
			// 環境センサ 温度
			tu = NewTemperatureUnit(v.DeviceName, byte(v.ControlID))
			//EnableOnlyTemperatureJudge = true
		}
		if CheckNeedUpdate(&unitStatus, v.ControlID) {
			if v.Category == strTemperatureAndHumidity {
				err := hu.UpdateData()
				if err != nil {
					notifyEnvironmentError(v.ID, v.Name, v.ChildDeviceID, overAllCommonTime)
					SetErrorToStatus(&unitStatus, v.ControlID)
					Logger.Writef(LOG_LEVEL_ERR, "Failed to get status from EnvironmentSensor(R1240) Connection:%02d-%02d Error:%s", v.ControlID, v.ControlChannel, err.Error())
					continue
				}
			} else {
				err := tu.UpdateData()
				if err != nil {
					notifyEnvironmentError(v.ID, v.Name, v.ChildDeviceID, overAllCommonTime)
					SetErrorToStatus(&unitStatus, v.ControlID)
					Logger.Writef(LOG_LEVEL_ERR, "Failed to get status from EnvironmentSensor(ADAM) Connection:%02d-%02d Error:%s", v.ControlID, v.ControlChannel, err.Error())
					continue
				}
			}
		} else if IsErrorStatus(&unitStatus, v.ControlID) {
			notifyEnvironmentError(v.ID, v.Name, v.ChildDeviceID, overAllCommonTime)
			continue
		}

		val := getEnvironmentValue(tu, hu, v)
		val.SensorID = v.ID
		val.Time = overAllCommonTime
		environmentSensorAlertStatus := EnvironmentSensorAlertStatus{
			ID:              val.SensorID,
			Temperature:     val.Temperature,
			Humidity:        val.Humidity,
			DiscomfortIndex: val.DiscomfortIndex,
			Updated:         overAllCommonTime,
		}

		// AND条件データベース登録
		if strings.Contains(v.DeviceName, "AND") {
			//if v.Category == strTemperatureAndHumidity {
			//if v.DeviceName[4:8] == "temp" || v.DeviceName[4:8] == "humi" {
			humiAndCnditionDBSave(v, val) // 温湿度センサー
			//} else {
			//	tempAndCnditionDBSave(v, val) // 温度センサー
			//}
		}

		if err := database.Save(&val); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save environment sensor status: %s", err.Error())
		}
		if err := database.Save(&environmentSensorAlertStatus); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save environment sensor alert status: %s", err.Error())
		}
		database.PresentValueEnvironmentCloudSaveDB(v, val)
		SaveDBChildDeviceCommunicationAlert(v.ChildDeviceID, strEnvironmentalSensor, COM_OK)
		UpdateCommunicationMailAlertStatus(v.ID, strEnvironmentalSensor, overAllCommonTime, COM_OK)
		Logger.Writef(LOG_LEVEL_DEBUG,
			"Get status from EnvironmentSensor Channel:%02d-%02d ID:%s Name:%s Category:%s Temperature:%.2f Humidity:%.2f DiscomfortIndex:%.2f",
			v.ControlID,
			v.ControlChannel,
			v.ID,
			v.Name,
			v.Category,
			val.Temperature,
			val.Humidity,
			val.DiscomfortIndex,
		)
	}
}

func humiAndCnditionDBSave(v EnvironmentSensor, val EnvironmentSensorStatus) {
	var deviceEnvironmentList DeviceEnvironmentInformation
	var controlCondition ControlConditions

	result := strings.ReplaceAll(v.DeviceName[2:], " ", "")
	arr1 := strings.Split(result, "_")
	arr2 := strings.Split(arr1[1][5:(len(arr1[1])-1)], ",")
	Logger.Writef(LOG_LEVEL_DEBUG, "AND condition Environment Sensor(Humi) : %v", arr2)
	//var tempANDCondition TempANDCondition
	var deviceID string
	var deviceName [8]string
	var humiID [8]string
	var humiVal [8]float64
	var humiCnt [8]string
	for i, value := range arr2 {
		arr3 := strings.Split(value, "-")
		//Logger.Writef(LOG_LEVEL_DEBUG, "AND condition Environment Sensor ID, Channel : %v", arr3)
		id, _ := strconv.Atoi(arr3[0])
		chn, _ := strconv.Atoi(arr3[1])
		if (int16(id) == v.ControlID) && (int16(chn) == v.ControlChannel) {
			// デバイスインフォメーションのidを取得
			err := database.GormDB.Where("sensor_id = ?", v.ID).Find(&deviceEnvironmentList).Error
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "deviceEnvironmentList Read DB:%s", err.Error())
				continue
			}
			deviceID = deviceEnvironmentList.DeviceID
			deviceName[i] = v.DeviceName
			humiID[i] = v.ID
			if strings.Contains(v.DeviceName, "humi") {
				humiVal[i] = val.Humidity
			} else if strings.Contains(v.DeviceName, "temp") {
				humiVal[i] = val.Temperature
			}

			// 制御条件の機器idを取得
			if len(arr1) > 2 {
				condname := arr1[2]
				err = database.GormDB.Where("name = ?", condname).Find(&controlCondition).Error
				if err != nil {
					Logger.Writef(LOG_LEVEL_ERR, "ControlConditions Read DB:%s", err.Error())
					continue
				}
				humiCnt[i] = controlCondition.TargetDevicesID
				// Logger.Writef(LOG_LEVEL_DEBUG, "AND condition Control Name : %s, ID : %s", condname, humiCnt[i])
			}
		}
	}
	var humiANDCondition HumiANDCondition
	//err := database.GormDB.Where("device_name = ?", v.DeviceName).First(&tempANDCondition).Error
	err := database.GormDB.Where("id = ?", deviceID).First(&humiANDCondition).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "humiANDConditions DB:%s", err.Error())
		//} else {
		//	Logger.Writef(LOG_LEVEL_DEBUG, "AND condition 1st DB select : %v", tempANDCondition)
	}
	humiANDCondition.ID = deviceID
	humiANDCondition.Number = len(arr2)
	if deviceName[0] != "" {
		humiANDCondition.DeviceName1 = deviceName[0]
	}
	if deviceName[1] != "" {
		humiANDCondition.DeviceName2 = deviceName[1]
	}
	if deviceName[2] != "" {
		humiANDCondition.DeviceName3 = deviceName[2]
	}
	if deviceName[3] != "" {
		humiANDCondition.DeviceName4 = deviceName[3]
	}
	if deviceName[4] != "" {
		humiANDCondition.DeviceName5 = deviceName[4]
	}
	if deviceName[5] != "" {
		humiANDCondition.DeviceName6 = deviceName[5]
	}
	if deviceName[6] != "" {
		humiANDCondition.DeviceName7 = deviceName[6]
	}
	if deviceName[7] != "" {
		humiANDCondition.DeviceName8 = deviceName[7]
	}
	if humiID[0] != "" {
		humiANDCondition.Humi1_ID = humiID[0]
	}
	if humiID[1] != "" {
		humiANDCondition.Humi2_ID = humiID[1]
	}
	if humiID[2] != "" {
		humiANDCondition.Humi3_ID = humiID[2]
	}
	if humiID[3] != "" {
		humiANDCondition.Humi4_ID = humiID[3]
	}
	if humiID[4] != "" {
		humiANDCondition.Humi5_ID = humiID[4]
	}
	if humiID[5] != "" {
		humiANDCondition.Humi6_ID = humiID[5]
	}
	if humiID[6] != "" {
		humiANDCondition.Humi7_ID = humiID[6]
	}
	if humiID[7] != "" {
		humiANDCondition.Humi8_ID = humiID[7]
	}
	if humiVal[0] != 0.0 {
		humiANDCondition.Humi1_Val = humiVal[0]
	}
	if humiVal[1] != 0.0 {
		humiANDCondition.Humi2_Val = humiVal[1]
	}
	if humiVal[2] != 0.0 {
		humiANDCondition.Humi3_Val = humiVal[2]
	}
	if humiVal[3] != 0.0 {
		humiANDCondition.Humi4_Val = humiVal[3]
	}
	if humiVal[4] != 0.0 {
		humiANDCondition.Humi5_Val = humiVal[4]
	}
	if humiVal[5] != 0.0 {
		humiANDCondition.Humi6_Val = humiVal[5]
	}
	if humiVal[6] != 0.0 {
		humiANDCondition.Humi7_Val = humiVal[6]
	}
	if humiVal[7] != 0.0 {
		humiANDCondition.Humi8_Val = humiVal[7]
	}
	if humiCnt[0] != "" {
		humiANDCondition.Humi1_Cnt = humiCnt[0]
	}
	if humiCnt[1] != "" {
		humiANDCondition.Humi2_Cnt = humiCnt[1]
	}
	if humiCnt[2] != "" {
		humiANDCondition.Humi3_Cnt = humiCnt[2]
	}
	if humiCnt[3] != "" {
		humiANDCondition.Humi4_Cnt = humiCnt[3]
	}
	if humiCnt[4] != "" {
		humiANDCondition.Humi5_Cnt = humiCnt[4]
	}
	if humiCnt[5] != "" {
		humiANDCondition.Humi6_Cnt = humiCnt[5]
	}
	if humiCnt[6] != "" {
		humiANDCondition.Humi7_Cnt = humiCnt[6]
	}
	if humiCnt[7] != "" {
		humiANDCondition.Humi8_Cnt = humiCnt[7]
	}
	if err := database.Save(&humiANDCondition); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to save humi AND condition: %s", err.Error())
	}
	//Logger.Writef(LOG_LEVEL_DEBUG, "AND condition data added(Humi) : %v", humiANDCondition)
}

/*
func tempAndCnditionDBSave(v EnvironmentSensor, val EnvironmentSensorStatus) {
	var deviceEnvironmentList DeviceEnvironmentInformation
	var controlCondition ControlConditions

	result := strings.ReplaceAll(v.DeviceName, " ", "")
	arr1 := strings.Split(result, "_")
	arr2 := strings.Split(arr1[1][5:(len(arr1[1])-1)], ",")
	Logger.Writef(LOG_LEVEL_DEBUG, "AND condition Environment Sensor(Temp) : %v", arr2)
	//var tempANDCondition TempANDCondition
	var deviceID string
	var deviceName [6]string
	var tempID [6]string
	var tempVal [6]float64
	var tempCnt [6]string
	for i, value := range arr2 {
		arr3 := strings.Split(value, "-")
		//Logger.Writef(LOG_LEVEL_DEBUG, "AND condition Environment Sensor ID, Channel : %v", arr3)
		id, _ := strconv.Atoi(arr3[0])
		chn, _ := strconv.Atoi(arr3[1])
		if (int16(id) == v.ControlID) && (int16(chn) == v.ControlChannel) {
			// デバイスインフォメーションのidを取得
			err := database.GormDB.Where("sensor_id = ?", v.ID).Find(&deviceEnvironmentList).Error
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "deviceEnvironmentList Read DB:%s", err.Error())
				continue
			}
			deviceID = deviceEnvironmentList.DeviceID
			deviceName[i] = v.DeviceName
			tempID[i] = v.ID
			tempVal[i] = val.Temperature

			// 制御条件の機器idを取得
			if len(arr1) > 2 {
				condname := arr1[2]
				err = database.GormDB.Where("name = ?", condname).Find(&controlCondition).Error
				if err != nil {
					Logger.Writef(LOG_LEVEL_ERR, "ControlConditions Read DB:%s", err.Error())
					continue
				}
				tempCnt[i] = controlCondition.TargetDevicesID
				Logger.Writef(LOG_LEVEL_DEBUG, "AND condition Control Name : %s, ID : %s", condname, tempCnt[i])
			}
		}
	}
	var tempANDCondition TempANDCondition
	//err := database.GormDB.Where("device_name = ?", v.DeviceName).First(&tempANDCondition).Error
	err := database.GormDB.Where("id = ?", deviceID).First(&tempANDCondition).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "tempANDConditions DB:%s", err.Error())
		//} else {
		//	Logger.Writef(LOG_LEVEL_DEBUG, "AND condition 1st DB select : %v", tempANDCondition)
	}
	tempANDCondition.ID = deviceID
	tempANDCondition.Number = len(arr2)
	if deviceName[0] != "" {
		tempANDCondition.DeviceName1 = deviceName[0]
	}
	if deviceName[1] != "" {
		tempANDCondition.DeviceName2 = deviceName[1]
	}
	if deviceName[2] != "" {
		tempANDCondition.DeviceName3 = deviceName[2]
	}
	if deviceName[3] != "" {
		tempANDCondition.DeviceName4 = deviceName[3]
	}
	if deviceName[4] != "" {
		tempANDCondition.DeviceName5 = deviceName[4]
	}
	if deviceName[5] != "" {
		tempANDCondition.DeviceName6 = deviceName[5]
	}
	if tempID[0] != "" {
		tempANDCondition.Temp1_ID = tempID[0]
	}
	if tempID[1] != "" {
		tempANDCondition.Temp2_ID = tempID[1]
	}
	if tempID[2] != "" {
		tempANDCondition.Temp3_ID = tempID[2]
	}
	if tempID[3] != "" {
		tempANDCondition.Temp4_ID = tempID[3]
	}
	if tempID[4] != "" {
		tempANDCondition.Temp5_ID = tempID[4]
	}
	if tempID[5] != "" {
		tempANDCondition.Temp6_ID = tempID[5]
	}
	if tempVal[0] != 0.0 {
		tempANDCondition.Temp1_Val = tempVal[0]
	}
	if tempVal[1] != 0.0 {
		tempANDCondition.Temp2_Val = tempVal[1]
	}
	if tempVal[2] != 0.0 {
		tempANDCondition.Temp3_Val = tempVal[2]
	}
	if tempVal[3] != 0.0 {
		tempANDCondition.Temp4_Val = tempVal[3]
	}
	if tempVal[4] != 0.0 {
		tempANDCondition.Temp5_Val = tempVal[4]
	}
	if tempVal[5] != 0.0 {
		tempANDCondition.Temp6_Val = tempVal[5]
	}
	if tempCnt[0] != "" {
		tempANDCondition.Temp1_Cnt = tempCnt[0]
	}
	if tempCnt[1] != "" {
		tempANDCondition.Temp2_Cnt = tempCnt[1]
	}
	if tempCnt[2] != "" {
		tempANDCondition.Temp3_Cnt = tempCnt[2]
	}
	if tempCnt[3] != "" {
		tempANDCondition.Temp4_Cnt = tempCnt[3]
	}
	if tempCnt[4] != "" {
		tempANDCondition.Temp5_Cnt = tempCnt[4]
	}
	if tempCnt[5] != "" {
		tempANDCondition.Temp6_Cnt = tempCnt[5]
	}
	if err := database.Save(&tempANDCondition); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to save temp AND condition: %s", err.Error())
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "AND condition data added(Temp) : %v", tempANDCondition)
}
*/

func notifyEnvironmentError(id, sensorName, ChildDeviceID string, time int64) {
	SaveDBChildDeviceCommunicationAlert(ChildDeviceID, strEnvironmentalSensor, COM_NG)
	UpdateCommunicationMailAlertStatus(id, strEnvironmentalSensor, time, COM_NG)
	SaveDBCommunicationAlert(id, sensorName, time, strEnvironmentalSensor)
}

/*
出力センサ処理
*/
func IOunitOutputProcess(outputContacts []OutputContact, overAllCommonTime int64, chOutput ChannelControl, deviceInformation []DeviceInformation) {
	// 機器情報の取得
	deviceinfo, _ := SearchDeviceInformation(chOutput.ID, deviceInformation)
	// 出力接点の情報取得
	outputContact, _ := SearchOutputContact(deviceinfo, outputContacts)

	// 温度センサーAND条件データ読み込み
	//var tempANDConditions []TempANDCondition
	//database.SelectAll(&tempANDConditions)

	// 出力接点の出力可否確認
	for _, data := range outputContact {
		// 出力フラグ
		var outValue int8

		if IsEnableChildUnit(data.ChildDeviceID) {
			// 間欠制御
			switch data.ID {
			case deviceinfo.IntermittenControlContactID:
				if data.ControlMethod == strStop {
					if data.Enable == 1 {
						outValue = chOutput.DeviceStop
					}
				} else {
					/* TODO: System Alert */
				}
			// 容量制御
			case deviceinfo.CapacityControlContactID:
				if data.ControlMethod == strCapacity {
					if data.Enable == 1 {
						outValue = chOutput.DeviceStop
					}
				} else {
					/* TODO: System Alert */
				}
			// デフロスト制御
			case deviceinfo.DefrostControlContactID:
				if data.Enable == 1 && chOutput.ControlType == StrControlRemoteDefrost {
					outValue = 1
				} else {
					/* TODO: System Alert */
				}
			default:
				/* TODO: System Alert */
			}
		} else {
			outValue = 0
		}
		/*
			var send_flag bool = false
			for _, val := range tempANDConditions {
				arr1 := strings.Split(val.Out_ID, "-")
				id, _ := strconv.Atoi(arr1[0])
				ch, _ := strconv.Atoi(arr1[1])
				if (id == int(data.ControlID)) && (ch == int(data.ControlChannel)) {
					send_flag = true
					if val.Temp1 && val.Temp2 && val.Temp3 && val.Temp4 && val.Temp5 {
						Output_CommunicationConfirmation(data, overAllCommonTime, outValue)
						Logger.Writef(LOG_LEVEL_DEBUG, "AND condition Set status to OutputContact Request:%02d-%02d Value:%d DataId:%s IntermittenControlContactID:%s ControlMethod:%s", data.ControlID, data.ControlChannel, outValue, data.ID, deviceinfo.IntermittenControlContactID, data.ControlMethod)
					} else {
						Output_CommunicationConfirmation(data, overAllCommonTime, 0)
					}
				}
			}

			if send_flag == false {
				Output_CommunicationConfirmation(data, overAllCommonTime, outValue)
			}	/*  */

		// 出力処理
		//		Logger.Writef(LOG_LEVEL_DEBUG, "Set status to OutputContact Request:%02d-%02d Value:%d DataId:%s IntermittenControlContactID:%s ControlMethod:%s", data.ControlID, data.ControlChannel, outValue, data.ID, deviceinfo.IntermittenControlContactID, data.ControlMethod)
		Output_CommunicationConfirmation(data, overAllCommonTime, outValue)
	}
}

/*
出力処理
*/
func Output_CommunicationConfirmation(outputContacts OutputContact, overAllCommonTime int64, outvalue int8) {
	// 子機無効の場合は出力のみ制御してアラートは発報しない
	if !IsEnableChildUnit(outputContacts.ChildDeviceID) {
		SaveDBChildDeviceCommunicationAlert(outputContacts.ChildDeviceID, strOutputContact, COM_OK)
		return
	}
	// IOユニット 出力コマンド送信
	err := IOunit_output(outputContacts.DeviceName, outputContacts.ControlID, outputContacts.ControlChannel, outvalue)
	if err != nil {
		// 通信異常アラート
		SaveDBChildDeviceCommunicationAlert(outputContacts.ChildDeviceID, strOutputContact, COM_NG)
		UpdateCommunicationMailAlertStatus(outputContacts.ID, strOutputContact, overAllCommonTime, COM_NG)
		SaveDBCommunicationAlert(outputContacts.ID, outputContacts.Name, overAllCommonTime, strOutputContact)
		Logger.Writef(LOG_LEVEL_ERR, "Failed to set status to OutputContact Connection(alert):%02d-%02d Error:%s", outputContacts.ControlID, outputContacts.ControlChannel, err.Error())
	} else {
		// ステータスをクラウドに保存（クラウド通知用と監視アラート用で２つ保存する）
		outputContactStatus := OutputContactStatus{overAllCommonTime, outputContacts.ID, outvalue}
		outputContactAlertStatus := OutputContactAlertStatus{outputContacts.ID, outvalue, time.Now().Unix()}
		if err := database.Save(&outputContactStatus); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save output contact status: %s", err.Error())
		}
		if err := database.Save(&outputContactAlertStatus); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save output contact alert status: %s", err.Error())
		}
		Logger.Writef(LOG_LEVEL_DEBUG, "Set status to OutputContact Connection(alert):%02d-%02d Value:%d", outputContacts.ControlID, outputContacts.ControlChannel, outvalue)
		SaveDBChildDeviceCommunicationAlert(outputContacts.ChildDeviceID, strOutputContact, COM_OK)
		UpdateCommunicationMailAlertStatus(outputContacts.ID, strOutputContact, overAllCommonTime, COM_OK)
	}
}

/*
IOユニット入力呼び出し
*/
func IOunit_input(deviceName string, id byte, channel int16) (byte, error) {
	io := NewIoUnit(deviceName, id) // 最初の引数がユニットの指定
	val, err := io.GetInputValue(uint16(channel))
	return val, err
}

/*
IOユニット出力呼び出し
*/
func IOunit_output(deviceName string, ControlID int16, ControlChannel int16, outvalue int8) error {
	io := NewIoUnit(deviceName, byte(ControlID))
	err := io.SetOutputStatus(uint16(ControlChannel), uint16(outvalue))
	return err
}

/*
電力モニタ通信呼び出し
*/
func getEnergyValue(eu EnergyUnit, ControlChannel int16) EnergyStatusElectricPower {
	energyGetValue := EnergyStatusElectricPower{}
	// 電圧取得
	voltage, err := eu.GetVoltage1(uint16(ControlChannel))
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "GetVoltage1 ID:%d:%s", eu.GetId(), err.Error())
	}
	energyGetValue.Voltage1 = voltage
	voltage, err = eu.GetVoltage2(uint16(ControlChannel))
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "GetVoltage2 ID:%d:%s", eu.GetId(), err.Error())
	}
	energyGetValue.Voltage2 = voltage
	voltage, err = eu.GetVoltage3(uint16(ControlChannel))
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "GetVoltage3 ID:%d:%s", eu.GetId(), err.Error())
	}
	energyGetValue.Voltage3 = voltage

	// 電流取得
	current, err := eu.GetCurrent1(uint16(ControlChannel))
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "GetCurrent1 ID:%d:%s", eu.GetId(), err.Error())
	}
	energyGetValue.Current1 = current
	current, err = eu.GetCurrent2(uint16(ControlChannel))
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "GetCurrent2 ID:%d:%s", eu.GetId(), err.Error())
	}
	energyGetValue.Current2 = current
	current, err = eu.GetCurrent3(uint16(ControlChannel))
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "GetCurrent3 ID:%d:%s", eu.GetId(), err.Error())
	}
	energyGetValue.Current3 = current

	// 電力取得
	ep, err := eu.GetEffectivePower(uint16(ControlChannel))
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "GetEffectivePower ID:%d:%s", eu.GetId(), err.Error())
	}
	energyGetValue.EffectivePower = ep

	// 周波数取得（CH1のみ取得可能、全CH共通）
	freq, err := eu.GetFrequency(1)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "GetFrequency ID:%d:%s", eu.GetId(), err.Error())
	}
	energyGetValue.Frequency = freq

	// 力率取得
	pf, err := eu.GetPowerFactor(uint16(ControlChannel))
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "GetPowerFactor ID:%d:%s", eu.GetId(), err.Error())
	}
	energyGetValue.PowerFactor = pf

	// 有効電力量
	epc, err := eu.GetEffectiveIntegratedElectricEnergy(uint16(ControlChannel))
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "GetEffectiveIntegratedElectricEnergy ID:%d:%s", eu.GetId(), err.Error())
	}
	energyGetValue.EffectivePowerConsumption = float64(epc)

	return energyGetValue
}

/*
温度モニタ通信呼び出し
*/
func getEnvironmentValue(tu TemperatureUnit, hu HumidityUnit, es EnvironmentSensor) EnvironmentSensorStatus {
	environmentSensorStatus := EnvironmentSensorStatus{}

	if es.Category == strTemperatureAndHumidity {
		// TODO 湿度取得
		//Logger.Writef(LOG_LEVEL_DEBUG, "温湿度センサー：デバイス名 : %s", es.DeviceName[4:8])
		if strings.Contains(es.DeviceName, "humi") {
			humi, err := hu.GetHumidity(uint16(es.ControlChannel))
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "Get Humidity(humi) ID:%d:%s", es.ControlID, err.Error())
			}
			environmentSensorStatus.Humidity = humi
		} else if strings.Contains(es.DeviceName, "temp") {
			temp, err := hu.GetTemperature(uint16(es.ControlChannel), es.CorrectionValue, es.CorrectionRatio)
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "Get Humidity(temp) ID:%d:%s", es.ControlID, err.Error())
			}
			environmentSensorStatus.Temperature = temp
		}
		// TODO 不快指数計算
		//environmentSensorStatus.DiscomfortIndex = CalcCurrentDiscomfortIndex(environmentSensorStatus.Temperature, environmentSensorStatus.Humidity)
	} else {
		// 温度取得
		temp, err := tu.GetTemperature(uint16(es.ControlChannel), es.CorrectionValue, es.CorrectionRatio)
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Get Temperature ID:%d:%s", es.ControlID, err.Error())
		}
		environmentSensorStatus.Temperature = temp
	}

	return environmentSensorStatus
}
