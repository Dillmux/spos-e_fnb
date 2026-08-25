package pos

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spos/backend/internal/middleware"
	"github.com/spos/backend/internal/models"
	"github.com/spos/backend/internal/utils"
	"gorm.io/gorm"
)

type POSHandler struct {
	db *gorm.DB
}

func NewPOSHandler(db *gorm.DB) *POSHandler {
	return &POSHandler{db: db}
}

// -----------------------------------------------------------------------------
// MANAJEMEN MEJA
// -----------------------------------------------------------------------------

// ListTables mengambil daftar seluruh meja pada outlet pengguna saat ini.
func (h *POSHandler) ListTables(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	var tables []models.Table
	if err := h.db.Where("outlet_id = ?", outletID).Order("table_number ASC").Find(&tables).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengambil daftar meja", err.Error())
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Daftar meja berhasil diambil", tables)
}

type CreateTableRequest struct {
	TableNumber string `json:"table_number"`
	Capacity    int    `json:"capacity"`
}

// CreateTable menambahkan meja baru ke dalam outlet.
func (h *POSHandler) CreateTable(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	var req CreateTableRequest
	if err := c.BodyParser(&req); err != nil || req.TableNumber == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Nomor meja atau payload tidak valid", nil)
	}

	if req.Capacity <= 0 {
		req.Capacity = 2
	}

	table := models.Table{
		OutletID:    outletID,
		TableNumber: req.TableNumber,
		Capacity:    req.Capacity,
		Status:      "available",
	}

	if err := h.db.Create(&table).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membuat meja baru", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusCreated, "Meja berhasil dibuat", table)
}

type UpdateTableRequest struct {
	TableNumber string `json:"table_number"`
	Capacity    int    `json:"capacity"`
	Status      string `json:"status"` // available, occupied, reserved
}

// UpdateTable memperbarui informasi meja milik outlet yang terotentikasi.
func (h *POSHandler) UpdateTable(c *fiber.Ctx) error {
	idStr := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID meja tidak valid", nil)
	}

	var req UpdateTableRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON tidak valid", err.Error())
	}

	var table models.Table
	if err := h.db.Where("id = ? AND outlet_id = ?", id, outletID).First(&table).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Meja tidak ditemukan pada outlet ini", nil)
	}

	if req.TableNumber != "" {
		table.TableNumber = req.TableNumber
	}
	if req.Capacity > 0 {
		table.Capacity = req.Capacity
	}
	if req.Status != "" {
		table.Status = req.Status
	}

	if err := h.db.Save(&table).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui informasi meja", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Informasi meja berhasil diperbarui", table)
}

// DeleteTable menghapus meja milik outlet yang terotentikasi.
func (h *POSHandler) DeleteTable(c *fiber.Ctx) error {
	idStr := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID meja tidak valid", nil)
	}

	var table models.Table
	if err := h.db.Where("id = ? AND outlet_id = ?", id, outletID).First(&table).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Meja tidak ditemukan pada outlet ini", nil)
	}

	if err := h.db.Delete(&table).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menghapus meja", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Meja berhasil dihapus", nil)
}

// -----------------------------------------------------------------------------
// KATALOG MENU & KATEGORI
// -----------------------------------------------------------------------------

// ListCategories mengambil daftar kategori menu untuk outlet pengguna.
func (h *POSHandler) ListCategories(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	var categories []models.Category
	if err := h.db.Where("outlet_id = ?", outletID).Find(&categories).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengambil daftar kategori", err.Error())
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Daftar kategori berhasil diambil", categories)
}

// CreateCategory membuat kategori menu baru pada outlet pengguna.
func (h *POSHandler) CreateCategory(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	var category models.Category
	if err := c.BodyParser(&category); err != nil || category.Name == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Nama kategori wajib diisi", nil)
	}
	category.OutletID = outletID
	if err := h.db.Create(&category).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membuat kategori baru", err.Error())
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Kategori berhasil dibuat", category)
}

