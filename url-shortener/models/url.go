// Package models defines the data structures (schemas) used throughout the app.
// GORM reads these structs and maps them to database tables automatically.
package models

import "time"

// URL represents a single shortened URL record stored in the database.
// The `gorm` struct tags tell GORM how to map each field to a DB column.
// The `json` struct tags control how the struct is serialized to JSON in API responses.
type URL struct {
	// ID is the auto-incrementing primary key managed by GORM.
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// OriginalURL is the full long URL the user submitted (e.g. https://example.com/some/long/path).
	OriginalURL string `gorm:"not null" json:"original_url"`

	// ShortCode is the unique 6–8 character identifier (e.g. "abc123").
	// The `uniqueIndex` constraint ensures no two rows can have the same short code.
	ShortCode string `gorm:"uniqueIndex;not null" json:"short_code"`

	// Clicks tracks how many times this short link has been visited.
	Clicks int `gorm:"default:0" json:"clicks"`

	// CreatedAt is automatically set by GORM when the record is first created.
	CreatedAt time.Time `json:"created_at"`
}
