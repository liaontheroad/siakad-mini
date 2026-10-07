package model

import "time"

type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"-"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"-"` 
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim" validate:"required,len=12,numeric"`
	Nama        string   `json:"nama" validate:"required,max=100"`
	Email       string   `json:"email" validate:"required,email"`
	Prodi       string   `json:"prodi" validate:"required,max=100"`
	Angkatan    int      `json:"angkatan" validate:"required,gte=1900,maxyear"`
	IPKTerakhir *float64 `json:"ipk_terakhir" validate:"omitempty,gte=0,lte=4"`
}

type UpdateStudentRequest struct {
	Nama        *string  `json:"nama,omitempty" validate:"omitempty,max=100"`
	Prodi       *string  `json:"prodi,omitempty" validate:"omitempty,max=100"`
	Angkatan    *int     `json:"angkatan,omitempty" validate:"omitempty,gte=1900,maxyear"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,gte=0,lte=4"`
	NIM         *string  `json:"nim,omitempty"` 
}

type StudentListQuery struct {
	Page     int     `json:"page"`
	PerPage  int     `json:"per_page"`
	Prodi    string  `json:"prodi"`
	Angkatan *int    `json:"angkatan"`
	Search   string  `json:"search"`
	Sort     string  `json:"sort"`
}

type EnrolledCourse struct {
	EnrollmentID  int    `json:"enrollment_id"`
	CourseID      int    `json:"course_id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	TahunAkademik string `json:"tahun_akademik"`
}

type StudentDetail struct {
	Student
	Courses  []EnrolledCourse `json:"courses"`
	TotalSKS int              `json:"total_sks"`
	BatasSKS int              `json:"batas_sks"`
}