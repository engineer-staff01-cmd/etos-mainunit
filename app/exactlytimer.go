package app

import (
	"sync"
	"time"
)

type Ticker struct {
	duration int64
	nextTime int64
	C        chan time.Time
	mu       sync.Mutex
}

type ExactlyTimer struct {
	ticker []*Ticker
	mu     sync.Mutex
	end    bool
}

func NewExactlyTimer() *ExactlyTimer {
	return new(ExactlyTimer)
}

func (et *ExactlyTimer) NewTicker(duration int64) *Ticker {
	ticker := new(Ticker)
	ticker.C = make(chan time.Time, TimerChannelBufferSize)
	ticker.duration = duration

	et.mu.Lock()
	et.ticker = append(et.ticker, ticker)
	et.mu.Unlock()
	return ticker
}

func (et *ExactlyTimer) Run() {
	Logger.Writef(LOG_LEVEL_DEBUG, "exactly timer start")

	go func() {
		et.Tick()
	}()
}

func (et *ExactlyTimer) Tick() {

	var errorSendMillsecond int64
	for {
		var nextDiffTime int64
		et.mu.Lock()
		if et.end {
			break
		}

		currentTime := time.Now()
		currentMillSecond := currentTime.UnixNano() / int64(time.Millisecond)
		for i := 0; i < len(et.ticker); i++ {
			et.ticker[i].mu.Lock()
			if et.ticker[i].nextTime == 0 {
				temp := currentMillSecond % et.ticker[i].duration
				if temp == 0 {
					nextDiffTime = 0
				} else {
					nextDiffTime = et.ticker[i].duration - temp
				}
				et.ticker[i].nextTime = currentMillSecond + nextDiffTime
			}

			if et.ticker[i].nextTime <= currentMillSecond {
				select {
				case et.ticker[i].C <- currentTime:
				default:
					//送信エラー出力は1秒以上空ける
					if currentMillSecond-errorSendMillsecond >= 1000 {
						Logger.Writef(LOG_LEVEL_WARNING, "timer send error [duration] =%v\n", et.ticker[i].duration)
						errorSendMillsecond = currentMillSecond
					}
				}
				et.ticker[i].nextTime += et.ticker[i].duration
			}
			et.ticker[i].mu.Unlock()
		}
		et.mu.Unlock()
		time.Sleep(1 * time.Millisecond)
	}

}

func (et *ExactlyTimer) Stop() {
	et.mu.Lock()
	et.end = true
	et.mu.Unlock()
}

func (ticker *Ticker) Reset(duration int64) {
	ticker.mu.Lock()
	if ticker.duration != duration {
		ticker.duration = duration
		ticker.nextTime = 0
	}
	ticker.mu.Unlock()
}

func (ticker *Ticker) Flush() {
	ticker.mu.Lock()
Loop:
	for {
		select {
		case <-ticker.C:
		default:
			break Loop
		}
	}
	ticker.mu.Unlock()
}
