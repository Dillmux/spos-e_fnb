package shift

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spos/backend/internal/models"
	"github.com/spos/backend/internal/utils"
	"gorm.io/gorm"
)

type ShiftHandler struct {
	db *gorm.DB
}

func NewShiftHandler(db *gorm.DB) *ShiftHandler {
	return &ShiftHandler{db: db}
}

type OpenShiftRequest struct {
	OpeningCash float64 `json:"opening_cash"`
	Notes       string  `json:"notes"`
}

// OpenShift membuka sesi shift laci kasir baru.
func (h *ShiftHandler) OpenShift(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	var req OpenShiftRequest
	if err := c.BodyParser(&req); err != nil || req.OpeningCash < 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "opening_cash harus >= 0", nil)
	}

	var activeShift models.Shift
	if err := h.db.Where("user_id = ? AND status = ?", userID, "open").First(&activeShift).Error; err == nil {
		return utils.SendSuccess(c, fiber.StatusOK, "Pengguna sudah memiliki shift yang sedang aktif", activeShift)
	}

	now := time.Now()
	shift := models.Shift{
		UserID:      userID,
		StartTime:   now,
		OpeningCash: req.OpeningCash,
		Status:      "open",
		Notes:       req.Notes,
	}

	if err := h.db.Create(&shift).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membuka shift", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusCreated, "Shift berhasil dibuka", shift)
}

// GetCurrentShift mengambil rincian detail shift kasir yang sedang aktif.
func (h *ShiftHandler) GetCurrentShift(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	var shift models.Shift
	if err := h.db.Where("user_id = ? AND status = ?", userID, "open").First(&shift).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Tidak ada shift aktif yang ditemukan untuk pengguna ini", nil)
	}

	var cashPaymentsTotal float64
	h.db.Model(&models.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.shift_id = ? AND payments.status = 'settlement' AND payments.payment_method = 'cash'", shift.ID).
		Select("COALESCE(SUM(payments.amount_paid - payments.change_due), 0)").
		Scan(&cashPaymentsTotal)

	var cashInTotal, cashOutTotal float64
	h.db.Model(&models.CashMovement{}).Where("shift_id = ? AND type = 'cash_in'", shift.ID).Select("COALESCE(SUM(amount), 0)").Scan(&cashInTotal)
	h.db.Model(&models.CashMovement{}).Where("shift_id = ? AND type = 'cash_out'", shift.ID).Select("COALESCE(SUM(amount), 0)").Scan(&cashOutTotal)

	expectedCash := shift.OpeningCash + cashPaymentsTotal + cashInTotal - cashOutTotal

	var movements []models.CashMovement
	h.db.Where("shift_id = ?", shift.ID).Find(&movements)

	return utils.SendSuccess(c, fiber.StatusOK, "Shift aktif berhasil diambil", fiber.Map{
		"shift":             shift,
		"cash_sales_total":  cashPaymentsTotal,
		"cash_in_total":     cashInTotal,
		"cash_out_total":    cashOutTotal,
		"expected_cash_now": expectedCash,
		"movements":         movements,
	})
}

type AddCashMovementRequest struct {
	Type   string  `json:"type"`   // in, out, cash_in, cash_out
	Amount float64 `json:"amount"`
	Notes  string  `json:"notes"`
	Reason string  `json:"reason"`
}

// AddCashMovement mencatat transaksi kas masuk atau kas keluar (petty cash) selama shift aktif.
func (h *ShiftHandler) AddCashMovement(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	var req AddCashMovementRequest
	if err := c.BodyParser(&req); err != nil || req.Amount <= 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "type, amount > 0, dan notes/reason wajib diisi", nil)
	}

	mType := req.Type
	if mType == "in" {
		mType = "cash_in"
	} else if mType == "out" {
		mType = "cash_out"
	}

	if mType != "cash_in" && mType != "cash_out" {
		return utils.SendError(c, fiber.StatusBadRequest, "type harus berupa cash_in atau cash_out", nil)
	}

	notes := req.Notes
	if notes == "" {
		notes = req.Reason
	}

	var activeShift models.Shift
	if err := h.db.Where("user_id = ? AND status = ?", userID, "open").First(&activeShift).Error; err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Tidak ada shift aktif untuk mencatat mutasi kas", nil)
	}

	movement := models.CashMovement{
		ShiftID:   activeShift.ID,
		UserID:    userID,
		Type:      mType,
		Amount:    req.Amount,
		Notes:     notes,
		CreatedAt: time.Now(),
	}

	if err := h.db.Create(&movement).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mencatat mutasi kas", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusCreated, "Mutasi kas berhasil dicatat", movement)
}

type CloseShiftRequest struct {
	ActualCash float64 `json:"actual_cash"`
	Notes      string  `json:"notes"`
}

// CloseShift merekonsiliasi fisik uang tunai di laci kasir dengan hitungan sistem dan menutup sesi shift.
func (h *ShiftHandler) CloseShift(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	var req CloseShiftRequest
	if err := c.BodyParser(&req); err != nil || req.ActualCash < 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "actual_cash harus >= 0", nil)
	}

	var shift models.Shift
	if err := h.db.Where("user_id = ? AND status = ?", userID, "open").First(&shift).Error; err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Tidak ada shift terbuka aktif untuk pengguna ini", nil)
	}

	var cashSales float64
	h.db.Model(&models.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.shift_id = ? AND payments.status = 'settlement' AND payments.payment_method = 'cash'", shift.ID).
		Select("COALESCE(SUM(payments.amount_paid - payments.change_due), 0)").
		Scan(&cashSales)

	var cashIn, cashOut float64
	h.db.Model(&models.CashMovement{}).Where("shift_id = ? AND type = 'cash_in'", shift.ID).Select("COALESCE(SUM(amount), 0)").Scan(&cashIn)
	h.db.Model(&models.CashMovement{}).Where("shift_id = ? AND type = 'cash_out'", shift.ID).Select("COALESCE(SUM(amount), 0)").Scan(&cashOut)

	expectedCash := shift.OpeningCash + cashSales + cashIn - cashOut
	diff := req.ActualCash - expectedCash

	now := time.Now()
	shift.EndTime = &now
	shift.ExpectedCash = &expectedCash
	shift.ActualCash = &req.ActualCash
	shift.CashDifference = &diff
	shift.Status = "closed"
	if req.Notes != "" {
		shift.Notes = req.Notes
	}

	if err := h.db.Save(&shift).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menutup shift", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Shift berhasil ditutup dan direkonsiliasi", shift)
}
