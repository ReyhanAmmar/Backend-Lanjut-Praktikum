package service

import (
    "context"
    "errors"
    "strconv"
    "strings"
    "time"
 
    "github.com/gofiber/fiber/v2"
 
    "api-students/app/model"
    "api-students/app/repository"
    "api-students/helper"
)
 
const refreshTokenBytes = 32

type AuthService struct {
    students    repository.StudentRepository
    tokens      repository.TokenRepository
    jwt         *helper.JWTManager
    perms       *helper.PermissionSet
    refreshTTL  time.Duration
}

func NewAuthService(
    students    repository.StudentRepository,
    tokens      repository.TokenRepository,
    jwtManager  *helper.JWTManager,
    perms       *helper.PermissionSet,
    refreshTTL  time.Duration,
) *AuthService {
    return &AuthService{
        students:   students,
        tokens:     tokens,
        jwt:        jwtManager,
        perms:      perms,
        refreshTTL: refreshTTL,
    }
}

func (s *AuthService) Register(c *fiber.Ctx) error {
    ctx, cancel := helper.RequestContext(c)
    defer cancel()
 
    var req model.RegisterRequest
    if err := c.BodyParser(&req); err != nil {
        return helper.BadRequest("body harus berupa JSON yang valid")
    }
 
    req.NIM = strings.TrimSpace(req.NIM)
    req.Name = strings.TrimSpace(req.Name)

    if errs := helper.ValidateStruct(req); len(errs) > 0 {
        return helper.Validation(errs)
    }

	hashed, err := helper.HashPassword(req.Password)
    if err != nil {
        return helper.Internal(err)
    }

	created, err := s.students.Create(ctx, model.Student{
        NIM:      req.NIM,
        Name:     req.Name,
        Grade:    req.Grade,
        Password: hashed,
        Role:     "user",
        IsActive: true,
    })
    if err != nil {
        if errors.Is(err, repository.ErrDuplicate) {
            return helper.Conflict("NIM sudah dipakai")
        }
        return helper.Internal(err)
    }

	return helper.Created(c, "pendaftaran berhasil", created,
        "/api/v1/students/"+strconv.Itoa(created.ID))
}

func (s *AuthService) Login(c *fiber.Ctx) error {
    ctx, cancel := helper.RequestContext(c)
    defer cancel()
 
    var req model.LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return helper.BadRequest("body harus berupa JSON yang valid")
    }
 
    if errs := helper.ValidateStruct(req); errs != nil {
        return helper.Validation(errs)
    }
 
    student, err := s.students.FindByNIM(ctx, strings.TrimSpace(req.NIM))
    if err != nil {
        if !errors.Is(err, repository.ErrNotFound) { return helper.Internal(err) }
        helper.VerifyDummyPassword(req.Password)
        return helper.Unauthorized("NIM atau password salah")
    }
 
    if !helper.VerifyPassword(student.Password, req.Password) {
        return helper.Unauthorized("NIM atau password salah")
    }
 
    if !student.IsActive {
        return helper.Forbidden("akun dinonaktifkan")
    }
 
    pair, err := s.issueTokenPair(ctx, student)
    if err != nil {
        return helper.Internal(err)
    }
 
    return helper.Success(c, fiber.StatusOK, "login berhasil", pair)
}

func (s *AuthService) Refresh(c *fiber.Ctx) error {
    ctx, cancel := helper.RequestContext(c)
    defer cancel()
 
    var req model.RefreshRequest
    if err := c.BodyParser(&req); err != nil {
        return helper.BadRequest("body harus berupa JSON yang valid")
    }
 
    if strings.TrimSpace(req.RefreshToken) == "" {
        return helper.BadRequest("refresh_token wajib diisi")
    }
 
    hash := helper.SHA256Hex(req.RefreshToken)
 
    stored, err := s.tokens.FindActive(ctx, hash)
    if err != nil {
        return helper.Unauthorized("refresh token tidak valid atau sudah kedaluwarsa")
    }
 
    student, err := s.students.FindByID(ctx, stored.StudentID)
    if err != nil {
        if !errors.Is(err, repository.ErrNotFound) { return helper.Internal(err) }
        return helper.Unauthorized("akun tidak dapat dipakai")
    }
    if !student.IsActive {
        return helper.Unauthorized("akun tidak dapat dipakai")
    }
 
    if err := s.tokens.Revoke(ctx, hash); err != nil {
        return helper.Internal(err)
    }
 
    pair, err := s.issueTokenPair(ctx, student)
    if err != nil {
        return helper.Internal(err)
    }
 
    return helper.Success(c, fiber.StatusOK, "token berhasil diperbarui", pair)
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
    ctx, cancel := helper.RequestContext(c)
    defer cancel()
 
    var req model.RefreshRequest
    if err := c.BodyParser(&req); err != nil {
        return helper.BadRequest("body harus berupa JSON yang valid")
    }
 
    if strings.TrimSpace(req.RefreshToken) != "" {
        _ = s.tokens.Revoke(ctx, helper.SHA256Hex(req.RefreshToken))
    }
 
    return helper.Success(c, fiber.StatusOK, "logout berhasil", nil)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
    ctx, cancel := helper.RequestContext(c)
    defer cancel()
 
    authStudent, ok := helper.CurrentStudent(c)
    if !ok {
        return helper.Unauthorized("belum terautentikasi")
    }
 
    student, err := s.students.FindByID(ctx, authStudent.StudentID)
    if err != nil {
        if !errors.Is(err, repository.ErrNotFound) { return helper.Internal(err) }
        return helper.Unauthorized("student tidak ditemukan")
    }
 
    return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", fiber.Map{
		"student":     student,
		"permissions": s.perms.PermissionsOf(student.Role),
	})
}

func (s *AuthService) issueTokenPair(
    ctx context.Context, student model.Student,
) (model.TokenPair, error) {
    accessToken, err := s.jwt.GenerateAccess(student)
    if err != nil {
        return model.TokenPair{}, err
    }
 
    refreshToken, err := helper.RandomToken(refreshTokenBytes)
    if err != nil {
        return model.TokenPair{}, err
    }

    err = s.tokens.Save(ctx, model.RefreshToken{
        StudentID:    student.ID,
        TokenHash: helper.SHA256Hex(refreshToken),
        ExpiresAt: time.Now().Add(s.refreshTTL),
    })
    if err != nil {
        return model.TokenPair{}, err
    }
 
    return model.TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        TokenType:    "Bearer",
        ExpiresIn:    int(s.jwt.AccessTTL().Seconds()),
    }, nil
}
