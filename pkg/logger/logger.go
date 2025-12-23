package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// LogLevel defines log levels.
type LogLevel string

const (
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	DEBUG LogLevel = "DEBUG"
)

// LogEntry represents a structured log line.
type LogEntry struct {
	Timestamp string   `json:"timestamp"`
	Level     LogLevel `json:"level"`
	ProxyID   string   `json:"proxy_id,omitempty"`
	Message   string   `json:"message"`
}

// Logger manages logging.
type Logger struct {
	mu          sync.Mutex
	file        *os.File
	ringBuffer  []LogEntry
	ringSize    int
	subscribers map[chan LogEntry]struct{}
}

// NewLogger creates a new logger.
func NewLogger(filePath string, bufferSize int) (*Logger, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:        f,
		ringBuffer:  make([]LogEntry, 0, bufferSize),
		ringSize:    bufferSize,
		subscribers: make(map[chan LogEntry]struct{}),
	}, nil
}

// Log writes a log entry.
func (l *Logger) Log(level LogLevel, proxyID, msg string) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		ProxyID:   proxyID,
		Message:   msg,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 1. Write to file
	if jsonBytes, err := json.Marshal(entry); err == nil {
		_, _ = l.file.Write(jsonBytes)
		_, _ = l.file.WriteString("\n")
	}

	// 2. Add to ring buffer
	if len(l.ringBuffer) >= l.ringSize {
		// Shift
		l.ringBuffer = l.ringBuffer[1:]
	}
	l.ringBuffer = append(l.ringBuffer, entry)

	// 3. Broadcast to subscribers
	for ch := range l.subscribers {
		select {
		case ch <- entry:
		default:
			// Drop if channel full to avoid blocking logger
		}
	}
	
	// Print to stdout for debug
	fmt.Printf("[%s] [%s] %s: %s\n", entry.Timestamp, entry.Level, entry.ProxyID, entry.Message)
}

// Subscribe returns a channel for live logs.
func (l *Logger) Subscribe() chan LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	ch := make(chan LogEntry, 100)
	l.subscribers[ch] = struct{}{}
	return ch
}

// Unsubscribe removes a subscriber.
func (l *Logger) Unsubscribe(ch chan LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.subscribers, ch)
	close(ch)
}

// GetRecent returns recent logs.
func (l *Logger) GetRecent(limit int) []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if limit > len(l.ringBuffer) {
		limit = len(l.ringBuffer)
	}
	
	// Return a copy
	out := make([]LogEntry, limit)
	start := len(l.ringBuffer) - limit
	copy(out, l.ringBuffer[start:])
	return out
}

func (l *Logger) Info(proxyID, msg string) {
	l.Log(INFO, proxyID, msg)
}

func (l *Logger) Error(proxyID, msg string) {
	l.Log(ERROR, proxyID, msg)
}

// Close closes the logger file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}
