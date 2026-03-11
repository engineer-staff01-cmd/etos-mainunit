package app

import (
	"bytes"
	"encoding/binary"
)

type WsZ3030 struct {
	ms *ModbusSerial
	id byte
}

var wsZ3030 []WsZ3030

func NewWsZ3030(deviceId byte, ms *ModbusSerial) *WsZ3030 {
	var ret *WsZ3030

	for i := range wsZ3030 {
		if wsZ3030[i].id == deviceId {
			ret = &wsZ3030[i]
			break
		}
	}

	if ret == nil {
		ret = new(WsZ3030)
		ret.id = deviceId
		ret.ms = ms
		wsZ3030 = append(wsZ3030, *ret)
	}

	return ret
}

func (ws *WsZ3030) GetBatteryVoltage() (float64, error) {
	var value uint16
	var stopBits int = 2
	results, err := ws.ms.ReadInputRegisters(ws.id, 0x2A, 1, stopBits)
	buf := bytes.NewReader(results)
	binary.Read(buf, binary.BigEndian, &value)
	voltage := float64(value) * 0.01

	return voltage, err
}

func (ws *WsZ3030) GetPulseCount() (uint32, error) {

	var value uint32
	var stopBits int = 2
	results, err := ws.ms.ReadInputRegisters(ws.id, 0x24, 2, stopBits)

	buf := bytes.NewReader(results)
	binary.Read(buf, binary.BigEndian, &value)

	return value, err
}

func (ws *WsZ3030) GetMeasurementElapsedTime() (uint16, error) {
	var value uint16
	var stopBits int = 2
	results, err := ws.ms.ReadInputRegisters(ws.id, 0x2B, 1, stopBits)
	buf := bytes.NewReader(results)
	binary.Read(buf, binary.BigEndian, &value)

	return value, err
}
