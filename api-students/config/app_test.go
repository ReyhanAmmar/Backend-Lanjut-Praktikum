package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"api-students/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func TestErrorHandlerStatusLogDanKerahasiaan(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	app := fiber.New(fiber.Config{ErrorHandler: newErrorHandler(logger)})
	app.Use(requestid.New())
	app.Get("/missing", func(*fiber.Ctx) error { return helper.NotFound("student tidak ditemukan") })
	app.Get("/broken", func(*fiber.Ctx) error { return helper.Internal(errors.New("rahasia: kolom database")) })
	for _, tc := range []struct{ path string; status int; code, level string }{
		{"/missing", 404, "NOT_FOUND", `"level":"WARN"`},
		{"/broken", 500, "INTERNAL_ERROR", `"level":"ERROR"`},
	} {
		resp, err := app.Test(httptest.NewRequest("GET", tc.path, nil))
		if err != nil { t.Fatal(err) }
		var body struct {
			Code string `json:"code"`
			Message string `json:"message"`
			RequestID string `json:"request_id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil { t.Fatal(err) }
		resp.Body.Close()
		if resp.StatusCode != tc.status || body.Code != tc.code || body.RequestID == "" {
			t.Fatalf("%s: status=%d body=%+v", tc.path, resp.StatusCode, body)
		}
		if strings.Contains(body.Message, "kolom database") || !strings.Contains(output.String(), tc.level) {
			t.Fatalf("pesan bocor atau log salah pada %s: %s", tc.path, output.String())
		}
	}
}
