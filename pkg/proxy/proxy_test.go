package proxy

import (
	"modbusproxy/pkg/logger"
	"modbusproxy/pkg/modbus"
	"net"
	"testing"
)

func TestSplitLogic_Integration(t *testing.T) {
	// Create a dummy target server
	targetListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock target: %v", err)
	}
	defer targetListener.Close()
	targetAddr := targetListener.Addr().String()

	go func() {
		for {
			conn, err := targetListener.Accept()
			if err != nil {
				return
			}
			go handleMockTarget(conn)
		}
	}()

	// Create proxy
	// Use a random port for proxy
	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start proxy listener: %v", err)
	}
	proxyAddr := proxyListener.Addr().String()
	proxyListener.Close() // We just wanted a free port, ProxyInstance will listen itself

	l, _ := logger.NewLogger("/dev/null", 100)
	p := NewProxyInstance("test", "test", proxyAddr, targetAddr, 10, l, nil) // MaxReadSize = 10 registers

	err = p.Start()
	if err != nil {
		t.Fatalf("Failed to start proxy: %v", err)
	}
	defer p.Stop()

	// Connect to proxy
	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("Failed to connect to proxy: %v", err)
	}
	defer conn.Close()

	// 1. Test Request within limit (5 registers)
	req1 := modbus.CreateReadRequest(1, 1, 3, 0, 5)
	_, err = conn.Write(req1)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	resp1, err := modbus.ReadFrame(conn)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Verify response 1
	if len(resp1) < 9 {
		t.Errorf("Response too short")
	}

	// 2. Test Request exceeding limit (25 registers). Max is 10.
	// Should split into 10, 10, 5.
	req2 := modbus.CreateReadRequest(2, 1, 3, 0, 25)
	_, err = conn.Write(req2)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	resp2, err := modbus.ReadFrame(conn)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Verify response 2
	// 25 registers * 2 = 50 bytes data.
	// Frame: MBAP(6) + Unit(1) + FC(1) + Len(1) + Data(50) = 59 bytes.
	if len(resp2) != 9+50 {
		t.Errorf("Split response length mismatch. Got %d, want %d", len(resp2), 59)
	}

	// Verify data content (Mock target returns 0xAA for all bytes)
	data := resp2[9:]
	for i, b := range data {
		if b != 0xAA {
			t.Errorf("Byte %d mismatch: got %x, want AA", i, b)
		}
	}
}

func handleMockTarget(conn net.Conn) {
	defer conn.Close()
	for {
		frame, err := modbus.ReadFrame(conn)
		if err != nil {
			return
		}

		// Parse request
		txID, unitID, fc, _, quantity, err := modbus.ParseReadRequest(frame)
		if err != nil {
			return
		}

		// Generate response with 0xAA
		data := make([]byte, quantity*2)
		for i := range data {
			data[i] = 0xAA
		}

		resp := modbus.CreateReadResponse(txID, unitID, fc, data)
		conn.Write(resp)
	}
}
