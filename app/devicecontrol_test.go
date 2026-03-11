package app_test

// import (
// 	"database/sql"
// 	"etos-mainunit/app"
// 	"fmt"
// 	"sync"
// 	"testing"
// 	"time"
// )

// func TestControlRun(t *testing.T) {
// 	/* 終了テスト */
// 	t.Run("ControlRun End Test", func(t *testing.T) {
// 		controlRunEndTest()
// 	})
// }

// func TestDataGet(t *testing.T) {
// 	db, err := app.NewDatabase() // DBの初期化
// 	if err != nil {
// 		app.Logger.Writef(app.LOG_LEVEL_ERR, "NewDatabase() %s", err.Error())
// 	}

// 	makeTestDB()

// 	chSendCloud := make(chan app.ChannelMessage, 10) // クラウドchanel

// 	device := app.NewDeviceControl() // デバイス制御
// 	var wg sync.WaitGroup
// 	device.Run(&wg, chSendCloud)

// 	/** 制御テスト 1つ **/
// 	/* 遠隔デフロスト制御 */
// 	t.Run("ControlRun RemoteDefrost Test", func(t *testing.T) {
// 		makeRemoteControlDB(0, "app.DeviceInformation")
// 		makeDemandControlDB(0)
// 		makeRemoteDefrostDB(1, "OutputContact2")
// 		makePresentValuePossibleControlDB("app.DeviceInformation")
// 		currentTime := time.Now()
// 		db.GormDB.Save(&app.ControlStatus{
// 			ID:               "app.DeviceInformation",
// 			ControlStartTime: 0,
// 			ControlEndTime:   currentTime.Unix() - 20,
// 		})
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunRemoteDefrostControlTest(t, device.out, "app.DeviceInformation", currentTime)
// 	})

// 	/* デフロスト制御 */
// 	testIoUnit := new(TestIoUnit)
// 	SetIoUnitTestInterface(testIoUnit)
// 	TestInputValue[0] = 1
// 	t.Run("ControlRun Defrost Test", func(t *testing.T) {
// 		makeRemoteDefrostDB(0, "OutputContact2")
// 		currentTime := time.Now()
// 		db.GormDB.Save(&app.ControlStatus{
// 			ID:               "app.DeviceInformation",
// 			ControlStartTime: 0,
// 			ControlEndTime:   currentTime.Unix() - 20,
// 		})
// 		device.to <- app.ChannelMessage{DataGet, "true", currentTime}
// 		controlRunDefrostControlTest(t, device.out, "app.DeviceInformation", currentTime)
// 	})

// 	TestInputValue[0] = 0
// 	/* 自動制御テスト制御(停止) */
// 	t.Run("ControlRun AutoControl Start Test", func(t *testing.T) {
// 		currentTime := time.Now()
// 		db.GormDB.Save(&app.ControlStatus{
// 			ID:               "app.DeviceInformation",
// 			ControlStartTime: 0,
// 			ControlEndTime:   currentTime.Unix() - 20,
// 		})
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunAutoControlStartTest(t, device.out, "app.DeviceInformation", currentTime)
// 	})

// 	/* 自動制御テスト 解除(稼働) */
// 	t.Run("ControlRun AutoControl Stop Test", func(t *testing.T) {
// 		currentTime := time.Now()
// 		makePresentValueImpossibleControlDB("app.DeviceInformation")
// 		db.GormDB.Save(&app.ControlStatus{
// 			ID:               "app.DeviceInformation",
// 			ControlStartTime: currentTime.Unix() - 20,
// 			ControlEndTime:   0,
// 		})
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunAutoControlStopTest(t, device.out, "app.DeviceInformation", currentTime)
// 	})

// 	/* デマンド制御テスト 制御(停止) */
// 	t.Run("ControlRun DemandControl Start Test", func(t *testing.T) {
// 		makeDemandControlDB(1)
// 		makePresentValuePossibleControlDB("app.DeviceInformation")
// 		makeCloudCommonStateDB(0)
// 		currentTime := time.Now()
// 		db.GormDB.Save(&app.ControlStatus{
// 			ID:               "app.DeviceInformation",
// 			ControlStartTime: 0,
// 			ControlEndTime:   currentTime.Unix() - 20,
// 		})
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunDemandStartTest(t, device.out, "app.DeviceInformation", currentTime)
// 	})

// 	/* デマンド制御テスト 解除(稼働)  */
// 	t.Run("ControlRun DemandControl Stop Test", func(t *testing.T) {
// 		makePresentValueImpossibleControlDB("app.DeviceInformation")
// 		makeCloudCommonStateDB(0)
// 		currentTime := time.Now()
// 		db.GormDB.Save(&app.ControlStatus{
// 			ID:               "app.DeviceInformation",
// 			ControlStartTime: currentTime.Unix() - 100,
// 			ControlEndTime:   0,
// 		})
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunDemandStopTest(t, device.out, "app.DeviceInformation", currentTime)
// 	})

// 	/* 遠隔制御テスト 制御(停止)  */
// 	t.Run("ControlRunRemoteControl Start Test", func(t *testing.T) {
// 		db.GormDB.Delete(&RemoteControl{})
// 		makeRemoteControlDB(1, "app.DeviceInformation")
// 		makeDemandControlDB(0)
// 		makeCloudCommonStateDB(0)
// 		currentTime := time.Now()
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunRemoteControlTest(t, device.out, "app.DeviceInformation", currentTime)
// 	})

// 	fmt.Println("Wait 3sec")
// 	time.Sleep(3 * time.Second)
// 	/* Demand=ON CommunicationError==1 Auto */
// 	t.Run("ControlRun Demand ON Communication Error Test", func(t *testing.T) {
// 		db.GormDB.Delete(&RemoteControl{})
// 		makeRemoteControlDB(0, "app.DeviceInformation")
// 		makeDemandControlDB(1)
// 		makePresentValuePossibleControlDB("app.DeviceInformation")
// 		makeCloudCommonStateDB(1)
// 		currentTime := time.Now()
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunAutoControlStartTest(t, device.out, "app.DeviceInformation", currentTime)
// 	})

// 	/** 制御テスト 3つ **/
// 	addapp.DeviceInformation()
// 	/* デマンド制御テスト*/
// 	t.Run("ControlRun DemandControl Test 3つ", func(t *testing.T) {
// 		db.GormDB.Delete(&RemoteControl{})
// 		makeRemoteControlDB(0, "app.DeviceInformation")
// 		makeDemandControlDB(1)
// 		makePresentValuePossibleControlDB("app.DeviceInformation")
// 		makePresentValueImpossibleControlDB("app.DeviceInformation2")
// 		makePresentValuePossibleControlDB("app.DeviceInformation3")
// 		makeCloudCommonStateDB(0)
// 		currentTime := time.Now()
// 		db.GormDB.Where("id = ?", "app.DeviceInformation").Delete(&app.ControlStatus{})
// 		db.GormDB.Save(&app.ControlStatus{
// 			ID:               "app.DeviceInformation",
// 			ControlStartTime: 0,
// 			ControlEndTime:   currentTime.Unix() - 20,
// 		})
// 		db.GormDB.Save(&DeviceStatus{
// 			DeviceID:      "app.DeviceInformation",
// 			Time:          0,
// 			Control:       0,
// 			Status:        "Auto",
// 			OccupancyRate: 0,
// 		})
// 		db.GormDB.Where("id = ?", "app.DeviceInformation3").Delete(&app.ControlStatus{})
// 		db.GormDB.Save(&app.ControlStatus{
// 			ID:               "app.DeviceInformation3",
// 			ControlStartTime: 0,
// 			ControlEndTime:   currentTime.Unix() - 20,
// 		})
// 		db.GormDB.Save(&DeviceStatus{
// 			DeviceID:      "app.DeviceInformation3",
// 			Time:          0,
// 			Control:       0,
// 			Status:        "Auto",
// 			OccupancyRate: 0,
// 		})
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunDemandStartTest(t, device.out, "app.DeviceInformation", currentTime)
// 		controlRunDemandStopTest(t, device.out, "app.DeviceInformation2", currentTime)
// 		controlRunDemandStartTest(t, device.out, "app.DeviceInformation3", currentTime)
// 	})

