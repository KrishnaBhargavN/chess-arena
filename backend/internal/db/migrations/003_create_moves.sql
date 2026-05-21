CREATE TABLE IF NOT EXISTS moves (
    id         SERIAL      PRIMARY KEY,
    game_id    UUID        NOT NULL REFERENCES games(id),
    ply        INT         NOT NULL,
    san        TEXT        NOT NULL,
    uci        TEXT        NOT NULL,
    from_sq    TEXT        NOT NULL,
    to_sq      TEXT        NOT NULL,
    played_by  UUID        REFERENCES users(id),
    played_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS moves_game_id_ply ON moves(game_id, ply);
