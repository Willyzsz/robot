package store

import (
	"context"
	"errors"
	"robot/internal/domain"
	"robot/pkg/apperr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type RuleStore struct {
	store *Store
}

var _ domain.RuleRepository = (*RuleStore)(nil)

func NewRuleStore(s *Store) *RuleStore {
	return &RuleStore{
		store: s,
	}
}

func (st *RuleStore) Insert(ctx context.Context, rule *domain.Rule) (domain.RuleID, error) {
	op := "Insert"

	query := `
		INSERT INTO rule
		(description, category_id)
		VALUES ($1, $2)
		RETURNING id
	`

	var id domain.RuleID
	err := st.store.db.QueryRow(ctx, query, rule.Description, rule.CategoryID).Scan(&id)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) && pgxErr.Code == ForeignKeyViolation {
			return 0, apperr.Wrap(op, "category_id does not reference an existing category", domain.ErrInvalidReference, apperr.Field{Name: "category_id", Value: rule.CategoryID})
		}
		return 0, apperr.Wrap(op, "unexpected error inserting rule", err, apperr.Field{Name: "database", Value: rule})
	}

	return id, nil
}

func (st *RuleStore) FindByID(ctx context.Context, id domain.RuleID) (*domain.Rule, error) {
	op := "FindByID"

	query := `
		SELECT id, description, category_id
		FROM rule
		WHERE id = $1
	`

	rule, err := scanRule(st.store.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "id", Value: id})
		}
		return nil, apperr.Wrap(op, "unexpected error selecting rule id", err, apperr.Field{Name: "id", Value: id})
	}
	return rule, nil
}

func (st *RuleStore) FindByCategoryID(ctx context.Context, id domain.CategoryID) ([]*domain.Rule, error) {
	return st.find(ctx, `
		SELECT id, description, category_id
		FROM rule
		WHERE category_id = $1
	`, id)
}

func (st *RuleStore) FindAll(ctx context.Context) ([]*domain.Rule, error) {
	return st.find(ctx, `
		SELECT id, description, category_id
		FROM rule
	`)
}

func (st *RuleStore) find(ctx context.Context, query string, args ...any) ([]*domain.Rule, error) {
	op := "find"

	rows, err := st.store.db.Query(ctx, query, args...)
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error selecting rules", err)
	}

	rules, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*domain.Rule, error) {
		return scanRule(row)
	})
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error collecting rules", err)
	}
	return rules, nil
}

type ruleRow interface {
	Scan(dest ...any) error
}

func scanRule(row ruleRow) (*domain.Rule, error) {
	var id domain.RuleID
	var description string
	var categoryID domain.CategoryID

	if err := row.Scan(&id, &description, &categoryID); err != nil {
		return nil, err
	}

	rule, err := domain.NewRule(description, categoryID)
	if err != nil {
		return nil, err
	}
	rule.ID = id
	return rule, nil
}
