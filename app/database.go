package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Database struct {
	GormDB *gorm.DB
}

const MicroFormat = "2006/01/02 15:04:05.000000"
const dbPath_A9E = "/vol_data/database/ecoRAMDAR.db"
const dbPath_G3L = "/ecoRAMDAR.db"

/*
制御状態の値
*/
const (
	// 制御状態の値
	StrControlAuto          = "AutoControl"       // 自動制御
	StrControlDemand        = "DemandControl"     // デマンド制御
	StrControlRemote        = "RemoteControl"     // 遠隔制御
	StrControlDefrost       = "Defrost"           // デフロスト制御
	StrControlRemoteDefrost = "RemoteDefrostStop" // 遠隔デフロスト制御
	StrControlRelease       = "Release"           // 解除

	// 監視カテゴリ
	strCurrent1      = "ElectricCurrent1" // 電流1
	strCurrent2      = "ElectricCurrent2" // 電流2
	strCurrent3      = "ElectricCurrent3" // 電流3
	strVoltage1      = "Voltage1"         // 電圧1
	strVoltage2      = "Voltage2"         // 電圧2
	strVoltage3      = "Voltage3"         // 電圧3
	strPowerFactor   = "PowerFactor"      // 力率
	strElectricPower = "ElectricPower"    // 電力
	strBreaker       = "Breaker"          // ブレーカー（入力監視）
	strControl       = "Control"          // ON,OFF制御（入出力監視）

	// 環境センサ
	strTemperature            = "Temperature"     // 温度
	strTemperatureAndHumidity = "Humidity"        // 湿度
	strDiscomfortIndex        = "DiscomfortIndex" // 不快指数

	// アラート判定
	strExceed = "More" // 上回り
	strBelow  = "Less" // 下回り
	strOn     = "ON"   // ON判定
	strOff    = "OFF"  // OFF判定

	// アラート発報/復旧
	StrExcess     = "Excess"  // 超過
	Strlimit      = "Limit"   // 限界
	StrBeVigilant = "Warn"    // 逼迫（警告）
	StrRecover    = "Recover" // 復帰
	StrNormal     = "Normal"  // 通常

	StrStatusExcess     = "超過"
	StrStatuslimit      = "限界"
	StrStatusBeVigilant = "警戒"
	StrStatusRecover    = "復帰"
	StrStatusNormal     = "通常"

	strOccurrence  = "Occurrence"  // 発生
	strRestoration = "Restoration" // 復旧

	strHeating = "Heating" // 暖房
	strCooling = "Cooling" // 冷房
	// strSolar   = "Solar"   // 太陽光
	strAuto = "Auto" // 自動（未使用）※クラウド側にて冷暖判断（冷房開始日などで判断）

	strIntermitten = "Intermitten" // 間欠
	strCapacity    = "Capacity"    // 容量
	strStop        = "Stop"        // 停止（間欠？）

	//クラウド送信type
	strchildIn       = "childIn"       // 入力（未使用）
	strchildOut      = "childOut"      // 出力
	strwatchalert    = "watchalert"    // アラート
	strdevicecontrol = "devicecontrol" // 制御
	strdemandPulse   = "demandPulse"   // デマンドパルスユニット

	// センサータイプ
	strEnergySensor        = "EnergySensor"        // エネルギーセンサー（電力）
	strEnvironmentalSensor = "EnvironmentalSensor" // 環境センサー
	strInputContact        = "InputContact"        // 入力接点
	strOutputContact       = "OutputContact"       // 出力接点
	strDemandPulseUnit     = "DemandPulse"         // デマンドパルス
	strChildUnit           = "ChildUnit"           // デマンドパルス

	// センサーカテゴリー
	strCategoryEnergySensor        = "エネルギーセンサー"
	strCategoryEnvironmentalSensor = "環境センサー"
	strCategoryInputContact        = "入力接点"
	strCategoryOutputContact       = "出力接点"
	strCategoryDemandPulseUnit     = "デマンドパルス"
	strCategoryChildUnit           = "子機"

	//クラウドデータ取得
	strMaster        = "Master"        // マスターデータ
	strManualControl = "ManualControl" // 手動制御命令
	strDefrost       = "Defrost"       // 手動デフロスト命令
	strReboot        = "Reboot"        // 親機再起動命令
	strDemand        = "Demand"        // デマンド制御命令
	strDummy         = "Dummy"         // チャンネルのmessageを使用しない場合用のダミーデータ
)

/*
バッテリーレベルのステータス
*/
const (
	VOLTAGE_OK int = 0
	VOLTAGE_NG int = 1
)

/*
UPSのアラート保持用の固定ID
*/
const (
	UPS_ALERT_STATUS_ID = "UPS"
)

const (
	UPS_STATUS_ONLINE      = "ONLINE"
	UPS_STATUS_BATTERY     = "ONBATT"
	UPS_STATUS_TRIM        = "TRIM"
	UPS_STATUS_BOOST       = "BOOST"
	UPS_STATUS_OVERLOAD    = "OVERLOAD"
	UPS_STATUS_LOWBATT     = "LOWBATT"
	UPS_STATUS_REPLACEBATT = "REPLACEBATT"
	UPS_STATUS_NOBATT      = "NOBATT"
	UPS_STATUS_COMMLOST    = "COMMLOST"
	UPS_STATUS_COMMERR     = "COMMERR"
)

/*
再起動コマンド保存用の固定ID
*/
const (
	REBOOT_COMMAND_ID = "REBOOT"
	DEMAND_CONTROL_ID = "DEMAND_CONTROL"
)

/*
更新周期
*/
const (
	//int64AllAggregationCycle = int64(30 * 1000) // 更新周期（子機通信）(msec)
	int64AllAggregationCycle     = int64(60 * 1000)    // 更新周期（子機通信）(msec)
	int64CommandAcquisitionCycle = int64(30 * 1000)    // 更新周期（未使用）(msec)
	demandTimeLimit              = float64(30)         // デマンド時限(分)
	demandSecondTimeLimit        = float64(1800)       // デマンド時限(秒)
	maxIntegration               = float64(4294967295) // デマンドmax値
)

