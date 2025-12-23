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
	ringBuffer  []LogEntry // Fixed-size circular buffer
	ringSize    int
	ringHead    int        // Write position
	ringCount   int        // Number of entries in buffer
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
		ringBuffer:  make([]LogEntry, bufferSize), // Pre-allocate full size
		ringSize:    bufferSize,
		ringHead:    0,
		ringCount:   0,
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

	// 2. Add to ring buffer (circular buffer, no allocations)
	l.ringBuffer[l.ringHead] = entry
	l.ringHead = (l.ringHead + 1) % l.ringSize
	if l.ringCount < l.ringSize {
		l.ringCount++
	}

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

	if limit > l.ringCount {
		limit = l.ringCount
	}
	if limit == 0 {
		return []LogEntry{}
	}

	out := make([]LogEntry, limit)

	// Calculate start position (most recent entries)
	// If buffer is not full, entries are at [0:ringCount]
	// If buffer is full, entries wrap around with head pointing to oldest
	if l.ringCount < l.ringSize {
		// Buffer not full yet, just copy the last 'limit' entries
		start := l.ringCount - limit
		copy(out, l.ringBuffer[start:l.ringCount])
	} else {
		// Buffer is full, need to handle wrap-around
		// Oldest entry is at ringHead, newest is at ringHead-1
		start := (l.ringHead - limit + l.ringSize) % l.ringSize
		if start < l.ringHead {
			// No wrap-around needed
			copy(out, l.ringBuffer[start:l.ringHead])
		} else {
			// Wrap-around: copy from start to end, then from 0 to ringHead
			n := copy(out, l.ringBuffer[start:])
			copy(out[n:], l.ringBuffer[:l.ringHead])
		}
	}

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
