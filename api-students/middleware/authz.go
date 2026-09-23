package middleware
 
import (
    "github.com/gofiber/fiber/v2"
 
    "api-students/helper"
)

func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        student, ok := helper.CurrentStudent(c)
        if !ok {
            return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
        }
 
        if !perms.Can(student.Role, permission) {
            return helper.Fail(c, fiber.StatusForbidden,
                "role "+student.Role+" tidak memiliki hak "+permission)
        }
 
        return c.Next()
    }
}

func RequireRole(roles ...string) fiber.Handler {
    allowed := make(map[string]struct{}, len(roles))
    for _, role := range roles {
        allowed[role] = struct{}{}
    }
 
    return func(c *fiber.Ctx) error {
        student, ok := helper.CurrentStudent(c)
        if !ok {
            return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
        }
 
        if _, granted := allowed[student.Role]; !granted {
            return helper.Fail(c, fiber.StatusForbidden,
                "role Anda tidak berhak mengakses endpoint ini")
        }
 
        return c.Next()
    }
}
