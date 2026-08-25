package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spos/backend/internal/config"
	"github.com/spos/backend/internal/database"
	"github.com/spos/backend/internal/middleware"
	"github.com/spos/backend/internal/modules/auth"
	"github.com/spos/backend/internal/modules/inventory"
	"github.com/spos/backend/internal/modules/payment"
	"github.com/spos/backend/internal/modules/pos"
	"github.com/spos/backend/internal/modules/report"
	"github.com/spos/backend/internal/modules/setup"
	"github.com/spos/backend/internal/modules/shift"
	"github.com/spos/backend/internal/utils"
)

func setupTestApp(t *testing.T) (*fiber.App, string) {
	dbFile := "test_spos.db"
	_ = os.Remove(dbFile)

	cfg := &config.Config{
		Port:                 "8088",
		DBDriver:             "sqlite",
		DBDsn:                dbFile,
		JWTSecret:            "test-jwt-secret-key-12345",
		MidtransServerKey:    "SB-Mid-server-testkey",
		MidtransClientKey:    "SB-Mid-client-testkey",
		MidtransIsProduction: false,
	}

	db, err := database.InitDB(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize test DB: %v", err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return utils.SendError(c, code, err.Error(), nil)
		},
	})

	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(middleware.SetupGuard(db))

	setupH := setup.NewSetupHandler(db, cfg.JWTSecret)
	authH := auth.NewAuthHandler(db, cfg.JWTSecret)
	posH := pos.NewPOSHandler(db)
	payH := payment.NewPaymentHandler(db, cfg)
	invH := inventory.NewInventoryHandler(db)
	shiftH := shift.NewShiftHandler(db)
	reportH := report.NewReportHandler(db)

	api := app.Group("/api/v1")

	// Setup
	api.Post("/setup/test-db", setupH.TestDB)
	api.Post("/setup/test-midtrans", setupH.TestMidtrans)
	api.Post("/setup/initialize", setupH.Initialize)

	// Auth
	api.Post("/auth/login", authH.Login)
	jwtGuard := middleware.JWTMiddleware(cfg.JWTSecret)
	api.Post("/auth/verify-pin", jwtGuard, authH.VerifyPIN)
	api.Get("/auth/me", jwtGuard, authH.Me)

	// POS
	posRoutes := api.Group("/pos", jwtGuard)
	posRoutes.Get("/tables", posH.ListTables)
	posRoutes.Post("/tables", posH.CreateTable)
	posRoutes.Put("/tables/:id", posH.UpdateTable)
	posRoutes.Get("/categories", posH.ListCategories)
	posRoutes.Post("/categories", posH.CreateCategory)
	posRoutes.Get("/products", posH.ListProducts)
	posRoutes.Post("/products", posH.CreateProduct)
	posRoutes.Put("/products/:id", posH.UpdateProduct)
	posRoutes.Post("/orders", posH.CreateOrder)
	posRoutes.Get("/orders", posH.ListOrders)
	posRoutes.Get("/orders/:id", posH.GetOrder)
	posRoutes.Post("/orders/:id/split", posH.SplitBill)
	posRoutes.Post("/orders/merge", posH.MergeBill)
	posRoutes.Post("/orders/:id/void", posH.VoidOrder)
	posRoutes.Put("/orders/items/:item_id/status", posH.UpdateKitchenStatus)

	// Payment
	payRoutes := api.Group("/payment")
	payRoutes.Post("/cash", jwtGuard, payH.CashCheckout)
	payRoutes.Post("/midtrans/charge", jwtGuard, payH.MidtransCharge)
	payRoutes.Post("/midtrans/webhook", payH.MidtransWebhook)

	// Inventory
	invRoutes := api.Group("/inventory", jwtGuard)
	invRoutes.Get("/ingredients", invH.ListIngredients)
	invRoutes.Post("/ingredients", invH.CreateIngredient)
	invRoutes.Put("/ingredients/:id", invH.UpdateIngredient)
	invRoutes.Post("/ingredients/:id/adjust", invH.AdjustStock)
	invRoutes.Get("/recipes/:product_id", invH.GetRecipe)
	invRoutes.Post("/recipes/:product_id", invH.SetRecipe)
	invRoutes.Get("/logs", invH.ListLogs)

	// Shift
	shiftRoutes := api.Group("/shift", jwtGuard)
	shiftRoutes.Post("/open", shiftH.OpenShift)
	shiftRoutes.Get("/current", shiftH.GetCurrentShift)
	shiftRoutes.Post("/cash-movement", shiftH.AddCashMovement)
	shiftRoutes.Post("/close", shiftH.CloseShift)

	// Report
	reportRoutes := api.Group("/report", jwtGuard)
	reportRoutes.Get("/daily-sales", reportH.DailySales)
	reportRoutes.Get("/gross-profit", reportH.GrossProfit)
	reportRoutes.Get("/top-selling", reportH.TopSelling)

	return app, dbFile
}