// UpdateCategory memperbarui nama kategori menu.
func (h *POSHandler) UpdateCategory(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	catID := c.Params("id")
	if catID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "ID Kategori tidak valid", nil)
	}

	var category models.Category
	if err := h.db.Where("id = ? AND outlet_id = ?", catID, outletID).First(&category).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Kategori tidak ditemukan", err.Error())
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Payload JSON tidak valid", err.Error())
	}

	if req.Name != "" {
		category.Name = req.Name
	}

	if err := h.db.Save(&category).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui kategori", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Kategori berhasil diperbarui", category)
}

// DeleteCategory menghapus kategori menu jika tidak terikat ke produk menu aktif.
func (h *POSHandler) DeleteCategory(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	catID := c.Params("id")
	if catID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "ID Kategori tidak valid", nil)
	}

	var category models.Category
	if err := h.db.Where("id = ? AND outlet_id = ?", catID, outletID).First(&category).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Kategori tidak ditemukan", err.Error())
	}

	var prodCount int64
	h.db.Model(&models.Product{}).Where("category_id = ?", catID).Count(&prodCount)
	if prodCount > 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "Kategori tidak dapat dihapus karena masih digunakan oleh produk menu aktif", nil)
	}

	if err := h.db.Delete(&category).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menghapus kategori", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Kategori berhasil dihapus", nil)
}

// ListProducts mengambil daftar produk menu beserta resep BOM dan kategori.
func (h *POSHandler) ListProducts(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	var products []models.Product
	err := h.db.Joins("JOIN categories ON categories.id = products.category_id").
		Where("categories.outlet_id = ?", outletID).
		Preload("Category").
		Preload("Recipes.Ingredient").
		Find(&products).Error

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengambil daftar produk menu", err.Error())
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Daftar produk menu berhasil diambil", products)
}

type ProductInput struct {
	CategoryID  interface{} `json:"category_id"`
	SKU         string      `json:"sku"`
	Name        string      `json:"name"`
	Price       float64     `json:"price"`
	COGS        float64     `json:"cogs"`
	IsAvailable *bool       `json:"is_available"`
}

func parseCategoryID(val interface{}) uint {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return uint(v)
	case int:
		return uint(v)
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			return uint(n)
		}
		numStr := ""
		for _, r := range v {
			if r >= '0' && r <= '9' {
				numStr += string(r)
			}
		}
		if numStr != "" {
			if n, err := strconv.Atoi(numStr); err == nil {
				return uint(n)
			}
		}
	}
	return 0
}

// CreateProduct menambahkan item produk menu baru ke dalam katalog outlet.
func (h *POSHandler) CreateProduct(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)

	var req ProductInput
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format payload menu tidak valid", err.Error())
	}

	if req.Name == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Nama menu wajib diisi", nil)
	}

	if req.Price < 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "Harga menu tidak boleh negatif", nil)
	}

	catID := parseCategoryID(req.CategoryID)
	if catID == 0 {
		var firstCat models.Category
		if err := h.db.Where("outlet_id = ?", outletID).First(&firstCat).Error; err == nil {
			catID = firstCat.ID
		} else {
			newCat := models.Category{
				OutletID: outletID,
				Name:     "Umum / Resto",
			}
			if err := h.db.Create(&newCat).Error; err == nil {
				catID = newCat.ID
			}
		}
	}

	isAvail := true
	if req.IsAvailable != nil {
		isAvail = *req.IsAvailable
	}

	product := models.Product{
		CategoryID:  catID,
		SKU:         req.SKU,
		Name:        req.Name,
		Price:       req.Price,
		COGS:        req.COGS,
		IsAvailable: isAvail,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.db.Create(&product).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membuat menu baru di database", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusCreated, "Menu berhasil dibuat", product)
}

