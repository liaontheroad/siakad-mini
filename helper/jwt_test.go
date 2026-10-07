package helper

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"siakad-mini/app/model"
)

func TestJWTManager_Success(t *testing.T) {
	m := NewJWTManager("super-secret-key-32-chars-minimum", "test-issuer", 1*time.Minute)
	user := model.User{ID: 5, Role: "admin"}

	tokenStr, exp, err := m.Generate(user)
	if err != nil {
		t.Fatalf("Generate gagal: %v", err)
	}
	if exp != 60 {
		t.Errorf("harap expires in 60 detik, dapat %d", exp)
	}

	authUser, err := m.Parse(tokenStr)
	if err != nil {
		t.Fatalf("Parse gagal: %v", err)
	}

	if authUser.UserID != 5 || authUser.Role != "admin" {
		t.Errorf("Klaim salah, dapat: %+v", authUser)
	}
}

func TestJWTManager_Expired(t *testing.T) {
	m := NewJWTManager("super-secret-key-32-chars-minimum", "test-issuer", -1*time.Minute)
	tokenStr, _, _ := m.Generate(model.User{ID: 1})

	_, err := m.Parse(tokenStr)
	if err == nil {
		t.Errorf("Token kedaluwarsa seharusnya ditolak")
	}
}

func TestJWTManager_WrongSecret(t *testing.T) {
	m1 := NewJWTManager("super-secret-key-32-chars-minimum", "test-issuer", 5*time.Minute)
	m2 := NewJWTManager("beda-secret-key-32-chars-minimum", "test-issuer", 5*time.Minute)

	tokenStr, _, _ := m1.Generate(model.User{ID: 1})

	_, err := m2.Parse(tokenStr)
	if err == nil {
		t.Errorf("Parse seharusnya gagal karena secret key berbeda")
	}
}

func TestJWTManager_AlgNone(t *testing.T) {
	claims := jwt.MapClaims{"sub": "1", "iss": "test-issuer", "exp": time.Now().Add(time.Minute).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	m := NewJWTManager("super-secret-key-32-chars-minimum", "test-issuer", 5*time.Minute)
	_, err := m.Parse(tokenStr)
	if err == nil {
		t.Errorf("Parse SEHARUSNYA menolak algoritma none")
	}
}