// Package memory provides an in-memory implementation of domain.Repository.
package memory

import (
	"sync"

	"github.com/harshity/url-shortner-go/internal/domain"
)

type repository struct {
	mu   sync.RWMutex
	data map[string]*domain.URL
}

// New returns a thread-safe in-memory repository.
func New() domain.Repository {
	return &repository{
		data: make(map[string]*domain.URL),
	}
}

func (r *repository) Save(url *domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[url.Code]; exists {
		return domain.ErrCodeConflict
	}
	// store a copy so the caller cannot mutate internal state
	cp := *url
	r.data[url.Code] = &cp
	return nil
}

func (r *repository) FindByCode(code string) (*domain.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, ok := r.data[code]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *url
	return &cp, nil
}

func (r *repository) IncrementClicks(code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	url, ok := r.data[code]
	if !ok {
		return domain.ErrNotFound
	}
	url.Clicks++
	return nil
}
