package prompt

import (
	"fmt"
	"strings"

	"github.com/vibe-composer/internal/analyzer"
)

// Build creates a creative Lyria music prompt from the analyzed input.
func Build(analysis *analyzer.AudioAnalysis, style string, voiceGender string, originalText string) string {
	var base string
	switch style {
	case "harsh":
		base = buildHarsh(analysis, originalText)
	case "hiphop":
		base = buildHiphop(analysis, originalText)
	case "pansori":
		base = buildPansori(analysis, originalText)
	default:
		base = buildFunny(analysis, originalText)
	}
	return base + voiceDirective(voiceGender)
}

// voiceDirective returns a prompt suffix for the requested voice gender.
func voiceDirective(gender string) string {
	switch gender {
	case "male":
		return "\n\nVOCAL DIRECTION: Use a MALE vocalist voice. The singer must be clearly male-sounding."
	case "female":
		return "\n\nVOCAL DIRECTION: Use a FEMALE vocalist voice. The singer must be clearly female-sounding."
	default:
		return ""
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

func buildHiphop(a *analyzer.AudioAnalysis, originalText string) string {
	intensityDesc := intensityToHiphopDesc(a.EmotionIntensity)
	vocalContext := ""
	if len(a.VocalTraits) > 0 {
		vocalContext = fmt.Sprintf(" The person was %s while expressing this.", strings.Join(a.VocalTraits, " and "))
	}

	return fmt.Sprintf(`Create a hard-hitting hip-hop track about this situation.

Original text (USE THIS for lyrics — keep the same language):
"%s"

Emotion analysis: %s (intensity: %s).%s
Context: %s

Musical direction:
- Genre: modern hip-hop — blend of trap beats, boom-bap drums, heavy 808 bass, and atmospheric synths
- Flow: confident, sharp bars with rhythmic switches — channel the frustration into swagger and dominance
- Lyrics: transform the anger into braggadocious diss-track energy — roast the source of frustration with clever wordplay, punchlines, and metaphors
- Include a catchy hook/chorus that slaps hard
- Build intensity: start with a menacing verse, escalate to an aggressive chorus, drop a fire bridge with double-time flow
- Add ad-libs ("yeah!", "let's go!", "uh!") where they hit hardest
- The overall vibe: the person went from being upset to absolutely owning the situation
- About 2 minutes long
- IMPORTANT: Write the lyrics in the SAME LANGUAGE as the "Original text" above. Do NOT translate to English. If the input is in Korean, the lyrics MUST be in Korean.

This track should make the listener feel like they just dropped the hardest diss track of the year.`,
		originalText,
		a.PrimaryEmotion,
		intensityDesc,
		vocalContext,
		a.Summary,
	)
}

func buildPansori(a *analyzer.AudioAnalysis, originalText string) string {
	intensityDesc := intensityToPansoriDesc(a.EmotionIntensity)
	vocalContext := ""
	if len(a.VocalTraits) > 0 {
		vocalContext = fmt.Sprintf(" The person was %s while expressing this.", strings.Join(a.VocalTraits, " and "))
	}

	return fmt.Sprintf(`Create a 판소리 (pansori) inspired dramatic musical piece about this situation.

Original text (USE THIS for lyrics — keep the same language):
"%s"

Emotion analysis: %s (intensity: %s).%s
Context: %s

Musical direction:
- Genre: 판소리 (Korean traditional pansori) — a solo vocalist performing dramatic narrative singing accompanied by a buk (barrel drum) providing rhythmic punctuation
- Vocals: use the distinctive pansori vocal techniques — 창 (chang, the singing), 아니리 (aniri, spoken narrative passages), and 추임새 (chuimsae, exclamatory reactions like "얼씨구!", "좋다!", "그렇지!")
- Singing style: deep, raw, guttural 수리성 (suriseong) voice with dramatic shifts between 우조 (ujo, majestic/solemn) and 계면조 (gyemyeonjo, sorrowful/lamenting) modes
- Structure: begin with an 아니리 (spoken narrative intro) setting the scene, then launch into passionate 창 (singing) passages that tell the story with theatrical emotional peaks
- Rhythm: use traditional 장단 (jangdan) rhythmic patterns — 중모리 for storytelling, 자진모리 for building tension, 휘모리 for climactic moments
- The overall feeling: a grand, theatrical retelling of the person's grievance as if it were an epic tale from Korean folklore — their frustration elevated to legendary proportions
- About 2 minutes long
- IMPORTANT: Write the lyrics PRIMARILY in Korean (한국어). This is a Korean traditional genre, so Korean lyrics are essential. If the original text is in another language, translate the sentiment into Korean for the lyrics while preserving the emotional core.

This should feel like a master 광대 (gwangdae, pansori performer) is dramatically recounting this person's woes to a spellbound audience, turning their everyday frustration into an epic Korean saga.`,
		originalText,
		a.PrimaryEmotion,
		intensityDesc,
		vocalContext,
		a.Summary,
	)
}

func intensityToHiphopDesc(intensity int) string {
	switch {
	case intensity >= 9:
		return "absolutely unhinged, ready-to-drop-the-mic level"
	case intensity >= 7:
		return "fired up, battle-rap energy"
	case intensity >= 5:
		return "heated, throwing-shade level"
	case intensity >= 3:
		return "cool and calculated, sly-diss level"
	default:
		return "chill but petty, subliminal-bars level"
	}
}

func intensityToPansoriDesc(intensity int) string {
	switch {
	case intensity >= 9:
		return "천지가 무너지는 (heaven-and-earth-shattering) fury"
	case intensity >= 7:
		return "한이 맺힌 (deeply resentful) anguish"
	case intensity >= 5:
		return "가슴이 답답한 (chest-tightening) frustration"
	case intensity >= 3:
		return "쓸쓸한 (melancholic) bitterness"
	default:
		return "한숨 섞인 (sigh-laden) discontent"
	}
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
