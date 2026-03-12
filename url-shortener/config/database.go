// Package config handles application-wide configuration,
// starting with the database connection and auto-migration.
package config

import (
	"log"
	"os"

	// glebarez/sqlite is a pure-Go SQLite driver that works without CGO.
	// This makes cross-compilation and Docker builds much simpler.
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"url-shortener/models"
)

// DB is the global database handle shared across the entire application.
// All repositories and controllers import this package to access it.
var DB *gorm.DB

// ConnectDB initialises the database connection.
//
// By default it uses SQLite and stores data in a local file called "urls.db".
// To switch to PostgreSQL later, replace the sqlite.Open(...) call with:
//
//	import "gorm.io/driver/postgres"
//	dsn := "host=localhost user=postgres password=secret dbname=urlshortener port=5432"
//	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
//
// The rest of the codebase stays identical — that is the power of GORM.
func ConnectDB() {
	var err error

	// Determine which database file to use.
	// The DB_PATH environment variable lets us override the default in Docker or tests.
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "urls.db"
	}

	// gorm.Config lets us customise logging behaviour.
	// In production you'd swap logger.Default for a structured logger (e.g. zerolog).
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Open the SQLite database at the resolved path.
	DB, err = gorm.Open(sqlite.Open(dbPath), gormConfig)
	if err != nil {
		// Fatal stops the process immediately — the app cannot run without a DB.
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established.")

	// AutoMigrate inspects the URL struct and creates (or updates) the `urls` table.
	// It never drops columns or data, making it safe to run on every startup.
	if err := DB.AutoMigrate(&models.URL{}); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	log.Println("Database migration completed successfully.")
}
