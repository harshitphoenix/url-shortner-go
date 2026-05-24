package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/harshity/url-shortner-go/internal/domain"
)

// URLHandler wires HTTP routes to the domain service.
type URLHandler struct {
	svc     domain.Service
	baseURL string
}

func NewURLHandler(svc domain.Service, baseURL string) *URLHandler {
	return &URLHandler{svc: svc, baseURL: baseURL}
}

// RegisterRoutes attaches all URL-shortener endpoints to mux.
// Route precedence (Go 1.22 ServeMux):
//
//	POST /shorten          → create
//	GET  /stats/{code}     → stats  (two-segment; more specific than /{code})
//	GET  /{code}           → redirect
//	GET  /health           → liveness probe
func (h *URLHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /shorten", h.Shorten)
	mux.HandleFunc("GET /stats/{code}", h.Stats)
	mux.HandleFunc("GET /{code}", h.Redirect)
	mux.HandleFunc("GET /health", h.Health)
}

// ── request / response DTOs ────────────────────────────────────────────────

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code        string `json:"code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type statsResponse struct {
	Code        string `json:"code"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	Clicks      int64  `json:"clicks"`
	CreatedAt   string `json:"created_at"`
}

// ── handlers ──────────────────────────────────────────────────────────────

func (h *URLHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := h.svc.Shorten(req.URL)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidURL) {
			respondError(w, http.StatusUnprocessableEntity, "invalid or unsupported URL")
			return
		}
		respondError(w, http.StatusInternalServerError, "could not shorten URL")
		return
	}

	respondJSON(w, http.StatusCreated, shortenResponse{
		Code:        u.Code,
		ShortURL:    h.baseURL + "/" + u.Code,
		OriginalURL: u.OriginalURL,
	})
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	u, err := h.svc.Resolve(code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalidCode) {
			http.NotFound(w, r)
			return
		}
		respondError(w, http.StatusInternalServerError, "could not resolve URL")
		return
	}

	http.Redirect(w, r, u.OriginalURL, http.StatusMovedPermanently)
}

func (h *URLHandler) Stats(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	u, err := h.svc.Stats(code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalidCode) {
			http.NotFound(w, r)
			return
		}
		respondError(w, http.StatusInternalServerError, "could not fetch stats")
		return
	}

	respondJSON(w, http.StatusOK, statsResponse{
		Code:        u.Code,
		OriginalURL: u.OriginalURL,
		ShortURL:    h.baseURL + "/" + u.Code,
		Clicks:      u.Clicks,
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *URLHandler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
