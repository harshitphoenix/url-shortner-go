package memory_test

import (
	"testing"

	"github.com/harshity/url-shortner-go/internal/domain"
	"github.com/harshity/url-shortner-go/internal/repository/memory"
)

func newURL(code, original string) *domain.URL {
	return &domain.URL{Code: code, OriginalURL: original}
}

func TestSave_AndFind(t *testing.T) {
	repo := memory.New()
	u := newURL("abc1234", "https://example.com")

	if err := repo.Save(u); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByCode("abc1234")
	if err != nil {
		t.Fatalf("FindByCode: %v", err)
	}
	if got.OriginalURL != u.OriginalURL {
		t.Errorf("OriginalURL = %q, want %q", got.OriginalURL, u.OriginalURL)
	}
}

func TestSave_ConflictReturnsError(t *testing.T) {
	repo := memory.New()
	u := newURL("abc1234", "https://example.com")

	_ = repo.Save(u)
	if err := repo.Save(u); err != domain.ErrCodeConflict {
		t.Errorf("expected ErrCodeConflict, got %v", err)
	}
}

func TestSave_IsolatesCopy(t *testing.T) {
	repo := memory.New()
	u := newURL("abc1234", "https://example.com")
	_ = repo.Save(u)

	u.OriginalURL = "https://mutated.com"

	got, _ := repo.FindByCode("abc1234")
	if got.OriginalURL == "https://mutated.com" {
		t.Error("Save did not copy: mutation affected stored value")
	}
}

func TestFindByCode_NotFound(t *testing.T) {
	repo := memory.New()
	_, err := repo.FindByCode("missing")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFindByCode_IsolatesCopy(t *testing.T) {
	repo := memory.New()
	_ = repo.Save(newURL("abc1234", "https://example.com"))

	got, _ := repo.FindByCode("abc1234")
	got.OriginalURL = "https://mutated.com"

	got2, _ := repo.FindByCode("abc1234")
	if got2.OriginalURL == "https://mutated.com" {
		t.Error("FindByCode did not copy: mutation affected stored value")
	}
}

func TestIncrementClicks(t *testing.T) {
	repo := memory.New()
	_ = repo.Save(newURL("abc1234", "https://example.com"))

	for i := 1; i <= 3; i++ {
		if err := repo.IncrementClicks("abc1234"); err != nil {
			t.Fatalf("IncrementClicks: %v", err)
		}
		got, _ := repo.FindByCode("abc1234")
		if got.Clicks != int64(i) {
			t.Errorf("Clicks = %d, want %d", got.Clicks, i)
		}
	}
}

func TestIncrementClicks_NotFound(t *testing.T) {
	repo := memory.New()
	if err := repo.IncrementClicks("missing"); err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
