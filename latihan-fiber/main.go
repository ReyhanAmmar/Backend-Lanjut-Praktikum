package main

import "fmt"

type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

func (s Student) GetInfo() string {
	return fmt.Sprintf("ID: %d | Nama: %s | Nilai: %.1f | Aktif: %v",
		s.ID, s.Name, s.Grade, s.IsActive)
}

func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	student := Student{
		ID:       1,
		Name:     "Budiono Siregar",
		Grade:    0,
		IsActive: false,
	}

	fmt.Println("=== Data awal ===")
	fmt.Println(student.GetInfo())

	student.Activate()
	student.UpdateGrade(88.5)

	fmt.Println("\n=== Setelah diaktifkan dan nilai diperbarui ===")
	fmt.Println(student.GetInfo())

	student.Deactivate()

	fmt.Println("\n=== Setelah dinonaktifkan ===")
	fmt.Println(student.GetInfo())
}