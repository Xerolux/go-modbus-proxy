package bigdata

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// MQTTBridge bridges Modbus data to MQTT for IoT integration.
type MQTTBridge struct {
	mu sync.RWMutex

	config   MQTTConfig
	enabled  bool
	stopChan chan struct{}

	// Message queue
	queue     chan *MQTTMessage
	queueSize int

	// Stats
	published uint64
	errors    uint64
}

// MQTTConfig holds MQTT configuration.
type MQTTConfig struct {
	Enabled       bool          `json:"enabled"`
	Broker        string        `json:"broker"`         // MQTT broker URL (e.g., tcp://localhost:1883)
	ClientID      string        `json:"client_id"`      // Client ID
	Username      string        `json:"username"`       // Username for authentication
	Password      string        `json:"password"`       // Password for authentication
	TopicPrefix   string        `json:"topic_prefix"`   // Topic prefix (e.g., modbridge/)
	QoS           byte          `json:"qos"`            // Quality of Service (0, 1, 2)
	Retained      bool          `json:"retained"`       // Retain messages
	CleanSession  bool          `json:"clean_session"`  // Clean session flag
	QueueSize     int           `json:"queue_size"`     // Internal queue size
	PublishInterval time.Duration `json:"publish_interval"` // Publish interval
}

// DefaultMQTTConfig returns default MQTT configuration.
func DefaultMQTTConfig() MQTTConfig {
	return MQTTConfig{
		Enabled:         false,
		Broker:          "tcp://localhost:1883",
		ClientID:        "modbridge",
		TopicPrefix:     "modbridge/",
		QoS:             1,
		Retained:        false,
		CleanSession:    true,
		QueueSize:       1000,
		PublishInterval: 1 * time.Second,
	}
}

// MQTTMessage represents an MQTT message.
type MQTTMessage struct {
	Topic    string
	Payload  []byte
	QoS      byte
	Retained bool
}

// ModbusDataMessage represents Modbus data for MQTT.
type ModbusDataMessage struct {
	DeviceID     string    `json:"device_id"`
	ProxyID      string    `json:"proxy_id"`
	ProxyName    string    `json:"proxy_name"`
	Timestamp    time.Time `json:"timestamp"`
	FunctionCode uint8     `json:"function_code"`
	Address      uint16    `json:"address"`
	Count        uint16    `json:"count"`
	Value        interface{} `json:"value"`
	DataType     string    `json:"data_type"` // uint16, int16, float32, etc.
	Unit         string    `json:"unit,omitempty"`
}

// NewMQTTBridge creates a new MQTT bridge.
func NewMQTTBridge(config MQTTConfig) *MQTTBridge {
	return &MQTTBridge{
		config:    config,
		enabled:   config.Enabled,
		stopChan:  make(chan struct{}),
		queue:     make(chan *MQTTMessage, config.QueueSize),
		queueSize: config.QueueSize,
	}
}

// Start starts the MQTT bridge.
func (mb *MQTTBridge) Start(ctx context.Context) {
	if !mb.enabled {
		log.Println("[MQTT] Disabled, not starting")
		return
	}

	log.Printf("[MQTT] Starting bridge (Broker: %s, ClientID: %s)", mb.config.Broker, mb.config.ClientID)

	// Start message processor
	go mb.processMessages(ctx)
}

// Stop stops the MQTT bridge.
func (mb *MQTTBridge) Stop() {
	close(mb.stopChan)
	close(mb.queue)
}

// PublishModbusData publishes Modbus data to MQTT.
func (mb *MQTTBridge) PublishModbusData(data ModbusDataMessage) error {
	if !mb.enabled {
		return nil
	}

	data.Timestamp = time.Now()

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Topic structure: modbridge/{proxy_id}/devices/{device_id}/data
	topic := fmt.Sprintf("%sproxies/%s/devices/%s/data",
		mb.config.TopicPrefix, data.ProxyID, data.DeviceID)

	msg := &MQTTMessage{
		Topic:    topic,
		Payload:  payload,
		QoS:      mb.config.QoS,
		Retained: mb.config.Retained,
	}

	return mb.enqueue(msg)
}

