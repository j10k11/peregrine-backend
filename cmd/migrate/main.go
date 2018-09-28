package main

import (
	"database/sql"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/Pigmice2733/peregrine-backend/internal/config"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var steps = flag.Int("steps", 0, "Number of steps to migrate. Leave unspecified to migrate all the way up or down.")
	var up = flag.Bool("up", false, "Migrate up. Cannot be used with -down.")
	var down = flag.Bool("down", false, "Migrate down. Cannot be used with -up.")
	var migrationstable = flag.String("migrationstable", "migrations", "Name of SQL table to store migrations in.")
	var basePath = flag.String("basePath", "", "Path to the etc directory where the config file is.")

	flag.Parse()

	// Neither are set or both are set
	if *up == *down {
		fmt.Printf("Error: must specify either -up or -down\n")
		return
	}

	c, err := config.Open(*basePath)
	if err != nil {
		fmt.Printf("Error: opening config: %v\n", err)
		return
	}

	db, err := sql.Open("postgres", c.Database.ConnectionInfo())
	if err != nil {
		fmt.Printf("Error: connecting to db: %v\n", err)
		return
	}
	defer db.Close()

	config := &postgres.Config{MigrationsTable: *migrationstable, DatabaseName: c.Database.Name}
	driver, err := postgres.WithInstance(db, config)
	if err != nil {
		fmt.Printf("Error: getting PostgreSQL driver: %v\n", err)
		return
	}

	migrationsSource := "file://"
	if *basePath != "" {
		migrationsSource += filepath.ToSlash(filepath.Clean(*basePath)) + "/"
	} else {
		migrationsSource += "./"
	}
	migrationsSource += "migrations"

	m, err := migrate.NewWithDatabaseInstance(
		migrationsSource,
		c.Database.Name, driver)
	if err != nil {
		fmt.Printf("Error: creating migrations: %v\n", err)
		return
	}

	defer m.Close()

	if *steps == 0 && *up {
		err = m.Up()
	} else if *steps == 0 && *down {
		err = m.Down()
	} else {
		if *down {
			*steps = -*steps
		}
		err = m.Steps(*steps)
	}

	if err != nil {
		fmt.Printf("Error: running migrations: %v\n", err)
		return
	}

	fmt.Println("Migrations successfully run")

	srcErr, dbErr := m.Close()
	if srcErr != nil {
		fmt.Printf("Error: closing migrations source: %v\n", srcErr)
	}
	if dbErr != nil {
		fmt.Printf("Error: closing connection to database: %v\n", dbErr)
	}

	if err = db.Close(); err != nil {
		fmt.Printf("Error: closing database: %v\n", err)
	}
}
