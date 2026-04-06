package server

import (
	"database/sql"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/Robi9/church-manager/internal/config"
	"github.com/Robi9/church-manager/internal/middleware"
	"github.com/Robi9/church-manager/internal/modules/auth"
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

	api := r.Group("/api")
	{
		members := api.Group("/members")
		members.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			members.POST("", handler.Create)
			members.GET("", handler.Find)
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
