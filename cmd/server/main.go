package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"google.golang.org/genai"

	"github.com/vibe-composer/internal/analyzer"
	"github.com/vibe-composer/internal/auth"
	"github.com/vibe-composer/internal/config"
	"github.com/vibe-composer/internal/db"
	"github.com/vibe-composer/internal/handler"
	"github.com/vibe-composer/internal/lyria"
	"github.com/vibe-composer/internal/lyricist"
	"github.com/vibe-composer/internal/storage"
	"github.com/vibe-composer/internal/worker"
)

func main() {
	// Setup structured logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Load .env file if present (ignored in production if missing)
	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found, using environment variables")
	}

	cfg := config.Load()

	if cfg.GoogleAPIKey == "" {
		slog.Error("GOOGLE_API_KEY is required")
		os.Exit(1)
	}
	if cfg.GCSBucket == "" {
		slog.Error("GCS_BUCKET is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	queries := db.NewQueries(pool)

	// Initialize GCS
	gcs, err := storage.NewGCS(ctx, cfg.GCSBucket)
	if err != nil {
		slog.Error("failed to initialize GCS", "error", err)
		os.Exit(1)
	}
	defer gcs.Close()

	// Initialize Gemini/GenAI client
	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.GoogleAPIKey,
	})
	if err != nil {
		slog.Error("failed to initialize GenAI client", "error", err)
		os.Exit(1)
	}

	// Initialize components
	audioAnalyzer := analyzer.NewAnalyzer(genaiClient)
	songLyricist := lyricist.NewLyricist(genaiClient)
	lyriaClient := lyria.NewClient(genaiClient)
	composer := worker.NewComposer(queries, audioAnalyzer, songLyricist, lyriaClient, gcs, cfg.GCSBucket)

	// Start background worker
	go composer.Start(ctx)

	// Setup HTTP handlers
	h := handler.New(queries, gcs, composer, cfg.GCSBucket)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// CORS for local development
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Serve static files
	webFS := http.Dir("web")
	r.Handle("/*", http.FileServer(webFS))

	// API routes (protected by basic auth)
	r.Route("/api", func(r chi.Router) {
		r.Use(auth.BasicAuth(cfg))

		r.Get("/me", h.GetMe)
		r.Post("/compose", h.Compose)
		r.Get("/compositions", h.ListCompositions)
		r.Get("/compositions/{id}", h.GetComposition)
		r.Get("/compositions/{id}/download", h.DownloadComposition)

		// Clip routes (숙성 — Maturation Recording)
		r.Post("/clips", h.CreateClip)
		r.Get("/clips", h.ListClips)
		r.Delete("/clips/{id}", h.DeleteClip)
		r.Get("/clips/{id}/download", h.DownloadClip)
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute, // Long timeout for music generation downloads
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	slog.Info("server starting", "addr", addr, "allowed_users", cfg.AllowedUsers)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
