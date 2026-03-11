package app

//mainスレッド
//スレッド起動

import (
	"embed"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// message type
const (
	End         uint32 = iota //message true or false(string)
	DataGet                   //messageは種別名
	Alert                     //messageは種別名
	Notify                    //messageは種別名
	FirstCommon               //message true or false(string)
	ErrorEnd
)

const (
	Defrost uint32 = iota
	Control
)

// 通信順番
const (
	PHASE_WAIT int = iota
	PHASE_CHILD_SEND
	PHASE_CHILD_WAIT
	PHASE_DEMAND_SEND
	PHASE_DEMAND_WAIT
	PHASE_WATCH_SEND
	PHASE_WATCH_WAIT
	PHASE_CONTROL_SEND
	PHASE_CONTROL_WAIT
	PHASE_END
)

const (
	// CycleMinSecond 更新サイクルの最小周期(60s)
	CycleMinSecond = 60
	// CycleMaxSecond 更新サイクルの最大周期(1h)
	CycleMaxSecond = 60 * 60 * 1
)

// チャネルバッファサイズの定数定義
// 各スレッドのメッセージ頻度と処理時間に基づいて設定
const (
	// CloudChannelBufferSize クラウド通信チャネルのバッファサイズ
	// 頻繁な通信と複数のデータタイプを扱うため大きめに設定
	CloudChannelBufferSizeTo   = 64 // main -> cloud
	CloudChannelBufferSizeFrom = 32 // cloud -> main

	// ChildUnitChannelBufferSize 子機通信チャネルのバッファサイズ
	// 複数の子機からのデータを扱うため中程度に設定
	ChildUnitChannelBufferSize = 32 // main <-> child

	// DemandPulseChannelBufferSize デマンドパルスチャネルのバッファサイズ
	// 定期的なデータ取得のため中程度に設定
	DemandPulseChannelBufferSize = 32 // main <-> demand

	// WatchAlertChannelBufferSize 監視アラートチャネルのバッファサイズ
	// アラートが同時発生する可能性があるため中程度に設定
	WatchAlertChannelBufferSize = 32 // main <-> watch

	// DeviceControlChannelBufferSize デバイス制御チャネルのバッファサイズ
	// 複数デバイスの制御コマンドを扱うため中程度に設定
	DeviceControlChannelBufferSize = 32 // main <-> control
	DeviceControlOutBufferSize     = 32 // control -> child

	// MailChannelBufferSize メール送信チャネルのバッファサイズ
	// アラートが大量発生する可能性があるため中程度に設定
	MailChannelBufferSize = 32 // watch -> mail

	// TimerChannelBufferSize タイマーチャネルのバッファサイズ
	// 高頻度のタイマーイベントのため小さめに設定（バックプレッシャーを防ぐ）
	TimerChannelBufferSize = 10 // timer -> main
)

// 時間関連の定数定義
const (
	// MillisecondsPerSecond 秒をミリ秒に変換する係数
	MillisecondsPerSecond = 1000

	// TimerIntervalSeconds タイマーの基本間隔（秒）
	TimerIntervalOneSecond   = 1  // 1秒
	TimerIntervalTenSeconds  = 10 // 10秒
	TimerIntervalThirtySeconds = 30 // 30秒

	// LoopTimeoutMilliseconds ループのタイムアウト値（ミリ秒）
	// SendDeviceControlFlagの待機ループの最大実行時間
	LoopTimeoutMilliseconds = 1000

	// SleepIntervalMilliseconds ループ内のスリープ間隔（ミリ秒）
	SleepIntervalMilliseconds = 1

	// ChannelUsageWarningThreshold チャネル使用率の警告閾値（パーセント）
	// この値を超えると警告ログを出力
	ChannelUsageWarningThreshold = 80.0
)

type ChannelMessage struct {
	messageType uint32
	message     string
	time        time.Time
}

type ChannelControl struct {
	ControlType string
	ID          string
	DeviceStop  int8
	time        int64
}

// Version情報
var VERSION = ""
var REVISION = ""
var ECORAMDAR_VERSION = VERSION + "(" + REVISION + ")"

// API接続先情報
var ENV = Production

// モデル情報
var MODEL = ""

var DeviceControlFlag bool = false
var SendDeviceControlFlag bool = false

//var RebootCommandFlag bool = false

func Run(fs_public embed.FS) int {
	return RunCustom(fs_public)
}

// GetHwSerial CPUのシリアル番号を取得
// 環境変数 AT_SERIAL_NUMBER から取得する
func GetHwSerial() (sn string, err error) {
	if MODEL == "A9E" {
		sn = os.Getenv("AT_SERIAL_NUMBER")
		if sn == "" {
			return "", errors.New("environment variable AT_SERIAL_NUMBER is not set")
		}
	} else if MODEL == "G3L" {
		out, err := exec.Command("grep", "Serial", "/proc/cpuinfo").Output()
		if err != nil {
			return "", errors.New("failed to get hardware serial number")
		}
		arr := strings.Split(string(out), " ")
		sn = arr[len(arr)-1]
		sn = strings.TrimRight(sn, "\n")
		sn = strings.TrimRight(sn, "\r")
	}

	return sn, nil
}

func RunCustom(fs_public embed.FS) int {
	//env, err := GetEnv(ENV)
	env, err := ReadEnv()
	if err != nil {
		panic("Env is unknown")
	}
	Logger.Writef(LOG_LEVEL_INFO, "ENV:%s", ENV)
	Logger.Writef(LOG_LEVEL_INFO, "ENV:%+v", env)

	Logger.Writef(LOG_LEVEL_INFO, "VERSION:%s", ECORAMDAR_VERSION)
	serialNumber, err := GetHwSerial()
	if err != nil {
		Logger.Write(LOG_LEVEL_CRIT, err.Error())
		panic("failed to get hardware serial number")
	}
	Logger.Writef(LOG_LEVEL_INFO, "HW SERIAL NUMBER:%s", serialNumber)

	cloud := NewCloud(serialNumber)            // クラウドスレッド
	child := NewChild()                        // 子機スレッド
	demand := NewDemandPulseUnit()             // デマンドパルスレッド
	watch := NewWatchAlert()                   // 監視
	device := NewDeviceControl()               // デバイス制御
	mail, err := NewMailThread(env, fs_public) // メール送信スレッド
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to NewMailThread() %s", err.Error())
	}

	db, err := NewDatabase() // DBの初期化
	if err != nil {
		Logger.Writef(LOG_LEVEL_CRIT, "Failed to NewDatabase() %s", err.Error())
		panic("failed to initialize database")
	}

	var wg sync.WaitGroup
	cloud.Run(env, &wg) // クラウド通信スレの実行

	startTime := time.Now()
	commonTime := time.Now()

	// 再起動コマンドを一度クリアする
	rebooted := RebootCommand{}
	if err := db.SelectOne(&rebooted); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select reboot command: %s", err.Error())
	}
	if i82b(rebooted.Enable) {
		Logger.Writef(LOG_LEVEL_INFO, "Detected reboot command. This device should have been rebooted.")
		cloud.to <- ChannelMessage{Notify, strReboot, startTime}
		rebootCommand := RebootCommand{
			ID:     REBOOT_COMMAND_ID,
			Enable: b2i8(false),
		}
		if err := db.Save(&rebootCommand); err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to save reboot command: %s", err.Error())
		}
		//		RebootCommandFlag = true
	}

	cloud.to <- ChannelMessage{DataGet, strMaster, startTime}

