package store

import (
	"context"
	"errors"
	"robot/internal/domain"
	"robot/pkg/apperr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type CategoryStore struct {
	store *Store
}

var _ domain.CategoryRepository = (*CategoryStore)(nil)

func NewCategoryStore(s *Store) *CategoryStore {
	return &CategoryStore{
		store: s,
	}
}

func (st *CategoryStore) Insert(ctx context.Context, category *domain.Category) (domain.CategoryID, error) {
	op := "Insert"

	query := `
		INSERT INTO category
		(name)
		VALUES ($1)
		RETURNING id
		`

	var id domain.CategoryID
	err := st.store.db.QueryRow(ctx, query, category.Name).Scan(&id)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) && pgxErr.Code == UniqueViolation {
			return 0, apperr.Wrap(op, "category already exists with name", domain.ErrAlreadyExists, apperr.Field{Name: "name", Value: category.Name})
		}
		return 0, apperr.Wrap(op, "unexpected error inserting category", err, apperr.Field{Name: "database", Value: category.Name})
	}
	return id, nil
}

func (st *CategoryStore) FindByID(ctx context.Context, id domain.CategoryID) (*domain.Category, error) {
	op := "FindByID"

	query := `
		SELECT id, name
		FROM category
		WHERE id = $1
	`
	var foundID domain.CategoryID
	var name string

	err := st.store.db.QueryRow(ctx, query, id).Scan(&foundID, &name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "id", Value: id})
		}
		return nil, apperr.Wrap(op, "unexpected error selecting id", err, apperr.Field{Name: "id", Value: id})
	}

	category, err := domain.NewCategory(name)
	if err != nil {
		return nil, err
	}
	category.ID = foundID
	return category, nil
}

func (st *CategoryStore) FindByName(ctx context.Context, name string) (*domain.Category, error) {
	op := "FindByName"

	query := `
		SELECT id, name
		FROM category
		WHERE name = $1
	`
	var id domain.CategoryID
	var foundName string

	err := st.store.db.QueryRow(ctx, query, name).Scan(&id, &foundName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "name", Value: name})
		}
		return nil, apperr.Wrap(op, "unexpected error selecting name", err, apperr.Field{Name: "name", Value: name})
	}

	category, err := domain.NewCategory(foundName)
	if err != nil {
		return nil, err
	}
	category.ID = id
	return category, nil
}

func (st *CategoryStore) FindAll(ctx context.Context) ([]*domain.Category, error) {
	op := "FindAll"

	query := `
		SELECT id, name
		FROM category
	`
	rows, err := st.store.db.Query(ctx, query)
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error listing categories", err, apperr.Field{Name: "database", Value: ""})
	}

	categories, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*domain.Category, error) {
		var id domain.CategoryID
		var name string

		err := row.Scan(&id, &name)
		if err != nil {
			return nil, apperr.Wrap(op, "something went wrong when scanning", err, apperr.Field{Name: "scan", Value: row})
		}
		category, err := domain.NewCategory(name)
		if err != nil {
			return nil, err
		}
		category.ID = id
		return category, nil
	})
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error collecting categories", err, apperr.Field{Name: "collect", Value: rows})
	}

	return categories, nil
}
