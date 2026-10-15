package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

// VerifyChapaWebhookSignature verifies the Chapa webhook signature against raw request body bytes.
func VerifyChapaWebhookSignature(requestBody []byte, signature string) error {
	secret := os.Getenv("CHAPA_WEBHOOK_SECRET")
	if secret == "" {
		return fmt.Errorf("CHAPA_WEBHOOK_SECRET environment variable not set")
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(requestBody)
	expectedHash := h.Sum(nil)

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}

	if !hmac.Equal(expectedHash, sigBytes) {
		return fmt.Errorf("invalid webhook signature")
	}

	return nil
}

// VerifyChapaWebhookSignatureFromMap verifies the Chapa webhook signature from a map
func VerifyChapaWebhookSignatureFromMap(requestBody map[string]interface{}, signature string) error {
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	return VerifyChapaWebhookSignature(jsonBytes, signature)
}
