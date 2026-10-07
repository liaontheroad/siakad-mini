package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type AuthService struct {
	users    repository.UserRepository
	students repository.StudentRepository
	jwt      *helper.JWTManager
}

func NewAuthService(u repository.UserRepository, s repository.StudentRepository, j *helper.JWTManager) *AuthService {
	return &AuthService{
		users:    u,
		students: s,
		jwt:      j,
	}
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := helper.BindAndValidate(c, &req); err != nil {
		return err 
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			helper.CheckPassword("", "dummy")
			return helper.Unauthorized("Email atau password salah")
		}
		return helper.Internal(err)
	}

	if !helper.CheckPassword(user.Password, req.Password) {
		return helper.Unauthorized("Email atau password salah")
	}

	tokenStr, expiresIn, err := s.jwt.Generate(user)
	if err != nil {
		return helper.Internal(err)
	}

	resp := model.LoginResponse{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        user,
	}

	return helper.Success(c, fiber.StatusOK, "Login berhasil", resp)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Anda belum login")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.users.FindByID(ctx, authUser.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("Akun Anda tidak ditemukan")
		}
		return helper.Internal(err)
	}

	result := fiber.Map{
		"user": user,
	}

	if user.Role == model.RoleMahasiswa {
		student, err := s.students.FindByUserID(ctx, user.ID)
		if err == nil {
			result["student"] = student
		}
	}

	return helper.Success(c, fiber.StatusOK, "Profil pengguna berhasil diambil", result)
}