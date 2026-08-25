package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spos/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	_ = db.AutoMigrate(&models.SystemConfig{}, &models.User{}, &models.Outlet{})
	return db
}

func TestGenerateAndValidateJWTToken(t *testing.T) {
	secret := "test-secret-key-1234567890"
	user := &models.User{
		ID:       "user-123",
		OutletID: "outlet-456",
		Username: "testcashier",
		Role:     "cashier",
	}

	tokenStr, err := GenerateJWTToken(user, secret)
	if err != nil {
		t.Fatalf("GenerateJWTToken error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("Expected non-empty token string")
	}

	app := fiber.New()
	app.Use(JWTMiddleware(secret))
	app.Get("/test", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		role := c.Locals("role")
		return c.JSON(fiber.Map{"user_id": userID, "role": role})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("App test error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}
}

func TestJWTMiddlewareMissingHeader(t *testing.T) {
	app := fiber.New()
	app.Use(JWTMiddleware("secret"))
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for missing header, got %d", resp.StatusCode)
	}
}

func TestRoleGuard(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("role", "cashier")
		return c.Next()
	})
	app.Get("/admin", RoleGuard("superadmin", "manager"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/pos", RoleGuard("cashier", "manager"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Test restricted route for cashier
	req1 := httptest.NewRequest("GET", "/admin", nil)
	resp1, _ := app.Test(req1)
	if resp1.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for cashier accessing admin route, got %d", resp1.StatusCode)
	}

	// Test allowed route for cashier
	req2 := httptest.NewRequest("GET", "/pos", nil)
	resp2, _ := app.Test(req2)
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for cashier accessing pos route, got %d", resp2.StatusCode)
	}
}

func TestVerifySupervisorPIN(t *testing.T) {
	db := setupTestDB(t)

	pin := "123456"
	pinHash, _ := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)

	supervisor := models.User{
		ID:          "sup-1",
		OutletID:    "out-1",
		Name:        "Test Manager",
		Username:    "manager1",
		PasswordHash: "pass",
		AuthPINHash:  string(pinHash),
		Role:        "manager",
		IsActive:    true,
	}
	db.Create(&supervisor)

	// Valid PIN check
	valid, user, err := VerifySupervisorPIN(db, "out-1", "123456")
	if err != nil || !valid || user == nil {
		t.Errorf("Expected valid PIN verification, got valid=%v, err=%v", valid, err)
	}

	// Invalid PIN check
	invalid, _, _ := VerifySupervisorPIN(db, "out-1", "999999")
	if invalid {
		t.Error("Expected invalid result for incorrect PIN")
	}
}

func TestSetupGuardAndPreWarmCache(t *testing.T) {
	db := setupTestDB(t)

	// Test pre-warm before initialization
	if InitSetupGuardCache(db) {
		t.Error("Expected InitSetupGuardCache to return false when app_initialized is missing")
	}

	app := fiber.New()
	app.Use(SetupGuard(db))
	app.Get("/api/v1/pos/products", func(c *fiber.Ctx) error { return c.SendString("ok") })
	app.Get("/api/v1/setup/status", func(c *fiber.Ctx) error { return c.SendString("setup ok") })

	// Uninitialized state: /api/v1/pos/products should return 403
	req1 := httptest.NewRequest("GET", "/api/v1/pos/products", nil)
	resp1, _ := app.Test(req1)
	if resp1.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 when uninitialized, got %d", resp1.StatusCode)
	}

	// Uninitialized state: /api/v1/setup/status should return 200
	req2 := httptest.NewRequest("GET", "/api/v1/setup/status", nil)
	resp2, _ := app.Test(req2)
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 for setup status route, got %d", resp2.StatusCode)
	}

	// Initialize app
	db.Create(&models.SystemConfig{Key: "app_initialized", Value: "true", CreatedAt: time.Now(), UpdatedAt: time.Now()})
	MarkAppAsInitialized()

	// Initialized state: /api/v1/pos/products should return 200
	req3 := httptest.NewRequest("GET", "/api/v1/pos/products", nil)
	resp3, _ := app.Test(req3)
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 after initialization, got %d", resp3.StatusCode)
	}
}

// Ensure unused import warnings are resolved if jwt is imported
var _ = jwt.RegisteredClaims{}
