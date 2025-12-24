package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkOptimizedProxy_Throughput benchmarks proxy throughput.
func BenchmarkOptimizedProxy_Throughput(b *testing.B) {
	// Start mock Modbus server
	mockServer := startMockServer(b)
	defer mockServer.Close()

	// Create optimized proxy
	proxy := NewOptimizedProxy(OptimizedConfig{
		ID:                "bench-proxy",
		Name:              "Benchmark Proxy",
		ListenAddr:        "127.0.0.1:15020",
		TargetAddr:        mockServer.Addr().String(),
		MaxConns:          10000,
		EnableDNSCache:    true,
		EnableBufferPool:  true,
		EnableKeepAlive:   true,
		EnableTCPOptimize: true,
	})

	ctx := context.Background()
	if err := proxy.Start(ctx); err != nil {
		b.Fatalf("Failed to start proxy: %v", err)
	}
	defer proxy.Stop()

	// Wait for proxy to start
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sendModbusRequest(b, "127.0.0.1:15020")
		}
	})

	b.StopTimer()
	stats := proxy.Stats()
	b.ReportMetric(stats.RequestsPerSec, "req/sec")
	b.ReportMetric(stats.ErrorRate*100, "error%")
}

// BenchmarkOptimizedProxy_Latency benchmarks proxy latency.
func BenchmarkOptimizedProxy_Latency(b *testing.B) {
	mockServer := startMockServer(b)
	defer mockServer.Close()

	proxy := NewOptimizedProxy(OptimizedConfig{
		ID:                "bench-proxy",
		Name:              "Benchmark Proxy",
		ListenAddr:        "127.0.0.1:15021",
		TargetAddr:        mockServer.Addr().String(),
		MaxConns:          1000,
		EnableDNSCache:    true,
		EnableBufferPool:  true,
		EnableKeepAlive:   true,
		EnableTCPOptimize: true,
	})

	ctx := context.Background()
	if err := proxy.Start(ctx); err != nil {
		b.Fatalf("Failed to start proxy: %v", err)
	}
	defer proxy.Stop()

	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		sendModbusRequest(b, "127.0.0.1:15021")
		latency := time.Since(start)
		b.ReportMetric(float64(latency.Microseconds()), "μs/op")
	}
}

// BenchmarkOptimizedProxy_Concurrent benchmarks concurrent connections.
func BenchmarkOptimizedProxy_Concurrent(b *testing.B) {
	mockServer := startMockServer(b)
	defer mockServer.Close()

	proxy := NewOptimizedProxy(OptimizedConfig{
		ID:                "bench-proxy",
		Name:              "Benchmark Proxy",
		ListenAddr:        "127.0.0.1:15022",
		TargetAddr:        mockServer.Addr().String(),
		MaxConns:          10000,
		EnableDNSCache:    true,
		EnableBufferPool:  true,
		EnableKeepAlive:   true,
		EnableTCPOptimize: true,
	})

	ctx := context.Background()
	if err := proxy.Start(ctx); err != nil {
		b.Fatalf("Failed to start proxy: %v", err)
	}
	defer proxy.Stop()

	time.Sleep(100 * time.Millisecond)

	concurrencyLevels := []int{10, 100, 1000, 5000}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrent-%d", concurrency), func(b *testing.B) {
			b.ResetTimer()

			var wg sync.WaitGroup
			var successCount uint64
			var errorCount uint64

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < b.N/concurrency; j++ {
						if err := sendModbusRequestWithError("127.0.0.1:15022"); err != nil {
							atomic.AddUint64(&errorCount, 1)
						} else {
							atomic.AddUint64(&successCount, 1)
						}
					}
				}()
			}

			wg.Wait()

			total := successCount + errorCount
			if total > 0 {
				b.ReportMetric(float64(errorCount)/float64(total)*100, "error%")
			}
		})
	}
}

// BenchmarkBufferPooling benchmarks buffer pool vs allocation.
func BenchmarkBufferPooling(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		mockServer := startMockServer(b)
		defer mockServer.Close()

		proxy := NewOptimizedProxy(OptimizedConfig{
			ID:               "bench-proxy",
			Name:             "Benchmark Proxy",
			ListenAddr:       "127.0.0.1:15023",
			TargetAddr:       mockServer.Addr().String(),
			EnableBufferPool: true,
		})

		ctx := context.Background()
		proxy.Start(ctx)
		defer proxy.Stop()
		time.Sleep(100 * time.Millisecond)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sendModbusRequest(b, "127.0.0.1:15023")
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		mockServer := startMockServer(b)
		defer mockServer.Close()

		proxy := NewOptimizedProxy(OptimizedConfig{
			ID:               "bench-proxy",
			Name:             "Benchmark Proxy",
			ListenAddr:       "127.0.0.1:15024",
			TargetAddr:       mockServer.Addr().String(),
			EnableBufferPool: false,
		})

		ctx := context.Background()
		proxy.Start(ctx)
		defer proxy.Stop()
		time.Sleep(100 * time.Millisecond)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sendModbusRequest(b, "127.0.0.1:15024")
		}
	})
}

// startMockServer starts a mock Modbus TCP server.
func startMockServer(tb testing.TB) net.Listener {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		tb.Fatalf("Failed to start mock server: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go handleMockConnection(conn)
		}
	}()

	return listener
}

// handleMockConnection handles a mock Modbus connection.
func handleMockConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 260)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		// Echo back (simple mock response)
		if n >= 8 {
			// Modbus TCP response (echo request with data)
			response := make([]byte, n+2)
			copy(response, buf[:n])
			response[n] = 0x00   // Add dummy data
			response[n+1] = 0x00

			conn.Write(response)
		}
	}
}

// sendModbusRequest sends a Modbus read holding registers request.
func sendModbusRequest(tb testing.TB, addr string) {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		tb.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Modbus TCP Read Holding Registers (0x03)
	// Transaction ID: 0x0001
	// Protocol ID: 0x0000
	// Length: 0x0006
	// Unit ID: 0x01
	// Function: 0x03 (Read Holding Registers)
	// Start Address: 0x0000
	// Count: 0x000A (10 registers)
	request := []byte{
		0x00, 0x01, // Transaction ID
		0x00, 0x00, // Protocol ID
		0x00, 0x06, // Length
		0x01,       // Unit ID
		0x03,       // Function Code
		0x00, 0x00, // Start Address
		0x00, 0x0A, // Count
	}

	if _, err := conn.Write(request); err != nil {
		tb.Fatalf("Failed to write: %v", err)
	}

	// Read response
	response := make([]byte, 260)
	if _, err := conn.Read(response); err != nil && err != io.EOF {
		tb.Fatalf("Failed to read: %v", err)
	}
}

// sendModbusRequestWithError is like sendModbusRequest but returns error instead of failing.
func sendModbusRequestWithError(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	request := []byte{
		0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03,
		0x00, 0x00, 0x00, 0x0A,
	}

	if _, err := conn.Write(request); err != nil {
		return err
	}

	response := make([]byte, 260)
	if _, err := conn.Read(response); err != nil && err != io.EOF {
		return err
	}

	return nil
}