loop:
	for {
		chCloudFlag := <-cloud.from
		switch chCloudFlag.messageType {
		case FirstCommon:
			break loop
		case End:
			cloud.to <- ChannelMessage{DataGet, strMaster, startTime}
		}
	}
	if err := loadFirmware(); err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "FWUpdateFlag: %t, Failed to loadFirmware : %s", FWUpdateFlag, err.Error())
	}

	child.Run(&wg, device.out, cloud.to) // 子機スレッドの実行
	demand.Run(&wg, cloud.to)            // デマンドパルス通信スレの実行
	watch.Run(&wg, cloud.to, mail.to)    // 監視スレッドの実行
	device.Run(&wg, cloud.to)            // 制御スレの実行
	mail.Run(&wg)                        // メール送信スレッドの実行

	et := NewExactlyTimer()

	t1s := et.NewTicker(TimerIntervalOneSecond * MillisecondsPerSecond) // 1秒おきに通知
	//	t30s := et.NewTicker(TimerIntervalThirtySeconds * MillisecondsPerSecond) // 30秒時限で開始
	t30s := et.NewTicker(int64AllAggregationCycle) // 集約サイクルで開始
	tMasters := et.NewTicker(TimerIntervalTenSeconds * MillisecondsPerSecond)
	tDemands := et.NewTicker(TimerIntervalTenSeconds * MillisecondsPerSecond)
	tRemotes := et.NewTicker(TimerIntervalTenSeconds * MillisecondsPerSecond)
	tDefrostes := et.NewTicker(TimerIntervalTenSeconds * MillisecondsPerSecond)
	tReboots := et.NewTicker(TimerIntervalTenSeconds * MillisecondsPerSecond)

	cycle := UpdateCycle{}
	if err := db.SelectOne(&cycle); err != nil {
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to select update cycle: %s", err.Error())
	}

	strPhase := []string{
		"PHASE_WAIT",
		"PHASE_CHILD_SEND",
		"PHASE_CHILD_WAIT",
		"PHASE_DEMAND_SEND",
		"PHASE_DEMAND_WAIT",
		"PHASE_WATCH_SEND",
		"PHASE_WATCH_WAIT",
		"PHASE_CONTROL_SEND",
		"PHASE_CONTROL_WAIT",
		"PHASE_END",
	}

	requestPhase := PHASE_END + 1

	// ウォッチドッグスタート
	wd := Watchdog{}
	wd.Start()
	// reboot_wait_time := 60 * time.Second

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	et.Run()

	rebootCmd := false

