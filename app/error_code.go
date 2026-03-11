package app

import (
	"fmt"
)

const UnknownErrorMessage = "unknown error"

const (
	ERCD_SUCCESS   uint32 = iota // E0000
	ERCD_UPS_ERROR        = 1    // E0001

	ERCD_CONTROL_CONDITION_ERROR = 100 // E0100
)

var errorMessages = map[uint32]string{
	ERCD_SUCCESS:                 "connect successfully to cloud. firmware: %s",
	ERCD_UPS_ERROR:               "UPS alert :%s",
	ERCD_CONTROL_CONDITION_ERROR: "Control release time is smaller than the required operation time",
}

// getErrorCode エラーコードから文字列を生成
func getErrorCode(code uint32) string {
	return fmt.Sprintf("E%04d", code)
}

// getErrorMessageFromCode エラーコードから対応するエラーメッセージを生成
func getErrorMessageFromCode(code uint32, args ...interface{}) string {
	message, ok := errorMessages[code]
	if !ok {
		return UnknownErrorMessage
	}
	return fmt.Sprintf(message, args...)
}

// GenerateSystemAlert エラーコードから対応するシステムアラートを生成
func GenerateSystemAlert(t int64, code uint32, args ...interface{}) SystemAlert {
	return SystemAlert{
		Time:      t,
		ErrorCode: getErrorCode(code),
		Message:   getErrorMessageFromCode(code, args...),
	}
}
