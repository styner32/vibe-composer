package lyria

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/genai"
)

// Result holds the output from Lyria music generation.
type Result struct {
	AudioData []byte
	MIMEType  string
	Lyrics    string
}

// Client wraps the Gemini GenAI SDK for Lyria music generation.
type Client struct {
	genaiClient *genai.Client
}

// NewClient creates a new Lyria client.
func NewClient(genaiClient *genai.Client) *Client {
	return &Client{genaiClient: genaiClient}
}

// GenerateMusic generates music from a text prompt using Lyria 3 Pro.
func (c *Client) GenerateMusic(ctx context.Context, prompt string) (*Result, error) {
	slog.Info("generating music with Lyria 3 Pro", "prompt_length", len(prompt))

	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"AUDIO", "TEXT"},
	}

	result, err := c.genaiClient.Models.GenerateContent(
		ctx,
		"lyria-3-pro-preview",
		[]*genai.Content{
			{Parts: []*genai.Part{{Text: prompt}}},
		},
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("calling Lyria 3 Pro: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Lyria 3 Pro")
	}

	res := &Result{}
	for _, part := range result.Candidates[0].Content.Parts {
		if part.Text != "" {
			if res.Lyrics != "" {
				res.Lyrics += "\n"
			}
			res.Lyrics += part.Text
		} else if part.InlineData != nil {
			res.AudioData = part.InlineData.Data
			res.MIMEType = part.InlineData.MIMEType
		}
	}

	if len(res.AudioData) == 0 {
		return nil, fmt.Errorf("no audio data in Lyria response")
	}

	slog.Info("music generated", "audio_size", len(res.AudioData), "has_lyrics", res.Lyrics != "")
	return res, nil
}
