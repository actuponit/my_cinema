package services

import (
	"cinema-server/utils"
	"fmt"
	"log"
	"time"
)

// The machine firmware subscribes to this topic; don't change it or the payload format.
const dispenseTopic = "chapa/webhook"

type VendingMachineServiceInterface interface {
	SendCommand(motorCodes string) error
	IsConnected() bool
}

type VendingMachineService struct {
	mqttClient utils.MQTTClientInterface
	retryDelay time.Duration
}

func NewVendingMachineService(mqttClient utils.MQTTClientInterface) *VendingMachineService {
	return &VendingMachineService{mqttClient: mqttClient, retryDelay: 200 * time.Millisecond}
}

func (s *VendingMachineService) IsConnected() bool {
	return s.mqttClient.IsConnected()
}

// SendCommand publishes the comma-separated motor codes, retrying 3 times (200, 400, 800 ms).
func (s *VendingMachineService) SendCommand(motorCodes string) error {
	err := s.mqttClient.Publish(dispenseTopic, motorCodes)
	for i, delay := 0, s.retryDelay; err != nil && i < 3; i, delay = i+1, delay*2 {
		log.Printf("MQTT publish failed (attempt %d): %v", i+1, err)
		time.Sleep(delay)
		err = s.mqttClient.Publish(dispenseTopic, motorCodes)
	}
	if err != nil {
		return fmt.Errorf("publish motor codes %q: %w", motorCodes, err)
	}
	return nil
}
