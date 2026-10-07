package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"
)

type Dependencies struct {
	Pool         *pgxpool.Pool
	JWT          *helper.JWTManager
	Auth         *service.AuthService
	Student      *service.StudentService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool))

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.Auth.Login)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.Auth.Me)

	students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	
	students.Get("/", middleware.RequireRole(model.RoleAdmin), deps.Student.List)
	students.Post("/", middleware.RequireRole(model.RoleAdmin), deps.Student.Create)
	students.Put("/:id", middleware.RequireRole(model.RoleAdmin), deps.Student.Update)
	students.Delete("/:id", middleware.RequireRole(model.RoleAdmin), deps.Student.Delete)

	students.Get("/:id", deps.Student.Get)
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