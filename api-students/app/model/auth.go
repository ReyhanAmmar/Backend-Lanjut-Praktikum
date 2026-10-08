package model

import "time"

type RegisterRequest struct {
	NIM      string  `json:"nim" validate:"required,nim"`
	Name     string  `json:"name" validate:"required,min=3,max=150,studentname"`
	Grade    float64 `json:"grade" validate:"gte=0,lte=100"`
	Password string  `json:"password" validate:"required,strongpassword"`
}

type LoginRequest struct {
	NIM      string `json:"nim"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type RefreshToken struct {
	ID        int64
	StudentID int
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type AuthStudent struct {
	StudentID int    `json:"student_id"`
	NIM       string `json:"nim"`
	Role      string `json:"role"`
}
