package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	NIM       string   `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateStudentRequest struct {
	NIM   string `json:"nim"`
	Name  string  `json:"name"`
	Grade *float64 `json:"grade"`
}

type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

type UpdateStudentRequest struct {
	NIM      *string  `json:"nim"`
	Name     *string  `json:"name"`
	Grade    *float64 `json:"grade"`
	IsActive *bool    `json:"is_active"`
}

type PatchStudentRequest struct {
	NIM      *string  `json:"nim"`
	Name     *string  `json:"name"`
	Grade    *float64 `json:"grade"`
	IsActive *bool    `json:"is_active"`
}

type WebResponse struct {
	Success bool   `json:"success"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
    Meta    *Meta  `json:"meta,omitempty"`
    Errors  any    `json:"errors,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
    Limit      int `json:"limit"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}

type ListQuery struct {
    Page     int
    Limit    int
    Search   string
    Sort     string
    Order    string
    IsActive *bool
	MinGrade *float64
	MaxGrade *float64
}

func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}

type RegisterRequest struct {
    NIM 	 string `json:"nim"`
    Name     string `json:"name"`
	Grade	 float64 `json:"grade"`
    Password string `json:"password"`
}

type LoginRequest struct {
    NIM string `json:"nim"`
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
    ID        	int64
    StudentID   int
    TokenHash 	string
    ExpiresAt 	time.Time
    RevokedAt 	*time.Time
    CreatedAt 	time.Time
}

type AuthStudent struct {
    StudentID   int    `json:"student_id"`
    NIM         string `json:"nim"`
    Role        string `json:"role"`
}
