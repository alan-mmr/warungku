package route

import (
	"warungku/app/service"
	"warungku/config"
	"warungku/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	cfg *config.Config,
	authService service.AuthService,
	categoryService service.CategoryService,
	productService service.ProductService,
	stockService service.StockService,
) {
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
	owner := protected.Group("/owner", middleware.RequireRole("owner"))
	if categoryService != nil {
		owner.Post("/categories", categoryService.Create)
		owner.Put("/categories/:id", categoryService.Update)
		owner.Delete("/categories/:id", categoryService.Delete)
	}
	if productService != nil {
		owner.Post("/products", productService.Create)
		owner.Put("/products/:id", productService.Update)
		owner.Delete("/products/:id", productService.Delete)
		owner.Post("/products/:id/variants", productService.CreateVariant)
	}
	if stockService != nil {
		owner.Get("/stock/movements", stockService.GetMovements)
		owner.Post("/stock/in", stockService.StockIn)
		owner.Post("/stock/adjust", stockService.StockAdjust)
	}

	// Business Scoped Endpoints (Role Guard: owner & cashier)
	pos := protected.Group("/pos", middleware.RequireRole("owner", "cashier"))
	if categoryService != nil {
		pos.Get("/categories", categoryService.FindAll)
		pos.Get("/categories/:id", categoryService.FindByID)
	}
	if productService != nil {
		pos.Get("/products", productService.FindAll)
		pos.Get("/products/:id", productService.FindByID)
	}
	if stockService != nil {
		pos.Get("/stock/low", stockService.GetLowStock)
	}
}
