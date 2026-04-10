package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/vibe-composer/internal/analyzer"
	"github.com/vibe-composer/internal/db"
	"github.com/vibe-composer/internal/lyria"
	"github.com/vibe-composer/internal/prompt"
	"github.com/vibe-composer/internal/storage"
)

// Job represents a music generation job.
type Job struct {
	CompositionID string
	Username      string
	InputText     string
	AudioData     []byte
	AudioMIME     string
	MusicStyle    string
}

// Composer processes music generation jobs in the background.
type Composer struct {
	jobs     chan Job
	queries  *db.Queries
	analyzer *analyzer.Analyzer
	lyria    *lyria.Client
	gcs      *storage.GCS
	bucket   string
}

// NewComposer creates a new background composer worker.
func NewComposer(
	queries *db.Queries,
	analyzer *analyzer.Analyzer,
	lyriaClient *lyria.Client,
	gcs *storage.GCS,
	bucket string,
) *Composer {
	return &Composer{
		jobs:     make(chan Job, 100),
		queries:  queries,
		analyzer: analyzer,
		lyria:    lyriaClient,
		gcs:      gcs,
		bucket:   bucket,
	}
}

// Submit adds a job to the processing queue.
func (c *Composer) Submit(job Job) {
	c.jobs <- job
}

// Start begins processing jobs. Call in a goroutine.
func (c *Composer) Start(ctx context.Context) {
	slog.Info("composer worker started")
	for {
		select {
		case <-ctx.Done():
			slog.Info("composer worker stopped")
			return
		case job := <-c.jobs:
			c.process(ctx, job)
		}
	}
}

func (c *Composer) process(ctx context.Context, job Job) {
	slog.Info("processing composition", "id", job.CompositionID, "username", job.Username)

	// Update status to generating
	if err := c.queries.UpdateStatus(ctx, job.CompositionID, "generating"); err != nil {
		slog.Error("failed to update status", "error", err)
		return
	}

	// Step 1: Analyze input (emotion detection)
	var analysis *analyzer.AudioAnalysis
	var err error

	if len(job.AudioData) > 0 {
		// Audio input: transcribe + detect vocal emotion
		analysis, err = c.analyzer.AnalyzeAudio(ctx, job.AudioData, job.AudioMIME)
		if err != nil {
			c.fail(ctx, job.CompositionID, fmt.Sprintf("audio analysis failed: %v", err))
			return
		}
		// Save transcription back to DB
		if err := c.queries.UpdateInputText(ctx, job.CompositionID, analysis.Transcription); err != nil {
			slog.Error("failed to save transcription", "error", err)
		}
	} else {
		// Text input: analyze sentiment
		analysis, err = c.analyzer.AnalyzeText(ctx, job.InputText)
		if err != nil {
			c.fail(ctx, job.CompositionID, fmt.Sprintf("text analysis failed: %v", err))
			return
		}
	}

	slog.Info("analysis complete",
		"emotion", analysis.PrimaryEmotion,
		"intensity", analysis.EmotionIntensity,
		"summary", analysis.Summary,
	)

	// Step 2: Build creative music prompt
	// Use original text for text input, or transcription for audio input
	originalText := job.InputText
	if originalText == "" && analysis.Transcription != "" {
		originalText = analysis.Transcription
	}
	musicPrompt := prompt.Build(analysis, job.MusicStyle, originalText)

	emotionJSON, _ := json.Marshal(analysis)
	if err := c.queries.UpdateMusicPrompt(ctx, job.CompositionID, musicPrompt, string(emotionJSON)); err != nil {
		slog.Error("failed to save music prompt", "error", err)
	}

	// Step 3: Generate music with Lyria
	slog.Info("calling Lyria 3 Pro", "prompt_length", len(musicPrompt))
	result, err := c.lyria.GenerateMusic(ctx, musicPrompt)
	if err != nil {
		c.fail(ctx, job.CompositionID, fmt.Sprintf("music generation failed: %v", err))
		return
	}

	// Step 4: Upload result to GCS
	ext := "mp3"
	if result.MIMEType == "audio/wav" {
		ext = "wav"
	}
	gcsPath := fmt.Sprintf("results/%s/%s.%s", job.Username, job.CompositionID, ext)
	contentType := result.MIMEType
	if contentType == "" {
		contentType = "audio/mpeg"
	}

	if err := c.gcs.Upload(ctx, gcsPath, result.AudioData, contentType); err != nil {
		c.fail(ctx, job.CompositionID, fmt.Sprintf("upload failed: %v", err))
		return
	}

	// Step 5: Update composition as done
	if err := c.queries.UpdateResult(ctx, job.CompositionID, gcsPath, result.Lyrics); err != nil {
		slog.Error("failed to update result", "error", err)
		return
	}

	slog.Info("composition complete",
		"id", job.CompositionID,
		"gcs_path", gcsPath,
		"audio_size", len(result.AudioData),
	)
}

func (c *Composer) fail(ctx context.Context, id, msg string) {
	slog.Error("composition failed", "id", id, "error", msg)
	if err := c.queries.UpdateError(ctx, id, msg); err != nil {
		slog.Error("failed to update error status", "error", err)
	}
}
