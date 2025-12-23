package modbus

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

// MBAPHeaderLength is the length of the Modbus Application Protocol header prefix (Transaction ID, Protocol ID, Length).
const MBAPHeaderLength = 6

// MaxFrameSize is the maximum Modbus TCP frame size
const MaxFrameSize = 260

// Buffer pool for frame allocations
var framePool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, MaxFrameSize)
		return &buf
	},
}

// getBuffer gets a buffer from the pool
func getBuffer() *[]byte {
	return framePool.Get().(*[]byte)
}

// putBuffer returns a buffer to the pool
func putBuffer(buf *[]byte) {
	framePool.Put(buf)
}

// ReadFrame reads a full Modbus TCP frame from the reader.
// Uses a pooled buffer internally to reduce allocations.
func ReadFrame(r io.Reader) ([]byte, error) {
	// Get a buffer from the pool
	bufPtr := getBuffer()
	defer putBuffer(bufPtr)
	buf := *bufPtr

	// Read the first 6 bytes (header) into the pooled buffer
	if _, err := io.ReadFull(r, buf[:MBAPHeaderLength]); err != nil {
		return nil, err
	}

	// Parse the length field (bytes 4 and 5)
	length := binary.BigEndian.Uint16(buf[4:6])

	// Sanity check for length
	if length == 0 || length > 300 {
		return nil, fmt.Errorf("invalid modbus length: %d", length)
	}

	// Read the payload into the pooled buffer
	totalLen := MBAPHeaderLength + int(length)
	if totalLen > MaxFrameSize {
		return nil, fmt.Errorf("frame too large: %d bytes", totalLen)
	}

	if _, err := io.ReadFull(r, buf[MBAPHeaderLength:totalLen]); err != nil {
		return nil, err
	}

	// Return a copy of the exact frame size (only allocation)
	frame := make([]byte, totalLen)
	copy(frame, buf[:totalLen])
	return frame, nil
}
