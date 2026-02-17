package handlers

import (
	"regexp"

	"deployment-logs/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var validNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9\-_]*$`)

func validateName(name string) string {
	if name == "" {
		return "name is required"
	}
	if !validNameRe.MatchString(name) {
		return "name must be lowercase alphanumeric with hyphens or underscores only (e.g. my-repo_name)"
	}
	return ""
}

func ListRepos(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var repos []models.RepositoryConfig
		if err := db.Find(&repos).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		// Omit SSH key from list response
		for i := range repos {
			repos[i].SSHKey = ""
		}
		return c.JSON(repos)
	}
}

func CreateRepo(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var repo models.RepositoryConfig
		if err := c.BodyParser(&repo); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		if msg := validateName(repo.Name); msg != "" {
			return c.Status(400).JSON(fiber.Map{"error": msg})
		}
		if err := db.Create(&repo).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		repo.SSHKey = ""
		return c.Status(201).JSON(repo)
	}
}

func GetRepo(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var repo models.RepositoryConfig
		if err := db.First(&repo, id).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}
		// Omit SSH key from GET response
		repo.SSHKey = ""
		return c.JSON(repo)
	}
}

func UpdateRepo(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var repo models.RepositoryConfig
		if err := db.First(&repo, id).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}

		// Preserve existing SSH key
		existingSSHKey := repo.SSHKey

		if err := c.BodyParser(&repo); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// If SSH key not provided in update, keep existing
		if repo.SSHKey == "" {
			repo.SSHKey = existingSSHKey
		}

		if msg := validateName(repo.Name); msg != "" {
			return c.Status(400).JSON(fiber.Map{"error": msg})
		}
		if err := db.Save(&repo).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		repo.SSHKey = ""
		return c.JSON(repo)
	}
}

func DeleteRepo(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := db.Delete(&models.RepositoryConfig{}, id).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"message": "deleted"})
	}
}
