package app_test

import (
	"etos-mainunit/app"
	"testing"
)

func TestNewChild(t *testing.T) {
	c := app.NewChild()

	if c == nil {
		t.Errorf("failed to NewChild")
	}
}

// func TestRunChild(t *testing.T) {
// 	var wg sync.WaitGroup

// 	chSendCloud := make(chan ChannelMessage, 10)      // クラウドchanel
// 	chReceiveControl := make(chan ChannelControl, 10) // DeviceControlChanel
// 	c := NewChild()

// 	go func() {
// 		c.Run(&wg, chReceiveControl, chSendCloud)
// 	}()

// 	c.to <- ChannelMessage{End, "true", time.Now()}
// 	wg.Wait()
// }

// func TestChildThreadRun(t *testing.T) {
// 	db, err := app.NewDatabase() // DBの初期化
// 	if err != nil {
// 		Logger.Writef(LOG_LEVEL_ERR, "NewDatabase() %s", err.Error())
// 		t.Errorf("Databaseエラー")
// 		return
// 	}

// 	currentTime := time.Now()
// 	chSendCloud := make(chan ChannelMessage, 10)      // クラウドchanel
// 	chReceiveControl := make(chan ChannelControl, 10) // DeviceControlchanel
// 	c := NewChild()

// 	go func() {
// 		c.childThreadRun(chReceiveControl, chSendCloud)
// 	}()

// 	/* Device Control Message Test */
// 	chReceiveControl <- ChannelControl{StrControlAuto, "Dummy", 0, currentTime.Unix()}
// 	controlMsg := <-chSendCloud
// 	if controlMsg.messageType != Alert {
// 		t.Errorf("messageTypeエラー")
// 	}
// 	if controlMsg.message != strchildOut {
// 		t.Errorf("messageエラー")
// 	}
// 	temp := currentTime.Unix()
// 	resultTime := time.Unix(temp, 0)
// 	if controlMsg.time != resultTime {
// 		t.Errorf("timeエラー")
// 	}

// 	/* Channel Message Test */
// 	c.to <- ChannelMessage{End, "true", currentTime}
// 	db.Close()
// }

// func TestreadChildUnit(t *testing.T) {
// 	t.Skip()
// }

// func TestOutputControlRequest(t *testing.T) {
// 	t.Skip()
// }

// type TestIoUnit struct{}

// var TestInputValue [8]byte
// var TestOutputStatus [8]byte

// func (r *TestIoUnit) UpdateData() error {
// 	return nil
// }

// func (r *TestIoUnit) GetInputValue(channel uint16) (byte, error) {
// 	return TestInputValue[channel-1], nil
// }

// func (r *TestIoUnit) GetOutputStatus(channel uint16) (byte, error) {
// 	return TestOutputStatus[channel-1], nil
// }

// func (r *TestIoUnit) SetOutputStatus(channel uint16, value uint16) error {
// 	TestOutputStatus[channel-1] = byte(value)
// 	return nil
// }
// func (r *TestIoUnit) Channels() int16 {
// 	return 0
// }

// func TestIOunitOutputProcess(t *testing.T) {
// 	var outputContacts []OutputContact
// 	var deviceInformations []DeviceInformation

// 	db, err := app.NewDatabase() // DBの初期化
// 	if err != nil {
// 		app.Logger.Writef(app.LOG_LEVEL_ERR, "NewDatabase() %s", err.Error())
// 		t.Errorf("Databaseエラー")
// 		return
// 	}

// 	testIoUnit := new(TestIoUnit)
// 	SetIoUnitTestInterface(testIoUnit)

// 	db.GormDB.Delete(&ChildUnit{})
// 	db.GormDB.Save(ChildUnit{
// 		ID:     "ChildUnit1",
// 		Enable: 1,
// 	})

// 	deviceInformations = append(deviceInformations, DeviceInformation{
// 		ID:                          "DeviceInformation1",
// 		Mode:                        "",
// 		StopElectricPower:           0,
// 		RequiredControllingTime:     0,
// 		EnergySensorID:              "",
// 		EnvironmentJudgmentCriteria: "",
// 		DefrostInputID:              "",
// 		IntermittenControlContactID: "OutputContact1",
// 		CapacityControlContactID:    "OutputContact2",
// 		DefrostControlContactID:     "OutputContact3",
// 	})

