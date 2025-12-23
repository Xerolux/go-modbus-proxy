package logger

import (
	"os"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	testFile := "test_logger.log"
	defer os.Remove(testFile)

	logger, err := NewLogger(testFile, 10)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	if logger == nil {
		t.Fatal("Logger is nil")
	}
}

func TestLogEntry(t *testing.T) {
	testFile := "test_log_entry.log"
	defer os.Remove(testFile)

	logger, _ := NewLogger(testFile, 10)
	defer logger.Close()

	logger.Info("TEST", "Test message")
	logger.Error("TEST", "Error message")

	logs := logger.GetRecent(10)
	if len(logs) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(logs))
	}

	if logs[0].Level != INFO {
		t.Errorf("Expected INFO level, got %s", logs[0].Level)
	}
	if logs[1].Level != ERROR {
		t.Errorf("Expected ERROR level, got %s", logs[1].Level)
	}
}

func TestRingBuffer(t *testing.T) {
	testFile := "test_ring_buffer.log"
	defer os.Remove(testFile)

	logger, _ := NewLogger(testFile, 3)
	defer logger.Close()

	// Add more entries than buffer size
	logger.Info("TEST", "Message 1")
	logger.Info("TEST", "Message 2")
	logger.Info("TEST", "Message 3")
	logger.Info("TEST", "Message 4")

	logs := logger.GetRecent(10)
	if len(logs) != 3 {
		t.Errorf("Expected 3 log entries (buffer size), got %d", len(logs))
	}

	// Should have messages 2, 3, 4 (1 was dropped)
	if logs[0].Message != "Message 2" {
		t.Errorf("Expected 'Message 2', got '%s'", logs[0].Message)
	}
}

func TestSubscribe(t *testing.T) {
	testFile := "test_subscribe.log"
	defer os.Remove(testFile)

	logger, _ := NewLogger(testFile, 10)
	defer logger.Close()

	ch := logger.Subscribe()
	defer logger.Unsubscribe(ch)

	// Log a message
	go logger.Info("TEST", "Subscribe test")

	// Wait for message
	select {
	case entry := <-ch:
		if entry.Message != "Subscribe test" {
			t.Errorf("Expected 'Subscribe test', got '%s'", entry.Message)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for log entry")
	}
}

func TestClose(t *testing.T) {
	testFile := "test_close.log"
	defer os.Remove(testFile)

	logger, _ := NewLogger(testFile, 10)
	logger.Info("TEST", "Before close")

	err := logger.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Multiple closes should not panic
	err = logger.Close()
	if err != nil {
		t.Errorf("Second close failed: %v", err)
	}
}
