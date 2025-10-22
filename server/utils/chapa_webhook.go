package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

// VerifyChapaWebhookSignature verifies the Chapa webhook signature
func VerifyChapaWebhookSignature(requestBody []byte, signature, signature2 string) error {
	fmt.Println("VerifyChapaWebhookSignature:", string(requestBody), signature)
	fmt.Println("VerifyChapaWebhookSignature2:", string(requestBody), signature2)
	// Get the secret key from environment variable
	secret := os.Getenv("CHAPA_WEBHOOK_SECRET")
	if secret == "" {
		return fmt.Errorf("CHAPA_WEBHOOK_SECRET environment variable not set")
	}

	// Create HMAC-SHA256 hash
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(secret))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	// Compare the expected hash with the received signature
	if expectedHash != signature && expectedHash != signature2 {
		return fmt.Errorf("invalid webhook signature: expected %s, got %s", expectedHash, signature)
	}

	return nil
}

// VerifyChapaWebhookSignatureFromMap verifies the Chapa webhook signature from a map
func VerifyChapaWebhookSignatureFromMap(requestBody map[string]interface{}, signature, signature2 string) error {
	// Convert map to JSON bytes
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	return VerifyChapaWebhookSignature(jsonBytes, signature, signature2)
}
