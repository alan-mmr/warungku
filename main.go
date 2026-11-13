package main

import (
	"log"

	"warungku/app/repository"
	"warungku/app/service"
	"warungku/config"
	"warungku/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 1. Load Configurations
	cfg := config.LoadConfig()

	// 2. Initialize Database Pool (PostgreSQL)
	dbPool, err := config.NewDatabasePool(cfg.DatabaseURL)
	if err != nil {
		log.Printf("[WARN] Database connection failed: %v", err)
		log.Println("[INFO] Pastikan PostgreSQL aktif dan database 'warungku_db' sudah dibuat.")
	} else {
		defer dbPool.Close()
		log.Println("[OK] Database connection pool established successfully.")
	}

	// 3. Initialize Repositories
	userRepo := repository.NewUserRepository(dbPool)
	customerRepo := repository.NewCustomerRepository(dbPool)

	// 4. Initialize Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTAccessExpMinutes)
	customerService := service.NewCustomerService(customerRepo)

	// 5. Initialize Fiber App & Middleware
	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// 6. Setup Routes
	route.SetupRoutes(app, cfg, authService, customerService)

	// 7. Start Server
	log.Printf("[READY] %s backend running on port :%s\n", cfg.AppName, cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}
}
