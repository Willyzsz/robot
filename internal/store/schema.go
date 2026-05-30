package store

import (
	"context"
	_ "embed"
)

//go:embed migration/schema.sql
var schema string

//go:embed migration/seed.sql
var seed string

func (st *Store) Migrate(ctx context.Context) error {
	_, err := st.db.Exec(ctx, schema)
	return err
}

func (st *Store) Seed(ctx context.Context) error {
	_, err := st.db.Exec(ctx, seed)
	return err
}
