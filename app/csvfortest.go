//go:build debug

package app

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"
)

var adam_mapParam = map[uint16]string{
	0x00: "Temperature"}

var ke1_mapParam = map[uint16]string{
	0x00:  "Voltage",
	0x0C:  "Current",
	0x24:  "PowerFactor",
	0x34:  "Frequency",
	0x38:  "EffectivePower",
	0x100: "EffectiveIntegratedElectricEnergy"}

var kmn1_mapParam = map[uint16]string{
	0x00:  "Voltage",
	0x06:  "Current",
	0x0C:  "PowerFactor",
	0x0E:  "Frequency",
	0x10:  "EffectivePower",
	0x200: "EffectiveIntegratedElectricEnergy"}

/*
func csv_write(ke1 *KE1) {
	var buf [6][64]byte
	var records [][]string

	f, err := os.Create("/root/test.csv") // 書き込む先のファイル
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "File Write :%s", err.Error())
	}

	for i := 0; i < numberOfVoltage; i++ {
		buf[0][i] = strconv.FormatFloat(ke1.voltage[i], 'f', -1, 64)
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "Write to record[0] : %+v", buf[0])
	records[0] = buf[0][:numberOfVoltage]

	for i := 0; i < numberOfCurrent; i++ {
		buf[1][i] = strconv.FormatFloat(ke1.current[i], 'f', -1, 64)
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "Write to record[1] : %+v", buf[1])
	records[1] = buf[1][:numberOfCurrent]

	buf[2][0] = strconv.FormatFloat(ke1.frequency, 'f', -1, 64)
	Logger.Writef(LOG_LEVEL_DEBUG, "Write to record[2] : %+v", buf[2])
	records[2] = buf[2][:1]

	for i := 0; i < numberOfPowerFactor; i++ {
		buf[3][i] = strconv.FormatFloat(ke1.powerFactor[i], 'f', -1, 64)
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "Write to record[3] : %+v", buf[3])
	records[3] = buf[3][:numberOfPowerFactor]

	for i := 0; i < numberOfEffectivePower; i++ {
		buf[4][i] = strconv.FormatFloat(ke1.effectivePower[i], 'f', -1, 64)
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "Write to record[4] : %+v", buf[4])
	records[4] = buf[4][:numberOfEffectivePower]

	for i := 0; i < numberOfIntegratedElectricEnergy; i++ {
		buf[5][i] = strconv.FormatFloat(ke1.integratedElectricEnergy[i], 'f', -1, 64)
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "Write to record[5] : %+v", buf[5])
	records[5] = buf[5][:numberOfIntegratedElectricEnergy]

	w := csv.NewWriter(f)
	w.WriteAll(records) // 一度にすべて書き込む
}
*/

var prevTime int64 = 0
var indexCnt = 0

const csvReadDuration = 60 * 1000

func adam_csv_read(address, quantity uint16) ([]byte, error) {
	var result []byte
	var buf [64]byte

	currentTime := time.Now()
	currentMillSecond := currentTime.UnixNano() / int64(time.Millisecond)

	if currentMillSecond > (prevTime + csvReadDuration) {
		indexCnt++
		prevTime = currentMillSecond
		if indexCnt > 13 {
			indexCnt = 0
		}
	}

	filename := "/root/" + adam_mapParam[address] + ".csv"
	//Logger.Writef(LOG_LEVEL_DEBUG, "CSV file open : %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "file open error:%+v", err.Error())
	}
	defer file.Close()

	r := csv.NewReader(file)
	rows, err := r.ReadAll() // csvを一度に全て読み込む

	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "file read error:%+v", err.Error())
	} else {
		Logger.Writef(LOG_LEVEL_DEBUG, "CSV file data : %#v", rows)
		for i, v := range rows[indexCnt] {
			//Logger.Writef(LOG_LEVEL_DEBUG, "CSV read data num: %d string: %s", i, v)
			in_dat, err := strconv.ParseInt(strings.TrimSpace(v), 10, 32)
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "String to Int error:%+v", err.Error())
				return result, err
			} else {
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i data: %d", in_dat)
				buf[i*2] = byte((in_dat >> 8) & 0xff)
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i result data: %d", buf[i*4+2])
				buf[i*2+1] = byte(in_dat & 0xff)
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i result data: %d", buf[i*4+3])
			}
		}
	}

	result = buf[:quantity*2]
	Logger.Writef(LOG_LEVEL_DEBUG, "CSV file read result : %+v", result)

	return result, err
}

