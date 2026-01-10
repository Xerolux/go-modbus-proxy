package modbus

import (
	"bytes"
	"testing"
)

func TestReadRequestHelpers(t *testing.T) {
	txID := uint16(12345)
	unitID := uint8(1)
	fc := uint8(3)
	startAddr := uint16(100)
	quantity := uint16(10)

	// Test CreateReadRequest
	frame := CreateReadRequest(txID, unitID, fc, startAddr, quantity)

	// Verify length
	if len(frame) != 12 {
		t.Errorf("Expected frame length 12, got %d", len(frame))
	}

	// Verify content
	if !IsReadRequest(frame) {
		t.Errorf("IsReadRequest returned false")
	}

	// Test ParseReadRequest
	parsedTxID, parsedUnitID, parsedFC, parsedStart, parsedQty, err := ParseReadRequest(frame)
	if err != nil {
		t.Fatalf("ParseReadRequest failed: %v", err)
	}

	if parsedTxID != txID {
		t.Errorf("TxID mismatch: got %d, want %d", parsedTxID, txID)
	}
	if parsedUnitID != unitID {
		t.Errorf("UnitID mismatch: got %d, want %d", parsedUnitID, unitID)
	}
	if parsedFC != fc {
		t.Errorf("FC mismatch: got %d, want %d", parsedFC, fc)
	}
	if parsedStart != startAddr {
		t.Errorf("StartAddr mismatch: got %d, want %d", parsedStart, startAddr)
	}
	if parsedQty != quantity {
		t.Errorf("Quantity mismatch: got %d, want %d", parsedQty, quantity)
	}
}

func TestReadResponseHelpers(t *testing.T) {
	txID := uint16(12345)
	unitID := uint8(1)
	fc := uint8(3)
	data := []byte{0x00, 0x01, 0x00, 0x02} // 2 registers

	// Test CreateReadResponse
	frame := CreateReadResponse(txID, unitID, fc, data)

	// Verify length (Header 6 + UnitID 1 + FC 1 + ByteCount 1 + Data 4 = 13)
	expectedLen := 6 + 1 + 1 + 1 + 4
	if len(frame) != expectedLen {
		t.Errorf("Expected frame length %d, got %d", expectedLen, len(frame))
	}

	// Test ParseReadResponse
	parsedData, err := ParseReadResponse(frame)
	if err != nil {
		t.Fatalf("ParseReadResponse failed: %v", err)
	}

	if !bytes.Equal(parsedData, data) {
		t.Errorf("Data mismatch: got %v, want %v", parsedData, data)
	}
}

func TestExceptionResponse(t *testing.T) {
	txID := uint16(12345)
	unitID := uint8(1)
	fc := uint8(3)
	code := uint8(ExceptionIllegalDataAddress)

	frame := ExceptionResponse(txID, unitID, fc, code)

	// Header 6 + UnitID 1 + FC 1 + Code 1 = 9
	if len(frame) != 9 {
		t.Errorf("Expected frame length 9, got %d", len(frame))
	}

	// Verify FC has high bit set
	if frame[7] != fc|0x80 {
		t.Errorf("Exception FC mismatch: got %02x, want %02x", frame[7], fc|0x80)
	}

	// Verify parse returns error
	_, err := ParseReadResponse(frame)
	if err == nil {
		t.Errorf("Expected error for exception frame, got nil")
	}
}
