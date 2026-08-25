package utils

import (
	"github.com/gofiber/fiber/v2"
)

// Response mendefinisikan struktur respon JSON standar yang seragam di seluruh endpoint API.
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// SendSuccess mengirimkan respon sukses bernilai JSON yang terstandardisasi.
func SendSuccess(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendError mengirimkan respon galat/error bernilai JSON yang terstandardisasi.
func SendError(c *fiber.Ctx, status int, message string, errors interface{}) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}
