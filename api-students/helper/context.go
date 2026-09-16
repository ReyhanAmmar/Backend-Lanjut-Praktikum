package helper

import "time"

import (
    "github.com/gofiber/fiber/v2"
 
    "api-students/app/model"
)

const LocalsAuthStudent = "authStudent"
 
func CurrentStudent(c *fiber.Ctx) (model.AuthStudent, bool) {
    student, ok := c.Locals(LocalsAuthStudent).(model.AuthStudent)
    return student, ok
}
