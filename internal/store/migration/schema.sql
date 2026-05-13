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
