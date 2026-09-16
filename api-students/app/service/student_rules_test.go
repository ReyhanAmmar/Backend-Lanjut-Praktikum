package service

import (
	"testing"

	"api-students/app/model"
)

func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
	}
	
	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d",
				tc.total, tc.limit, tc.want, got)
		}
	}
}

func TestValidateCreate(t *testing.T) {
	grade := 88.0
	valid := model.CreateStudentRequest{NIM: "123456", Name: "Sari Dewi", Grade: &grade}
	if errs := ValidateCreate(valid); len(errs) != 0 {
		t.Errorf("input valid seharusnya tidak menghasilkan error, dapat: %v", errs)
	}

	invalidGrade := -10.0
	invalid := model.CreateStudentRequest{NIM: "", Name: "", Grade: &invalidGrade}
	errs := ValidateCreate(invalid)
	if errs["nim"] == "" {
		t.Error("nim kosong seharusnya menghasilkan error")
	}
	if errs["name"] == "" {
		t.Error("name kosong seharusnya menghasilkan error")
	}
	if errs["grade"] == "" {
		t.Error("grade di luar rentang 0-100 seharusnya menghasilkan error")
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{ID: 1, NIM: "123456", Name: "Sari Dewi", Grade: 80, IsActive: true}
	inactive := false

	result, errs := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})
	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}
	if result.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if result.Name != "Sari Dewi" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}

	wrongGrade := 200.0
	_, errs2 := ApplyPatch(initial, model.PatchStudentRequest{Grade: &wrongGrade})
	if errs2["grade"] == "" {
		t.Error("grade di luar rentang seharusnya ditolak walau lewat PATCH")
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