// UpdateProduct memperbarui informasi produk menu pada outlet terotentikasi.
func (h *POSHandler) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)

	var product models.Product
	err := h.db.Joins("JOIN categories ON categories.id = products.category_id").
		Where("products.id = ? AND categories.outlet_id = ?", id, outletID).
		First(&product).Error
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Menu tidak ditemukan pada outlet ini", nil)
	}

	var req ProductInput
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format payload update menu tidak valid", err.Error())
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Price >= 0 {
		product.Price = req.Price
	}
	if req.COGS >= 0 {
		product.COGS = req.COGS
	}
	if req.SKU != "" {
		product.SKU = req.SKU
	}
	if req.IsAvailable != nil {
		product.IsAvailable = *req.IsAvailable
	}

	catID := parseCategoryID(req.CategoryID)
	if catID > 0 {
		product.CategoryID = catID
	}

	product.UpdatedAt = time.Now()

	if err := h.db.Save(&product).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui menu", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Menu berhasil diperbarui", product)
}

// DeleteProduct menghapus produk menu secara permanen dari outlet terotentikasi.
func (h *POSHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)

	var product models.Product
	err := h.db.Joins("JOIN categories ON categories.id = products.category_id").
		Where("products.id = ? AND categories.outlet_id = ?", id, outletID).
		First(&product).Error
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Menu tidak ditemukan pada outlet ini", nil)
	}

	if err := h.db.Delete(&product).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menghapus menu", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Menu berhasil dihapus permanen", nil)
}

// -----------------------------------------------------------------------------
// ORDERS & BILL OPERATIONS (CREATE, SPLIT, MERGE, VOID, KDS)
// -----------------------------------------------------------------------------

type CreateOrderItemRequest struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	ItemNotes string  `json:"item_notes"`
}

type CreateOrderRequest struct {
	TableID        interface{}              `json:"table_id"`
	ServiceType    string                   `json:"service_type"` // dine_in, take_away, delivery
	CustomerName   string                   `json:"customer_name"`
	DiscountAmount float64                  `json:"discount_amount"`
	Items          []CreateOrderItemRequest `json:"items"`
}

