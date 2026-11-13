package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type NotificationWorker struct {
	students      repository.StudentRepository
	notifications repository.NotificationRepository
	tx            *repository.TxManager
	mailer        helper.Mailer
	logger        *slog.Logger
}

func NewNotificationWorker(
	students repository.StudentRepository,
	notifications repository.NotificationRepository,
	tx *repository.TxManager,
	mailer helper.Mailer,
	logger *slog.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		students:      students,
		notifications: notifications,
		tx:            tx,
		mailer:        mailer,
		logger:        logger,
	}
}

func (w *NotificationWorker) HandleSend(ctx context.Context, t *asynq.Task) error {
	var job model.NotificationJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return fmt.Errorf("%w: payload task rusak: %v", asynq.SkipRetry, err)
	}

	var targetStudentID int
	var msg model.Message

	if job.EventType == model.EventCourseQuotaChanged {
		var event model.CourseQuotaChangedEvent
		if err := json.Unmarshal(job.Payload, &event); err != nil {
			return fmt.Errorf("%w: payload quota event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
		}
		targetStudentID = job.StudentID
		student, err := w.students.FindByID(ctx, targetStudentID)
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: mahasiswa %d tidak ditemukan", asynq.SkipRetry, targetStudentID)
		}
		if err != nil {
			return err
		}
		msg = BuildQuotaChangedMessage(event, student)
	} else {
		var event model.EnrollmentEvent
		if err := json.Unmarshal(job.Payload, &event); err != nil {
			return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
		}
		targetStudentID = event.StudentID
		student, err := w.students.FindByID(ctx, targetStudentID)
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: mahasiswa %d tidak ditemukan", asynq.SkipRetry, targetStudentID)
		}
		if err != nil {
			return err
		}
		var errBuild error
		msg, errBuild = BuildMessage(job.EventType, event, student)
		if errBuild != nil {
			return fmt.Errorf("%w: %v", asynq.SkipRetry, errBuild)
		}
	}

	alreadySent := false
	err := w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		claimed, err := w.notifications.Claim(ctx, model.Notification{
			OutboxID:  job.OutboxID,
			StudentID: targetStudentID,
			Channel:   "email",
			Subject:   msg.Subject,
			Body:      msg.Body,
		})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return nil
		}
		return w.mailer.Send(ctx, msg)
	})

	if errors.Is(err, helper.ErrPermanentDelivery) {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if err != nil {
		return err
	}

	retry, _ := asynq.GetRetryCount(ctx)
	w.logger.Info("notification_processed",
		slog.Int64("outbox_id", job.OutboxID),
		slog.String("event", job.EventType),
		slog.Int("student_id", targetStudentID),
		slog.Int("retry", retry),
		slog.Bool("already_sent", alreadySent))
	return nil
}