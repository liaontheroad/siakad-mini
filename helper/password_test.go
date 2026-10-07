package helper

import (
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	plain := "Rahasia123!"

	hash1, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("gagal hash: %v", err)
	}

	hash2, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("gagal hash ke-2: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("bcrypt seharusnya menghasilkan salt unik, tetapi hash1 == hash2")
	}

	if !CheckPassword(hash1, plain) {
		t.Errorf("CheckPassword gagal mengenali password yang benar")
	}

	if CheckPassword(hash1, "salah") {
		t.Errorf("CheckPassword lolos padahal password salah")
	}
}

func TestCheckPassword_Dummy(t *testing.T) {
	if CheckPassword("", "dummy") {
		t.Errorf("CheckPassword dengan hash kosong tidak boleh bernilai true")
	}
}