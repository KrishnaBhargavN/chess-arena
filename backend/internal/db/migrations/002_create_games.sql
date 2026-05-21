CREATE TABLE IF NOT EXISTS games (
    id             UUID        PRIMARY KEY,
    player_a_id    UUID        REFERENCES users(id),
    player_b_id    UUID        REFERENCES users(id),
    player_a_color TEXT,
    player_b_color TEXT,
    status         TEXT        NOT NULL DEFAULT 'waiting',
    current_fen    TEXT        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at    TIMESTAMPTZ
);
