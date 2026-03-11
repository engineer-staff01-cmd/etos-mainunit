package app

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const numberOfAnalogInputPort = 8

//const IOUnitWatchdogTime = 300
//const IOUnitWatchdogEnable = 1

type R1240 struct {
	ms              *ModbusSerial
	temperature     [numberOfAnalogInputPort]float64
	humidity        [numberOfAnalogInputPort]float64
	id              byte
	CorrectionValue float64 // 補正量[℃]
	CorrectionRatio float64 // 補正倍率
}

var r1240 []R1240

func NewR1240(deviceId byte, ms *ModbusSerial) *R1240 {
	var ret *R1240

	for i := range r1240 {
		if r1240[i].id == deviceId {
			ret = &r1240[i]
			break
		}
	}

	if ret == nil {
		ret = new(R1240)
		ret.id = deviceId
		ret.ms = ms
		// 初期化時の補正値は影響がないように+0℃と+1倍にしておく
		ret.CorrectionValue = 0
		ret.CorrectionRatio = 1
		r1240 = append(r1240, *ret)
	}

	return ret
}

func (r1240 *R1240) Channels() int16 {
	return numberOfAnalogInputPort
}

func (r1240 *R1240) UpdateData() (ret error) {
	ret = nil
	var stopBits int = 2
	//results, err := r1240.ms.ReadInputRegisters(r1240.id, 0x00, numberOfAnalogInputPort*2, stopBits)
	results, err := r1240.ms.ReadInputRegisters(r1240.id, 0x02C0, numberOfAnalogInputPort, stopBits)
	if err == nil {
		for i := 0; i < numberOfAnalogInputPort; i++ {
			value := r1240.calculateTemperature(results[i*2 : (i+1)*2])
			r1240.temperature[i] = value
			value = r1240.calculateHumidity(results[i*2 : (i+1)*2])
			r1240.humidity[i] = value
			//Logger.Writef(LOG_LEVEL_DEBUG, "Humidity Data : %f. ch : %d", r1240.humidity[i], i)
		}
	} else {
		ret = err
		Logger.Writef(LOG_LEVEL_ERR, "R1240 Read Error")
		return ret
	}
	/*	if err != nil {
			return err
		} else if len(results) != 4*numberOfAnalogInputPort {
			return fmt.Errorf("received data is illegal!")
		}
		for i := 0; i < numberOfAnalogInputPort; i++ {
			r1240.inputState[i] = results[i*2+1]
		}	/*  */

	return nil
}

func (r1240 *R1240) calculateTemperature(data []byte) float64 {
	var temp_max float64 = 80.0
	var temp_min float64 = -30.0
	var resolution float64 = 65536.0 // 16bit
	//var resolution float64 = float64(0x0000FFFE - 0x00003333) // 4-20mAの測定レンジ
	var value uint16
	//Logger.Writef(LOG_LEVEL_DEBUG, "loadTemplate(sensorTemplateFilepath) :%#v, resolution : %f", data, resolution)

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &value)
	//Logger.Writef(LOG_LEVEL_DEBUG, "Temperature AD value :%d(%x), resolution : %f", value, value, resolution)
	return float64(value)*(temp_max-temp_min)/resolution + temp_min // Type Code 0x29
	//return float64(value)/327.675 - 50.0 // Type Code 0x20
}

func (r1240 *R1240) calculateHumidity(data []byte) float64 {
	var humi_max float64 = 100.0
	var humi_min float64 = 0.0
	var resolution float64 = 65536.0 // 16bit
	//var resolution float64 = float64(0x0000FFFE - 0x00003333) // 4-20mAの測定レンジ
	var value uint16
	//Logger.Writef(LOG_LEVEL_DEBUG, "loadTemplate(sensorTemplateFilepath) :%#v", data)

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &value)
	//Logger.Writef(LOG_LEVEL_DEBUG, "Humidity AD value :%d(%x), resolution : %f", value, value, resolution)
	return float64(value)*(humi_max-humi_min)/resolution + humi_min // Type Code 0x29
	//return float64(value)/327.675 - 50.0 // Type Code 0x20
}

func (r1240 *R1240) GetTemperature(channel uint16, value, ratio NullFloat64) (float64, error) {

	if channel > 8 || channel == 0 {
		return 0, fmt.Errorf("Error:%s", "portNumber is illegal!\n")
	}

	var t = r1240.temperature[channel-1]
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

func (r1240 *R1240) GetHumidity(channel uint16) (float64, error) {

	if channel > 8 || channel == 0 {
		return 0, fmt.Errorf("Error:%s", "portNumber is illegal!\n")
	}

	return r1240.humidity[channel-1], nil
}

func (r1240 *R1240) convertUint32Value(data []byte) uint32 {
	var value uint32

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &value)

	return value
}

func (r1240 *R1240) convertInt32Value(data []byte) int32 {
	var value int32

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &value)

	return value
}
