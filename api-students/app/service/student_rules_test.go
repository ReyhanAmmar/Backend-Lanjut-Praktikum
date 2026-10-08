package service

import (
    "testing"
    "api-students/app/model"
    "api-students/helper"
)

func TestStudentValidation(t *testing.T) {
    grade := 88.0
    valid := model.CreateStudentRequest{NIM:"123456", Name:"Sari Dewi", Grade:&grade}
    if errors := helper.ValidateStruct(valid); len(errors) != 0 { t.Fatalf("valid: %v",errors) }
    badGrade := -10.0
    bad := model.CreateStudentRequest{NIM:"",Name:"",Grade:&badGrade}
    errors := helper.ValidateStruct(bad)
    for _, field := range []string{"nim","name","grade"} {
        if errors[field] == "" { t.Errorf("%s not rejected: %v",field,errors) }
    }
    empty := ""
    if helper.ValidateStruct(model.PatchStudentRequest{Name:&empty})["name"] == "" { t.Fatal("empty name accepted") }
    if errors := helper.ValidateStruct(model.PatchStudentRequest{IsActive:new(bool)}); errors != nil { t.Fatalf("omitted name rejected: %v", errors) }
}

func TestApplyPatch(t *testing.T) {
    initial := model.Student{ID:1,NIM:"123456",Name:"Sari Dewi",Grade:80,IsActive:true}
    inactive := false
    result := ApplyPatch(initial,model.PatchStudentRequest{IsActive:&inactive})
    if result.IsActive || result.Name != initial.Name { t.Fatalf("unexpected patch: %+v",result) }
    if !IsEmptyPatch(model.PatchStudentRequest{}) { t.Fatal("empty patch accepted") }
}
