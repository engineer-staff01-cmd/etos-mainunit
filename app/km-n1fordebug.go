//go:build debug

package app

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	numberOfVoltage                  = 3
	numberOfCurrent                  = 3
	numberOfPowerFactor              = 1
	numberOfEffectivePower           = 1
	numberOfIntegratedElectricEnergy = 1
)

type KMN1 struct {
	ms                       *ModbusSerial
	voltage                  [numberOfVoltage]float64
	current                  [numberOfCurrent]float64
	frequency                float64
	powerFactor              [numberOfPowerFactor]float64
	effectivePower           [numberOfEffectivePower]float64
	integratedElectricEnergy [numberOfIntegratedElectricEnergy]float64
	id                       byte
}

var kmn1 []KMN1

func NewKMN1(deviceId byte, ms *ModbusSerial) *KMN1 {
	var ret *KMN1

	for i := range kmn1 {
		if kmn1[i].id == deviceId {
			ret = &kmn1[i]
			break
		}
	}

	if ret == nil {
		ret = new(KMN1)
		ret.id = deviceId
		ret.ms = ms
		kmn1 = append(kmn1, *ret)
	}

	return ret
}

func (kmn1 *KMN1) GetId() byte {

	return kmn1.id
}

func (kmn1 *KMN1) UpdateData() (ret error) {
	/* Get Voltage */
	//	results, err := kmn1.ms.ReadHoldingRegisters(kmn1.id, 0, 2*numberOfVoltage)
	results, err := kmn1_csv_read(0, 2*numberOfVoltage)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "CSV file read error:%+v", err.Error())
	}
	Logger.Writef(LOG_LEVEL_DEBUG, "KMN1 Data Voltage : %+v", results)
	if err == nil {
		for i := 0; i < numberOfVoltage; i++ {
			value := kmn1.convertUint32Value(results[i*4 : (i+1)*4])
			kmn1.voltage[i] = float64(value) / 10
		}
	} else {
		ret = err
	}

	/* Get Current */
	//	results, err = kmn1.ms.ReadHoldingRegisters(kmn1.id+1, 0x0C, 2*numberOfCurrent)
	results, err = kmn1_csv_read(0x0C, 2*numberOfCurrent)
	Logger.Writef(LOG_LEVEL_DEBUG, "KMN1 Data Current : %+v", results)
	if err == nil {
		for i := 0; i < numberOfCurrent; i++ {
			value := kmn1.convertUint32Value(results[i*4 : (i+1)*4])
			kmn1.current[i] = float64(value) / 1000
		}
	} else {
		ret = err
	}

	/* Get Frequency */
	//	results, err = kmn1.ms.ReadHoldingRegisters(kmn1.id, 0x34, 2)
	results, err = kmn1_csv_read(0x34, 2)
	Logger.Writef(LOG_LEVEL_DEBUG, "KMN1 Data frequency : %+v", results)
	if err == nil {
		value := kmn1.convertUint32Value(results)
		kmn1.frequency = float64(value) / 10
	} else {
		ret = err
	}

	/* Get PowerFactor */
	//	results, err = kmn1.ms.ReadHoldingRegisters(kmn1.id+1, 0x24, 2*numberOfPowerFactor)
	results, err = kmn1_csv_read(0x24, 2*numberOfPowerFactor)
	Logger.Writef(LOG_LEVEL_DEBUG, "KMN1 Data PowerFactor : %+v", results)
	if err == nil {
		for i := 0; i < numberOfPowerFactor; i++ {
			value := kmn1.convertInt32Value(results[i*4 : (i+1)*4])
			kmn1.powerFactor[i] = float64(value) / 100
		}
	} else {
		ret = err
	}

	/* Get EffectivePower */
	//	results, err = kmn1.ms.ReadHoldingRegisters(kmn1.id+1, 0x38, 2*numberOfEffectivePower)
	results, err = kmn1_csv_read(0x38, 2*numberOfEffectivePower)
	Logger.Writef(LOG_LEVEL_DEBUG, "KMN1 Data EffectivePower : %+v", results)
	if err == nil {
		for i := 0; i < numberOfEffectivePower; i++ {
			value := kmn1.convertInt32Value(results[i*4 : (i+1)*4])
			kmn1.effectivePower[i] = float64(value) / 10
		}
	} else {
		ret = err
	}

	/* Get EffectiveIntegratedElectricEnergy */
	//	results, err = kmn1.ms.ReadHoldingRegisters(kmn1.id+1, 0x100, 2*numberOfIntegratedElectricEnergy)
	results, err = kmn1_csv_read(0x100, 2*numberOfIntegratedElectricEnergy)
	Logger.Writef(LOG_LEVEL_DEBUG, "KMN1 Data IntegratedElectricEnergy : %+v", results)
	if err == nil {
		for i := 0; i < numberOfIntegratedElectricEnergy; i++ {
			kmn1.integratedElectricEnergy[i] = float64(kmn1.convertUint32Value(results[i*4 : (i+1)*4]))
		}
	} else {
		ret = err
	}

	//	csv_write(kmn1)
	Logger.Writef(LOG_LEVEL_DEBUG, "KMN1 Data : %+v", kmn1)
	return
}

func (kmn1 *KMN1) GetVoltage1(id uint16) (float64, error) {

	return kmn1.voltage[0], nil
}

func (kmn1 *KMN1) GetVoltage2(id uint16) (float64, error) {

	return kmn1.voltage[1], nil
}

func (kmn1 *KMN1) GetVoltage3(id uint16) (float64, error) {

	return kmn1.voltage[2], nil
}

func (kmn1 *KMN1) GetCurrent1(id uint16) (float64, error) {
	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	return kmn1.current[0], nil
}

func (kmn1 *KMN1) GetCurrent2(id uint16) (float64, error) {
	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	return kmn1.current[1], nil
}

func (kmn1 *KMN1) GetCurrent3(id uint16) (float64, error) {
	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	return kmn1.current[2], nil
}

func (kmn1 *KMN1) GetFrequency(id uint16) (float64, error) {
	return kmn1.frequency, nil
}

func (kmn1 *KMN1) GetPowerFactor(id uint16) (float64, error) {

	var index uint16

	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	if id < 3 {
		index = id - 1
	} else {
		index = id + 1
	}

	//	return kmn1.powerFactor[index], nil
	return kmn1.powerFactor[0], nil
}

func (kmn1 *KMN1) GetEffectivePower(id uint16) (float64, error) {
	var index uint16

	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	if id < 3 {
		index = id - 1
	} else {
		index = id + 1
	}

	//	return kmn1.effectivePower[index], nil
	return kmn1.effectivePower[0], nil
}

func (kmn1 *KMN1) GetEffectiveIntegratedElectricEnergy(id uint16) (float64, error) {
	var index uint16

	if id > 4 || id == 0 {
		return 0, fmt.Errorf("Error: %s\n", "id is illegal!")
	}

	if id < 3 {
		index = id - 1
	} else {
		index = id + 1
	}

	//	return kmn1.integratedElectricEnergy[index], nil
	return kmn1.integratedElectricEnergy[0], nil
}

func (kmn1 *KMN1) convertUint32Value(data []byte) uint32 {
	var value uint32

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &value)

	return value
}

func (kmn1 *KMN1) convertInt32Value(data []byte) int32 {
	var value int32

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &value)

	return value
}