// PublishRegisterValue publishes a single register value.
func (mb *MQTTBridge) PublishRegisterValue(proxyID, deviceID string, address uint16, value interface{}, dataType, unit string) error {
	if !mb.enabled {
		return nil
	}

	data := ModbusDataMessage{
		DeviceID:  deviceID,
		ProxyID:   proxyID,
		Timestamp: time.Now(),
		Address:   address,
		Value:     value,
		DataType:  dataType,
		Unit:      unit,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Topic structure: modbridge/{proxy_id}/devices/{device_id}/registers/{address}
	topic := fmt.Sprintf("%sproxies/%s/devices/%s/registers/%d",
		mb.config.TopicPrefix, proxyID, deviceID, address)

	msg := &MQTTMessage{
		Topic:    topic,
		Payload:  payload,
		QoS:      mb.config.QoS,
		Retained: true, // Retain register values
	}

	return mb.enqueue(msg)
}

// PublishProxyStatus publishes proxy status.
func (mb *MQTTBridge) PublishProxyStatus(proxyID, proxyName, status string, activeConnections int64) error {
	if !mb.enabled {
		return nil
	}

	statusMsg := map[string]interface{}{
		"proxy_id":           proxyID,
		"proxy_name":         proxyName,
		"status":             status,
		"active_connections": activeConnections,
		"timestamp":          time.Now(),
	}

	payload, err := json.Marshal(statusMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	// Topic: modbridge/{proxy_id}/status
	topic := fmt.Sprintf("%sproxies/%s/status", mb.config.TopicPrefix, proxyID)

	msg := &MQTTMessage{
		Topic:    topic,
		Payload:  payload,
		QoS:      mb.config.QoS,
		Retained: true,
	}

	return mb.enqueue(msg)
}

// PublishSystemMetrics publishes system-wide metrics.
func (mb *MQTTBridge) PublishSystemMetrics(metrics map[string]interface{}) error {
	if !mb.enabled {
		return nil
	}

	metrics["timestamp"] = time.Now()

	payload, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	topic := fmt.Sprintf("%ssystem/metrics", mb.config.TopicPrefix)

	msg := &MQTTMessage{
		Topic:    topic,
		Payload:  payload,
		QoS:      0, // Use QoS 0 for frequent metrics
		Retained: false,
	}

	return mb.enqueue(msg)
}

// enqueue adds a message to the queue.
func (mb *MQTTBridge) enqueue(msg *MQTTMessage) error {
	select {
	case mb.queue <- msg:
		return nil
	default:
		mb.mu.Lock()
		mb.errors++
		mb.mu.Unlock()
		return fmt.Errorf("queue full (%d messages)", mb.queueSize)
	}
}

// processMessages processes messages from the queue.
func (mb *MQTTBridge) processMessages(ctx context.Context) {
	ticker := time.NewTicker(mb.config.PublishInterval)
	defer ticker.Stop()

	batch := make([]*MQTTMessage, 0, 10)

	for {
		select {
		case <-ctx.Done():
			mb.flushBatch(batch)
			return

		case <-mb.stopChan:
			mb.flushBatch(batch)
			return

		case msg, ok := <-mb.queue:
			if !ok {
				mb.flushBatch(batch)
				return
			}

			batch = append(batch, msg)

			// Flush batch periodically
			if len(batch) >= 10 {
				mb.flushBatch(batch)
				batch = make([]*MQTTMessage, 0, 10)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				mb.flushBatch(batch)
				batch = make([]*MQTTMessage, 0, 10)
			}
		}
	}
}

// flushBatch publishes a batch of messages.
func (mb *MQTTBridge) flushBatch(batch []*MQTTMessage) {
	if len(batch) == 0 {
		return
	}

	for _, msg := range batch {
		if err := mb.publishMessage(msg); err != nil {
			log.Printf("[MQTT] Failed to publish message to %s: %v", msg.Topic, err)
			mb.mu.Lock()
			mb.errors++
			mb.mu.Unlock()
		} else {
			mb.mu.Lock()
			mb.published++
			mb.mu.Unlock()
		}
	}
}

// publishMessage publishes a single message to MQTT.
func (mb *MQTTBridge) publishMessage(msg *MQTTMessage) error {
	// In a real implementation, this would:
	// 1. Use paho.mqtt.golang library
	// 2. Connect to broker (if not connected)
	// 3. Publish message
	// 4. Handle reconnection

	// For now, just log
	log.Printf("[MQTT] Would publish to %s: %d bytes", msg.Topic, len(msg.Payload))

	// Example of what the real implementation would look like:
	/*
		opts := mqtt.NewClientOptions()
		opts.AddBroker(mb.config.Broker)
		opts.SetClientID(mb.config.ClientID)
		opts.SetUsername(mb.config.Username)
		opts.SetPassword(mb.config.Password)
		opts.SetCleanSession(mb.config.CleanSession)

		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			return token.Error()
		}
		defer client.Disconnect(250)

		token := client.Publish(msg.Topic, msg.QoS, msg.Retained, msg.Payload)
		token.Wait()
		return token.Error()
	*/

	return nil
}

// GetStats returns bridge statistics.
func (mb *MQTTBridge) GetStats() (published, errors uint64, queueLen int) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.published, mb.errors, len(mb.queue)
}

// Subscribe subscribes to MQTT topics for bi-directional communication.
func (mb *MQTTBridge) Subscribe(topic string, handler func(topic string, payload []byte)) error {
	if !mb.enabled {
		return fmt.Errorf("MQTT bridge is disabled")
	}

	// In a real implementation, this would subscribe to topics
	// and invoke the handler when messages arrive

	log.Printf("[MQTT] Would subscribe to %s", topic)
	return nil
}

// MQTTCommandHandler handles MQTT commands (for bi-directional communication).
type MQTTCommandHandler struct {
	bridge *MQTTBridge
}

// NewMQTTCommandHandler creates a new MQTT command handler.
func NewMQTTCommandHandler(bridge *MQTTBridge) *MQTTCommandHandler {
	return &MQTTCommandHandler{
		bridge: bridge,
	}
}

// HandleWriteCommand handles a write command from MQTT.
// Topic: modbridge/{proxy_id}/devices/{device_id}/write
func (mch *MQTTCommandHandler) HandleWriteCommand(topic string, payload []byte) error {
	var cmd struct {
		Address  uint16      `json:"address"`
		Value    interface{} `json:"value"`
		DataType string      `json:"data_type"`
	}

	if err := json.Unmarshal(payload, &cmd); err != nil {
		return fmt.Errorf("failed to unmarshal command: %w", err)
	}

	log.Printf("[MQTT Command] Write to address %d: %v (type: %s)",
		cmd.Address, cmd.Value, cmd.DataType)

	// TODO: Execute Modbus write operation

	return nil
}

// HandleProxyCommand handles proxy control commands.
// Topic: modbridge/{proxy_id}/control
func (mch *MQTTCommandHandler) HandleProxyCommand(topic string, payload []byte) error {
	var cmd struct {
		Action string `json:"action"` // start, stop, restart
	}

	if err := json.Unmarshal(payload, &cmd); err != nil {
		return fmt.Errorf("failed to unmarshal command: %w", err)
	}

	log.Printf("[MQTT Command] Proxy control: %s", cmd.Action)

	// TODO: Execute proxy control action

	return nil
}