/*
制御時温度設定
※この設定値を超えた場合は時間制御のみを行う
*/
const (
	maxLimitTemperature = float64(100)  // 温度上限
	minLimitTemperature = float64(-100) // 温度下限
)

const (
	COM_OK int = 0
	COM_NG int = 1
)

/*
null 許容型の構造体を宣言
Unmarshal で null が来たことがわかるようにするため

	type NullFloat64 struct {
		Float64 float64
		Valid   bool // Valid is true if Float64 is not NULL
	}

	Float64 = 数値
	Valid = true:float64格納, false:null格納
*/
type NullFloat64 struct {
	sql.NullFloat64
}

// GO -> JSON
func (s NullFloat64) MarshalJSON() ([]byte, error) {
	if s.Valid {
		return json.Marshal(s.Float64)
	}
	return []byte("null"), nil
}

// JSON -> GO
func (s *NullFloat64) UnmarshalJSON(data []byte) error {
	var v float64
	if string(data) == "null" {
		s.Valid = false
		return nil
	}
	if err := json.Unmarshal(data, &v); err != nil {
		s.Valid = false
		return err
	}
	s.Float64 = v
	s.Valid = true
	return nil
}

type BaseInformation struct {
	ID                        string  // 拠点識別子
	Name                      string  // 拠点名
	ContactElectricPower      float64 // 契約電力[kW]
	LimitElectricPower        float64 // 限界電力[kW]
	TargetElectricPower       float64 // 目標電力[kW]
	CancellationElectricPower float64 // 復帰電力[kW]
	InitialElectricPower      float64 // 初期電力[kW]
	SendGridAPIKey            string  // メール送信用APIキー
}

type ChildUnit struct {
	ID       string // 子機識別子
	Name     string // 子機名
	Wireless bool   // 子機有線・無線接続設定
	Enable   int8   // 子機有効設定
}

type DeviceInformation struct {
	ID                          string  // 機器識別子
	Name                        string  // 機器名
	Mode                        string  // 冷暖モード
	StopElectricPower           float64 // 稼働判断電力[W]
	RequiredControllingTime     int64   // 稼働必須時間[Sec]
	EnergySensorID              string  // エネルギーセンサー識別
	EnvironmentJudgmentCriteria string  // 環境センサー判断基準（未使用：複数必要なためDeviceEnvironmentInformationを使用）
	DefrostInputID              string  // デフロスト入力接点識別子
	IntermittenControlContactID string  // 制御接点（間欠）出力接点識別子
	CapacityControlContactID    string  // 制御接点（容量）出力接点識別子
	DefrostControlContactID     string  // デフロスト制御接点 出力接点識別子
}

type DeviceEnvironmentInformation struct {
	DeviceID string // 機器識別
	SensorID string // 環境センサー識別
}

type DemandPulseUnit struct {
	ID                     string  // デマンドパルス識別
	Name                   string  // デマンドパルス名
	DeviceName             string  // デバイス名
	ElectricEnergyPerPulse float64 // １パルスあたりの電力量
	ControlID              int16   // 接続先（ID）
	Threshold              float64 // 電池残量アラートのしきい値[％]
	Enabled                int8    // デマンドパルスの電池残量アラート通知 ON/OFF
	JudgementAbnormalTime  int64   // 異常継続時間（秒）
	Message                string  // メッセージ
}

type EnergySensor struct {
	ID                        string  // エネルギーセンサー識別
	Name                      string  // エネルギーセンサー名
	ChildDeviceID             string  // 接続子機ID
	DeviceName                string  // デバイス名
	Voltage                   float64 // 電圧[V]
	PowerFactor               float64 // 力率
	RatedPowerConsumptionCool float64 // 定格消費電力（冷房）[W]
	RatedPowerConsumptionWarm float64 // 定格消費電力（暖房）[W]
	ControlID                 int16   // 接続先ID
	ControlChannel            int16   // 接続先CH
	Enable                    int8    // 有効
}

type EnvironmentSensor struct {
	ID              string      // 環境センサー識別子
	Name            string      // 環境センサー名
	ChildDeviceID   string      // 接続子機ID
	DeviceName      string      // デバイス名
	Category        string      // 機器カテゴリ
	CorrectionValue NullFloat64 // 補正量[℃]
	CorrectionRatio NullFloat64 // 補正倍率
	ControlID       int16       // 接続先ID
	ControlChannel  int16       // 接続先CH
	Enable          int8        // 有効
}

type InputContact struct {
	ID             string // 入力接点識別
	Name           string // 入力接点名
	ChildDeviceID  string // 接続子機ID
	DeviceName     string // デバイス名
	Category       string // 入力カテゴリ
	ControlID      int16  // 接続先ID
	ControlChannel int16  // 接続先CH
	Enable         int8   // 有効
}

type OutputContact struct {
	ID             string // 出力接点識別
	Name           string // 出力接点名
	ChildDeviceID  string // 接続子機ID
	DeviceName     string // デバイス名
	Category       string // 出力カテゴリ
	ControlMethod  string // 制御方法
	ControlID      int16  // 接続先ID
	ControlChannel int16  // 接続先CH
	Enable         int8   // 有効
}

type UpdateCycle struct {
	ID            string // 機器識別子
	Master        int64  // マスターデータの同期頻度[Sec]
	ManualControl int64  // 手動制御命令の同期頻度[Sec]
	Defrost       int64  // 手動デフロスト命令の同期頻度[Sec]
	Reboot        int64  // 親機再起動の同期頻度[Sec]
	Demand        int64  // デマンド制御命令の同期頻度[Sec]
}

// アップデート情報
type FirmwareUpdateInfo struct {
	Name string `gorm:"primary_key"` // ファイル名
	URL  string // URL
}

// メール宛先
type UsersMailAddress struct {
	Email                 string `gorm:"primary_key"` // 宛先
	EnableBatteryAlert    int8   // 電池残量アラートをメール通知するかどうか
	EnableConnectionAlert int8   // 通信アラートをメール通知するかどうか
	EnableDemandAlert     int8   // デマンドアラートをメール通知するかどうか
	EnableSensorAlert     int8   // センサー監視アラートをメール通知するかどうか
	EnableSystemAlert     int8   // 親機システムアラートをメール通知するかどうか
	EnableUpsAlert        int8   // 親機電源アラートをメール通知するかどうか
}

