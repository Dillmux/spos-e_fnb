package config

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config menyimpan seluruh parameter konfigurasi aplikasi runtime SPOS-E.
type Config struct {
	Port                 string
	DBDriver             string
	DBDsn                string
	JWTSecret            string
	CorsAllowedOrigins   string
	MidtransServerKey    string
	MidtransClientKey    string
	MidtransIsProduction bool
}

func generateRandomSecret(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("spos-jwt-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// LoadConfig membaca berkas .env dan variabel lingkungan (environment variables).
func LoadConfig() *Config {
	// Memuat secara otomatis berkas .env jika tersedia di direktori kerja
	loadDotEnv(".env")

	port := getEnv("PORT", "8080")
	dbDriver := getEnv("DB_DRIVER", "sqlite")
	dbDsn := getEnv("DB_DSN", "spos.db")
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "*")
	jwtSecret := getEnv("JWT_SECRET", "")
	appEnv := getEnv("APP_ENV", "development")

	if jwtSecret == "" {
		jwtSecret = generateRandomSecret(32)
		log.Println("[SECURITY NOTICE] JWT_SECRET tidak terkonfigurasi. Kunci kriptografi aman 64-karakter berhasil di-generate secara otomatis.")
	} else if appEnv == "production" && (jwtSecret == "spos-super-secret-jwt-key-2026" || len(jwtSecret) < 16) {
		log.Println("[SECURITY WARNING] Peringatan Keamanan Produk: JWT_SECRET yang digunakan di mode produksi terlalu pendek atau menggunakan nilai default!")
	}

	midtransServerKey := getEnv("MIDTRANS_SERVER_KEY", "SB-Mid-server-placeholder")
	midtransClientKey := getEnv("MIDTRANS_CLIENT_KEY", "SB-Mid-client-placeholder")
	midtransIsProdStr := getEnv("MIDTRANS_IS_PRODUCTION", "true")
	midtransIsProd, _ := strconv.ParseBool(midtransIsProdStr)

	return &Config{
		Port:                 port,
		DBDriver:             dbDriver,
		DBDsn:                dbDsn,
		JWTSecret:            jwtSecret,
		CorsAllowedOrigins:   corsOrigins,
		MidtransServerKey:    midtransServerKey,
		MidtransClientKey:    midtransClientKey,
		MidtransIsProduction: midtransIsProd,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

// loadDotEnv membaca pasangan key-value dari berkas .env dan mengisinya ke environment variables jika belum tersedia.
func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return // Berkas tidak ditemukan atau tidak dapat dibaca, lewati secara aman
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Hapus tanda kutip jika ada
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return
	}
}
