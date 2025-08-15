package utils

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"crypto/hmac"
	"crypto/sha256"
)

func TestVerifyChapaWebhookSignature(t *testing.T) {
	// Set test secret
	os.Setenv("CHAPA_WEBHOOK_SECRET", "test_secret_key")
	defer os.Unsetenv("CHAPA_WEBHOOK_SECRET")

	// Test data
	requestBody := []byte(`{"amount":"100","currency":"ETB","tx_ref":"test_ref"}`)

	// Create expected signature
	h := hmac.New(sha256.New, []byte("test_secret_key"))
	h.Write(requestBody)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Test valid signature
	err := VerifyChapaWebhookSignature(requestBody, expectedSignature)
	if err != nil {
		t.Errorf("Expected no error for valid signature, got: %v", err)
	}

	// Test invalid signature
	err = VerifyChapaWebhookSignature(requestBody, "invalid_signature")
	if err == nil {
		t.Error("Expected error for invalid signature, got nil")
	}

	// Test missing secret
	os.Unsetenv("CHAPA_WEBHOOK_SECRET")
	err = VerifyChapaWebhookSignature(requestBody, expectedSignature)
	if err == nil {
		t.Error("Expected error for missing secret, got nil")
	}
}

func TestVerifyChapaWebhookSignatureFromMap(t *testing.T) {
	// Set test secret
	os.Setenv("CHAPA_WEBHOOK_SECRET", "test_secret_key")
	defer os.Unsetenv("CHAPA_WEBHOOK_SECRET")

	// Test data
	requestBody := map[string]interface{}{
		"amount":   "100",
		"currency": "ETB",
		"tx_ref":   "test_ref",
	}

	// Create expected signature
	jsonBytes, _ := json.Marshal(requestBody)
	h := hmac.New(sha256.New, []byte("test_secret_key"))
	h.Write(jsonBytes)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Test valid signature
	err := VerifyChapaWebhookSignatureFromMap(requestBody, expectedSignature)
	if err != nil {
		t.Errorf("Expected no error for valid signature, got: %v", err)
	}

	// Test invalid signature
	err = VerifyChapaWebhookSignatureFromMap(requestBody, "invalid_signature")
	if err == nil {
		t.Error("Expected error for invalid signature, got nil")
	}
}
