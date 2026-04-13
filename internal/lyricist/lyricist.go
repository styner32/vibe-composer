package lyricist

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/vibe-composer/internal/analyzer"
	"google.golang.org/genai"
)

// Lyricist generates song lyrics from extracted emotional data using Gemini.
// It enforces the NSZ privacy contract: lyrics must never reveal the original story.
type Lyricist struct {
	client *genai.Client
}

// NewLyricist creates a new Lyricist.
func NewLyricist(client *genai.Client) *Lyricist {
	return &Lyricist{client: client}
}

const systemPrompt = `You are a lyricist for "No Scratch Zone," a platform that transforms
raw emotional venting into artistic creations.

Your job is to write song lyrics based on extracted emotional data —
never based on the original words the user wrote or said.

ABSOLUTE RULES
--------------
1. Do NOT use any word that appears in the original user input.
2. Do NOT mention names, job titles, or locations.
3. Do NOT describe what actually happened — emotion and atmosphere only.
4. The listener must feel the emotion without knowing the story.
5. Write ALL lyrics in the language specified in the "language" field.
   If language is "Korean", every line of lyrics MUST be in Korean.
   If language is "English", every line MUST be in English.
   NEVER translate to a different language. This is non-negotiable.

LYRIC TYPE INSTRUCTIONS
-----------------------
Type "arc" (기승전결):
  Verse 1 — voicing grievance / injustice
  Verse 2 — unleashing anger
  Verse 3 — transcendence / declaration of resilience

Type "immersion" (감정 만끽):
  No arc. Linger in the emotion.
  All events referenced through metaphor only.
  Prioritize nature imagery, symbolism, repetition.

METER
-----
If genre includes pansori or minyo:
  Follow 3·4 or 4·4 syllabic meter.
  Use exclamatory particles: 으이고, 허이, 에라, 아이고, 얼쑤.

If genre includes hip-hop or rap:
  End-rhyme each bar. Structure in 16-bar units.
  Hook = the emotional peak in one line.

If genre is funny/comedy:
  Exaggerate everything to absurd proportions.
  Turn the feeling into comedic catharsis.

If genre is harsh/metal:
  Raw, visceral imagery. Channel rage into empowerment.

OUTPUT FORMAT
-------------
Return only the lyrics. No explanation, no title header.
Label each section: [Verse 1] [Hook] [Verse 2] [Bridge] etc.`

// GenerateLyrics creates song lyrics from extracted emotional metadata.
// The original text is NOT passed — only the analyzed emotion data.
// lyricType should be "arc" or "immersion". If empty, it's auto-derived from style.
func (l *Lyricist) GenerateLyrics(ctx context.Context, analysis *analyzer.AudioAnalysis, style string, lyricType string) (string, error) {
	if lyricType == "" {
		lyricType = chooseLyricType(style)
	}
	keywords := strings.Join(analysis.AbstractKeywords, ", ")
	if keywords == "" {
		keywords = "(none extracted)"
	}

	language := analysis.Language
	if language == "" {
		language = "Korean" // default fallback
	}

	speechStyle := analysis.SpeechStyle
	if speechStyle == "" {
		speechStyle = "natural and conversational"
	}

	userPrompt := fmt.Sprintf(`Generate song lyrics from this emotional data:

- emotion_type: %s
- emotion_intensity: %d (scale 1-10)
- abstract_keywords: [%s]
- genre: %s
- lyric_type: %s
- language: %s
- speech_style: %s
- emotional_context: %s

CRITICAL: Write the ENTIRE lyrics in %s. Do NOT use any other language.
IMPORTANT: Match the speech_style above — if the person was raw and casual, the lyrics should feel raw and casual. If they were sarcastic, the lyrics should drip with sarcasm. Preserve their VOICE and ATTITUDE in the lyrics.
Remember: write ONLY from the emotional metadata above. The listener must feel the emotion without knowing the story.`,
		analysis.PrimaryEmotion,
		analysis.EmotionIntensity,
		keywords,
		style,
		lyricType,
		language,
		speechStyle,
		analysis.Summary,
		language,
	)

	slog.Info("generating lyrics",
		"emotion", analysis.PrimaryEmotion,
		"intensity", analysis.EmotionIntensity,
		"style", style,
		"lyric_type", lyricType,
		"keywords", keywords,
	)

	result, err := l.client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		[]*genai.Content{
			{Parts: []*genai.Part{{Text: userPrompt}}},
		},
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: systemPrompt}},
			},
			Temperature: genai.Ptr(float32(0.9)), // Higher creativity for lyrics
		},
	)
	if err != nil {
		return "", fmt.Errorf("calling Gemini for lyrics: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no lyrics generated from Gemini")
	}

	lyrics := result.Candidates[0].Content.Parts[0].Text
	slog.Info("lyrics generated", "length", len(lyrics))

	return lyrics, nil
}

// chooseLyricType provides a sensible default when the user doesn't specify.
func chooseLyricType(style string) string {
	switch style {
	case "hiphop":
		return "arc" // Hip-hop fits the 기승전결 narrative arc
	case "pansori":
		return "immersion" // Pansori fits emotional immersion
	case "harsh":
		return "arc" // Metal builds from rage to empowerment
	case "funny":
		return "arc" // Comedy benefits from a narrative arc
	default:
		return "arc"
	}
}
