package utils

import (
	"cinema-server/config"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTClientInterface defines the interface for MQTT operations
type MQTTClientInterface interface {
	Connect() error
	Disconnect()
	Publish(topic, message string) error
	Subscribe(topic string, handler mqtt.MessageHandler) error
	IsConnected() bool
}

// MQTTClient implements MQTT operations
type MQTTClient struct {
	client mqtt.Client
	opts   *mqtt.ClientOptions
}

// NewMQTTClient creates a new MQTT client instance
func NewMQTTClient(config config.MQTTConfig) *MQTTClient {
	opts := mqtt.NewClientOptions().AddBroker(config.BrokerURL)

	if config.ClientID != "" {
		opts.SetClientID(config.ClientID)
	}

	if config.Username != "" {
		opts.SetUsername(config.Username)
		opts.SetPassword(config.Password)
	}

	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("Connected to MQTT broker")
	})

	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(1 * time.Minute)
	opts.SetConnectTimeout(config.Timeout)

	return &MQTTClient{
		opts: opts,
	}
}

// NewMQTTClientFromConfig creates a new MQTT client instance from config.MQTTConfig
func NewMQTTClientFromConfig(config config.MQTTConfig) *MQTTClient {
	opts := mqtt.NewClientOptions().AddBroker(config.BrokerURL)

	if config.ClientID != "" {
		opts.SetClientID(config.ClientID)
	}

	if config.Username != "" {
		opts.SetUsername(config.Username)
		opts.SetPassword(config.Password)
	}

	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("Connected to MQTT broker")
	})

	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(1 * time.Minute)
	opts.SetConnectTimeout(config.Timeout)

	return &MQTTClient{
		opts: opts,
	}
}

// Connect establishes connection to MQTT broker
func (m *MQTTClient) Connect() error {
	m.client = mqtt.NewClient(m.opts)
	token := m.client.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Disconnect closes the MQTT connection
func (m *MQTTClient) Disconnect() {
	if m.client != nil && m.client.IsConnected() {
		m.client.Disconnect(250)
	}
}

// Publish sends a message to a specific topic
func (m *MQTTClient) Publish(topic, message string) error {
	if !m.IsConnected() {
		return mqtt.ErrNotConnected
	}

	token := m.client.Publish(topic, 0, false, message)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	log.Printf("Published message to topic %s: %s", topic, message)
	return nil
}

// Subscribe subscribes to a topic with a message handler
func (m *MQTTClient) Subscribe(topic string, handler mqtt.MessageHandler) error {
	if !m.IsConnected() {
		return mqtt.ErrNotConnected
	}

	token := m.client.Subscribe(topic, 0, handler)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	log.Printf("Subscribed to topic: %s", topic)
	return nil
}

// IsConnected checks if the client is connected to the broker
func (m *MQTTClient) IsConnected() bool {
	return m.client != nil && m.client.IsConnected()
}

// DefaultMQTTConfig returns a default configuration for MQTT
func DefaultMQTTConfig() config.MQTTConfig {
	return config.MQTTConfig{
		BrokerURL: "tcp://test.mosquitto.org:1883",
		ClientID:  "",
		Timeout:   100 * time.Second,
	}
}
