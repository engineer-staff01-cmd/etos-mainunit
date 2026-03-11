package app

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const NumberOfTemperature = 6

type ADAM4015 struct {
	ms              *ModbusSerial
	temperature     [NumberOfTemperature]float64
	id              byte
	CorrectionValue float64 // 補正量[℃]
	CorrectionRatio float64 // 補正倍率
}

var adam4015 []ADAM4015

func NewAdAM4015(deviceId byte, ms *ModbusSerial) *ADAM4015 {
	var ret *ADAM4015

	for i := range adam4015 {
		if adam4015[i].id == deviceId {
			ret = &adam4015[i]
			break
		}
	}

	if ret == nil {
		ret = new(ADAM4015)
		ret.id = deviceId
		ret.ms = ms
		// 初期化時の補正値は影響がないように+0℃と+1倍にしておく
		ret.CorrectionValue = 0
		ret.CorrectionRatio = 1
		adam4015 = append(adam4015, *ret)
	}

	return ret
}

func (adam4015 *ADAM4015) InitSensorType() {
	var stopBits int = 2
	var address uint16 = 0x00C8
	var code uint16 = 0x0029
	var quantity uint16 = 12
	var quantityHalf uint16 = quantity / 2
	var setValue = []byte{0x00, 0x29, 0x00, 0x29, 0x00, 0x29, 0x00, 0x29, 0x00, 0x29, 0x00, 0x29} // Pt Type -200～200℃
	var err error
	var results []byte

	// PTタイプ変更
	adam4015.ms.WriteMultipleRegister(adam4015.id, address, quantity, setValue, stopBits)
	_, err = adam4015.ms.WriteSingleRegister(adam4015.id, address, code, stopBits)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Temperature Sensor Type Code  Write Error : %v, Device ID : %d", err, adam4015.id)
	}

	results, err = adam4015.ms.ReadHoldingRegisters(adam4015.id, address, quantityHalf, stopBits)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Temperature Sensor Type Code Read Error : %v, Device ID : %d", err, adam4015.id)
	} else {
		Logger.Writef(LOG_LEVEL_DEBUG, "Temperature Sensor Type Code : %#v, Device ID : %d", results, adam4015.id)
	}
}

func (adam4015 *ADAM4015) UpdateData() error {
	var stopBits int = 2
	results, err := adam4015.ms.ReadHoldingRegisters(adam4015.id, 0, NumberOfTemperature, stopBits)
	//adam4015.ms.handler.StopBits = 2
	//	results, err := adam_csv_read(0, NumberOfTemperature)
	//	Logger.Writef(LOG_LEVEL_DEBUG, "loadTemplate(sensorTemplateFilepath) :%+v", results)
	if err == nil {
		for i := 0; i < NumberOfTemperature; i++ {
			adam4015.temperature[i] = adam4015.calculateTemperature(results[i*2 : (i+1)*2])
		}
	}

	return err
}

func (adam4015 *ADAM4015) GetTemperature(channel uint16, value, ratio NullFloat64) (float64, error) {

	if channel > 7 || channel == 0 {
		return 0, fmt.Errorf("Error:%s", "channel is illegal!\n")
	}

	var t = adam4015.temperature[channel-1]
	var v = float64(0.0)
	var r = float64(1.0)
	if value.Valid {
		v = value.Float64
	}
	if ratio.Valid {
		r = ratio.Float64
	}
	// 温度(℃) = 測定温度(℃) * 補正倍率 + 補正量(℃)
	t = (t * r) + v

	return t, nil
}

func (adam4015 *ADAM4015) calculateTemperature(data []byte) float64 {
	var temp_max float64 = 200.0
	var temp_min float64 = -200.0
	var resolution float64 = 65536.0 // 16bit
	var value uint16
	//Logger.Writef(LOG_LEVEL_DEBUG, "loadTemplate(sensorTemplateFilepath) :%#v", data)

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &value)
	//Logger.Writef(LOG_LEVEL_DEBUG, "loadTemplate(sensorTemplateFilepath) :%x", value)
	return float64(value)*(temp_max-temp_min)/resolution + temp_min // Type Code 0x29
	//return float64(value)/327.675 - 50.0 // Type Code 0x20
}
