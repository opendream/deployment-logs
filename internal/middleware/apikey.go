package middleware

import (
	"deployment-logs/internal/config"

	"github.com/gofiber/fiber/v2"
)

// APIKeyAuth checks X-API-Key header against cfg.AppKey.
// If AppKey is empty, the endpoint is open (no auth required).
func APIKeyAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if cfg.AppKey == "" {
			return c.Next()
		}
		key := c.Get("X-API-Key")
		if key == "" || key != cfg.AppKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or missing API key"})
		}
		return c.Next()
	}
}
