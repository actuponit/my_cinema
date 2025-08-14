package config

import (
	"os"
	"strconv"
	"time"
)

// MQTTConfig holds configuration for MQTT client
type MQTTConfig struct {
	BrokerURL string
	ClientID  string
	Username  string
	Password  string
	Timeout   time.Duration
}

// NewMQTTConfig creates a new MQTT configuration from environment variables
func NewMQTTConfig() *MQTTConfig {
	timeoutStr := os.Getenv("MQTT_TIMEOUT_SECONDS")
	timeout := 30 * time.Second // default timeout

	if timeoutStr != "" {
		if seconds, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(seconds) * time.Second
		}
	}

	return &MQTTConfig{
		BrokerURL: getEnvOrDefault("MQTT_BROKER_URL", "tcp://test.mosquitto.org:1883"),
		ClientID:  getEnvOrDefault("MQTT_CLIENT_ID", "cinema-server"),
		Username:  os.Getenv("MQTT_USERNAME"),
		Password:  os.Getenv("MQTT_PASSWORD"),
		Timeout:   timeout,
	}
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