// 	/* 出力接点 Enable Test */
// 	outputContacts = append(outputContacts, OutputContact{
// 		ID:             "OutputContact1",
// 		DeviceName:     "UnitTest1",
// 		Category:       "",
// 		ControlMethod:  strStop,
// 		ControlID:      0,
// 		ControlChannel: 1,
// 		Enable:         1,
// 	})
// 	outputContacts = append(outputContacts, OutputContact{
// 		ID:             "OutputContact2",
// 		DeviceName:     "UnitTest2",
// 		Category:       "",
// 		ControlMethod:  strCapacity,
// 		ControlID:      0,
// 		ControlChannel: 2,
// 		Enable:         1,
// 	})
// 	outputContacts = append(outputContacts, OutputContact{
// 		ID:             "OutputContact3",
// 		DeviceName:     "UnitTest3",
// 		Category:       "Defrost",
// 		ControlMethod:  strCapacity,
// 		ControlID:      0,
// 		ControlChannel: 3,
// 		Enable:         1,
// 	})
// 	chOutput := ChannelControl{
// 		ControlType: StrControlAuto,
// 		ID:          "DeviceInformation1",
// 		DeviceStop:  1,
// 		time:        0,
// 	}
// 	currentTIme := time.Now().Unix()
// 	IOunitOutputProcess(outputContacts, currentTIme, chOutput, deviceInformations)
// 	if TestOutputStatus[0] != 1 {
// 		t.Errorf("制御(停止)エラー")
// 	}
// 	if TestOutputStatus[1] != 1 {
// 		t.Errorf("制御(容量)エラー")
// 	}
// 	if TestOutputStatus[2] != 0 {
// 		t.Errorf("制御(デフロスト)エラー")
// 	}

// 	/* 遠隔デフロスト制御テスト */
// 	chOutput = ChannelControl{
// 		ControlType: StrControlRemoteDefrost,
// 		ID:          "DeviceInformation1",
// 		DeviceStop:  0,
// 		time:        0,
// 	}
// 	IOunitOutputProcess(outputContacts, currentTIme, chOutput, deviceInformations)
// 	if TestOutputStatus[0] != 0 {
// 		t.Errorf("制御(停止)エラー")
// 	}
// 	if TestOutputStatus[1] != 0 {
// 		t.Errorf("制御(容量)エラー")
// 	}
// 	if TestOutputStatus[2] != 1 {
// 		t.Errorf("制御(デフロスト)エラー")
// 	}

// 	/* 子機無効テスト */
// 	chOutput = ChannelControl{
// 		ControlType: StrControlAuto,
// 		ID:          "DeviceInformation1",
// 		DeviceStop:  1,
// 		time:        0,
// 	}
// 	db.GormDB.Delete(&ChildUnit{})
// 	db.GormDB.Save(ChildUnit{
// 		ID:     "ChildUnit1",
// 		Enable: 0,
// 	})
// 	IOunitOutputProcess(outputContacts, currentTIme, chOutput, deviceInformations)
// 	if TestOutputStatus[0] != 0 {
// 		t.Errorf("制御(停止)エラー")
// 	}
// 	if TestOutputStatus[1] != 0 {
// 		t.Errorf("制御(容量)エラー")
// 	}
// 	if TestOutputStatus[2] != 0 {
// 		t.Errorf("制御(デフロスト)エラー")
// 	}