/* Control Struct */
type ControlConditions struct {
	TargetDevicesID             string      `gorm:"primary_key"` // 機器識別
	Name                        string      // 制御名
	ControlTime                 int64       // 制御時間
	ReleaseControlTime          int64       // 制御解除時間
	ControlStartTemperature     NullFloat64 // 制御開始温度
	ControlStopTemperature      NullFloat64 // 制御開始温度
	ControlStartHumidity        NullFloat64 // 制御開始湿度
	ControlStopHumidity         NullFloat64 // 制御終了湿度
	ControlStartDiscomfortIndex NullFloat64 // 制御開始不快指数
	ControlStopDiscomfortIndex  NullFloat64 // 制御終了不快指数
	DemandStartTemperature      NullFloat64 // デマンド制御開始温度
	DemandStopTemperature       NullFloat64 // デマンド制御終了温度
	DemandStartHumidity         NullFloat64 // デマンド制御開始湿度
	DemandStopHumidity          NullFloat64 // デマンド制御終了湿度
	DemandStartDiscomfortIndex  NullFloat64 // デマンド制御開始不快指数
	DemandStopDiscomfortIndex   NullFloat64 // デマンド制御終了不快指数
	ControlEnable               int8        // 制御対象
	DemandEnable                int8        // デマンド制御対象
}

type MonitorConditions struct {
	ID                    string  // センサー監視識別
	Name                  string  // センサー監視条件名
	EnergySensorID        string  // エネルギーセンサー識別
	EnvironmentSensorID   string  // 環境センサー識別
	InputContactID        string  // 入力接点識別
	OutputContactID       string  // 出力接点識別
	MonitorCategory       string  // 監視対象カテゴリ
	Threshold             float64 // 閾値
	JudgementMethod       string  // 異常判断
	JudgementAbnormalTime int64   // 異常継続時間[秒]
	Message               string  // メッセージ
	Enable                int8    // 有効
}

type DemandControl struct {
	ID            string // 固定ID (DEMAND_CONTROL_ID)
	DemandControl int8   // デマンド制御命令
}

type RemoteControl struct {
	TargetDeviceId string `gorm:"primary_key"` // 出力接点識別
	ControlCommand int8   // 手動制御命令
}

type RemoteDefrostCommand struct {
	OutputContactID string `gorm:"primary_key"` // 出力接点識別
	StopCommand     int8   // 停止命令
}

type RebootCommand struct {
	ID     string // 固定ID (REBOOT_COMMAND_ID)
	Enable int8   // 再起動命令
}

/* Measurement Struct */
type DemandPulseCount struct {
	Time  int64 `gorm:"primary_key"` // 更新時刻
	Count int64 // パルスカウント数
}

type EnergyStatusDemand struct {
	Time                   int64   `gorm:"primary_key"` // 時限開始時刻
	CurrentElectricPower   float64 // 現在電力[kW]	// float64だとデータベースに保存するときに「Inf」になる
	PredictedElectricPower float64 // 予測電力[kW]	// float64だとデータベースに保存するときに「Inf」になる
	AdjustedElectricPower  float64 // 調整電力[kW]	// float64だとデータベースに保存するときに「Inf」になる
	//	CurrentElectricPower   float32 // 現在電力[kW]
	//	PredictedElectricPower float32 // 予測電力[kW]
	//	AdjustedElectricPower  float32 // 調整電力[kW]
	TimeLeft    int64  // 残り時間[Sec]
	AlarmStatus string // アラート状況
}

type EnergyAlertStatusDemand struct {
	Time                   int64   `gorm:"primary_key"` // 時限開始時刻
	CurrentElectricPower   float64 // 現在電力[kW]
	PredictedElectricPower float64 // 予測電力[kW]
	AdjustedElectricPower  float64 // 調整電力[kW]
	ContactElectricPower   float64 // 限界電力[kW]
	TargetElectricPower    float64 // 契約電力[kW]
	TimeLeft               int64   // 残り時間[Sec]
	AlarmStatus            string  // アラート状況
}

type PowerPerMinutes struct {
	Time             int64   `gorm:"primary_key"` // 取得時間
	PowerIntegration float64 // 現在積算値
	Power            float64 // 電力増加量
}

// エネルギーセンサーデータの登録用
type EnergyStatusElectricPower struct {
	//	EnergySensorID               string  `gorm:"primary_key"` // エネルギーセンサー識別子
	EnergySensorID               string  // エネルギーセンサー識別子
	Time                         int64   // 取得時間
	EffectivePower               float64 // 有効電力[W]
	EffectivePowerConsumption    float64 // 有効電力量[Wh]
	PreEffectivePowerConsumption float64 // 有効電力量[Wh]
	Current1                     float64 // 電流1[A]
	Current2                     float64 // 電流2[A]
	Current3                     float64 // 電流3[A]
	Voltage1                     float64 // 電圧1[V]
	Voltage2                     float64 // 電圧2[V]
	Voltage3                     float64 // 電圧3[V]
	Frequency                    float64 // 周波数[Hz]
	PowerFactor                  float64 // 力率
}

// エネルギーセンサー監視アラート用
type EnergyAlertStatusElectricPower struct {
	ID                           string  // エネルギーセンサー識別子
	EffectivePower               float64 // 有効電力[W]
	EffectivePowerConsumption    float64 // 有効電力量[Wh]
	PreEffectivePowerConsumption float64 // 有効電力量[Wh]
	Current1                     float64 // 電流1[A]
	Current2                     float64 // 電流2[A]
	Current3                     float64 // 電流3[A]
	Voltage1                     float64 // 電圧1[V]
	Voltage2                     float64 // 電圧2[V]
	Voltage3                     float64 // 電圧3[V]
	Frequency                    float64 // 周波数[Hz]
	PowerFactor                  float64 // 力率
	Updated                      int64   // 更新時間
}

