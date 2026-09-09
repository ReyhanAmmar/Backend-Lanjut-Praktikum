package route

import (
	"context"
	"time"
 
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
 
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
)

func Register(app *fiber.App, pool *pgxpool.Pool, studentService *service.StudentService) {
	api := app.Group("/api/v1")
 
	api.Get("/health", healthCheck(pool))
 
	students := api.Group("/students", middleware.RequireJSON)
	students.Get("/", studentService.List)
	students.Get("/:id", studentService.Get)
	students.Post("/", studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", studentService.Delete)
}