func TestCompleteSPOSFlow(t *testing.T) {
	app, dbFile := setupTestApp(t)
	defer os.Remove(dbFile)

	// 1. Verify SetupGuard blocks uninitialized access
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pos/tables", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden before setup, got %d", resp.StatusCode)
	}

	// 2. Initialize Setup
	initPayload := map[string]interface{}{
		"outlet": map[string]interface{}{
			"name":               "Testing Resto",
			"tagline":            "Modern Taste",
			"address":            "Jl. Testing No 1",
			"phone":              "0812345678",
			"business_type":      "resto",
			"service_modes":      []string{"dine_in", "take_away"},
			"tax_percentage":     10.0,
			"service_percentage": 5.0,
			"is_tax_inclusive":   false,
		},
		"user": map[string]interface{}{
			"name":     "Super Admin",
			"username": "admin",
			"email":    "admin@test.com",
			"password": "password123",
			"auth_pin": "123456",
		},
	}
	body, _ := json.Marshal(initPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/setup/initialize", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK on initialize, got %d", resp.StatusCode)
	}

	// 3. Verify initialize endpoint is locked after setup
	req = httptest.NewRequest(http.MethodPost, "/api/v1/setup/initialize", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden on re-initialization, got %d", resp.StatusCode)
	}

	// 4. Authenticate Login
	loginPayload := map[string]string{
		"username": "admin",
		"password": "password123",
	}
	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK on login, got %d", resp.StatusCode)
	}

	var loginResp struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp.Data.Token
	if token == "" {
		t.Fatal("Token returned is empty")
	}

	authHeader := "Bearer " + token

	// 5. Open Shift
	shiftPayload := map[string]interface{}{"opening_cash": 200000.0}
	body, _ = json.Marshal(shiftPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/shift/open", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201 Created on open shift, got %d", resp.StatusCode)
	}

	// 6. Create Table
	tablePayload := map[string]interface{}{"table_number": "T-01", "capacity": 4}
	body, _ = json.Marshal(tablePayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/pos/tables", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201 Created on create table, got %d", resp.StatusCode)
	}

	var tableResp struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&tableResp)
	tableID := tableResp.Data.ID

	// 7. Create Category & Ingredient & Product
	catPayload := map[string]interface{}{"name": "Coffee"}
	body, _ = json.Marshal(catPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/pos/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)

	var catResp struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&catResp)

	// Create Ingredient: Espresso Beans
	ingPayload := map[string]interface{}{
		"name":            "Espresso Beans",
		"unit":            "gram",
		"current_stock":   1000.0,
		"min_stock_alert": 100.0,
		"cost_per_unit":   200.0, // Rp 200 / gram
	}
	body, _ = json.Marshal(ingPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/inventory/ingredients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)

	var ingResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&ingResp)
	ingredientID := ingResp.Data.ID

	// Create Product: Cappuccino
	prodPayload := map[string]interface{}{
		"category_id": catResp.Data.ID,
		"name":        "Cappuccino",
		"price":       30000.0,
		"sku":         "CAP-01",
	}
	body, _ = json.Marshal(prodPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/pos/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)

	var prodResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&prodResp)
	productID := prodResp.Data.ID

	// Set Product Recipe (BOM): 18g beans per Cappuccino
	recipePayload := map[string]interface{}{
		"recipes": []map[string]interface{}{
			{"ingredient_id": ingredientID, "quantity_required": 18.0},
		},
	}
	body, _ = json.Marshal(recipePayload)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/inventory/recipes/%s", productID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK on set recipe, got %d", resp.StatusCode)
	}

	// 8. Create Order
	orderPayload := map[string]interface{}{
		"table_id":      tableID,
		"service_type":  "dine_in",
		"customer_name": "Budi",
		"items": []map[string]interface{}{
			{
				"product_id": productID,
				"quantity":   2,
				"unit_price": 30000.0,
				"item_notes": "Extra hot",
			},
		},
	}
	body, _ = json.Marshal(orderPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/pos/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201 Created on create order, got %d", resp.StatusCode)
	}

	var orderResp struct {
		Data struct {
			ID          string  `json:"id"`
			TotalAmount float64 `json:"total_amount"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&orderResp)
	orderID := orderResp.Data.ID

	// 9. Process Cash Payment Checkout
	cashPayload := map[string]interface{}{
		"order_id":    orderID,
		"amount_paid": 100000.0,
	}
	body, _ = json.Marshal(cashPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/payment/cash", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK on cash payment, got %d", resp.StatusCode)
	}

	// 10. Verify BOM deduction (1000g - (2 cups * 18g) = 964g remaining)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/inventory/ingredients", nil)
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)

	var ingListResp struct {
		Data []struct {
			ID           string  `json:"id"`
			CurrentStock float64 `json:"current_stock"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&ingListResp)

	if len(ingListResp.Data) > 0 {
		remStock := ingListResp.Data[0].CurrentStock
		if remStock != 964.0 {
			t.Fatalf("Expected remaining stock to be 964.0 after 2 cups deduction, got %.2f", remStock)
		}
	}

	// 11. Close Shift Reconcile
	closeShiftPayload := map[string]interface{}{
		"actual_cash": 200000.0 + orderResp.Data.TotalAmount,
		"notes":       "All matched",
	}
	body, _ = json.Marshal(closeShiftPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/shift/close", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK on close shift, got %d", resp.StatusCode)
	}

	// 12. Daily Sales & Gross Profit Report Verification
	req = httptest.NewRequest(http.MethodGet, "/api/v1/report/daily-sales", nil)
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK on daily sales report, got %d", resp.StatusCode)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/report/gross-profit", nil)
	req.Header.Set("Authorization", authHeader)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK on gross profit report, got %d", resp.StatusCode)
	}
}