// 入力接点データの登録用
type InputContactStatus struct {
	Time     int64  // 取得時間
	SensorID string // 入力接点識別子
	Status   int8   // 状態
}

// 入力接点監視アラート用
type InputContactAlertStatus struct {
	ID      string // 入力接点識別子
	Status  int8   // 状態
	Updated int64  // 取得時間
}

// 出力接点データの登録用
type OutputContactStatus struct {
	Time     int64  // 取得時間
	SensorID string // 出力接点識別子
	Status   int8   // 状態
}

// センサー監視条件アラート用
type OutputContactAlertStatus struct {
	ID      string // 出力接点識別子
	Status  int8   // 状態
	Updated int64  // 取得時間
}

// 環境センサーデータの登録用
type EnvironmentSensorStatus struct {
	Time            int64   // 取得時間
	SensorID        string  // 環境センサー識別子
	Temperature     float64 // 温度[℃]
	Humidity        float64 // 湿度[%]
	DiscomfortIndex float64 // 不快指数
}

// 環境センサー監視アラート用
type EnvironmentSensorAlertStatus struct {
	ID              string  // 環境センサー識別子
	Temperature     float64 // 温度[℃]
	Humidity        float64 // 湿度[%]
	DiscomfortIndex float64 // 不快指数
	Updated         int64   // 更新時間
}

type DeviceStatus struct {
	Time           int64   // 取得時間
	DeviceID       string  // 機器識別子
	Control        int8    // 制御命令
	Status         string  // 制御状況
	OccupancyRate  float64 // 稼働率
	ElectricPower  float64 // 電力(W)
	ElectricEnergy float64 // 電力量[Wh]
	Operating      bool    // 稼働状況
}

type SensorMonitorAlert struct {
	//	Time                int64   `gorm:"primary_key"` // アラート時間
	Time                int64   // アラート時間
	MonitorConditionsID string  `gorm:"primary_key"` // 監視条件識別子
	OccurredAt          int64   // 発生時間
	SensorID            string  // センサー識別子
	Status              string  // ステータス
	Value               float64 // 発生値
	Kind                string  // 種別
}

type CommunicationAlert struct {
	Time     int64  // アラート時間
	SensorID string // センサー識別子
	Kind     string // 種別
}

type BatteryLevelAlert struct {
	Time     int64  // アラート時間
	SensorID string // センサー識別子
	kind     string // 種別
}

type UpsAlert struct {
	Time    int64  // アラート時間
	Message string // メッセージ
}

type UpsAlertStatus struct {
	ID         string // ID、固定値 UPS
	Time       int64  // アラート時間
	Message    string // メッセージ
	LastStatus string // ステータス ONLINE or ONBATT or その他
}

type SystemAlert struct {
	Time      int64  // 発生時刻
	ErrorCode string // エラーコード
	Message   string // メッセージ
}

/*	現在値 	*/
type PresentValue struct {
	ID                      string  // 機器ID
	CurrentPower            float64 // 電力(W)
	CurrentPowerIntegration float64 // 電力量(Wh)
	CurrentTemperature      float64 // 温度
	CurrentHumidity         float64 // 湿度
	CurrentDiscomfortIndex  float64 // 不快指数
}

/*	制御状態 */
type ControlStatus struct {
	ID               string //機器ID
	ControlStartTime int64  // 制御開始時刻
	ControlEndTime   int64  // 制御終了時刻
	Status           string
	Control          int8
}

/* 遠隔制御解除(解除通知用) */
type RemoteControlReleaseStatus struct {
	ID string //機器ID
}

/*	デマンド逼迫内使用データ	*/
type DemandTightData struct {
	StartTime               int64   // 開始時間
	CurrentPowerIntegration float64 // 現在積算値
	StartIntegration        float64 // デマンド時限開始時積算値
	TargetPower             float64 // 目標電力
	PrevPowerIntegration    float64 // 前回積算値
}

// クラウド通信不可
type CloudCommonState struct {
	CommunicationError int
}

// 子機別センサー通信状況（0=正常、1=異常）
type ChildDeviceCommonState struct {
	ID                string // 子機ID
	InputContact      int    // 入力接点
	OutputContact     int    // 出力接点
	EnergySensor      int    // エネルギーセンサー
	EnvironmentSensor int    // 環境センサー
}

// デマンド逼迫内演算の戻り値
type ReturnValueCurrentPower struct {
	currentPower              float64 // 現在電力量
	leftTime                  float64 // 残り時間
	powerIncrease             float64 // 1分間の現在電力の増加電力量
	predictedPower            float64 // 予測電力量
	targetCurrentPower        float64 // 目標現在電力
	adjustedPower             float64 // 調整電力
	LimitElectricPower        float64 // 限界電力
	ContactElectricPower      float64 // 契約電力
	TargetElectricPower       float64 // 目標電力
	CancellationElectricPower float64 // 復帰電力
	InitialElectricPower      float64 // 初期電力
}

// デマンド計測用
type DemandStatus struct {
	PulseUnitCount uint32  // パルスカウント数
	Voltage        float64 // 電圧[V]
}

type DemandAlertStatus struct {
	ID     string // デマンドパルスID
	Time   int64  // アラート時間
	Status string
}

// センサ監視状態
type SensorMonitorAlertStatus struct {
	ID         string // 監視対象識別子
	Category   string // 監視対象カテゴリ
	StartTime  int64  // 異常開始時間
	LastStatus string // 前回ステータス
	OccurredAt int64  // 異常発生時刻
}

// 通信異常監視状態
type CommunicationMailAlertStatus struct {
	ID         string // 通信デバイス識別子(子機、センサー、接点、デマンドパルスのどれかのID)
	DeviceType string // デバイスタイプ
	Time       int64  // アラート時間（このアラートを保存した時間）
	LastStatus int    // ステータス (= 最新の通信状態)
	OccurredAt int64  // 異常発生時刻
	SentMailAt int64  // 最後にメールを送信する判断をした時刻
}

// 電池残量異常監視状態
type BatteryLevelMailAlertStatus struct {
	ID         string // デマンドパルスID
	Time       int64  // アラート時間（このアラートを保存した時間）
	LastStatus int    // ステータス (= 最新の通信状態)
	OccurredAt int64  // 異常発生時刻
	SentMailAt int64  // 最後にメールを送信する判断をした時刻
}