// 	/* Auto制御テスト*/
// 	t.Run("ControlRun AutoControl Test 3つ", func(t *testing.T) {
// 		db.GormDB.Delete(&RemoteControl{})
// 		makeRemoteControlDB(0, "app.DeviceInformation")
// 		makeRemoteControlDB(0, "app.DeviceInformation2")
// 		makeRemoteControlDB(0, "app.DeviceInformation3")
// 		makeDemandControlDB(0)
// 		makePresentValuePossibleControlDB("app.DeviceInformation")
// 		makePresentValueImpossibleControlDB("app.DeviceInformation2")
// 		makePresentValuePossibleControlDB("app.DeviceInformation3")
// 		makeCloudCommonStateDB(0)
// 		currentTime := time.Now()
// 		db.GormDB.Save(&app.ControlStatus{
// 			ID:               "app.DeviceInformation",
// 			ControlStartTime: 0,
// 			ControlEndTime:   currentTime.Unix() - 20,
// 		})
// 		db.GormDB.Where("device_id = ?", "app.DeviceInformation").Delete(&DeviceStatus{})
// 		db.GormDB.Save(&DeviceStatus{
// 			DeviceID:      "app.DeviceInformation",
// 			Time:          0,
// 			Control:       0,
// 			Status:        "Auto",
// 			OccupancyRate: 0,
// 		})
// 		db.GormDB.Save(&app.ControlStatus{
// 			ID:               "app.DeviceInformation2",
// 			ControlStartTime: currentTime.Unix() - 20,
// 			ControlEndTime:   0,
// 		})
// 		db.GormDB.Where("device_id = ?", "app.DeviceInformation2").Delete(&DeviceStatus{})
// 		db.GormDB.Save(&DeviceStatus{
// 			DeviceID:      "app.DeviceInformation2",
// 			Time:          0,
// 			Control:       1,
// 			Status:        "Auto",
// 			OccupancyRate: 0,
// 		})
// 		db.GormDB.Save(&app.ControlStatus{
// 			ID:               "app.DeviceInformation3",
// 			ControlStartTime: 0,
// 			ControlEndTime:   currentTime.Unix() - 20,
// 		})
// 		db.GormDB.Where("device_id = ?", "app.DeviceInformation3").Delete(&DeviceStatus{})
// 		db.GormDB.Save(&DeviceStatus{
// 			DeviceID:      "app.DeviceInformation3",
// 			Time:          0,
// 			Control:       0,
// 			Status:        "Auto",
// 			OccupancyRate: 0,
// 		})
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunAutoControlStartTest(t, device.out, "app.DeviceInformation", currentTime)
// 		controlRunAutoControlStopTest(t, device.out, "app.DeviceInformation2", currentTime)
// 		controlRunAutoControlStartTest(t, device.out, "app.DeviceInformation3", currentTime)
// 	})

// 	/* RemoteControlテスト*/
// 	t.Run("ControlRun RemoteControl Start Test 3つ", func(t *testing.T) {
// 		db.GormDB.Delete(&RemoteControl{})
// 		makeRemoteControlDB(1, "app.DeviceInformation")
// 		makeRemoteControlDB(1, "app.DeviceInformation2")
// 		makeRemoteControlDB(1, "app.DeviceInformation3")
// 		makeDemandControlDB(1)
// 		makePresentValuePossibleControlDB("app.DeviceInformation")
// 		makePresentValueImpossibleControlDB("app.DeviceInformation2")
// 		makePresentValuePossibleControlDB("app.DeviceInformation3")
// 		makeCloudCommonStateDB(0)
// 		currentTime := time.Now()
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunRemoteControlTest(t, device.out, "app.DeviceInformation", currentTime)
// 		controlRunRemoteControlTest(t, device.out, "app.DeviceInformation2", currentTime)
// 		controlRunRemoteControlTest(t, device.out, "app.DeviceInformation3", currentTime)
// 	})

// 	/** 状態遷移テスト **/
// 	/* 上で済み */
// 	/* 遠隔デフロスト　-> デフロスト */
// 	/* デフロスト -> 自動制御 */
// 	/* 自動制御 -> デマンド */
// 	/* デマンド -> 遠隔制御 */
// 	/* 遠隔制御 -> 自動制御 */
// 	/* デマンド -> 自動制御 */
// 	/* 自動制御 -> 遠隔制御 */

// 	/* 新規 */
// 	/* 遠隔制御 -> デフロスト */
// 	fmt.Println("Wait 3sec")
// 	time.Sleep(3 * time.Second)
// 	db.GormDB.Where("id = ?", "app.DeviceInformation2").Delete(&app.DeviceInformation{})
// 	db.GormDB.Where("id = ?", "app.DeviceInformation3").Delete(&app.DeviceInformation{})
// 	TestInputValue[0] = 1
// 	t.Run("ControlRun Defrost Test", func(t *testing.T) {
// 		makeRemoteControlDB(0, "app.DeviceInformation")
// 		makeRemoteControlDB(0, "app.DeviceInformation2")
// 		makeRemoteControlDB(0, "app.DeviceInformation3")
// 		makeRemoteDefrostDB(0, "OutputContact2")
// 		currentTime := time.Now()
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunDefrostControlTest(t, device.out, "app.DeviceInformation", currentTime)
// 	})

// 	/* デフロスト -> 遠隔制御 */
// 	TestInputValue[0] = 0
// 	fmt.Println("Wait 3sec")
// 	time.Sleep(3 * time.Second)
// 	/* RemoteControlテスト*/
// 	t.Run("ControlRun RemoteControl Start Test 3つ", func(t *testing.T) {
// 		db.GormDB.Delete(&RemoteControl{})
// 		makeRemoteControlDB(1, "app.DeviceInformation")
// 		makeDemandControlDB(1)
// 		makePresentValuePossibleControlDB("app.DeviceInformation")
// 		makeCloudCommonStateDB(0)
// 		currentTime := time.Now()
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunRemoteControlTest(t, device.out, "app.DeviceInformation", currentTime)
// 	})

// 	/* 遠隔制御 -> デマンド */
// 	/* デマンド制御テスト*/
// 	time.Sleep(3 * time.Second)
// 	t.Run("ControlRun DemandControl Test 3つ", func(t *testing.T) {
// 		db.GormDB.Delete(&RemoteControl{})
// 		makeRemoteControlDB(0, "app.DeviceInformation")
// 		makeDemandControlDB(1)
// 		makePresentValuePossibleControlDB("app.DeviceInformation")
// 		makeCloudCommonStateDB(0)
// 		currentTime := time.Now()
// 		device.to <- ChannelMessage{DataGet, "true", currentTime}
// 		controlRunDemandStartTest(t, device.out, "app.DeviceInformation", currentTime)
// 	})

// 	db.Close()
// }

// func controlRunEndTest() {
// 	var channelMessage ChannelMessage

// 	chSendCloud := make(chan ChannelMessage, 10) // クラウドchanel

// 	device := NewDeviceControl() // デバイス制御
// 	var wg sync.WaitGroup
// 	device.Run(&wg, chSendCloud)

// 	/* End Test */
// 	channelMessage.messageType = End
// 	device.to <- ChannelMessage{End, "true", time.Now()}
// }

