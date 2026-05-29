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

type TeamStore struct {
	store *Store
}

var _ domain.TeamRepository = (*TeamStore)(nil)

func NewTeamStore(s *Store) *TeamStore {
	return &TeamStore{
		store: s,
	}
}

func (st *TeamStore) Insert(ctx context.Context, team *domain.Team) (domain.TeamID, error) {
	op := "Insert"

	query := `
		INSERT INTO team
		(name, school, grade, teacher, category_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id domain.TeamID
	err := st.store.db.QueryRow(ctx, query,
		team.Name,
		team.School,
		team.Grade,
		team.Teacher,
		team.CategoryID,
	).Scan(&id)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			switch pgxErr.Code {
			case UniqueViolation:
				return 0, apperr.Wrap(op, "team already exists with name", domain.ErrAlreadyExists, apperr.Field{Name: "name", Value: team.Name})
			case ForeignKeyViolation:
				return 0, apperr.Wrap(op, "category_id does not reference an existing category", domain.ErrInvalidReference, apperr.Field{Name: "category_id", Value: team.CategoryID})
			}
		}
		return 0, apperr.Wrap(op, "unexpected error inserting team", err, apperr.Field{Name: "database", Value: team.Name})
	}

	return id, nil
}

func (st *TeamStore) FindByID(ctx context.Context, id domain.TeamID) (*domain.Team, error) {
	op := "FindByID"

	query := `
		SELECT id, name, school, grade, teacher, category_id
		FROM team
		WHERE id = $1
	`
	var foundID domain.TeamID
	var name, school, grade, teacher string
	var categoryID domain.CategoryID

	err := st.store.db.QueryRow(ctx, query, id).Scan(&foundID, &name, &school, &grade, &teacher, &categoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "id", Value: id})
		}
		return nil, apperr.Wrap(op, "unexpected error selecting id", err, apperr.Field{Name: "id", Value: id})
	}

	team, err := domain.NewTeam(name, school, grade, teacher, categoryID)
	if err != nil {
		return nil, err
	}
	team.ID = foundID
	return team, nil
}

func (st *TeamStore) FindByName(ctx context.Context, name string) (*domain.Team, error) {
	op := "FindByName"

	query := `
		SELECT id, name, school, grade, teacher, category_id
		FROM team
		WHERE name = $1
	`
	var id domain.TeamID
	var foundName, school, grade, teacher string
	var categoryID domain.CategoryID

	err := st.store.db.QueryRow(ctx, query, name).Scan(&id, &foundName, &school, &grade, &teacher, &categoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "name", Value: name})
		}
		return nil, apperr.Wrap(op, "unexpected error selecting name", err, apperr.Field{Name: "name", Value: name})
	}

	team, err := domain.NewTeam(name, school, grade, teacher, categoryID)
	if err != nil {
		return nil, err
	}

	team.ID = id
	return team, nil
}

func (st *TeamStore) Find(ctx context.Context, t domain.TeamQuery) ([]*domain.Team, error) {
	op := "Find"
	args, query := st.buildQuery(t)

	rows, err := st.store.db.Query(ctx, query, args...)
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error selecting team", err, apperr.Field{Name: "database", Value: t})
	}

	teams, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*domain.Team, error) {
		var foundID domain.TeamID
		var name, school, grade, teacher string
		var categoryID domain.CategoryID

		err := row.Scan(&foundID, &name, &school, &grade, &teacher, &categoryID)
		if err != nil {
			return nil, apperr.Wrap(op, "unexpected error scanning team", err, apperr.Field{Name: "scan", Value: row})
		}

		team, err := domain.NewTeam(name, school, grade, teacher, categoryID)
		if err != nil {
			return nil, err
		}

		team.ID = foundID
		return team, nil
	})

	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error collecting teams", err, apperr.Field{Name: "collect", Value: t})
	}
	return teams, nil
}

func (st *TeamStore) FindAll(ctx context.Context) ([]*domain.Team, error) {
	op := "FindAll"

	query := `
		SELECT id, name, school, grade, teacher, category_id
		FROM team
	`
	rows, err := st.store.db.Query(ctx, query)
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error selecting teams", err, apperr.Field{Name: "database", Value: nil})
	}

	teams, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*domain.Team, error) {
		var id domain.TeamID
		var name, school, grade, teacher string
		var categoryID domain.CategoryID

		err := row.Scan(&id, &name, &school, &grade, &teacher, &categoryID)
		if err != nil {
			return nil, apperr.Wrap(op, "unexpected error scanning team", err, apperr.Field{Name: "scan", Value: row})
		}

		team, err := domain.NewTeam(name, school, grade, teacher, categoryID)
		if err != nil {
			return nil, err
		}

		team.ID = id
		return team, nil
	})

	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error collecting teams", err, apperr.Field{Name: "collect", Value: nil})
	}
	return teams, nil
}

func (st *TeamStore) buildQuery(t domain.TeamQuery) ([]any, string) {
	query := `
		SELECT id, name, school, grade, teacher, category_id
		FROM team
		WHERE 1=1
	`
	var args []any

	if t.Name != "" {
		args = append(args, t.Name)
		query += fmt.Sprintf(" AND name = $%d", len(args))
	}

	if t.School != "" {
		args = append(args, t.School)
		query += fmt.Sprintf(" AND school = $%d", len(args))
	}

	if t.Grade != "" {
		args = append(args, t.Grade)
		query += fmt.Sprintf(" AND grade = $%d", len(args))
	}

	if t.Teacher != "" {
		args = append(args, t.Teacher)
		query += fmt.Sprintf(" AND teacher = $%d", len(args))
	}

	if t.CategoryID != 0 {
		args = append(args, t.CategoryID)
		query += fmt.Sprintf(" AND category_id = $%d", len(args))
	}

	return args, query
}
