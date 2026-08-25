package utils

import (
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestVerifyMidtransSignature(t *testing.T) {
	orderID := "INV-20260825-001"
	statusCode := "200"
	grossAmount := "50000.00"
	serverKey := "SB-Mid-server-secret"

	raw := orderID + statusCode + grossAmount + serverKey
	hasher := sha512.New()
	hasher.Write([]byte(raw))
	expectedSignature := hex.EncodeToString(hasher.Sum(nil))

	// Test valid signature
	isValid := VerifyMidtransSignature(orderID, statusCode, grossAmount, serverKey, expectedSignature)
	if !isValid {
		t.Error("Expected VerifyMidtransSignature to return true for valid signature")
	}

	// Test invalid signature
	isInvalid := VerifyMidtransSignature(orderID, statusCode, grossAmount, serverKey, "invalid_signature_hash")
	if isInvalid {
		t.Error("Expected VerifyMidtransSignature to return false for invalid signature")
	}
}

func TestSendSuccessAndSendErrorHelpers(t *testing.T) {
	app := fiber.New()

	app.Get("/success", func(c *fiber.Ctx) error {
		return SendSuccess(c, fiber.StatusOK, "Berhasil Ditampilkan", fiber.Map{"key": "value"})
	})

	app.Get("/error", func(c *fiber.Ctx) error {
		return SendError(c, fiber.StatusBadRequest, "Payload Tidak Valid", "detail error")
	})

	// Test success response format
	reqSuccess := httptest.NewRequest("GET", "/success", nil)
	respSuccess, err := app.Test(reqSuccess)
	if err != nil || respSuccess.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for SendSuccess helper, got %d", respSuccess.StatusCode)
	}

	// Test error response format
	reqError := httptest.NewRequest("GET", "/error", nil)
	respError, err := app.Test(reqError)
	if err != nil || respError.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for SendError helper, got %d", respError.StatusCode)
	}
}
