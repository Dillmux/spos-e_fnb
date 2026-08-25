package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spos/backend/internal/middleware"
	"github.com/spos/backend/internal/models"
	"github.com/spos/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
}

func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type VerifyPINRequest struct {
	PIN string `json:"pin"`
}

// Login mengautentikasi kredensial pengguna dan mengembalikan token JWT bearer.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Payload JSON tidak valid", err.Error())
	}

	if req.Username == "" || req.Password == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Username dan password wajib diisi", nil)
	}

	var user models.User
	err := h.db.Preload("Outlet").Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Username atau password salah", nil)
	}

	if !user.IsActive {
		return utils.SendError(c, fiber.StatusForbidden, "Akun pengguna tidak aktif", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Username atau password salah", nil)
	}

	token, err := middleware.GenerateJWTToken(&user, h.jwtSecret)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membuat token akses", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Login berhasil", fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":        user.ID,
			"name":      user.Name,
			"username":  user.Username,
			"email":     user.Email,
			"role":      user.Role,
			"outlet_id": user.OutletID,
			"outlet":    user.Outlet,
		},
	})
}

// RefreshToken memperbarui token JWT akses tanpa memerlukan login ulang.
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Tidak terotentikasi", nil)
	}

	var user models.User
	if err := h.db.Preload("Outlet").Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Pengguna tidak ditemukan atau tidak aktif", nil)
	}

	newToken, err := middleware.GenerateJWTToken(&user, h.jwtSecret)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui token akses", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Token akses berhasil diperbarui", fiber.Map{
		"token": newToken,
	})
}


// VerifyPIN memvalidasi 6-digit PIN supervisor/manager untuk otorisasi tindakan khusus.
func (h *AuthHandler) VerifyPIN(c *fiber.Ctx) error {
	var req VerifyPINRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Permintaan JSON tidak valid", err.Error())
	}

	if req.PIN == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "PIN wajib diisi", nil)
	}

	outletID, _ := c.Locals("outlet_id").(string)
	if outletID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Konteks otentikasi tidak valid", nil)
	}

	valid, supervisor, err := middleware.VerifySupervisorPIN(h.db, outletID, req.PIN)
	if err != nil || !valid {
		return utils.SendError(c, fiber.StatusForbidden, "PIN supervisor tidak valid", nil)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "PIN supervisor berhasil diverifikasi", fiber.Map{
		"verified_by": supervisor.Name,
		"role":        supervisor.Role,
	})
}

// Me mengembalikan profil pengguna yang sedang terotentikasi dan status shift aktifnya.
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Tidak terotentikasi", nil)
	}

	var user models.User
	if err := h.db.Preload("Outlet").Where("id = ?", userID).First(&user).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Pengguna tidak ditemukan", err.Error())
	}

	var activeShift models.Shift
	hasShift := h.db.Where("user_id = ? AND status = ?", userID, "open").First(&activeShift).Error == nil

	var shiftData interface{}
	if hasShift {
		shiftData = activeShift
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Profil pengguna berhasil diambil", fiber.Map{
		"user": fiber.Map{
			"id":        user.ID,
			"name":      user.Name,
			"username":  user.Username,
			"email":     user.Email,
			"role":      user.Role,
			"outlet_id": user.OutletID,
			"outlet":    user.Outlet,
		},
		"has_active_shift": hasShift,
		"active_shift":     shiftData,
	})
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	AuthPIN  string `json:"auth_pin"`
	IsActive *bool  `json:"is_active"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
	AuthPIN  string `json:"auth_pin,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// ListUsers mengambil daftar seluruh akun staf/pengguna pada outlet terotentikasi.
func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)
	var users []models.User
	if err := h.db.Where("outlet_id = ?", outletID).Order("created_at DESC").Find(&users).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal mengambil daftar pengguna", err.Error())
	}

	userResponses := make([]fiber.Map, 0, len(users))
	for _, u := range users {
		userResponses = append(userResponses, fiber.Map{
			"id":         u.ID,
			"outlet_id":  u.OutletID,
			"name":       u.Name,
			"username":   u.Username,
			"email":      u.Email,
			"role":       u.Role,
			"is_active":  u.IsActive,
			"has_pin":    u.AuthPINHash != "",
			"created_at": u.CreatedAt,
			"updated_at": u.UpdatedAt,
		})
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Daftar pengguna berhasil diambil", userResponses)
}

