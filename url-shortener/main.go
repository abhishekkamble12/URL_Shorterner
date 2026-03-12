// main.go is the entry point of the URL Shortener API.
//
// Startup sequence:
//  1. Connect to the database and run migrations.
//  2. Create a Gin HTTP engine with sensible defaults.
//  3. Register all API routes.
//  4. Start listening on the configured port.
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"url-shortener/config"
	"url-shortener/routes"
)

func main() {
	// ----------------------------------------------------------------
	// 1. Database initialisation
	// ----------------------------------------------------------------
	// ConnectDB opens the SQLite file (or PostgreSQL in production),
	// creates the `urls` table if it does not exist, and stores the
	// connection in the config.DB global so controllers can use it.
	config.ConnectDB()

	// ----------------------------------------------------------------
	// 2. Gin engine setup
	// ----------------------------------------------------------------
	// gin.Default() creates a router pre-configured with:
	//   - Logger middleware  → prints each request to stdout
	//   - Recovery middleware → catches panics and returns 500 instead of crashing
	//
	// In production you might use gin.New() and add your own structured logger.
	r := gin.Default()

	// Increase the max multipart memory limit (optional, good practice).
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	// ----------------------------------------------------------------
	// 3. Route registration
	// ----------------------------------------------------------------
	// All endpoints are declared in routes/routes.go.
	routes.RegisterRoutes(r)

	// ----------------------------------------------------------------
	// 4. Start the HTTP server
	// ----------------------------------------------------------------
	// PORT can be overridden via environment variable, which is the
	// standard 12-factor app pattern for containerised deployments.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("URL Shortener API starting on port %s ...", port)

	// r.Run() blocks until the server is stopped or crashes.
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
