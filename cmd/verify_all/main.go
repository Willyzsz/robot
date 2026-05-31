package main

import (
	"context"
	"fmt"
	"log"
	"robot/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	connString, err := store.ConnStringFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback(ctx)

	robotsCreated, err := tx.Exec(ctx, `
		INSERT INTO robot (team_id, is_valid)
		SELECT t.id, true
		FROM team t
		WHERE NOT EXISTS (
			SELECT 1
			FROM robot r
			WHERE r.team_id = t.id
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	rulesAdded, err := tx.Exec(ctx, `
		INSERT INTO robot_valid_rule (robot_id, rule_id)
		SELECT r.id, rule.id
		FROM robot r
		JOIN team t ON t.id = r.team_id
		JOIN rule ON rule.category_id = t.category_id
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		log.Fatal(err)
	}

	robotsVerified, err := tx.Exec(ctx, `
		UPDATE robot
		SET is_valid = true
	`)
	if err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("created_robots=%d added_valid_rules=%d verified_robots=%d\n",
		robotsCreated.RowsAffected(),
		rulesAdded.RowsAffected(),
		robotsVerified.RowsAffected(),
	)
}
