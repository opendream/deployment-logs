package handlers

import (
	"time"

	"deployment-logs/internal/models"
	"deployment-logs/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func Login(db *gorm.DB, jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}

		var user models.User
		if err := db.Where("username = ?", body.Username).First(&user).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}

		if !services.CheckPassword(user.PasswordHash, body.Password) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":      user.ID,
			"username": user.Username,
			"role":     user.Role,
			"exp":      time.Now().Add(72 * time.Hour).Unix(),
		})
		signed, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
		}

		return c.JSON(fiber.Map{
			"token": signed,
			"user": fiber.Map{
				"id":       user.ID,
				"username": user.Username,
				"role":     user.Role,
			},
		})
	}
}

func Me() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"id":       c.Locals("userID"),
			"username": c.Locals("username"),
			"role":     c.Locals("role"),
		})
	}
}
