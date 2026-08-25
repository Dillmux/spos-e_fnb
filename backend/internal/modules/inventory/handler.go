package inventory

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spos/backend/internal/models"
	"github.com/spos/backend/internal/utils"
	"gorm.io/gorm"
)

type InventoryHandler struct {
	db *gorm.DB
}

func NewInventoryHandler(db *gorm.DB) *InventoryHandler {
	return &InventoryHandler{db: db}
}

type IngredientInput struct {
	Name          string  `json:"name"`
	Unit          string  `json:"unit"`
	Stock         float64 `json:"stock"`
	CurrentStock  float64 `json:"current_stock"`
	MinStock      float64 `json:"min_stock"`
	MinStockAlert float64 `json:"min_stock_alert"`
	COGSPerUnit   float64 `json:"cogs_per_unit"`
	CostPerUnit   float64 `json:"cost_per_unit"`
}

// ListIngredients mengambil daftar bahan baku dan tingkat stok untuk outlet saat ini.
func (h *InventoryHandler) ListIngredients(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	var ingredients []models.Ingredient
	if err := h.db.Where("outlet_id = ?", outletID).Order("name ASC").Find(&ingredients).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengambil daftar bahan baku", err.Error())
	}

	ingResponses := make([]fiber.Map, 0, len(ingredients))
	for _, ing := range ingredients {
		ingResponses = append(ingResponses, fiber.Map{
			"id":              ing.ID,
			"outlet_id":       ing.OutletID,
			"name":            ing.Name,
			"unit":            ing.Unit,
			"stock":           ing.CurrentStock,
			"current_stock":   ing.CurrentStock,
			"min_stock":       ing.MinStockAlert,
			"min_stock_alert": ing.MinStockAlert,
			"cogs_per_unit":   ing.CostPerUnit,
			"cost_per_unit":   ing.CostPerUnit,
			"updated_at":      ing.UpdatedAt,
		})
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Daftar bahan baku berhasil diambil", ingResponses)
}

// CreateIngredient mendaftarkan item bahan baku baru ke dalam outlet.
func (h *InventoryHandler) CreateIngredient(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	var req IngredientInput
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Nama bahan baku wajib diisi", nil)
	}

	if req.Unit == "" {
		req.Unit = "gram"
	}

	stock := req.CurrentStock
	if stock == 0 && req.Stock != 0 {
		stock = req.Stock
	}

	minStock := req.MinStockAlert
	if minStock == 0 && req.MinStock != 0 {
		minStock = req.MinStock
	}

	cost := req.CostPerUnit
	if cost == 0 && req.COGSPerUnit != 0 {
		cost = req.COGSPerUnit
	}

	ingredient := models.Ingredient{
		OutletID:      outletID,
		Name:          req.Name,
		Unit:          req.Unit,
		CurrentStock:  stock,
		MinStockAlert: minStock,
		CostPerUnit:   cost,
		UpdatedAt:     time.Now(),
	}

	if err := h.db.Create(&ingredient).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membuat bahan baku di database", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusCreated, "Bahan baku berhasil dibuat", fiber.Map{
		"id":              ingredient.ID,
		"outlet_id":       ingredient.OutletID,
		"name":            ingredient.Name,
		"unit":            ingredient.Unit,
		"stock":           ingredient.CurrentStock,
		"current_stock":   ingredient.CurrentStock,
		"min_stock":       ingredient.MinStockAlert,
		"min_stock_alert": ingredient.MinStockAlert,
		"cogs_per_unit":   ingredient.CostPerUnit,
		"cost_per_unit":   ingredient.CostPerUnit,
		"updated_at":      ingredient.UpdatedAt,
	})
}

