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
func VerifyChapaWebhookSignature(requestBody []byte, signature string) error {
	// Get the secret key from environment variable
	secret := os.Getenv("CHAPA_WEBHOOK_SECRET")
	if secret == "" {
		return fmt.Errorf("CHAPA_WEBHOOK_SECRET environment variable not set")
	}

	// Create HMAC-SHA256 hash
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(requestBody)
	expectedHash := hex.EncodeToString(h.Sum(nil))

	// Compare the expected hash with the received signature
	if expectedHash != signature {
		return fmt.Errorf("invalid webhook signature: expected %s, got %s", expectedHash, signature)
	}

	return nil
}

// VerifyChapaWebhookSignatureFromMap verifies the Chapa webhook signature from a map
func VerifyChapaWebhookSignatureFromMap(requestBody map[string]interface{}, signature string) error {
	// Convert map to JSON bytes
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	return VerifyChapaWebhookSignature(jsonBytes, signature)
}
