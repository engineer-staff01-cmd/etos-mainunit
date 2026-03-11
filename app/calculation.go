package app

import (
	"time"
)

// 計算関連の定数定義
const (
	// WattsPerKilowatt ワットからキロワットへの変換係数
	WattsPerKilowatt = 1000

	// DemandPeriodConversionFactor デマンド時限の変換係数（60分/30分）
	// 30分間のデータを60分間の電力に変換するための係数
	DemandPeriodConversionFactor = 2

	// SecondsPerMinute 分から秒への変換係数
	SecondsPerMinute = 60
)

/*
エネルギーセンサ電力量
*/
func CalcElectricEnergy(activePower float64) (val float64) {
	val = activePower / 3600
	return val
}

var StartIntegration float64 = 0.0
var start_time float64 = 0.0

/*
デマンド逼迫用演算（デマンド時限開始からの積算値）
*/
func CalcDemandPower(arg *DemandTightData, baseInformation BaseInformation, overAllCommonTime int64, ElectricEnergyPerPulse float64) (ret ReturnValueCurrentPower) {
	// 経過時間
	var elapsedTime float64

	tm := time.Unix(overAllCommonTime, 0)
	min := tm.Minute()
	//sec := tm.Second()
	// 初回時は直近の開始時限に時刻を設定する
	if arg.StartTime == 0 {
		arg.StartIntegration = arg.CurrentPowerIntegration
		arg.PrevPowerIntegration = arg.CurrentPowerIntegration
		if 0 <= min && min < 30 {
			arg.StartTime = time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), 0, 0, 0, time.Local).Unix()
		} else {
			arg.StartTime = time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), 30, 0, 0, time.Local).Unix()
		}
	}

	//Logger.Writef(LOG_LEVEL_DEBUG, "デマンド時限経過時間：%s\n", time.Unix(arg.StartTime, 0).Format(MicroFormat))
	elapsedTime = float64((overAllCommonTime - arg.StartTime) / 60)
	Logger.Writef(LOG_LEVEL_DEBUG, "デマンド時限経過時間 elapsedTime：%f", elapsedTime)
	elapsedSecondTime := float64(overAllCommonTime - arg.StartTime)
	Logger.Writef(LOG_LEVEL_DEBUG, "デマンド時限経過時間 elapsedSecondTime%f", elapsedSecondTime)
	// 現在電力
	if demandTimeLimit <= elapsedTime {
		arg.StartIntegration = arg.PrevPowerIntegration
		//親機が稼働停止していて、30分以上経過していることも考慮する
		if 0 <= min && min < 30 {
			arg.StartTime = time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), 0, 0, 0, time.Local).Unix()
		} else {
			arg.StartTime = time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), 30, 0, 0, time.Local).Unix()
		}
		elapsedTime = 0
		elapsedSecondTime = 0
	}

	if arg.StartIntegration <= arg.CurrentPowerIntegration {
		ret.currentPower = arg.CurrentPowerIntegration - arg.StartIntegration
	} else { //0クリア時
		//ret.currentPower = (maxIntegration - arg.CurrentPowerIntegration) + arg.StartIntegration
		ret.currentPower = maxIntegration - arg.StartIntegration + arg.CurrentPowerIntegration // issue #44の修正
	}
	// 現在電力(W)＝パルス数 × 1パルスあたりの電力量(Wh）× 2(60分/30分)
	// kWにするために1000で割る
	ret.currentPower = (ret.currentPower * ElectricEnergyPerPulse * DemandPeriodConversionFactor) / WattsPerKilowatt
	arg.PrevPowerIntegration = arg.CurrentPowerIntegration
	/*
		// 1分間の現在電力の増加電力量
		var ppm PowerPerMinutes

		result := database.GormDB.First(&ppm)
		if result.Error != nil {
			Logger.Writef(LOG_LEVEL_ERR, "CalcDemandPower :%s", result.Error)
		}

		if sec < 10 {
			if ppm.PowerIntegration == 0 {
				ret.powerIncrease = 0
				ppm.Power = 0
			} else {
				ret.powerIncrease = arg.CurrentPowerIntegration - ppm.PowerIntegration // パルス数
				ppm.Power = ret.powerIncrease
			}
			ppm.PowerIntegration = arg.CurrentPowerIntegration
			ppm.Time = overAllCommonTime

			tx := database.GormDB.Begin()
			tx.Delete(PowerPerMinutes{})
			tx.Save(&ppm)
			Logger.Writef(LOG_LEVEL_DEBUG, "ppm:%+v", ppm)
			tx.Commit()
		} else {
			ret.powerIncrease = ppm.Power
		}		/*  */

	// 1分間の電力増加量を傾きの平均として算出
	if elapsedTime == 0 || elapsedTime == 30 {
		ret.powerIncrease = 0
		//StartIntegration = arg.CurrentPowerIntegration
		StartIntegration = ret.currentPower
		start_time = 0
	} else if elapsedTime > 0 {
		if StartIntegration == 0 {
			//StartIntegration = arg.CurrentPowerIntegration
			StartIntegration = ret.currentPower
			start_time = elapsedTime
			ret.powerIncrease = 0
		} else {
			//ret.powerIncrease = (arg.CurrentPowerIntegration - StartIntegration) / (elapsedTime - start_time)
			if elapsedTime-start_time == 0 {
				ret.powerIncrease = 0
			} else {
				ret.powerIncrease = (ret.currentPower - StartIntegration) / (elapsedTime - start_time)
			}
		}
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "1分間の電力増加量 : %.2f, 現在積算値 : %.2f, デマンド時限開始時積算値 : %.2f, 経過時間 : %.2f, 開始時間 : %.2f", ret.powerIncrease, arg.CurrentPowerIntegration, StartIntegration, elapsedTime, start_time)

	// パルス数に1パルス当たりの電力量をかけ、W ⇒ kWに変換
	//ret.powerIncrease = ret.powerIncrease * ElectricEnergyPerPulse / 1000

	// 残り時間
	mLeftTime := demandTimeLimit - elapsedTime
	ret.leftTime = demandSecondTimeLimit - elapsedSecondTime
	Logger.Writef(LOG_LEVEL_DEBUG, "mLeftTime:%.2f ret.leftTime%.2f", mLeftTime, ret.leftTime)
	// 予測電力量
	ret.predictedPower = ret.currentPower + (ret.powerIncrease * mLeftTime)
	Logger.Writef(LOG_LEVEL_DEBUG, "予測電力量 : %.2f, 現在電力量 : %.2f, 1分間の電力増加量 : %.2f, 残り時間 : %.2f", ret.predictedPower, ret.currentPower, ret.powerIncrease, mLeftTime)
	// 目標現在電力
	ret.targetCurrentPower =
		((baseInformation.TargetElectricPower-baseInformation.InitialElectricPower)/demandTimeLimit)*elapsedTime +
			baseInformation.InitialElectricPower
	Logger.Writef(LOG_LEVEL_DEBUG, "目標現在電力 : %.2f, 目標電力 : %.2f, 初期電力 : %.2f, 経過時間 : %.2f", ret.targetCurrentPower, baseInformation.TargetElectricPower, baseInformation.InitialElectricPower, elapsedTime)

	// 調整電力
	//ret.adjustedPower = (ret.predictedPower - arg.TargetPower) / mLeftTime * demandTimeLimit
	ret.adjustedPower = ret.currentPower - ret.targetCurrentPower
	Logger.Writef(LOG_LEVEL_DEBUG, "調整電力 : %.2f, 現在量 : %.2f, 現在目標電力 : %.2f", ret.adjustedPower, ret.currentPower, ret.targetCurrentPower)

	// 契約電力
	ret.ContactElectricPower = baseInformation.ContactElectricPower
	// 限界電力
	ret.LimitElectricPower = baseInformation.LimitElectricPower
	// 目標電力
	ret.TargetElectricPower = baseInformation.TargetElectricPower
	// 復帰電力
	ret.CancellationElectricPower = baseInformation.CancellationElectricPower
	// 初期電力
	ret.InitialElectricPower = baseInformation.InitialElectricPower

	return ret
}

/*
稼働率
引数：電力(W),定格消費電力
戻り値：稼働率
*/
func CalcOccupancyRate(ElectricPower float64, RatedPowerConsumption float64) (val float64) {
	if RatedPowerConsumption == 0 {
		return 0
	}
	val = ElectricPower / RatedPowerConsumption
	return val
}

/*
不快指数
引数：温度、湿度
戻り値：不快指数
*/
func CalcCurrentDiscomfortIndex(temperature float64, humidity float64) (val float64) {
	if humidity == 0 {
		return 0
	}
	val = 0.81*temperature + 0.01*humidity*(0.99*temperature-14.3) + 46.3
	return val
}