// アラートメール専用
type SensorMonitorMailAlert struct {
	//	Time                int64   `gorm:"primary_key"` // アラート時間
	Time                int64   // アラート時間
	MonitorConditionsID string  `gorm:"primary_key"` // 監視条件識別子
	OccurredAt          int64   // 発生時間
	SensorID            string  // センサー識別子
	Status              string  // ステータス
	Value               float64 // 発生値
	Kind                string  // 種別
	AlertMessage        string  // エラーメッセージ
}

type CommunicationMailAlert struct {
	Time       int64  // アラート時間
	SensorID   string // センサー識別子
	SensorName string // センサー名
	Kind       string // 種別
}

type BatteryLevelMailAlert struct {
	Time                  int64  // アラート時間
	SensorID              string // センサー識別子
	kind                  string // 種別
	JudgementAbnormalTime int64  // 異常継続時間
}

type UpsMailAlert struct {
	Time   int64  // アラート時間
	Status string // ステータス
}

type DemandStatusMailAlert struct {
	Time                   int64   `gorm:"primary_key"` // 時限開始時刻
	CurrentElectricPower   float64 // 現在電力[kW]
	PredictedElectricPower float64 // 予測電力[kW]
	AdjustedElectricPower  float64 // 調整電力[kW]
	TargetElectricPower    float64 // 目標電力[kW]
	LimitElectricPower     float64 // 限界電力[kW]
	ContractElectricPower  float64 // 契約電力[kW]
	AlarmStatus            string  // アラート状況
}

type TempANDCondition struct {
	ID          string  `gorm:"primary_key"` // 温度センサーの機器ID
	DeviceName1 string  // デバイス名
	DeviceName2 string  // デバイス名
	DeviceName3 string  // デバイス名
	DeviceName4 string  // デバイス名
	DeviceName5 string  // デバイス名
	DeviceName6 string  // デバイス名
	Number      int     // 温度センサーの数
	Temp1_ID    string  // 温度センサーのセンサーID
	Temp2_ID    string  // 温度センサーのセンサーID
	Temp3_ID    string  // 温度センサーのセンサーID
	Temp4_ID    string  // 温度センサーのセンサーID
	Temp5_ID    string  // 温度センサーのセンサーID
	Temp6_ID    string  // 温度センサーのセンサーID
	Temp1_Val   float64 // 温度センサー測定値
	Temp2_Val   float64 // 温度センサー測定値
	Temp3_Val   float64 // 温度センサー測定値
	Temp4_Val   float64 // 温度センサー測定値
	Temp5_Val   float64 // 温度センサー測定値
	Temp6_Val   float64 // 温度センサー測定値
	Temp1_Cnt   string  // 温度センサー制御条件ID
	Temp2_Cnt   string  // 温度センサー制御条件ID
	Temp3_Cnt   string  // 温度センサー制御条件ID
	Temp4_Cnt   string  // 温度センサー制御条件ID
	Temp5_Cnt   string  // 温度センサー制御条件ID
	Temp6_Cnt   string  // 温度センサー制御条件ID
}

type HumiANDCondition struct {
	ID          string  `gorm:"primary_key"` // 温湿度センサーの機器ID
	DeviceName1 string  // デバイス名
	DeviceName2 string  // デバイス名
	DeviceName3 string  // デバイス名
	DeviceName4 string  // デバイス名
	DeviceName5 string  // デバイス名
	DeviceName6 string  // デバイス名
	DeviceName7 string  // デバイス名
	DeviceName8 string  // デバイス名
	Number      int     // 温湿度センサーの数
	Humi1_ID    string  // 温湿度センサーのセンサーID
	Humi2_ID    string  // 温湿度センサーのセンサーID
	Humi3_ID    string  // 温湿度センサーのセンサーID
	Humi4_ID    string  // 温湿度センサーのセンサーID
	Humi5_ID    string  // 温湿度センサーのセンサーID
	Humi6_ID    string  // 温湿度センサーのセンサーID
	Humi7_ID    string  // 温湿度センサーのセンサーID
	Humi8_ID    string  // 温湿度センサーのセンサーID
	Humi1_Val   float64 // 温湿度センサー測定値
	Humi2_Val   float64 // 温湿度センサー測定値
	Humi3_Val   float64 // 温湿度センサー測定値
	Humi4_Val   float64 // 温湿度センサー測定値
	Humi5_Val   float64 // 温湿度センサー測定値
	Humi6_Val   float64 // 温湿度センサー測定値
	Humi7_Val   float64 // 温湿度センサー測定値
	Humi8_Val   float64 // 温湿度センサー測定値
	Humi1_Cnt   string  // 温湿度センサー制御条件ID
	Humi2_Cnt   string  // 温湿度センサー制御条件ID
	Humi3_Cnt   string  // 温湿度センサー制御条件ID
	Humi4_Cnt   string  // 温湿度センサー制御条件ID
	Humi5_Cnt   string  // 温湿度センサー制御条件ID
	Humi6_Cnt   string  // 温湿度センサー制御条件ID
	Humi7_Cnt   string  // 温湿度センサー制御条件ID
	Humi8_Cnt   string  // 温湿度センサー制御条件ID
}

var database *Database

