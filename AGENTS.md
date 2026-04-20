# AGENTS.md — Vibe Composer (No Scratch Zone)

> **NSZ = No Scratch Zone**: A platform that transforms raw emotional venting
> into artistic music — the listener feels the emotion without knowing the story.

---

## 1. Core Principle

### What the app does

Users submit raw negative feelings (text or audio) — frustration about a coworker,
anger at a situation, exhaustion from daily life. The system transforms those feelings
into a full-length song (~2 min) with vocals, lyrics, and instrumentation.

### The privacy contract

The generated lyrics must **never** reveal the original story. This is non-negotiable.

| Extract | Discard |
|---------|---------|
| Emotion type (resentment, anger, exhaustion, loneliness) | Names, job titles, locations |
| Emotion intensity (scale 1–5) | Specific descriptions of what happened |
| Abstract keywords only — NOT the literal words | Every word that appears verbatim in the original input |

**Example transformation:**
```
"my boss ignored me" → ["invisibility", "echo", "disappearing"]
```

The goal: a listener should **feel the emotion** without knowing the story.

---

## 2. Lyric Types

### Type 1 — Arc (기승전결)

A narrative arc across 3 verses:

| Section | Role |
|---------|------|
| Verse 1 | Voicing grievance / injustice |
| Verse 2 | Unleashing anger |
| Verse 3 | Transcendence — "I will rise" / "I have already won" |

Best fit: hip-hop beat + pansori vocal, rap verse + pansori hook

### Type 2 — Immersion (감정 만끽)

No narrative arc. Pure emotional immersion:

- Events referenced only through metaphor — **never described directly**
- Emotion is explored slowly, lingered in, savored
- Heavy use of nature imagery and symbolism

Best fit: minyo (민요) melody, pansori 창, traditional vocal style

---

## 3. Rhythm & Meter Rules

### For hip-hop / rap lyrics

- End-rhyme on the final syllable of each bar
- Structure in 16-bar units
- Hook = the emotional peak compressed into a single line
- Use internal rhyme and alliteration where possible

### For pansori / minyo lyrics

- Follow **3·4 or 4·4 syllabic meter** (inherited from Korean folk poetry)
  - Example 3·4: `에이 모르겠다 / 다 버리고파`
  - Example 4·4: `하늘이 높아 / 땅도 깊어라`
- Use exclamatory particles for emotional emphasis:
  `으이고`, `허이`, `에라`, `아이고`, `얼쑤`
- Build intensity through repetition with variation

### Genre hybrid combinations

| Combo | Hook | Verses |
|-------|------|--------|
| Hip-hop beat + pansori vocal | pansori 창 | rap bars |
| Minyo base + rap | 창 (climax only) | contemporary rap or sung melody |

---

## 4. Project Structure

```
vibe-composer/
├── AGENTS.md                          # This file — project spec & guidelines
├── README.md                          # User-facing readme
├── Makefile                           # Build & dev commands
├── Dockerfile                         # Multi-stage Go build
├── docker-compose.yml                 # Local PostgreSQL
├── go.mod / go.sum                    # Go module dependencies
├── .env.example                       # Environment variable template
│
├── cmd/
│   ├── server/
│   │   └── main.go                    # Server entrypoint — wires all dependencies
│   └── migrate/
│       ├── main.go                    # DB migration runner (embeds SQL files)
│       └── migrations/
│           ├── 001_init.sql           # Initial schema (compositions table)
│           ├── 002_add_voice_gender.sql
│           ├── 003_add_generated_lyrics.sql
│           └── 004_add_lyric_type.sql
│
├── internal/
│   ├── analyzer/
│   │   └── audio.go                   # Gemini Flash — emotion detection, transcription, abstract keyword extraction
│   ├── auth/
│   │   └── basic.go                   # Basic auth middleware, username context extraction
│   ├── config/
│   │   └── config.go                  # Env-based config loading, allowed user list
│   ├── db/
│   │   ├── db.go                      # PostgreSQL connection pool (pgx)
│   │   ├── queries.go                 # Composition CRUD operations
│   │   └── migrations/                # Reference copy of migrations (not used at runtime)
│   │       ├── 001_init.sql
│   │       └── 002_add_voice_gender.sql
│   ├── handler/
│   │   └── handler.go                 # HTTP API handlers (chi router)
│   ├── lyria/
│   │   └── client.go                  # Lyria 3 Pro music generation client
│   ├── lyricist/
│   │   └── lyricist.go                # Gemini Flash — privacy-safe lyrics generation from emotion data
│   ├── prompt/
│   │   └── builder.go                 # Builds Lyria music prompts from pre-generated lyrics + analysis
│   ├── storage/
│   │   └── gcs.go                     # Google Cloud Storage upload/download
│   └── worker/
│       └── composer.go                # Background job processing (in-memory channel queue)
│
├── web/
│   ├── index.html                     # Single-page app shell
│   ├── style.css                      # Design system — dark glassmorphism theme
│   └── app.js                         # Frontend logic — auth, compose, library, player
│
└── infra/
    ├── main.tf                        # Terraform provider config
    ├── variables.tf                   # Input variables
    ├── terraform.tfvars.example       # Example variable values
    ├── cloud_run.tf                   # Cloud Run service definition
    ├── cloud_sql.tf                   # Cloud SQL PostgreSQL instance
    ├── storage.tf                     # GCS bucket
    ├── iam.tf                         # Service account & permissions
    ├── networking.tf                  # VPC connector for Cloud SQL
    ├── secrets.tf                     # Secret Manager configs
    └── outputs.tf                     # Terraform outputs
```

