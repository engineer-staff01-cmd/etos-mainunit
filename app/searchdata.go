package app

import "fmt"

/*
	機器情報の取得
*/
func SearchDeviceInformation(id string, deviceInformation []DeviceInformation) (ret DeviceInformation, err error) {
	for _, v := range deviceInformation {
		if v.ID == id {
			ret = v
			return ret, nil
		}
	}
	err = fmt.Errorf("record not found")
	return ret, err
}

/*
	機器用の電力センサ情報の取得
*/
func SearchenergySensor(deviceInformation DeviceInformation, energySensores []EnergySensor) (ret EnergySensor, err error) {
	for _, v := range energySensores {
		if v.ID == deviceInformation.EnergySensorID {
			ret = v
			return ret, nil
		}
	}
	err = fmt.Errorf("record not found")
	return ret, err
}

/*
	機器用の環境センサ情報の取得
*/
func SearchEnvironmentSensor(id string, environmentSensores []EnvironmentSensor) (ret EnvironmentSensor, err error) {
	for _, v := range environmentSensores {
		if v.ID == id {
			ret = v
			return ret, nil
		}
	}
	err = fmt.Errorf("record not found")
	return ret, err
}

/*
	機器用の制御情報の取得
*/
func SearchControlConditions(deviceInformation DeviceInformation, controlConditionses []ControlConditions) (ret ControlConditions, err error) {
	for _, v := range controlConditionses {
		if v.TargetDevicesID == deviceInformation.ID {
			ret = v
			return ret, nil
		}
	}
	return ret, err
}

/*
	機器用の出力接点情報の取得
*/
func SearchOutputContact(deviceInformation DeviceInformation, outputContact []OutputContact) (ret []OutputContact, err error) {
	for _, v := range outputContact {

		// 間欠制御判定
		if v.ID == deviceInformation.IntermittenControlContactID {
			ret = append(ret, v)
		}

		// 容量制御判定
		if v.ID == deviceInformation.CapacityControlContactID {
			ret = append(ret, v)
		}

		// デフロスト制御判定
		if v.ID == deviceInformation.DefrostControlContactID {
			ret = append(ret, v)
		}
	}

	if len(ret) > 0 {
		return ret, nil
	}

	err = fmt.Errorf("record not found")
	return ret, err
}

/*
	機器用の出力接点情報の取得
*/
func SearchDefrostContact(deviceInformation DeviceInformation, outputContact []OutputContact) (ret OutputContact, err error) {
	for _, v := range outputContact {
		// デフロスト制御判定
		if v.ID == deviceInformation.DefrostControlContactID {
			ret = v
			return ret, nil
		}
	}
	err = fmt.Errorf("record not found")
	return ret, err
}

/*
	機器用の遠隔制御命令の取得
*/
func SearchRemoteControl(deviceInformation DeviceInformation, remotecontroles []RemoteControl) (ret *RemoteControl, err error) {
	for _, j := range remotecontroles {
		if j.TargetDeviceId == deviceInformation.ID {
			return &j, nil
		}
	}
	err = fmt.Errorf("record not found")
	return nil, err
}

/*
	機器用の環境センサステータス取得
*/
func SearchEnvironmentSensorStatus(id string, time int64, environmentSensorStatuses []EnvironmentSensorStatus) (ret EnvironmentSensorStatus, err error) {
	for _, v := range environmentSensorStatuses {
		if v.SensorID == id {
			ret = v
			return ret, nil
		}
	}
	err = fmt.Errorf("record not found")
	return ret, err
}

/*
	機器用のエネルギーセンサーステータス取得
*/
func SearchEnergyStatusElectricPower(id string, time int64, energyStatusElectricPowers []EnergyAlertStatusElectricPower) (ret EnergyAlertStatusElectricPower, err error) {
	for _, v := range energyStatusElectricPowers {
		if v.ID == id {
			ret = v
			return ret, nil
		}
	}
	err = fmt.Errorf("record not found")
	return ret, err
}

/*
	機器用の入力IOステータス取得
*/
func SearchInputContactStatus(id string, time int64, inputContactAlertStatus []InputContactAlertStatus) (ret InputContactAlertStatus, err error) {
	for _, v := range inputContactAlertStatus {
		if v.ID == id {
			ret = v
			return ret, nil
		}
	}
	err = fmt.Errorf("record not found")
	return ret, err
}

/*
	機器用の出力IOステータス取得
*/
func SearchOutputContactAlertStatus(id string, time int64, outputContactAlertStatus []OutputContactAlertStatus) (ret OutputContactAlertStatus, err error) {
	for _, v := range outputContactAlertStatus {
		if v.ID == id {
			ret = v
			return ret, nil
		}
	}
	return ret, fmt.Errorf("record not found")
}

/*
	機器用の制御状態取得
*/
func SearchDeviceStatus(deviceInformation DeviceInformation, deviceStatuses []DeviceStatus) (ret DeviceStatus) {
	for _, v := range deviceStatuses {
		if v.DeviceID == deviceInformation.ID {
			ret = v
			break
		}
	}

	/* 見つからなかった場合 */
	if ret.DeviceID != deviceInformation.ID {
		ret.DeviceID = deviceInformation.ID
	}

	return ret
}

/*
	定格消費電力の取得
*/
func SearchRatedPowerConsumption(deviceInfo DeviceInformation, energySensor EnergySensor) (ret float64) {
	if deviceInfo.Mode == strHeating {
		ret = energySensor.RatedPowerConsumptionWarm
	} else {
		ret = energySensor.RatedPowerConsumptionCool
	}
	return ret
}

func SearchControlStatus(deviceInformation DeviceInformation, controlStatuses []ControlStatus) (ret ControlStatus) {

	for _, v := range controlStatuses {
		if v.ID == deviceInformation.ID {
			ret = v
			break
		}
	}

	/* 見つからなかった場合 */
	if ret.ID != deviceInformation.ID {
		ret.ID = deviceInformation.ID
	}

	return ret
}

func SearchPresentValue(deviceInformation DeviceInformation, values []PresentValue) (ret PresentValue, err error) {
	for _, v := range values {
		if v.ID == deviceInformation.ID {
			ret = v
			return ret, nil
		}
	}
	err = fmt.Errorf("record not found")
	return ret, err
}

func SearchRemoteDefrostCommand(deviceInformation DeviceInformation, remoteDefrostCommands []RemoteDefrostCommand) (ret RemoteDefrostCommand, err error) {
	for _, v := range remoteDefrostCommands {
		if v.OutputContactID == deviceInformation.DefrostControlContactID {
			ret = v
			return ret, nil
		}
	}
	err = fmt.Errorf("record not found")
	return ret, err
}
