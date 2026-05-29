CREATE TABLE IF NOT EXISTS category (
    id  SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS rule (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    category_id INTEGER NOT NULL,
    CONSTRAINT fk_category 
        FOREIGN KEY (category_id) REFERENCES category(id)
);

CREATE TABLE IF NOT EXISTS team (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    school TEXT NOT NULL,
    grade TEXT NOT NULL,
    teacher TEXT NOT NULL,
    category_id INTEGER NOT NULL,
    CONSTRAINT fk_category
        FOREIGN KEY (category_id) REFERENCES category(id)
);

CREATE TABLE IF NOT EXISTS member (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT,
    is_leader BOOLEAN NOT NULL DEFAULT false,
    team_id INTEGER NOT NULL,
    CONSTRAINT fk_team
        FOREIGN KEY (team_id) REFERENCES team(id)
);

CREATE TABLE IF NOT EXISTS "match" (
    id SERIAL PRIMARY KEY,
    team_a_id INTEGER NOT NULL,
    team_b_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    CONSTRAINT fk_match_team_a
        FOREIGN KEY (team_a_id) REFERENCES team(id),
    CONSTRAINT fk_match_team_b
        FOREIGN KEY (team_b_id) REFERENCES team(id),
    CONSTRAINT fk_match_category
        FOREIGN KEY (category_id) REFERENCES category(id),
    CONSTRAINT chk_match_distinct_teams
        CHECK (team_a_id <> team_b_id)
);

CREATE TABLE IF NOT EXISTS match_queue (
    match_id INTEGER NOT NULL,
    team_id INTEGER NOT NULL,
    position INTEGER NOT NULL,
    PRIMARY KEY (match_id, team_id),
    CONSTRAINT fk_match_queue_match
        FOREIGN KEY (match_id) REFERENCES "match"(id) ON DELETE CASCADE,
    CONSTRAINT fk_match_queue_team
        FOREIGN KEY (team_id) REFERENCES team(id),
    CONSTRAINT uq_match_queue_position
        UNIQUE (match_id, position)
);

CREATE TABLE IF NOT EXISTS result (
    id SERIAL PRIMARY KEY,
    winner_team_id INTEGER NOT NULL,
    result_time TIMESTAMPTZ,
    match_id INTEGER NOT NULL UNIQUE,
    CONSTRAINT fk_result_winner
        FOREIGN KEY (winner_team_id) REFERENCES team(id),
    CONSTRAINT fk_result_match
        FOREIGN KEY (match_id) REFERENCES "match"(id) ON DELETE CASCADE
);