mainloop:
	for {
		select {
		case q := <-quit:
			Logger.Writef(LOG_LEVEL_INFO, "Received shutdown signal:%v", q)
			wd.End()
			break mainloop

		case <-t1s.C:
			wd.Clear()
			t1s.Flush()

		case t := <-t30s.C:
			//			var prevPhase = PHASE_WAIT
			if requestPhase <= PHASE_END {
				Logger.Writef(LOG_LEVEL_CRIT, "Data collection did not complete within %d seconds(%s)", int64AllAggregationCycle/MillisecondsPerSecond, strPhase[requestPhase])
				//				prevPhase = requestPhase
				requestPhase = PHASE_WAIT
				DeviceControlFlag = false
			}
			t30s.Reset(int64AllAggregationCycle)
			commonTime = t
			requestPhase = PHASE_CHILD_SEND
			//			if prevPhase != PHASE_WAIT {
			//				requestPhase = prevPhase + 1
			//			} else {
			//				requestPhase = PHASE_CHILD_SEND
			//			}
			t30s.Flush()

		case t := <-tMasters.C:
			Logger.Writef(LOG_LEVEL_DEBUG, "==================== Cloud tMasters Start ====================")
			msg := ChannelMessage{DataGet, strMaster, t}
			SendChannelMessageSafely(cloud.to, msg, false)
			c := UpdateCycle{}
			if err := db.SelectOne(&c); err != nil {
				Logger.Writef(LOG_LEVEL_WARNING, "Failed to select update cycle: %s", err.Error())
			}
			if c.ID != "" {
				//タイマー更新
				tMasters.Reset(Clamp(CycleMinSecond, CycleMaxSecond, cycle.Master) * MillisecondsPerSecond)

				// タイマーが更新されていれば一度イベントを発行する
				if cycle.Demand != c.Demand {
					tDemands.Reset(1)
				}
				if cycle.ManualControl != c.ManualControl {
					tRemotes.Reset(1)
				}
				if cycle.Defrost != c.Defrost {
					tDefrostes.Reset(1)
				}
				if cycle.Reboot != c.Reboot {
					tReboots.Reset(1)
				}
				// タイマー値更新
				cycle = c
			}
			tMasters.Flush()
		case t := <-tDemands.C:
			Logger.Writef(LOG_LEVEL_DEBUG, "==================== Cloud tDemands Start ====================")
			tDemands.Reset(Clamp(CycleMinSecond, CycleMaxSecond, cycle.Demand) * MillisecondsPerSecond)
			msg := ChannelMessage{DataGet, strDemand, t}
			SendChannelMessageSafely(cloud.to, msg, false)
			tDemands.Flush()
		case t := <-tRemotes.C:
			Logger.Writef(LOG_LEVEL_DEBUG, "==================== Cloud tRemotes Start ====================")
			tRemotes.Reset(Clamp(CycleMinSecond, CycleMaxSecond, cycle.ManualControl) * MillisecondsPerSecond)
			msg := ChannelMessage{DataGet, strManualControl, t}
			SendChannelMessageSafely(cloud.to, msg, false)
			tRemotes.Flush()
		case t := <-tDefrostes.C:
			Logger.Writef(LOG_LEVEL_DEBUG, "==================== Cloud tDefrostes Start ====================")
			tDefrostes.Reset(Clamp(CycleMinSecond, CycleMaxSecond, cycle.Defrost) * MillisecondsPerSecond)
			msg := ChannelMessage{DataGet, strDefrost, t}
			SendChannelMessageSafely(cloud.to, msg, false)
			tDefrostes.Flush()
		case t := <-tReboots.C:
			Logger.Writef(LOG_LEVEL_DEBUG, "==================== Cloud tReboots Start ====================")
			tReboots.Reset(Clamp(CycleMinSecond, CycleMaxSecond, cycle.Reboot) * MillisecondsPerSecond)
			// マニュアル操作 -> 主装置再起動の設定値をチェック
			rebootCommand := RebootCommand{}
			if err := db.SelectOne(&rebootCommand); err != nil {
				Logger.Writef(LOG_LEVEL_WARNING, "Failed to select reboot command: %s", err.Error())
			}
			if i82b(rebootCommand.Enable) {
				rebootCmd = true
				// ウォッチドッグによる再起動を行うためあえて wd.End() は行わない
				if FWUpdateFlag {
					Logger.Writef(LOG_LEVEL_INFO, "FWUpdateFlag is true. Rebooting...")
					wd.End()
				}
				break mainloop
			}
			msg := ChannelMessage{DataGet, strReboot, t}
			SendChannelMessageSafely(cloud.to, msg, true)
			tReboots.Flush()
		default:
			var ch_msg ChannelMessage

			switch requestPhase {
			case PHASE_CHILD_SEND:
				// 子機データ取得タイミング
				Logger.Writef(LOG_LEVEL_DEBUG, "==================== ChildUit Start ====================\n")
				SendChannelMessageSafely(child.to, ChannelMessage{DataGet, "true", commonTime}, true)
				requestPhase++

			case PHASE_CHILD_WAIT:
				// 子機データ取得end
				select {
				case chmsg := <-child.from:
					if chmsg.messageType == End {
						requestPhase++
					}
				default:
				}

			case PHASE_DEMAND_SEND:
				// デマンドパルスユニット監視タイミング
				Logger.Writef(LOG_LEVEL_DEBUG, "==================== DemandPulse Start ====================\n")
				SendChannelMessageSafely(demand.to, ChannelMessage{DataGet, "true", commonTime}, true)
				requestPhase++

			case PHASE_DEMAND_WAIT:
				// デマンドパルス取得end
				select {
				case chmsg := <-demand.from:
					if chmsg.messageType == End {
						requestPhase++
					}
				default:
				}

			case PHASE_WATCH_SEND:
				// 監視タイミング
				Logger.Writef(LOG_LEVEL_DEBUG, "==================== WatchAlert Start:%s ====================\n", commonTime.Format(MicroFormat))
				SendChannelMessageSafely(watch.to, ChannelMessage{DataGet, "true", commonTime}, true)
				requestPhase++

			case PHASE_WATCH_WAIT:
				// 監視終了
				select {
				case chmsg := <-watch.from:
					if chmsg.messageType == End {
						requestPhase++
					}
				default:
				}

			case PHASE_CONTROL_SEND:
				// 制御タイミング
				Logger.Writef(LOG_LEVEL_DEBUG, "==================== Control Start ====================\n")
				DeviceControlFlag = false
				SendChannelMessageSafely(device.to, ChannelMessage{DataGet, "true", commonTime}, true)
				requestPhase++

			case PHASE_CONTROL_WAIT:
				// 制御終了
				if DeviceControlFlag == true {
					select {
					case ch_msg = <-device.from:
						//Logger.Writef(LOG_LEVEL_DEBUG, "==================== Control End(main) : %+v", chmsg)
						if ch_msg.messageType == End || ch_msg.messageType == ErrorEnd {
							msg := ChannelMessage{Alert, strdevicecontrol, ch_msg.time}
							SendChannelMessageSafely(cloud.to, msg, false)
							Logger.Writef(LOG_LEVEL_DEBUG, "==================== Send Device Control msg : %+v", msg)
							DeviceControlFlag = false
							requestPhase++
						}
					default:
					}
				}

			case PHASE_END:
				Logger.Writef(LOG_LEVEL_DEBUG, "==================== Flag Replase:%dms ====================\n", time.Since(commonTime).Milliseconds())
				i := 0
				for {
					wd.Clear()
					if SendDeviceControlFlag == false || i >= LoopTimeoutMilliseconds {
						Logger.Writef(LOG_LEVEL_DEBUG, "==================== SendDeviceControlFlag loop end : %dms \n", i)
						break
						/*} else if i == 100 {
						msg := ChannelMessage{Alert, strdevicecontrol, ch_msg.time}
						SendChannelMessageSafely(cloud.to, msg, false)
						Logger.Writef(LOG_LEVEL_DEBUG, "==================== Send Device Control msg(2nd) : %+v", msg)	/*  */
					}
					i++
					time.Sleep(SleepIntervalMilliseconds * time.Millisecond)
				}

				requestPhase++

			}
		}
	}

	// 以下正常終了処理
	et.Stop()
	msg := ChannelMessage{End, strDummy, time.Time{}}
	SendChannelMessageSafely(cloud.to, msg, false)
	SendChannelMessageSafely(child.to, msg, false)
	SendChannelMessageSafely(demand.to, msg, false)
	SendChannelMessageSafely(watch.to, msg, false)
	SendChannelMessageSafely(device.to, msg, false)
	SendChannelMessageSafely(mail.to, msg, false)
	Logger.Writef(LOG_LEVEL_INFO, "Waiting shutdown threads...")
	wg.Wait()

	db.Close()
	Logger.Writef(LOG_LEVEL_INFO, "Shutdown successfully. elapsed_second:%f", time.Since(startTime).Seconds())
	// mainloop を抜ける前に待機し、再起動準備の猶予を確保
	if FWUpdateFlag == false && rebootCmd == true {
		Logger.Writef(LOG_LEVEL_INFO, "FWUpdateFlag is false. Waiting for watchdog timer to expire...")
		wd.WaitUntilTimeout()
		return -1
	}
	return 0
}