// UpdateIngredient memperbarui metadata atau biaya per unit bahan baku dalam outlet terkait.
func (h *InventoryHandler) UpdateIngredient(c *fiber.Ctx) error {
	id := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)

	var ingredient models.Ingredient
	if err := h.db.Where("id = ? AND outlet_id = ?", id, outletID).First(&ingredient).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Bahan baku tidak ditemukan", nil)
	}

	var req IngredientInput
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON tidak valid", err.Error())
	}

	if req.Name != "" {
		ingredient.Name = req.Name
	}
	if req.Unit != "" {
		ingredient.Unit = req.Unit
	}

	if req.CurrentStock > 0 {
		ingredient.CurrentStock = req.CurrentStock
	} else if req.Stock > 0 {
		ingredient.CurrentStock = req.Stock
	}

	if req.MinStockAlert > 0 {
		ingredient.MinStockAlert = req.MinStockAlert
	} else if req.MinStock > 0 {
		ingredient.MinStockAlert = req.MinStock
	}

	if req.CostPerUnit > 0 {
		ingredient.CostPerUnit = req.CostPerUnit
	} else if req.COGSPerUnit > 0 {
		ingredient.CostPerUnit = req.COGSPerUnit
	}

	ingredient.UpdatedAt = time.Now()

	if err := h.db.Save(&ingredient).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui bahan baku", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Bahan baku berhasil diperbarui", fiber.Map{
		"id":              ingredient.ID,
		"outlet_id":       ingredient.OutletID,
		"name":            ingredient.Name,
		"unit":            ingredient.Unit,
		"stock":           ingredient.CurrentStock,
		"current_stock":   ingredient.CurrentStock,
		"min_stock":       ingredient.MinStockAlert,
		"min_stock_alert": ingredient.MinStockAlert,
		"cogs_per_unit":   ingredient.CostPerUnit,
		"cost_per_unit":   ingredient.CostPerUnit,
		"updated_at":      ingredient.UpdatedAt,
	})
}

// DeleteIngredient menghapus bahan baku secara permanen dari inventaris outlet.
func (h *InventoryHandler) DeleteIngredient(c *fiber.Ctx) error {
	id := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)

	var ingredient models.Ingredient
	if err := h.db.Where("id = ? AND outlet_id = ?", id, outletID).First(&ingredient).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Bahan baku tidak ditemukan", nil)
	}

	if err := h.db.Delete(&ingredient).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menghapus bahan baku", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Bahan baku berhasil dihapus", fiber.Map{"id": id})
}

type StockAdjustmentRequest struct {
	ChangeType     string  `json:"change_type"` // restock, waste, adjustment
	QuantityChange float64 `json:"quantity_change"`
	Notes          string  `json:"notes"`
}

// AdjustStock menangani penyesuaian stok manual (restock, pembuangan/waste, koreksi stok) dengan aman.
func (h *InventoryHandler) AdjustStock(c *fiber.Ctx) error {
	id := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)

	var req StockAdjustmentRequest
	if err := c.BodyParser(&req); err != nil || req.ChangeType == "" || req.QuantityChange == 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "change_type dan quantity_change (bukan nol) wajib diisi", nil)
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ingredient models.Ingredient
	if err := tx.Where("id = ? AND outlet_id = ?", id, outletID).First(&ingredient).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusNotFound, "Bahan baku tidak ditemukan", nil)
	}

	newStock := ingredient.CurrentStock + req.QuantityChange
	if newStock < 0 {
		newStock = 0
	}

	ingredient.CurrentStock = newStock
	now := time.Now()
	ingredient.UpdatedAt = now

	if err := tx.Save(&ingredient).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui stok bahan baku", err.Error())
	}

	logEntry := models.InventoryLog{
		IngredientID:   ingredient.ID,
		ChangeType:     req.ChangeType,
		QuantityChange: req.QuantityChange,
		RemainingStock: newStock,
		Notes:          req.Notes,
		CreatedAt:      now,
	}

	if err := tx.Create(&logEntry).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mencatat log audit inventaris", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menyelesaikan transaksi penyesuaian stok", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Penyesuaian stok berhasil disimpan", fiber.Map{
		"ingredient_id":   ingredient.ID,
		"new_stock":       newStock,
		"quantity_change": req.QuantityChange,
		"change_type":     req.ChangeType,
	})
}

