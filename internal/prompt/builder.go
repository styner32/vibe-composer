package prompt

import (
	"fmt"
	"strings"

	"github.com/vibe-composer/internal/analyzer"
)

// Build creates a creative Lyria music prompt from the analyzed input.
func Build(analysis *analyzer.AudioAnalysis, style string, originalText string) string {
	switch style {
	case "harsh":
		return buildHarsh(analysis, originalText)
	default:
		return buildFunny(analysis, originalText)
	}
}

func buildFunny(a *analyzer.AudioAnalysis, originalText string) string {
	intensityDesc := intensityToFunnyDesc(a.EmotionIntensity)
	vocalContext := ""
	if len(a.VocalTraits) > 0 {
		vocalContext = fmt.Sprintf(" The person was %s while expressing this.", strings.Join(a.VocalTraits, " and "))
	}

	return fmt.Sprintf(`Create a hilarious, over-the-top comedy song about this situation.

Original text (USE THIS for lyrics — keep the same language):
"%s"

Emotion analysis: %s (intensity: %s).%s
Context: %s

Musical direction:
- Genre: comedic pop/folk with silly sound effects and playful instrumentation
- Vocals: exaggerated, theatrical, absurdly dramatic — like a Broadway actor having a meltdown over something trivial
- Lyrics: turn the anger into pure absurdist comedy — exaggerate everything to ridiculous proportions
- Make it catchy with a memorable chorus that satirizes the situation
- Include at least one spoken-word breakdown where the singer gets increasingly petty
- The overall tone should make someone who was angry about this burst out laughing
- About 2 minutes long
- IMPORTANT: Write the lyrics in the SAME LANGUAGE as the "Original text" above. Do NOT translate to English. If the input is in Korean, the lyrics MUST be in Korean.

The angrier the original feeling, the funnier and more ridiculous the song should be.`,
		originalText,
		a.PrimaryEmotion,
		intensityDesc,
		vocalContext,
		a.Summary,
	)
}

func buildHarsh(a *analyzer.AudioAnalysis, originalText string) string {
	intensityDesc := intensityToHarshDesc(a.EmotionIntensity)
	vocalContext := ""
	if len(a.VocalTraits) > 0 {
		vocalContext = fmt.Sprintf(" The person was %s.", strings.Join(a.VocalTraits, " and "))
	}

	return fmt.Sprintf(`Create an intense, cathartic, powerful song about this situation.

Original text (USE THIS for lyrics — keep the same language):
"%s"

Emotion analysis: %s (intensity: %s).%s
Context: %s

Musical direction:
- Genre: aggressive rock/metal with heavy distortion, pounding drums, and raw energy
- Vocals: powerful, raw, screaming catharsis — channel every ounce of rage
- Lyrics: turn the pain into empowerment, transform victimhood into strength
- Build from seething verses to an explosive chorus
- Include a bridge that's a moment of eerie calm before the final eruption
- The overall message: "I will not be broken by this"
- About 2 minutes long
- IMPORTANT: Write the lyrics in the SAME LANGUAGE as the "Original text" above. Do NOT translate to English. If the input is in Korean, the lyrics MUST be in Korean.

This song should feel like screaming into a void and hearing the void scream back in solidarity.`,
		originalText,
		a.PrimaryEmotion,
		intensityDesc,
		vocalContext,
		a.Summary,
	)
}

func intensityToFunnyDesc(intensity int) string {
	switch {
	case intensity >= 9:
		return "volcanic, absolutely nuclear-level"
	case intensity >= 7:
		return "seriously heated, steam-coming-out-of-ears level"
	case intensity >= 5:
		return "notably irritated, eye-twitching level"
	case intensity >= 3:
		return "mildly annoyed, passive-aggressive level"
	default:
		return "slightly peeved, heavy-sighing level"
	}
}

func intensityToHarshDesc(intensity int) string {
	switch {
	case intensity >= 9:
		return "white-hot, apocalyptic fury"
	case intensity >= 7:
		return "burning, barely-contained rage"
	case intensity >= 5:
		return "simmering, dark resentment"
	case intensity >= 3:
		return "cold, calculated bitterness"
	default:
		return "quiet, smoldering discontent"
	}
}
