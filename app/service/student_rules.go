package service

import (
	"fmt"
	"math"
	"time"

	"siakad-mini/app/model"
)

func ParseStudentListQuery(page, perPage int, prodi string, angkatan *int, search, sort string) model.StudentListQuery {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	} else if perPage > 50 {
		perPage = 50
	}
	return model.StudentListQuery{
		Page:     page,
		PerPage:  perPage,
		Prodi:    prodi,
		Angkatan: angkatan,
		Search:   search,
		Sort:     sort,
	}
}

func MaxSKS(ipk float64) int {
	scaled := int(math.Round(ipk * 100))
	if scaled >= 300 {
		return 24
	}
	if scaled >= 250 {
		return 21
	}
	return 18
}

func CurrentTahunAkademik(now time.Time) string {
	m := now.Month()
	y := now.Year()
	if m >= time.August {
		return fmt.Sprintf("%d/%d-Ganjil", y, y+1)
	}
	if m == time.January {
		return fmt.Sprintf("%d/%d-Ganjil", y-1, y)
	}
	return fmt.Sprintf("%d/%d-Genap", y-1, y)
}

func ValidateUpdate(req model.UpdateStudentRequest) map[string][]string {
	errs := make(map[string][]string)
	if req.NIM != nil {
		errs["nim"] = []string{"NIM tidak boleh diubah"}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}