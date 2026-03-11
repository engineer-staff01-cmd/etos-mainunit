package app

import "strings"

const UNIT_TEST_NAME = "UnitTest"

const (
	CT_225A = "225A"
	CT_100A = "100A"
)

type EnergyUnit interface {
	GetId() byte
	setCTtype(ctType []byte)
	UpdateData() error
	GetVoltage1(id uint16) (float64, error)
	GetVoltage2(id uint16) (float64, error)
	GetVoltage3(id uint16) (float64, error)
	GetCurrent1(id uint16) (float64, error)
	GetCurrent2(id uint16) (float64, error)
	GetCurrent3(id uint16) (float64, error)
	GetFrequency(id uint16) (float64, error)
	GetPowerFactor(id uint16) (float64, error)
	GetEffectivePower(id uint16) (float64, error)
	GetEffectiveIntegratedElectricEnergy(id uint16) (float64, error)
}

type TemperatureUnit interface {
	InitSensorType()
	UpdateData() error
	GetTemperature(channel uint16, value, ratio NullFloat64) (float64, error)
}

type HumidityUnit interface {
	UpdateData() error
	GetTemperature(channel uint16, value, ratio NullFloat64) (float64, error)
	GetHumidity(channel uint16) (float64, error)
}

type IoUnit interface {
	UpdateData() error
	GetInputValue(portNumber uint16) (byte, error)
	GetOutputStatus(portNumber uint16) (byte, error)
	SetOutputStatus(portNumber uint16, value uint16) error
	Channels() int16
	SetWatchdog() error
}

type PulseUnit interface {
	GetBatteryVoltage() (float64, error)
	GetPulseCount() (uint32, error)
	GetMeasurementElapsedTime() (uint16, error)
}

var testTemperatureUnit TemperatureUnit
var testHumidityUnit HumidityUnit
var testIoUnit IoUnit

func SetTemperatureTestInterface(tu TemperatureUnit) {
	testTemperatureUnit = tu
}

func SetIoUnitTestInterface(iu IoUnit) {
	testIoUnit = iu
}

func NewEnergyUnit(deviceName string, deviceId byte) EnergyUnit {
	var ret EnergyUnit

	ms := NewModbusSerial()
	if strings.Contains(deviceName, UNIT_TEST_NAME) {
		ret = nil
	} else {
		ret = NewKMN1(deviceId, ms)
	}
	return ret
}

func NewTemperatureUnit(deviceName string, deviceId byte) TemperatureUnit {
	var ret TemperatureUnit

	ms := NewModbusSerial()
	if strings.Contains(deviceName, UNIT_TEST_NAME) {
		ret = testTemperatureUnit
	} else {
		ret = NewAdAM4015(deviceId, ms)
	}
	return ret
}

func NewHumidityUnit(deviceName string, deviceId byte) HumidityUnit {
	var ret HumidityUnit

	ms := NewModbusSerial()
	if strings.Contains(deviceName, UNIT_TEST_NAME) {
		ret = testHumidityUnit
	} else {
		ret = NewR1240(deviceId, ms)
	}
	return ret
}

func NewIoUnit(deviceName string, deviceId byte) IoUnit {
	var ret IoUnit

	ms := NewModbusSerial()
	if strings.Contains(deviceName, UNIT_TEST_NAME) {
		ret = testIoUnit
	} else {
		ret = NewR1212(deviceId, ms)
	}
	return ret
}

func NewPulseUnit(deviceName string, deviceId byte) PulseUnit {
	var ret PulseUnit

	ms := NewModbusSerial()
	switch deviceName {
	case UNIT_TEST_NAME:
		ret = nil
	case DEVICE_WS_Z3030:
		ret = NewWsZ3030(deviceId, ms)
	case DEVICE_R1212:
		ret = NewR1212(deviceId, ms)
	default:
		ret = nil
	}
	return ret
}
