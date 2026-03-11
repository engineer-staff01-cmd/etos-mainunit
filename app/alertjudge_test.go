package app_test

import (
	"etos-mainunit/app"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	db, err := app.NewDatabase()
	if err != nil {
		log.Printf("NewDatabase() %s", err.Error())
		os.Exit(-1)
	}

	code := m.Run()

	db.Close()
	os.Remove("ecoRAMDAR.db")
	os.Exit(code)
}

// func TestDemandJudge(t *testing.T) {
// 	type args struct {
// 		demandPulseID     string
// 		demandData        ReturnValueCurrentPower
// 		overAllCommonTime int64
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		{
// 			"test:" + StrExcess,
// 			args{
// 				"",
// 				ReturnValueCurrentPower{
// 					currentPower:              20,
// 					leftTime:                  0,
// 					powerIncrease:             0,
// 					predictedPower:            0,
// 					targetCurrentPower:        0,
// 					adjustedPower:             0,
// 					LimitElectricPower:        0,
// 					ContactElectricPower:      0,
// 					TargetElectricPower:       20,
// 					CancellationElectricPower: 0,
// 					InitialElectricPower:      0,
// 				},
// 				1,
// 			},
// 			StrExcess,
// 		},
// 		{
// 			"test:" + Strlimit,
// 			args{
// 				"",
// 				ReturnValueCurrentPower{
// 					currentPower:              20,
// 					leftTime:                  0,
// 					powerIncrease:             0,
// 					predictedPower:            0,
// 					targetCurrentPower:        0,
// 					adjustedPower:             0,
// 					LimitElectricPower:        20,
// 					ContactElectricPower:      30,
// 					TargetElectricPower:       10,
// 					CancellationElectricPower: 0,
// 					InitialElectricPower:      0,
// 				},
// 				1,
// 			},
// 			Strlimit,
// 		},
// 		{
// 			"test:" + StrBeVigilant,
// 			args{
// 				"",
// 				ReturnValueCurrentPower{
// 					currentPower:              20,
// 					leftTime:                  0,
// 					powerIncrease:             0,
// 					predictedPower:            0,
// 					targetCurrentPower:        0,
// 					adjustedPower:             0,
// 					LimitElectricPower:        30,
// 					ContactElectricPower:      30,
// 					TargetElectricPower:       30,
// 					CancellationElectricPower: 0,
// 					InitialElectricPower:      0,
// 				},
// 				1,
// 			},
// 			StrBeVigilant,
// 		},
// 		{
// 			"test" + StrNormal,
// 			args{
// 				"",
// 				ReturnValueCurrentPower{
// 					currentPower:              20,
// 					leftTime:                  0,
// 					powerIncrease:             0,
// 					predictedPower:            0,
// 					targetCurrentPower:        0,
// 					adjustedPower:             -1,
// 					LimitElectricPower:        30,
// 					ContactElectricPower:      30,
// 					TargetElectricPower:       30,
// 					CancellationElectricPower: 0,
// 					InitialElectricPower:      0,
// 				},
// 				1,
// 			},
// 			StrNormal,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			app.DemandJudge(tt.args.demandPulseID, tt.args.demandData, tt.args.overAllCommonTime)
// 			var got app.EnergyStatusDemand
// 			database.GormDB.First(&got)
// 			fmt.Printf("%+v\n", got)
// 			if !strings.Contains(tt.want, got.AlarmStatus) {
// 				t.Errorf("DemandJudge() = %+v, want %s", got, tt.want)
// 			}
// 			database.GormDB.Delete(&got)
// 		})
// 	}
// }

// func TestInputContactJudge(t *testing.T) {
// 	type args struct {
// 		m                        MonitorConditions
// 		status                   InputContactAlertStatus
// 		inputContact             InputContact
// 		overAllCommonTime        int64
// 		sensorMonitorAlertStatus SensorMonitorAlertStatus
// 	}
// 	tests := []struct {
// 		name       string
// 		args       args
// 		want       bool
// 		wantStatus string
// 	}{
// 		{
// 			name: "on:" + "init(\"\")->" + strOccurrence,
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strInputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 0,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: InputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  1,
// 					Updated: 100,
// 				},
// 				inputContact: InputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strInputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 1,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strInputContact,
// 					StartTime:  0,
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       true,
// 			wantStatus: strOccurrence,
// 		},
// 		{
// 			name: "on:" + "init(\"\")->init(\"\")",
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strInputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 1,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: InputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  0,
// 					Updated: 100,
// 				},
// 				inputContact: InputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strInputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 1,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strInputContact,
// 					StartTime:  0,
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       false,
// 			wantStatus: "",
// 		},
// 		{
// 			name: "on:" + strRestoration + "->" + strOccurrence,
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strInputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 1,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: InputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  1,
// 					Updated: 100,
// 				},
// 				inputContact: InputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strInputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 1,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strInputContact,
// 					StartTime:  0,
// 					LastStatus: strRestoration,
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       true,
// 			wantStatus: strOccurrence,
// 		},
// 		{
// 			name: "on:" + strRestoration + "->" + strRestoration,
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strInputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 1,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: InputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  0,
// 					Updated: 100,
// 				},
// 				inputContact: InputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strInputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 1,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strInputContact,
// 					StartTime:  0,
// 					LastStatus: strRestoration,
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       false,
// 			wantStatus: strRestoration,
// 		},
// 		{
// 			name: "on:" + strOccurrence + "->" + strOccurrence,
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strInputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 1,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: InputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  1,
// 					Updated: 100,
// 				},
// 				inputContact: InputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strInputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 2,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strInputContact,
// 					StartTime:  0,
// 					LastStatus: strOccurrence,
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       true,
// 			wantStatus: strOccurrence,
// 		},
// 		{
// 			name: "on:" + strOccurrence + "->" + strRestoration,
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strInputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 1,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: InputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  0,
// 					Updated: 100,
// 				},
// 				inputContact: InputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strInputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 10,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strInputContact,
// 					StartTime:  0,
// 					LastStatus: strOccurrence,
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       false,
// 			wantStatus: strRestoration,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var alert SensorMonitorAlert
// 			var alertStatus SensorMonitorAlertStatus

// 			alertStatus.ID = tt.args.sensorMonitorAlertStatus.ID
// 			alertStatus.Category = tt.args.sensorMonitorAlertStatus.Category
// 			alertStatus.StartTime = 0
// 			alertStatus.LastStatus = tt.args.sensorMonitorAlertStatus.LastStatus
// 			database.Save(&alertStatus)

// 			if tt.args.m.JudgementAbnormalTime > 0 {
// 				// しきい値オーバー
// 				InputContactJudge(tt.args.m, tt.args.status, tt.args.inputContact, tt.args.overAllCommonTime)

// 				// 異常継続時間前
// 				InputContactJudge(tt.args.m, tt.args.status, tt.args.inputContact, tt.args.overAllCommonTime)
// 				result := database.GormDB.First(&alert)
// 				if result.Error == nil {
// 					// この時点ではアラート発生していないので、DBに無いのが正解
// 					t.Errorf("InputContactJudge() = %+v, JudgementAbnormalTime", alert)
// 				}
// 				result = database.GormDB.First(&alertStatus)
// 				if result.Error != nil {
// 					t.Errorf("InputContactJudge() = %+v, SensorMonitorAlertStatus", alertStatus)
// 				}
// 				if !strings.Contains(alertStatus.LastStatus, tt.args.sensorMonitorAlertStatus.LastStatus) {
// 					// この時点ではアラート発生していないので、初期値と同じでなければエラー
// 					t.Errorf("InputContactJudge() = %+v, JudgementAbnormalTime", alertStatus)
// 				}

// 				// 待機
// 				time.Sleep(time.Duration(tt.args.m.JudgementAbnormalTime+1) * time.Second)
// 			}

// 			// 異常継続時間後
// 			if got := InputContactJudge(tt.args.m, tt.args.status, tt.args.inputContact, tt.args.overAllCommonTime); got != tt.want {
// 				t.Errorf("InputContactJudge() = %v, want %v", got, tt.want)
// 			}

// 			// アラート情報判定
// 			result := database.GormDB.Where("id = ?", tt.args.inputContact.ID).First(&alertStatus)
// 			if result.Error != nil {
// 				t.Errorf("InputContactJudge() = %+v, %s", alertStatus, result.Error.Error())
// 			} else if !strings.Contains(alertStatus.LastStatus, tt.wantStatus) {
// 				fmt.Printf("%+v\n", alertStatus)
// 				t.Errorf("InputContactJudge() = %+v, wantStatus:%s", alertStatus, tt.wantStatus)
// 			} else if alertStatus.StartTime != 0 {
// 				t.Errorf("InputContactJudge() = %+v, wantStartTime:%s", alertStatus, "0")
// 			}

// 			database.GormDB.Delete(&SensorMonitorAlert{})
// 			database.GormDB.Delete(&SensorMonitorAlertStatus{})
// 		})
// 	}
// }

// func TestOutputContactJudge(t *testing.T) {
// 	type args struct {
// 		m                        MonitorConditions
// 		status                   OutputContactAlertStatus
// 		outputContact            OutputContact
// 		overAllCommonTime        int64
// 		sensorMonitorAlertStatus SensorMonitorAlertStatus
// 	}
// 	tests := []struct {
// 		name       string
// 		args       args
// 		want       bool
// 		wantStatus string
// 	}{
// 		{
// 			name: "on:" + "init(\"\")->" + strOccurrence,
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strOutputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 0,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: OutputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  1,
// 					Updated: 100,
// 				},
// 				outputContact: OutputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strOutputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 1,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strOutputContact,
// 					StartTime:  0,
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       true,
// 			wantStatus: strOccurrence,
// 		},
// 		{
// 			name: "on:" + "init(\"\")->init(\"\")",
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strOutputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 1,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: OutputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  0,
// 					Updated: 100,
// 				},
// 				outputContact: OutputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strOutputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 1,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strOutputContact,
// 					StartTime:  0,
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       false,
// 			wantStatus: "",
// 		},
// 		{
// 			name: "on:" + strRestoration + "->" + strOccurrence,
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strOutputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 1,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: OutputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  1,
// 					Updated: 100,
// 				},
// 				outputContact: OutputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strOutputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 1,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strOutputContact,
// 					StartTime:  0,
// 					LastStatus: strRestoration,
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       true,
// 			wantStatus: strOccurrence,
// 		},
// 		{
// 			name: "on:" + strRestoration + "->" + strRestoration,
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strOutputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 1,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: OutputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  0,
// 					Updated: 100,
// 				},
// 				outputContact: OutputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strOutputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 1,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strOutputContact,
// 					StartTime:  0,
// 					LastStatus: strRestoration,
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       false,
// 			wantStatus: strRestoration,
// 		},
// 		{
// 			name: "on:" + strOccurrence + "->" + strOccurrence,
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strOutputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 1,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: OutputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  1,
// 					Updated: 100,
// 				},
// 				outputContact: OutputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strOutputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 2,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strOutputContact,
// 					StartTime:  0,
// 					LastStatus: strOccurrence,
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       true,
// 			wantStatus: strOccurrence,
// 		},
// 		{
// 			name: "on:" + strOccurrence + "->" + strRestoration,
// 			args: args{
// 				m: MonitorConditions{
// 					ID:                    "qwertyuiop",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       strOutputContact,
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 1,
// 					Message:               "",
// 					Enable:                0,
// 				},
// 				status: OutputContactAlertStatus{
// 					ID:      "1234567890",
// 					Status:  0,
// 					Updated: 100,
// 				},
// 				outputContact: OutputContact{
// 					ID:             "1234567890",
// 					DeviceName:     "R1212",
// 					Category:       strOutputContact,
// 					ControlID:      0,
// 					ControlChannel: 0,
// 					Enable:         0,
// 				},
// 				overAllCommonTime: 10,
// 				sensorMonitorAlertStatus: SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   strOutputContact,
// 					StartTime:  0,
// 					LastStatus: strOccurrence,
// 					OccurredAt: 0,
// 				},
// 			},
// 			want:       false,
// 			wantStatus: strRestoration,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var alert SensorMonitorAlert
// 			var alertStatus SensorMonitorAlertStatus

// 			alertStatus.ID = tt.args.sensorMonitorAlertStatus.ID
// 			alertStatus.Category = tt.args.sensorMonitorAlertStatus.Category
// 			alertStatus.StartTime = 0
// 			alertStatus.LastStatus = tt.args.sensorMonitorAlertStatus.LastStatus
// 			database.Save(&alertStatus)

// 			if tt.args.m.JudgementAbnormalTime > 0 {
// 				// しきい値オーバー
// 				OutputContactJudge(tt.args.m, tt.args.status, tt.args.outputContact, tt.args.overAllCommonTime)

// 				// 異常継続時間前
// 				OutputContactJudge(tt.args.m, tt.args.status, tt.args.outputContact, tt.args.overAllCommonTime)
// 				result := database.GormDB.First(&alert)
// 				if result.Error == nil {
// 					// この時点ではアラート発生していないので、DBに無いのが正解
// 					t.Errorf("OutputContactJudge() = %+v, JudgementAbnormalTime", alert)
// 				}
// 				result = database.GormDB.First(&alertStatus)
// 				if result.Error != nil {
// 					t.Errorf("OutputContactJudge() = %+v, SensorMonitorAlertStatus", alertStatus)
// 				}
// 				if !strings.Contains(alertStatus.LastStatus, tt.args.sensorMonitorAlertStatus.LastStatus) {
// 					// この時点ではアラート発生していないので、初期値と同じでなければエラー
// 					t.Errorf("OutputContactJudge() = %+v, JudgementAbnormalTime", alertStatus)
// 				}

// 				// 待機
// 				time.Sleep(time.Duration(tt.args.m.JudgementAbnormalTime+1) * time.Second)
// 			}

// 			if got := OutputContactJudge(tt.args.m, tt.args.status, tt.args.outputContact, tt.args.overAllCommonTime); got != tt.want {
// 				t.Errorf("OutputContactJudge() = %v, want %v", got, tt.want)
// 			}

// 			result := database.GormDB.Where("id = ?", tt.args.outputContact.ID).First(&alertStatus)
// 			if result.Error != nil {
// 				t.Errorf("OutputContactJudge() = %+v, %s", alertStatus, result.Error.Error())
// 			} else if !strings.Contains(alertStatus.LastStatus, tt.wantStatus) {
// 				fmt.Printf("%+v\n", alertStatus)
// 				t.Errorf("OutputContactJudge() = %+v, wantStatus:%s", alertStatus, tt.wantStatus)
// 			}

// 			database.GormDB.Delete(&alert)
// 			database.GormDB.Delete(&alertStatus)
// 		})
// 	}
// }

func TestPowerOutageJudge(t *testing.T) {
	const otherStatus = "OTHER"
	testOutput := func(status string) string {
		return fmt.Sprintf("qwertyui\nSTATUS   : %s\n1234", status)
	}

	type args struct {
		oldStatus         string
		output            string
		overAllCommonTime int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: `"" -> ONBATT`,
			args: args{
				oldStatus: "",
				output:    testOutput(app.UPS_STATUS_BATTERY),
			},
			want: true,
		},
		{
			name: `"" -> ONLINE`,
			args: args{
				oldStatus: "",
				output:    testOutput(app.UPS_STATUS_ONLINE),
			},
			want: false,
		},
		{
			name: `"" -> OTHER`,
			args: args{
				oldStatus: "",
				output:    testOutput(otherStatus),
			},
			want: true,
		},
		{
			name: `ONLINE -> ONLINE`,
			args: args{
				oldStatus: app.UPS_STATUS_ONLINE,
				output:    testOutput(app.UPS_STATUS_ONLINE),
			},
			want: false,
		},
		{
			name: `ONBATT -> ONBATT`,
			args: args{
				oldStatus: app.UPS_STATUS_BATTERY,
				output:    testOutput(app.UPS_STATUS_BATTERY),
			},
			want: false,
		},
		{
			name: `OTHER -> OTHER`,
			args: args{
				oldStatus: otherStatus,
				output:    testOutput(otherStatus),
			},
			want: false,
		},
		{
			name: `ONBATT -> ONLINE`,
			args: args{
				oldStatus: app.UPS_STATUS_BATTERY,
				output:    testOutput(app.UPS_STATUS_ONLINE),
			},
			want: false,
		},
		{
			name: `ONBATT -> OTHER`,
			args: args{
				oldStatus: app.UPS_STATUS_BATTERY,
				output:    testOutput(otherStatus),
			},
			want: true,
		},
		{
			name: `ONLINE -> ONBATT`,
			args: args{
				oldStatus: app.UPS_STATUS_ONLINE,
				output:    testOutput(app.UPS_STATUS_BATTERY),
			},
			want: true,
		},
		{
			name: `ONLINE -> OTHER`,
			args: args{
				oldStatus: app.UPS_STATUS_ONLINE,
				output:    testOutput(otherStatus),
			},
			want: true,
		},
		{
			name: `OTHER -> ONLINE`,
			args: args{
				oldStatus: otherStatus,
				output:    testOutput(app.UPS_STATUS_ONLINE),
			},
			want: false,
		},
		{
			name: `OTHER -> ONBATT`,
			args: args{
				oldStatus: otherStatus,
				output:    testOutput(app.UPS_STATUS_BATTERY),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := app.PowerOutageJudge(tt.args.oldStatus, tt.args.output, tt.args.overAllCommonTime); got != tt.want {
				t.Errorf("PowerOutageJudge() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestBatteryLevelAlertJudge(t *testing.T) {
// 	type args struct {
// 		startStatus     string
// 		demandPulseUnit DemandPulseUnit
// 		demandStatus    DemandStatus
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		{
// 			"under:" + strRestoration + "->" + strOccurrence,
// 			args{
// 				strRestoration,
// 				DemandPulseUnit{
// 					ID:                     "12345667890",
// 					DeviceName:             "",
// 					ElectricEnergyPerPulse: 0,
// 					ControlID:              0,
// 					Threshold:              25,
// 					Enabled:                0,
// 					JudgementAbnormalTime:  0,
// 					Message:                "",
// 				},
// 				DemandStatus{
// 					PulseUnitCount: 1,
// 					Voltage:        2.925, // しきい値 = (3.6 - (3.6-2.7)*(1.0-float64(Threshold/100)))
// 				},
// 			},
// 			strOccurrence,
// 		},
// 		{
// 			"under:" + strRestoration + "->" + strRestoration,
// 			args{
// 				strRestoration,
// 				DemandPulseUnit{
// 					ID:                     "12345667890",
// 					DeviceName:             "",
// 					ElectricEnergyPerPulse: 0,
// 					ControlID:              0,
// 					Threshold:              25,
// 					Enabled:                0,
// 					JudgementAbnormalTime:  1,
// 					Message:                "",
// 				},
// 				DemandStatus{
// 					PulseUnitCount: 0,
// 					Voltage:        2.926, // しきい値 = (3.6 - (3.6-2.7)*(1.0-float64(Threshold/100)))
// 				},
// 			},
// 			strRestoration,
// 		},
// 		{
// 			"under:" + strOccurrence + "->" + strRestoration,
// 			args{
// 				strOccurrence,
// 				DemandPulseUnit{
// 					ID:                     "12345667890",
// 					DeviceName:             "",
// 					ElectricEnergyPerPulse: 0,
// 					ControlID:              0,
// 					Threshold:              25,
// 					Enabled:                0,
// 					JudgementAbnormalTime:  1,
// 					Message:                "",
// 				},
// 				DemandStatus{
// 					PulseUnitCount: 0,
// 					Voltage:        3.6, // しきい値 = (3.6 - (3.6-2.7)*(1.0-float64(Threshold/100)))
// 				},
// 			},
// 			strRestoration,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var alert BatteryLevelAlert
// 			var alertStatus SensorMonitorAlertStatus

// 			alertStatus.ID = tt.args.demandPulseUnit.ID
// 			alertStatus.Category = strDemandPulseUnit
// 			alertStatus.StartTime = 0
// 			alertStatus.LastStatus = tt.args.startStatus
// 			database.Save(&alertStatus)
// 			overAllCommonTime := time.Now().Unix()

// 			// しきい値オーバー
// 			BatteryLevelAlertJudge(tt.args.demandPulseUnit, tt.args.demandStatus, overAllCommonTime)

// 			// 異常継続時間前
// 			BatteryLevelAlertJudge(tt.args.demandPulseUnit, tt.args.demandStatus, overAllCommonTime)
// 			result := database.GormDB.First(&alert)
// 			if result.Error == nil {
// 				t.Errorf("BatteryLevelAlertJudge() = %+v, JudgementAbnormalTime", alert)
// 			}
// 			result = database.GormDB.First(&alertStatus)
// 			if result.Error != nil {
// 				t.Errorf("BatteryLevelAlertJudge() = %+v, SensorMonitorAlertStatus", alertStatus)
// 			}
// 			if alertStatus.LastStatus != tt.args.startStatus {
// 				t.Errorf("BatteryLevelAlertJudge() = %+v, JudgementAbnormalTime", alertStatus)
// 			}

// 			// 待機
// 			time.Sleep(time.Duration(tt.args.demandPulseUnit.JudgementAbnormalTime+1) * time.Second)

// 			// 異常継続時間後
// 			BatteryLevelAlertJudge(tt.args.demandPulseUnit, tt.args.demandStatus, overAllCommonTime)
// 			result = database.GormDB.First(&alert)
// 			if result.Error != nil {
// 				if tt.want != strRestoration {
// 					t.Errorf("BatteryLevelAlertJudge() = %s", result.Error.Error())
// 				}
// 			} else if strings.Contains(alert.kind, strDemandPulseUnit) {
// 				t.Errorf("BatteryLevelAlertJudge() = alert kind:%s", alert.kind)
// 			}

// 			result = database.GormDB.First(&alertStatus)
// 			if result.Error != nil {
// 				t.Errorf("BatteryLevelAlertJudge() = %+v, SensorMonitorAlertStatus", alertStatus)
// 			} else if alertStatus.LastStatus != tt.want {
// 				t.Errorf("BatteryLevelAlertJudge() = %+v, want:%s", alertStatus, tt.want)
// 			}

// 			database.GormDB.Delete(&alert)
// 			database.GormDB.Delete(&alertStatus)
// 		})
// 	}
// }

// func TestSaveDBCommunicationAlert(t *testing.T) {
// 	type args struct {
// 		id                string
// 		name              string
// 		overAllCommonTime int64
// 		kind              string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			SaveDBCommunicationAlert(tt.args.id, tt.args.name, tt.args.overAllCommonTime, tt.args.kind)
// 		})
// 	}
// }

// func Test_sensorAlertCreate(t *testing.T) {
// 	type args struct {
// 		id                string
// 		alertStatus       *SensorMonitorAlertStatus
// 		value             float64
// 		overAllCommonTime int64
// 		monitorid         string
// 		kind              string
// 	}
// 	tests := []struct {
// 		name      string
// 		args      args
// 		wantAlert SensorMonitorAlert
// 	}{}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if gotAlert := sensorAlertCreate(tt.args.id, tt.args.alertStatus, tt.args.value, tt.args.overAllCommonTime, tt.args.monitorid, tt.args.kind); !reflect.DeepEqual(gotAlert, tt.wantAlert) {
// 				t.Errorf("sensorAlertCreate() = %v, want %v", gotAlert, tt.wantAlert)
// 			}
// 		})
// 	}
// }

// func Test_EnergyJudge(t *testing.T) {
// 	type args struct {
// 		m                 MonitorConditions
// 		status            EnergyAlertStatusElectricPower
// 		energysensor      EnergySensor
// 		deviceid          string
// 		overAllCommonTime int64
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantRet bool
// 	}{}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if gotRet := EnergyJudge(tt.args.m, tt.args.status, tt.args.energysensor, tt.args.deviceid, tt.args.overAllCommonTime); gotRet != tt.wantRet {
// 				t.Errorf("EnergyJudge() = %v, want %v", gotRet, tt.wantRet)
// 			}
// 		})
// 	}
// }

// func Test_EnvironmentJudge(t *testing.T) {
// 	type args struct {
// 		m                 MonitorConditions
// 		status            EnvironmentSensorStatus
// 		sensor            EnvironmentSensor
// 		deviceid          string
// 		overAllCommonTime int64
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantRet bool
// 	}{}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if gotRet := EnvironmentJudge(tt.args.m, tt.args.status, tt.args.sensor, tt.args.deviceid, tt.args.overAllCommonTime); gotRet != tt.wantRet {
// 				t.Errorf("EnvironmentJudge() = %v, want %v", gotRet, tt.wantRet)
// 			}
// 		})
// 	}
// }

// func Test_underSensorValueJudge(t *testing.T) {
// 	type args struct {
// 		value                 float64
// 		threshold             float64
// 		alertStatus           *SensorMonitorAlertStatus
// 		JudgementAbnormalTime int64
// 	}
// 	tests := []struct {
// 		name        string
// 		args        args
// 		want        bool
// 		alertStatus *SensorMonitorAlertStatus
// 	}{
// 		{
// 			name: "OFF -> ON (starttime=0)",
// 			args: args{
// 				value:     0,
// 				threshold: 0.1,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  0,
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 				JudgementAbnormalTime: 1,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: ""},
// 		},
// 		{
// 			name: "OFF -> ON (now()-starttime < JudgementAbnormalTime)",
// 			args: args{
// 				value:     0,
// 				threshold: 0.01,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  time.Now().Unix(),
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 				JudgementAbnormalTime: 100,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: ""},
// 		},
// 		{
// 			name: "OFF -> ON (now()-starttime > JudgementAbnormalTime)",
// 			args: args{
// 				value:     0,
// 				threshold: 0.001,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  10,
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 				JudgementAbnormalTime: 0,
// 			},
// 			want:        true,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "OFF -> ON (JudgementAbnormalTime=0)",
// 			args: args{
// 				value:     0,
// 				threshold: 0.0001,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  0,
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 				JudgementAbnormalTime: 0,
// 			},
// 			want:        true,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "OFF -> ON (OccurredAt!=0)",
// 			args: args{
// 				value:     0,
// 				threshold: 0.0001,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  0,
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 1,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "OFF -> ON (OccurredAt!=0)",
// 			args: args{
// 				value:     0,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  time.Now().Unix(),
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 0,
// 			},
// 			want:        true,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "ON  -> OFF(starttime=0)",
// 			args: args{
// 				value:     0.1,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  0,
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 1,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "ON  -> OFF(now()-starttime < JudgementAbnormalTime)",
// 			args: args{
// 				value:     0.01,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  time.Now().Unix(),
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 100,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "ON  -> OFF(now()-starttime > JudgementAbnormalTime)",
// 			args: args{
// 				value:     0.001,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  10,
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 1,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strRestoration},
// 		},
// 		{
// 			name: "ON  -> OFF(JudgementAbnormalTime=0)",
// 			args: args{
// 				value:     0.0001,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  time.Now().Unix(),
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 0,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strRestoration},
// 		},
// 		{
// 			name: "ON  -> OFF(OccurredAt!=0)",
// 			args: args{
// 				value:     0.0001,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  0,
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 1,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "ON  -> OFF(OccurredAt!=0)",
// 			args: args{
// 				value:     0.0001,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  time.Now().Unix(),
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 0,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strRestoration},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := underSensorValueJudge(tt.args.value, tt.args.threshold, tt.args.alertStatus, tt.args.JudgementAbnormalTime); got != tt.want {
// 				t.Errorf("underSensorValueJudge() = %v, want %v", got, tt.want)
// 			} else {
// 				log.Printf("%+v\n", tt.args.alertStatus)
// 				if tt.alertStatus.LastStatus != tt.args.alertStatus.LastStatus {
// 					t.Errorf("underSensorValueJudge() = LastStatus error %+v %+v", tt.args.alertStatus, tt.alertStatus)
// 				}
// 			}
// 		})
// 	}
// }

// func Test_overSensorValueJudge(t *testing.T) {
// 	type args struct {
// 		value                 float64
// 		threshold             float64
// 		alertStatus           *SensorMonitorAlertStatus
// 		JudgementAbnormalTime int64
// 	}
// 	tests := []struct {
// 		name        string
// 		args        args
// 		want        bool
// 		alertStatus *SensorMonitorAlertStatus
// 	}{
// 		{
// 			name: "OFF -> ON (starttime=0)",
// 			args: args{
// 				value:     0.1,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  0,
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 				JudgementAbnormalTime: 1,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: ""},
// 		},
// 		{
// 			name: "OFF -> ON (now()-starttime < JudgementAbnormalTime)",
// 			args: args{
// 				value:     0.01,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  time.Now().Unix(),
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 				JudgementAbnormalTime: 100,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: ""},
// 		},
// 		{
// 			name: "OFF -> ON (now()-starttime > JudgementAbnormalTime)",
// 			args: args{
// 				value:     0.001,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  10,
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 				JudgementAbnormalTime: 0,
// 			},
// 			want:        true,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "OFF -> ON (JudgementAbnormalTime=0)",
// 			args: args{
// 				value:     0.0001,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  0,
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 				JudgementAbnormalTime: 0,
// 			},
// 			want:        true,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "OFF -> ON (OccurredAt!=0)",
// 			args: args{
// 				value:     0.0001,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  0,
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 1,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "OFF -> ON (OccurredAt!=0)",
// 			args: args{
// 				value:     0,
// 				threshold: 0,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  time.Now().Unix(),
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 0,
// 			},
// 			want:        true,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "ON  -> OFF(starttime=0)",
// 			args: args{
// 				value:     0,
// 				threshold: 0.1,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  0,
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 1,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "ON  -> OFF(now()-starttime < JudgementAbnormalTime)",
// 			args: args{
// 				value:     0,
// 				threshold: 0.01,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  time.Now().Unix(),
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 100,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "ON  -> OFF(now()-starttime > JudgementAbnormalTime)",
// 			args: args{
// 				value:     0,
// 				threshold: 0.001,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  10,
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 1,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strRestoration},
// 		},
// 		{
// 			name: "ON  -> OFF(JudgementAbnormalTime=0)",
// 			args: args{
// 				value:     0,
// 				threshold: 0.0001,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  time.Now().Unix(),
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 0,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strRestoration},
// 		},
// 		{
// 			name: "ON  -> OFF(OccurredAt!=0)",
// 			args: args{
// 				value:     0,
// 				threshold: 0.0001,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  0,
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 1,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strOccurrence},
// 		},
// 		{
// 			name: "ON  -> OFF(OccurredAt!=0)",
// 			args: args{
// 				value:     0,
// 				threshold: 0.0001,
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "",
// 					Category:   "",
// 					StartTime:  time.Now().Unix(),
// 					LastStatus: strOccurrence,
// 					OccurredAt: time.Now().Unix(),
// 				},
// 				JudgementAbnormalTime: 0,
// 			},
// 			want:        false,
// 			alertStatus: &SensorMonitorAlertStatus{LastStatus: strRestoration},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := overSensorValueJudge(tt.args.value, tt.args.threshold, tt.args.alertStatus, tt.args.JudgementAbnormalTime); got != tt.want {
// 				t.Errorf("overSensorValueJudge() = %v, want %v", got, tt.want)
// 			} else {
// 				log.Printf("%+v\n", tt.args.alertStatus)
// 				if tt.alertStatus.LastStatus != tt.args.alertStatus.LastStatus {
// 					t.Errorf("overSensorValueJudge() = LastStatus error %+v %+v", tt.args.alertStatus, tt.alertStatus)
// 				}
// 			}
// 		})
// 	}
// }

