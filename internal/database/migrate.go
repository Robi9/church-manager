package database

import (
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL string) {
	log.Println("running migrations...")
	m, err := migrate.New(
		"file://./migrations",
		databaseURL,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		sourceErr, databaseErr := m.Close()
		if sourceErr != nil {
			log.Printf("closing migration source: %v", sourceErr)
		}
		if databaseErr != nil {
			log.Printf("closing migration database: %v", databaseErr)
		}
	}()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}

	log.Println("migrations applied successfully")
}
