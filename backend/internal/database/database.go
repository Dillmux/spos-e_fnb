package database

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/spos/backend/internal/config"
	"github.com/spos/backend/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ParseMySQLDSN menormalisasi string koneksi MySQL/MariaDB dari format URL (mysql://user:pass@host:port/db) maupun format DSN go-sql-driver (user:pass@tcp(host:port)/db).
func ParseMySQLDSN(dsn string) string {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return dsn
	}

	if strings.HasPrefix(dsn, "mysql://") || strings.HasPrefix(dsn, "mariadb://") {
		u, err := url.Parse(dsn)
		if err == nil {
			user := u.User.Username()
			pass, _ := u.User.Password()
			host := u.Host
			dbName := strings.TrimPrefix(u.Path, "/")

			queryParams := u.Query()
			tlsVal := queryParams.Get("tls")
			if tlsVal == "" {
				sslMode := strings.ToUpper(queryParams.Get("ssl-mode"))
				if sslMode == "REQUIRED" || sslMode == "TRUE" {
					tlsVal = "true"
				} else if sslMode == "DISABLED" || sslMode == "FALSE" {
					tlsVal = "false"
				} else if sslMode == "SKIP-VERIFY" {
					tlsVal = "skip-verify"
				} else {
					tlsVal = "true"
				}
			}

			if pass != "" {
				dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s", user, pass, host, dbName, tlsVal)
			} else if user != "" {
				dsn = fmt.Sprintf("%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s", user, host, dbName, tlsVal)
			}
			return dsn
		}
	}

	// Format DSN standar Go: Tambahkan parseTime=True jika belum ada
	if !strings.Contains(dsn, "parseTime=") {
		if strings.Contains(dsn, "?") {
			dsn += "&parseTime=True"
		} else {
			dsn += "?parseTime=True"
		}
	}
	return dsn
}

// InitDB menginisialisasi koneksi database GORM berdasarkan konfigurasi driver (SQLite/PostgreSQL/MySQL) dan menjalankan auto-migrasi skema.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch strings.ToLower(cfg.DBDriver) {
	case "postgres", "postgresql":
		dialector = postgres.Open(cfg.DBDsn)
	case "mysql", "mariadb":
		dialector = mysql.Open(ParseMySQLDSN(cfg.DBDsn))
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

	driverName := strings.ToLower(cfg.DBDriver)
	sqlDB, err := db.DB()
	if err == nil {
		if driverName == "sqlite" || driverName == "sqlite3" || driverName == "" {
			// Aktifkan WAL mode & PRAGMA optimasi konkurensi SQLite
			_ = db.Exec("PRAGMA journal_mode = WAL;").Error
			_ = db.Exec("PRAGMA busy_timeout = 5000;").Error
			_ = db.Exec("PRAGMA synchronous = NORMAL;").Error
			_ = db.Exec("PRAGMA cache_size = -2000;").Error

			sqlDB.SetMaxOpenConns(5)
			sqlDB.SetMaxIdleConns(2)
			log.Printf("[DATABASE] SQLite dikonfigurasi dengan WAL mode (Pool: MaxOpen=5, MaxIdle=2)")
		} else {
			sqlDB.SetMaxOpenConns(100)
			sqlDB.SetMaxIdleConns(10)
			log.Printf("[DATABASE] PostgreSQL dikonfigurasi (Pool: MaxOpen=100, MaxIdle=10)")
		}
		sqlDB.SetConnMaxLifetime(1 * time.Hour)
	}

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

