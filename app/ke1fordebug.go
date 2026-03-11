//go:build debug

package app

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	numberOfVoltage                  = 3
	numberOfCurrent                  = 12
	numberOfPowerFactor              = 8
	numberOfEffectivePower           = 8
	numberOfIntegratedElectricEnergy = 8
)

type KE1 struct {
	ms                       *ModbusSerial
	voltage                  [numberOfVoltage]float64
	current                  [numberOfCurrent]float64
	frequency                float64
	powerFactor              [numberOfPowerFactor]float64
	effectivePower           [numberOfEffectivePower]float64
	integratedElectricEnergy [numberOfIntegratedElectricEnergy]float64
	id                       byte
}

var ke1 []KE1

func NewKE1(deviceId byte, ms *ModbusSerial) *KE1 {
	var ret *KE1

	for i := range ke1 {
		if ke1[i].id == deviceId {
			ret = &ke1[i]
			break
		}
	}

	if ret == nil {
		ret = new(KE1)
		ret.id = deviceId
		ret.ms = ms
		ke1 = append(ke1, *ret)
	}

	return ret
}

func (ke1 *KE1) GetId() byte {

	return ke1.id
}

func (ke1 *KE1) UpdateData() (ret error) {
	/* Get Voltage */
	//	results, err := ke1.ms.ReadHoldingRegisters(ke1.id, 0, 2*numberOfVoltage)
	results, err := ke1_csv_read(0, 2*numberOfVoltage)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "CSV file read error:%+v", err.Error())
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "KE1 Data Voltage : %+v", results)
	if err == nil {
		for i := 0; i < numberOfVoltage; i++ {
			value := ke1.convertUint32Value(results[i*4 : (i+1)*4])
			ke1.voltage[i] = float64(value) / 10
		}
	} else {
		ret = err
	}

	/* Get Current */
	//	results, err = ke1.ms.ReadHoldingRegisters(ke1.id+1, 0x0C, 2*numberOfCurrent)
	results, err = ke1_csv_read(0x0C, 2*numberOfCurrent)
	Logger.Writef(LOG_LEVEL_DEBUG, "KE1 Data Current : %+v", results)
	if err == nil {
		for i := 0; i < numberOfCurrent; i++ {
			value := ke1.convertUint32Value(results[i*4 : (i+1)*4])
			ke1.current[i] = float64(value) / 1000
		}
	} else {
		ret = err
	}

	/* Get Frequency */
	//	results, err = ke1.ms.ReadHoldingRegisters(ke1.id, 0x34, 2)
	results, err = ke1_csv_read(0x34, 2)
	Logger.Writef(LOG_LEVEL_DEBUG, "KE1 Data frequency : %+v", results)
	if err == nil {
		value := ke1.convertUint32Value(results)
		ke1.frequency = float64(value) / 10
	} else {
		ret = err
	}

	/* Get PowerFactor */
	//	results, err = ke1.ms.ReadHoldingRegisters(ke1.id+1, 0x24, 2*numberOfPowerFactor)
	results, err = ke1_csv_read(0x24, 2*numberOfPowerFactor)
	Logger.Writef(LOG_LEVEL_DEBUG, "KE1 Data PowerFactor : %+v", results)
	if err == nil {
		for i := 0; i < numberOfPowerFactor; i++ {
			value := ke1.convertInt32Value(results[i*4 : (i+1)*4])
			ke1.powerFactor[i] = float64(value) / 100
		}
	} else {
		ret = err
	}

	/* Get EffectivePower */
	//	results, err = ke1.ms.ReadHoldingRegisters(ke1.id+1, 0x38, 2*numberOfEffectivePower)
	results, err = ke1_csv_read(0x38, 2*numberOfEffectivePower)
	Logger.Writef(LOG_LEVEL_DEBUG, "KE1 Data EffectivePower : %+v", results)
	if err == nil {
		for i := 0; i < numberOfEffectivePower; i++ {
			value := ke1.convertInt32Value(results[i*4 : (i+1)*4])
			ke1.effectivePower[i] = float64(value) / 10
		}
	} else {
		ret = err
	}

	/* Get EffectiveIntegratedElectricEnergy */
	//	results, err = ke1.ms.ReadHoldingRegisters(ke1.id+1, 0x100, 2*numberOfIntegratedElectricEnergy)
	results, err = ke1_csv_read(0x100, 2*numberOfIntegratedElectricEnergy)
	Logger.Writef(LOG_LEVEL_DEBUG, "KE1 Data IntegratedElectricEnergy : %+v", results)
	if err == nil {
		for i := 0; i < numberOfIntegratedElectricEnergy; i++ {
			ke1.integratedElectricEnergy[i] = float64(ke1.convertUint32Value(results[i*4 : (i+1)*4]))
		}
	} else {
		ret = err
	}

	//	csv_write(ke1)
	Logger.Writef(LOG_LEVEL_DEBUG, "KE1 Data : %+v", ke1)
	return
}

func (ke1 *KE1) GetVoltage1(id uint16) (float64, error) {

	return ke1.voltage[0], nil
}

func (ke1 *KE1) GetVoltage2(id uint16) (float64, error) {

	return ke1.voltage[1], nil
}

func (ke1 *KE1) GetVoltage3(id uint16) (float64, error) {

	return ke1.voltage[2], nil
}

func (ke1 *KE1) GetCurrent1(id uint16) (float64, error) {
	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	return ke1.current[(id-1)*3], nil
}

func (ke1 *KE1) GetCurrent2(id uint16) (float64, error) {
	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	return ke1.current[(id-1)*3+1], nil
}

func (ke1 *KE1) GetCurrent3(id uint16) (float64, error) {
	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	return ke1.current[(id-1)*3+2], nil
}

func (ke1 *KE1) GetFrequency(id uint16) (float64, error) {
	return ke1.frequency, nil
}

func (ke1 *KE1) GetPowerFactor(id uint16) (float64, error) {

	var index uint16

	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	if id < 3 {
		index = id - 1
	} else {
		index = id + 1
	}

	return ke1.powerFactor[index], nil
}

func (ke1 *KE1) GetEffectivePower(id uint16) (float64, error) {
	var index uint16

	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	if id < 3 {
		index = id - 1
	} else {
		index = id + 1
	}

	return ke1.effectivePower[index], nil
}

func (ke1 *KE1) GetEffectiveIntegratedElectricEnergy(id uint16) (float64, error) {
	var index uint16

	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	if id < 3 {
		index = id - 1
	} else {
		index = id + 1
	}

	return ke1.integratedElectricEnergy[index], nil
}

func (ke1 *KE1) convertUint32Value(data []byte) uint32 {
	var value uint32

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &value)

	return value
}

func (ke1 *KE1) convertInt32Value(data []byte) int32 {
	var value int32

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &value)

	return value
}
