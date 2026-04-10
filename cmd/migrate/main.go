package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		slog.Error("failed to read migrations", "error", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			slog.Error("failed to read migration", "file", entry.Name(), "error", err)
			os.Exit(1)
		}

		slog.Info("running migration", "file", entry.Name())
		if _, err := pool.Exec(ctx, string(data)); err != nil {
			slog.Error("migration failed", "file", entry.Name(), "error", err)
			os.Exit(1)
		}
	}

	fmt.Println("✅ all migrations applied successfully")
}