func (h *POSHandler) CreateOrder(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	outletID, _ := c.Locals("outlet_id").(string)

	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil || len(req.Items) == 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "Order must contain at least one item", nil)
	}

	// Parse TableID flexibly (supports float64, string, int)
	var tableID *uint
	if req.TableID != nil {
		switch v := req.TableID.(type) {
		case float64:
			u := uint(v)
			tableID = &u
		case string:
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				u := uint(parsed)
				tableID = &u
			}
		case int:
			u := uint(v)
			tableID = &u
		}
	}

	// Verify active shift for current user
	var activeShift models.Shift
	if err := h.db.Where("user_id = ? AND status = ?", userID, "open").First(&activeShift).Error; err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Cashier must open a shift before processing orders", nil)
	}

	// Fetch Outlet tax & service config
	var outlet models.Outlet
	if err := h.db.First(&outlet, "id = ?", outletID).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to fetch outlet settings", err.Error())
	}

	if req.ServiceType == "" {
		req.ServiceType = "dine_in"
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var subtotal float64
	var orderItems []models.OrderItem
	now := time.Now()

	// 1. Calculate Aggregate BOM Ingredient Demand across ALL items in order
	type IngDemand struct {
		Name         string
		Unit         string
		CurrentStock float64
		TotalNeeded  float64
	}
	ingredientDemands := make(map[string]*IngDemand)

	for _, itemReq := range req.Items {
		var product models.Product
		if err := h.db.Preload("Recipes.Ingredient").First(&product, "id = ?", itemReq.ProductID).Error; err == nil {
			if !product.IsAvailable {
				tx.Rollback()
				return utils.SendError(c, fiber.StatusBadRequest, fmt.Sprintf("Menu '%s' sedang dinonaktifkan / tidak dapat dipesan", product.Name), nil)
			}
			for _, recipe := range product.Recipes {
				if recipe.Ingredient != nil {
					needed := float64(itemReq.Quantity) * recipe.QuantityRequired
					if d, exists := ingredientDemands[recipe.IngredientID]; exists {
						d.TotalNeeded += needed
					} else {
						ingredientDemands[recipe.IngredientID] = &IngDemand{
							Name:         recipe.Ingredient.Name,
							Unit:         recipe.Ingredient.Unit,
							CurrentStock: recipe.Ingredient.CurrentStock,
							TotalNeeded:  needed,
						}
					}
				}
			}
		}

		itemSubtotal := float64(itemReq.Quantity) * itemReq.UnitPrice
		subtotal += itemSubtotal
		orderItems = append(orderItems, models.OrderItem{
			ProductID:     itemReq.ProductID,
			Quantity:      itemReq.Quantity,
			UnitPrice:     itemReq.UnitPrice,
			Subtotal:      itemSubtotal,
			ItemNotes:     itemReq.ItemNotes,
			KitchenStatus: "queued",
			CreatedAt:     now,
		})
	}

	// 2. Verify Aggregate Demand against Available Current Stock
	for _, d := range ingredientDemands {
		if d.CurrentStock <= 0 || d.CurrentStock < d.TotalNeeded {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusBadRequest, fmt.Sprintf("Pesanan ditolak! Total kebutuhan bahan baku '%s' untuk seluruh pesanan (%.2f %s) melebihi sisa stok yang tersedia (%.2f %s)", d.Name, d.TotalNeeded, d.Unit, d.CurrentStock, d.Unit), nil)
		}
	}

	// Calculate Tax and Service Percentage
	taxableAmount := subtotal - req.DiscountAmount
	if taxableAmount < 0 {
		taxableAmount = 0
	}

	serviceAmount := 0.0
	if outlet.ServicePercentage > 0 {
		serviceAmount = taxableAmount * (outlet.ServicePercentage / 100.0)
	}

	taxAmount := 0.0
	if outlet.TaxPercentage > 0 && !outlet.IsTaxInclusive {
		taxAmount = (taxableAmount + serviceAmount) * (outlet.TaxPercentage / 100.0)
	}

	totalAmount := taxableAmount + serviceAmount + taxAmount

	// Hitung Nomor Antrean Harian (di-reset setiap hari per outlet)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var maxQueue int
	tx.Model(&models.Order{}).
		Where("outlet_id = ? AND created_at >= ? AND created_at < ?", outletID, startOfDay, endOfDay).
		Select("COALESCE(MAX(queue_number), 0)").
		Scan(&maxQueue)

	queueNum := maxQueue + 1

	// Generate Invoice Number: INV/YYYYMMDD/RAND
	invNum := fmt.Sprintf("INV/%s/%04d", now.Format("20060102"), rand.Intn(10000))

	order := models.Order{
		InvoiceNumber:  invNum,
		OutletID:       outletID,
		ShiftID:        activeShift.ID,
		UserID:         userID,
		TableID:        tableID,
		ServiceType:    req.ServiceType,
		CustomerName:   req.CustomerName,
		Subtotal:       subtotal,
		DiscountAmount: req.DiscountAmount,
		TaxAmount:      taxAmount,
		ServiceAmount:  serviceAmount,
		TotalAmount:    totalAmount,
		QueueNumber:    queueNum,
		Status:         "open",
		CreatedAt:      now,
		UpdatedAt:      now,
		OrderItems:     orderItems,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create order", err.Error())
	}

	// Automatically mark table occupied if tableID is present
	if tableID != nil && *tableID > 0 {
		tx.Model(&models.Table{}).Where("id = ?", *tableID).Update("status", "occupied")
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to commit order creation", err.Error())
	}

	h.db.Preload("OrderItems.Product").Preload("Table").First(&order, "id = ?", order.ID)
	utils.GetSSEBroker().Broadcast("order_created", fmt.Sprintf("{\"order_id\":\"%s\",\"outlet_id\":\"%s\"}", order.ID, outletID))
	return utils.SendSuccess(c, fiber.StatusCreated, "Order created successfully", order)
}

