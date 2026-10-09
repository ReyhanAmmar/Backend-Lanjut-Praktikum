package middleware

import (
	"bytes"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"api-students/helper"
	"github.com/gofiber/fiber/v2"
)

func TestRequestLoggerMencatatStatusError(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(err.(*helper.AppError).Status).SendString("ditolak")
	}})
	app.Use(RequestLogger(logger))
	app.Get("/students/999", func(*fiber.Ctx) error { return helper.NotFound("student tidak ditemukan") })
	resp, err := app.Test(httptest.NewRequest("GET", "/students/999", nil))
	if err != nil { t.Fatal(err) }
	resp.Body.Close()
	if resp.StatusCode != 404 || !strings.Contains(output.String(), `"status":404`) {
		t.Fatalf("status respons/log berbeda: %d %s", resp.StatusCode, output.String())
	}
}