// func controlRunRemoteDefrostControlTest(t *testing.T, receiveChMsg <-chan app.ChannelMessage, targetID string, currentTime time.Time) {
// 	fmt.Println("controlRunRemoteDefrostControlTest check")

// 	msg := <-receiveChMsg

// 	expectMsg := app.ChannelMessage{StrControlRemoteDefrost, targetID, 0, currentTime.Unix()}
// 	expectControlStatus := app.ControlStatus{targetID, 0, currentTime.Unix(), StrControlRemoteDefrost, 0}
// 	expectDeviceStatus := DeviceStatus{currentTime.Unix(), targetID, 0, StrControlRemoteDefrost, 0, 0, 0, true}

// 	controlRunCheckStatus(t, msg, expectMsg, expectControlStatus, expectDeviceStatus, 0)

// }

// func controlRunDefrostControlTest(t *testing.T, receiveChMsg <-chan app.ChannelMessage, targetID string, currentTime time.Time) {
// 	fmt.Println("controlRunDefrostControlTest check")

// 	msg := <-receiveChMsg

// 	expectMsg := app.ChannelMessage{StrControlDefrost, targetID, 0, currentTime.Unix()}
// 	expectControlStatus := app.ControlStatus{targetID, 0, currentTime.Unix(), StrControlDefrost, 0}
// 	expectDeviceStatus := DeviceStatus{currentTime.Unix(), targetID, 0, StrControlDefrost, 0, 0, 0, true}

// 	controlRunCheckStatus(t, msg, expectMsg, expectControlStatus, expectDeviceStatus, 0)

// }

// func controlRunAutoControlStartTest(t *testing.T, receiveChMsg <-chan app.ChannelMessage, targetID string, currentTime time.Time) {
// 	fmt.Println("controlRunAutoControlStartTest check")

// 	msg := <-receiveChMsg

// 	expectMsg := app.ChannelMessage{StrControlAuto, targetID, 1, currentTime.Unix()}
// 	expectControlStatus := app.ControlStatus{targetID, currentTime.Unix(), 0, StrControlAuto, 0}
// 	expectDeviceStatus := DeviceStatus{currentTime.Unix(), targetID, 1, StrControlAuto, 0, 0, 0, true}

// 	controlRunCheckStatus(t, msg, expectMsg, expectControlStatus, expectDeviceStatus, 1)
// }

// func controlRunAutoControlStopTest(t *testing.T, receiveChMsg <-chan app.ChannelMessage, targetID string, currentTime time.Time) {
// 	fmt.Println("controlRunAutoControlStopTest Check")

// 	msg := <-receiveChMsg

// 	expectMsg := app.ChannelMessage{StrControlAuto, targetID, 0, currentTime.Unix()}
// 	expectControlStatus := app.ControlStatus{targetID, 0, currentTime.Unix(), StrControlAuto, 0}
// 	expectDeviceStatus := DeviceStatus{currentTime.Unix(), targetID, 0, StrControlAuto, 0, 0, 0, true}

// 	controlRunCheckStatus(t, msg, expectMsg, expectControlStatus, expectDeviceStatus, 0)
// }

// func controlRunDemandStartTest(t *testing.T, receiveChMsg <-chan app.ChannelMessage, targetID string, currentTime time.Time) {
// 	fmt.Println("controlRunDemandStartTest Check")

// 	msg := <-receiveChMsg

// 	expectMsg := app.ChannelMessage{StrControlDemand, targetID, 1, currentTime.Unix()}
// 	expectControlStatus := app.ControlStatus{targetID, currentTime.Unix(), 0, StrControlDemand, 0}
// 	expectDeviceStatus := DeviceStatus{currentTime.Unix(), targetID, 1, StrControlDemand, 0, 0, 0, true}

// 	controlRunCheckStatus(t, msg, expectMsg, expectControlStatus, expectDeviceStatus, 1)
// }

// func controlRunDemandStopTest(t *testing.T, receiveChMsg <-chan app.ChannelMessage, targetID string, currentTime time.Time) {
// 	fmt.Println("controlRunDemandStopTest Check")

// 	msg := <-receiveChMsg

// 	expectMsg := app.ChannelMessage{StrControlDemand, targetID, 0, currentTime.Unix()}
// 	expectControlStatus := app.ControlStatus{targetID, 0, currentTime.Unix(), StrControlDemand, 0}
// 	expectDeviceStatus := DeviceStatus{currentTime.Unix(), targetID, 0, StrControlDemand, 0, 0, 0, true}

// 	controlRunCheckStatus(t, msg, expectMsg, expectControlStatus, expectDeviceStatus, 0)
// }

// func controlRunRemoteControlTest(t *testing.T, receiveChMsg <-chan app.ChannelMessage, targetID string, currentTime time.Time) {
// 	fmt.Println("controlRunRemoteControlTest Check")
// 	msg := <-receiveChMsg

// 	expectMsg := app.ChannelMessage{StrControlRemote, targetID, 0, currentTime.Unix()}
// 	expectControlStatus := app.ControlStatus{targetID, 0, currentTime.Unix(), StrControlRemote, 0}
// 	expectDeviceStatus := DeviceStatus{currentTime.Unix(), targetID, 1, StrControlRemote, 0, 0, 0, true}

// 	controlRunCheckStatus(t, msg, expectMsg, expectControlStatus, expectDeviceStatus, 2)
// }

// func controlRunCheckStatus(t *testing.T, receiveMsg app.ChannelMessage, expectMsg app.ChannelMessage,
// 	expectaControlStatus app.ControlStatus, expectDeviceStatus DeviceStatus, deviceStop int8) {
// 	var ControlStatus app.ControlStatus
// 	var deviceStatus DeviceStatus

// 	if receiveMsg.ControlType != expectMsg.ControlType {
// 		t.Errorf("メッセージControlTypeエラー recv=%v, expect=%v", receiveMsg.ControlType, expectMsg.ControlType)
// 	}
// 	if receiveMsg.time != expectMsg.time {
// 		t.Errorf("メッセージTimeエラー msg.time=%v, currentTime=%v", receiveMsg.time, expectMsg.time)
// 	}
// 	if deviceStop == 0 || deviceStop == 1 {
// 		if receiveMsg.DeviceStop != deviceStop {
// 			t.Errorf("状態エラー ID=%v, current=%v, expect=%v", expectMsg.ID, receiveMsg.DeviceStop, deviceStop)
// 		}
// 	}
// 	if receiveMsg.ID != expectMsg.ID {
// 		t.Errorf("IDエラー")
// 	}
// 	result := db.GormDB.First(&app.ControlStatus, "id = ?", expectapp.ControlStatus.ID)
// 	if result.Error != nil {
// 		t.Errorf("app.ControlStatus DBエラー targetID=%v, %v", expectMsg.ID, result.Error.Error())
// 	} else {
// 		if app.ControlStatus.ID != expectapp.ControlStatus.ID {
// 			t.Errorf("app.ControlStatus IDエラー")
// 		}
// 		if app.ControlStatus.Status != expectapp.ControlStatus.Status {
// 			t.Errorf("app.ControlStatus Statusエラー=%v, expect=%v", app.ControlStatus.Status, expectapp.ControlStatus.Status)
// 		}
// 		if deviceStop == 1 || expectDeviceStatus.Status == StrControlRemoteDefrost {
// 			if app.ControlStatus.ControlStartTime != expectapp.ControlStatus.ControlStartTime {
// 				t.Errorf("app.ControlStatus TIMEエラー ID=%v, app.ControlStatus.ControlStartTime=%v, expectapp.ControlStatus.ControlStartTime=%v", app.ControlStatus.ID, app.ControlStatus.ControlStartTime, expectapp.ControlStatus.ControlStartTime)
// 			}
// 		}
// 		/* else if deviceStop == 0 {
// 			if app.ControlStatus.ControlEndTime != expectapp.ControlStatus.ControlEndTime {
// 				t.Errorf("app.ControlStatus TIMEエラー app.ControlStatus.ControlEndTime=%v, expectapp.ControlStatus.ControlEndTime=%v", app.ControlStatus.ControlEndTime, expectapp.ControlStatus.ControlEndTime)
// 			}
// 		} */
// 	}
// 	result = db.GormDB.First(&deviceStatus, "device_id = ?", expectapp.ControlStatus.ID)
// 	if result.Error != nil {
// 		t.Errorf("DeviceStatus DBエラー targetID=%v, %v", receiveMsg.ID, result.Error.Error())
// 	} else {
// 		if deviceStatus.DeviceID != expectDeviceStatus.DeviceID {
// 			t.Errorf("DeviceStatus IDエラー %v", deviceStatus)
// 		}
// 		if deviceStatus.Control != expectDeviceStatus.Control {
// 			t.Errorf("DeviceStatus ID=%v, Controlエラー currentStatus=%v, expectStatus=%v", expectDeviceStatus.DeviceID, deviceStatus.Control, expectDeviceStatus.Control)
// 		}
// 		if deviceStatus.Status != expectDeviceStatus.Status {
// 			t.Errorf("DeviceStatus Statusエラー %v, %v", deviceStatus, expectDeviceStatus)
// 		}
// 	}
// }

