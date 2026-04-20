package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/genai"
)

// AudioAnalysis holds the result of analyzing an audio recording.
type AudioAnalysis struct {
	Transcription    string   `json:"transcription"`
	PrimaryEmotion   string   `json:"primary_emotion"`
	EmotionIntensity int      `json:"emotion_intensity"`
	VocalTraits      []string `json:"vocal_traits"`
	AbstractKeywords []string `json:"abstract_keywords"`
	SpeechStyle      string   `json:"speech_style"`
	Language         string   `json:"language"`
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
  "abstract_keywords": ["metaphorical", "words", "that", "capture", "the", "feeling"],
  "speech_style": "describe how the person speaks — their tone, register, and attitude",
  "language": "Korean",
  "summary": "brief description of the emotional context and what they're upset about"
}

Rules:
- emotion_intensity is 1-10 (1=mild, 10=explosive rage)
- vocal_traits should describe HOW they sound, not WHAT they say
- transcription should be as accurate as possible
- abstract_keywords should be 3-6 metaphorical or symbolic words that capture the FEELING, not the literal situation. Example: "my boss ignored me" → ["invisibility", "echo", "glass wall", "hollow room"]. Do NOT use any words from the original input.
- speech_style should capture the person's WAY of speaking, NOT what they said. Examples: "raw and casual, uses slang and profanity, speaks in 반말", "dry and sarcastic, deadpan delivery", "explosive and dramatic, exaggerates everything", "quiet and bitter, understated but cutting". This should describe their tone, register (formal/informal), attitude, and personality.
- language should be the language the person is speaking in (e.g. "Korean", "English", "Japanese", "Spanish")
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

	cleaned := cleanJSON(text)
	var analysis AudioAnalysis
	if err := json.Unmarshal([]byte(cleaned), &analysis); err != nil {
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
  "abstract_keywords": ["metaphorical", "words", "that", "capture", "the", "feeling"],
  "speech_style": "describe how the person writes — their tone, register, and attitude",
  "language": "Korean",
  "summary": "brief description of what they're upset about"
}

Rules:
- emotion_intensity is 1-10 (1=mild, 10=explosive rage)
- vocal_traits should be empty for text input
- abstract_keywords should be 3-6 metaphorical or symbolic words that capture the FEELING, not the literal situation. Example: "my boss ignored me" → ["invisibility", "echo", "glass wall", "hollow room"]. Do NOT use any words from the original input.
- speech_style should capture the person's WAY of writing, NOT what they wrote. Examples: "raw and casual, uses slang and profanity, writes in 반말", "dry and sarcastic, deadpan tone", "explosive and dramatic, exaggerates everything", "quiet and bitter, understated but cutting", "crude street talk, no filter". This should describe their tone, register (formal/informal/반말/존댓말), attitude, and personality.
- language should be the language the text is written in (e.g. "Korean", "English", "Japanese", "Spanish")
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

	cleaned := cleanJSON(respText)
	var analysis AudioAnalysis
	if err := json.Unmarshal([]byte(cleaned), &analysis); err != nil {
		return nil, fmt.Errorf("parsing analysis JSON: %w (raw: %s)", err, respText)
	}

	return &analysis, nil
}

// cleanJSON extracts the first valid JSON object from a string.
// Gemini sometimes returns trailing characters (extra braces, whitespace)
// after the JSON object, causing parse failures.
func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	// Strip markdown code fences if present
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}

	// Find the first '{' and its matching '}'
	start := strings.Index(s, "{")
	if start == -1 {
		return s
	}

	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inString {
			escaped = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if c == '{' {
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}

	return s[start:]
}
