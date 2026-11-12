package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName              string
	AppPort              string
	AppEnv               string
	DatabaseURL          string
	JWTSecret            string
	JWTAccessExpMinutes  int
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		AppName:             getEnv("APP_NAME", "WarungKu"),
		AppPort:             getEnv("APP_PORT", "3000"),
		AppEnv:              getEnv("APP_ENV", "development"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://ASUS@localhost:5432/warungku_db?sslmode=disable"),
		JWTSecret:           getEnv("JWT_SECRET", "supersecret_warungku_jwt_key_2026_dev"),
		JWTAccessExpMinutes: 60,
	}
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
