package store

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	UniqueViolation string = "23505"
	ForeignKeyViolation string = "23503"
)

type Store struct {
	db *pgxpool.Pool
}

func ConnStringFromEnv() string {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	db := os.Getenv("POSTGRES_DB")


	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user,
		password,
		host,
		port,
		db,
	)
}

func Open(ctx context.Context) (*Store, error) {
	pool, err := pgxpool.New(ctx, ConnStringFromEnv())
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Store{db: pool}, nil
}

func (st *Store) Close() {
	st.db.Close()
}