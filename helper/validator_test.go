package helper

import (
	"testing"
	"time"
)

type sampleRequest struct {
	NIM      string `json:"nim" validate:"required,len=12,numeric"`
	Email    string `json:"email" validate:"required,email"`
	Angkatan int    `json:"angkatan" validate:"required,maxyear"`
}

func TestValidate_Lolos(t *testing.T) {
	req := sampleRequest{NIM: "187221000001", Email: "a@b.com", Angkatan: time.Now().Year()}
	if errs := Validate(req); errs != nil {
		t.Fatalf("seharusnya lolos, dapat %v", errs)
	}
}

func TestValidate_NamaFieldMemakaiNamaJSON(t *testing.T) {
	errs := Validate(sampleRequest{NIM: "123", Email: "bukan-email", Angkatan: time.Now().Year() + 1})

	for _, field := range []string{"nim", "email", "angkatan"} {
		msgs, ada := errs[field]
		if !ada || len(msgs) == 0 {
			t.Errorf("harap ada error untuk field %q, dapat %v", field, errs)
		}
	}
}

func TestValidate_PesanBerupaArray(t *testing.T) {
	errs := Validate(sampleRequest{NIM: "123", Email: "a@b.com", Angkatan: 2022})

	got := errs["nim"]
	if len(got) != 1 || got[0] != "harus tepat 12 karakter" {
		t.Errorf("pesan nim tidak sesuai: %v", got)
	}
}

func TestNewMeta(t *testing.T) {
	cases := []struct{ total, perPage, want int }{
		{0, 10, 0}, {1, 10, 1}, {10, 10, 1}, {11, 10, 2}, {20, 10, 2},
	}
	for _, tc := range cases {
		if got := NewMeta(1, tc.perPage, tc.total).LastPage; got != tc.want {
			t.Errorf("total=%d per_page=%d: harap %d, dapat %d", tc.total, tc.perPage, tc.want, got)
		}
	}
}
