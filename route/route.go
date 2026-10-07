package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/helper"
)

func Register(app *fiber.App, pool *pgxpool.Pool) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(pool))
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("Database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "Server dan database berjalan", nil)
	}
}
