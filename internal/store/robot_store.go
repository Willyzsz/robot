package store

import (
	"context"
	"errors"
	"fmt"
	"robot/internal/domain"
	"robot/pkg/apperr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type RobotStore struct {
	store *Store
}

var _ domain.RobotRepository = (*RobotStore)(nil)

func NewRobotStore(s *Store) *RobotStore {
	return &RobotStore{
		store: s,
	}
}

func (st *RobotStore) Insert(ctx context.Context, robot *domain.Robot) (domain.RobotID, error) {
	op := "Insert"

	tx, err := st.store.db.Begin(ctx)
	if err != nil {
		return 0, apperr.Wrap(op, "unexpected error beginning robot transaction", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO robot
		(team_id, is_valid)
		VALUES ($1, $2)
		RETURNING id
	`

	var id domain.RobotID
	err = tx.QueryRow(ctx, query, robot.TeamID, robot.IsValid).Scan(&id)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) && pgxErr.Code == ForeignKeyViolation {
			return 0, apperr.Wrap(op, "team_id does not reference an existing team", domain.ErrInvalidReference, apperr.Field{Name: "team_id", Value: robot.TeamID})
		}
		return 0, apperr.Wrap(op, "unexpected error inserting robot", err, apperr.Field{Name: "database", Value: robot})
	}

	for _, ruleID := range robot.ValidRules {
		_, err = tx.Exec(ctx, `
			INSERT INTO robot_valid_rule
			(robot_id, rule_id)
			VALUES ($1, $2)
		`, id, ruleID)
		if err != nil {
			var pgxErr *pgconn.PgError
			if errors.As(err, &pgxErr) && pgxErr.Code == ForeignKeyViolation {
				return 0, apperr.Wrap(op, "rule_id does not reference an existing rule", domain.ErrInvalidReference, apperr.Field{Name: "rule_id", Value: ruleID})
			}
			return 0, apperr.Wrap(op, "unexpected error inserting robot valid rule", err, apperr.Field{Name: "rule_id", Value: ruleID})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, apperr.Wrap(op, "unexpected error committing robot transaction", err)
	}

	return id, nil
}

func (st *RobotStore) FindByID(ctx context.Context, id domain.RobotID) (*domain.Robot, error) {
	op := "FindByID"

	query := `
		SELECT id, team_id, is_valid
		FROM robot
		WHERE id = $1
	`
	robot, err := scanRobot(st.store.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "id", Value: id})
		}
		return nil, apperr.Wrap(op, "unexpected error selecting robot id", err, apperr.Field{Name: "id", Value: id})
	}

	if err := st.hydrateValidRules(ctx, robot); err != nil {
		return nil, err
	}
	return robot, nil
}

func (st *RobotStore) Find(ctx context.Context, q domain.RobotQuery) ([]*domain.Robot, error) {
	op := "Find"
	args, query := st.buildQuery(q)

	rows, err := st.store.db.Query(ctx, query, args...)
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error selecting robots", err, apperr.Field{Name: "database", Value: q})
	}
	defer rows.Close()

	var robots []*domain.Robot
	for rows.Next() {
		robot, err := scanRobot(rows)
		if err != nil {
			return nil, apperr.Wrap(op, "unexpected error scanning robot", err, apperr.Field{Name: "scan", Value: q})
		}
		if err := st.hydrateValidRules(ctx, robot); err != nil {
			return nil, err
		}
		robots = append(robots, robot)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Wrap(op, "unexpected error collecting robots", err, apperr.Field{Name: "collect", Value: q})
	}

	return robots, nil
}

func (st *RobotStore) FindAll(ctx context.Context) ([]*domain.Robot, error) {
	return st.Find(ctx, domain.RobotQuery{})
}

func (st *RobotStore) AddValidRule(ctx context.Context, robotID domain.RobotID, ruleID domain.RuleID) error {
	op := "AddValidRule"

	_, err := st.store.db.Exec(ctx, `
		INSERT INTO robot_valid_rule
		(robot_id, rule_id)
		VALUES ($1, $2)
	`, robotID, ruleID)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			switch pgxErr.Code {
			case UniqueViolation:
				return apperr.Wrap(op, "robot already has this valid rule", domain.ErrAlreadyExists, apperr.Field{Name: "rule_id", Value: ruleID})
			case ForeignKeyViolation:
				return apperr.Wrap(op, "robot_id or rule_id does not reference an existing row", domain.ErrInvalidReference, apperr.Field{Name: "robot_id", Value: robotID}, apperr.Field{Name: "rule_id", Value: ruleID})
			}
		}
		return apperr.Wrap(op, "unexpected error inserting robot valid rule", err, apperr.Field{Name: "robot_id", Value: robotID}, apperr.Field{Name: "rule_id", Value: ruleID})
	}

	return nil
}

func (st *RobotStore) SetValidity(ctx context.Context, robotID domain.RobotID, isValid bool) error {
	op := "SetValidity"

	commandTag, err := st.store.db.Exec(ctx, `
		UPDATE robot
		SET is_valid = $2
		WHERE id = $1
	`, robotID, isValid)
	if err != nil {
		return apperr.Wrap(op, "unexpected error updating robot validity", err, apperr.Field{Name: "robot_id", Value: robotID})
	}
	if commandTag.RowsAffected() == 0 {
		return apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "robot_id", Value: robotID})
	}
	return nil
}

func (st *RobotStore) hydrateValidRules(ctx context.Context, robot *domain.Robot) error {
	op := "hydrateValidRules"

	rows, err := st.store.db.Query(ctx, `
		SELECT rule_id
		FROM robot_valid_rule
		WHERE robot_id = $1
		ORDER BY rule_id
	`, robot.ID)
	if err != nil {
		return apperr.Wrap(op, "unexpected error selecting robot valid rules", err, apperr.Field{Name: "robot_id", Value: robot.ID})
	}

	rules, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.RuleID, error) {
		var ruleID domain.RuleID
		if err := row.Scan(&ruleID); err != nil {
			return 0, err
		}
		return ruleID, nil
	})
	if err != nil {
		return apperr.Wrap(op, "unexpected error collecting robot valid rules", err, apperr.Field{Name: "robot_id", Value: robot.ID})
	}

	robot.ValidRules = rules
	return nil
}

func (st *RobotStore) buildQuery(q domain.RobotQuery) ([]any, string) {
	query := `
		SELECT id, team_id, is_valid
		FROM robot
		WHERE 1=1
	`
	var args []any

	if q.TeamID != 0 {
		args = append(args, q.TeamID)
		query += fmt.Sprintf(" AND team_id = $%d", len(args))
	}
	if q.IsValid != nil {
		args = append(args, *q.IsValid)
		query += fmt.Sprintf(" AND is_valid = $%d", len(args))
	}

	return args, query
}

type robotRow interface {
	Scan(dest ...any) error
}

func scanRobot(row robotRow) (*domain.Robot, error) {
	var id domain.RobotID
	var teamID domain.TeamID
	var isValid bool

	if err := row.Scan(&id, &teamID, &isValid); err != nil {
		return nil, err
	}

	robot, err := domain.NewRobot(teamID, nil)
	if err != nil {
		return nil, err
	}
	robot.ID = id
	robot.IsValid = isValid
	return robot, nil
}