// func makeTestDB() {

// 	tx := db.GormDB.Begin()

// 	tx.Delete(&BaseInformation{})
// 	tx.Delete(&ChildUnit{})
// 	tx.Delete((&DemandPulseUnit{}))
// 	tx.Delete(&EnergySensor{})
// 	tx.Delete(&EnvironmentSensor{})
// 	tx.Delete(&InputContact{})
// 	tx.Delete(&OutputContact{})
// 	tx.Delete(&app.DeviceInformation{})
// 	tx.Delete(&app.ControlConditions{})
// 	tx.Delete(&RemoteControl{})
// 	tx.Delete(&DemandControl{})

// 	tx.Delete(&PresentValue{})
// 	tx.Delete(&DeviceStatus{})
// 	tx.Delete(&app.ControlStatus{})

// 	tx.Create(&BaseInformation{ID: "192168", ContactElectricPower: 300, TargetElectricPower: 280})

// 	tx.Create(&ChildUnit{
// 		ID:     "1",
// 		Enable: 1,
// 	})
// 	tx.Create(&ChildUnit{
// 		ID:     "2",
// 		Enable: 1,
// 	})

// 	tx.Create(&DemandPulseUnit{
// 		ID:                     "DemandPulseUnit",
// 		ElectricEnergyPerPulse: 1,
// 	})
// 	tx.Create(&EnergySensor{
// 		ID:          "EnergySensor",
// 		DeviceName:  "KE1",
// 		Voltage:     200,
// 		PowerFactor: 1.0,
// 		//ControlMethod:             "Stop",
// 		RatedPowerConsumptionCool: 450,
// 		RatedPowerConsumptionWarm: 460,
// 		ControlID:                 5,
// 		ControlChannel:            1,
// 		Enable:                    1,
// 	})
// 	tx.Create(&EnvironmentSensor{
// 		ID:              "EnvironmentSensor",
// 		DeviceName:      "ADAM4015",
// 		Category:        "Temperature",
// 		CorrectionValue: NullFloat64{sql.NullFloat64{1.0, true}},
// 		CorrectionRatio: NullFloat64{sql.NullFloat64{0.01, true}},
// 		ControlID:       10,
// 		ControlChannel:  1,
// 		Enable:          1,
// 	})
// 	tx.Create(&InputContact{
// 		ID:             "InputContact1",
// 		DeviceName:     "UnitTest1",
// 		Category:       "Defrost",
// 		ControlID:      1,
// 		ControlChannel: 1,
// 		Enable:         1,
// 	})
// 	tx.Create(&InputContact{
// 		ID:             "InputContact2",
// 		DeviceName:     "R1212",
// 		Category:       "Breaker",
// 		ControlID:      1,
// 		ControlChannel: 2,
// 		Enable:         1,
// 	})
// 	tx.Create(&OutputContact{
// 		ID:             "OutputContact1",
// 		DeviceName:     "R1212",
// 		Category:       "Control",
// 		ControlMethod:  "Stop",
// 		ControlID:      1,
// 		ControlChannel: 1,
// 		Enable:         1,
// 	})
// 	tx.Create(&OutputContact{
// 		ID:             "OutputContact2",
// 		DeviceName:     "R1212",
// 		Category:       "DefrostControl",
// 		ControlMethod:  "Stop",
// 		ControlID:      1,
// 		ControlChannel: 2,
// 		Enable:         1,
// 	})
// 	tx.Create(&app.DeviceInformation{
// 		ID:                      "app.DeviceInformation",
// 		Mode:                    "Auto",
// 		StopElectricPower:       260,
// 		RequiredControllingTime: 2,
// 		EnergySensorID:          "EnergySensor",
// 		//EnvironmentSensorID:         "EnvironmentSensor",
// 		DefrostInputID:              "InputContact1",
// 		IntermittenControlContactID: "",
// 		CapacityControlContactID:    "OutputContact1",
// 		DefrostControlContactID:     "OutputContact2",
// 	})
// 	tx.Create(&app.ControlConditions{
// 		TargetDevicesID:             "app.DeviceInformation",
// 		ControlTime:                 5,
// 		ReleaseControlTime:          10,
// 		ControlStartTemperature:     app.NullFloat64{sql.NullFloat64{Float64: 35, Valid: true}},
// 		ControlStopTemperature:      app.NullFloat64{sql.NullFloat64{Float64: 20, Valid: true}},
// 		ControlStartHumidity:        app.NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}},
// 		ControlStopHumidity:         app.NullFloat64{sql.NullFloat64{Float64: 30, Valid: true}},
// 		ControlStartDiscomfortIndex: app.NullFloat64{sql.NullFloat64{Float64: 75, Valid: true}},
// 		ControlStopDiscomfortIndex:  app.NullFloat64{sql.NullFloat64{Float64: 65, Valid: true}},
// 		DemandStartTemperature:      app.NullFloat64{sql.NullFloat64{Float64: 30, Valid: true}},
// 		DemandStopTemperature:       app.NullFloat64{sql.NullFloat64{Float64: 25, Valid: true}},
// 		DemandStartHumidity:         app.NullFloat64{sql.NullFloat64{Float64: 52, Valid: true}},
// 		DemandStopHumidity:          app.NullFloat64{sql.NullFloat64{Float64: 32, Valid: true}},
// 		DemandStartDiscomfortIndex:  app.NullFloat64{sql.NullFloat64{Float64: 73, Valid: true}},
// 		DemandStopDiscomfortIndex:   app.NullFloat64{sql.NullFloat64{Float64: 63, Valid: true}},
// 		ControlEnable:               1,
// 		DemandEnable:                1,
// 	})
// 	tx.Commit()
// }

// func addDeviceInformation() {

// 	tx := db.GormDB.Begin()

// 	tx.Create(&ChildUnit{
// 		ID:     "3",
// 		Enable: 1,
// 	})