---

## 5. Architecture Overview

```
┌──────────────┐    ┌──────────────┐    ┌─────────────────────┐
│   Frontend   │───▶│   Handler    │───▶│   Composer Worker   │
│  (web/)      │    │  (handler/)  │    │   (worker/)         │
│  HTML/JS/CSS │    │  REST API    │    │                     │
└──────────────┘    └──────┬───────┘    │  ┌───────────────┐  │
                           │            │  │ 1. Analyzer    │  │
                    ┌──────▼───────┐    │  │  (Gemini)     │  │
                    │   Database   │    │  ├───────────────┤  │
                    │   (db/)      │◄───│  │ 2. Lyricist   │  │
                    │  PostgreSQL  │    │  │  (Gemini)     │  │
                    └──────────────┘    │  ├───────────────┤  │
                                       │  │ 3. Prompt     │  │
                                       │  │  Builder      │  │
                                       │  ├───────────────┤  │
                                       │  │ 4. Lyria 3    │  │
                                       │  │  Pro          │  │
                                       │  └───────┬───────┘  │
                                       └──────────│──────────┘
                                                  │
                                           ┌──────▼──────┐
                                           │    GCS      │
                                           │  (storage/) │
                                           └─────────────┘
```

### Package responsibilities

| Package | Path | Responsibility |
|---------|------|---------------|
| `auth` | `internal/auth/` | Basic auth middleware, username extraction from context |
| `config` | `internal/config/` | Env-based config loading, allowed user list |
| `db` | `internal/db/` | PostgreSQL queries (pgx), composition CRUD |
| `handler` | `internal/handler/` | HTTP API handlers (chi router) |
| `analyzer` | `internal/analyzer/` | Gemini Flash — emotion detection, transcription, abstract keyword extraction |
| `lyricist` | `internal/lyricist/` | Gemini Flash — generates privacy-safe lyrics from extracted emotion data |
| `prompt` | `internal/prompt/` | Builds Lyria music prompts from pre-generated lyrics + analysis |
| `lyria` | `internal/lyria/` | Lyria 3 Pro music generation client |
| `storage` | `internal/storage/` | GCS upload/download |
| `worker` | `internal/worker/` | Background job processing (in-memory channel queue) |
| `web` | `web/` | Frontend — vanilla HTML/CSS/JS, dark glassmorphism theme |

### Request flow

1. User submits text or audio via `POST /api/compose`
2. **Handler** validates input, creates DB record, uploads audio to GCS if needed
3. **Composer worker** picks up the job:
   - **Step 1 — Analyzer** (Gemini Flash) → transcribes audio + detects emotion, intensity, abstract keywords, vocal traits
   - **Step 2 — Lyricist** (Gemini Flash) → generates privacy-safe lyrics from extracted metadata only (original text is discarded)
   - **Step 3 — Prompt Builder** → crafts a Lyria music prompt using generated lyrics + analysis
   - **Step 4 — Lyria 3 Pro** → generates full-length song with vocals
4. Generated lyrics saved to DB immediately (visible to user while music generates)
5. Final audio uploaded to GCS, DB updated with status `done`

### Privacy pipeline

