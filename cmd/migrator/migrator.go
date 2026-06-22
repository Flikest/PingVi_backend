package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cassandra"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/joho/godotenv"
)

func main() {
	env := flag.String("env", "sso.local.env", "environment")
	db := flag.String("db", "postgresql", "name of the database for driver generation")
	flag.Parse()

	if err := godotenv.Load(*env); err != nil {
		panic("failed to load environment variables 😫😫😫")
	}

	switch *db {
	case "postgresql":
		migratePostgres()
	case "scylladb":
		migrateScyllaDB()
	default:
		panic(fmt.Sprintf("unsupported database type: %s", *db))
	}

	log.Println("migrations up 🚀🚀🚀")
}

func migratePostgres() {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_CONNECTION_PATH"))
	if err != nil {
		panic(fmt.Sprintf("error with connect database: %v", err))
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic("error creating driver")
	}

	m, err := migrate.NewWithDatabaseInstance(
		os.Getenv("MIGRATIONS_PATH"),
		"postgres", driver)
	if err != nil {
		panic(fmt.Sprintf("error with creating migrations: %v 🐖🐖🐖", err))
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("migrations not apply")
			return
		}
		panic(fmt.Sprintf("error migrations up: %v 🐷🐷🐷", err))
	}
}

func migrateScyllaDB() {
	connString := os.Getenv("SCYLLADB_CONN_STRING")
	if connString == "" {
		panic("SCYLLADB_CONN_STRING is not set 😫😫😫")
	}

	if !strings.Contains(connString, "x-multi-statement") {
		if strings.Contains(connString, "?") {
			connString += "&x-multi-statement=true"
		} else {
			connString += "?x-multi-statement=true"
		}
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")

	m, err := migrate.New(
		migrationsPath,
		connString,
	)
	if err != nil {
		panic(fmt.Sprintf("error with creating migrations: %v 🐖🐖🐖", err))
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("migrations not apply")
			return
		}
		panic(fmt.Sprintf("error migrations up: %v 🐷🐷🐷", err))
	}
}
