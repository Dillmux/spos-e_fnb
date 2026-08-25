package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/spos/backend/internal/config"
	"github.com/spos/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB menginisialisasi koneksi database GORM berdasarkan konfigurasi driver (SQLite/PostgreSQL) dan menjalankan auto-migrasi skema.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch strings.ToLower(cfg.DBDriver) {
	case "postgres", "postgresql":
		dialector = postgres.Open(cfg.DBDsn)
	case "sqlite", "sqlite3":
		dialector = sqlite.Open(cfg.DBDsn)
	default:
		dialector = sqlite.Open(cfg.DBDsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungkan driver database (%s): %w", cfg.DBDriver, err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(1 * time.Hour)
	}

	log.Printf("[DATABASE] Koneksi berhasil dibuat menggunakan driver: %s (Pool: MaxOpen=100, MaxIdle=10)", cfg.DBDriver)

	// Jalankan auto-migrasi skema tabel untuk seluruh entitas sistem
	if err := MigrateSchema(db); err != nil {
		return nil, fmt.Errorf("gagal menjalankan auto-migrasi: %w", err)
	}

	return db, nil
}

// MigrateSchema mengeksekusi AutoMigrate GORM dan pembuatan indeks untuk seluruh model sistem.
func MigrateSchema(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.SystemConfig{},
		&models.Outlet{},
		&models.User{},
		&models.Shift{},
		&models.CashMovement{},
		&models.Category{},
		&models.Product{},
		&models.Ingredient{},
		&models.ProductRecipe{},
		&models.InventoryLog{},
		&models.Table{},
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
	)
	if err != nil {
		return err
	}

	createIndexes(db)
	log.Println("[DATABASE] Auto-migrasi skema tabel berhasil diselesaikan.")
	return nil
}

// createIndexes memastikan indeks majemuk (composite index) dibuat secara eksplisit untuk performa kueri transaksi.
func createIndexes(db *gorm.DB) {
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_orders_status_created ON orders (status, created_at);").Error
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_order_items_kitchen ON order_items (kitchen_status, created_at);").Error
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_inventory_logs_ingredient ON inventory_logs (ingredient_id, created_at);").Error
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_shifts_user_status ON shifts (user_id, status);").Error
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_system_configs_key ON system_configs (key);").Error
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_products_category ON products (category_id);").Error
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_orders_outlet_queue ON orders (outlet_id, created_at, queue_number);").Error
}