// NewDatabase .
func NewDatabase() (*Database, error) {
	if database != nil {
		return nil, nil
	}

	dbPath := ""
	if MODEL == "A9E" {
		dbPath = dbPath_A9E
	} else {
		dbPath = dbPath_G3L
	}

	gormDB, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	// 内部保持用ステータス削除
	gormDB.DropTableIfExists(
		&EnergyStatusElectricPower{},
		&InputContactAlertStatus{},
		&OutputContactAlertStatus{},
		&EnergyAlertStatusElectricPower{},
		&EnvironmentSensorAlertStatus{},
		&UpsAlertStatus{},
		&ChildDeviceCommonState{},
		&CommunicationMailAlertStatus{},
		&BatteryLevelMailAlertStatus{},
		&SensorMonitorAlertStatus{},
		&TempANDCondition{}, // 温度センサーAND条件
		&HumiANDCondition{}, // 温湿度センサーAND条件

		&SensorMonitorAlert{},     // v0.43では必須。それ以降ではコメントアウトでもよい
		&SensorMonitorMailAlert{}, // v0.43では必須。それ以降ではコメントアウトでもよい
	)
	gormDB.AutoMigrate(
		// マスターデータ系(クラウドから取得)
		&BaseInformation{},
		&UpdateCycle{},
		&UsersMailAddress{},
		&FirmwareUpdateInfo{},
		&ChildUnit{},
		&DeviceInformation{},
		&DeviceEnvironmentInformation{},
		&DemandPulseUnit{},
		&EnergySensor{},
		&EnvironmentSensor{},
		&InputContact{},
		&OutputContact{},
		&ControlConditions{},
		&MonitorConditions{},

		// マスターデータ(更新周期系、クラウドから取得)
		&DemandControl{},
		&RemoteControl{},
		&RemoteDefrostCommand{},
		&RebootCommand{},

		// 内部専用データ
		&DemandPulseCount{},
		&EnergyStatusDemand{},
		&EnergyStatusElectricPower{},
		&InputContactStatus{},
		&OutputContactStatus{},
		&EnvironmentSensorStatus{},
		&DeviceStatus{},

		&PresentValue{},
		&ControlStatus{},
		&RemoteControlReleaseStatus{},
		&CloudCommonState{},
		&DemandTightData{},
		&PowerPerMinutes{},
		&ChildDeviceCommonState{},

		// アラート系
		&InputContactAlertStatus{},
		&OutputContactAlertStatus{},
		&CommunicationAlert{},
		&BatteryLevelAlert{},
		&UpsAlert{},

		// メールアラート系
		&SensorMonitorMailAlert{},
		&CommunicationMailAlert{},
		&DemandStatusMailAlert{},
		&BatteryLevelMailAlert{},
		&UpsMailAlert{},
		&SensorMonitorAlert{},
		&SystemAlert{},

		// アラートで使用するステータス
		&UpsAlertStatus{},
		&EnergyAlertStatusDemand{},
		&EnergyAlertStatusElectricPower{},
		&EnvironmentSensorAlertStatus{},
		&SensorMonitorAlertStatus{},
		&DemandAlertStatus{},

		&BatteryLevelMailAlertStatus{},
		&CommunicationMailAlertStatus{},

		// 温度センサーAND条件用
		&TempANDCondition{},
		&HumiANDCondition{},
	)

	database = &Database{
		GormDB: gormDB,
	}
	return database, nil
}

// Close .
func (db *Database) Close() {
	if database != nil {
		db.GormDB.Close()
		database = nil
	}
}

func deleteAllData(tx *gorm.DB) {
	models := []interface{}{
		&BaseInformation{},
		&UpdateCycle{},
		&UsersMailAddress{},
		&FirmwareUpdateInfo{},
		&ChildUnit{},
		&DeviceInformation{},
		&DeviceEnvironmentInformation{},
		&DemandPulseUnit{},
		&EnergySensor{},
		&EnvironmentSensor{},
		&InputContact{},
		&OutputContact{},
		&ControlConditions{},
		&MonitorConditions{},
	}

	for _, model := range models {
		if err := tx.Delete(model).Error; err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "%T delete DB : %s", model, err.Error())
		} else {
			Logger.Writef(LOG_LEVEL_DEBUG, "%T delete DB", model)
		}
	}
}

// SelectAll 引数で指定した models に DB から全件取得したデータを設定する
// エラーが発生した場合はログに記録し、エラーを返す
// ErrRecordNotFound の場合は nil を返す（データが存在しないことは正常な状態として扱う）
func (db *Database) SelectAll(models interface{}) error {
	if err := database.GormDB.Find(models).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Logger.Writef(LOG_LEVEL_DEBUG, "%T not found", models)
			return nil // レコードが見つからない場合は正常として扱う
		}
		Logger.Writef(LOG_LEVEL_ERR, "%T Read DB:%s", models, err.Error())
		return err
	}
	return nil
}

// SelectOne 引数で指定した model に DB から取得したデータを1件設定する
// エラーが発生した場合はログに記録し、エラーを返す
// ErrRecordNotFound の場合は nil を返す（データが存在しないことは正常な状態として扱う）
func (db *Database) SelectOne(model interface{}) error {
	if err := database.GormDB.Take(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Logger.Writef(LOG_LEVEL_DEBUG, "%T not found", model)
			return nil // レコードが見つからない場合は正常として扱う
		}
		Logger.Writef(LOG_LEVEL_ERR, "%T Read DB:%s", model, err.Error())
		return err
	}
	return nil
}

// SelectByQuery 引数で指定した model に DB から取得したデータを1件設定する
func (db *Database) SelectByQuery(model interface{}, query interface{}, args ...interface{}) {
	if err := database.GormDB.Where(query, args...).Take(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Logger.Writef(LOG_LEVEL_DEBUG, "%T not found", model)
		} else {
			Logger.Writef(LOG_LEVEL_ERR, "%T Read DB:%s", model, err.Error())
		}
	}
	return
}

// Save 引数で指定した model を DB に保存する
// エラーが発生した場合はログに記録し、エラーを返す
func (db *Database) Save(model interface{}) error {
	if err := db.GormDB.Save(model).Error; err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "%T Save DB:%s", model, err.Error())
		return err
	}
	return nil
}

// Delete 引数で指定した model を DB から削除する
// エラーが発生した場合はログに記録し、エラーを返す
func (db *Database) Delete(model, query interface{}, args ...interface{}) error {
	if err := database.GormDB.Where(query, args...).Delete(model).Error; err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "%T Delete DB:%s", model, err.Error())
		return err
	}
	return nil
}

