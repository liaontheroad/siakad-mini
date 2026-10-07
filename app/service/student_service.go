package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type StudentService struct {
	students repository.StudentRepository
}

func NewStudentService(s repository.StudentRepository) *StudentService {
	return &StudentService{students: s}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 10)
	prodi := c.Query("prodi")
	search := c.Query("search")
	sort := c.Query("sort")

	var angkatan *int
	if rawAngkatan := c.Query("angkatan"); rawAngkatan != "" {
		var a int
		if _, err := fmtSscanfInt(rawAngkatan, &a); err == nil {
			angkatan = &a
		} else {
			return helper.Validation(map[string][]string{"angkatan": {"angkatan harus berupa angka"}})
		}
	}

	q := ParseStudentListQuery(page, perPage, prodi, angkatan, search, sort)

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	list, total, err := s.students.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	meta := helper.NewMeta(q.Page, q.PerPage, total)
	return helper.SuccessList(c, "Daftar mahasiswa berhasil diambil", list, meta)
}

func fmtSscanfInt(s string, v *int) (int, error) {
	return fmt.Sscanf(s, "%d", v)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := helper.BindAndValidate(c, &req); err != nil {
		return err
	}

	ipk := 0.00
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	hashedPassword, err := helper.HashPassword(req.NIM) 
	if err != nil {
		return helper.Internal(err)
	}

	userEntity := model.User{
		Email:    req.Email,
		Password: hashedPassword,
		Role:     model.RoleMahasiswa,
	}

	studentEntity := model.Student{
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: ipk,
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	created, err := s.students.Create(ctx, userEntity, studentEntity)
	if err != nil {
		var dupErr *repository.DuplicateFieldError
		if errors.As(err, &dupErr) {
			return helper.Validation(map[string][]string{
				dupErr.Field: {fmt.Sprintf("%s sudah terdaftar", dupErr.Field)},
			})
		}
		return helper.Internal(err)
	}

	return helper.Created(c, "Mahasiswa berhasil didaftarkan", created, "/api/v1/students/"+itoa(created.ID))
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

func (s *StudentService) Get(c *fiber.Ctx) error {
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

	if authUser.Role == model.RoleMahasiswa {
		myStudent, err := s.students.FindByUserID(ctx, authUser.UserID)
		if err != nil || myStudent.ID != id {
			return helper.Forbidden("Anda hanya boleh mengakses data sendiri")
		}
	}

	student, err := s.students.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	tahunAkademik := c.Query("tahun_akademik")
	if tahunAkademik == "" {
		tahunAkademik = CurrentTahunAkademik(time.Now())
	}

	courses, err := s.students.ListEnrollmentsForStudent(ctx, student.ID, tahunAkademik)
	if err != nil {
		return helper.Internal(err)
	}

	totalSKS := 0
	for _, course := range courses {
		totalSKS += course.SKS
	}

	detail := model.StudentDetail{
		Student:  student,
		Courses:  courses,
		TotalSKS: totalSKS,
		BatasSKS: MaxSKS(student.IPKTerakhir),
	}

	return helper.Success(c, fiber.StatusOK, "Detail mahasiswa berhasil diambil", detail)
}

func (s *StudentService) Update(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	var req model.UpdateStudentRequest
	if err := helper.BindAndValidate(c, &req); err != nil {
		return err
	}

	if errs := ValidateUpdate(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	existing, err := s.students.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if req.Nama != nil {
		existing.Nama = *req.Nama
	}
	if req.Prodi != nil {
		existing.Prodi = *req.Prodi
	}
	if req.Angkatan != nil {
		existing.Angkatan = *req.Angkatan
	}
	if req.IPKTerakhir != nil {
		existing.IPKTerakhir = *req.IPKTerakhir
	}

	updated, err := s.students.Update(ctx, existing)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", updated)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	err = s.students.SoftDelete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}