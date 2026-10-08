package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil { return err }
	q, err := helper.ParseCursorQuery(c)
	if err != nil { return err }
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	students, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}
	hasMore := len(students) > q.Limit
	if hasMore { students = students[:q.Limit] }
	meta := &model.CursorMeta{Limit:q.Limit, HasMore:hasMore}
	if hasMore { last := students[len(students)-1]; meta.NextCursor = helper.EncodeCursor(last.CreatedAt,last.ID) }
	if format == helper.FormatCSV { return helper.WriteStudentsCSV(c, students) }
	return helper.SuccessCursor(c, "daftar student berhasil diambil", students, meta)
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentStudent(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	
	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "gagal mengambil data student")
	}

	if !CanAccessStudent(current,student.OwnerID,s.perms,"student:read:any") {
		return helper.Forbidden("tidak berhak mengakses data student lain")
	}

	return helper.Success(c, fiber.StatusOK, "student ditemukan", student)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentStudent(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	newStudent, err := s.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    *req.Grade,
		IsActive: true,
		OwnerID:  current.StudentID,
	})
	if err != nil {
		return translateError(err, "gagal menyimpan student")
	}

	return helper.Created(c, "student berhasil dibuat", newStudent,
		"/api/v1/students/"+strconv.Itoa(newStudent.ID))
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentStudent(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "gagal mengambil data student")
	}

	if !CanAccessStudent(current, student.OwnerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data student lain")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	hasil, err := s.repo.Update(ctx, model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     strings.TrimSpace(req.Name),
		Grade:    req.Grade,
		IsActive: req.IsActive,
	})
	if err != nil {
		return translateError(err, "gagal memperbarui student")
	}

	return helper.Success(c, fiber.StatusOK, "student berhasil diganti seluruhnya", hasil)
}

func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentStudent(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "gagal mengambil data student")
	}

	if !CanAccessStudent(current, student.OwnerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data student lain")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}
	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}
	updated := ApplyPatch(student, req)

	updated.OwnerID = student.OwnerID

	hasil, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateError(err, "gagal memperbarui student")
	}

	return helper.Success(c, fiber.StatusOK, "student berhasil diperbarui sebagian", hasil)
}

func (s *StudentService) AssignRole(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentStudent(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := ValidateAssignRole(current, id, req, s.perms); len(errs) > 0 {
		return helper.Validation(errs)
	}

	result, err := s.repo.UpdateRole(ctx, id, strings.TrimSpace(req.Role))
	if err != nil {
		return translateError(err, "gagal mengubah role student")
	}

	return helper.Success(c, fiber.StatusOK, "role student berhasil diubah", result)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentStudent(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	// Punya permission menghapus tidak berarti boleh menghapus akun sendiri
	if current.StudentID == id {
		return helper.Forbidden("tidak boleh menghapus akun sendiri")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(err, "gagal menghapus student")
	}

	return helper.NoContent(c)
}

func translateError(err error, pesanUmum string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("NIM sudah dipakai")
	default:
		return helper.Internal(err)
	}
}
