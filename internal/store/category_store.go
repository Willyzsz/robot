package store

import (
	"context"
	"errors"
	"robot/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type CategoryStore struct {
	store *Store
}

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
			return 0, domain.NewRobotErr(op, "", "name", category.Name, domain.ErrAlreadyExists, "category already exists with name")
		}
		return 0, domain.NewRobotErr(op, "", "database", category.Name, err, "unexpected error inserting category")
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
			return nil, domain.NewRobotErr(op, "", "id", id, domain.ErrNotFound, "")
		}
		return nil, domain.NewRobotErr(op, "", "id", id, err, "unexpected error selecting id")
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
			return nil, domain.NewRobotErr(op, "", "name", name, domain.ErrNotFound, "")
		}
		return nil, domain.NewRobotErr(op, "", "name", name, err, "unexpected error selecting name")
	}

	category, err := domain.NewCategory(foundName)
	if err != nil {
		return nil, err
	}
	category.ID = id
	return category, nil
}
