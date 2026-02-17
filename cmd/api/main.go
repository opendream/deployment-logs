package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strings"

	"deployment-logs/internal/config"
	"deployment-logs/internal/database"
	"deployment-logs/internal/handlers"
	"deployment-logs/internal/middleware"
	"deployment-logs/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if cfg.JWTSecret == "" {
		b := make([]byte, 32)
		rand.Read(b)
		cfg.JWTSecret = hex.EncodeToString(b)
		log.Println("Generated ephemeral JWT secret (set JWT_SECRET env var for persistence)")
	}

	db, err := database.Setup(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	services.SeedAdminUser(db, cfg.AdminPassword)

	app := fiber.New(fiber.Config{
		ReadBufferSize: 16384,
	})
	app.Use(cors.New())

	// When base path is set, redirect /basepath to /basepath/
	if cfg.BasePath != "" {
		app.Use(func(c *fiber.Ctx) error {
			if c.Path() == cfg.BasePath {
				return c.Redirect(cfg.BasePath+"/", fiber.StatusMovedPermanently)
			}
			return c.Next()
		})
	}

	bp := cfg.BasePath

	// Debug middleware
	app.Use(func(c *fiber.Ctx) error {
		log.Printf("DEBUG: %s %s", c.Method(), c.Path())
		return c.Next()
	})

	// Public routes
	app.Post(bp+"/api/auth/login", handlers.Login(db, cfg.JWTSecret))
	app.Get(bp+"/api/public/logs/repo/:repo_name", handlers.GetLogsByRepo(db))
	app.Post(bp+"/api/public/trigger", middleware.APIKeyAuth(cfg), handlers.Trigger(db, cfg))

	// Public read routes (with optional auth so frontend knows if user is admin)
	optAuth := app.Group(bp+"/api", middleware.OptionalJWTAuth(cfg.JWTSecret))
	optAuth.Get("/auth/me", handlers.Me())
	optAuth.Get("/repos", handlers.ListRepos(db))
	optAuth.Get("/repos/:id", handlers.GetRepo(db))
	optAuth.Get("/logs", handlers.ListLogs(db))
	optAuth.Get("/logs/repo/:repo_name", handlers.GetLogsByRepo(db))
	optAuth.Get("/logs/:id", handlers.GetLog(db))

	// Authenticated routes
	api := app.Group(bp+"/api", middleware.JWTAuth(cfg.JWTSecret))

	admin := api.Group("", middleware.RequireAdmin())

	adminRepos := admin.Group("/repos")
	adminRepos.Post("/", handlers.CreateRepo(db))
	adminRepos.Put("/:id", handlers.UpdateRepo(db))
	adminRepos.Delete("/:id", handlers.DeleteRepo(db))

	admin.Post("/trigger", handlers.Trigger(db, cfg))
	admin.Delete("/logs/repo/:repo_name", handlers.ClearLogsByRepo(db))

	settings := admin.Group("/settings")
	settings.Get("/", handlers.GetSettings(cfg))
	settings.Put("/", handlers.UpdateSettings(cfg))

	// Static assets
	app.Static(bp+"/assets", "./web/dist/assets")

	// SPA fallback: serve index.html with injected base path
	serveIndex := func(c *fiber.Ctx) error {
		data, err := os.ReadFile("./web/dist/index.html")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		html := strings.Replace(string(data), "<head>",
			`<head><script>window.__BASE_PATH__="`+cfg.BasePath+`"</script>`, 1)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}

	app.Get(bp+"/*", serveIndex)

	log.Fatal(app.Listen(":" + cfg.Port))
}
