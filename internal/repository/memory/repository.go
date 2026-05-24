// Package memory provides an in-memory implementation of domain.Repository.
package memory

import (
	"sync"

	"github.com/harshity/url-shortner-go/internal/domain"
)

type repository struct {
	mu      sync.RWMutex
	byCode  map[string]*domain.URL
	byURL   map[string]*domain.URL
}

// New returns a thread-safe in-memory repository.
func New() domain.Repository {
	return &repository{
		byCode: make(map[string]*domain.URL),
		byURL:  make(map[string]*domain.URL),
	}
}

func (r *repository) Save(url *domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byCode[url.Code]; exists {
		return domain.ErrCodeConflict
	}
	cp := *url
	r.byCode[url.Code] = &cp
	r.byURL[url.OriginalURL] = &cp
	return nil
}

func (r *repository) FindByCode(code string) (*domain.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, ok := r.byCode[code]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *url
	return &cp, nil
}

func (r *repository) FindByURL(originalURL string) (*domain.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, ok := r.byURL[originalURL]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *url
	return &cp, nil
}

func (r *repository) IncrementClicks(code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	url, ok := r.byCode[code]
	if !ok {
		return domain.ErrNotFound
	}
	url.Clicks++
	return nil
}
