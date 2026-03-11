package app

import (
	"os"
	"sync"
	"time"
)

type Watchdog struct {
	fp           *os.File
	lastClearTime time.Time
	timeout      time.Duration
	mu           sync.Mutex
}

// Start ウォッチドッグスタート
func (wd *Watchdog) Start() {
	Logger.Write(LOG_LEVEL_DEBUG, "Start Watchdog")
	var err error
	wd.fp, err = os.OpenFile("/dev/watchdog", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		Logger.Write(LOG_LEVEL_ERR, "Start Watchdog: "+err.Error())
	}
	// デフォルトのタイムアウトは60秒
	wd.timeout = 60 * time.Second
	wd.lastClearTime = time.Now()
}

// Clear ウォッチドッグクリア
func (wd *Watchdog) Clear() {
	wd.mu.Lock()
	wd.lastClearTime = time.Now()
	wd.mu.Unlock()
	_, err := wd.fp.Write([]byte{0})
	if err != nil {
		Logger.Write(LOG_LEVEL_ERR, "Clear Watchdog: "+err.Error())
	}
}

// GetRemainingTime ウォッチドッグタイマーの残り時間を取得
func (wd *Watchdog) GetRemainingTime() time.Duration {
	wd.mu.Lock()
	defer wd.mu.Unlock()
	elapsed := time.Since(wd.lastClearTime)
	remaining := wd.timeout - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// WaitUntilTimeout ウォッチドッグタイマーが0になるまで待機し、残り時間をログに出力
func (wd *Watchdog) WaitUntilTimeout() {
	Logger.Writef(LOG_LEVEL_INFO, "Waiting for watchdog timer to expire...: %.1f seconds", wd.timeout.Seconds())
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		remaining := wd.GetRemainingTime()
		if remaining <= 0 {
			Logger.Writef(LOG_LEVEL_INFO, "Watchdog timer expired (remaining: 0s)")
			break
		}
		// Logger.Writef(LOG_LEVEL_INFO, "Watchdog timer remaining: %.1f seconds", remaining.Seconds())
		<-ticker.C
	}
}

// End ウォッチドッグ終了
func (wd *Watchdog) End() {
	Logger.Write(LOG_LEVEL_DEBUG, "End Watchdog")
	if _, err := wd.fp.Write([]byte("V")); err != nil {
		// Make sure the file descriptor is closed even if Magic Close fails.
		Logger.Write(LOG_LEVEL_ERR, "End Watchdog: "+err.Error())
		_ = wd.fp.Close()
		return
	}
	_ = wd.fp.Close()
}
