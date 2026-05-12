package domain

import "context"

type CategoryRepo interface {
	Insert(ctx context.Context, category Category) (CategoryID, error)
}