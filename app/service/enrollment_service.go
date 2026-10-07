package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type EnrollmentService struct {
	enrollments repository.EnrollmentRepository
	students    repository.StudentRepository
}

func NewEnrollmentService(e repository.EnrollmentRepository, s repository.StudentRepository) *EnrollmentService {
	return &EnrollmentService{
		enrollments: e,
		students:    s,
	}
}

func (s *EnrollmentService) Create(c *fiber.Ctx) error {
	var req model.CreateEnrollmentRequest
	if err := helper.BindAndValidate(c, &req); err != nil {
		return err
	}

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Anda belum login")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.students.FindByUserID(ctx, authUser.UserID)
	if err != nil {
		return helper.Forbidden("Hanya mahasiswa yang dapat mengambil KRS")
	}

	maxSKS := MaxSKS(student.IPKTerakhir)

	in := repository.CreateEnrollmentInput{
		StudentID:     student.ID,
		CourseID:      req.CourseID,
		TahunAkademik: req.TahunAkademik,
		MaxSKS:        maxSKS,
	}

	en, err := s.enrollments.CreateWithRules(ctx, in)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAlreadyEnrolled):
			return helper.Conflict("Mata kuliah ini sudah Anda ambil pada tahun akademik tersebut")
		case errors.Is(err, repository.ErrQuotaFull):
			return helper.Unprocessable("Kuota mata kuliah sudah penuh")
		case errors.Is(err, repository.ErrCourseNotFound):
			return helper.Validation(map[string][]string{"course_id": {"Mata kuliah tidak ditemukan"}})
		}

		var sksErr *repository.SKSExceededError
		if errors.As(err, &sksErr) {
			msg := FormatSKSMessage(sksErr.Batas, sksErr.Terpakai, sksErr.Sisa, sksErr.SKSBaru, sksErr.IPK)
			return helper.Unprocessable(msg)
		}

		return helper.Internal(err)
	}

	return helper.Created(c, "Mata kuliah berhasil diambil", en, "/api/v1/enrollments/"+itoa(en.ID))
}

func (s *EnrollmentService) Delete(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("Anda belum login")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	en, err := s.enrollments.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Data KRS tidak ditemukan")
		}
		return helper.Internal(err)
	}

	student, err := s.students.FindByID(ctx, en.StudentID)
	if err != nil || student.UserID != authUser.UserID {
		return helper.Forbidden("Anda hanya boleh membatalkan KRS milik sendiri")
	}

	err = s.enrollments.Delete(ctx, id)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}