package setup

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/spos/backend/internal/database"
	"github.com/spos/backend/internal/middleware"
	"github.com/spos/backend/internal/models"
	"github.com/spos/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SetupHandler struct {
	db        *gorm.DB
	jwtSecret string
}

func NewSetupHandler(db *gorm.DB, jwtSecret string) *SetupHandler {
	return &SetupHandler{db: db, jwtSecret: jwtSecret}
}

type TestDBRequest struct {
	Driver           string `json:"driver"`
	DSN              string `json:"dsn"`
	ConnectionString string `json:"connection_string"`
}

type TestMidtransRequest struct {
	ServerKey string `json:"server_key"`
}

type ActiveModulesStruct struct {
	EnableTables   bool `json:"enable_tables"`
	EnableKDS      bool `json:"enable_kds"`
	EnableWLD      bool `json:"enable_wld"`
	EnableBOM      bool `json:"enable_bom"`
	EnableMidtrans bool `json:"enable_midtrans"`
}

type InitializeRequest struct {
	Outlet struct {
		Name              string   `json:"name"`
		Tagline           string   `json:"tagline"`
		Address           string   `json:"address"`
		Phone             string   `json:"phone"`
		LogoURL           string   `json:"logo_url"`
		BusinessType      string   `json:"business_type"`
		ServiceModes      []string `json:"service_modes"`
		TaxPercentage     float64  `json:"tax_percentage"`
		ServicePercentage float64  `json:"service_percentage"`
		IsTaxInclusive    bool     `json:"is_tax_inclusive"`
		ReceiptFooter     string   `json:"receipt_footer"`
	} `json:"outlet"`
	User struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		AuthPIN  string `json:"auth_pin"`
	} `json:"user"`

	ActiveModules ActiveModulesStruct `json:"active_modules"`

	MidtransServerKey    string `json:"midtrans_server_key"`
	MidtransClientKey    string `json:"midtrans_client_key"`
	MidtransIsProduction bool   `json:"midtrans_is_production"`
	DBDriver             string `json:"db_driver"`
	DBConnectionString string `json:"db_connection_string"`

	// Fallback flat JSON payload
	OutletName        string   `json:"outlet_name"`
	OutletTagline     string   `json:"outlet_tagline"`
	OutletAddress     string   `json:"outlet_address"`
	OutletPhone       string   `json:"outlet_phone"`
	BusinessType      string   `json:"business_type"`
	ServiceModes      []string `json:"service_modes"`
	TaxPercentage     float64  `json:"tax_percentage"`
	ServicePercentage float64  `json:"service_percentage"`
	IsTaxInclusive    bool     `json:"is_tax_inclusive"`

	AdminName     string `json:"admin_name"`
	AdminUsername string `json:"admin_username"`
	AdminEmail    string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`
	AdminPIN      string `json:"admin_pin"`
}

// saveSysConfig menyimpan atau memperbarui entri pasangan kunci-nilai konfigurasi sistem.
func saveSysConfig(tx *gorm.DB, key string, val string) error {
	now := time.Now()
	var sysConf models.SystemConfig
	err := tx.Where("key = ?", key).First(&sysConf).Error
	if err != nil {
		sysConf = models.SystemConfig{
			Key:       key,
			Value:     val,
			CreatedAt: now,
			UpdatedAt: now,
		}
		return tx.Create(&sysConf).Error
	}
	sysConf.Value = val
	sysConf.UpdatedAt = now
	return tx.Save(&sysConf).Error
}

// GetStatus memeriksa apakah sistem telah diinisialisasi dan memuat konfigurasi modul aktif.
func (h *SetupHandler) GetStatus(c *fiber.Ctx) error {
	var sysConf models.SystemConfig
	err := h.db.Where("key = ?", "app_initialized").First(&sysConf).Error
	isInitialized := err == nil && sysConf.Value == "true"

	var outletName string
	var outlet models.Outlet
	if isInitialized {
		if err := h.db.First(&outlet).Error; err == nil {
			outletName = outlet.Name
		}
	}

	var sysConfigs []models.SystemConfig
	h.db.Find(&sysConfigs)
	featureFlags := map[string]bool{
		"enable_tables":   true,
		"enable_kds":      true,
		"enable_wld":      true,
		"enable_bom":      true,
		"enable_midtrans": true,
	}
	hasMidtransKey := false
	if serverKey := os.Getenv("MIDTRANS_SERVER_KEY"); serverKey != "" && !strings.Contains(serverKey, "placeholder") {
		hasMidtransKey = true
	}

	for _, cfg := range sysConfigs {
		if strings.HasPrefix(cfg.Key, "enable_") {
			featureFlags[cfg.Key] = (cfg.Value == "true")
		}
		if cfg.Key == "midtrans_server_key" && cfg.Value != "" && !strings.Contains(cfg.Value, "placeholder") {
			hasMidtransKey = true
		}
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Status setup berhasil diambil", fiber.Map{
		"initialized":         isInitialized,
		"db_connected":        true,
		"outlet_name":         outletName,
		"outlet":              outlet,
		"feature_flags":       featureFlags,
		"midtrans_configured": hasMidtransKey,
	})
}

// TestDB memverifikasi koneksi string driver database.
func (h *SetupHandler) TestDB(c *fiber.Ctx) error {
	var req TestDBRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Payload JSON tidak valid", err.Error())
	}

	if req.Driver == "" {
		req.Driver = "sqlite"
	}
	if req.DSN == "" && req.ConnectionString != "" {
		req.DSN = req.ConnectionString
	}
	if req.DSN == "" {
		req.DSN = "spos_test.db"
	}

	var dialector gorm.Dialector
	switch strings.ToLower(req.Driver) {
	case "postgres", "postgresql":
		dialector = postgres.Open(req.DSN)
	default:
		dialector = sqlite.Open(req.DSN)
	}

	testConn, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Uji koneksi database gagal", err.Error())
	}

	sqlDB, err := testConn.DB()
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Gagal membaca koneksi DB", err.Error())
	}

	if pingErr := sqlDB.Ping(); pingErr != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Ping database gagal", pingErr.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Uji koneksi database berhasil", fiber.Map{
		"driver":    req.Driver,
		"status":    "connected",
		"connected": true,
	})
}

// TestMidtrans memverifikasi kelayakan format Midtrans Server Key.
func (h *SetupHandler) TestMidtrans(c *fiber.Ctx) error {
	var req TestMidtransRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Payload JSON tidak valid", err.Error())
	}

	if strings.TrimSpace(req.ServerKey) == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Midtrans server_key tidak boleh kosong", nil)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Format konfigurasi Midtrans valid", fiber.Map{
		"server_key_length": len(req.ServerKey),
		"status":            "valid",
		"valid":             true,
	})
}

// updateDotEnv memperbarui atau menambahkan kunci-kunci variabel pada berkas .env
func updateDotEnv(updates map[string]string) error {
	envPath := ".env"
	var lines []string

	if file, err := os.Open(envPath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("gagal membaca berkas .env: %w", err)
		}
	}

	foundKeys := make(map[string]bool)
	var newLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		replaced := false
		for key, val := range updates {
			if strings.HasPrefix(trimmed, key+"=") {
				newLines = append(newLines, fmt.Sprintf("%s=%s", key, val))
				foundKeys[key] = true
				replaced = true
				break
			}
		}
		if !replaced {
			newLines = append(newLines, line)
		}
	}

	for key, val := range updates {
		if !foundKeys[key] && val != "" {
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, val))
		}
	}

	content := strings.Join(newLines, "\n") + "\n"
	return os.WriteFile(envPath, []byte(content), 0644)
}

// Initialize melakukan inisialisasi awal outlet, pengguna superadmin, dan mengunci setup guard.
func (h *SetupHandler) Initialize(c *fiber.Ctx) error {
	var req InitializeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Body permintaan JSON tidak valid", err.Error())
	}

	if req.Outlet.Name == "" && req.OutletName != "" {
		req.Outlet.Name = req.OutletName
		req.Outlet.Tagline = req.OutletTagline
		req.Outlet.Address = req.OutletAddress
		req.Outlet.Phone = req.OutletPhone
		req.Outlet.BusinessType = req.BusinessType
		req.Outlet.ServiceModes = req.ServiceModes
		req.Outlet.TaxPercentage = req.TaxPercentage
		req.Outlet.ServicePercentage = req.ServicePercentage
		req.Outlet.IsTaxInclusive = req.IsTaxInclusive
	}
	if req.User.Username == "" && req.AdminUsername != "" {
		req.User.Name = req.AdminName
		req.User.Username = req.AdminUsername
		req.User.Email = req.AdminEmail
		req.User.Password = req.AdminPassword
		req.User.AuthPIN = req.AdminPIN
	}
	if req.User.Email == "" && req.User.Username != "" {
		req.User.Email = req.User.Username + "@spos.local"
	}

	if req.Outlet.Name == "" || req.User.Username == "" || req.User.Password == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Field wajib tidak lengkap: nama outlet, username pengguna, dan password", nil)
	}

	// 0. Jika driver database target ditentukan (misal PostgreSQL/Supabase), alihkan koneksi aktif & update .env
	targetDB := h.db
	targetDriver := strings.ToLower(req.DBDriver)
	targetDSN := req.DBConnectionString

	if targetDriver != "" && targetDSN != "" {
		var dialector gorm.Dialector
		switch targetDriver {
		case "postgres", "postgresql":
			dialector = postgres.Open(targetDSN)
		default:
			dialector = sqlite.Open(targetDSN)
		}

		newConn, err := gorm.Open(dialector, &gorm.Config{})
		if err != nil {
			return utils.SendError(c, fiber.StatusBadRequest, "Gagal terhubung ke database target ("+targetDriver+"): "+err.Error(), nil)
		}

		sqlDB, err := newConn.DB()
		if err != nil || sqlDB.Ping() != nil {
			pingMsg := "Tidak dapat menghubungi server database"
			if err != nil {
				pingMsg = err.Error()
			}
			return utils.SendError(c, fiber.StatusBadRequest, "Ping database target ("+targetDriver+") gagal", pingMsg)
		}

		// Jalankan migrasi skema tabel secara otomatis ke database target (misal Supabase PostgreSQL)
		if err := database.MigrateSchema(newConn); err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "Gagal auto-migrasi pada database target: "+err.Error(), nil)
		}

		// Perbarui file .env dan variabel lingkungan aplikasi secara permanen
		envUpdates := map[string]string{
			"DB_DRIVER": targetDriver,
			"DB_DSN":    targetDSN,
		}
		if req.MidtransServerKey != "" {
			envUpdates["MIDTRANS_SERVER_KEY"] = req.MidtransServerKey
			os.Setenv("MIDTRANS_SERVER_KEY", req.MidtransServerKey)
		}
		if req.MidtransClientKey != "" {
			envUpdates["MIDTRANS_CLIENT_KEY"] = req.MidtransClientKey
			os.Setenv("MIDTRANS_CLIENT_KEY", req.MidtransClientKey)
		}

		_ = updateDotEnv(envUpdates)
		os.Setenv("DB_DRIVER", targetDriver)
		os.Setenv("DB_DSN", targetDSN)

		// Alihkan pointer h.db internal aplikasi secara langsung ke koneksi database target
		*h.db = *newConn
		targetDB = h.db
	}

	tx := targetDB.Begin()
	if tx.Error != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memulai transaksi database", tx.Error.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Buat Outlet awal
	serviceModesStr := `["dine_in","take_away"]`
	if len(req.Outlet.ServiceModes) > 0 {
		serviceModesStr = `["` + strings.Join(req.Outlet.ServiceModes, `","`) + `"]`
	}

	outlet := models.Outlet{
		Name:              req.Outlet.Name,
		Tagline:           req.Outlet.Tagline,
		Address:           req.Outlet.Address,
		Phone:             req.Outlet.Phone,
		LogoURL:           req.Outlet.LogoURL,
		BusinessType:      req.Outlet.BusinessType,
		ServiceModes:      serviceModesStr,
		TaxPercentage:     req.Outlet.TaxPercentage,
		ServicePercentage: req.Outlet.ServicePercentage,
		IsTaxInclusive:    req.Outlet.IsTaxInclusive,
		ReceiptFooter:     req.Outlet.ReceiptFooter,
	}

	if err := tx.Create(&outlet).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membuat outlet awal", err.Error())
	}

	// 2. Hash Password dan PIN untuk pengguna Superadmin
	passHash, err := bcrypt.GenerateFromPassword([]byte(req.User.Password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal meng-hash password", err.Error())
	}

	var pinHashStr string
	if req.User.AuthPIN != "" {
		pinHash, err := bcrypt.GenerateFromPassword([]byte(req.User.AuthPIN), bcrypt.DefaultCost)
		if err == nil {
			pinHashStr = string(pinHash)
		}
	}

	user := models.User{
		OutletID:     outlet.ID,
		Name:         req.User.Name,
		Username:     req.User.Username,
		Email:        req.User.Email,
		PasswordHash: string(passHash),
		AuthPINHash:  pinHashStr,
		Role:         "superadmin",
		IsActive:     true,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membuat pengguna superadmin", err.Error())
	}

	// 3. Simpan status inisialisasi dan konfigurasi modul aktif
	if err := saveSysConfig(tx, "app_initialized", "true"); err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menyimpan konfigurasi app_initialized", err.Error())
	}
	saveSysConfig(tx, "enable_tables", fmt.Sprintf("%t", req.ActiveModules.EnableTables))
	saveSysConfig(tx, "enable_kds", fmt.Sprintf("%t", req.ActiveModules.EnableKDS))
	saveSysConfig(tx, "enable_wld", fmt.Sprintf("%t", req.ActiveModules.EnableWLD))
	saveSysConfig(tx, "enable_bom", fmt.Sprintf("%t", req.ActiveModules.EnableBOM))
	saveSysConfig(tx, "enable_midtrans", fmt.Sprintf("%t", req.ActiveModules.EnableMidtrans))

	if req.MidtransServerKey != "" {
		saveSysConfig(tx, "midtrans_server_key", req.MidtransServerKey)
	}
	if req.MidtransClientKey != "" {
		saveSysConfig(tx, "midtrans_client_key", req.MidtransClientKey)
	}

	isProdStr := fmt.Sprintf("%t", req.MidtransIsProduction)
	saveSysConfig(tx, "midtrans_is_production", isProdStr)

	// Tulis konfigurasi utama & MIDTRANS_IS_PRODUCTION ke .env
	envUpdatesFinal := map[string]string{
		"MIDTRANS_IS_PRODUCTION": isProdStr,
	}
	if req.MidtransServerKey != "" {
		envUpdatesFinal["MIDTRANS_SERVER_KEY"] = req.MidtransServerKey
	}
	if req.MidtransClientKey != "" {
		envUpdatesFinal["MIDTRANS_CLIENT_KEY"] = req.MidtransClientKey
	}
	_ = updateDotEnv(envUpdatesFinal)
	os.Setenv("MIDTRANS_IS_PRODUCTION", isProdStr)

	if err := tx.Commit().Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengomit transaksi database", err.Error())
	}

	middleware.MarkAppAsInitialized()

	token, _ := middleware.GenerateJWTToken(&user, h.jwtSecret)

	return utils.SendSuccess(c, fiber.StatusOK, "Sistem SPOS berhasil diinisialisasi", fiber.Map{
		"outlet_id":   outlet.ID,
		"user_id":     user.ID,
		"username":    user.Username,
		"role":        user.Role,
		"token":       token,
		"initialized": true,
	})
}

type UpdateSettingsRequest struct {
	SupervisorPIN        string  `json:"supervisor_pin"`
	OutletName           string  `json:"outlet_name"`
	OutletTagline        string  `json:"outlet_tagline"`
	OutletAddress        string  `json:"outlet_address"`
	OutletPhone          string  `json:"outlet_phone"`
	ReceiptFooter        string  `json:"receipt_footer"`
	TaxPercentage        float64 `json:"tax_percentage"`
	ServicePercentage    float64 `json:"service_percentage"`
	IsTaxInclusive       bool    `json:"is_tax_inclusive"`
	EnableTables         *bool   `json:"enable_tables"`
	EnableKDS            *bool   `json:"enable_kds"`
	EnableWLD            *bool   `json:"enable_wld"`
	EnableBOM            *bool   `json:"enable_bom"`
	EnableMidtrans       *bool   `json:"enable_midtrans"`
	MidtransServerKey    string  `json:"midtrans_server_key"`
	MidtransClientKey    string  `json:"midtrans_client_key"`
	MidtransIsProduction *bool   `json:"midtrans_is_production"`
}

// GetSettings mengembalikan konfigurasi lengkap outlet, kredensial Midtrans, dan feature flags untuk Superadmin.
func (h *SetupHandler) GetSettings(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)

	var outlet models.Outlet
	if err := h.db.Where("id = ?", outletID).First(&outlet).Error; err != nil {
		h.db.First(&outlet)
	}

	var sysConfigs []models.SystemConfig
	h.db.Find(&sysConfigs)

	featureFlags := map[string]bool{
		"enable_tables":   true,
		"enable_kds":      true,
		"enable_wld":      true,
		"enable_bom":      true,
		"enable_midtrans": true,
	}

	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	clientKey := os.Getenv("MIDTRANS_CLIENT_KEY")
	if strings.Contains(serverKey, "placeholder") {
		serverKey = ""
	}
	if strings.Contains(clientKey, "placeholder") {
		clientKey = ""
	}
	isProd := os.Getenv("MIDTRANS_IS_PRODUCTION") == "true"

	for _, cfg := range sysConfigs {
		if strings.HasPrefix(cfg.Key, "enable_") {
			featureFlags[cfg.Key] = (cfg.Value == "true")
		}
		if cfg.Key == "midtrans_server_key" {
			if strings.Contains(cfg.Value, "placeholder") {
				serverKey = ""
			} else {
				serverKey = cfg.Value
			}
		}
		if cfg.Key == "midtrans_client_key" {
			if strings.Contains(cfg.Value, "placeholder") {
				clientKey = ""
			} else {
				clientKey = cfg.Value
			}
		}
		if cfg.Key == "midtrans_is_production" {
			isProd = (cfg.Value == "true")
		}
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Pengaturan sistem berhasil diambil", fiber.Map{
		"outlet":                 outlet,
		"feature_flags":          featureFlags,
		"midtrans_server_key":    serverKey,
		"midtrans_client_key":    clientKey,
		"midtrans_is_production": isProd,
	})
}

// UpdateSettings memperbarui kredensial outlet, Midtrans, dan feature flags. Wajib diverifikasi dengan PIN Supervisor.
func (h *SetupHandler) UpdateSettings(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)

	var req UpdateSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Body JSON tidak valid", err.Error())
	}

	if strings.TrimSpace(req.SupervisorPIN) == "" {
		return utils.SendError(c, fiber.StatusForbidden, "PIN Supervisor wajib diisi untuk menyimpan perubahan pengaturan", nil)
	}

	// Memverifikasi PIN Supervisor
	valid, _, err := middleware.VerifySupervisorPIN(h.db, outletID, req.SupervisorPIN)
	if err != nil || !valid {
		return utils.SendError(c, fiber.StatusForbidden, "Otorisasi Gagal: PIN Supervisor tidak valid", nil)
	}

	// 1. Perbarui data Outlet
	var outlet models.Outlet
	if err := h.db.Where("id = ?", outletID).First(&outlet).Error; err != nil {
		if err := h.db.First(&outlet).Error; err != nil {
			return utils.SendError(c, fiber.StatusNotFound, "Data outlet tidak ditemukan", err.Error())
		}
	}

	if req.OutletName != "" {
		outlet.Name = req.OutletName
	}
	outlet.Tagline = req.OutletTagline
	outlet.Address = req.OutletAddress
	outlet.Phone = req.OutletPhone
	outlet.ReceiptFooter = req.ReceiptFooter
	outlet.TaxPercentage = req.TaxPercentage
	outlet.ServicePercentage = req.ServicePercentage
	outlet.IsTaxInclusive = req.IsTaxInclusive
	outlet.UpdatedAt = time.Now()

	if err := h.db.Save(&outlet).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui profil outlet", err.Error())
	}

	// 2. Perbarui SystemConfig & Env Variables
	envUpdates := make(map[string]string)

	if req.EnableTables != nil {
		saveSysConfig(h.db, "enable_tables", fmt.Sprintf("%t", *req.EnableTables))
	}
	if req.EnableKDS != nil {
		saveSysConfig(h.db, "enable_kds", fmt.Sprintf("%t", *req.EnableKDS))
	}
	if req.EnableWLD != nil {
		saveSysConfig(h.db, "enable_wld", fmt.Sprintf("%t", *req.EnableWLD))
	}
	if req.EnableBOM != nil {
		saveSysConfig(h.db, "enable_bom", fmt.Sprintf("%t", *req.EnableBOM))
	}
	if req.EnableMidtrans != nil {
		saveSysConfig(h.db, "enable_midtrans", fmt.Sprintf("%t", *req.EnableMidtrans))
	}

	if req.MidtransServerKey != "" {
		saveSysConfig(h.db, "midtrans_server_key", req.MidtransServerKey)
		envUpdates["MIDTRANS_SERVER_KEY"] = req.MidtransServerKey
		os.Setenv("MIDTRANS_SERVER_KEY", req.MidtransServerKey)
	}

	if req.MidtransClientKey != "" {
		saveSysConfig(h.db, "midtrans_client_key", req.MidtransClientKey)
		envUpdates["MIDTRANS_CLIENT_KEY"] = req.MidtransClientKey
		os.Setenv("MIDTRANS_CLIENT_KEY", req.MidtransClientKey)
	}

	if req.MidtransIsProduction != nil {
		isProdStr := fmt.Sprintf("%t", *req.MidtransIsProduction)
		saveSysConfig(h.db, "midtrans_is_production", isProdStr)
		envUpdates["MIDTRANS_IS_PRODUCTION"] = isProdStr
		os.Setenv("MIDTRANS_IS_PRODUCTION", isProdStr)
	}

	if len(envUpdates) > 0 {
		_ = updateDotEnv(envUpdates)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Pengaturan sistem berhasil diperbarui", fiber.Map{
		"outlet_id": outlet.ID,
		"updated":   true,
	})
}