func (h *POSHandler) ListOrders(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	status := c.Query("status")
	shiftID := c.Query("shift_id")
	dateStr := c.Query("date")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	query := h.db.Where("outlet_id = ?", outletID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if shiftID != "" {
		query = query.Where("shift_id = ?", shiftID)
	}

	if dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			endOfDay := startOfDay.Add(24 * time.Hour)
			query = query.Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay)
		}
	} else {
		if startDateStr != "" {
			if st, err := time.Parse("2006-01-02", startDateStr); err == nil {
				startOfDay := time.Date(st.Year(), st.Month(), st.Day(), 0, 0, 0, 0, time.UTC)
				query = query.Where("created_at >= ?", startOfDay)
			}
		}
		if endDateStr != "" {
			if et, err := time.Parse("2006-01-02", endDateStr); err == nil {
				endOfDay := time.Date(et.Year(), et.Month(), et.Day(), 23, 59, 59, 999999999, time.UTC)
				query = query.Where("created_at <= ?", endOfDay)
			}
		}
	}

	var orders []models.Order
	err := query.Preload("Table").
		Preload("OrderItems.Product").
		Preload("Payments").
		Preload("User").
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengambil daftar pesanan", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Daftar pesanan berhasil diambil", orders)
}

// GetOrder mengambil detail pesanan spesifik yang terasosiasi dengan outlet pengguna terotentikasi.
func (h *POSHandler) GetOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)

	var order models.Order
	err := h.db.Where("id = ? AND outlet_id = ?", id, outletID).
		Preload("Table").
		Preload("OrderItems.Product").
		Preload("Payments").
		Preload("User").
		First(&order).Error

	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Pesanan tidak ditemukan pada outlet ini", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Detail pesanan berhasil diambil", order)
}

type SplitBillItemRequest struct {
	OrderItemID string `json:"order_item_id"`
	Quantity    int    `json:"quantity"`
}

type SplitBillRequest struct {
	Items []SplitBillItemRequest `json:"items"`
}

