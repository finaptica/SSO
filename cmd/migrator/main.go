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

func main() {
	var migrationsPath, dbUser, dbPassword, dbHost, dbPort, dbName string

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&dbUser, "db-user", "", "database user")
	flag.StringVar(&dbPassword, "db-password", "", "database password")
	flag.StringVar(&dbHost, "db-host", "", "database host")
	flag.StringVar(&dbPort, "db-port", "", "database port")
	flag.StringVar(&dbName, "db-name", "", "target database name")
	flag.Parse()

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Printf("failed to open connection with db: %s", err.Error())
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		fmt.Printf("failed to get postgres driver with instance: %s", err.Error())
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
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
