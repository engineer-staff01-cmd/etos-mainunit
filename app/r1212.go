package app

import (
	"fmt"
)

const numberOfInputPort = 8

const IOUnitWatchdogTime = 300
const IOUnitWatchdogEnable = 1

type R1212 struct {
	ms         *ModbusSerial
	inputState [numberOfInputPort]byte
	id         byte
}

var r1212 []R1212

func NewR1212(deviceId byte, ms *ModbusSerial) *R1212 {
	var ret *R1212

	for i := range r1212 {
		if r1212[i].id == deviceId {
			ret = &r1212[i]
			break
		}
	}

	if ret == nil {
		ret = new(R1212)
		ret.id = deviceId
		ret.ms = ms
		r1212 = append(r1212, *ret)
	}

	return ret
}

func (r *R1212) Channels() int16 {
	return numberOfInputPort
}

func (r *R1212) UpdateData() error {
	var stopBits int = 2
	results, err := r.ms.ReadInputRegisters(r.id, 0x00, numberOfInputPort, stopBits)
	if err != nil {
		return err
	} else if len(results) != 2*numberOfInputPort {
		return fmt.Errorf("Error:%s", "received data is illegal!")
	}
	for i := 0; i < numberOfInputPort; i++ {
		r.inputState[i] = results[i*2+1]
	}

	return nil
}

func (r *R1212) GetInputValue(channel uint16) (byte, error) {

	if channel > 8 || channel == 0 {
		return 0, fmt.Errorf("Error:%s", "portNumber is illegal!")
	}

	return r.inputState[channel-1], nil
}

func (r *R1212) GetOutputStatus(channel uint16) (byte, error) {

	if channel > 8 || channel == 0 {
		return 0, fmt.Errorf("Error:%s", "portNumber is illegal!")
	}
	portNumber := channel - 1
	address := 0x0140 + portNumber
	var stopBits int = 2
	results, err := r.ms.ReadCoils(r.id, address, 1, stopBits)

	if err != nil {
		return 0, err
	} else {
		return results[0], err
	}
}

func (r *R1212) SetOutputStatus(channel uint16, value uint16) error {
	var setValue, address uint16

	if channel > 8 || channel == 0 {
		return fmt.Errorf("Error:%s", "portNumber is illegal!")
	}
	portNumber := channel - 1
	if value == 0 {
		setValue = 0x0000
	} else {
		setValue = 0xFF00
	}
	address = portNumber + 0x140
	var stopBits int = 2
	_, err := r.ms.WriteSingleCoil(r.id, address, setValue, stopBits)

	return err
}

func (r *R1212) GetBatteryVoltage() (float64, error) {
	return 12.0, nil
}

func (r *R1212) GetPulseCount() (uint32, error) {
	// In1 固定
	var stopBits int = 2
	results, err := r.ms.ReadInputRegisters(r.id, 0x0020, 2, stopBits)
	if err != nil {
		return 0, err
	}

	// 上位下位反転
	var value uint32 = 0
	value += uint32(results[2]) << 24
	value += uint32(results[3]) << 16
	value += uint32(results[0]) << 8
	value += uint32(results[1])

	return value, nil
}

func (r *R1212) GetMeasurementElapsedTime() (uint16, error) {
	return 0, fmt.Errorf("wired connection")
}

func (r *R1212) SetWatchdog() error {
	var address, quantity uint16
	var err error

	address = 0x7565 // SYS_modbusWatchdogTimeout
	quantity = 1
	var stopBits int = 2

	//	Logger.Writef(LOG_LEVEL_DEBUG, "IOUNIT Watchdog Setup Start")
	results, err := r.ms.ReadHoldingRegisters(r.id, address, quantity, stopBits)
	Logger.Writef(LOG_LEVEL_DEBUG, "SYS_modbusWatchdogTimeout:%d", results)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to read Watchdog Time:%02d Error:%s", r.id, err.Error())
		return err
	}

	// 上位下位反転
	var value uint32 = 0
	//	value += uint32(results[2]) << 24
	//	value += uint32(results[3]) << 16
	value += uint32(results[0]) << 8
	value += uint32(results[1])

	Logger.Writef(LOG_LEVEL_DEBUG, "Watchdog Timeout:%d sec", value)
	if value != IOUnitWatchdogTime {
		_, err = r.ms.WriteSingleRegister(r.id, address, IOUnitWatchdogTime, stopBits)
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to Write Watchdog time:%02d Error:%s", r.id, err.Error())
			return err
		}
	}

	//	Logger.Writef(LOG_LEVEL_DEBUG, "Watchdog Funtion read")
	address = 0x7564 // SYS_modbusWatchdogFuntion
	results, err = r.ms.ReadHoldingRegisters(r.id, address, quantity, stopBits)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "Failed to Write Watchdog Function:%02d Error:%s", r.id, err.Error())
		return err
	}

	Logger.Writef(LOG_LEVEL_DEBUG, "SYS_modbusWatchdogFuntion read:%d", results[1])

	//	value = 0
	//	value += uint32(results[1]) << 8
	//	value += uint32(results[0])

	if results[1] != IOUnitWatchdogEnable {
		_, err = r.ms.WriteSingleRegister(r.id, address, IOUnitWatchdogEnable, stopBits)
		if err != nil {
			Logger.Writef(LOG_LEVEL_ERR, "Failed to Write Watchdog Enable:%02d Error:%s", r.id, err.Error())
			return err
		}
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "SYS_modbusWatchdogFuntion set:%d", IOUnitWatchdogEnable)

	address = 0x7566 // SYS_modbusWatchdogStatus
	_, err = r.ms.WriteSingleRegister(r.id, address, 0, stopBits)
	Logger.Writef(LOG_LEVEL_DEBUG, "SYS_modbusWatchdogStatus set:%d", 0)

	return err
}
