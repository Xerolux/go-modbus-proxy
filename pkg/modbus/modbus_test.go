package modbus

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestReadFrame(t *testing.T) {
	// Construct a valid frame
	// TransID(2) + ProtoID(2) + Length(2) + UnitID(1) + PDU(n)
	transID := uint16(123)
	protoID := uint16(0)
	unitID := uint8(1)
	pdu := []byte{0x03, 0x00, 0x01, 0x00, 0x01} // Read Holding Registers
	length := uint16(1 + len(pdu))              // UnitID + PDU

	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, transID)
	_ = binary.Write(buf, binary.BigEndian, protoID)
	_ = binary.Write(buf, binary.BigEndian, length)
	_ = buf.WriteByte(unitID)
	_, _ = buf.Write(pdu)

	frame, err := ReadFrame(buf)
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}

	expectedLen := 6 + int(length)
	if len(frame) != expectedLen {
		t.Errorf("Expected frame length %d, got %d", expectedLen, len(frame))
	}

	if frame[6] != unitID {
		t.Errorf("Expected UnitID %d, got %d", unitID, frame[6])
	}
}

func TestReadFrameShort(t *testing.T) {
	buf := bytes.NewBuffer([]byte{0x00, 0x01}) // Too short for header
	_, err := ReadFrame(buf)
	if err == nil {
		t.Fatal("Expected error for short buffer, got nil")
	}
}
