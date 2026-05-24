package service

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/harshity/url-shortner-go/internal/domain"
	"github.com/harshity/url-shortner-go/pkg/generator"
)

type urlService struct {
	repo    domain.Repository
	baseURL string
	codeLen int
}

// New returns a URLService that satisfies domain.Service.
func New(repo domain.Repository, baseURL string, codeLen int) domain.Service {
	return &urlService{
		repo:    repo,
		baseURL: strings.TrimRight(baseURL, "/"),
		codeLen: codeLen,
	}
}

func (s *urlService) Shorten(originalURL string) (*domain.URL, error) {
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	if existing, err := s.repo.FindByURL(originalURL); err == nil {
		return existing, nil
	}

	code, err := generator.Code(s.codeLen)
	if err != nil {
		return nil, fmt.Errorf("generating code: %w", err)
	}

	u := &domain.URL{
		Code:        code,
		OriginalURL: originalURL,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.Save(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *urlService) Resolve(code string) (*domain.URL, error) {
	if !isValidCode(code) {
		return nil, domain.ErrInvalidCode
	}

	u, err := s.repo.FindByCode(code)
	if err != nil {
		return nil, err
	}

	// increment in background-safe way; ignore error since redirect still works
	_ = s.repo.IncrementClicks(code)
	return u, nil
}

func (s *urlService) Stats(code string) (*domain.URL, error) {
	if !isValidCode(code) {
		return nil, domain.ErrInvalidCode
	}
	return s.repo.FindByCode(code)
}

// validateURL rejects anything that is not an absolute http/https URL.
func validateURL(raw string) error {
	if raw == "" {
		return domain.ErrInvalidURL
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return domain.ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return domain.ErrInvalidURL
	}
	if parsed.Host == "" {
		return domain.ErrInvalidURL
	}
	return nil
}

// isValidCode allows only base-62 characters within a sane length range.
func isValidCode(code string) bool {
	if len(code) < 4 || len(code) > 20 {
		return false
	}
	for _, c := range code {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