// GetRecipe mengambil pemetaan resep BOM (Bill of Materials) untuk suatu produk.
func (h *InventoryHandler) GetRecipe(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	outletID, _ := c.Locals("outlet_id").(string)

	var product models.Product
	err := h.db.Joins("JOIN categories ON categories.id = products.category_id").
		Where("products.id = ? AND categories.outlet_id = ?", productID, outletID).
		First(&product).Error
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Produk tidak ditemukan pada outlet ini", nil)
	}

	var recipes []models.ProductRecipe
	if err := h.db.Preload("Ingredient").Where("product_id = ?", productID).Find(&recipes).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengambil resep produk", err.Error())
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Resep produk berhasil diambil", recipes)
}

type SetRecipeItemRequest struct {
	IngredientID     string  `json:"ingredient_id"`
	QuantityRequired float64 `json:"quantity_required"`
}

type SetRecipeRequest struct {
	Recipes []SetRecipeItemRequest `json:"recipes"`
}

// SetRecipe mengatur pemetaan bahan resep BOM untuk suatu produk dan menghitung ulang HPP (COGS) secara otomatis.
func (h *InventoryHandler) SetRecipe(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	outletID, _ := c.Locals("outlet_id").(string)

	var req SetRecipeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format JSON tidak valid", err.Error())
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var product models.Product
	err := tx.Joins("JOIN categories ON categories.id = products.category_id").
		Where("products.id = ? AND categories.outlet_id = ?", productID, outletID).
		First(&product).Error
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusNotFound, "Produk tidak ditemukan pada outlet ini", nil)
	}

	// Hapus resep lama untuk produk ini
	if err := tx.Where("product_id = ?", productID).Delete(&models.ProductRecipe{}).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membersihkan resep lama", err.Error())
	}

	var calculatedCOGS float64
	var newRecipes []models.ProductRecipe

	for _, item := range req.Recipes {
		var ingredient models.Ingredient
		if err := tx.Where("id = ? AND outlet_id = ?", item.IngredientID, outletID).First(&ingredient).Error; err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusBadRequest, "Bahan baku tidak ditemukan: "+item.IngredientID, nil)
		}

		recipe := models.ProductRecipe{
			ProductID:        productID,
			IngredientID:     item.IngredientID,
			QuantityRequired: item.QuantityRequired,
		}
		newRecipes = append(newRecipes, recipe)

		calculatedCOGS += item.QuantityRequired * ingredient.CostPerUnit
	}

	if len(newRecipes) > 0 {
		if err := tx.Create(&newRecipes).Error; err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menyimpan resep produk baru", err.Error())
		}
	}

	// Perbarui HPP (COGS) hasil kalkulasi otomatis pada produk
	product.COGS = calculatedCOGS
	product.UpdatedAt = time.Now()
	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui HPP produk", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menyimpan perubahan resep", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Resep produk berhasil diperbarui dan HPP berhasil dihitung", fiber.Map{
		"product_id":      product.ID,
		"calculated_cogs": calculatedCOGS,
		"recipes_count":   len(newRecipes),
	})
}

// ListLogs mengambil riwayat log audit mutasi inventaris stok.
func (h *InventoryHandler) ListLogs(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	var logs []models.InventoryLog
	err := h.db.Joins("JOIN ingredients ON ingredients.id = inventory_logs.ingredient_id").
		Where("ingredients.outlet_id = ?", outletID).
		Preload("Ingredient").
		Preload("OrderItem.Product").
		Order("inventory_logs.created_at DESC").
		Limit(100).
		Find(&logs).Error

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengambil log audit inventaris", err.Error())
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Log audit inventaris berhasil diambil", logs)
}
