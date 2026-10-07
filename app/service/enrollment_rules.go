package service

import (
	"fmt"
)

func CheckSKS(current, add, max int) (bool, int) {
	sisa := max - current
	if current+add > max {
		return false, sisa
	}
	return true, sisa
}

func FormatSKSMessage(max, terpakai, sisa, sksBaru int, ipk float64) string {
	return fmt.Sprintf("Total SKS melebihi batas %d SKS (IPK %.2f). SKS terpakai %d, sisa %d, mata kuliah ini %d SKS.", max, ipk, terpakai, sisa, sksBaru)
}