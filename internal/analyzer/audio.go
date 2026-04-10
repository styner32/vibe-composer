package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"google.golang.org/genai"
)

// AudioAnalysis holds the result of analyzing an audio recording.
type AudioAnalysis struct {
	Transcription    string   `json:"transcription"`
	PrimaryEmotion   string   `json:"primary_emotion"`
	EmotionIntensity int      `json:"emotion_intensity"`
	VocalTraits      []string `json:"vocal_traits"`
	Summary          string   `json:"summary"`
}

// Analyzer uses Gemini Flash to analyze audio for transcription and emotion.
type Analyzer struct {
	client *genai.Client
}

// NewAnalyzer creates a new audio analyzer.
func NewAnalyzer(client *genai.Client) *Analyzer {
	return &Analyzer{client: client}
}

const audioAnalysisPrompt = `Analyze this audio recording of a person expressing negative feelings or frustration.
Return a JSON object with these exact fields:
{
  "transcription": "exact text of what the person is saying",
  "primary_emotion": "one of: anger, frustration, sadness, disgust, anxiety, contempt, resentment",
  "emotion_intensity": 7,
  "vocal_traits": ["shouting", "sarcastic", "crying", "whispering", "trembling", "aggressive"],
  "summary": "brief description of the emotional context and what they're upset about"
}

Rules:
- emotion_intensity is 1-10 (1=mild, 10=explosive rage)
- vocal_traits should describe HOW they sound, not WHAT they say
- transcription should be as accurate as possible
- Only return valid JSON, no markdown formatting`

// AnalyzeAudio sends audio data to Gemini Flash for transcription and emotion detection.
func (a *Analyzer) AnalyzeAudio(ctx context.Context, audioData []byte, mimeType string) (*AudioAnalysis, error) {
	slog.Info("analyzing audio", "size", len(audioData), "mimeType", mimeType)

	result, err := a.client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		[]*genai.Content{
			{
				Parts: []*genai.Part{
					{Text: audioAnalysisPrompt},
					{InlineData: &genai.Blob{
						Data:     audioData,
						MIMEType: mimeType,
					}},
				},
			},
		},
		&genai.GenerateContentConfig{
			Temperature:      genai.Ptr(float32(0.2)),
			ResponseMIMEType: "application/json",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("calling Gemini Flash: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini Flash")
	}

	text := result.Candidates[0].Content.Parts[0].Text
	slog.Info("audio analysis response", "raw", text)

	var analysis AudioAnalysis
	if err := json.Unmarshal([]byte(text), &analysis); err != nil {
		return nil, fmt.Errorf("parsing analysis JSON: %w (raw: %s)", err, text)
	}

	return &analysis, nil
}

// AnalyzeText uses Gemini Flash to detect emotion from text input.
func (a *Analyzer) AnalyzeText(ctx context.Context, text string) (*AudioAnalysis, error) {
	slog.Info("analyzing text", "length", len(text))

	prompt := fmt.Sprintf(`Analyze this text from a person expressing negative feelings:

"%s"

Return a JSON object with these exact fields:
{
  "transcription": "the original text as-is",
  "primary_emotion": "one of: anger, frustration, sadness, disgust, anxiety, contempt, resentment",
  "emotion_intensity": 7,
  "vocal_traits": [],
  "summary": "brief description of what they're upset about"
}

Rules:
- emotion_intensity is 1-10 (1=mild, 10=explosive rage)
- vocal_traits should be empty for text input
- Only return valid JSON, no markdown formatting`, text)

	result, err := a.client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		[]*genai.Content{
			{Parts: []*genai.Part{{Text: prompt}}},
		},
		&genai.GenerateContentConfig{
			Temperature:      genai.Ptr(float32(0.2)),
			ResponseMIMEType: "application/json",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("calling Gemini Flash: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini Flash")
	}

	respText := result.Candidates[0].Content.Parts[0].Text

	var analysis AudioAnalysis
	if err := json.Unmarshal([]byte(respText), &analysis); err != nil {
		return nil, fmt.Errorf("parsing analysis JSON: %w (raw: %s)", err, respText)
	}

	return &analysis, nil
}
