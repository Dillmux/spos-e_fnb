package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spos/backend/internal/models"
	"github.com/spos/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type JWTClaims struct {
	UserID   string `json:"user_id"`
	OutletID string `json:"outlet_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWTToken membuat token akses JWT baru untuk pengguna yang terotentikasi.
func GenerateJWTToken(user *models.User, jwtSecret string) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		OutletID: user.OutletID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// JWTMiddleware memvalidasi token Bearer pada header Authorization.
func JWTMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.SendError(c, fiber.StatusUnauthorized, "Header Authorization tidak ditemukan", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return utils.SendError(c, fiber.StatusUnauthorized, "Format header Authorization tidak valid. Format yang diharapkan: Bearer <token>", nil)
		}

		tokenString := parts[1]
		claims := &JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode penandatanganan tidak diharapkan: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return utils.SendError(c, fiber.StatusUnauthorized, "Token JWT tidak valid atau telah kedaluwarsa", nil)
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("outlet_id", claims.OutletID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		c.Locals("jwt_claims", claims)

		return c.Next()
	}
}

// RoleGuard membatasi eksekusi endpoint hanya untuk role pengguna yang diizinkan.
func RoleGuard(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			return utils.SendError(c, fiber.StatusForbidden, "Akses ditolak: role pengguna tidak terdefinisi", nil)
		}

		for _, allowed := range allowedRoles {
			if strings.EqualFold(userRole, allowed) {
				return c.Next()
			}
		}

		return utils.SendError(c, fiber.StatusForbidden, fmt.Sprintf("Akses ditolak untuk role '%s'", userRole), nil)
	}
}

// VerifySupervisorPIN memeriksa apakah 6-digit PIN yang diberikan cocok dengan supervisor/manager/superadmin aktif di outlet terkait.
func VerifySupervisorPIN(db *gorm.DB, outletID, pin string) (bool, *models.User, error) {
	if outletID == "" || pin == "" {
		return false, nil, nil
	}
	var supervisors []models.User
	err := db.Where("outlet_id = ? AND role IN ('manager', 'superadmin') AND is_active = ?", outletID, true).Find(&supervisors).Error
	if err != nil {
		return false, nil, err
	}

	for _, supervisor := range supervisors {
		if supervisor.AuthPINHash != "" {
			if bcrypt.CompareHashAndPassword([]byte(supervisor.AuthPINHash), []byte(pin)) == nil {
				return true, &supervisor, nil
			}
		}
	}

	return false, nil, nil
}