// 	/* Device2 */
// 	tx.Create(&DemandPulseUnit{
// 		ID:                     "DemandPulseUnit2",
// 		ElectricEnergyPerPulse: 1,
// 	})
// 	tx.Create(&EnergySensor{
// 		ID:          "EnergySensor2",
// 		DeviceName:  "KE1",
// 		Voltage:     200,
// 		PowerFactor: 1.0,
// 		//ControlMethod:             "Stop",
// 		RatedPowerConsumptionCool: 450,
// 		RatedPowerConsumptionWarm: 460,
// 		ControlID:                 5,
// 		ControlChannel:            1,
// 		Enable:                    1,
// 	})
// 	tx.Create(&EnvironmentSensor{
// 		ID:              "EnvironmentSensor2",
// 		DeviceName:      "ADAM4015",
// 		Category:        "Temperature",
// 		CorrectionValue: NullFloat64{sql.NullFloat64{Float64: 1.0, Valid: true}},
// 		CorrectionRatio: NullFloat64{sql.NullFloat64{Float64: 0.01, Valid: true}},
// 		ControlID:       10,
// 		ControlChannel:  1,
// 		Enable:          1,
// 	})
// 	tx.Create(&InputContact{
// 		ID:             "InputContact3",
// 		DeviceName:     "R1212",
// 		Category:       "Defrost",
// 		ControlID:      1,
// 		ControlChannel: 1,
// 		Enable:         1,
// 	})
// 	tx.Create(&InputContact{
// 		ID:             "InputContact4",
// 		DeviceName:     "R1212",
// 		Category:       "Breaker",
// 		ControlID:      1,
// 		ControlChannel: 2,
// 		Enable:         1,
// 	})
// 	tx.Create(&OutputContact{
// 		ID:             "OutputContact3",
// 		DeviceName:     "R1212",
// 		Category:       "Control",
// 		ControlMethod:  "Stop",
// 		ControlID:      1,
// 		ControlChannel: 1,
// 		Enable:         1,
// 	})
// 	tx.Create(&OutputContact{
// 		ID:             "OutputContact4",
// 		DeviceName:     "R1212",
// 		Category:       "DefrostControl",
// 		ControlMethod:  "Stop",
// 		ControlID:      1,
// 		ControlChannel: 2,
// 		Enable:         1,
// 	})
// 	tx.Create(&app.DeviceInformation{
// 		ID:                      "app.DeviceInformation2",
// 		Mode:                    "Auto",
// 		StopElectricPower:       260,
// 		RequiredControllingTime: 2,
// 		EnergySensorID:          "EnergySensor2",
// 		//EnvironmentSensorID:         "EnvironmentSensor2",
// 		DefrostInputID:              "InputContact3",
// 		IntermittenControlContactID: "",
// 		CapacityControlContactID:    "OutputContact3",
// 		DefrostControlContactID:     "OutputContact4",
// 	})
// 	tx.Create(&app.ControlConditions{
// 		TargetDevicesID:             "app.DeviceInformation2",
// 		ControlTime:                 5,
// 		ReleaseControlTime:          10,
// 		ControlStartTemperature:     NullFloat64{sql.NullFloat64{Float64: 35, Valid: true}},
// 		ControlStopTemperature:      NullFloat64{sql.NullFloat64{Float64: 20, Valid: true}},
// 		ControlStartHumidity:        NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}},
// 		ControlStopHumidity:         NullFloat64{sql.NullFloat64{Float64: 30, Valid: true}},
// 		ControlStartDiscomfortIndex: NullFloat64{sql.NullFloat64{Float64: 75, Valid: true}},
// 		ControlStopDiscomfortIndex:  NullFloat64{sql.NullFloat64{Float64: 65, Valid: true}},
// 		DemandStartTemperature:      NullFloat64{sql.NullFloat64{Float64: 30, Valid: true}},
// 		DemandStopTemperature:       NullFloat64{sql.NullFloat64{Float64: 25, Valid: true}},
// 		DemandStartHumidity:         NullFloat64{sql.NullFloat64{Float64: 52, Valid: true}},
// 		DemandStopHumidity:          NullFloat64{sql.NullFloat64{Float64: 32, Valid: true}},
// 		DemandStartDiscomfortIndex:  NullFloat64{sql.NullFloat64{Float64: 73, Valid: true}},
// 		DemandStopDiscomfortIndex:   NullFloat64{sql.NullFloat64{Float64: 63, Valid: true}},
// 		ControlEnable:               1,
// 		DemandEnable:                1,
// 	})

// 	/* Device3 */
// 	tx.Create(&DemandPulseUnit{
// 		ID:                     "DemandPulseUnit3",
// 		ElectricEnergyPerPulse: 1,
// 	})
// 	tx.Create(&EnergySensor{
// 		ID:          "EnergySensor3",
// 		DeviceName:  "KE1",
// 		Voltage:     200,
// 		PowerFactor: 1.0,
// 		//ControlMethod:             "Stop",
// 		RatedPowerConsumptionCool: 450,
// 		RatedPowerConsumptionWarm: 460,
// 		ControlID:                 5,
// 		ControlChannel:            1,
// 		Enable:                    1,
// 	})
// 	tx.Create(&EnvironmentSensor{
// 		ID:              "EnvironmentSensor3",
// 		DeviceName:      "ADAM4015",
// 		Category:        "Temperature",
// 		CorrectionValue: NullFloat64{sql.NullFloat64{Float64: 1.0, Valid: true}},
// 		CorrectionRatio: NullFloat64{sql.NullFloat64{Float64: 0.01, Valid: true}},
// 		ControlID:       10,
// 		ControlChannel:  1,
// 		Enable:          1,
// 	})
// 	tx.Create(&InputContact{
// 		ID:             "InputContact5",
// 		DeviceName:     "R1212",
// 		Category:       "Defrost",
// 		ControlID:      1,
// 		ControlChannel: 1,
// 		Enable:         1,
// 	})
// 	tx.Create(&InputContact{
// 		ID:             "InputContact6",
// 		DeviceName:     "R1212",
// 		Category:       "Breaker",
// 		ControlID:      1,
// 		ControlChannel: 2,
// 		Enable:         1,
// 	})
// 	tx.Create(&OutputContact{
// 		ID:             "OutputContact5",
// 		DeviceName:     "R1212",
// 		Category:       "Control",
// 		ControlMethod:  "Stop",
// 		ControlID:      1,
// 		ControlChannel: 1,
// 		Enable:         1,
// 	})
// 	tx.Create(&OutputContact{
// 		ID:             "OutputContact6",
// 		DeviceName:     "R1212",
// 		Category:       "DefrostControl",
// 		ControlMethod:  "Stop",
// 		ControlID:      1,
// 		ControlChannel: 2,
// 		Enable:         1,
// 	})
// 	tx.Create(&app.DeviceInformation{
// 		ID:                      "app.DeviceInformation3",
// 		Mode:                    "Auto",
// 		StopElectricPower:       260,
// 		RequiredControllingTime: 2,
// 		EnergySensorID:          "EnergySensor3",
// 		//EnvironmentSensorID:         "EnvironmentSensor3",
// 		DefrostInputID:              "InputContact5",
// 		IntermittenControlContactID: "",
// 		CapacityControlContactID:    "OutputContact5",
// 		DefrostControlContactID:     "OutputContact6",
// 	})
// 	tx.Create(&app.ControlConditions{
// 		TargetDevicesID:             "app.DeviceInformation3",
// 		ControlTime:                 5,
// 		ReleaseControlTime:          10,
// 		ControlStartTemperature:     NullFloat64{sql.NullFloat64{Float64: 35, Valid: true}},
// 		ControlStopTemperature:      NullFloat64{sql.NullFloat64{Float64: 20, Valid: true}},
// 		ControlStartHumidity:        NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}},
// 		ControlStopHumidity:         NullFloat64{sql.NullFloat64{Float64: 30, Valid: true}},
// 		ControlStartDiscomfortIndex: NullFloat64{sql.NullFloat64{Float64: 75, Valid: true}},
// 		ControlStopDiscomfortIndex:  NullFloat64{sql.NullFloat64{Float64: 65, Valid: true}},
// 		DemandStartTemperature:      NullFloat64{sql.NullFloat64{Float64: 30, Valid: true}},
// 		DemandStopTemperature:       NullFloat64{sql.NullFloat64{Float64: 25, Valid: true}},
// 		DemandStartHumidity:         NullFloat64{sql.NullFloat64{Float64: 52, Valid: true}},
// 		DemandStopHumidity:          NullFloat64{sql.NullFloat64{Float64: 32, Valid: true}},
// 		DemandStartDiscomfortIndex:  NullFloat64{sql.NullFloat64{Float64: 73, Valid: true}},
// 		DemandStopDiscomfortIndex:   NullFloat64{sql.NullFloat64{Float64: 63, Valid: true}},
// 		ControlEnable:               1,
// 		DemandEnable:                1,
// 	})

