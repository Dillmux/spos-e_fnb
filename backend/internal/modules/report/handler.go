package report

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spos/backend/internal/models"
	"github.com/spos/backend/internal/utils"
	"gorm.io/gorm"
)

type ReportHandler struct {
	db *gorm.DB
}

func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

// DailySales menghitung pendapatan harian, jumlah pesanan, pajak, service charge, serta rincian metode pembayaran.
func (h *ReportHandler) DailySales(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	dateStr := c.Query("date")

	var targetDate time.Time
	var err error
	if dateStr != "" {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return utils.SendError(c, fiber.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", nil)
		}
	} else {
		targetDate = time.Now()
	}

	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var totalRevenue float64
	var totalTax float64
	var totalService float64
	var orderCount int64

	h.db.Model(&models.Order{}).
		Where("outlet_id = ? AND status = 'paid' AND created_at >= ? AND created_at < ?", outletID, startOfDay, endOfDay).
		Count(&orderCount)

	h.db.Model(&models.Order{}).
		Where("outlet_id = ? AND status = 'paid' AND created_at >= ? AND created_at < ?", outletID, startOfDay, endOfDay).
		Select("COALESCE(SUM(total_amount), 0), COALESCE(SUM(tax_amount), 0), COALESCE(SUM(service_amount), 0)").
		Row().Scan(&totalRevenue, &totalTax, &totalService)

	// Rincian berdasarkan metode pembayaran
	type PaymentBreakdown struct {
		PaymentMethod string  `json:"payment_method"`
		TotalAmount   float64 `json:"total_amount"`
		Count         int64   `json:"count"`
	}

	var breakdowns []PaymentBreakdown
	h.db.Model(&models.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.outlet_id = ? AND payments.status = 'settlement' AND orders.created_at >= ? AND orders.created_at < ?", outletID, startOfDay, endOfDay).
		Select("payments.payment_method, COALESCE(SUM(payments.amount_paid - payments.change_due), 0) as total_amount, COUNT(payments.id) as count").
		Group("payments.payment_method").
		Scan(&breakdowns)

	return utils.SendSuccess(c, fiber.StatusOK, "Laporan penjualan harian berhasil dibuat", fiber.Map{
		"date":               startOfDay.Format("2006-01-02"),
		"total_orders":       orderCount,
		"total_revenue":      totalRevenue,
		"total_tax":          totalTax,
		"total_service":      totalService,
		"payment_breakdowns": breakdowns,
	})
}

// GrossProfit menghitung laba kotor berdasarkan omzet penjualan bersih dibandingkan HPP (COGS) produk dari resep BOM.
func (h *ReportHandler) GrossProfit(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	now := time.Now()
	startDate := now.AddDate(0, 0, -30)
	endDate := now

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, parsed.Location())
		}
	}

	var totalRevenue float64
	h.db.Model(&models.Order{}).
		Where("outlet_id = ? AND status = 'paid' AND created_at >= ? AND created_at <= ?", outletID, startDate, endDate).
		Select("COALESCE(SUM(total_amount - tax_amount - service_amount), 0)").
		Scan(&totalRevenue)

	// Hitung total HPP dari order_items dan HPP produk
	var totalCOGS float64
	h.db.Model(&models.OrderItem{}).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("orders.outlet_id = ? AND orders.status = 'paid' AND orders.created_at >= ? AND orders.created_at <= ?", outletID, startDate, endDate).
		Select("COALESCE(SUM(order_items.quantity * products.cogs), 0)").
		Scan(&totalCOGS)

	grossProfit := totalRevenue - totalCOGS
	var profitMargin float64
	if totalRevenue > 0 {
		profitMargin = (grossProfit / totalRevenue) * 100.0
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Laporan laba kotor berhasil dibuat", fiber.Map{
		"start_date":    startDate.Format("2006-01-02"),
		"end_date":      endDate.Format("2006-01-02"),
		"net_revenue":   totalRevenue,
		"total_cogs":    totalCOGS,
		"gross_profit":  grossProfit,
		"profit_margin": profitMargin,
	})
}