// SplitBill memisahkan item tagihan pesanan ke dalam pesanan anak (terhubung via ParentOrderID) pada outlet yang sama.
func (h *POSHandler) SplitBill(c *fiber.Ctx) error {
	parentOrderID := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)

	var req SplitBillRequest
	if err := c.BodyParser(&req); err != nil || len(req.Items) == 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "Pilih setidaknya satu item untuk dipisahkan", nil)
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var parentOrder models.Order
	if err := tx.Where("id = ? AND outlet_id = ?", parentOrderID, outletID).Preload("OrderItems").First(&parentOrder).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusNotFound, "Pesanan induk tidak ditemukan pada outlet ini", nil)
	}

	if parentOrder.Status == "paid" || parentOrder.Status == "cancelled" {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusBadRequest, "Tidak dapat memisah tagihan yang sudah dibayar atau dibatalkan", nil)
	}

	var childItems []models.OrderItem
	now := time.Now()

	for _, splitReq := range req.Items {
		var origItem models.OrderItem
		if err := tx.First(&origItem, "id = ? AND order_id = ?", splitReq.OrderItemID, parentOrderID).Error; err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusBadRequest, "Item tidak ditemukan dalam pesanan induk", nil)
		}

		if splitReq.Quantity >= origItem.Quantity {
			origItem.OrderID = ""
			childItems = append(childItems, origItem)
			tx.Delete(&origItem)
		} else {
			origItem.Quantity -= splitReq.Quantity
			origItem.Subtotal = float64(origItem.Quantity) * origItem.UnitPrice
			tx.Save(&origItem)

			childItems = append(childItems, models.OrderItem{
				ProductID:     origItem.ProductID,
				Quantity:      splitReq.Quantity,
				UnitPrice:     origItem.UnitPrice,
				Subtotal:      float64(splitReq.Quantity) * origItem.UnitPrice,
				ItemNotes:     origItem.ItemNotes,
				KitchenStatus: origItem.KitchenStatus,
				CreatedAt:     now,
			})
		}
	}

	var outlet models.Outlet
	tx.First(&outlet, "id = ?", parentOrder.OutletID)

	var childSubtotal float64
	for _, ci := range childItems {
		childSubtotal += ci.Subtotal
	}

	childTaxable := childSubtotal
	childService := 0.0
	if outlet.ServicePercentage > 0 {
		childService = childTaxable * (outlet.ServicePercentage / 100.0)
	}
	childTax := 0.0
	if outlet.TaxPercentage > 0 && !outlet.IsTaxInclusive {
		childTax = (childTaxable + childService) * (outlet.TaxPercentage / 100.0)
	}
	childTotal := childTaxable + childService + childTax

	childInv := fmt.Sprintf("%s-S1", parentOrder.InvoiceNumber)

	childOrder := models.Order{
		InvoiceNumber:  childInv,
		ParentOrderID:  &parentOrder.ID,
		OutletID:       parentOrder.OutletID,
		ShiftID:        parentOrder.ShiftID,
		UserID:         parentOrder.UserID,
		TableID:        parentOrder.TableID,
		ServiceType:    parentOrder.ServiceType,
		CustomerName:   fmt.Sprintf("%s (Split)", parentOrder.CustomerName),
		Subtotal:       childSubtotal,
		DiscountAmount: 0,
		TaxAmount:      childTax,
		ServiceAmount:  childService,
		TotalAmount:    childTotal,
		QueueNumber:    parentOrder.QueueNumber,
		Status:         "open",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := tx.Create(&childOrder).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membuat pesanan pisah tagihan baru", err.Error())
	}

	for i := range childItems {
		childItems[i].OrderID = childOrder.ID
		if err := tx.Create(&childItems[i]).Error; err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mendaftarkan item ke pesanan pisah tagihan", err.Error())
		}
	}

	var remainingItems []models.OrderItem
	tx.Where("order_id = ?", parentOrder.ID).Find(&remainingItems)

	if len(remainingItems) == 0 {
		parentOrder.Status = "split"
	} else {
		var remSubtotal float64
		for _, ri := range remainingItems {
			remSubtotal += ri.Subtotal
		}
		remTaxable := remSubtotal - parentOrder.DiscountAmount
		if remTaxable < 0 {
			remTaxable = 0
		}
		remService := 0.0
		if outlet.ServicePercentage > 0 {
			remService = remTaxable * (outlet.ServicePercentage / 100.0)
		}
		remTax := 0.0
		if outlet.TaxPercentage > 0 && !outlet.IsTaxInclusive {
			remTax = (remTaxable + remService) * (outlet.TaxPercentage / 100.0)
		}
		parentOrder.Subtotal = remSubtotal
		parentOrder.ServiceAmount = remService
		parentOrder.TaxAmount = remTax
		parentOrder.TotalAmount = remTaxable + remService + remTax
	}
	parentOrder.UpdatedAt = now
	tx.Save(&parentOrder)

	if err := tx.Commit().Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menyelesaikan transaksi pisah tagihan", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Tagihan berhasil dipisahkan", fiber.Map{
		"parent_order": parentOrder,
		"child_order":  childOrder,
	})
}

type MergeBillRequest struct {
	TargetOrderID  string   `json:"target_order_id"`
	SourceOrderIDs []string `json:"source_order_ids"`
}

