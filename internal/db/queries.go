package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Composition represents a music composition request and its result.
type Composition struct {
	ID               string    `json:"id"`
	Username         string    `json:"username"`
	Status           string    `json:"status"`
	InputType        string    `json:"input_type"`
	InputText        *string   `json:"input_text,omitempty"`
	InputAudioURL    *string   `json:"input_audio_url,omitempty"`
	Emotion          *string   `json:"emotion,omitempty"`
	MusicStyle       string    `json:"music_style"`
	VoiceGender      string    `json:"voice_gender"`
	LyricType        string    `json:"lyric_type"`
	MusicPrompt      *string   `json:"music_prompt,omitempty"`
	GeneratedLyrics  *string   `json:"generated_lyrics,omitempty"`
	ResultURL        *string   `json:"result_url,omitempty"`
	ResultLyrics     *string   `json:"result_lyrics,omitempty"`
	ErrorMessage     *string   `json:"error_message,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Queries provides database operations for compositions.
type Queries struct {
	pool *pgxpool.Pool
}

// NewQueries creates a new Queries instance.
func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

// CreateComposition inserts a new composition and returns its ID.
func (q *Queries) CreateComposition(ctx context.Context, c *Composition) (string, error) {
	var id string
	err := q.pool.QueryRow(ctx,
		`INSERT INTO compositions (username, status, input_type, input_text, input_audio_url, emotion, music_style, voice_gender, lyric_type, music_prompt)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id`,
		c.Username, c.Status, c.InputType, c.InputText, c.InputAudioURL, c.Emotion, c.MusicStyle, c.VoiceGender, c.LyricType, c.MusicPrompt,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("inserting composition: %w", err)
	}
	return id, nil
}

// GetComposition retrieves a composition by ID.
func (q *Queries) GetComposition(ctx context.Context, id string) (*Composition, error) {
	c := &Composition{}
	err := q.pool.QueryRow(ctx,
		`SELECT id, username, status, input_type, input_text, input_audio_url, emotion,
		        music_style, voice_gender, lyric_type, music_prompt, generated_lyrics, result_url, result_lyrics, error_message,
		        created_at, updated_at
		 FROM compositions WHERE id = $1`, id,
	).Scan(
		&c.ID, &c.Username, &c.Status, &c.InputType, &c.InputText, &c.InputAudioURL,
		&c.Emotion, &c.MusicStyle, &c.VoiceGender, &c.LyricType, &c.MusicPrompt, &c.GeneratedLyrics, &c.ResultURL, &c.ResultLyrics,
		&c.ErrorMessage, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting composition: %w", err)
	}
	return c, nil
}

// ListCompositions retrieves all compositions for a user, most recent first.
func (q *Queries) ListCompositions(ctx context.Context, username string) ([]*Composition, error) {
	rows, err := q.pool.Query(ctx,
		`SELECT id, username, status, input_type, input_text, input_audio_url, emotion,
		        music_style, voice_gender, lyric_type, music_prompt, generated_lyrics, result_url, result_lyrics, error_message,
		        created_at, updated_at
		 FROM compositions WHERE username = $1
		 ORDER BY created_at DESC
		 LIMIT 50`, username,
	)
	if err != nil {
		return nil, fmt.Errorf("listing compositions: %w", err)
	}
	defer rows.Close()

	var compositions []*Composition
	for rows.Next() {
		c := &Composition{}
		if err := rows.Scan(
			&c.ID, &c.Username, &c.Status, &c.InputType, &c.InputText, &c.InputAudioURL,
			&c.Emotion, &c.MusicStyle, &c.VoiceGender, &c.LyricType, &c.MusicPrompt, &c.GeneratedLyrics, &c.ResultURL, &c.ResultLyrics,
			&c.ErrorMessage, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning composition: %w", err)
		}
		compositions = append(compositions, c)
	}
	return compositions, nil
}

// HasActiveGeneration checks if a user has an in-progress generation.
func (q *Queries) HasActiveGeneration(ctx context.Context, username string) (bool, error) {
	var count int
	err := q.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM compositions
		 WHERE username = $1 AND status IN ('pending', 'generating')`, username,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking active generation: %w", err)
	}
	return count > 0, nil
}

// UpdateStatus updates the status of a composition.
func (q *Queries) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := q.pool.Exec(ctx,
		`UPDATE compositions SET status = $1, updated_at = now() WHERE id = $2`,
		status, id,
	)
	return err
}

// UpdateResult updates a composition with the generation result.
func (q *Queries) UpdateResult(ctx context.Context, id, resultURL, lyrics string) error {
	_, err := q.pool.Exec(ctx,
		`UPDATE compositions SET status = 'done', result_url = $1, result_lyrics = $2, updated_at = now()
		 WHERE id = $3`,
		resultURL, lyrics, id,
	)
	return err
}

// UpdateError marks a composition as failed with an error message.
func (q *Queries) UpdateError(ctx context.Context, id, errMsg string) error {
	_, err := q.pool.Exec(ctx,
		`UPDATE compositions SET status = 'failed', error_message = $1, updated_at = now()
		 WHERE id = $2`,
		errMsg, id,
	)
	return err
}

// UpdateMusicPrompt sets the music prompt and emotion for a composition.
func (q *Queries) UpdateMusicPrompt(ctx context.Context, id, prompt, emotion string) error {
	_, err := q.pool.Exec(ctx,
		`UPDATE compositions SET music_prompt = $1, emotion = $2, updated_at = now()
		 WHERE id = $3`,
		prompt, emotion, id,
	)
	return err
}

// UpdateInputText sets the transcribed text for an audio composition.
func (q *Queries) UpdateInputText(ctx context.Context, id, text string) error {
	_, err := q.pool.Exec(ctx,
		`UPDATE compositions SET input_text = $1, updated_at = now()
		 WHERE id = $2`,
		text, id,
	)
	return err
}

// UpdateGeneratedLyrics saves the pre-generated lyrics (from the lyricist step) for a composition.
func (q *Queries) UpdateGeneratedLyrics(ctx context.Context, id, lyrics string) error {
	_, err := q.pool.Exec(ctx,
		`UPDATE compositions SET generated_lyrics = $1, updated_at = now()
		 WHERE id = $2`,
		lyrics, id,
	)
	return err
}
