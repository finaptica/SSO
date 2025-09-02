package main

import (
	"errors"
	"flag"
	"fmt"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// "host=localhost port=5432 user=fin_admin password=12345678FinAdmin dbname=finaptica sslmode=disable"
func main() {
	var migrationsPath string

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.Parse()

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	db, _ := sql.Open("postgres", "postgres://fin_admin:12345678FinAdmin@localhost:5432/finaptica?sslmode=enable")
	driver, _ := postgres.WithInstance(db, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	m.Up()
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}

	fmt.Println("migrations applied successfully")
}
