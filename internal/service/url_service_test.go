package service_test

import (
	"errors"
	"testing"

	"github.com/harshity/url-shortner-go/internal/domain"
	"github.com/harshity/url-shortner-go/internal/repository/memory"
	"github.com/harshity/url-shortner-go/internal/service"
)

func newSvc() domain.Service {
	return service.New(memory.New(), "http://localhost:8080", 7)
}

func TestShorten_ReturnsURL(t *testing.T) {
	svc := newSvc()
	u, err := svc.Shorten("https://example.com")
	if err != nil {
		t.Fatalf("Shorten: %v", err)
	}
	if u.Code == "" {
		t.Error("expected non-empty code")
	}
	if u.OriginalURL != "https://example.com" {
		t.Errorf("OriginalURL = %q", u.OriginalURL)
	}
	if len(u.Code) != 7 {
		t.Errorf("code length = %d, want 7", len(u.Code))
	}
}

func TestShorten_InvalidURLs(t *testing.T) {
	svc := newSvc()
	cases := []string{
		"",
		"not-a-url",
		"ftp://example.com",
		"//example.com",
		"http://",
	}
	for _, raw := range cases {
		_, err := svc.Shorten(raw)
		if !errors.Is(err, domain.ErrInvalidURL) {
			t.Errorf("Shorten(%q) = %v, want ErrInvalidURL", raw, err)
		}
	}
}

func TestShorten_ValidURLs(t *testing.T) {
	svc := newSvc()
	cases := []string{
		"https://example.com",
		"http://example.com/path?q=1#anchor",
		"https://sub.domain.co.uk/a/b/c",
	}
	for _, raw := range cases {
		if _, err := svc.Shorten(raw); err != nil {
			t.Errorf("Shorten(%q) unexpected error: %v", raw, err)
		}
	}
}

func TestResolve_IncrementsClicks(t *testing.T) {
	svc := newSvc()
	u, _ := svc.Shorten("https://example.com")

	for i := 1; i <= 3; i++ {
		resolved, err := svc.Resolve(u.Code)
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if resolved.OriginalURL != "https://example.com" {
			t.Errorf("OriginalURL = %q", resolved.OriginalURL)
		}
	}

	stats, _ := svc.Stats(u.Code)
	if stats.Clicks != 3 {
		t.Errorf("Clicks = %d, want 3", stats.Clicks)
	}
}

func TestResolve_NotFound(t *testing.T) {
	svc := newSvc()
	_, err := svc.Resolve("aaaaaaa")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestResolve_InvalidCode(t *testing.T) {
	svc := newSvc()
	cases := []string{"", "ab!", "a b", "abc"}
	for _, code := range cases {
		_, err := svc.Resolve(code)
		if !errors.Is(err, domain.ErrInvalidCode) {
			t.Errorf("Resolve(%q) = %v, want ErrInvalidCode", code, err)
		}
	}
}

func TestStats_NotFound(t *testing.T) {
	svc := newSvc()
	_, err := svc.Stats("aaaaaaa")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestStats_InvalidCode(t *testing.T) {
	svc := newSvc()
	_, err := svc.Stats("!bad")
	if !errors.Is(err, domain.ErrInvalidCode) {
		t.Errorf("expected ErrInvalidCode, got %v", err)
	}
}
