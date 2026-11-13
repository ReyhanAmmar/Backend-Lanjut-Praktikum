package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
	"api-students/config"
)

func main() {
	outboxID := flag.Int64("outbox", 0, "ID baris outbox yang ingin dikirim ulang")
	copies := flag.Int("copies", 1, "Jumlah salinan task yang dikirim ke Redis")
	flag.Parse()

	if *outboxID <= 0 {
		fmt.Println("Gunakan: go run ./tools/redeliver -outbox <id> -copies <n>")
		os.Exit(1)
	}

	config.LoadEnv()
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.GetEnv("DB_USER", "postgres"),
		config.GetEnv("DB_PASSWORD", "123"),
		config.GetEnv("DB_HOST", "localhost"),
		config.GetEnv("DB_PORT", "5432"),
		config.GetEnv("DB_NAME", "students"),
		config.GetEnv("DB_SSLMODE", "disable"),
	)

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Koneksi DB gagal: %v", err)
	}
	defer pool.Close()

	var eventType string
	var payload []byte
	err = pool.QueryRow(context.Background(),
		"SELECT event_type, payload FROM outbox WHERE id = $1", *outboxID).Scan(&eventType, &payload)
	if err != nil {
		log.Fatalf("Ambil data outbox %d gagal: %v", *outboxID, err)
	}

	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()

	// Jika ini event perubahan kuota yang punya StudentIDs
	if eventType == model.EventCourseQuotaChanged {
		var event model.CourseQuotaChangedEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Fatalf("Unmarshal payload quota gagal: %v", err)
		}

		for _, studentID := range event.StudentIDs {
			job := model.NotificationJob{
				OutboxID:  *outboxID,
				EventType: eventType,
				Payload:   payload,
				StudentID: studentID,
			}
			jobBytes, _ := json.Marshal(job)

			for c := 1; c <= *copies; c++ {
				task := asynq.NewTask(
					model.TaskSendNotification,
					jobBytes,
					asynq.Queue(model.QueueNotifications),
					asynq.TaskID(fmt.Sprintf("redeliver-outbox-%d-student-%d-copy-%d", *outboxID, studentID, c)),
				)
				if _, err := client.Enqueue(task); err != nil {
					log.Printf("Enqueue copy %d untuk student %d gagal: %v", c, studentID, err)
				} else {
					fmt.Printf("Task student %d copy %d berhasil di-enqueue\n", studentID, c)
				}
			}
		}
		return
	}

	// Untuk event reguler (1 mahasiswa)
	job := model.NotificationJob{
		OutboxID:  *outboxID,
		EventType: eventType,
		Payload:   payload,
	}
	jobBytes, _ := json.Marshal(job)

	for c := 1; c <= *copies; c++ {
		task := asynq.NewTask(
			model.TaskSendNotification,
			jobBytes,
			asynq.Queue(model.QueueNotifications),
			asynq.TaskID(fmt.Sprintf("redeliver-outbox-%d-copy-%d", *outboxID, c)),
		)
		if _, err := client.Enqueue(task); err != nil {
			log.Printf("Enqueue copy %d gagal: %v", c, err)
		} else {
			fmt.Printf("Task outbox %d copy %d berhasil di-enqueue\n", *outboxID, c)
		}
	}
}

