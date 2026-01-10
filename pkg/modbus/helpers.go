package modbus

import (
	"encoding/binary"
	"fmt"
)

const (
	// Modbus Function Codes
	FuncReadHoldingRegisters = 0x03
	FuncReadInputRegisters   = 0x04

	// Modbus Exceptions
	ExceptionIllegalFunction = 0x01
	ExceptionIllegalDataAddress = 0x02
	ExceptionIllegalDataValue = 0x03
	ExceptionSlaveDeviceFailure = 0x04
)

// IsReadRequest checks if the frame is a Read Holding/Input Registers request.
func IsReadRequest(frame []byte) bool {
	if len(frame) < 12 {
		return false
	}
	fc := frame[7]
	return fc == FuncReadHoldingRegisters || fc == FuncReadInputRegisters
}

// ParseReadRequest extracts parameters from a read request frame.
func ParseReadRequest(frame []byte) (txID uint16, unitID uint8, fc uint8, startAddr uint16, quantity uint16, err error) {
	if len(frame) < 12 {
		return 0, 0, 0, 0, 0, fmt.Errorf("frame too short")
	}
	txID = binary.BigEndian.Uint16(frame[0:2])
	unitID = frame[6]
	fc = frame[7]
	startAddr = binary.BigEndian.Uint16(frame[8:10])
	quantity = binary.BigEndian.Uint16(frame[10:12])
	return
}

// CreateReadRequest constructs a Modbus TCP read request frame.
func CreateReadRequest(txID uint16, unitID uint8, fc uint8, startAddr uint16, quantity uint16) []byte {
	frame := make([]byte, 12)
	binary.BigEndian.PutUint16(frame[0:2], txID)
	// Protocol ID is 0
	frame[2] = 0
	frame[3] = 0
	// Length is 6 (UnitID + FC + StartAddr + Quantity)
	binary.BigEndian.PutUint16(frame[4:6], 6)
	frame[6] = unitID
	frame[7] = fc
	binary.BigEndian.PutUint16(frame[8:10], startAddr)
	binary.BigEndian.PutUint16(frame[10:12], quantity)
	return frame
}

// ParseReadResponse extracts data from a read response frame.
func ParseReadResponse(frame []byte) ([]byte, error) {
	if len(frame) < 9 {
		return nil, fmt.Errorf("frame too short")
	}
	// length := binary.BigEndian.Uint16(frame[4:6])
	fc := frame[7]
	if fc >= 0x80 {
		return nil, fmt.Errorf("modbus exception: %d", frame[8])
	}
	byteCount := frame[8]
	if len(frame) < 9+int(byteCount) {
		return nil, fmt.Errorf("frame data incomplete")
	}
	return frame[9 : 9+byteCount], nil
}

// CreateReadResponse constructs a Modbus TCP read response frame.
func CreateReadResponse(txID uint16, unitID uint8, fc uint8, data []byte) []byte {
	byteCount := len(data)
	length := 3 + byteCount // UnitID(1) + FC(1) + ByteCount(1) + Data(N)
	frame := make([]byte, 6+length)

	binary.BigEndian.PutUint16(frame[0:2], txID)
	frame[2] = 0
	frame[3] = 0
	binary.BigEndian.PutUint16(frame[4:6], uint16(length))
	frame[6] = unitID
	frame[7] = fc
	frame[8] = uint8(byteCount)
	copy(frame[9:], data)
	return frame
}

// ExceptionResponse constructs a Modbus TCP exception response.
func ExceptionResponse(txID uint16, unitID uint8, fc uint8, exceptionCode uint8) []byte {
	length := 3 // UnitID + FC + ExceptionCode
	frame := make([]byte, 6+length)

	binary.BigEndian.PutUint16(frame[0:2], txID)
	frame[2] = 0
	frame[3] = 0
	binary.BigEndian.PutUint16(frame[4:6], uint16(length))
	frame[6] = unitID
	frame[7] = fc | 0x80
	frame[8] = exceptionCode
	return frame
}
