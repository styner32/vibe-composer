package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"cloud.google.com/go/storage"
)

// GCS wraps Google Cloud Storage operations.
type GCS struct {
	client *storage.Client
	bucket string
}

// NewGCS creates a new GCS client.
func NewGCS(ctx context.Context, bucket string) (*GCS, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating GCS client: %w", err)
	}
	return &GCS{client: client, bucket: bucket}, nil
}

// Upload writes data to a GCS object.
func (g *GCS) Upload(ctx context.Context, path string, data []byte, contentType string) error {
	wc := g.client.Bucket(g.bucket).Object(path).NewWriter(ctx)
	wc.ContentType = contentType
	if _, err := wc.Write(data); err != nil {
		wc.Close()
		return fmt.Errorf("writing to GCS: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("closing GCS writer: %w", err)
	}
	slog.Info("uploaded to GCS", "path", path, "size", len(data))
	return nil
}

// Download reads data from a GCS object.
func (g *GCS) Download(ctx context.Context, path string) ([]byte, string, error) {
	obj := g.client.Bucket(g.bucket).Object(path)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("getting object attrs: %w", err)
	}

	rc, err := obj.NewReader(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("creating GCS reader: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, "", fmt.Errorf("reading from GCS: %w", err)
	}
	return data, attrs.ContentType, nil
}

// SignedURL generates a signed URL for downloading a GCS object.
func (g *GCS) SignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	url, err := g.client.Bucket(g.bucket).SignedURL(path, &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expiry),
	})
	if err != nil {
		return "", fmt.Errorf("generating signed URL: %w", err)
	}
	return url, nil
}

// Close closes the GCS client.
func (g *GCS) Close() error {
	return g.client.Close()
}
