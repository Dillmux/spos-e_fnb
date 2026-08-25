package middleware

import (
	"strings"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/spos/backend/internal/models"
	"github.com/spos/backend/internal/utils"
	"gorm.io/gorm"
)

var isAppInitializedCached int32

// MarkAppAsInitialized memperbarui cache memori bahwa aplikasi telah terinisialisasi.
func MarkAppAsInitialized() {
	atomic.StoreInt32(&isAppInitializedCached, 1)
}

// InitSetupGuardCache memeriksa DB saat bootstrapping aplikasi untuk mem-prewarm atomic cache.
func InitSetupGuardCache(db *gorm.DB) bool {
	var config models.SystemConfig
	if err := db.Where("key = ?", "app_initialized").First(&config).Error; err == nil && config.Value == "true" {
		atomic.StoreInt32(&isAppInitializedCached, 1)
		return true
	}
	return false
}

// SetupGuard memeriksa apakah aplikasi telah diinisialisasi.
// Memanfaatkan in-memory caching untuk mengeliminasi kueri SQL berulang pada setiap request API.
func SetupGuard(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Lewati guard untuk aset statis / rute frontend (non-API)
		if !strings.HasPrefix(path, "/api/") {
			return c.Next()
		}

		isInitialized := atomic.LoadInt32(&isAppInitializedCached) == 1

		if !isInitialized {
			var config models.SystemConfig
			err := db.Where("key = ?", "app_initialized").First(&config).Error
			if err == nil && config.Value == "true" {
				isInitialized = true
				atomic.StoreInt32(&isAppInitializedCached, 1)
			}
		}

		if !isInitialized {
			// Aplikasi BELUM diinisialisasi: Hanya izinkan rute /api/v1/setup/*
			if strings.HasPrefix(path, "/api/v1/setup/") {
				return c.Next()
			}
			return utils.SendError(c, fiber.StatusForbidden, "Inisialisasi sistem diperlukan. Silakan selesaikan setup via /api/v1/setup/initialize", nil)
		}

		// Aplikasi SUDAH diinisialisasi: Kunci endpoint inisialisasi dan tes koneksi setup secara permanen
		if path == "/api/v1/setup/initialize" || strings.HasPrefix(path, "/api/v1/setup/test-") {
			return utils.SendError(c, fiber.StatusForbidden, "Aplikasi telah terinisialisasi dan modul setup telah dikunci demi keamanan.", nil)
		}

		return c.Next()
	}
}