```
Raw user input
    ↓
┌─────────────────────────────────┐
│ Step 1: Emotion Extraction      │
│ (analyzer — Gemini Flash)       │
│                                 │
│ Input: raw text / audio         │
│ Output:                         │
│   - emotion_type                │
│   - emotion_intensity (1-10)    │
│   - abstract_keywords[]         │
│   - vocal_traits[]              │
│   - summary                     │
└────────────┬────────────────────┘
             ↓
   ⛔ Original text is DISCARDED here
             ↓
┌─────────────────────────────────┐
│ Step 2: Lyric Generation        │
│ (lyricist — Gemini Flash)       │
│                                 │
│ Input: extracted metadata ONLY  │
│ Output: privacy-safe lyrics     │
│ (saved to DB as generated_lyrics)│
└────────────┬────────────────────┘
             ↓
┌─────────────────────────────────┐
│ Step 3: Music Generation        │
│ (prompt builder + Lyria 3 Pro)  │
│                                 │
│ Input: generated lyrics + style │
│ Output: full song with vocals   │
└─────────────────────────────────┘
```

---

## 6. AI System Prompt — Lyric Generation

The lyricist system prompt lives in `internal/lyricist/lyricist.go`:

```
You are a lyricist for "No Scratch Zone," a platform that transforms
raw emotional venting into artistic creations.

Your job is to write song lyrics based on extracted emotional data —
never based on the original words the user wrote or said.

INPUT FORMAT
------------
- emotion_type: {e.g. resentment, anger, exhaustion}
- emotion_intensity: {1–5}
- abstract_keywords: {metaphor seeds extracted from raw input}
- genre: {hiphop+pansori | minyo+rap | funny | harsh}
- lyric_type: {1 = arc | 2 = immersion}

ABSOLUTE RULES
--------------
1. Do NOT use any word that appears in the original user input.
2. Do NOT mention names, job titles, or locations.
3. Do NOT describe what actually happened — emotion and atmosphere only.
4. The listener must feel the emotion without knowing the story.

LYRIC TYPE INSTRUCTIONS
-----------------------
Type 1 (Arc):
  Verse 1 — voicing grievance
  Verse 2 — unleashing anger
  Verse 3 — transcendence / declaration of resilience

Type 2 (Immersion):
  No arc. Linger in the emotion.
  All events referenced through metaphor only.
  Prioritize nature imagery, symbolism, repetition.

METER
-----
If genre includes pansori or minyo:
  Follow 3·4 or 4·4 syllabic meter.
  Use exclamatory particles: 으이고, 허이, 에라, 아이고.

If genre includes hip-hop or rap:
  End-rhyme each bar. Structure in 16-bar units.
  Hook = the emotional peak in one line.

OUTPUT FORMAT
-------------
Return only the lyrics. No explanation, no title header.
Label each section: [Verse 1] [Hook] [Verse 2] [Bridge] etc.
```

---

## 7. Example Input → Output

### Raw user input (never passed to AI)

```
오늘 팀장이 회의에서 내 의견을 완전히 무시했어.
세 번이나 말했는데 그냥 넘어가더니,
다른 사람이 똑같은 말 하니까 좋은 아이디어라고.
```

### Extraction step (what gets passed to AI)

```json
{
  "emotion_type": "resentment, invisibility",
  "emotion_intensity": 4,
  "abstract_keywords": ["echo", "glass wall", "disappearing voice", "hollow room"],
  "genre": "hiphop+pansori",
  "lyric_type": 1
}
```

### Expected output style

```
[Verse 1 — grievance]
벽에 대고 말하는 기분 / 소리가 흩어져
내 목소리 유리 속에 갇혀 / 아무도 못 들어

[Hook — pansori 창]
으이고, 허이 — 메아리도 없는 이 방
에라, 내 소리 들어라

[Verse 2 — anger]
...

[Verse 3 — transcendence]
...
```

---

## 8. Current Music Styles

| Style | Key | Description |
|-------|-----|-------------|
| Comedy | `funny` | Absurdist comedy songs — exaggerated theatrical vocals |
| Metal | `harsh` | Cathartic heavy rock/metal — raw screaming energy |
| Hip-hop | `hiphop` | Trap beats, 808 bass, diss-track energy |
| Pansori | `pansori` | Korean traditional dramatic singing with buk drum |
| Dark Drill | `drill` | Cold-fury UK Drill — menacing warning with sliding 808s |
| Jazz-hop | `jazzhop` | Sarcastic lo-fi jazz-hop — cynical burnout with dark humor |
| Epic Pansori | `epic_pansori` | Cinematic orchestral pansori — legendary-scale wrath |
| K-Swagger | `kswagger` | Pansori × Trap fusion — NSZ signature style, ultimate dominance |

