package middleware

import (
	"strings"

	"warungku/helper"

	"github.com/gofiber/fiber/v2"
)

const (
	CtxUserID     = "user_id"
	CtxBusinessID = "business_id"
	CtxRole       = "role"
)

func RequireAuth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.TranslateError(c, helper.ErrUnauthorized)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return helper.TranslateError(c, helper.ErrUnauthorized)
		}

		payload, err := helper.ParseAccessToken(parts[1], jwtSecret)
		if err != nil {
			return helper.TranslateError(c, helper.ErrUnauthorized)
		}

		c.Locals(CtxUserID, payload.UserID)
		if payload.BusinessID != nil {
			c.Locals(CtxBusinessID, *payload.BusinessID)
		}
		c.Locals(CtxRole, payload.Role)

		return c.Next()
	}
}

func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals(CtxRole).(string)
		if !ok || userRole == "" {
			return helper.TranslateError(c, helper.ErrForbidden)
		}

		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return helper.TranslateError(c, helper.ErrForbidden)
	}
}
