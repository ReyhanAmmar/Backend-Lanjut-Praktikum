package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
)

type Dependencies struct {
	Pool               *pgxpool.Pool
	Mongo              *mongo.Client
	JWT                *helper.JWTManager
	Permissions        *helper.PermissionSet
	StudentService     *service.StudentService
	AuthService        *service.AuthService
	CourseService      *service.CourseService
	SupervisorService  *service.SupervisorService
	PortfolioService   *service.PortfolioService 
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool, deps.Mongo))

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	users := api.Group("/students",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	perms := deps.Permissions

	users.Get("/", middleware.RequirePermission(perms, "student:list"), deps.StudentService.List)
	users.Post("/", middleware.RequirePermission(perms, "student:create"), deps.StudentService.Create)
	users.Patch("/:id/role", middleware.RequirePermission(perms, "role:assign"), deps.StudentService.AssignRole)
	users.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), deps.StudentService.Delete)

	users.Get("/:id", deps.StudentService.Get)
	users.Put("/:id", deps.StudentService.Replace)
	users.Patch("/:id", deps.StudentService.Patch)

	courses := api.Group("/courses",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	
	courses.Get("/", deps.CourseService.List)
	courses.Get("/mine", deps.CourseService.Mine)
	courses.Get("/:id", deps.CourseService.Get)
	courses.Post("/:id/enroll", deps.CourseService.Enroll)
	courses.Delete("/:id/enroll", deps.CourseService.Drop)
	courses.Patch("/:id", middleware.RequirePermission(perms, "course:manage"), deps.CourseService.UpdateQuota)

	enrollments := api.Group("/enrollments",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	enrollments.Post("/swap", deps.CourseService.Swap)

	supervisors := api.Group("/supervisors",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	supervisors.Get("/", deps.SupervisorService.List)
	supervisors.Get("/:id", deps.SupervisorService.Get)
	supervisors.Post("/:id/assign", middleware.RequirePermission(perms, "supervision:manage"), deps.SupervisorService.Assign)
	supervisors.Patch("/:id", middleware.RequirePermission(perms, "supervision:manage"), deps.SupervisorService.UpdateQuota)

	supervisions := api.Group("/supervisions",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
		
	supervisions.Post("/switch", middleware.RequirePermission(perms, "supervision:manage"), deps.SupervisorService.Switch)

	studentPortfolios := api.Group("/students/:id/portfolio",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	studentPortfolios.Get("/", deps.PortfolioService.Get)
	studentPortfolios.Put("/", deps.PortfolioService.Upsert)
	studentPortfolios.Post("/projects", deps.PortfolioService.AddProject)
	studentPortfolios.Delete("/projects/:projectId", deps.PortfolioService.DeleteProject)

	api.Get("/portfolios/stats/technologies", deps.PortfolioService.TopTechnologies)
}

func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi", err)
		}
		if mongoClient != nil {
			if err := mongoClient.Ping(ctx, nil); err != nil {
				return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi", err)
			}
		}
		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", nil)
	}
}