Each style has its own prompt builder function in `internal/prompt/builder.go`.

---

## 9. Database Schema

### `compositions` table

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID (PK) | Auto-generated |
| `username` | TEXT | Authenticated user |
| `status` | TEXT | `pending` → `generating` → `done` / `failed` |
| `input_type` | TEXT | `text` or `audio` |
| `input_text` | TEXT | User's raw text (or transcription for audio) |
| `input_audio_url` | TEXT | GCS path for uploaded audio |
| `emotion` | TEXT | JSON blob from analyzer |
| `music_style` | TEXT | `funny`, `harsh`, `hiphop`, `pansori`, `drill`, `jazzhop`, `epic_pansori`, `kswagger` |
| `voice_gender` | TEXT | `male`, `female`, `any` |
| `lyric_type` | TEXT | `arc`, `immersion` |
| `music_prompt` | TEXT | Final prompt sent to Lyria |
| `generated_lyrics` | TEXT | Privacy-safe lyrics from the lyricist step |
| `result_url` | TEXT | GCS path for generated song |
| `result_lyrics` | TEXT | Lyrics returned by Lyria |
| `error_message` | TEXT | Error details if `status = failed` |
| `created_at` | TIMESTAMPTZ | Record creation time |
| `updated_at` | TIMESTAMPTZ | Last update time |

### Migrations

Migrations live in `cmd/migrate/migrations/` and are embedded via `//go:embed`.
They run sequentially by filename on `make migrate`.

---

## 10. API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/me` | Current authenticated user info |
| `POST` | `/api/compose` | Submit text/audio for music generation |
| `GET` | `/api/compositions` | List user's compositions (most recent first, limit 50) |
| `GET` | `/api/compositions/{id}` | Get single composition details |
| `GET` | `/api/compositions/{id}/download` | Stream/download generated music file |

### `POST /api/compose` form fields

| Field | Type | Required | Values |
|-------|------|----------|--------|
| `text` | string | if no audio | Raw text input |
| `audio` | file | if no text | Audio file (max 25MB) |
| `style` | string | no | `funny` (default), `harsh`, `hiphop`, `pansori`, `drill`, `jazzhop`, `epic_pansori`, `kswagger` |
| `voice` | string | no | `any` (default), `male`, `female` |
| `lyric_type` | string | no | `arc` (default), `immersion` |

---

## 11. Development Guidelines

### Running locally

```bash
make db-up          # Start PostgreSQL
cp .env.example .env  # Configure credentials
export $(cat .env | xargs)
make migrate        # Apply DB migrations
make run            # Start server at :8080
```

### Auth

- Invite-only: usernames configured via `ALLOWED_USERS` env var (comma-separated)
- Basic auth with shared password (`AUTH_PASSWORD`)
- One active generation per user at a time

### Key invariants

1. **Privacy first** — Generated lyrics must never reveal the original story
2. **Language preservation** — Lyrics must be in the same language as user input
3. **One-at-a-time** — Users cannot start a new generation while one is in progress
4. **Structured logging** — Use `log/slog` throughout, not `fmt.Println` or `log`

### Infrastructure

- **Runtime**: Google Cloud Run
- **Database**: Cloud SQL (PostgreSQL)
- **Storage**: Google Cloud Storage
- **IaC**: Terraform (see `infra/`)
- **Container**: Multi-stage Dockerfile with Go build

---

## 12. Future Roadmap

### Personalized 말투 (Speech Style Learning)

Currently, `speech_style` is extracted per-request from the user's input and discarded after
lyrics generation. In the future, the system should **learn each user's 말투 over time** to
produce lyrics that consistently match their personality across all compositions.

**Planned approach:**

1. **Accumulate speech profiles** — After each emotion extraction, store the `speech_style`
   result in a per-user profile table (e.g. `user_speech_profiles`).
2. **Build a composite voice** — Aggregate multiple `speech_style` entries into a stable
   user voice description (e.g. "consistently casual 반말, sarcastic, uses internet slang,
   occasionally dramatic").
3. **Inject profile into lyricist** — Pass the learned voice profile alongside per-request
   `speech_style` so lyrics feel like "them" even on short inputs where tone is ambiguous.
4. **Allow manual override** — Users may want to set or adjust their 말투 profile in settings.

**Privacy note:** The speech profile stores *how* they speak (tone, register, attitude),
never *what* they said. This is consistent with the NSZ contract.
