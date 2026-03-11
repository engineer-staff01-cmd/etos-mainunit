package app

import (
	"os/exec"
	"strings"
)

// Clamp min ~ max の範囲内の値を返す
func Clamp(min, max, value int64) int64 {
	if value <= min {
		value = min
	}
	if max <= value {
		value = max
	}
	return value
}

func GetUPSStatus() (string, error) {
	bytes, err := exec.Command("apcaccess").Output()
	if err != nil {
		return "STATUS   : COMMERR", err
	}
	return string(bytes), nil
}

// ReadPowerOutageStatus apcaccessの出力文字列からステータスを取り出す
func ReadPowerOutageStatus(output string) string {
	lines := strings.Split(output, "\n")

	for _, l := range lines {
		if strings.Contains(l, "STATUS") {
			arr := strings.Split(l, ":")
			if len(arr) < 2 {
				continue
			}
			return strings.TrimSpace(arr[1])
		}
	}
	return ""
}

// IsBatteryPowered 電源供給がバッテリーの場合 trueを返す
func IsBatteryPowered() bool {
	output, _ := GetUPSStatus()
	if ReadPowerOutageStatus(output) == UPS_STATUS_BATTERY {
		return true
	}
	return false
}
