package app

import (
	"fmt"
	"log/syslog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var jst *time.Location

func init() {
	var err error
	jst, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		// タイムゾーンの取得に失敗した場合はUTCを使用
		jst = time.UTC
	}
}

type UnitLog struct {
	file      *os.File
	syslog    *syslog.Writer
	tag       string
	mu        sync.Mutex
	logDir    string
}

// ログレベル
const (
	LOG_LEVEL_EMERG int = iota
	LOG_LEVEL_ALERT
	LOG_LEVEL_CRIT
	LOG_LEVEL_ERR
	LOG_LEVEL_WARNING
	LOG_LEVEL_NOTICE
	LOG_LEVEL_INFO
	LOG_LEVEL_DEBUG
)

var ul *UnitLog
var Logger = NewUnitLog("ecoRAMDAR")

// コンストラクタ代わり
func NewUnitLog(tag string) *UnitLog {
	if ul == nil {
		ul = new(UnitLog)
		ul.tag = tag
		ul.logDir = "/vol_data/log"
		
		// ログディレクトリが存在しない場合は作成
		if err := os.MkdirAll(ul.logDir, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create log directory %s: %v", ul.logDir, err))
		}
		
		// ログファイルを開く（追記モード）
		logPath := fmt.Sprintf("%s/%s.log", ul.logDir, tag)
		var err error
		ul.file, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic(fmt.Sprintf("Failed to open log file %s: %v", logPath, err))
		}
		
		// syslogに接続
		ul.syslog, err = syslog.New(syslog.LOG_USER|syslog.LOG_INFO, tag)
		if err != nil {
			// syslogへの接続に失敗してもファイルログは継続
			// エラーは無視（syslogが利用できない環境でも動作するように）
		}
	}
	return ul
}

// ログ書き込み（フォーマット指定）
func (l *UnitLog) Writef(level int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	l.Write(level, msg)
}

// ログ書き込み
func (l *UnitLog) Write(level int, msg string) {
	switch level {
	case LOG_LEVEL_EMERG:
		l.Emerg(msg)
	case LOG_LEVEL_ALERT:
		l.Alert(msg)
	case LOG_LEVEL_CRIT:
		l.Crit(msg)
	case LOG_LEVEL_ERR:
		l.Err(msg)
	case LOG_LEVEL_WARNING:
		l.Warning(msg)
	case LOG_LEVEL_NOTICE:
		l.Notice(msg)
	case LOG_LEVEL_INFO:
		l.Info(msg)
	case LOG_LEVEL_DEBUG:
		l.Debug(msg)
	}
}

// getCallerFuncName 呼び出し元の関数名を取得（ログ関数とWrite/Writefをスキップ）
func (l *UnitLog) getCallerFuncName() string {
	// スキップする関数名のリスト
	skipFuncs := []string{
		".Emerg", ".Alert", ".Crit", ".Err", ".Warning", ".Notice", ".Info", ".Debug",
		".Write", ".Writef", ".writeLog", ".getCallerFuncName",
	}
	
	// 最大10フレームまで追跡
	for skip := 2; skip < 12; skip++ {
		pc, _, _, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		fullName := fn.Name()
		
		// スキップ対象の関数かチェック
		shouldSkip := false
		for _, skipFunc := range skipFuncs {
			if strings.HasSuffix(fullName, skipFunc) {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}
		
		// パッケージ名とレシーバー名を除去して関数名のみを取得
		// 例: "github.com/A9E/etos-mainunit/app.(*UnitLog).SomeFunction" -> "SomeFunction"
		// 例: "github.com/A9E/etos-mainunit/app.SomeFunction" -> "SomeFunction"
		parts := strings.Split(fullName, ".")
		if len(parts) > 0 {
			funcName := parts[len(parts)-1]
			// レシーバー名を除去（例: "(*UnitLog).SomeFunction" -> "SomeFunction"）
			if idx := strings.Index(funcName, ")"); idx >= 0 {
				if idx+1 < len(funcName) && funcName[idx+1] == '.' {
					funcName = funcName[idx+2:]
				} else if idx+1 < len(funcName) {
					funcName = funcName[idx+1:]
				}
			}
			return funcName
		}
	}
	return l.tag // デフォルトはtagを使用
}

// getSyslogPriority ログレベルをsyslogの優先度に変換
func (l *UnitLog) getSyslogPriority(level string) syslog.Priority {
	switch level {
	case "EMERG":
		return syslog.LOG_EMERG
	case "ALERT":
		return syslog.LOG_ALERT
	case "CRITICAL":
		return syslog.LOG_CRIT
	case "ERROR":
		return syslog.LOG_ERR
	case "WARNING":
		return syslog.LOG_WARNING
	case "NOTICE":
		return syslog.LOG_NOTICE
	case "INFO":
		return syslog.LOG_INFO
	case "DEBUG":
		return syslog.LOG_DEBUG
	default:
		return syslog.LOG_INFO
	}
}

// writeLog ログをファイルとsyslogに書き込む（スレッドセーフ）
func (l *UnitLog) writeLog(level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	funcName := l.getCallerFuncName()
	timestamp := time.Now().In(jst).Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("%s [%s] %s: %s\n", timestamp, level, funcName, msg)
	
	if MODEL == "A9E" {
		// ファイルに書き込み
		if l.file != nil {
			if _, err := l.file.WriteString(logLine); err != nil {
				// ログ書き込みエラーは無視（無限ループを防ぐため）
			}
			// 即座にディスクに書き込む
			l.file.Sync()
		}
	}
	
	// syslogに書き込み
	if l.syslog != nil {
		syslogMsg := fmt.Sprintf("[%s] %s: %s", level, funcName, msg)
		syslogPriority := l.getSyslogPriority(level)
		
		switch syslogPriority {
		case syslog.LOG_EMERG:
			_ = l.syslog.Emerg(syslogMsg)
		case syslog.LOG_ALERT:
			_ = l.syslog.Alert(syslogMsg)
		case syslog.LOG_CRIT:
			_ = l.syslog.Crit(syslogMsg)
		case syslog.LOG_ERR:
			_ = l.syslog.Err(syslogMsg)
		case syslog.LOG_WARNING:
			_ = l.syslog.Warning(syslogMsg)
		case syslog.LOG_NOTICE:
			_ = l.syslog.Notice(syslogMsg)
		case syslog.LOG_INFO:
			_ = l.syslog.Info(syslogMsg)
		case syslog.LOG_DEBUG:
			_ = l.syslog.Debug(syslogMsg)
		}
	}
}

// Emergency（レベル０：最上位）
func (l *UnitLog) Emerg(msg string) {
	l.writeLog("EMERG", msg)
}

// Alert(レベル１)
func (l *UnitLog) Alert(msg string) {
	l.writeLog("ALERT", msg)
}

// Critical(レベル２)
func (l *UnitLog) Crit(msg string) {
	l.writeLog("CRITICAL", msg)
}

// Error（レベル３）
func (l *UnitLog) Err(msg string) {
	l.writeLog("ERROR", msg)
}

// Warning（レベル４）
func (l *UnitLog) Warning(msg string) {
	l.writeLog("WARNING", msg)
}

// Notice（レベル５）
func (l *UnitLog) Notice(msg string) {
	l.writeLog("NOTICE", msg)
}

// Information（レベル６）
func (l *UnitLog) Info(msg string) {
	l.writeLog("INFO", msg)
}

// Debug（レベル７）
func (l *UnitLog) Debug(msg string) {
	l.writeLog("DEBUG", msg)
}
