package config

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/helper"
	"siakad-mini/middleware"
	"siakad-mini/route"
)

func NewApp(logger *slog.Logger, deps route.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "SIAKAD Mini"),
		ErrorHandler: newErrorHandler(logger),
		BodyLimit:    1 * 1024 * 1024,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS", "http://localhost:3000"))
	route.Register(app, deps)

	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("Endpoint tidak ditemukan")
	})

	return app
}

func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	production := GetEnv("APP_ENV", "development") == "production"

	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		message := "Terjadi kesalahan pada server"
		var fieldErrs map[string][]string

		var appErr *helper.AppError
		var fiberErr *fiber.Error

		switch {
		case errors.As(err, &appErr):
			status, message, fieldErrs = appErr.Status, appErr.Message, appErr.Errors
		case errors.As(err, &fiberErr):
			status, message = fiberErr.Code, fiberErr.Message
		}

		if status >= fiber.StatusInternalServerError {
			requestID, _ := c.Locals("requestid").(string)
			logger.Error("server_error",
				slog.String("request_id", requestID),
				slog.String("method", c.Method()),
				slog.String("path", c.Path()),
				slog.Int("status", status),
				slog.String("error", err.Error()),
			)
			if status == fiber.StatusInternalServerError {
				message = "Terjadi kesalahan pada server"
				if !production {
					fieldErrs = map[string][]string{"debug": {err.Error()}}
				}
			}
		}

		return helper.WriteError(c, status, message, fieldErrs)
	}
}