// CreateUser membuat akun staf/pengguna baru dalam outlet.
func (h *AuthHandler) CreateUser(c *fiber.Ctx) error {
	outletID, _ := c.Locals("outlet_id").(string)

	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Body JSON tidak valid", err.Error())
	}

	if req.Name == "" || req.Username == "" || req.Password == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Nama, username, dan password wajib diisi", nil)
	}

	if req.Role == "" {
		req.Role = "cashier"
	}

	var existing models.User
	if err := h.db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return utils.SendError(c, fiber.StatusConflict, "Username sudah digunakan", nil)
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal meng-hash password", err.Error())
	}

	var pinHashStr string
	if req.AuthPIN != "" {
		pinHash, err := bcrypt.GenerateFromPassword([]byte(req.AuthPIN), bcrypt.DefaultCost)
		if err == nil {
			pinHashStr = string(pinHash)
		}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	newUser := models.User{
		OutletID:     outletID,
		Name:         req.Name,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(passHash),
		AuthPINHash:  pinHashStr,
		Role:         req.Role,
		IsActive:     isActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.db.Create(&newUser).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal membuat pengguna baru di database", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusCreated, "Pengguna berhasil dibuat", fiber.Map{
		"id":        newUser.ID,
		"name":      newUser.Name,
		"username":  newUser.Username,
		"email":     newUser.Email,
		"role":      newUser.Role,
		"is_active": newUser.IsActive,
		"has_pin":   newUser.AuthPINHash != "",
	})
}

// UpdateUser memperbarui informasi akun pengguna yang ada.
func (h *AuthHandler) UpdateUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)

	var user models.User
	if err := h.db.Where("id = ? AND outlet_id = ?", userID, outletID).First(&user).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Pengguna tidak ditemukan", nil)
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Body JSON tidak valid", err.Error())
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Username != "" && req.Username != user.Username {
		var existing models.User
		if err := h.db.Where("username = ? AND id != ?", req.Username, userID).First(&existing).Error; err == nil {
			return utils.SendError(c, fiber.StatusConflict, "Username sudah digunakan", nil)
		}
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.Password != "" {
		passHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err == nil {
			user.PasswordHash = string(passHash)
		}
	}
	if req.AuthPIN != "" {
		pinHash, err := bcrypt.GenerateFromPassword([]byte(req.AuthPIN), bcrypt.DefaultCost)
		if err == nil {
			user.AuthPINHash = string(pinHash)
		}
	}

	user.UpdatedAt = time.Now()

	if err := h.db.Save(&user).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui data pengguna", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Pengguna berhasil diperbarui", fiber.Map{
		"id":        user.ID,
		"name":      user.Name,
		"username":  user.Username,
		"email":     user.Email,
		"role":      user.Role,
		"is_active": user.IsActive,
		"has_pin":   user.AuthPINHash != "",
	})
}

// DeleteUser menghapus akun pengguna secara permanen dari outlet terotentikasi.
func (h *AuthHandler) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	outletID, _ := c.Locals("outlet_id").(string)

	currentUserID, _ := c.Locals("user_id").(string)
	if userID == currentUserID {
		return utils.SendError(c, fiber.StatusBadRequest, "Tidak dapat menghapus akun Anda sendiri yang sedang digunakan", nil)
	}

	var user models.User
	if err := h.db.Where("id = ? AND outlet_id = ?", userID, outletID).First(&user).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Pengguna tidak ditemukan", nil)
	}

	if err := h.db.Delete(&user).Error; err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Gagal menghapus pengguna dari database", err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Pengguna berhasil dihapus secara permanen", fiber.Map{
		"id": user.ID,
	})
}
