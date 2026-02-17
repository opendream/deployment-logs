package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	app.Post("/api/auth/login", func(c *fiber.Ctx) error {
		return c.SendString("OK-login")
	})
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendString("SPA")
	})
	log.Fatal(app.Listen(":4903"))
}
