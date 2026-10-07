package service

import (
	"testing"
)

func TestCheckSKS(t *testing.T) {
	ok, sisa := CheckSKS(18, 4, 22)
	if !ok || sisa != 4 {
		t.Errorf("Harap ok=true, sisa=4, dapat ok=%v, sisa=%d", ok, sisa)
	}

	ok, sisa = CheckSKS(20, 4, 21)
	if ok || sisa != 1 {
		t.Errorf("Harap ok=false, sisa=1, dapat ok=%v, sisa=%d", ok, sisa)
	}
}

func TestFormatSKSMessage(t *testing.T) {
	msg := FormatSKSMessage(21, 18, 3, 4, 2.75)
	expected := "Total SKS melebihi batas 21 SKS (IPK 2.75). SKS terpakai 18, sisa 3, mata kuliah ini 4 SKS."
	if msg != expected {
		t.Errorf("Pesan salah:\nDapat: %s\nHarap: %s", msg, expected)
	}
}