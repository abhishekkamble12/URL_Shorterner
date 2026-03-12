// Package controllers holds the HTTP handler functions (controllers).
// Each function maps 1-to-1 with an API endpoint defined in routes/routes.go.
// Controllers are responsible for:
//   1. Parsing and validating the request.
//   2. Performing the business logic (via GORM).
//   3. Returning a structured JSON response.
package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"url-shortener/config"
	"url-shortener/models"
	"url-shortener/utils"
)

// -----------------------------------------------------------------------
// Request / Response payload structs
// -----------------------------------------------------------------------

// ShortenRequest is the JSON body expected by POST /shorten.
// The `binding:"required,url"` tag makes Gin validate the field automatically:
//   - "required" → field must be present and non-empty
//   - "url"      → value must be a valid URL format
type ShortenRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// ShortenResponse is the JSON body returned after successfully creating a short URL.
type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

// -----------------------------------------------------------------------
// Helper: baseURL
// -----------------------------------------------------------------------

// baseURL returns the server's base address used when building short URLs.
// It reads the BASE_URL environment variable so the value can be changed
// without recompiling (e.g. to a real domain in production).
// Falls back to http://localhost:8080 during local development.
func baseURL() string {
	if url := os.Getenv("BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

// -----------------------------------------------------------------------
// GET / — Landing page
// -----------------------------------------------------------------------

func HomePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "URL Shortener API",
		"endpoints": gin.H{
			"POST /api/shorten":              "Create a short URL",
			"GET  /api/stats/:shortcode":     "View click analytics",
			"GET  /api/urls":                 "List all shortened URLs",
			"GET  /api/health":               "Health check",
			"GET  /:shortcode":               "Redirect to original URL",
		},
	})
}

// -----------------------------------------------------------------------
// POST /api/shorten — Create a new short URL
// -----------------------------------------------------------------------

// ShortenURL handles POST /shorten.
// It validates the incoming URL, generates a unique short code,
// persists the record, and returns the full short URL to the client.
func ShortenURL(c *gin.Context) {
	var req ShortenRequest

	// ShouldBindJSON parses the JSON body and runs the validation rules
	// declared in the struct tags. It returns an error on parse or validation failure.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. Please provide a valid 'url' field."})
		return
	}

	// Generate a short code that is guaranteed not to collide with any existing one.
	shortCode := utils.GenerateUniqueCode(config.DB)

	// Build the URL model that GORM will persist.
	newURL := models.URL{
		OriginalURL: req.URL,
		ShortCode:   shortCode,
		Clicks:      0,
	}

	// Insert the record into the database.
	if result := config.DB.Create(&newURL); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL. Please try again."})
		return
	}

	// Return the full shortened URL.
	c.JSON(http.StatusCreated, ShortenResponse{
		ShortURL: baseURL() + "/" + shortCode,
	})
}

// -----------------------------------------------------------------------
// GET /:shortcode — Redirect to original URL
// -----------------------------------------------------------------------

// RedirectURL handles GET /:shortcode.
// It looks up the short code, increments the click counter,
// and issues an HTTP 302 redirect to the original URL.
func RedirectURL(c *gin.Context) {
	shortCode := c.Param("shortcode")

	var urlRecord models.URL

	// Look up the URL by its short code.
	result := config.DB.Where("short_code = ?", shortCode).First(&urlRecord)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error."})
		}
		return
	}

	// Atomically increment the click counter.
	// Using DB.Model(&urlRecord).UpdateColumn ensures only the `clicks` column
	// is updated and CreatedAt / other columns are not touched.
	config.DB.Model(&urlRecord).UpdateColumn("clicks", gorm.Expr("clicks + 1"))

	// 302 Found — temporary redirect so the browser always checks back with us.
	c.Redirect(http.StatusFound, urlRecord.OriginalURL)
}

// -----------------------------------------------------------------------
// GET /stats/:shortcode — Return analytics for a short URL
// -----------------------------------------------------------------------

// GetStats handles GET /stats/:shortcode.
// It returns the original URL, click count, and creation timestamp.
func GetStats(c *gin.Context) {
	shortCode := c.Param("shortcode")

	var urlRecord models.URL

	result := config.DB.Where("short_code = ?", shortCode).First(&urlRecord)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error."})
		}
		return
	}

	// Return a clean analytics summary.
	c.JSON(http.StatusOK, gin.H{
		"original_url": urlRecord.OriginalURL,
		"shortcode":    urlRecord.ShortCode,
		"clicks":       urlRecord.Clicks,
		"created_at":   urlRecord.CreatedAt.Format("2006-01-02"),
	})
}

// -----------------------------------------------------------------------
// GET /urls — List all shortened URLs
// -----------------------------------------------------------------------

// ListURLs handles GET /urls.
// It returns every stored URL record, most recently created first.
// In a production system you would add pagination (limit/offset query params)
// to avoid loading millions of rows into memory.
func ListURLs(c *gin.Context) {
	var urls []models.URL

	// ORDER BY id DESC puts the newest records at the top of the list.
	if result := config.DB.Order("id desc").Find(&urls); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve URLs."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(urls),
		"urls":  urls,
	})
}
