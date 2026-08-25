package utils

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"
)

// VerifyMidtransSignature memvalidasi kunci tanda tangan (signature key) notifikasi webhook Midtrans menggunakan algoritma SHA-512.
// Algoritma: SHA512(order_id + status_code + gross_amount + server_key)
func VerifyMidtransSignature(orderID, statusCode, grossAmount, serverKey, signature string) bool {
	raw := fmt.Sprintf("%s%s%s%s", orderID, statusCode, grossAmount, serverKey)
	hash := sha512.Sum512([]byte(raw))
	calculatedSignature := hex.EncodeToString(hash[:])
	return strings.EqualFold(calculatedSignature, signature)
}

// MidtransNotification merepresentasikan struktur payload postback webhook standar dari Midtrans.
type MidtransNotification struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status,omitempty"`
	Currency          string `json:"currency,omitempty"`
}

// MidtransChargeRequest merepresentasikan payload pembuatan transaksi penagihan dinamis Midtrans Core API.
type MidtransChargeRequest struct {
	PaymentType        string                   `json:"payment_type"`
	TransactionDetails MidtransTransactionDetail `json:"transaction_details"`
	CustomerDetails    *MidtransCustomerDetail  `json:"customer_details,omitempty"`
}

type MidtransTransactionDetail struct {
	OrderID     string  `json:"order_id"`
	GrossAmount int64   `json:"gross_amount"`
}

type MidtransCustomerDetail struct {
	FirstName string `json:"first_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}
