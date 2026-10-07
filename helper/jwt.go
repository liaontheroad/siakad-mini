package helper

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"siakad-mini/app/model"
)

type JWTManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

func NewJWTManager(secret, issuer string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
	}
}

func (m *JWTManager) Generate(user model.User) (string, int, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", user.ID), 
		"role": user.Role,
		"iss":  m.issuer,
		"iat":  now.Unix(),
		"exp":  now.Add(m.ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}

	return signedToken, int(m.ttl.Seconds()), nil
}

func (m *JWTManager) Parse(tokenStr string) (model.AuthUser, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritma tidak dikenal: %v", token.Header["alg"])
		}
		return m.secret, nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
	)

	if err != nil || !token.Valid {
		return model.AuthUser{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return model.AuthUser{}, fmt.Errorf("klaim token tidak valid")
	}

	var userID int
	if sub, ok := claims["sub"].(string); ok {
		fmt.Sscanf(sub, "%d", &userID)
	}
	role, _ := claims["role"].(string)

	return model.AuthUser{
		UserID: userID,
		Role:   role,
	}, nil
}