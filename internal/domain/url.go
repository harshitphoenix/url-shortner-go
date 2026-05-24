package domain

import "time"

// URL is the core business entity.
type URL struct {
	Code        string
	OriginalURL string
	CreatedAt   time.Time
	Clicks      int64
}

// Repository defines the persistence contract for URLs.
type Repository interface {
	Save(url *URL) error
	FindByCode(code string) (*URL, error)
	IncrementClicks(code string) error
}

// Service defines the business-logic contract for URL operations.
type Service interface {
	Shorten(originalURL string) (*URL, error)
	Resolve(code string) (*URL, error)
	Stats(code string) (*URL, error)
}
