package handlers

import (
	"deployment-logs/internal/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetSettings(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		appKey, _ := models.GetSetting(db, "app_key")
		masked := ""
		if len(appKey) > 4 {
			masked = strings.Repeat("*", len(appKey)-4) + appKey[len(appKey)-4:]
		} else if len(appKey) > 0 {
			masked = strings.Repeat("*", len(appKey))
		}
		return c.JSON(fiber.Map{"app_key": masked})
	}
}

func UpdateSettings(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			AppKey string `json:"app_key"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		if err := models.SetSetting(db, "app_key", body.AppKey); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to save setting"})
		}
		return c.JSON(fiber.Map{"message": "updated"})
	}
}