// 	tx.Commit()
// }

// func makeRemoteDefrostDB(enable int8, targetID string) {
// 	db.GormDB.Where("output_contact_id = ?", targetID).Delete(&RemoteDefrostCommand{})
// 	db.GormDB.Save(&RemoteDefrostCommand{
// 		OutputContactID: targetID,
// 		StopCommand:     enable,
// 	})
// }

// func makeRemoteControlDB(enable int8, targetID string) {
// 	db.GormDB.Where("target_device_id = ?", targetID).Delete(&RemoteControl{})
// 	db.GormDB.Save(&RemoteControl{
// 		TargetDeviceId: targetID,
// 		ControlCommand: enable,
// 	})
// }

// func makeDemandControlDB(enable int8) {
// 	db.GormDB.Delete(&DemandControl{})
// 	db.GormDB.Create(&DemandControl{
// 		DemandControl: enable,
// 	})
// }

// func makeCloudCommonStateDB(enable int) {
// 	db.GormDB.Delete(&CloudCommonState{})
// 	db.GormDB.Save(&CloudCommonState{
// 		CommunicationError: enable,
// 	})
// }

// func makePresentValuePossibleControlDB(targetID string) {
// 	db.GormDB.Where("id = ?", targetID).Delete(&PresentValue{})
// 	db.GormDB.Save(&PresentValue{
// 		ID:                     targetID,
// 		CurrentPower:           300,
// 		CurrentTemperature:     -20,
// 		CurrentHumidity:        20,
// 		CurrentDiscomfortIndex: 20,
// 	})
// }

// func makePresentValueImpossibleControlDB(targetID string) {
// 	db.GormDB.Where("id = ?", targetID).Delete(&PresentValue{})
// 	db.GormDB.Save(&PresentValue{
// 		ID:                     targetID,
// 		CurrentPower:           10,
// 		CurrentTemperature:     50,
// 		CurrentHumidity:        70,
// 		CurrentDiscomfortIndex: 60,
// 	})
// }

// func TestdemandControlJudge(t *testing.T) {
// 	var deviceInfo app.DeviceInformation
// 	var controlData app.ControlConditions
// 	var value PresentValue
// 	var ControlStatus app.ControlStatus

// 	db, err := app.NewDatabase() // DBの初期化
// 	if err != nil {
// 		app.Logger.Writef(app.LOG_LEVEL_ERR, "NewDatabase() %s", err.Error())
// 	}

// 	controlData.DemandStartTemperature = NullFloat64{sql.NullFloat64{Float64: 35, Valid: true}}
// 	controlData.DemandStartHumidity = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.DemandStartDiscomfortIndex = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.DemandStopTemperature = NullFloat64{sql.NullFloat64{Float64: 35, Valid: true}}
// 	controlData.DemandStopHumidity = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.DemandStopDiscomfortIndex = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.ControlEnable = 1
// 	controlData.DemandEnable = 1

// 	value.CurrentHumidity = 40
// 	value.CurrentDiscomfortIndex = 40
// 	value.CurrentTemperature = 30
// 	value.CurrentPower = 200

// 	deviceInfo.StopElectricPower = 100
// 	deviceInfo.RequiredControllingTime = 2
// 	ControlStatus.ControlEndTime = time.Now().Unix() - 20

// 	/* デマンド開始
// 	 */
// 	ControlStatus.Status = StrControlDemand

// 	ControlStatus.Control = 0
// 	demandControlJudge(deviceInfo, controlData, value, &ControlStatus)
// 	if ControlStatus.Control != 1 {
// 		t.Errorf("デマンド開始判定エラー")
// 	}

// 	ControlStatus.Control = 0
// 	ControlStatus.ControlEndTime = time.Now().Unix() - 1
// 	demandControlJudge(deviceInfo, controlData, value, &ControlStatus)
// 	if ControlStatus.Control != 0 {
// 		t.Errorf("デマンド開始判定エラー")
// 	}

// 	ControlStatus.Control = 0
// 	ControlStatus.ControlEndTime = time.Now().Unix() - 100
// 	value.CurrentPower = 50
// 	demandControlJudge(deviceInfo, controlData, value, &ControlStatus)
// 	if ControlStatus.Control != 0 {
// 		t.Errorf("デマンド開始判定エラー")
// 	}

// 	/* デマンド終了 */
// 	ControlStatus.Control = 1
// 	demandControlJudge(deviceInfo, controlData, value, &ControlStatus)
// 	if ControlStatus.Control != 1 {
// 		t.Errorf("デマンド終了判定エラー")
// 	}

// 	ControlStatus.Control = 1
// 	value.CurrentTemperature = 40
// 	demandControlJudge(deviceInfo, controlData, value, &ControlStatus)
// 	if ControlStatus.Control != 0 {
// 		t.Errorf("デマンド終了判定エラー")
// 	}
// 	db.Close()
// }

// func TestautoControlJudge(t *testing.T) {
// 	var deviceInfo app.DeviceInformation
// 	var controlData app.ControlConditions
// 	var value app.PresentValue
// 	var ControlStatus app.ControlStatus

// 	db, err := app.NewDatabase() // DBの初期化
// 	if err != nil {
// 		app.Logger.Writef(app.LOG_LEVEL_ERR, "NewDatabase() %s", err.Error())
// 	}

// 	currentTime := time.Now().Unix()
// 	controlData.ControlStartTemperature = app.NullFloat64{sql.NullFloat64{Float64: 35, Valid: true}}
// 	controlData.ControlStartHumidity = app.NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.ControlStartDiscomfortIndex = app.NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.ControlStopTemperature = app.NullFloat64{sql.NullFloat64{Float64: 35, Valid: true}}
// 	controlData.ControlStopHumidity = app.NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.ControlStopDiscomfortIndex = app.NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.ControlEnable = 1

// 	value.CurrentHumidity = 40
// 	value.CurrentDiscomfortIndex = 40
// 	value.CurrentTemperature = 30
// 	value.CurrentPower = 150

// 	deviceInfo.StopElectricPower = 100
// 	deviceInfo.RequiredControllingTime = 2
// 	app.ControlStatus.ControlEndTime = time.Now().Unix() - 100

// 	/** 自動制御開始 **/
// 	app.ControlStatus.Status = StrControlAuto

// 	/* 自動制御開始 */
// 	controlData.ControlTime = 10
// 	controlData.ReleaseControlTime = 10
// 	app.ControlStatus.Control = 0
// 	autoControlJudge(deviceInfo, controlData, value, &app.ControlStatus, currentTime)
// 	if app.ControlStatus.Control != 1 {
// 		t.Errorf("自動制御開始判定エラー")
// 	}

// 	/* 自動制御開始(温度オーバー) */
// 	app.ControlStatus.Control = 0
// 	controlData.ControlTime = 10
// 	controlData.ReleaseControlTime = 10
// 	app.ControlStatus.ControlEndTime = time.Now().Unix() - 100
// 	value.CurrentTemperature = 50
// 	autoControlJudge(deviceInfo, controlData, value, &app.ControlStatus, currentTime)
// 	if app.ControlStatus.Control != 0 {
// 		t.Errorf("自動制御開始判定エラー")
// 	}