// func Test_sensorValueJudge(t *testing.T) {
// 	type args struct {
// 		id                string
// 		m                 MonitorConditions
// 		alertStatus       *SensorMonitorAlertStatus
// 		value             float64
// 		overAllCommonTime int64
// 		kind              string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want bool
// 	}{
// 		{
// 			name: "OFF -> ON(発報) -> ON(発報なし)",
// 			args: args{
// 				id: "",
// 				m: MonitorConditions{
// 					ID:                    "1234567890",
// 					EnergySensorID:        "",
// 					EnvironmentSensorID:   "",
// 					InputContactID:        "1234567890",
// 					OutputContactID:       "",
// 					MonitorCategory:       "",
// 					Threshold:             0,
// 					JudgementMethod:       strOn,
// 					JudgementAbnormalTime: 10,
// 					Message:               "test",
// 					Enable:                1,
// 				},
// 				alertStatus: &SensorMonitorAlertStatus{
// 					ID:         "1234567890",
// 					Category:   "",
// 					StartTime:  0,
// 					LastStatus: "",
// 					OccurredAt: 0,
// 				},
// 				value:             0,
// 				overAllCommonTime: time.Now().Unix(),
// 				kind:              strInputContact,
// 			},
// 			want: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			if got := sensorValueJudge(tt.args.id, tt.args.m, tt.args.alertStatus, tt.args.value, tt.args.overAllCommonTime, tt.args.kind); got != false {
// 				t.Errorf("sensorValueJudge() = %v, want %v", got, false)
// 			}
// 			if got := sensorValueJudge(tt.args.id, tt.args.m, tt.args.alertStatus, tt.args.value, tt.args.overAllCommonTime, tt.args.kind); got != false {
// 				t.Errorf("sensorValueJudge() = %v, want %v", got, false)
// 			}

