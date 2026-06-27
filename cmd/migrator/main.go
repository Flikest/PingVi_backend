package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cassandra"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	env := flag.String("env", "./sso.local.env", "environment")
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
	db, err := sql.Open("pgx", os.Getenv("POSTGRES_CONNECTION_PATH"))
	if err != nil {
		panic(fmt.Sprintf("error with connect database: %v", err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("error pinging database: %v", err))
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(fmt.Sprintf("error creating driver: %v", err))
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		panic("MIGRATIONS_PATH is not set")
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		panic(fmt.Sprintf("error getting absolute path: %v", err))
	}

	log.Printf("Migrations path: %s", absPath)

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
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
	if migrationsPath == "" {
		panic("MIGRATIONS_PATH is not set")
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		panic(fmt.Sprintf("error getting absolute path: %v", err))
	}

	log.Printf("Migrations path: %s", absPath)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", absPath),
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
