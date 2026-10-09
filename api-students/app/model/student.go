package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	OwnerID   int       `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type AssignRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

type CreateStudentRequest struct {
	NIM   string   `json:"nim" validate:"required,nim"`
	Name  string   `json:"name" validate:"required,min=3,max=150,studentname"`
	Grade *float64 `json:"grade" validate:"required,gte=0,lte=100"`
}

type ReplaceStudentRequest struct {
	NIM      string  `json:"nim" validate:"required,nim"`
	Name     string  `json:"name" validate:"required,min=3,max=150,studentname"`
	Grade    float64 `json:"grade" validate:"gte=0,lte=100"`
	IsActive bool    `json:"is_active"`
}

type UpdateStudentRequest struct {
	NIM      *string  `json:"nim" validate:"omitnil,required,nim"`
	Name     *string  `json:"name" validate:"omitnil,min=3,max=150,studentname"`
	Grade    *float64 `json:"grade" validate:"omitnil,gte=0,lte=100"`
	IsActive *bool    `json:"is_active"`
}

type PatchStudentRequest struct {
	NIM      string   `json:"nim,omitempty" validate:"omitnil,required,nim"`
	Name     *string  `json:"name" validate:"omitnil,min=3,max=150,studentname"`
	Grade    *float64 `json:"grade" validate:"omitnil,gte=0,lte=100"`
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

type ErrorResponse struct {
    Success   bool              `json:"success"`
    Code      string            `json:"code"`
    Message   string            `json:"message"`
    Fields    map[string]string `json:"fields,omitempty"`
    RequestID string            `json:"request_id,omitempty"`
}

type Cursor struct { CreatedAt time.Time; ID int }
type CursorQuery struct { Limit int; Search string; IsActive *bool; After *Cursor }
type CursorMeta struct {
	Limit int `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore bool `json:"has_more"`
}
