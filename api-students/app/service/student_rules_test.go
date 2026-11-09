package service

import (
	"testing"

	"api-students/app/model"
)

func TestApplyPatch(t *testing.T) {
	initial := model.Student{ID: 1, NIM: "123456", Name: "Sari Dewi", Grade: 80, IsActive: true}
	inactive := false

	result := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})
	if result.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if result.Name != "Sari Dewi" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestIsEmptyPatch(t *testing.T) {
	if !IsEmptyPatch(model.PatchStudentRequest{}) {
		t.Error("request tanpa field apa pun seharusnya dianggap kosong")
	}

	grade := 90.0
	if IsEmptyPatch(model.PatchStudentRequest{Grade: &grade}) {
		t.Error("request berisi grade seharusnya tidak dianggap kosong")
	}
}