// 			time.Sleep(time.Duration(tt.args.m.JudgementAbnormalTime))

// 			if got := sensorValueJudge(tt.args.id, tt.args.m, tt.args.alertStatus, tt.args.value, tt.args.overAllCommonTime, tt.args.kind); got != tt.want {
// 				t.Errorf("sensorValueJudge() = %v, want %v", got, tt.want)
// 			}

// 			if got := sensorValueJudge(tt.args.id, tt.args.m, tt.args.alertStatus, tt.args.value, tt.args.overAllCommonTime, tt.args.kind); got != false {
// 				t.Errorf("sensorValueJudge() = %v, want %v", got, false)
// 			}

// 			time.Sleep(time.Duration(tt.args.m.JudgementAbnormalTime))

// 			if got := sensorValueJudge(tt.args.id, tt.args.m, tt.args.alertStatus, tt.args.value, tt.args.overAllCommonTime, tt.args.kind); got != tt.want {
// 				t.Errorf("sensorValueJudge() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestShouldSendMailByTime(t *testing.T) {
	type args struct {
		Now           int64
		OccuredAt     int64
		SentMailAt    int64
		FirstInterval int64
		AfterInterval int64
	}

	const (
		hasNotSentMail = int64(0)
		fiveMinutes    = 60 * 5
		oneDay         = 60 * 60 * 24
	)

	var (
		baseTime20220110_1200 = time.Date(2022, 1, 10, 12, 0, 0, 0, time.Local).Unix()
		baseTime20220110_1203 = time.Date(2022, 1, 10, 12, 3, 0, 0, time.Local).Unix()
		baseTime20220110_1205 = time.Date(2022, 1, 10, 12, 5, 0, 0, time.Local).Unix()
		baseTime20220110_1210 = time.Date(2022, 1, 10, 12, 10, 0, 0, time.Local).Unix()
		baseTime20220111_1200 = time.Date(2022, 1, 11, 12, 0, 0, 0, time.Local).Unix()
		baseTime20220111_1205 = time.Date(2022, 1, 11, 12, 5, 0, 0, time.Local).Unix()
		baseTime20220111_1300 = time.Date(2022, 1, 11, 13, 5, 0, 0, time.Local).Unix()
	)

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "異常発生から経過時間0秒",
			args: args{
				Now:           baseTime20220110_1200,
				OccuredAt:     baseTime20220110_1200,
				SentMailAt:    hasNotSentMail,
				FirstInterval: fiveMinutes,
				AfterInterval: oneDay,
			},
			want: false,
		},
		{
			name: "異常発生から経過時間3分",
			args: args{
				Now:           baseTime20220110_1203,
				OccuredAt:     baseTime20220110_1200,
				SentMailAt:    hasNotSentMail,
				FirstInterval: fiveMinutes,
				AfterInterval: oneDay,
			},
			want: false,
		},
		{
			name: "異常発生から経過時間5分",
			args: args{
				Now:           baseTime20220110_1205,
				OccuredAt:     baseTime20220110_1200,
				SentMailAt:    hasNotSentMail,
				FirstInterval: fiveMinutes,
				AfterInterval: oneDay,
			},
			want: true,
		},
		{
			name: "異常発生から経過時間10分 (初回メール送信済)",
			args: args{
				Now:           baseTime20220110_1210,
				OccuredAt:     baseTime20220110_1200,
				SentMailAt:    baseTime20220110_1205,
				FirstInterval: fiveMinutes,
				AfterInterval: oneDay,
			},
			want: false,
		},
		{
			name: "異常発生から経過時間10分 (初回メール未送信)",
			args: args{
				Now:           baseTime20220110_1210,
				OccuredAt:     baseTime20220110_1200,
				SentMailAt:    hasNotSentMail,
				FirstInterval: fiveMinutes,
				AfterInterval: oneDay,
			},
			want: true,
		},
		{
			name: "異常発生から経過時間1日 (初回メール送信済)",
			args: args{
				Now:           baseTime20220111_1200,
				OccuredAt:     baseTime20220110_1200,
				SentMailAt:    baseTime20220110_1205,
				FirstInterval: fiveMinutes,
				AfterInterval: oneDay,
			},
			want: false,
		},
		{
			name: "異常発生から経過時間1日 (初回メール送信済)",
			args: args{
				Now:           baseTime20220111_1200,
				OccuredAt:     baseTime20220110_1200,
				SentMailAt:    baseTime20220110_1205,
				FirstInterval: fiveMinutes,
				AfterInterval: oneDay,
			},
			want: false,
		},
		{
			name: "異常発生から経過時間1日5分 (初回メール送信済)",
			args: args{
				Now:           baseTime20220111_1205,
				OccuredAt:     baseTime20220110_1200,
				SentMailAt:    baseTime20220110_1205,
				FirstInterval: fiveMinutes,
				AfterInterval: oneDay,
			},
			want: true,
		},
		{
			name: "異常発生から経過時間1日1時間 (2回目メール送信済)",
			args: args{
				Now:           baseTime20220111_1300,
				OccuredAt:     baseTime20220110_1200,
				SentMailAt:    baseTime20220111_1205,
				FirstInterval: fiveMinutes,
				AfterInterval: oneDay,
			},
			want: false,
		},
		{
			name: "異常発生から経過時間1日1時間 (2回目メール未送信)",
			args: args{
				Now:           baseTime20220111_1300,
				OccuredAt:     baseTime20220110_1200,
				SentMailAt:    baseTime20220110_1205,
				FirstInterval: fiveMinutes,
				AfterInterval: oneDay,
			},
			want: true,
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := app.ShouldSendMailByTime(tt.args.Now, tt.args.OccuredAt, tt.args.SentMailAt, tt.args.FirstInterval, tt.args.AfterInterval); got != tt.want {
				t.Errorf("shouldSendMailByTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
