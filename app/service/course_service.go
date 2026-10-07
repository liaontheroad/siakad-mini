package service

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type CourseService struct {
	courses repository.CourseRepository
}

func NewCourseService(c repository.CourseRepository) *CourseService {
	return &CourseService{courses: c}
}

func (s *CourseService) List(c *fiber.Ctx) error {
	var semester *int
	if rawSem := c.Query("semester"); rawSem != "" {
		var sem int
		if _, err := fmtSscanfInt(rawSem, &sem); err != nil {
			return helper.Validation(map[string][]string{"semester": {"semester harus berupa angka"}})
		}
		semester = &sem
	}

	search := c.Query("search")
	availableOnly := c.Query("available") == "true"
	tahunAkademik := c.Query("tahun_akademik")
	if tahunAkademik == "" {
		tahunAkademik = CurrentTahunAkademik(time.Now())
	}

	q := model.CourseListQuery{
		Semester:      semester,
		Search:        search,
		AvailableOnly: availableOnly,
		TahunAkademik: tahunAkademik,
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	list, err := s.courses.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "Daftar mata kuliah berhasil diambil", list)
}