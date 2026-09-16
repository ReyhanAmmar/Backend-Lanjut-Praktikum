package helper

import (
    "errors"
    "fmt"
    "strconv"
    "time"
 
    "github.com/golang-jwt/jwt/v5"
 
    "api-students/app/model"
)

var (
	ErrInvalidToken = errors.New("token tidak valid")
	ErrExpiredToken = errors.New("token sudah kedaluwarsa")
)

type accessClaims struct {
	NIM  string `json:"nim"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

func NewJWTManager(secret, issuer string, accessTTL time.Duration) *JWTManager {
    return &JWTManager{secret: []byte(secret), issuer: issuer, accessTTL: accessTTL}
}

func (m *JWTManager) AccessTTL() time.Duration {return m.accessTTL}

func (m *JWTManager) GenerateAccess(student model.Student) (string, error) {
	now := time.Now()

	claims := accessClaims{
		NIM:  student.NIM,
		Role: "student",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(student.ID),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) Parse(tokenString string) (model.AuthStudent, error) {
	claims := &accessClaims{}

	token, err := jwt.ParseWithClaims(tokenString,claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("algoritma tidak diharapkan: %v", t.Header["alg"])
			}
			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return model.AuthStudent{}, ErrExpiredToken
		}
		return model.AuthStudent{}, ErrInvalidToken
	}

	if !token.Valid {
		return model.AuthStudent{}, ErrInvalidToken
	}

	studentID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return model.AuthStudent{}, ErrInvalidToken
	}

	return model.AuthStudent{
		StudentID: studentID,
		NIM:       claims.NIM,
		Role:      claims.Role,
	}, nil
}