// MergeBill menggabungkan item pesanan dari beberapa pesanan asal ke pesanan tujuan pada outlet terotentikasi.
func (h *POSHandler) MergeBill(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)

	var req MergeBillRequest
	if err := c.BodyParser(&req); err != nil || req.TargetOrderID == "" || len(req.SourceOrderIDs) == 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "Pesanan tujuan dan pesanan asal wajib diisi", nil)
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var targetOrder models.Order
	if err := tx.Where("id = ? AND outlet_id = ?", req.TargetOrderID, outletID).First(&targetOrder).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusNotFound, "Pesanan tujuan tidak ditemukan pada outlet ini", nil)
	}

	for _, srcID := range req.SourceOrderIDs {
		if srcID == req.TargetOrderID {
			continue
		}

		var srcOrder models.Order
		if err := tx.Where("id = ? AND outlet_id = ?", srcID, outletID).First(&srcOrder).Error; err != nil {
			continue
		}

		if err := tx.Model(&models.OrderItem{}).Where("order_id = ?", srcID).Update("order_id", req.TargetOrderID).Error; err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memindahkan item pesanan", err.Error())
		}
		tx.Model(&models.Order{}).Where("id = ?", srcID).Updates(map[string]interface{}{
			"status":     "cancelled",
			"updated_at": time.Now(),
		})
	}

	var allItems []models.OrderItem
	tx.Where("order_id = ?", targetOrder.ID).Find(&allItems)

	var subtotal float64
	for _, item := range allItems {
		subtotal += item.Subtotal
	}

	var outlet models.Outlet
	tx.First(&outlet, "id = ?", targetOrder.OutletID)

	taxable := subtotal - targetOrder.DiscountAmount
	if taxable < 0 {
		taxable = 0
	}
	service := 0.0
	if outlet.ServicePercentage > 0 {
		service = taxable * (outlet.ServicePercentage / 100.0)
	}
	tax := 0.0
	if outlet.TaxPercentage > 0 && !outlet.IsTaxInclusive {
		tax = (taxable + service) * (outlet.TaxPercentage / 100.0)
	}

	targetOrder.Subtotal = subtotal
	targetOrder.ServiceAmount = service
	targetOrder.TaxAmount = tax
	targetOrder.TotalAmount = taxable + service + tax
	targetOrder.UpdatedAt = time.Now()

	if err := tx.Save(&targetOrder).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui pesanan gabungan tujuan", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menyelesaikan transaksi gabung tagihan", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Pesanan berhasil digabungkan", targetOrder)
}

type VoidOrderRequest struct {
	PIN    string `json:"pin"`
	Reason string `json:"reason"`
}

func restoreBOMStockOnVoid(tx *gorm.DB, orderID string) error {
	var items []models.OrderItem
	if err := tx.Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return err
	}

	ingredientRestorations := make(map[string]float64)
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
			restoreQty := float64(item.Quantity) * recipe.QuantityRequired
			ingredientRestorations[recipe.IngredientID] += restoreQty
			note := fmt.Sprintf("Void order '%s' (%d pcs x %.2f)", productName, item.Quantity, recipe.QuantityRequired)
			ingredientItemNotes[recipe.IngredientID] = append(ingredientItemNotes[recipe.IngredientID], note)
		}
	}

	for ingredientID, totalRestoreQty := range ingredientRestorations {
		var ingredient models.Ingredient
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&ingredient, "id = ?", ingredientID).Error; err != nil {
			return fmt.Errorf("ingredient %s not found: %w", ingredientID, err)
		}

		restoredStock := ingredient.CurrentStock + totalRestoreQty
		ingredient.CurrentStock = restoredStock
		ingredient.UpdatedAt = now

		if err := tx.Save(&ingredient).Error; err != nil {
			return fmt.Errorf("failed to restore ingredient stock for %s: %w", ingredient.Name, err)
		}

		noteStr := strings.Join(ingredientItemNotes[ingredientID], "; ")
		if len(noteStr) > 250 {
			noteStr = noteStr[:247] + "..."
		}

		logEntry := models.InventoryLog{
			IngredientID:   ingredient.ID,
			ChangeType:     "void_restoration",
			QuantityChange: totalRestoreQty,
			RemainingStock: restoredStock,
			Notes:          "POS void stock restoration: " + noteStr,
			CreatedAt:      now,
		}

		if err := tx.Create(&logEntry).Error; err != nil {
			return fmt.Errorf("failed to record restoration log for %s: %w", ingredient.Name, err)
		}
	}

	return nil
}

