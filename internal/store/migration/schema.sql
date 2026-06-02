CREATE TABLE IF NOT EXISTS category (
    id  SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_account (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'visitante',
    password_hash TEXT NOT NULL
);

ALTER TABLE user_account
    ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'visitante';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_user_role'
    ) THEN
        ALTER TABLE user_account
            ADD CONSTRAINT chk_user_role
            CHECK (role IN ('juez', 'visitante', 'arbitro', 'admin', 'dev'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS user_category (
    user_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, category_id),
    CONSTRAINT fk_user_category_user
        FOREIGN KEY (user_id) REFERENCES user_account(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_category_category
        FOREIGN KEY (category_id) REFERENCES category(id) ON DELETE CASCADE
);

WITH user_seed(username, name, role, password_hash) AS (
    VALUES
        ('admin', 'Administrador', 'admin', '$2a$10$ba/zJUnNZvs/M4qMlKnyu.VDooIhiy4ElnkTPTo6HXW0SmjeBsN/y'),
        ('dev', 'Desarrollador', 'dev', '$2a$10$ba/zJUnNZvs/M4qMlKnyu.VDooIhiy4ElnkTPTo6HXW0SmjeBsN/y'),
        ('juez', 'Juez', 'juez', '$2a$10$ba/zJUnNZvs/M4qMlKnyu.VDooIhiy4ElnkTPTo6HXW0SmjeBsN/y'),
        ('arbitro', 'Arbitro', 'arbitro', '$2a$10$ba/zJUnNZvs/M4qMlKnyu.VDooIhiy4ElnkTPTo6HXW0SmjeBsN/y'),
        ('visitante', 'Visitante', 'visitante', '$2a$10$ba/zJUnNZvs/M4qMlKnyu.VDooIhiy4ElnkTPTo6HXW0SmjeBsN/y')
)
INSERT INTO user_account (username, name, role, password_hash)
SELECT username, name, role, password_hash
FROM user_seed
ON CONFLICT (username) DO UPDATE
SET name = EXCLUDED.name,
    role = EXCLUDED.role,
    password_hash = EXCLUDED.password_hash;

INSERT INTO user_category (user_id, category_id)
SELECT u.id, c.id
FROM user_account u
CROSS JOIN category c
WHERE u.role IN ('juez', 'arbitro')
ON CONFLICT (user_id, category_id) DO NOTHING;

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
    is_internal BOOLEAN NOT NULL DEFAULT false,
    category_id INTEGER NOT NULL,
    CONSTRAINT fk_category
        FOREIGN KEY (category_id) REFERENCES category(id)
);

ALTER TABLE team
    ADD COLUMN IF NOT EXISTS is_internal BOOLEAN NOT NULL DEFAULT false;

UPDATE team
SET is_internal = true
WHERE lower(grade) IN ('utnc nivel tsu', 'utnc nivel ingeniería', 'utnc nivel ingenieria');

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
    is_internal BOOLEAN NOT NULL DEFAULT false,
    bracket_id TEXT,
    bracket_key TEXT,
    bracket_round INTEGER,
    bracket_slot INTEGER,
    status TEXT NOT NULL DEFAULT 'ready',
    CONSTRAINT fk_match_team_a
        FOREIGN KEY (team_a_id) REFERENCES team(id),
    CONSTRAINT fk_match_team_b
        FOREIGN KEY (team_b_id) REFERENCES team(id),
    CONSTRAINT fk_match_category
        FOREIGN KEY (category_id) REFERENCES category(id),
    CONSTRAINT chk_match_distinct_teams
        CHECK (team_a_id <> team_b_id),
    CONSTRAINT chk_match_status
        CHECK (status IN ('pending', 'ready', 'completed', 'bye'))
);

ALTER TABLE "match"
    ALTER COLUMN team_a_id DROP NOT NULL,
    ALTER COLUMN team_b_id DROP NOT NULL;

ALTER TABLE "match"
    ADD COLUMN IF NOT EXISTS bracket_id TEXT,
    ADD COLUMN IF NOT EXISTS bracket_key TEXT,
    ADD COLUMN IF NOT EXISTS bracket_round INTEGER,
    ADD COLUMN IF NOT EXISTS bracket_slot INTEGER,
    ADD COLUMN IF NOT EXISTS is_internal BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'ready';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_match_status'
    ) THEN
        ALTER TABLE "match"
            ADD CONSTRAINT chk_match_status
            CHECK (status IN ('pending', 'ready', 'completed', 'bye'));
    END IF;
END $$;

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
    eliminated_team_id INTEGER,
    result_time_seconds INTEGER,
    match_id INTEGER NOT NULL UNIQUE,
    CONSTRAINT ck_result_time_seconds
        CHECK (result_time_seconds IS NULL OR result_time_seconds >= 0),
    CONSTRAINT ck_result_eliminated_not_winner
        CHECK (eliminated_team_id IS NULL OR eliminated_team_id <> winner_team_id),
    CONSTRAINT fk_result_winner
        FOREIGN KEY (winner_team_id) REFERENCES team(id),
    CONSTRAINT fk_result_eliminated
        FOREIGN KEY (eliminated_team_id) REFERENCES team(id),
    CONSTRAINT fk_result_match
        FOREIGN KEY (match_id) REFERENCES "match"(id) ON DELETE CASCADE
);

ALTER TABLE result
    ADD COLUMN IF NOT EXISTS eliminated_team_id INTEGER,
    ADD COLUMN IF NOT EXISTS result_time_seconds INTEGER;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_result_time_seconds'
    ) THEN
        ALTER TABLE result
            ADD CONSTRAINT ck_result_time_seconds
            CHECK (result_time_seconds IS NULL OR result_time_seconds >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_result_eliminated_not_winner'
    ) THEN
        ALTER TABLE result
            ADD CONSTRAINT ck_result_eliminated_not_winner
            CHECK (eliminated_team_id IS NULL OR eliminated_team_id <> winner_team_id);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_result_eliminated'
    ) THEN
        ALTER TABLE result
            ADD CONSTRAINT fk_result_eliminated
            FOREIGN KEY (eliminated_team_id) REFERENCES team(id);
    END IF;
END $$;

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
