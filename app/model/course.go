package model

type Course struct {
	ID       int    `json:"id"`
	KodeMK   string `json:"kode_mk"`
	NamaMK   string `json:"nama_mk"`
	SKS      int    `json:"sks"`
	Semester int    `json:"semester"`
	Kuota    int    `json:"kuota"`
}

type CourseView struct {
	Course
	Terisi     int `json:"terisi"`
	SisaKuota  int `json:"sisa_kuota"`
}

type CourseListQuery struct {
	Semester       *int   `json:"semester"`
	Search         string `json:"search"`
	AvailableOnly  bool   `json:"available_only"`
	TahunAkademik  string `json:"tahun_akademik"`
}