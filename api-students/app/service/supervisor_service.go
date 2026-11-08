package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type SupervisorService struct {
	supervisors repository.SupervisorRepository
	tx          *repository.TxManager
}

func NewSupervisorService(supervisors repository.SupervisorRepository, tx *repository.TxManager) *SupervisorService {
	return &SupervisorService{supervisors: supervisors, tx: tx}
}

func (s *SupervisorService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	list, err := s.supervisors.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar dosen pembimbing berhasil diambil", list)
}

func (s *SupervisorService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	sp, err := s.supervisors.FindByID(ctx, id)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "dosen pembimbing ditemukan", sp)
}

func (s *SupervisorService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id dosen harus berupa angka positif")
	}

	var req model.AssignSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var updated model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		var err error
		updated, err = repo.IncrementSupervised(ctx, supervisorID)
		if err != nil {
			return err
		}

		return repo.InsertSupervision(ctx, req.StudentID, supervisorID)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Created(c, "berhasil menetapkan dosen pembimbing", updated,
		"/api/v1/supervisors/"+strconv.Itoa(supervisorID))
}

func (s *SupervisorService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.SwitchSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}
	if req.FromSupervisorID == req.ToSupervisorID {
		return helper.Validation(map[string]string{
			"to_supervisor_id": "tidak boleh sama dengan from_supervisor_id",
		})
	}

	var newSupervisor model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}
		if err := repo.LockSupervisors(ctx, req.FromSupervisorID, req.ToSupervisorID); err != nil {
			return err
		}
		if err := repo.DeleteSupervision(ctx, req.StudentID, req.FromSupervisorID); err != nil {
			return err
		}
		if _, err := repo.DecrementSupervised(ctx, req.FromSupervisorID); err != nil {
			return err
		}

		var err error
		newSupervisor, err = repo.IncrementSupervised(ctx, req.ToSupervisorID)
		if err != nil {
			return err
		}

		return repo.InsertSupervision(ctx, req.StudentID, req.ToSupervisorID)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "berhasil mengganti dosen pembimbing", newSupervisor)
}

func (s *SupervisorService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateSupervisorQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.supervisors.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "kuota pembimbing berhasil diubah", updated)
}

func translateSupervisorError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data pembimbing atau mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrSupervisorFull):
		return helper.NewError(fiber.StatusConflict, "SUPERVISOR_FULL", "kuota dosen pembimbing sudah penuh")
	case errors.Is(err, repository.ErrAlreadySupervised):
		return helper.NewError(fiber.StatusConflict, "ALREADY_SUPERVISED", "mahasiswa sudah memiliki pembimbing tugas akhir")
	case errors.Is(err, repository.ErrNotSupervised):
		return helper.NewError(fiber.StatusConflict, "NOT_SUPERVISED", "mahasiswa tidak terdaftar pada dosen pembimbing tersebut")
	case errors.Is(err, repository.ErrSupervisorVersionConflict):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT", "data pembimbing sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrSupervisorQuotaBelowFilled):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_SUPERVISED", "kuota tidak boleh lebih kecil dari jumlah mahasiswa yang sedang dibimbing")
	default:
		return helper.Internal(err)
	}
}