func SendChannelMessageSafely(channel chan ChannelMessage, msg ChannelMessage, flush bool) {
	/* flush channel */
	if flush {
	Loop:
		for {
			select {
			case <-channel:
			default:
				break Loop
			}
		}
	}

	select {
	case channel <- msg:
		// 送信成功
		// チャネル使用率が高い場合に警告（デバッグ用）
		channelLen := len(channel)
		channelCap := cap(channel)
		if channelCap > 0 {
			usagePercent := float64(channelLen) / float64(channelCap) * 100
			if usagePercent > ChannelUsageWarningThreshold {
				Logger.Writef(LOG_LEVEL_DEBUG, "Channel usage high: %d/%d (%.1f%%)", channelLen, channelCap, usagePercent)
			}
		}
	default:
		// チャネルが満杯で送信失敗（非ブロッキング）
		channelCap := cap(channel)
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to send ChannelMessage: channel is full (capacity=%d), messageType=%d, message=%s", channelCap, msg.messageType, msg.message)
	}
}

func SendChannelControlSafely(channel chan ChannelControl, msg ChannelControl, flush bool) {
	/* flush channel */
	if flush {
	Loop:
		for {
			select {
			case <-channel:
			default:
				break Loop
			}
		}
	}

	select {
	case channel <- msg:
		// 送信成功
		// チャネル使用率が高い場合に警告（デバッグ用）
		channelLen := len(channel)
		channelCap := cap(channel)
		if channelCap > 0 {
			usagePercent := float64(channelLen) / float64(channelCap) * 100
			if usagePercent > ChannelUsageWarningThreshold {
				Logger.Writef(LOG_LEVEL_DEBUG, "ChannelControl usage high: %d/%d (%.1f%%)", channelLen, channelCap, usagePercent)
			}
		}
	default:
		// チャネルが満杯で送信失敗（非ブロッキング）
		channelCap := cap(channel)
		Logger.Writef(LOG_LEVEL_WARNING, "Failed to send ChannelControl: channel is full (capacity=%d), ControlType=%s, ID=%s", channelCap, msg.ControlType, msg.ID)
	}
}
