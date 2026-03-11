package app

import (
	"fmt"
	"strings"
	"sync"
)

//デマンドパルス監視

type DemandPulseUnitThread struct {
	to   chan ChannelMessage
	from chan ChannelMessage
}

const (
	DEVICE_WS_Z3030 = "WS_Z3030"
	DEVICE_R1212    = "R1212"
)

func demandpulseunitDataGet(chmsg ChannelMessage) {

	overAllCommonTime := chmsg.time.Unix()

	//---------------------
	// DBよりデータ取得
	//---------------------

	var baseInformation BaseInformation
	database.SelectOne(&baseInformation)

	var demandPulseUnit DemandPulseUnit
	database.SelectOne(&demandPulseUnit)

	var demandTightData DemandTightData
	database.SelectOne(&demandTightData)

	//---------------------
	// データ更新
	//---------------------

	// デマンド有無判定
	if demandPulseUnit.ID != "" {
		// デマンド機器よりデータ取得
		demandStatus, err := ReadDemand(demandPulseUnit.DeviceName, byte(demandPulseUnit.ControlID))
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				// 通信異常
				UpdateCommunicationMailAlertStatus(demandPulseUnit.ID, strDemandPulseUnit, overAllCommonTime, COM_NG)
				SaveDBCommunicationAlert(demandPulseUnit.ID, demandPulseUnit.Name, overAllCommonTime, strDemandPulseUnit)
				Logger.Writef(LOG_LEVEL_ERR, "Failed to get status from DemandPulse Connection:%02d Error:%s",
					demandPulseUnit.ControlID,
					err.Error())
			}
		} else {
			Logger.Writef(LOG_LEVEL_DEBUG, "demandStatus:%+v", demandStatus)
			// 目標電力の更新
			demandTightData.TargetPower = baseInformation.TargetElectricPower
			// 現在値の更新
			demandTightData.CurrentPowerIntegration = float64(demandStatus.PulseUnitCount)
			// デマンド逼迫判断
			demandData := CalcDemandPower(&demandTightData, baseInformation, overAllCommonTime, demandPulseUnit.ElectricEnergyPerPulse)
			Logger.Writef(LOG_LEVEL_DEBUG, "デマンドデータ demandData : %+v", demandData)

			// デマンドパルスカウント
			demandPulseCount := DemandPulseCount{overAllCommonTime, int64(demandStatus.PulseUnitCount)}

			// database save demand status
			database.DemandTightDataSaveDB(demandTightData)

			// database saving demand count
			database.DemandPulseCountCloudSaveDB(demandPulseCount)

			// デマンド監視
			DemandJudge(demandPulseUnit.ID, demandData, overAllCommonTime)

			// デマンドパルス アラート通知 有効
			if demandPulseUnit.Enabled == 1 {
				// 電池残量監視
				BatteryLevelAlertJudge(demandPulseUnit, demandStatus, overAllCommonTime)
			}

			UpdateCommunicationMailAlertStatus(demandPulseUnit.ID, strDemandPulseUnit, overAllCommonTime, COM_OK)
		}
	}
}

func NewDemandPulseUnit() (d *DemandPulseUnitThread) {
	d = new(DemandPulseUnitThread)
	d.to = make(chan ChannelMessage, DemandPulseChannelBufferSize)   // デマンド(main -> demand)
	d.from = make(chan ChannelMessage, DemandPulseChannelBufferSize) // デマンド(main <- demand)
	return
}

func (d *DemandPulseUnitThread) Run(wg *sync.WaitGroup, sendchCloudMsg chan ChannelMessage) {
	Logger.Writef(LOG_LEVEL_DEBUG, "Start DemandPulseUnitThread")
	wg.Add(1) // goroutine起動前にカウントを増やす
	go func() {
		d.demandPulseUnitThreadRun(sendchCloudMsg)
		Logger.Writef(LOG_LEVEL_DEBUG, "Stop DemandPulseUnitThread")
		wg.Done()
	}()
}

func (d *DemandPulseUnitThread) demandPulseUnitThreadRun(sendchCloudMsg chan ChannelMessage) {
	for {
		chmsg := <-d.to
		switch chmsg.messageType {
		case End:
			return

		case DataGet:
			demandpulseunitDataGet(chmsg)

			// cloudスレに送る
			msg := ChannelMessage{Alert, strdemandPulse, chmsg.time}
			SendChannelMessageSafely(sendchCloudMsg, msg, false)

			// childUnitスレ終了
			msg = ChannelMessage{End, "", chmsg.time}
			SendChannelMessageSafely(d.from, msg, false)
		}
	}
}

func ReadDemand(deviceName string, id byte) (demandStatus DemandStatus, err error) {
	pu := NewPulseUnit(deviceName, id)
	if pu == nil {
		return demandStatus, fmt.Errorf("deviceName:%s", deviceName)
	}
	count, err := pu.GetPulseCount()
	if err != nil {
		return demandStatus, fmt.Errorf("failed to GetPulseCount%v", err)
	}
	demandStatus.PulseUnitCount = count

	voltage, err := pu.GetBatteryVoltage()
	if err != nil {
		return demandStatus, fmt.Errorf("failed to GetPulseCount%v", err)
	}
	demandStatus.Voltage = voltage

	return demandStatus, err
}
