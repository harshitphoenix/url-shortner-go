package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/harshity/url-shortner-go/config"
	"github.com/harshity/url-shortner-go/internal/handler"
	"github.com/harshity/url-shortner-go/internal/repository/memory"
	"github.com/harshity/url-shortner-go/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()

	// wire up layers: repository → service → handler
	repo := memory.New()
	svc := service.New(repo, cfg.BaseURL, cfg.CodeLen)

	mux := http.NewServeMux()
	h := handler.NewURLHandler(svc, cfg.BaseURL)
	h.RegisterRoutes(mux)

	// middleware stack (outermost first): Logger → MaxBody → RequireJSON → mux
	stack := handler.Logger(
		handler.MaxBody(1 << 20)( // 1 MB body limit
			handler.RequireJSON(mux),
		),
	)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      stack,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("server starting", "addr", srv.Addr, "base_url", cfg.BaseURL)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
}
