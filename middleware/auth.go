package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/helper"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Set("WWW-Authenticate", "Bearer")
			return helper.Unauthorized("Token tidak ada, tidak valid, atau kedaluwarsa")
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		authUser, err := jwtManager.Parse(tokenStr)
		if err != nil {
			c.Set("WWW-Authenticate", "Bearer")
			return helper.Unauthorized("Token tidak ada, tidak valid, atau kedaluwarsa")
		}

		helper.SetCurrentUser(c, authUser)

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("Anda belum login")
		}

		for _, allowedRole := range roles {
			if user.Role == allowedRole {
				return c.Next()
			}
		}

		return helper.Forbidden("Anda tidak berhak mengakses endpoint ini")
	}
}