package services

import (
	"cinema-server/utils"
	"log"
)

// VendingMachineServiceInterface defines the interface for vending machine operations
type VendingMachineServiceInterface interface {
	StartMotor(vendingMachineID int) error
	StopMotor(vendingMachineID int) error
	SendCommand(vendingMachineID int, command string) error
}

// VendingMachineService implements vending machine operations
type VendingMachineService struct {
	mqttClient utils.MQTTClientInterface
}

// NewVendingMachineService creates a new vending machine service instance
func NewVendingMachineService(mqttClient utils.MQTTClientInterface) *VendingMachineService {
	return &VendingMachineService{
		mqttClient: mqttClient,
	}
}

// StartMotor sends a command to start the motor of a specific vending machine
func (s *VendingMachineService) StartMotor(vendingMachineID int) error {
	return s.SendCommand(vendingMachineID, "start_motor")
}

// StopMotor sends a command to stop the motor of a specific vending machine
func (s *VendingMachineService) StopMotor(vendingMachineID int) error {
	return s.SendCommand(vendingMachineID, "stop_motor")
}

// SendCommand sends a custom command to a specific vending machine
func (s *VendingMachineService) SendCommand(vendingMachineID int, command string) error {
	// Connect to MQTT broker
	if err := s.mqttClient.Connect(); err != nil {
		log.Printf("Failed to connect to MQTT broker: %v", err)
		return err
	}
	defer s.mqttClient.Disconnect()

	// Define the topic with vending machine ID
	topic := "chapa/webhook"

	// Create a structured message (you can modify this based on your needs)
	message := command

	// Publish the message
	if err := s.mqttClient.Publish(topic, message); err != nil {
		log.Printf("Failed to publish MQTT message: %v", err)
		return err
	}

	log.Printf("Successfully sent command '%s' to vending machine %d", command, vendingMachineID)
	return nil
}
