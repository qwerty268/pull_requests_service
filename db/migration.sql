CREATE TABLE IF NOT EXISTS "user" (
    user_id   TEXT PRIMARY KEY,
    username  TEXT UNIQUE NOT NULL,
    team_name TEXT,
    is_active BOOLEAN DEFAULT false
);

CREATE TABLE IF NOT EXISTS team (
    team_name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS team_user_map (
    team_name TEXT REFERENCES team(team_name)  ON DELETE CASCADE,
    user_id TEXT REFERENCES "user"(user_id)  ON DELETE CASCADE,
    PRIMARY KEY (team_name, user_id)
);

CREATE TABLE IF NOT EXISTS pull_request (
    pull_request_id     TEXT PRIMARY KEY,
    pull_request_name   TEXT NOT NULL,
    author_id           TEXT REFERENCES "user"(user_id),
    is_merged           BOOLEAN DEFAULT FALSE NOT NULL,
    assigned_reviewers  TEXT[],
    created_at          TIMESTAMPTZ NOT NULL,
    merged_at           TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS pr_reviewers_map (
    pull_request_id TEXT REFERENCES pull_request(pull_request_id) ON DELETE CASCADE,
    user_id TEXT REFERENCES "user"(user_id) ON DELETE CASCADE
);
