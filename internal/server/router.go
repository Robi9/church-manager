package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"github.com/Robi9/church-manager/internal/middleware"
	"github.com/Robi9/church-manager/internal/modules/auth"
	"github.com/Robi9/church-manager/internal/modules/member"
)

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// 🧱 Layers assembly
	repo := member.NewRepository(db)
	service := member.NewService(repo)
	handler := member.NewHandler(service)

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)

	api := r.Group("/api")
	{
		members := api.Group("/members")
		members.Use(middleware.AuthMiddleware())
		{
			members.POST("", handler.Create)
			members.GET("", handler.Find)
		}
	}

	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
	}

	return r
}
