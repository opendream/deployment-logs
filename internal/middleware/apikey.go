package middleware

import (
	"deployment-logs/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// APIKeyAuth checks X-API-Key header against the app_key stored in DB settings.
// If no app_key is configured, access is denied.
func APIKeyAuth(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		appKey, err := models.GetSetting(db, "app_key")
		if err != nil || appKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "API key not configured"})
		}
		key := c.Get("X-API-Key")
		if key == "" || key != appKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or missing API key"})
		}
		return c.Next()
	}
}
