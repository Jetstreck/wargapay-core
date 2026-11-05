package service

import (
	"crypto/sha512"
	"encoding/hex"
	"os"
	"testing"
)

func TestMidtransSignatureVerification(t *testing.T) {
	serverKey := "SB-Mid-server-test-key"
	os.Setenv("MIDTRANS_SERVER_KEY", serverKey)
	defer os.Unsetenv("MIDTRANS_SERVER_KEY")

	svc := NewMidtransService()

	orderID := "WP-DUES-1-12345678"
	statusCode := "200"
	grossAmount := "50000.00"

	// Calculate correct signature
	raw := orderID + statusCode + grossAmount + serverKey
	hash := sha512.Sum512([]byte(raw))
	expectedSignature := hex.EncodeToString(hash[:])

	// Verify with valid signature
	if !svc.VerifySignature(orderID, statusCode, grossAmount, expectedSignature) {
		t.Errorf("Expected signature to be valid, but got false")
	}

	// Verify with invalid signature
	if svc.VerifySignature(orderID, statusCode, grossAmount, "invalid_signature") {
		t.Errorf("Expected signature to be invalid, but got true")
	}
}

