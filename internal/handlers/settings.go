package handlers

import (
	"deployment-logs/internal/config"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetSettings(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		masked := ""
		if len(cfg.AppKey) > 4 {
			masked = strings.Repeat("*", len(cfg.AppKey)-4) + cfg.AppKey[len(cfg.AppKey)-4:]
		} else if len(cfg.AppKey) > 0 {
			masked = strings.Repeat("*", len(cfg.AppKey))
		}
		return c.JSON(fiber.Map{"app_key": masked})
	}
}

func UpdateSettings(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			AppKey string `json:"app_key"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		cfg.AppKey = body.AppKey
		return c.JSON(fiber.Map{"message": "updated"})
	}
}
