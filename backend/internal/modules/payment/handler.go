package payment

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spos/backend/internal/config"
	"github.com/spos/backend/internal/models"
	"github.com/spos/backend/internal/utils"
	"gorm.io/gorm"
)

type MidtransNotificationPayload struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	TransactionStatus string `json:"transaction_status"`
	SignatureKey      string `json:"signature_key"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
}

type PaymentHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewPaymentHandler(db *gorm.DB, cfg *config.Config) *PaymentHandler {
	return &PaymentHandler{
		db:  db,
		cfg: cfg,
	}
}

type CashPaymentRequest struct {
	OrderID    string  `json:"order_id"`
	AmountPaid float64 `json:"amount_paid"`
}

// CashCheckout memproses pelunasan transaksi tunai secara transaksional, memperbarui status pesanan, dan mengosongkan/mengisi stok bahan baku via BOM.
func (h *PaymentHandler) CashCheckout(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	var req CashPaymentRequest
	if err := c.BodyParser(&req); err != nil || req.OrderID == "" || req.AmountPaid <= 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "order_id dan amount_paid (lebih besar dari 0) wajib diisi", nil)
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var order models.Order
	if err := tx.Preload("OrderItems").Where("id = ? AND outlet_id = ?", req.OrderID, outletID).First(&order).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusNotFound, "Pesanan tidak ditemukan pada outlet ini", nil)
	}

	if order.Status == "paid" {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusBadRequest, "Pesanan ini sudah dilunasi sebelumnya", nil)
	}

	if req.AmountPaid < order.TotalAmount {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusBadRequest, fmt.Sprintf("Pembayaran tidak mencukupi. Diharuskan: %.2f, Dibayar: %.2f", order.TotalAmount, req.AmountPaid), nil)
	}

	changeDue := req.AmountPaid - order.TotalAmount
	now := time.Now()

	// 1. Buat catatan Pembayaran (Payment)
	payment := models.Payment{
		OrderID:       order.ID,
		PaymentMethod: "cash",
		AmountPaid:    req.AmountPaid,
		ChangeDue:     changeDue,
		Status:        "settlement",
		SettledAt:     &now,
	}

	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mencatat pembayaran tunai", err.Error())
	}

	// 2. Perbarui status pesanan menjadi paid
	order.Status = "paid"
	order.UpdatedAt = now
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui status pesanan", err.Error())
	}

	// 3. Perbarui status meja menjadi terisi (occupied) jika pesanan terhubung dengan meja
	if order.TableID != nil && *order.TableID > 0 {
		if err := tx.Model(&models.Table{}).Where("id = ?", *order.TableID).Update("status", "occupied").Error; err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui status meja", err.Error())
		}
	}

	// 4. Pemotongan Stok Bahan Baku BOM secara Otomatis & Transaksional
	if err := executeBOMStockDeduction(tx, order.ID); err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memotong stok bahan baku inventaris", err.Error())
	}

	// Komit Transaksi Database
	if err := tx.Commit().Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengomit transaksi pembayaran", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Pembayaran tunai berhasil diselesaikan", fiber.Map{
		"payment_id":  payment.ID,
		"order_id":    order.ID,
		"total":       order.TotalAmount,
		"amount_paid": req.AmountPaid,
		"change_due":  changeDue,
		"status":      "paid",
	})
}

type MidtransChargeRequest struct {
	OrderID     string `json:"order_id"`
	PaymentType string `json:"payment_type"` // qris, bank_transfer, gopay
}

type midtransChargePayload struct {
	PaymentType        string                    `json:"payment_type"`
	TransactionDetails midtransTransactionDetail `json:"transaction_details"`
	Qris               *midtransQrisDetail       `json:"qris,omitempty"`
}

type midtransTransactionDetail struct {
	OrderID     string  `json:"order_id"`
	GrossAmount float64 `json:"gross_amount"`
}

type midtransQrisDetail struct {
	Acquirer string `json:"acquirer,omitempty"`
}

type midtransResponseAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

type midtransChargeAPIResponse struct {
	StatusCode        string                   `json:"status_code"`
	StatusMessage     string                   `json:"status_message"`
	TransactionID     string                   `json:"transaction_id"`
	OrderID           string                   `json:"order_id"`
	GrossAmount       string                   `json:"gross_amount"`
	PaymentType       string                   `json:"payment_type"`
	TransactionStatus string                   `json:"transaction_status"`
	Actions           []midtransResponseAction `json:"actions"`
	QRString          string                   `json:"qr_string"`
}

type midtransStatusAPIResponse struct {
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	PaymentType       string `json:"payment_type"`
	TransactionStatus string `json:"transaction_status"`
	SignatureKey      string `json:"signature_key"`
}

func getMidtransServerKey(db *gorm.DB, cfg *config.Config) string {
	if cfg != nil && cfg.MidtransServerKey != "" && !strings.Contains(cfg.MidtransServerKey, "placeholder") {
		return cfg.MidtransServerKey
	}
	if envKey := os.Getenv("MIDTRANS_SERVER_KEY"); envKey != "" && !strings.Contains(envKey, "placeholder") {
		return envKey
	}
	var sysConf models.SystemConfig
	if db != nil {
		if err := db.Where("key = ?", "midtrans_server_key").First(&sysConf).Error; err == nil && sysConf.Value != "" && !strings.Contains(sysConf.Value, "placeholder") {
			return sysConf.Value
		}
	}
	return ""
}

// MidtransCharge menghasilkan payload penagihan dinamis QRIS / Midtrans Core API.
func (h *PaymentHandler) MidtransCharge(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)

	var req MidtransChargeRequest
	if err := c.BodyParser(&req); err != nil || req.OrderID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "order_id wajib diisi", nil)
	}

	var order models.Order
	if err := h.db.Where("id = ? AND outlet_id = ?", req.OrderID, outletID).First(&order).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Pesanan tidak ditemukan pada outlet ini", nil)
	}

	if order.Status == "paid" {
		return utils.SendError(c, fiber.StatusBadRequest, "Pesanan ini sudah dilunasi sebelumnya", nil)
	}

	if req.PaymentType == "" {
		req.PaymentType = "qris_dynamic"
	}

	serverKey := getMidtransServerKey(h.db, h.cfg)
	isProd := h.cfg != nil && h.cfg.MidtransIsProduction

	baseURL := "https://api.sandbox.midtrans.com"
	if isProd {
		baseURL = "https://api.midtrans.com"
	}

	cleanInvoice := strings.ReplaceAll(order.InvoiceNumber, "/", "-")
	midtransOrderID := fmt.Sprintf("%s-%d", cleanInvoice, time.Now().Unix())

	var qrisURL string
	var qrString string
	var midtransTxID string = midtransOrderID

	// 1. Jika Midtrans Server Key dikonfigurasi, eksekusi pemanggilan HTTP resmi ke Midtrans Core API (/v2/charge)
	if serverKey != "" {
		payload := midtransChargePayload{
			PaymentType: "qris",
			TransactionDetails: midtransTransactionDetail{
				OrderID:     midtransOrderID,
				GrossAmount: order.TotalAmount,
			},
			Qris: &midtransQrisDetail{
				Acquirer: "gopay",
			},
		}

		payloadBytes, _ := json.Marshal(payload)
		reqURL := fmt.Sprintf("%s/v2/charge", baseURL)
		httpReq, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(payloadBytes))
		if err == nil {
			httpReq.Header.Set("Accept", "application/json")
			httpReq.Header.Set("Content-Type", "application/json")
			authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(serverKey+":"))
			httpReq.Header.Set("Authorization", authHeader)

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(httpReq)
			if err == nil {
				defer resp.Body.Close()
				bodyBytes, _ := io.ReadAll(resp.Body)
				var apiRes midtransChargeAPIResponse
				if err := json.Unmarshal(bodyBytes, &apiRes); err == nil && (apiRes.StatusCode == "200" || apiRes.StatusCode == "201") {
					midtransTxID = apiRes.TransactionID
					qrString = apiRes.QRString
					for _, act := range apiRes.Actions {
						if act.Name == "generate-qr-code" {
							qrisURL = act.URL
							break
						}
					}
				}
			}
		}
	}

	// 2. Fallback jika offline / mode simulasi lokal tanpa API Key Midtrans
	if qrisURL == "" {
		if qrString == "" {
			qrString = fmt.Sprintf("00020101021226680014ID.LINKAJA.WWW0118936009110021000001520458125303360540%.2f5802ID5911SPOS RESTO6007JAKARTA61051211062070703A016304", order.TotalAmount)
		}
		qrisURL = fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s", url.QueryEscape(qrString))
	}

	// 3. Catat entri pembayaran pending ke database
	now := time.Now()
	payment := models.Payment{
		OrderID:               order.ID,
		PaymentMethod:         "midtrans_dynamic",
		AmountPaid:            order.TotalAmount,
		ChangeDue:             0,
		MidtransTransactionID: midtransTxID,
		MidtransPaymentType:   req.PaymentType,
		Status:                "pending",
		CreatedAt:             now,
	}

	if err := h.db.Create(&payment).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menginisialisasi catatan pembayaran Midtrans", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Penagihan dinamis Midtrans berhasil dibuat", fiber.Map{
		"payment_id":              payment.ID,
		"order_id":                order.ID,
		"midtrans_transaction_id": midtransTxID,
		"gross_amount":            order.TotalAmount,
		"payment_type":            req.PaymentType,
		"qris_url":                qrisURL,
		"qr_string":               qrString,
		"status":                  "pending",
		"is_configured":           serverKey != "",
	})
}

// MidtransWebhook memvalidasi tanda tangan SHA-512 dan memproses pelunasan transaksi & pemotongan stok BOM secara transaksional.
func (h *PaymentHandler) MidtransWebhook(c *fiber.Ctx) error {
	var notif utils.MidtransNotification
	if err := c.BodyParser(&notif); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Payload notifikasi webhook tidak valid", err.Error())
	}

	if notif.OrderID == "" || notif.StatusCode == "" || notif.GrossAmount == "" || notif.SignatureKey == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Komponen penandatanganan webhook wajib tidak lengkap", nil)
	}

	// 1. Validasi Wajib Tanda Tangan SHA-512 Notifikasi Webhook
	serverKey := getMidtransServerKey(h.db, h.cfg)
	isValid := utils.VerifyMidtransSignature(notif.OrderID, notif.StatusCode, notif.GrossAmount, serverKey, notif.SignatureKey)
	if !isValid {
		return utils.SendError(c, fiber.StatusUnauthorized, "Validasi tanda tangan Midtrans gagal", nil)
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var order models.Order
	if err := tx.Preload("OrderItems").First(&order, "id = ?", notif.OrderID).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusNotFound, "Pesanan yang dirujuk dalam webhook tidak ditemukan", nil)
	}

	now := time.Now()

	if notif.TransactionStatus == "settlement" || notif.TransactionStatus == "capture" {
		if order.Status != "paid" {
			order.Status = "paid"
			order.UpdatedAt = now
			if err := tx.Save(&order).Error; err != nil {
				tx.Rollback()
				return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui status pesanan", err.Error())
			}

			if order.TableID != nil && *order.TableID > 0 {
				tx.Model(&models.Table{}).Where("id = ?", *order.TableID).Update("status", "occupied")
			}

			if err := executeBOMStockDeduction(tx, order.ID); err != nil {
				tx.Rollback()
				return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memotong stok bahan baku saat pelunasan webhook", err.Error())
			}

			var payment models.Payment
			if err := tx.Where("order_id = ?", order.ID).First(&payment).Error; err != nil {
				payment = models.Payment{
					OrderID:               order.ID,
					PaymentMethod:         "midtrans_dynamic",
					AmountPaid:            order.TotalAmount,
					MidtransTransactionID: notif.TransactionID,
					MidtransPaymentType:   notif.PaymentType,
					Status:                "settlement",
					SettledAt:             &now,
					CreatedAt:             now,
				}
				tx.Create(&payment)
			} else {
				payment.Status = "settlement"
				payment.SettledAt = &now
				payment.MidtransTransactionID = notif.TransactionID
				payment.MidtransPaymentType = notif.PaymentType
				tx.Save(&payment)
			}
		}
	} else if notif.TransactionStatus == "expire" || notif.TransactionStatus == "cancel" || notif.TransactionStatus == "deny" {
		tx.Model(&models.Payment{}).Where("order_id = ?", order.ID).Update("status", notif.TransactionStatus)
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengomit transaksi webhook Midtrans", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Webhook Midtrans berhasil diproses", fiber.Map{
		"order_id":           order.ID,
		"transaction_status": notif.TransactionStatus,
	})
}

// GetPaymentStatus memeriksa status pembayaran terbaru untuk pesanan tertentu.
func (h *PaymentHandler) GetPaymentStatus(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	outletID, _ := c.Locals("outlet_id").(string)

	if orderID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "order_id wajib diisi", nil)
	}

	var payment models.Payment
	err := h.db.Joins("JOIN orders ON orders.id = payments.order_id").
		Where("payments.order_id = ? AND orders.outlet_id = ?", orderID, outletID).
		Order("payments.id desc").
		First(&payment).Error

	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Status pembayaran tidak ditemukan untuk pesanan ini", nil)
	}

	// Cek status ke API Server Midtrans secara langsung jika transaksi masih pending dan API Key tersedia
	serverKey := getMidtransServerKey(h.db, h.cfg)
	if payment.Status == "pending" && serverKey != "" && payment.MidtransTransactionID != "" {
		isProd := h.cfg != nil && h.cfg.MidtransIsProduction
		baseURL := "https://api.sandbox.midtrans.com"
		if isProd {
			baseURL = "https://api.midtrans.com"
		}

		statusReqURL := fmt.Sprintf("%s/v2/%s/status", baseURL, payment.MidtransTransactionID)
		req, err := http.NewRequest("GET", statusReqURL, nil)
		if err == nil {
			authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(serverKey+":"))
			req.Header.Set("Authorization", authHeader)
			req.Header.Set("Accept", "application/json")

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				bodyBytes, _ := io.ReadAll(resp.Body)
				var stRes midtransStatusAPIResponse
				if err := json.Unmarshal(bodyBytes, &stRes); err == nil && (stRes.TransactionStatus == "settlement" || stRes.TransactionStatus == "capture") {
					now := time.Now()
					tx := h.db.Begin()
					var order models.Order
					if tx.First(&order, "id = ?", payment.OrderID).Error == nil && order.Status != "paid" {
						order.Status = "paid"
						order.UpdatedAt = now
						tx.Save(&order)
						if order.TableID != nil && *order.TableID > 0 {
							tx.Model(&models.Table{}).Where("id = ?", *order.TableID).Update("status", "occupied")
						}
						_ = executeBOMStockDeduction(tx, order.ID)
					}
					payment.Status = "settlement"
					payment.SettledAt = &now
					tx.Save(&payment)
					tx.Commit()
				}
			}
		}
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Status pembayaran berhasil diambil", payment)
}

// executeBOMStockDeduction menjalankan pengurangan stok bahan baku (BOM) secara otomatis pada inventory_logs dan ingredients.
func executeBOMStockDeduction(tx *gorm.DB, orderID string) error {
	var items []models.OrderItem
	if err := tx.Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return err
	}

	ingredientDeductions := make(map[string]float64)
	ingredientItemNotes := make(map[string][]string)

	now := time.Now()

	for _, item := range items {
		var product models.Product
		productName := "Product"
		if err := tx.First(&product, "id = ?", item.ProductID).Error; err == nil {
			productName = product.Name
		}

		var recipes []models.ProductRecipe
		if err := tx.Where("product_id = ?", item.ProductID).Find(&recipes).Error; err != nil {
			return err
		}

		for _, recipe := range recipes {
			deductQty := float64(item.Quantity) * recipe.QuantityRequired
			ingredientDeductions[recipe.IngredientID] += deductQty
			note := fmt.Sprintf("Menu '%s' (%d pcs x %.2f)", productName, item.Quantity, recipe.QuantityRequired)
			ingredientItemNotes[recipe.IngredientID] = append(ingredientItemNotes[recipe.IngredientID], note)
		}
	}

	for ingredientID, totalDeductQty := range ingredientDeductions {
		var ingredient models.Ingredient
		// Penguncian baris (row locking FOR UPDATE) untuk mencegah race condition pada checkout bersamaan
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&ingredient, "id = ?", ingredientID).Error; err != nil {
			return fmt.Errorf("bahan baku %s tidak ditemukan: %w", ingredientID, err)
		}

		remainingStock := ingredient.CurrentStock - totalDeductQty
		ingredient.CurrentStock = remainingStock
		ingredient.UpdatedAt = now

		if err := tx.Save(&ingredient).Error; err != nil {
			return fmt.Errorf("gagal memperbarui stok bahan baku %s: %w", ingredient.Name, err)
		}

		noteStr := strings.Join(ingredientItemNotes[ingredientID], "; ")
		if len(noteStr) > 250 {
			noteStr = noteStr[:247] + "..."
		}

		logEntry := models.InventoryLog{
			IngredientID:   ingredient.ID,
			ChangeType:     "sale_deduction",
			QuantityChange: -totalDeductQty,
			RemainingStock: remainingStock,
			Notes:          "Pengurangan penjualan POS: " + noteStr,
			CreatedAt:      now,
		}

		if err := tx.Create(&logEntry).Error; err != nil {
			return fmt.Errorf("gagal mencatat log inventaris untuk %s: %w", ingredient.Name, err)
		}
	}

	return nil
}
