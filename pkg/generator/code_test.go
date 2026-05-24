package generator_test

import (
	"testing"

	"github.com/harshity/url-shortner-go/pkg/generator"
)

const base62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func TestCode_Length(t *testing.T) {
	for _, length := range []int{4, 7, 12, 20} {
		code, err := generator.Code(length)
		if err != nil {
			t.Fatalf("Code(%d) error: %v", length, err)
		}
		if len(code) != length {
			t.Errorf("Code(%d) = %q, want length %d", length, code, length)
		}
	}
}

func TestCode_Charset(t *testing.T) {
	allowed := make(map[rune]bool, len(base62))
	for _, c := range base62 {
		allowed[c] = true
	}

	code, err := generator.Code(50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range code {
		if !allowed[c] {
			t.Errorf("Code contains invalid character %q", c)
		}
	}
}

func TestCode_Uniqueness(t *testing.T) {
	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		code, err := generator.Code(7)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if seen[code] {
			t.Fatalf("duplicate code generated: %q", code)
		}
		seen[code] = true
	}
}