func ke1_csv_read(address, quantity uint16) ([]byte, error) {
	var result []byte
	var buf [64]byte

	filename := "/root/" + ke1_mapParam[address] + ".csv"
	//Logger.Writef(LOG_LEVEL_DEBUG, "CSV file open : %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "file open error:%+v", err.Error())
	}
	defer file.Close()

	r := csv.NewReader(file)
	rows, err := r.ReadAll() // csvを一度に全て読み込む
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "file read error:%+v", err.Error())
	} else {
		Logger.Writef(LOG_LEVEL_DEBUG, "CSV file data : %#v", rows)
		for i, v := range rows[0] {
			//Logger.Writef(LOG_LEVEL_DEBUG, "CSV read data num: %d string: %s", i, v)
			in_dat, err := strconv.ParseInt(strings.TrimSpace(v), 10, 32)
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "String to Int error:%+v", err.Error())
				return result, err
			} else {
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i data: %d", in_dat)
				buf[i*4] = byte((in_dat >> 24) & 0xff)
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i result data: %d", buf[i*4])
				buf[i*4+1] = byte((in_dat >> 16) & 0xff)
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i result data: %d", buf[i*4+1])
				buf[i*4+2] = byte((in_dat >> 8) & 0xff)
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i result data: %d", buf[i*4+2])
				buf[i*4+3] = byte(in_dat & 0xff)
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i result data: %d", buf[i*4+3])
			}
		}
	}

	result = buf[:quantity*2]
	Logger.Writef(LOG_LEVEL_DEBUG, "CSV file read result : %+v", result)
	return result, err
}

func kmn1_csv_read(address, quantity uint16) ([]byte, error) {
	var result []byte
	var buf [64]byte

	filename := "/root/" + kmn1_mapParam[address] + ".csv"
	//Logger.Writef(LOG_LEVEL_DEBUG, "CSV file open : %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "file open error:%+v", err.Error())
	}
	defer file.Close()

	r := csv.NewReader(file)
	rows, err := r.ReadAll() // csvを一度に全て読み込む
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "file read error:%+v", err.Error())
	} else {
		Logger.Writef(LOG_LEVEL_DEBUG, "CSV file data : %#v", rows)
		for i, v := range rows[0] {
			//Logger.Writef(LOG_LEVEL_DEBUG, "CSV read data num: %d string: %s", i, v)
			in_dat, err := strconv.ParseInt(strings.TrimSpace(v), 10, 32)
			if err != nil {
				Logger.Writef(LOG_LEVEL_ERR, "String to Int error:%+v", err.Error())
				return result, err
			} else {
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i data: %d", in_dat)
				buf[i*4] = byte((in_dat >> 24) & 0xff)
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i result data: %d", buf[i*4])
				buf[i*4+1] = byte((in_dat >> 16) & 0xff)
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i result data: %d", buf[i*4+1])
				buf[i*4+2] = byte((in_dat >> 8) & 0xff)
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i result data: %d", buf[i*4+2])
				buf[i*4+3] = byte(in_dat & 0xff)
				//Logger.Writef(LOG_LEVEL_DEBUG, "CSV A to i result data: %d", buf[i*4+3])
			}
		}
	}

	result = buf[:quantity*2]
	Logger.Writef(LOG_LEVEL_DEBUG, "CSV file read result : %+v", result)
	return result, err
}

/*
func prompt_ui(deviceID byte, address, quantity uint16) ([]byte, error) {
	var result []byte
	var in_str [12]string
	var in_dat [12]int
	var buf [64]byte
	var err error

	// 入力が不正な場合errorを返す関数を作成
	validate := func(input string) error {
		_, err := strconv.ParseInt(input, 10, 32)
		if err != nil {
			return errors.New("Invalid number")
		}
		return nil
	}

	for i := 0; i < (quantity / 2); i++ {
		// インタラクションの表示やバリデーションを設定
		prompt := promptui.Prompt{
			Label:    mapParam[address] + "(" + strconv.Itoa(i) + ")", // 表示する文言
			Validate: validate,                                             // validate
		}

		in_str[i], err = prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			i--
		} else {
			in_dat[i], _ = strconv.Atoi(strings.TrimSpace(in_str[i]), 10, 0)
			fmt.Printf("You choose %d\n", in_dat[i])
			buf[i*4] = byte((in_dat[i] >> 24) & 0xff)
			buf[i*4+1] = byte((in_dat[i] >> 16) & 0xff)
			buf[i*4+2] = byte((in_dat[i] >> 8) & 0xff)
			buf[i*4+3] = byte(in_dat[i] & 0xff)
		}
	}

	result = buf[:quantity*2]
	return result, err
}
*/
