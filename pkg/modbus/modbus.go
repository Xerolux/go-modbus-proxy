package modbus

import (
	"encoding/binary"
	"fmt"
	"io"
)

// MBAPHeaderLength is the length of the Modbus Application Protocol header prefix (Transaction ID, Protocol ID, Length).
const MBAPHeaderLength = 6

// ReadFrame reads a full Modbus TCP frame from the reader.
func ReadFrame(r io.Reader) ([]byte, error) {
	// Read the first 6 bytes to get the length
	header := make([]byte, MBAPHeaderLength)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	// Parse the length field (bytes 4 and 5)
	length := binary.BigEndian.Uint16(header[4:6])

	// The length field represents the number of bytes following.
	// We need to read 'length' bytes.
	// Sanity check for length (Modbus PDU max is usually 253 + 1 Unit ID = 254, but let's allow a bit more just in case, standard says 260 bytes max total frame)
	if length == 0 || length > 300 {
		return nil, fmt.Errorf("invalid modbus length: %d", length)
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	// Combine header and payload
	frame := append(header, payload...)
	return frame, nil
}