// RemoteControlDeleteAndSaveDB 手動制御によるデバイス制御用のデータ削除・追加
func (db *Database) RemoteControlDeleteAndSaveDB(r []RemoteControl) {
	tx := database.GormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "RemoteControl Begin DB:%s", err.Error())
		return
	}
	if err := tx.Delete(&RemoteControl{}).Error; err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "RemoteControl Delete DB:%s", err.Error())
		tx.Rollback()
		return
	}
	for _, v := range r {
		if err := tx.Save(&v).Error; err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "RemoteControl Create DB:%s", err.Error())
			tx.Rollback()
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "RemoteControl Commit DB:%s", err.Error())
		tx.Rollback()
		return
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "RemoteControl Delete And Save DB")
}

/*
DemandPulseCount write setting data
*/
func (db *Database) DemandPulseCountCloudSaveDB(demandPulseCount DemandPulseCount) {
	tx := db.GormDB.Begin()
	tx.Delete(&demandPulseCount)
	err := tx.Save(&demandPulseCount).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "demandPulseCount Save DB:%+v", err.Error())
	}
	tx.Commit()
	// Logger.Writef(LOG_LEVEL_DEBUG, "demandPulseCount Cloud Save DB : %+v", demandPulseCount)
}

/*
EnergyStatusElectricPower write setting data
*/
func (db *Database) EnergyStatusElectricPowerSaveDB(energyStatusElectricPower EnergyStatusElectricPower) {
	var e EnergyAlertStatusElectricPower
	db.GormDB.Where("id = ?", energyStatusElectricPower.EnergySensorID).First(&e) // レコードが無い場合は0として考える
	energyStatusElectricPower.PreEffectivePowerConsumption = e.EffectivePowerConsumption
	err := db.GormDB.Save(&energyStatusElectricPower).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "EnergyStatusElectricPower Save DB:%+v", err.Error())
	}
	// Logger.Writef(LOG_LEVEL_DEBUG, "EnergyStatusElectricPower Cloud Save DB : %+v", energyStatusElectricPower)
}

/*
EnergyAlertStatusElectricPowerSaveDB write setting data
*/
func (db *Database) EnergyAlertStatusElectricPowerSaveDB(energyAlertStatusElectricPower EnergyAlertStatusElectricPower) {
	var e EnergyAlertStatusElectricPower
	db.GormDB.Where("id = ?", energyAlertStatusElectricPower.ID).First(&e) // レコードが無い場合は0として考える
	energyAlertStatusElectricPower.PreEffectivePowerConsumption = e.EffectivePowerConsumption
	err := db.GormDB.Save(&energyAlertStatusElectricPower).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "energyAlertStatusElectricPower Save DB:%+v", err.Error())
	}
	// Logger.Writef(LOG_LEVEL_DEBUG, "energyAlertStatusElectricPower Cloud Save DB : %+v", energyAlertStatusElectricPower)
}

/*
PresentValueEnvironmentCloudSaveDB write setting data
*/
func (db *Database) PresentValueEnvironmentCloudSaveDB(v EnvironmentSensor, s EnvironmentSensorStatus) {
	// デバイスインフォメーションのidを取得
	var deviceEnvironmentList []DeviceEnvironmentInformation
	err := database.GormDB.Where("sensor_id = ?", v.ID).Find(&deviceEnvironmentList).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "deviceEnvironmentList Cloud Read DB:%s", err.Error())
		return
	}

	for _, data := range deviceEnvironmentList {
		deviceid := data.DeviceID
		var value PresentValue
		result := database.GormDB.Where("id = ?", deviceid).First(&value)
		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "record not found") {
				// DB上に無い場合もあるためログしない
				// Logger.Writef(LOG_LEVEL_ERR, "PresentValue Read EnergyJudge DB:%s", result.Error.Error())
			} else {
				Logger.Writef(LOG_LEVEL_ERR, "PresentValue Read environment DB:%s", result.Error.Error())
			}
		} else {
			//			Logger.Writef(LOG_LEVEL_DEBUG, "PresentValue Read DB:%+v", result)
		}
		value.ID = deviceid
		// value.SensorType = v.Category
		value.CurrentTemperature = s.Temperature
		value.CurrentHumidity = s.Humidity
		value.CurrentDiscomfortIndex = s.DiscomfortIndex
		err = database.GormDB.Save(&value).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "PresentValue Save environment DB:%s", err.Error())
		} else {
			//Logger.Writef(LOG_LEVEL_DEBUG, "PresentValue Save environment DB:%+v", value)
		}
	}
}

/*
PresentValueEnergyCloudSaveDB write setting data
*/
func (db *Database) PresentValueEnergyCloudSaveDB(v EnergySensor, s EnergyAlertStatusElectricPower) {

	var deviceInformation DeviceInformation
	err := database.GormDB.Where("energy_sensor_id = ?", v.ID).First(&deviceInformation).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "deviceInformation Cloud Read DB:id=%s, %s", v.ID, err.Error())
	} else {
		deviceid := deviceInformation.ID
		var value PresentValue
		result := database.GormDB.Where("id = ?", deviceid).First(&value)
		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "record not found") {
				// DB上に無い場合もあるためログしない
				// Logger.Writef(LOG_LEVEL_ERR, "PresentValue Read energy DB:%s", result.Error.Error())
			} else {
				Logger.Writef(LOG_LEVEL_ERR, "PresentValue Read energy DB:%s", result.Error.Error())
			}
		} else {
			//			Logger.Writef(LOG_LEVEL_DEBUG, "PresentValue Read DB:%+v", result)
		}
		value.ID = deviceid
		value.CurrentPower = s.EffectivePower
		value.CurrentPowerIntegration = s.EffectivePowerConsumption

		// value.SensorType = strEnergySensor
		err = database.GormDB.Save(&value).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "PresentValue Save energy DB:%s", err.Error())
		} else {
			//			Logger.Writef(LOG_LEVEL_DEBUG, "PresentValue Save energy DB:%+v", value)
		}
	}
}

