CREATE TABLE polls (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    public_id   TEXT NOT NULL UNIQUE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_polls_public_id ON polls(public_id);

CREATE TABLE poll_options (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    poll_id  INTEGER NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    label    TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_poll_options_poll_id ON poll_options(poll_id);

CREATE TABLE votes (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    poll_id  INTEGER NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    name     TEXT NOT NULL,
    voted_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_votes_poll_id ON votes(poll_id);

CREATE TABLE vote_responses (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    vote_id   INTEGER NOT NULL REFERENCES votes(id) ON DELETE CASCADE,
    option_id INTEGER NOT NULL REFERENCES poll_options(id) ON DELETE CASCADE,
    available INTEGER NOT NULL DEFAULT 0,
    UNIQUE(vote_id, option_id)
);
CREATE INDEX idx_vote_responses_vote_id ON vote_responses(vote_id);
