package app

import (
	"sync"
	"time"

	"github.com/goburrow/modbus"
)

type ModbusSerial struct {
	mu      sync.Mutex
	handler *modbus.RTUClientHandler
}

const modbusSerialDevice_A9E = "/dev/ttyrpmsg1"
const modbusSerialDevice_G3L = "/dev/ttymxc1"

var ms *ModbusSerial

func NewModbusSerial() *ModbusSerial {

	if ms == nil {
		ms = new(ModbusSerial)

		modbusSerialDevice := ""
		if MODEL == "A9E" {
			modbusSerialDevice = modbusSerialDevice_A9E
		} else if MODEL == "G3L" {
			modbusSerialDevice = modbusSerialDevice_G3L
		} else {
			panic("MODEL is unknown")
		}

		ms.handler = modbus.NewRTUClientHandler(modbusSerialDevice)
		ms.handler.BaudRate = 38400
		ms.handler.DataBits = 8
		ms.handler.Parity = "N"
		ms.handler.StopBits = 2
		ms.handler.SlaveId = 1
		//ms.handler.Timeout = 300 * time.Millisecond
		ms.handler.Timeout = 400 * time.Millisecond
	}
	return ms
}

func (ms *ModbusSerial) ReadCoils(deviceID byte, address, quantity uint16, stopBits int) ([]byte, error) {

	ms.mu.Lock()
	ms.handler.SlaveId = deviceID
	ms.handler.StopBits = stopBits
	err := ms.handler.Connect()

	client := modbus.NewClient(ms.handler)
	results, err := client.ReadCoils(address, quantity)
	ms.handler.Close()
	ms.mu.Unlock()

	return results, err
}

func (ms *ModbusSerial) ReadHoldingRegisters(deviceID byte, address, quantity uint16, stopBits int) ([]byte, error) {

	ms.mu.Lock()
	ms.handler.SlaveId = deviceID
	ms.handler.StopBits = stopBits
	err := ms.handler.Connect()

	client := modbus.NewClient(ms.handler)
	results, err := client.ReadHoldingRegisters(address, quantity)
	ms.handler.Close()
	ms.mu.Unlock()

	return results, err
}

func (ms *ModbusSerial) ReadDiscreteInputs(deviceID byte, address, quantity uint16, stopBits int) ([]byte, error) {

	ms.mu.Lock()
	ms.handler.SlaveId = deviceID
	ms.handler.StopBits = stopBits
	err := ms.handler.Connect()

	client := modbus.NewClient(ms.handler)
	results, err := client.ReadDiscreteInputs(address, quantity)
	ms.handler.Close()
	ms.mu.Unlock()

	return results, err
}

func (ms *ModbusSerial) ReadInputRegisters(deviceID byte, address, quantity uint16, stopBits int) ([]byte, error) {

	ms.mu.Lock()
	ms.handler.SlaveId = deviceID
	ms.handler.StopBits = stopBits
	err := ms.handler.Connect()

	client := modbus.NewClient(ms.handler)
	results, err := client.ReadInputRegisters(address, quantity)
	ms.handler.Close()
	ms.mu.Unlock()

	return results, err
}

func (ms *ModbusSerial) WriteSingleCoil(device_id byte, address, value uint16, stopBits int) ([]byte, error) {

	ms.mu.Lock()
	ms.handler.SlaveId = device_id
	ms.handler.StopBits = stopBits
	err := ms.handler.Connect()

	client := modbus.NewClient(ms.handler)
	results, err := client.WriteSingleCoil(address, value)
	ms.handler.Close()
	ms.mu.Unlock()

	return results, err
}

func (ms *ModbusSerial) WriteSingleRegister(device_id byte, address, value uint16, stopBits int) ([]byte, error) {

	ms.mu.Lock()
	ms.handler.SlaveId = device_id
	ms.handler.StopBits = stopBits
	err := ms.handler.Connect()

	client := modbus.NewClient(ms.handler)
	results, err := client.WriteSingleRegister(address, value)
	ms.handler.Close()
	ms.mu.Unlock()

	return results, err
}

func (ms *ModbusSerial) WriteMultipleRegister(device_id byte, address uint16, quantity uint16, value []byte, stopBits int) ([]byte, error) {

	ms.mu.Lock()
	ms.handler.SlaveId = device_id
	ms.handler.StopBits = stopBits
	err := ms.handler.Connect()

	client := modbus.NewClient(ms.handler)
	//	results, err := client.WriteSingleRegister(address, value)
	results, err := client.WriteMultipleRegisters(address, quantity, value)
	ms.handler.Close()
	ms.mu.Unlock()

	return results, err
}