// VoidOrder membatalkan pesanan dengan verifikasi PIN Supervisor dan batasan usia pesanan maksimal 30 menit.
func (h *POSHandler) VoidOrder(c *fiber.Ctx) error {
	orderID := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)

	var req VoidOrderRequest
	if err := c.BodyParser(&req); err != nil || req.PIN == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "PIN Supervisor wajib diisi untuk membatalkan pesanan", nil)
	}

	valid, supervisor, err := middleware.VerifySupervisorPIN(h.db, outletID, req.PIN)
	if err != nil || !valid {
		return utils.SendError(c, fiber.StatusForbidden, "PIN Supervisor tidak valid", nil)
	}

	var order models.Order
	if err := h.db.Where("id = ? AND outlet_id = ?", orderID, outletID).First(&order).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Pesanan tidak ditemukan pada outlet ini", nil)
	}

	if order.Status == "cancelled" || order.Status == "void" {
		return utils.SendError(c, fiber.StatusBadRequest, "Pesanan ini sudah dibatalkan sebelumnya", nil)
	}

	if time.Since(order.CreatedAt) > 30*time.Minute {
		return utils.SendError(c, fiber.StatusBadRequest, "Pesanan tidak dapat dibatalkan karena telah berusia lebih dari 30 menit", nil)
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if order.Status == "paid" || order.Status == "completed" {
		if err := restoreBOMStockOnVoid(tx, order.ID); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengembalikan stok bahan baku saat void pesanan", err)
		}
	}

	order.Status = "cancelled"
	order.UpdatedAt = time.Now()
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membatalkan pesanan", err.Error())
	}

	// Release table if dine_in
	if order.TableID != nil && *order.TableID > 0 {
		tx.Model(&models.Table{}).Where("id = ?", *order.TableID).Update("status", "available")
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Transaction commit failed", err.Error())
	}

	utils.GetSSEBroker().Broadcast("order_updated", fmt.Sprintf("{\"order_id\":\"%s\",\"status\":\"cancelled\"}", order.ID))


	return utils.SendSuccess(c, fiber.StatusOK, "Pesanan berhasil dibatalkan", fiber.Map{
		"order_id":    order.ID,
		"status":      "cancelled",
		"voided_by":   supervisor.Name,
		"void_reason": req.Reason,
	})
}

type UpdateKitchenStatusRequest struct {
	KitchenStatus string `json:"kitchen_status"` // queued, cooking, ready, served
}

// UpdateKitchenStatus updates KDS preparation state of order items.
func (h *POSHandler) UpdateKitchenStatus(c *fiber.Ctx) error {
	itemID := c.Params("item_id")
	var req UpdateKitchenStatusRequest
	if err := c.BodyParser(&req); err != nil || req.KitchenStatus == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "kitchen_status is required", nil)
	}

	var item models.OrderItem
	if err := h.db.First(&item, "id = ?", itemID).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Order item not found", nil)
	}

	item.KitchenStatus = req.KitchenStatus
	if err := h.db.Save(&item).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update kitchen status", err.Error())
	}

	utils.GetSSEBroker().Broadcast("kitchen_status_updated", fmt.Sprintf("{\"order_id\":\"%s\",\"item_id\":\"%s\",\"kitchen_status\":\"%s\"}", item.OrderID, item.ID, item.KitchenStatus))

	return utils.SendSuccess(c, fiber.StatusOK, "Kitchen status updated", item)
}

// CompleteOrderItems marks all order items as 'served' when order pickup is confirmed in WLD.
func (h *POSHandler) CompleteOrderItems(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if err := h.db.Model(&models.OrderItem{}).Where("order_id = ?", orderID).Update("kitchen_status", "served").Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to complete order pickup", err.Error())
	}
	utils.GetSSEBroker().Broadcast("order_updated", fmt.Sprintf("{\"order_id\":\"%s\",\"status\":\"served\"}", orderID))
	return utils.SendSuccess(c, fiber.StatusOK, "Order marked as completed/served", fiber.Map{"order_id": orderID})
}

