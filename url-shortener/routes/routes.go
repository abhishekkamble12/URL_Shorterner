// Package routes wires HTTP endpoints to their controller functions.
// Keeping routes in a dedicated file makes it easy to see the full API
// surface at a glance and add middleware (auth, rate limiting, CORS) later.
package routes

import (
	"github.com/gin-gonic/gin"

	"url-shortener/controllers"
)

// RegisterRoutes attaches all application routes to the provided Gin engine.
// It is called once during application startup in main.go.
//
// Route summary:
//
//	GET    /                       → landing page with API usage info
//	POST   /api/shorten            → create a new short URL
//	GET    /api/stats/:shortcode   → fetch click analytics
//	GET    /api/urls               → list all stored URLs
//	GET    /api/health             → simple liveness probe
//	GET    /:shortcode             → redirect to the original URL
func RegisterRoutes(r *gin.Engine) {
	r.GET("/", controllers.HomePage)

	// All API endpoints live under /api/ so they never collide with the
	// GET /:shortcode redirect wildcard at the root level.
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		api.POST("/shorten", controllers.ShortenURL)
		api.GET("/stats/:shortcode", controllers.GetStats)
		api.GET("/urls", controllers.ListURLs)
	}

	r.GET("/:shortcode", controllers.RedirectURL)
}
