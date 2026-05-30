CREATE TABLE IF NOT EXISTS category (
    id  SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS rule (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'characteristic',
    category_id INTEGER NOT NULL,
    CONSTRAINT chk_rule_type
        CHECK (type IN ('characteristic', 'restriction')),
    CONSTRAINT fk_category 
        FOREIGN KEY (category_id) REFERENCES category(id)
);

ALTER TABLE rule
    ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'characteristic';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_rule_type'
    ) THEN
        ALTER TABLE rule
            ADD CONSTRAINT chk_rule_type
            CHECK (type IN ('characteristic', 'restriction'));
    END IF;
END $$;

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
    team_a_id INTEGER,
    team_b_id INTEGER,
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

ALTER TABLE "match"
    ALTER COLUMN team_a_id DROP NOT NULL,
    ALTER COLUMN team_b_id DROP NOT NULL;

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

CREATE TABLE IF NOT EXISTS robot (
    id SERIAL PRIMARY KEY,
    team_id INTEGER NOT NULL,
    is_valid BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT fk_robot_team
        FOREIGN KEY (team_id) REFERENCES team(id)
);

CREATE TABLE IF NOT EXISTS robot_valid_rule (
    robot_id INTEGER NOT NULL,
    rule_id INTEGER NOT NULL,
    PRIMARY KEY (robot_id, rule_id),
    CONSTRAINT fk_robot_valid_rule_robot
        FOREIGN KEY (robot_id) REFERENCES robot(id) ON DELETE CASCADE,
    CONSTRAINT fk_robot_valid_rule_rule
        FOREIGN KEY (rule_id) REFERENCES rule(id)
);