/*
CommunicationAlert write setting data
*/
func (db *Database) CommunicationAlertCloudSaveDB(overAllCommonTime int64, communicationAlert CommunicationAlert) {
	latestStatus := db.CommunicationMailAlertStatusReadDB(communicationAlert.SensorID)
	if latestStatus == nil {
		Logger.Write(LOG_LEVEL_DEBUG, "CommunicationMailAlertStatusReadDB: not found")
		return
	}

	// 通信異常検知時刻から5分後に送信する
	shouldSendAlert := func(now, occuredAt int64) bool {
		diff := now - occuredAt
		var fiveMinutes int64 = 5 * 60
		if diff > fiveMinutes {
			return true
		}
		return false
	}

	if shouldSendAlert(overAllCommonTime, latestStatus.OccurredAt) {
		if err := db.GormDB.Save(&communicationAlert).Error; err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "communicationAlert Save DB:%+v", err.Error())
		} else {
			//			Logger.Writef(LOG_LEVEL_DEBUG, "CommunicationAlert Cloud Save DB:%+v", communicationAlert)
		}
	}
}

/*
cloud common error write setting data
*/
func (db *Database) CloudCommonStateSaveDB(cloudCommonState CloudCommonState) {
	tx := db.GormDB.Begin()
	err := tx.Delete(&cloudCommonState).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "cloudComState DB Delete:%s", err.Error())
		tx.Rollback()
	} else {
		err := tx.Save(&cloudCommonState).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "cloudCommonState Save DB: %+v", err.Error())
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}
	// Logger.Writef(LOG_LEVEL_DEBUG, "CloudCommonState Cloud Save DB : %+v", cloudCommonState)
}

/*
デマンド逼迫内使用データ write setting data
*/
func (db *Database) DemandTightDataSaveDB(demandTightData DemandTightData) {
	tx := db.GormDB.Begin()
	tx.Delete(&demandTightData)
	err := tx.Save(&demandTightData).Error
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "demandTightData Save DB:%+v", err.Error())
	}
	tx.Commit()
	// Logger.Writef(LOG_LEVEL_DEBUG, "demandTightData Cloud Save DB : %+v", demandTightData)
}

// FirmwareUpdateInfoReadDB 設定ファームウェア情報のデータ読込
func (db *Database) FirmwareUpdateInfoReadDB() (f FirmwareUpdateInfo) {
	if err := database.GormDB.First(&f).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Logger.Writef(LOG_LEVEL_DEBUG, "%T not found", f)
		} else {
			Logger.Writef(LOG_LEVEL_ERR, "%T Read DB:%s", f, err.Error())
		}
	}
	return
}

// CommunicationMailAlertStatusReadDB 最新の通信状態保持用データ読込
func (db *Database) CommunicationMailAlertStatusReadDB(communicationDeviceID string) *CommunicationMailAlertStatus {
	var status CommunicationMailAlertStatus
	err := db.GormDB.Where("id = ?", communicationDeviceID).Take(&status).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Logger.Writef(LOG_LEVEL_DEBUG, "%T not found", status)
		} else {
			Logger.Writef(LOG_LEVEL_ERR, "%T Read DB:%s", status, err.Error())
		}
		return nil
	}
	//	Logger.Writef(LOG_LEVEL_DEBUG, "CommunicationMailAlertStatus Read DB:%+v", status)
	return &status
}

// BatteryLevelMailAlertStatusReadDB 最新の電圧状態の異常ステータス保持用データ読込
func (db *Database) BatteryLevelMailAlertStatusReadDB(demandPulseID string) *BatteryLevelMailAlertStatus {
	var status BatteryLevelMailAlertStatus
	err := db.GormDB.Where("id = ?", demandPulseID).Take(&status).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Logger.Writef(LOG_LEVEL_DEBUG, "%T not found", status)
		} else {
			Logger.Writef(LOG_LEVEL_ERR, "%T Read DB:%s", status, err.Error())
		}
		return nil
	}
	//	Logger.Writef(LOG_LEVEL_DEBUG, "BatteryLevelMailAlertStatus Read DB:%+v", status)
	return &status
}

// CommunicationMailAlertSaveDB 通信異常時メールアラート用のデータ保存
func (db *Database) CommunicationMailAlertSaveDB(communicationMailAlert CommunicationMailAlert) {
	latestStatus := db.CommunicationMailAlertStatusReadDB(communicationMailAlert.SensorID)
	if latestStatus == nil {
		Logger.Write(LOG_LEVEL_DEBUG, "CommunicationMailAlertCloudSaveDB: not found")
		return
	}

	// 通信異常検知時刻から5分後、それ以降は24時間経過毎に送信する
	shouldSendMail := ShouldSendCommunicationMailAlert(communicationMailAlert.Time, latestStatus.OccurredAt, latestStatus.SentMailAt)
	if shouldSendMail {
		err := db.GormDB.Save(&communicationMailAlert).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "communicationMailAlert Save DB:%s", err.Error())
		} else {
			//			Logger.Writef(LOG_LEVEL_DEBUG, "communicationMailAlert Save DB:%+v", communicationMailAlert)
		}

		latestStatus.SentMailAt = communicationMailAlert.Time
		database.Save(latestStatus)
	}
}

// BatteryLevelMailAlertSaveDB 電圧異常時メールアラート用のデータ保存
func (db *Database) BatteryLevelMailAlertSaveDB(batteryLevelMailAlert BatteryLevelMailAlert) {
	latestStatus := db.BatteryLevelMailAlertStatusReadDB(batteryLevelMailAlert.SensorID)
	if latestStatus == nil {
		Logger.Write(LOG_LEVEL_CRIT, "BatteryLevelMailAlertStatusReadDB: not found")
		return
	}

	// 通信異常検知時刻から異常継続時間経過後、それ以降は24時間経過毎に送信する
	shouldSendMail := ShouldSendBatteryLevelMailAlert(batteryLevelMailAlert.Time, latestStatus.OccurredAt, latestStatus.SentMailAt, batteryLevelMailAlert.JudgementAbnormalTime)
	if shouldSendMail {
		err := db.GormDB.Save(&batteryLevelMailAlert).Error
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "batteryLevelMailAlert Save DB:%s", err.Error())
		} else {
			Logger.Writef(LOG_LEVEL_DEBUG, "batteryLevelMailAlert Save DB:%+v", batteryLevelMailAlert)
		}

		latestStatus.SentMailAt = batteryLevelMailAlert.Time
		database.Save(latestStatus)
	}
}
