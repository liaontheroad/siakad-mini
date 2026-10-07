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
	DeletedAt   *time.Time `json:"-"` // nil = aktif
}
