# Vibe Composer 🎵

> Turn your rage into a banger — AI-powered music from your negative feelings

Vibe Composer takes your frustrations, anger, and negative emotions — as text or audio — and transforms them into hilarious comedy songs or cathartic heavy metal anthems using **Lyria 3 Pro**.

## Features

- 📝 **Text input**: Type out your frustrations
- 🎤 **Audio input**: Record or upload an angry voice memo
- 🧠 **Emotion detection**: AI analyzes vocal tone, intensity, and emotional context
- 😂 **Funny mode**: Transforms rage into absurdist comedy songs
- 🤘 **Harsh mode**: Channels anger into cathartic metal anthems
- 🔒 **Invite-only**: Only users with invited usernames can access
- 🚫 **No parallel generation**: One composition at a time per user

## Tech Stack

- **Backend**: Go + chi router + pgx (PostgreSQL)
- **AI**: Google Gemini API (Lyria 3 Pro + Gemini Flash)
- **Storage**: Google Cloud Storage
- **Frontend**: Vanilla HTML/CSS/JS (dark glassmorphism theme)

## Setup

### 1. Prerequisites
- Go 1.21+
- Docker (for PostgreSQL)
- Google Cloud account with:
  - Gemini API key
  - GCS bucket created

### 2. Start PostgreSQL
```bash
make db-up
```

### 3. Configure Environment
```bash
cp .env.example .env
# Edit .env with your credentials
```

### 4. Run the Server
```bash
# Source your env vars
export $(cat .env | xargs)
make run
```

### 5. Open in Browser
Navigate to `http://localhost:8080` and login with your invited username.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/me` | Current user info |
| `POST` | `/api/compose` | Submit text/audio for music generation |
| `GET` | `/api/compositions` | List user's compositions |
| `GET` | `/api/compositions/:id` | Get composition details |
| `GET` | `/api/compositions/:id/download` | Download generated music |

## How It Works

1. User submits text or audio expressing negative feelings
2. **Gemini Flash** analyzes the input — transcribes audio, detects emotion (anger, frustration, etc.), measures intensity (1-10), identifies vocal traits (shouting, sarcastic, etc.)
3. **Prompt Builder** crafts a creative music prompt based on the analysis + chosen style (funny/harsh)
4. **Lyria 3 Pro** generates a full-length song (~2 min) with vocals, lyrics, and instrumentation
5. Result is uploaded to GCS and made available for streaming/download
