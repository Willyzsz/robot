package store

import (
	"context"
	"errors"
	"robot/internal/domain"
	"robot/pkg/apperr"

	"github.com/jackc/pgx/v5"
)

type UserStore struct {
	store *Store
}

var _ domain.UserRepository = (*UserStore)(nil)

func NewUserStore(s *Store) *UserStore {
	return &UserStore{store: s}
}

func (st *UserStore) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	op := "FindByUsername"

	user, err := scanUser(st.store.db.QueryRow(ctx, `
		SELECT id, username, name, role, password_hash
		FROM user_account
		WHERE username = $1
	`, username))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "username", Value: username})
		}
		return nil, apperr.Wrap(op, "unexpected error selecting user", err, apperr.Field{Name: "username", Value: username})
	}
	if err := st.hydrateUserCategories(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (st *UserStore) hydrateUserCategories(ctx context.Context, user *domain.User) error {
	rows, err := st.store.db.Query(ctx, `
		SELECT category_id
		FROM user_category
		WHERE user_id = $1
		ORDER BY category_id
	`, user.ID)
	if err != nil {
		return apperr.Wrap("hydrateUserCategories", "unexpected error selecting user categories", err, apperr.Field{Name: "user_id", Value: user.ID})
	}
	defer rows.Close()

	user.CategoryIDs = []domain.CategoryID{}
	for rows.Next() {
		var categoryID domain.CategoryID
		if err := rows.Scan(&categoryID); err != nil {
			return apperr.Wrap("hydrateUserCategories", "unexpected error scanning user category", err, apperr.Field{Name: "user_id", Value: user.ID})
		}
		user.CategoryIDs = append(user.CategoryIDs, categoryID)
	}
	if err := rows.Err(); err != nil {
		return apperr.Wrap("hydrateUserCategories", "unexpected error collecting user categories", err, apperr.Field{Name: "user_id", Value: user.ID})
	}
	return nil
}

type userRow interface {
	Scan(dest ...any) error
}

func scanUser(row userRow) (*domain.User, error) {
	var id domain.UserID
	var username, name, passwordHash string
	var role domain.UserRole

	if err := row.Scan(&id, &username, &name, &role, &passwordHash); err != nil {
		return nil, err
	}

	user, err := domain.NewUser(username, name, role, passwordHash)
	if err != nil {
		return nil, err
	}
	user.ID = id
	return user, nil
}