// 	/* 出力接点 Disable Test */
// 	db.GormDB.Delete(&ChildUnit{})
// 	db.GormDB.Save(ChildUnit{
// 		ID:     "ChildUnit1",
// 		Enable: 1,
// 	})
// 	outputContacts = outputContacts[:0]
// 	outputContacts = append(outputContacts, OutputContact{
// 		ID:             "OutputContact1",
// 		DeviceName:     "UnitTest1",
// 		Category:       "",
// 		ControlMethod:  strStop,
// 		ControlID:      0,
// 		ControlChannel: 1,
// 		Enable:         0,
// 	})
// 	outputContacts = append(outputContacts, OutputContact{
// 		ID:             "OutputContact2",
// 		DeviceName:     "UnitTest2",
// 		Category:       "",
// 		ControlMethod:  strCapacity,
// 		ControlID:      0,
// 		ControlChannel: 2,
// 		Enable:         0,
// 	})
// 	outputContacts = append(outputContacts, OutputContact{
// 		ID:             "OutputContact3",
// 		DeviceName:     "UnitTest3",
// 		Category:       "Defrost",
// 		ControlMethod:  strCapacity,
// 		ControlID:      0,
// 		ControlChannel: 3,
// 		Enable:         1,
// 	})
// 	IOunitOutputProcess(outputContacts, currentTIme, chOutput, deviceInformations)
// 	if TestOutputStatus[0] != 0 {
// 		t.Errorf("制御(停止)エラー")
// 	}
// 	if TestOutputStatus[1] != 0 {
// 		t.Errorf("制御(容量)エラー")
// 	}
// 	if TestOutputStatus[2] != 0 {
// 		t.Errorf("制御(デフロスト)エラー")
// 	}

// 	db.Close()
// }

// func TestIsEnableChildUnit(t *testing.T) {
// 	db, err := app.NewDatabase() // DBの初期化
// 	if err != nil {
// 		app.Logger.Writef(app.LOG_LEVEL_ERR, "NewDatabase() %s", err.Error())
// 		t.Errorf("Databaseエラー")
// 		return
// 	}
// 	/* 異常系 */
// 	result := IsEnableChildUnit("Temperature")
// 	if result {
// 		t.Errorf("異常系 Disableエラー")
// 	}

// 	result = IsEnableChildUnit("Temperature")
// 	if result {
// 		t.Errorf("異常系 Disableエラー")
// 	}
// 	db.GormDB.Save(ChildUnit{
// 		ID:     "1",
// 		Enable: 1,
// 	})
// 	result = IsEnableChildUnit("Temperature")
// 	if result {
// 		t.Errorf("異常系 Disableエラー")
// 	}

// 	/* 正常系 */
// 	db.GormDB.Delete(&ChildUnit{})
// 	db.GormDB.Save(ChildUnit{
// 		ID:     "1",
// 		Enable: 0,
// 	})

// 	result = IsEnableChildUnit("Temperature")
// 	if result {
// 		t.Errorf("Disableエラー")
// 	}

// 	db.GormDB.Where("id = ?", 1).Delete(&ChildUnit{})
// 	db.GormDB.Save(ChildUnit{
// 		ID:     "1",
// 		Enable: 1,
// 	})
// 	result = IsEnableChildUnit("Temperature")
// 	if !result {
// 		t.Errorf("Enableエラー")
// 	}
// 	db.Close()
// }

// func TestIOunitInputProcess(t *testing.T) {
// 	t.Skip()
// }

// func TestEnergyUnitProcess(t *testing.T) {
// 	t.Skip()
// }

// func TestEnvironmentUnitProcess(t *testing.T) {
// 	t.Skip()
// }

// func TestOutput_CommunicationConfirmation(t *testing.T) {
// 	t.Skip()
// }

// func TestIOunit_input(t *testing.T) {
// 	t.Skip()
// }

// func TestIOunit_output(t *testing.T) {
// 	t.Skip()
// }

// func TestGetEnergyValue(t *testing.T) {
// 	t.Skip()
// }

// /** 温度テスト **/

// type TestTemperature struct{}
// type TestTemperature2 struct{}

// const testTemperatureValue float64 = 11.11
// const testMinusTemperatureValue float64 = -22.22

// func TestGetTemperatureValue(t *testing.T) {
// 	t.Skip()
// }

// func (tt *TestTemperature) UpdateData() error {
// 	return nil
// }

// func (tt *TestTemperature) GetTemperature(channel uint16, value, ratio sql.NullFloat64) (float64, error) {
// 	return testTemperatureValue, nil
// }

// func (tt2 *TestTemperature2) UpdateData() error {
// 	return nil
// }
// func (tt2 *TestTemperature2) GetTemperature(channel uint16, value, ratio sql.NullFloat64) (float64, error) {
// 	return testMinusTemperatureValue, nil
// }

// func TestTemperaturehumidityUnit(t *testing.T) {
// 	t.Skip()
// }
