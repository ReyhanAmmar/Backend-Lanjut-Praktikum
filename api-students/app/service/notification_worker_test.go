package service

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/hibiken/asynq"

	"api-students/app/model"
	"api-students/app/repository"
)

type fakeStudents struct {
	repository.StudentRepository
	students map[int]model.Student
}

func (f *fakeStudents) FindByID(_ context.Context, id int) (model.Student, error) {
	s, ok := f.students[id]
	if !ok {
		return model.Student{}, repository.ErrNotFound
	}
	return s, nil
}

func newTask(t *testing.T, eventType string, event any) *asynq.Task {
	t.Helper()
	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("menyusun payload event: %v", err)
	}
	job, err := json.Marshal(model.NotificationJob{OutboxID: 1, EventType: eventType, Payload: payload})
	if err != nil {
		t.Fatalf("menyusun job: %v", err)
	}
	return asynq.NewTask(model.TaskSendNotification, job)
}

func TestHandleSend_PermanentFailures(t *testing.T) {
	worker := NewNotificationWorker(
		&fakeStudents{students: map[int]model.Student{7: {ID: 7, Name: "mhsuji", Email: "mhsuji@student.unair.ac.id"}}},
		nil, nil, nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	tests := []struct {
		name string
		task *asynq.Task
	}{
		{name: "payload task bukan JSON", task: asynq.NewTask(model.TaskSendNotification, []byte("bukan json"))},
		{name: "user sudah dihapus", task: newTask(t, model.EventEnrollmentCreated, model.EnrollmentEvent{StudentID: 99})},
		{name: "jenis event tidak dikenal", task: newTask(t, "enrollment.unknown", model.EnrollmentEvent{StudentID: 7})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := worker.HandleSend(context.Background(), tt.task)
			if err != asynq.SkipRetry {
				t.Errorf("error = %v, seharusnya membungkus asynq.SkipRetry", err)
			}
		})
	}
}