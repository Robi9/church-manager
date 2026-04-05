package main

import (
	"github.com/Robi9/church-manager/internal/config"
	"github.com/Robi9/church-manager/internal/database"
	"github.com/Robi9/church-manager/internal/server"
)

func main() {
	cfg := config.Load()

	database.RunMigrations(cfg.DatabaseURL)

	db := database.NewConnection(cfg.DatabaseURL)

	r := server.SetupRouter(db)
	r.Run(":" + cfg.Port)
}
