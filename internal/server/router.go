package server

import (
	"database/sql"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/Robi9/church-manager/internal/config"
	"github.com/Robi9/church-manager/internal/middleware"
	"github.com/Robi9/church-manager/internal/modules/auth"
	"github.com/Robi9/church-manager/internal/modules/dashboard"
	"github.com/Robi9/church-manager/internal/modules/member"
)

func SetupRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	repo := member.NewRepository(db)
	service := member.NewService(repo)
	handler := member.NewHandler(service)

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService)

	dashboardRepo := dashboard.NewRepository(db)
	dashboardService := dashboard.NewService(dashboardRepo)
	dashboardHandler := dashboard.NewHandler(dashboardService)

	api := r.Group("/api")
	{
		members := api.Group("/members")
		members.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			members.POST("", handler.Create)
			members.POST("/check-duplicates", handler.CheckDuplicates)
			members.GET("", handler.Find)
			members.GET("/:id", handler.GetByID)
			members.PUT("/:id", handler.Update)
			members.DELETE("/:id", handler.Delete)
			members.POST("/import", handler.Import)
			members.GET("/import/template", handler.DownloadTemplate)
			members.GET(
				"/import/errors/:jobID",
				handler.DownloadImportErrors,
			)
		}

		dashboardGroup := api.Group("/dashboard")
		dashboardGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			dashboardGroup.GET("/stats", dashboardHandler.GetStats)
		}
	}

	loginLimiter := middleware.NewRateLimiter(5, time.Minute)

	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", loginLimiter.Middleware(), authHandler.Login)
	}

	return r
}
