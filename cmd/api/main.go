package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Robi9/church-manager/internal/config"
	"github.com/Robi9/church-manager/internal/database"
	"github.com/Robi9/church-manager/internal/modules/auth"
	"github.com/Robi9/church-manager/internal/server"
)

func main() {
	cfg := config.Load()

	database.RunMigrations(cfg.DatabaseURL)

	db := database.NewConnection(cfg.DatabaseURL)
	defer db.Close()

	authService := auth.NewService(auth.NewRepository(db), cfg.JWTSecret)
	if err := authService.EnsureInitialAdmin(cfg.InitialAdminEmail, cfg.InitialAdminPassword); err != nil {
		log.Fatal(err)
	}

	r := server.SetupRouter(db, cfg)
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("API listening on port %s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
