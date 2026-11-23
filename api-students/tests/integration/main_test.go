package main

package integration_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"api-students/app/model"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	_ = godotenv.Load(filepath.Join("..", "..", ".env.test"))

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn != "" {
		fmt.Println("TEST_DATABASE_URL tidak diisi: integration test dilewati")
		os.Exit(0)
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		fmt.Println("TEST_DATABASE_URL tidak valid:", err)
		os.Exit(1)
	}

	if !strings.HasSuffix(cfg.ConnConfig.Database, "_test") {
		fmt.Printf("menolak berjalan: database %q tidak berakhiran _test\n", cfg.ConnConfig.Database)
		os.Exit(1)
	}

	cfg.MaxConns = 1

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err == nil {
		err = pool.Ping(ctx)
	}
	if err != nil {
		fmt.Println("gagal terhubung ke database test:", err)
		os.Exit(1)
	}

	if err := applyMigrations(ctx); err != nil {
		fmt.Println("gagal menjalankan migration:", err)
		os.Exit(1)
	}

	code := m.Run()
	pool.Close()
	os.Exit(code)
}

func applyMigrations(ctx context.Context) error {
	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.sql"))
	if err != nil {
		return err
	}
	slices.Sort(files)

	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(f), err)
		}
	}
	return nil
}

func uniqueSuffix(t *testing.T) string {
	t.Helper()
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("membuat akhiran acak: %v", err)
	}
	return hex.EncodeToString(b)
}

// seedStudent membuat satu mahasiswa khusus untuk test yang memanggilnya.
//
// t.Cleanup mendaftarkan pembersihan yang dijalankan SETELAH test selesai
// — berhasil maupun gagal. Seluruh data turunannya (enrollments,
// notifications) ikut terhapus lewat ON DELETE CASCADE.
func seedStudent(t *testing.T) model.Student {
	t.Helper()
	ctx := context.Background()
	s := uniqueSuffix(t)
	u := model.Student{
		NIM:   "99" + s[:6],
		Name:  "it_" + s,
		Email: "it_" + s + "@student.unair.ac.id",
		Role:  "user",
	}

	err := pool.QueryRow(ctx,
		`INSERT INTO students (nim, name, email, password, role) VALUES ($1, $2, $3, '!', 'user') RETURNING id`,
		u.NIM, u.Name, u.Email).Scan(&u.ID)
	if err != nil {
		t.Fatalf("seed mahasiswa: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM outbox WHERE (payload->>'student_id')::INT = $1`, u.ID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM students WHERE id = $1`, u.ID)
	})
	return u
}

// seedCourse membuat satu kelas khusus untuk test yang memanggilnya.
func seedCourse(t *testing.T, credits, quota int) model.Course {
	t.Helper()
	ctx := context.Background()
	c := model.Course{Code: "IT" + uniqueSuffix(t), Name: "Kelas Uji", Credits: credits, Quota: quota}

	err := pool.QueryRow(ctx,
		`INSERT INTO courses (code, name, credits, quota) VALUES ($1, $2, $3, $4) RETURNING id`,
		c.Code, c.Name, c.Credits, c.Quota).Scan(&c.ID)
	if err != nil {
		t.Fatalf("seed kelas: %v", err)
	}

	defer func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM courses WHERE id = $1`, c.ID)
	}()
	return c
}