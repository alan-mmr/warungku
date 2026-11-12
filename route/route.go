package route

import (
	"warungku/app/service"
	"warungku/config"
	"warungku/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, cfg *config.Config, authService service.AuthService) {
	api := app.Group("/api/v1")

	// Public Health Check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"code":    200,
			"status":  "OK",
			"app":     cfg.AppName,
			"version": "1.0.0",
		})
	})

	// Public Auth Endpoints
	auth := api.Group("/auth")
	auth.Post("/register", authService.RegisterOwner)
	auth.Post("/login", authService.Login)

	// Protected Endpoints
	protected := api.Group("", middleware.RequireAuth(cfg.JWTSecret))
	protected.Get("/auth/me", authService.GetMe)

	// Admin Platform Endpoints (Role Guard: admin only)
	_ = protected.Group("/admin", middleware.RequireRole("admin"))

	// Owner Only Endpoints (Role Guard: owner only)
	_ = protected.Group("/owner", middleware.RequireRole("owner"))

	// Business Scoped Endpoints (Role Guard: owner & cashier)
	_ = protected.Group("/pos", middleware.RequireRole("owner", "cashier"))
}
