package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
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

//go:embed public/*
var embeddedPublic embed.FS

func setupStaticSPA(app *fiber.App) {
	fsys, err := fs.Sub(embeddedPublic, "public")
	if err != nil {
		log.Printf("[EMBED] Warning: Failed to load embedded filesystem sub-path: %v", err)
		return
	}
	app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(fsys),
		Index:        "index.html",
		NotFoundFile: "index.html",
	}))
}

func main() {
	// 1. Memuat Konfigurasi Aplikasi
	cfg := config.LoadConfig()
	log.Printf("[STARTUP] Memulai SPOS Server pada port :%s", cfg.Port)

	// 2. Inisialisasi Koneksi Database Multi-Driver & Jalankan Migrasi Otomatis
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("[FATAL] Inisialisasi database gagal: %v", err)
	}

	// Pre-warm setup guard cache untuk mengeliminasi slow SQL query saat request awal
	if middleware.InitSetupGuardCache(db) {
		log.Println("[STARTUP] Setup guard cache terisi: Aplikasi telah terinisialisasi.")
	} else {
		log.Println("[STARTUP] Setup guard cache terisi: Aplikasi belum terinisialisasi.")
	}

	// 3. Inisialisasi Aplikasi Fiber
	app := fiber.New(fiber.Config{
		AppName:       "SPOS F&B Engine v1.0",
		CaseSensitive: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return utils.SendError(c, code, err.Error(), nil)
		},
	})

	// 4. Registrasi Middleware Standar Global
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CorsAllowedOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// 5. Registrasi Middleware Guard Inisialisasi Aplikasi
	app.Use(middleware.SetupGuard(db))

	// 6. Instansiasi Handler Domain
	setupH := setup.NewSetupHandler(db, cfg.JWTSecret)
	authH := auth.NewAuthHandler(db, cfg.JWTSecret)
	posH := pos.NewPOSHandler(db)
	payH := payment.NewPaymentHandler(db, cfg)
	invH := inventory.NewInventoryHandler(db)
	shiftH := shift.NewShiftHandler(db)
	reportH := report.NewReportHandler(db)

	// 7. Registrasi Rute API
	api := app.Group("/api/v1")

	// --- MODUL SETUP ---
	setupRoutes := api.Group("/setup")
	setupRoutes.Get("/status", setupH.GetStatus)
	setupRoutes.Post("/test-db", setupH.TestDB)
	setupRoutes.Post("/test-midtrans", setupH.TestMidtrans)
	setupRoutes.Post("/initialize", setupH.Initialize)

	// --- MODUL AUTENTIKASI ---
	authLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return utils.SendError(c, fiber.StatusTooManyRequests, "Terlalu banyak percobaan otentikasi. Silakan coba lagi dalam 1 menit.", nil)
		},
	})

	authRoutes := api.Group("/auth")
	authRoutes.Post("/login", authLimiter, authH.Login)
	jwtGuard := middleware.JWTMiddleware(cfg.JWTSecret)
	authRoutes.Post("/refresh", jwtGuard, authH.RefreshToken)
	authRoutes.Post("/verify-pin", jwtGuard, authLimiter, authH.VerifyPIN)
	authRoutes.Get("/me", jwtGuard, authH.Me)

	// Manajemen Pengguna
	adminGuard := middleware.RoleGuard("superadmin", "manager")
	authRoutes.Get("/users", jwtGuard, adminGuard, authH.ListUsers)
	authRoutes.Post("/users", jwtGuard, adminGuard, authH.CreateUser)
	authRoutes.Put("/users/:id", jwtGuard, adminGuard, authH.UpdateUser)
	authRoutes.Delete("/users/:id", jwtGuard, adminGuard, authH.DeleteUser)

	// --- MODUL KASIR POS ---
	posRoutes := api.Group("/pos", jwtGuard)

	// Manajemen Meja
	posRoutes.Get("/tables", posH.ListTables)
	posRoutes.Post("/tables", middleware.RoleGuard("superadmin", "manager"), posH.CreateTable)
	posRoutes.Put("/tables/:id", middleware.RoleGuard("superadmin", "manager"), posH.UpdateTable)
	posRoutes.Delete("/tables/:id", middleware.RoleGuard("superadmin", "manager"), posH.DeleteTable)

	// Katalog Menu & Kategori
	posRoutes.Get("/categories", posH.ListCategories)
	posRoutes.Post("/categories", middleware.RoleGuard("superadmin", "manager"), posH.CreateCategory)
	posRoutes.Put("/categories/:id", middleware.RoleGuard("superadmin", "manager"), posH.UpdateCategory)
	posRoutes.Delete("/categories/:id", middleware.RoleGuard("superadmin", "manager"), posH.DeleteCategory)
	posRoutes.Get("/products", posH.ListProducts)
	posRoutes.Post("/products", middleware.RoleGuard("superadmin", "manager"), posH.CreateProduct)
	posRoutes.Put("/products/:id", middleware.RoleGuard("superadmin", "manager"), posH.UpdateProduct)
	posRoutes.Delete("/products/:id", middleware.RoleGuard("superadmin", "manager"), posH.DeleteProduct)

	// Pesanan & Operasi Tagihan (Pisah/Gabung/Batal)
	posRoutes.Post("/orders", posH.CreateOrder)
	posRoutes.Get("/orders", posH.ListOrders)
	posRoutes.Get("/orders/:id", posH.GetOrder)
	posRoutes.Post("/orders/:id/split", posH.SplitBill)
	posRoutes.Post("/orders/merge", posH.MergeBill)
	posRoutes.Post("/orders/:id/void", posH.VoidOrder)
	posRoutes.Put("/orders/:id/complete", posH.CompleteOrderItems)
	posRoutes.Put("/orders/items/:item_id/status", posH.UpdateKitchenStatus)

	// --- MODUL PEMBAYARAN ---
	webhookLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return utils.SendError(c, fiber.StatusTooManyRequests, "Terlalu banyak permintaan webhook", nil)
		},
	})

	payRoutes := api.Group("/payment")
	payRoutes.Post("/cash", jwtGuard, payH.CashCheckout)
	payRoutes.Post("/midtrans/charge", jwtGuard, payH.MidtransCharge)
	payRoutes.Post("/midtrans/webhook", webhookLimiter, payH.MidtransWebhook) // Validasi webhook ditangani via SHA-512 & Rate Limiter
	payRoutes.Post("/midtrans/notification", webhookLimiter, payH.MidtransWebhook)
	payRoutes.Get("/status/:order_id", jwtGuard, payH.GetPaymentStatus)

	// --- MODUL INVENTARIS & STOK ---
	invRoutes := api.Group("/inventory", jwtGuard, middleware.RoleGuard("superadmin", "manager"))
	invRoutes.Get("/ingredients", invH.ListIngredients)
	invRoutes.Post("/ingredients", invH.CreateIngredient)
	invRoutes.Put("/ingredients/:id", invH.UpdateIngredient)
	invRoutes.Delete("/ingredients/:id", invH.DeleteIngredient)
	invRoutes.Post("/ingredients/:id/adjust", invH.AdjustStock)
	invRoutes.Get("/recipes/:product_id", invH.GetRecipe)
	invRoutes.Post("/recipes/:product_id", invH.SetRecipe)
	invRoutes.Get("/logs", invH.ListLogs)

	// --- MODUL SHIFT KASIR ---
	for _, path := range []string{"/shift", "/shifts"} {
		sr := api.Group(path, jwtGuard)
		sr.Post("/open", shiftH.OpenShift)
		sr.Get("/current", shiftH.GetCurrentShift)
		sr.Get("/active", shiftH.GetCurrentShift)
		sr.Post("/cash-movement", shiftH.AddCashMovement)
		sr.Post("/movements", shiftH.AddCashMovement)
		sr.Post("/close", shiftH.CloseShift)
	}

	// --- MODUL LAPORAN ---
	reportRoutes := api.Group("/report", jwtGuard, middleware.RoleGuard("superadmin", "manager"))
	reportRoutes.Get("/daily-sales", reportH.DailySales)
	reportRoutes.Get("/gross-profit", reportH.GrossProfit)
	reportRoutes.Get("/top-selling", reportH.TopSelling)
	reportRoutes.Get("/export/daily-sales", reportH.ExportDailySalesCSV)
	reportRoutes.Get("/export/gross-profit", reportH.ExportGrossProfitCSV)

	// --- MODUL REAL-TIME EVENTS (SSE) ---
	api.Get("/events/stream", utils.GetSSEBroker().StreamHandler)

	// 8. Menyajikan Frontend SPA Terembed
	setupStaticSPA(app)

	// 9. Pendengar Shutdown Berkelanjutan (Graceful Shutdown)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", cfg.Port)); err != nil {
			log.Printf("[SERVER] Berhenti: %v", err)
		}
	}()

	log.Printf("[SERVER] Engine API SPOS berhasil dijalankan pada port :%s", cfg.Port)

	<-sigChan
	log.Println("[SERVER] Mematikan server secara aman...")
	_ = app.Shutdown()
	log.Println("[SERVER] Server berhasil dimatikan secara bersih.")
}
