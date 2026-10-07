package service

import (
	"testing"
	"time"
)

func TestMaxSKS(t *testing.T) {
	cases := []struct {
		ipk  float64
		want int
	}{
		{4.00, 24},
		{3.00, 24},
		{2.99, 21},
		{2.50, 21},
		{2.49, 18},
		{0.00, 18},
	}

	for _, tc := range cases {
		got := MaxSKS(tc.ipk)
		if got != tc.want {
			t.Errorf("IPK %.2f: harap %d SKS, dapat %d", tc.ipk, tc.want, got)
		}
	}
}

func TestCurrentTahunAkademik(t *testing.T) {
	jan := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	if got := CurrentTahunAkademik(jan); got != "2025/2026-Ganjil" {
		t.Errorf("Januari 2026: harap 2025/2026-Ganjil, dapat %s", got)
	}

	feb := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	if got := CurrentTahunAkademik(feb); got != "2025/2026-Genap" {
		t.Errorf("Februari 2026: harap 2025/2026-Genap, dapat %s", got)
	}

	aug := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
	if got := CurrentTahunAkademik(aug); got != "2026/2027-Ganjil" {
		t.Errorf("Agustus 2026: harap 2026/2027-Ganjil, dapat %s", got)
	}
}

func TestParseStudentListQuery_Normalization(t *testing.T) {
	q1 := ParseStudentListQuery(0, 0, "", nil, "", "")
	if q1.Page != 1 || q1.PerPage != 10 {
		t.Errorf("Normalisasi bawah gagal: %+v", q1)
	}

	q2 := ParseStudentListQuery(1, 1000, "", nil, "", "")
	if q2.PerPage != 50 {
		t.Errorf("Normalisasi batas atas per_page gagal, harap 50 dapat %d", q2.PerPage)
	}
}