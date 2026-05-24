package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/harshity/url-shortner-go/internal/handler"
	"github.com/harshity/url-shortner-go/internal/repository/memory"
	"github.com/harshity/url-shortner-go/internal/service"
)

const baseURL = "http://localhost:8080"

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	svc := service.New(memory.New(), baseURL, 7)
	mux := http.NewServeMux()
	handler.NewURLHandler(svc, baseURL).RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func postShorten(t *testing.T, srv *httptest.Server, body string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/shorten",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /shorten: %v", err)
	}
	return resp
}

// ── /health ───────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

// ── POST /shorten ─────────────────────────────────────────────────────────────

func TestShorten_Created(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp := postShorten(t, srv, `{"url":"https://example.com"}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["code"] == "" {
		t.Error("expected non-empty code")
	}
	if body["short_url"] == "" {
		t.Error("expected non-empty short_url")
	}
	if body["original_url"] != "https://example.com" {
		t.Errorf("original_url = %q", body["original_url"])
	}
}

func TestShorten_InvalidURL(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp := postShorten(t, srv, `{"url":"not-a-url"}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", resp.StatusCode)
	}
}

func TestShorten_BadJSON(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp := postShorten(t, srv, `not json`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

// ── GET /{code} ───────────────────────────────────────────────────────────────

func TestRedirect_MovedPermanently(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp := postShorten(t, srv, `{"url":"https://example.com"}`)
	defer resp.Body.Close()

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	code := body["code"]

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	r, err := client.Get(srv.URL + "/" + code)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusMovedPermanently {
		t.Errorf("status = %d, want 301", r.StatusCode)
	}
	if loc := r.Header.Get("Location"); loc != "https://example.com" {
		t.Errorf("Location = %q, want https://example.com", loc)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/aaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// ── GET /stats/{code} ────────────────────────────────────────────────────────

func TestStats_OK(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp := postShorten(t, srv, `{"url":"https://example.com"}`)
	defer resp.Body.Close()

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	code := body["code"]

	r, err := http.Get(srv.URL + "/stats/" + code)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", r.StatusCode)
	}

	var stats map[string]any
	json.NewDecoder(r.Body).Decode(&stats)
	if stats["code"] != code {
		t.Errorf("code = %v, want %q", stats["code"], code)
	}
	if stats["clicks"].(float64) != 0 {
		t.Errorf("clicks = %v, want 0", stats["clicks"])
	}
}

func TestStats_NotFound(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/stats/aaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}
