package handlers

import (
	"deployment-logs/internal/config"
	"deployment-logs/internal/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Trigger(db *gorm.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			RepoName string `json:"repo_name"`
			Branch   string `json:"branch"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		if body.RepoName == "" || body.Branch == "" {
			return c.Status(400).JSON(fiber.Map{"error": "repo_name and branch are required"})
		}

		log, err := services.Generate(db, cfg, body.RepoName, body.Branch)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(log)
	}
}