// TopSelling menghitung daftar produk terlaris berdasarkan kuantitas terjual dan omzet pendapatan.
func (h *ReportHandler) TopSelling(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	dateStr := c.Query("date")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	query := h.db.Model(&models.OrderItem{}).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("orders.outlet_id = ? AND orders.status = 'paid'", outletID)

	if dateStr != "" {
		if targetDate, err := time.Parse("2006-01-02", dateStr); err == nil {
			startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
			endOfDay := startOfDay.Add(24 * time.Hour)
			query = query.Where("orders.created_at >= ? AND orders.created_at < ?", startOfDay, endOfDay)
		}
	} else if startDateStr != "" && endDateStr != "" {
		if startDate, err1 := time.Parse("2006-01-02", startDateStr); err1 == nil {
			if endDate, err2 := time.Parse("2006-01-02", endDateStr); err2 == nil {
				endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
				query = query.Where("orders.created_at >= ? AND orders.created_at <= ?", startDate, endOfDay)
			}
		}
	}

	type TopProduct struct {
		ProductID    string  `json:"product_id"`
		ProductName  string  `json:"product_name"`
		TotalSold    int64   `json:"total_sold"`
		TotalRevenue float64 `json:"total_revenue"`
	}

	var topProducts []TopProduct
	err := query.
		Select("order_items.product_id, products.name as product_name, SUM(order_items.quantity) as total_sold, SUM(order_items.subtotal) as total_revenue").
		Group("order_items.product_id, products.name").
		Order("total_sold DESC").
		Limit(10).
		Scan(&topProducts).Error

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menghitung produk terlaris", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Daftar produk terlaris berhasil diambil", topProducts)
}

// ExportDailySalesCSV mengekspor berkas CSV rekapitulasi penjualan harian.
func (h *ReportHandler) ExportDailySalesCSV(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	dateStr := c.Query("date")

	var targetDate time.Time
	if dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = t
		} else {
			targetDate = time.Now()
		}
	} else {
		targetDate = time.Now()
	}

	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var orders []models.Order
	h.db.Where("outlet_id = ? AND status = 'paid' AND created_at >= ? AND created_at < ?", outletID, startOfDay, endOfDay).
		Preload("OrderItems.Product").
		Preload("Payments").
		Order("created_at ASC").
		Find(&orders)

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=Daily_Sales_%s.csv", startOfDay.Format("20060102")))

	csvContent := "No. Faktur,Waktu Transaksi,Nama Pelanggan,Mode Layanan,Subtotal,Diskon,Pajak,Service,Total Akhir,Metode Pembayaran\n"
	for _, o := range orders {
		payMethod := "Tunai"
		if len(o.Payments) > 0 {
			payMethod = o.Payments[0].PaymentMethod
		}
		csvContent += fmt.Sprintf("%s,%s,\"%s\",%s,%.2f,%.2f,%.2f,%.2f,%.2f,%s\n",
			o.InvoiceNumber,
			o.CreatedAt.Format("15:04:05"),
			o.CustomerName,
			o.ServiceType,
			o.Subtotal,
			o.DiscountAmount,
			o.TaxAmount,
			o.ServiceAmount,
			o.TotalAmount,
			payMethod,
		)
	}

	return c.SendString(csvContent)
}

// ExportGrossProfitCSV mengekspor berkas CSV rekapitulasi laba kotor dan HPP.
func (h *ReportHandler) ExportGrossProfitCSV(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	now := time.Now()
	startDate := now.AddDate(0, 0, -30)
	endDate := now

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, parsed.Location())
		}
	}

	type ItemCOGS struct {
		ProductName  string  `json:"product_name"`
		TotalQty     int64   `json:"total_qty"`
		SellingPrice float64 `json:"selling_price"`
		UnitCOGS     float64 `json:"unit_cogs"`
		TotalRevenue float64 `json:"total_revenue"`
		TotalCOGS    float64 `json:"total_cogs"`
	}

	var items []ItemCOGS
	h.db.Model(&models.OrderItem{}).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("orders.outlet_id = ? AND orders.status = 'paid' AND orders.created_at >= ? AND orders.created_at <= ?", outletID, startDate, endDate).
		Select("products.name as product_name, SUM(order_items.quantity) as total_qty, products.price as selling_price, products.cogs as unit_cogs, SUM(order_items.subtotal) as total_revenue, SUM(order_items.quantity * products.cogs) as total_cogs").
		Group("products.name, products.price, products.cogs").
		Scan(&items)

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=Gross_Profit_%s_%s.csv", startDate.Format("20060102"), endDate.Format("20060102")))

	csvContent := "Nama Menu,Total Terjual,Harga Jual Unit,HPP (COGS) Unit,Total Omzet Bersih,Total HPP,Laba Kotor Unit,Margin Laba (%)\n"
	for _, it := range items {
		gross := it.TotalRevenue - it.TotalCOGS
		margin := 0.0
		if it.TotalRevenue > 0 {
			margin = (gross / it.TotalRevenue) * 100.0
		}
		csvContent += fmt.Sprintf("\"%s\",%d,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f%%\n",
			it.ProductName,
			it.TotalQty,
			it.SellingPrice,
			it.UnitCOGS,
			it.TotalRevenue,
			it.TotalCOGS,
			gross,
			margin,
		)
	}

	return c.SendString(csvContent)
}


