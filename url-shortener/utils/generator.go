// Package utils provides small, self-contained helper functions
// that are reused across the application.
package utils

import (
	"math/rand"
	"time"

	"gorm.io/gorm"

	"url-shortener/models"
)

// charset is the pool of characters used when building a short code.
// Using only lowercase letters and digits keeps URLs clean and URL-safe.
const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// shortCodeLength controls how long each generated code will be.
// 7 characters from a 36-character alphabet gives 36^7 ≈ 78 billion combinations,
// which is more than enough for any real-world use-case.
const shortCodeLength = 7

// seededRand is a random source seeded once at package initialisation time.
// Using a fixed seed (time.Now().UnixNano()) ensures different codes each run.
var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// generateCode builds a single random short code of the configured length.
// It is unexported because callers should use GenerateUniqueCode instead.
func generateCode() string {
	b := make([]byte, shortCodeLength)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateUniqueCode creates a collision-free short code by repeatedly
// generating candidates until one that does not yet exist in the database
// is found.
//
// In practice, collisions are astronomically rare, so this loop almost
// always terminates on the first iteration.
//
// Parameters:
//   - db: the active GORM database handle used to check for existing codes.
//
// Returns the unique short code string.
func GenerateUniqueCode(db *gorm.DB) string {
	for {
		code := generateCode()

		// Try to find an existing URL record with this code.
		var existing models.URL
		result := db.Where("short_code = ?", code).First(&existing)

		// gorm.ErrRecordNotFound means the code is available — return it.
		if result.Error == gorm.ErrRecordNotFound {
			return code
		}

		// Any other error (e.g. DB connection issue) also means the code was
		// not confirmed as taken, so we try a new one to stay safe.
	}
}