// 	/* 自動制御開始(コントロール時間内) */
// 	app.ControlStatus.Control = 0
// 	app.ControlStatus.ControlEndTime = time.Now().Unix() - 3
// 	value.CurrentTemperature = 30
// 	autoControlJudge(deviceInfo, controlData, value, &app.ControlStatus, currentTime)
// 	if app.ControlStatus.Control != 0 {
// 		t.Errorf("自動制御開始判定エラー")
// 	}

// 	/* 自動制御開始(温度制御、コントロール時間内) */
// 	controlData.ControlTime = 0
// 	controlData.ReleaseControlTime = 0
// 	autoControlJudge(deviceInfo, controlData, value, &app.ControlStatus, currentTime)
// 	if app.ControlStatus.Control != 1 {
// 		t.Errorf("自動制御開始判定エラー")
// 	}

// 	/* 自動制御開始(稼働率=停止) */
// 	app.ControlStatus.Control = 0
// 	controlData.ControlTime = 10
// 	controlData.ReleaseControlTime = 10
// 	app.ControlStatus.ControlEndTime = time.Now().Unix() - 100
// 	value.CurrentPower = 50
// 	autoControlJudge(deviceInfo, controlData, value, &app.ControlStatus, currentTime)
// 	if app.ControlStatus.Control != 0 {
// 		t.Errorf("自動制御開始判定エラー")
// 	}

// 	/** 自動制御終了 **/

// 	/* 制御時間内 */
// 	app.ControlStatus.Control = 1
// 	app.ControlStatus.ControlStartTime = time.Now().Unix() - 5
// 	value.CurrentPower = 50
// 	autoControlJudge(deviceInfo, controlData, value, &app.ControlStatus, currentTime)
// 	if app.ControlStatus.Control != 1 {
// 		t.Errorf("自動制御終了判定エラー")
// 	}

// 	/* 温度制御 */
// 	app.ControlStatus.Control = 1
// 	controlData.ControlTime = 0
// 	controlData.ReleaseControlTime = 0
// 	autoControlJudge(deviceInfo, controlData, value, &app.ControlStatus, currentTime)
// 	if app.ControlStatus.Control != 1 {
// 		t.Errorf("自動制御終了判定エラー")
// 	}

// 	/* 温度オーバー */
// 	app.ControlStatus.Control = 1
// 	value.CurrentTemperature = 40
// 	autoControlJudge(deviceInfo, controlData, value, &app.ControlStatus, currentTime)
// 	if app.ControlStatus.Control != 0 {
// 		t.Errorf("自動制御終了判定エラー")
// 	}

// 	/* 制御時間超過 */
// 	app.ControlStatus.Control = 1
// 	value.CurrentTemperature = 30
// 	controlData.ControlTime = 50
// 	app.ControlStatus.ControlStartTime = time.Now().Unix() - 100
// 	autoControlJudge(deviceInfo, controlData, value, &app.ControlStatus, currentTime)
// 	if app.ControlStatus.Control != 0 {
// 		t.Errorf("自動制御終了判定エラー")
// 	}

// 	/* 温度制御(制御時間超過) */
// 	app.ControlStatus.Control = 1
// 	controlData.ControlTime = 0
// 	controlData.ReleaseControlTime = 0
// 	value.CurrentTemperature = 30
// 	app.ControlStatus.ControlStartTime = time.Now().Unix() - 100
// 	autoControlJudge(deviceInfo, controlData, value, &app.ControlStatus, currentTime)
// 	if app.ControlStatus.Control != 1 {
// 		t.Errorf("自動制御終了判定エラー")
// 	}

// 	db.Close()
// }

// func TestdemandControlDeviceStopJudge(t *testing.T) {
// 	var controlData app.ControlConditions
// 	var value PresentValue
// 	var result int8

// 	controlData.DemandStartTemperature = NullFloat64{sql.NullFloat64{Float64: 35, Valid: true}}
// 	controlData.DemandStartHumidity = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.DemandStartDiscomfortIndex = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.ControlEnable = 1
// 	controlData.DemandEnable = 1

// 	value.CurrentHumidity = 40
// 	value.CurrentDiscomfortIndex = 40
// 	/** 温度テスト **/
// 	/* 温度上 */
// 	value.CurrentTemperature = 40
// 	result = demandControlDeviceStopJudge(controlData, value)
// 	if result != 0 {
// 		t.Errorf("制御開始判定エラー(温度上)")
// 	}

// 	/* 温度同じ */
// 	value.CurrentTemperature = 35
// 	result = demandControlDeviceStopJudge(controlData, value)
// 	if result != 1 {
// 		t.Errorf("制御開始判定エラー(温度同じ)")
// 	}

// 	/* 温度下 */
// 	value.CurrentTemperature = 30
// 	result = demandControlDeviceStopJudge(controlData, value)
// 	if result != 1 {
// 		t.Errorf("制御開始判定エラー(温度下)")
// 	}

// 	if !EnableOnlyTemperatureJudge {
// 		/** 湿度テスト **/
// 		/* 湿度上 */
// 		value.CurrentHumidity = 60
// 		result = demandControlDeviceStopJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(湿度上)")
// 		}

// 		/* 湿度同じ */
// 		value.CurrentHumidity = 50
// 		result = demandControlDeviceStopJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(湿度同じ)")
// 		}

// 		/* 湿度下 */
// 		value.CurrentHumidity = 30
// 		result = demandControlDeviceStopJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(湿度下)")
// 		}

// 		/** 不快指数テスト **/
// 		/* 不快指数上 */
// 		value.CurrentDiscomfortIndex = 60
// 		result = demandControlDeviceStopJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(不快指数上)")
// 		}

// 		/* 不快指数同じ */
// 		value.CurrentDiscomfortIndex = 50
// 		result = demandControlDeviceStopJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(不快指数同じ)")
// 		}

// 		/* 不快指数下 */
// 		value.CurrentDiscomfortIndex = 30
// 		result = demandControlDeviceStopJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(不快指数下)")
// 		}
// 	}
// }

// func TestdemandControlDeviceReleaseJudge(t *testing.T) {
// 	var controlData app.ControlConditions
// 	var value PresentValue
// 	var result int8

// 	controlData.DemandStopTemperature = NullFloat64{sql.NullFloat64{Float64: 35, Valid: true}}
// 	controlData.DemandStopHumidity = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.DemandStopDiscomfortIndex = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.ControlEnable = 1
// 	controlData.DemandEnable = 1

// 	value.CurrentHumidity = 40
// 	value.CurrentDiscomfortIndex = 40
// 	/** 温度テスト **/
// 	/* 温度上 */
// 	value.CurrentTemperature = 40
// 	result = demandControlDeviceReleaseJudge(controlData, value)
// 	if result != 1 {
// 		t.Errorf("制御開始判定エラー(温度上)")
// 	}

// 	/* 温度同じ */
// 	value.CurrentTemperature = 35
// 	result = demandControlDeviceReleaseJudge(controlData, value)
// 	if result != 0 {
// 		t.Errorf("制御開始判定エラー(温度同じ)")
// 	}

// 	/* 温度下 */
// 	value.CurrentTemperature = 30
// 	result = demandControlDeviceReleaseJudge(controlData, value)
// 	if result != 0 {
// 		t.Errorf("制御開始判定エラー(温度下)")
// 	}

// 	if !EnableOnlyTemperatureJudge {
// 		/** 湿度テスト **/
// 		/* 湿度上 */
// 		value.CurrentHumidity = 60
// 		result = demandControlDeviceReleaseJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(湿度上)")
// 		}

