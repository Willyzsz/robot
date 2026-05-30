WITH category_seed(name) AS (
    VALUES
        ('Sumo'),
        ('Velocista')
)
INSERT INTO category (name)
SELECT name
FROM category_seed
ON CONFLICT (name) DO NOTHING;

WITH rule_seed(category_name, description, type) AS (
    VALUES
        ('Sumo', 'Robot must fit inside 10cm x 10cm', 'characteristic'),
        ('Sumo', 'Robot must weigh 500g or less', 'characteristic'),
        ('Sumo', 'Robot must not use a jammer', 'restriction'),
        ('Sumo', 'Robot must not damage the opponent', 'restriction'),
        ('Velocista', 'Robot must start behind the line', 'characteristic'),
        ('Velocista', 'Robot must follow the track autonomously', 'characteristic'),
        ('Velocista', 'Robot must not leave parts on the track', 'restriction')
)
INSERT INTO rule (description, type, category_id)
SELECT rs.description, rs.type, c.id
FROM rule_seed rs
JOIN category c ON c.name = rs.category_name
WHERE NOT EXISTS (
    SELECT 1
    FROM rule r
    WHERE r.description = rs.description
      AND r.category_id = c.id
);

WITH team_seed(name, school, grade, teacher, category_name) AS (
    VALUES
        ('Sumo Alpha', 'Escuela Norte', '10', 'Prof. Ana', 'Sumo'),
        ('Sumo Beta', 'Escuela Centro', '11', 'Prof. Luis', 'Sumo'),
        ('Sumo No Members', 'Escuela Sur', '9', 'Prof. Carla', 'Sumo'),
        ('Velocista Flash', 'Instituto Rapido', '12', 'Prof. Mario', 'Velocista'),
        ('Velocista Pending', 'Instituto Tecnico', '8', 'Prof. Elena', 'Velocista')
)
INSERT INTO team (name, school, grade, teacher, category_id)
SELECT ts.name, ts.school, ts.grade, ts.teacher, c.id
FROM team_seed ts
JOIN category c ON c.name = ts.category_name
ON CONFLICT (name) DO NOTHING;

WITH member_seed(team_name, name, email, is_leader) AS (
    VALUES
        ('Sumo Alpha', 'Alex Alpha', 'alex.alpha@example.com', true),
        ('Sumo Alpha', 'Ari Alpha', 'ari.alpha@example.com', false),
        ('Sumo Beta', 'Bruno Beta', 'bruno.beta@example.com', true),
        ('Sumo Beta', 'Bianca Beta', 'bianca.beta@example.com', false),
        ('Velocista Flash', 'Valeria Flash', 'valeria.flash@example.com', true),
        ('Velocista Pending', 'Victor Pending', 'victor.pending@example.com', true)
)
INSERT INTO member (name, email, is_leader, team_id)
SELECT ms.name, ms.email, ms.is_leader, t.id
FROM member_seed ms
JOIN team t ON t.name = ms.team_name
WHERE NOT EXISTS (
    SELECT 1
    FROM member m
    WHERE m.team_id = t.id
      AND m.email = ms.email
);

WITH robot_seed(team_name, is_valid) AS (
    VALUES
        ('Sumo Alpha', true),
        ('Sumo Beta', false),
        ('Velocista Flash', true),
        ('Velocista Pending', false)
)
INSERT INTO robot (team_id, is_valid)
SELECT t.id, rs.is_valid
FROM robot_seed rs
JOIN team t ON t.name = rs.team_name
WHERE NOT EXISTS (
    SELECT 1
    FROM robot r
    WHERE r.team_id = t.id
);

WITH robot_rule_seed(team_name, rule_description) AS (
    VALUES
        ('Sumo Alpha', 'Robot must fit inside 10cm x 10cm'),
        ('Sumo Alpha', 'Robot must weigh 500g or less'),
        ('Sumo Alpha', 'Robot must not use a jammer'),
        ('Sumo Alpha', 'Robot must not damage the opponent'),
        ('Sumo Beta', 'Robot must fit inside 10cm x 10cm'),
        ('Sumo Beta', 'Robot must not use a jammer'),
        ('Velocista Flash', 'Robot must start behind the line'),
        ('Velocista Flash', 'Robot must follow the track autonomously'),
        ('Velocista Flash', 'Robot must not leave parts on the track'),
        ('Velocista Pending', 'Robot must start behind the line')
)
INSERT INTO robot_valid_rule (robot_id, rule_id)
SELECT rb.id, r.id
FROM robot_rule_seed rrs
JOIN team t ON t.name = rrs.team_name
JOIN robot rb ON rb.team_id = t.id
JOIN rule r ON r.description = rrs.rule_description
JOIN category c ON c.id = t.category_id AND c.id = r.category_id
ON CONFLICT (robot_id, rule_id) DO NOTHING;

UPDATE robot rb
SET is_valid = rule_counts.passed_rules = rule_counts.total_rules
FROM (
    SELECT
        rb.id AS robot_id,
        COUNT(DISTINCT rvr.rule_id) AS passed_rules,
        COUNT(DISTINCT r.id) AS total_rules
    FROM robot rb
    JOIN team t ON t.id = rb.team_id
    JOIN rule r ON r.category_id = t.category_id
    LEFT JOIN robot_valid_rule rvr ON rvr.robot_id = rb.id AND rvr.rule_id = r.id
    GROUP BY rb.id
) rule_counts
WHERE rb.id = rule_counts.robot_id;
