package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"siakad-mini/app/repository"
	"siakad-mini/app/service"
	"siakad-mini/config"
	"siakad-mini/database"
	"siakad-mini/helper"
	"siakad-mini/route"
)

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < 32 {
		logger.Error("JWT_SECRET tidak valid", slog.String("error", "Harus diatur di .env dan minimal 32 karakter"))
		os.Exit(1)
	}
	jwtIssuer := config.GetEnv("JWT_ISSUER", "siakad-mini")
	jwtTTLMin := config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 60)
	jwtManager := helper.NewJWTManager(jwtSecret, jwtIssuer, time.Duration(jwtTTLMin)*time.Minute)

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	
	authService := service.NewAuthService(userRepo, studentRepo, jwtManager)
	studentService := service.NewStudentService(studentRepo)

	deps := route.Dependencies{
		Pool:    pool,
		JWT:     jwtManager,
		Auth:    authService,
		Student: studentService,
	}

	app := config.NewApp(logger, deps)
	port := config.GetEnv("APP_PORT", "3000")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("server berjalan", slog.String("port", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal berhenti diterima, menutup server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal menutup server dengan rapi", slog.String("error", err.Error()))
	}
	logger.Info("server berhenti dengan rapi")
}