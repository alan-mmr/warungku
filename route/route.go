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
	customerService service.CustomerService,
	txService service.TransactionService,
	debtService service.DebtService,
	expenseService service.ExpenseService,
	reportService service.ReportService,
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
	if txService != nil {
		owner.Post("/transactions/:id/void", txService.Void)
	}
	if expenseService != nil {
		owner.Get("/expenses", expenseService.FindAll)
		owner.Post("/expenses", expenseService.Create)
		owner.Get("/expenses/:id", expenseService.FindByID)
		owner.Put("/expenses/:id", expenseService.Update)
		owner.Delete("/expenses/:id", expenseService.Delete)
	}
	if reportService != nil {
		owner.Get("/reports/monthly", reportService.MonthlyReport)
		owner.Get("/reports/top-products", reportService.TopProducts)
	}

	// Business Scoped Endpoints (Role Guard: owner & cashier)
	pos := protected.Group("/pos", middleware.RequireRole("owner", "cashier"))
	if customerService != nil {
		pos.Get("/customers", customerService.FindAll)
		pos.Post("/customers", customerService.Create)
		pos.Get("/customers/:id", customerService.FindByID)
		pos.Put("/customers/:id", customerService.Update)
	}
	if txService != nil {
		pos.Post("/transactions", txService.Create)
		pos.Get("/transactions", txService.FindAll)
		pos.Get("/transactions/:id", txService.FindByID)
	}
	if debtService != nil {
		pos.Get("/debts", debtService.FindAll)
		pos.Get("/debts/:id", debtService.FindByID)
		pos.Post("/debts/:id/payments", debtService.Pay)
	}
	if reportService != nil {
		pos.Get("/reports/daily", reportService.DailyReport)
	}
}
