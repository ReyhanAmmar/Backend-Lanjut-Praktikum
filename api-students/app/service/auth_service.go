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
    students      repository.StudentRepository
    tokens     repository.TokenRepository
    jwt        *helper.JWTManager
    refreshTTL time.Duration
}

func NewAuthService(
    students repository.StudentRepository,
    tokens repository.TokenRepository,
    jwtManager *helper.JWTManager,
    refreshTTL time.Duration,
) *AuthService {
    return &AuthService{
        students: students, tokens: tokens, jwt: jwtManager, refreshTTL: refreshTTL,
    }
}

func (s *AuthService) Register(c *fiber.Ctx) error {
    ctx, cancel := helper.RequestContext(c)
    defer cancel()
 
    var req model.RegisterRequest
    if err := c.BodyParser(&req); err != nil {
        return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
    }
 
    req.NIM = strings.TrimSpace(req.NIM)
    req.Name = strings.TrimSpace(req.Name)

    if errs := ValidateRegister(req); len(errs) > 0 {
        return helper.FailValidation(c, errs)
    }

	hashed, err := helper.HashPassword(req.Password)
    if err != nil {
        return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
    }

	created, err := s.students.Create(ctx, model.Student{
        NIM:      req.NIM,
        Name:     req.Name,
        Password: hashed,
        IsActive: true,
    })
    if err != nil {
        if errors.Is(err, repository.ErrDuplicate) {
            return helper.Fail(c, fiber.StatusConflict, "NIM sudah dipakai")
        }
        return helper.Fail(c, fiber.StatusInternalServerError, "gagal mendaftarkan student")
    }

	return helper.Created(c, "pendaftaran berhasil", created,
        "/api/v1/students/"+strconv.Itoa(created.ID))
}

func (s *AuthService) Login(c *fiber.Ctx) error {
    ctx, cancel := helper.RequestContext(c)
    defer cancel()
 
    var req model.LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
    }
 
    if errs := ValidateLogin(req); len(errs) > 0 {
        return helper.FailValidation(c, errs)
    }
 
    user, err := s.students.FindByNIM(ctx, strings.TrimSpace(req.NIM))
    if err != nil {
        helper.VerifyDummyPassword(req.Password)
        return helper.Fail(c, fiber.StatusUnauthorized, "NIM atau password salah")
    }
 
    if !helper.VerifyPassword(user.Password, req.Password) {
        return helper.Fail(c, fiber.StatusUnauthorized, "NIM atau password salah")
    }
 
    if !user.IsActive {
        return helper.Fail(c, fiber.StatusForbidden, "akun dinonaktifkan")
    }
 
    pair, err := s.issueTokenPair(ctx, user)
    if err != nil {
        return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token")
    }
 
    return helper.Success(c, fiber.StatusOK, "login berhasil", pair)
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
    ctx, cancel := helper.RequestContext(c)
    defer cancel()
 
    var req model.RefreshRequest
    if err := c.BodyParser(&req); err != nil {
        return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
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
        return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
    }
 
    student, err := s.students.FindByID(ctx, authStudent.StudentID)
    if err != nil {
        return helper.Fail(c, fiber.StatusUnauthorized, "student tidak ditemukan")
    }
 
    return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", student)
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