// 		/* 湿度同じ */
// 		value.CurrentHumidity = 50
// 		result = demandControlDeviceReleaseJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(湿度同じ)")
// 		}

// 		/* 湿度下 */
// 		value.CurrentHumidity = 30
// 		result = demandControlDeviceReleaseJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(湿度下)")
// 		}

// 		/** 不快指数テスト **/
// 		/* 不快指数上 */
// 		value.CurrentDiscomfortIndex = 60
// 		result = demandControlDeviceReleaseJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(不快指数上)")
// 		}

// 		/* 不快指数同じ */
// 		value.CurrentDiscomfortIndex = 50
// 		result = demandControlDeviceReleaseJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(不快指数同じ)")
// 		}

// 		/* 不快指数下 */
// 		value.CurrentDiscomfortIndex = 30
// 		result = demandControlDeviceReleaseJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(不快指数下)")
// 		}
// 	}
// }

// func TestautoControlDeviceStopJudge(t *testing.T) {
// 	var controlData app.ControlConditions
// 	var value PresentValue
// 	var result int8

// 	controlData.ControlStartTemperature = NullFloat64{sql.NullFloat64{Float64: 35, Valid: true}}
// 	controlData.ControlStartHumidity = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.ControlStartDiscomfortIndex = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.ControlEnable = 1

// 	value.CurrentHumidity = 40
// 	value.CurrentDiscomfortIndex = 40
// 	/** 温度テスト **/
// 	/* 温度上 */
// 	value.CurrentTemperature = 40
// 	result = autoControlDeviceStopJudge(controlData, value)
// 	if result != 0 {
// 		t.Errorf("制御開始判定エラー(温度上)")
// 	}

// 	/* 温度同じ */
// 	value.CurrentTemperature = 35
// 	result = autoControlDeviceStopJudge(controlData, value)
// 	if result != 1 {
// 		t.Errorf("制御開始判定エラー(温度同じ)")
// 	}

// 	/* 温度下 */
// 	value.CurrentTemperature = 30
// 	result = autoControlDeviceStopJudge(controlData, value)
// 	if result != 1 {
// 		t.Errorf("制御開始判定エラー(温度下)")
// 	}

// 	if !EnableOnlyTemperatureJudge {
// 		/** 湿度テスト **/
// 		/* 湿度上 */
// 		value.CurrentHumidity = 60
// 		result = autoControlDeviceStopJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(湿度上)")
// 		}

// 		/* 湿度同じ */
// 		value.CurrentHumidity = 50
// 		result = autoControlDeviceStopJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(湿度同じ)")
// 		}

// 		/* 湿度下 */
// 		value.CurrentHumidity = 30
// 		result = autoControlDeviceStopJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(湿度下)")
// 		}

// 		/** 不快指数テスト **/
// 		/* 不快指数上 */
// 		value.CurrentDiscomfortIndex = 60
// 		result = autoControlDeviceStopJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(不快指数上)")
// 		}

// 		/* 不快指数同じ */
// 		value.CurrentDiscomfortIndex = 50
// 		result = autoControlDeviceStopJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(不快指数同じ)")
// 		}

// 		/* 不快指数下 */
// 		value.CurrentDiscomfortIndex = 30
// 		result = autoControlDeviceStopJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(不快指数下)")
// 		}
// 	}

// 	/** 間欠のみ制御 **/
// 	controlData.ControlStartTemperature = NullFloat64{sql.NullFloat64{Float64: 150, Valid: true}}
// 	value.CurrentTemperature = 40
// 	value.CurrentDiscomfortIndex = 60
// 	value.CurrentHumidity = 60
// 	result = autoControlDeviceStopJudge(controlData, value)
// 	if result != 1 {
// 		t.Errorf("制御開始判定エラー(間欠のみ制御 温度上)")
// 	}
// 	controlData.ControlStartTemperature = NullFloat64{sql.NullFloat64{Float64: -150, Valid: true}}
// 	result = autoControlDeviceStopJudge(controlData, value)
// 	if result != 1 {
// 		t.Errorf("制御開始判定エラー(間欠のみ制御 温度下)")
// 	}
// }
// func TestautoControlDeviceReleaseJudge(t *testing.T) {
// 	var controlData app.ControlConditions
// 	var value PresentValue
// 	var result int8

// 	controlData.ControlStopTemperature = NullFloat64{sql.NullFloat64{Float64: 35, Valid: true}}
// 	controlData.ControlStopHumidity = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.ControlStopDiscomfortIndex = NullFloat64{sql.NullFloat64{Float64: 50, Valid: true}}
// 	controlData.ControlEnable = 1

// 	value.CurrentHumidity = 40
// 	value.CurrentDiscomfortIndex = 40
// 	/** 温度テスト **/
// 	/* 温度上 */
// 	value.CurrentTemperature = 40
// 	result = autoControlDeviceReleaseJudge(controlData, value)
// 	if result != 1 {
// 		t.Errorf("制御開始判定エラー(温度上)")
// 	}

// 	/* 温度同じ */
// 	value.CurrentTemperature = 35
// 	result = autoControlDeviceReleaseJudge(controlData, value)
// 	if result != 0 {
// 		t.Errorf("制御開始判定エラー(温度同じ)")
// 	}

// 	/* 温度下 */
// 	value.CurrentTemperature = 30
// 	result = autoControlDeviceReleaseJudge(controlData, value)
// 	if result != 0 {
// 		t.Errorf("制御開始判定エラー(温度下)")
// 	}

// 	if !EnableOnlyTemperatureJudge {
// 		/** 湿度テスト **/
// 		/* 湿度上 */
// 		value.CurrentHumidity = 60
// 		result = autoControlDeviceReleaseJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(湿度上)")
// 		}

// 		/* 湿度同じ */
// 		value.CurrentHumidity = 50
// 		result = autoControlDeviceReleaseJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(湿度同じ)")
// 		}

// 		/* 湿度下 */
// 		value.CurrentHumidity = 30
// 		result = autoControlDeviceReleaseJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(湿度下)")
// 		}

// 		/** 不快指数テスト **/
// 		/* 不快指数上 */
// 		value.CurrentDiscomfortIndex = 60
// 		result = autoControlDeviceReleaseJudge(controlData, value)
// 		if result != 1 {
// 			t.Errorf("制御開始判定エラー(不快指数上)")
// 		}

// 		/* 不快指数同じ */
// 		value.CurrentDiscomfortIndex = 50
// 		result = autoControlDeviceReleaseJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(不快指数同じ)")
// 		}

// 		/* 不快指数下 */
// 		value.CurrentDiscomfortIndex = 30
// 		result = autoControlDeviceReleaseJudge(controlData, value)
// 		if result != 0 {
// 			t.Errorf("制御開始判定エラー(不快指数下)")
// 		}
// 	}

// 	/** 間欠のみ制御 **/
// 	controlData.ControlStartTemperature = NullFloat64{sql.NullFloat64{Float64: 150, Valid: true}}
// 	value.CurrentTemperature = 40
// 	value.CurrentDiscomfortIndex = 60
// 	value.CurrentHumidity = 60
// 	result = autoControlDeviceReleaseJudge(controlData, value)
// 	if result != 1 {
// 		t.Errorf("制御開始判定エラー(間欠のみ制御 温度上)")
// 	}
// 	controlData.ControlStartTemperature = NullFloat64{sql.NullFloat64{Float64: -150, Valid: true}}
// 	result = autoControlDeviceReleaseJudge(controlData, value)
// 	if result != 1 {
// 		t.Errorf("制御開始判定エラー(間欠のみ制御 温度下)")
// 	}
// }
