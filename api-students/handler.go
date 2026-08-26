package main

import (
    "sort"
    "strconv"
    "strings"
    "time"
 
    "github.com/gofiber/fiber/v2"
)

var students []Student
var nextStudentID = 1

func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}
	return -1
}

func cocokPencarian(s Student, kata string) bool {
    kata = strings.ToLower(kata)
    return strings.Contains(strings.ToLower(s.Name), kata) ||
        strings.Contains(strings.ToLower(s.NIM), kata)
}

func paramStudentID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// GET /students
func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)

	result := []Student{}
	for _, s := range students {
		if q.IsActive != nil && s.IsActive != *q.IsActive {
			continue	
		}
		if q.Search != "" && !cocokPencarian(s, q.Search) {
            continue
        }
		result = append(result, s)
	}

	sort.SliceStable(result, func(i, j int) bool {
		var less bool
		switch q.Sort {
			case "id":
				less = result[i].ID < result[j].ID
			case "nim":
				less = result[i].NIM < result[j].NIM
			case "name":
				less = result[i].Name < result[j].Name
			case "grade":
				less = result[i].Grade < result[j].Grade
			case "created_at":
				less = result[i].CreatedAt.Before(result[j].CreatedAt)
			default:
				less = result[i].ID < result[j].ID
		}
		if q.Order == "desc" {
			return !less
		}
		return less
	})

	total :=len(result)
	totalPages := (total + q.Limit - 1) / q.Limit
    mulai := (q.Page - 1) * q.Limit
    if mulai > total {
        mulai = total
    }
    akhir := mulai + q.Limit
    if akhir > total {
        akhir = total
    }
 
    return okList(c, "daftar mahasiswa berhasil diambil", result[mulai:akhir], &Meta{
        Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
    })
}

// GET /students/:id
func getStudent(c *fiber.Ctx) error {
	id, valid := paramStudentID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka bulat positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	return ok(c, "mahasiswa ditemukan", students[i])
}

// POST
func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := map[string]string{}
	if req.NIM == "" {
		errs["nim"] = "NIM wajib diisi"
	}
	if req.Name == "" {
		errs["name"] = "nama mahasiswa wajib diisi"
	}
	if req.Grade == nil {
		errs["grade"] = "nilai (grade) wajib diisi"
	} else if *req.Grade < 0.0 || *req.Grade > 100.0 {
		errs["grade"] = "nilai (grade) harus berada di antara 0.00 dan 100.00"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}
	for _, s := range students {
		if strings.EqualFold(s.NIM, req.NIM) {
			return fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
		}
	}

	newStudent := Student{
		ID:        nextStudentID,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     *req.Grade,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	students = append(students, newStudent)
	nextStudentID++

	return created(c, "mahasiswa berhasil didaftarkan", newStudent,
		"/api/v1/students/"+strconv.Itoa(newStudent.ID))
}

//PUT
func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramStudentID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka bulat positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := map[string]string{}
	if req.NIM == "" {
		errs["nim"] = "NIM wajib diisi pada pembaruan PUT"
	}
	if req.Name == "" {
		errs["name"] = "nama wajib diisi pada pembaruan PUT"
	}
	if req.Grade < 0.0 || req.Grade > 100.0 {
		errs["grade"] = "nilai (grade) harus berada di rentang 0.00 - 100.00"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	students[i].NIM = req.NIM
	students[i].Name = req.Name
	students[i].Grade = req.Grade
	students[i].IsActive = req.IsActive

	return ok(c, "data mahasiswa berhasil diganti seluruhnya", students[i])
}

// PATCH
func patchStudent(c *fiber.Ctx) error {
	id, valid := paramStudentID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka bulat positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM != nil {
        if strings.TrimSpace(*req.NIM) == "" {
            return failValidation(c, map[string]string{"nim": "NIM sudah digunakan oleh mahasiswa lain"})
        }
        students[i].NIM = *req.NIM
	}
	if req.Name != nil {
        if !strings.Contains(*req.Name, "@") {
            return failValidation(c, map[string]string{"name": "nama tidak boleh dikosongkan"})
        }
        students[i].Name = *req.Name
    }
	if req.Grade != nil {
        if *req.Grade < 0.0 || *req.Grade > 100.0 {
            return failValidation(c, map[string]string{"grade": "nilai harus berada di rentang 0.00 - 100.00"})
        }
        students[i].Grade = *req.Grade
    }
    if req.IsActive != nil {
        students[i].IsActive = *req.IsActive
    }

	return ok(c, "data mahasiswa berhasil diperbarui sebagian", students[i])
}

// DELETE
func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramStudentID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka bulat positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	students = append(students[:i], students[i+1:]...)

	return noContent(c)
}
	