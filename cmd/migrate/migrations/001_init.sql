CREATE TABLE IF NOT EXISTS compositions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    input_type      TEXT NOT NULL,
    input_text      TEXT,
    input_audio_url TEXT,
    emotion         TEXT,
    music_style     TEXT NOT NULL DEFAULT 'funny',
    music_prompt    TEXT,
    result_url      TEXT,
    result_lyrics   TEXT,
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_compositions_username ON compositions(username);
