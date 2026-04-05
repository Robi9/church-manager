package main

import (
	"fmt"

	"github.com/Robi9/church-manager/internal/config"
	"github.com/Robi9/church-manager/internal/database"
	"github.com/Robi9/church-manager/internal/server"
	"golang.org/x/crypto/bcrypt"
)

func main() {

	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	fmt.Println(string(hash))

	cfg := config.Load()

	database.RunMigrations(cfg.DatabaseURL)

	db := database.NewConnection(cfg.DatabaseURL)

	r := server.SetupRouter(db)
	r.Run(":" + cfg.Port)
}
