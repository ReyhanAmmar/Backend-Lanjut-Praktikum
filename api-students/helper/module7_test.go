package helper

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"api-students/app/model"
	"github.com/gofiber/fiber/v2"
)

func TestValidationStudentDanPassword(t *testing.T) {
	grade := 82.0
	valid := model.CreateStudentRequest{NIM: "434241061", Name: "Reyhan Ammar", Grade: &grade}
	if fields := ValidateStruct(valid); fields != nil {
		t.Fatalf("student valid ditolak: %v", fields)
	}
	bad := model.CreateStudentRequest{NIM: "abc", Name: "", Grade: &grade}
	fields := ValidateStruct(bad)
	if fields["nim"] == "" || fields["name"] == "" {
		t.Fatalf("pelanggaran nim dan name tidak lengkap: %v", fields)
	}
	if PasswordStrength("Kuat12345") != "" || PasswordStrength("password123") != "password terlalu umum" {
		t.Fatal("aturan password tidak sesuai")
	}
	if got := Validation(fields).Status; got != fiber.StatusUnprocessableEntity {
		t.Fatalf("status validasi = %d, ingin 422", got)
	}
	empty := ""
	if ValidateStruct(model.PatchStudentRequest{Name: &empty})["name"] == "" {
		t.Fatal("PATCH name kosong diterima")
	}
	if fields := ValidateStruct(model.PatchStudentRequest{Grade: &grade}); fields != nil {
		t.Fatalf("PATCH tanpa name ditolak: %v", fields)
	}
}

func TestCursorDanCSV(t *testing.T) {
	stamp := time.Date(2026, 9, 13, 14, 12, 40, 0, time.UTC)
	cursor := EncodeCursor(stamp, 8)
	decoded, err := DecodeCursor(cursor)
	if err != nil || decoded.ID != 8 || !decoded.CreatedAt.Equal(stamp) {
		t.Fatalf("cursor tidak kembali utuh: %+v, %v", decoded, err)
	}
	if _, err := DecodeCursor("bukanbase64!!"); err == nil {
		t.Fatal("cursor rusak diterima")
	}

	app := fiber.New()
	app.Get("/students", func(c *fiber.Ctx) error {
		return WriteStudentsCSV(c, []model.Student{{ID: 8, NIM: "434241061", Name: "A, B", Grade: 82, IsActive: true, CreatedAt: stamp}})
	})
	resp, err := app.Test(httptest.NewRequest("GET", "/students", nil))
	if err != nil { t.Fatal(err) }
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil { t.Fatal(err) }
	if resp.StatusCode != 200 || !strings.Contains(string(body), "id,nim,name,grade,is_active,created_at") ||
		!strings.Contains(string(body), `"A, B"`) || !strings.Contains(string(body), "434241061") {
		t.Fatalf("CSV kosong atau rusak: status=%d body=%q", resp.StatusCode, body)
	}
}
