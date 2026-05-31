package store

import (
	"context"
	"errors"
	"fmt"
	"robot/internal/domain"
	"robot/pkg/apperr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type ResultStore struct {
	store *Store
}

var _ domain.ResultRepository = (*ResultStore)(nil)

func NewResultStore(s *Store) *ResultStore {
	return &ResultStore{
		store: s,
	}
}

func (st *ResultStore) Insert(ctx context.Context, result *domain.Result) (domain.ResultID, error) {
	op := "Insert"

	query := `
		INSERT INTO result
		(winner_team_id, eliminated_team_id, result_time_seconds, match_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var resultTimeSeconds any
	if result.Time != nil {
		resultTimeSeconds = result.Time.TotalSeconds()
	}
	var eliminatedTeamID any
	if result.EliminatedTeamID != nil {
		eliminatedTeamID = *result.EliminatedTeamID
	}

	var id domain.ResultID
	err := st.store.db.QueryRow(ctx, query,
		result.Winner,
		eliminatedTeamID,
		resultTimeSeconds,
		result.MatchID,
	).Scan(&id)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			switch pgxErr.Code {
			case UniqueViolation:
				return 0, apperr.Wrap(op, "match already has a result", domain.ErrAlreadyExists, apperr.Field{Name: "match_id", Value: result.MatchID})
			case ForeignKeyViolation:
				return 0, apperr.Wrap(op, "result references a missing match, winner team, or eliminated team", domain.ErrInvalidReference, apperr.Field{Name: "result", Value: result})
			}
		}
		return 0, apperr.Wrap(op, "unexpected error inserting result", err, apperr.Field{Name: "database", Value: result})
	}

	return id, nil
}

func (st *ResultStore) FindByID(ctx context.Context, id domain.ResultID) (*domain.Result, error) {
	op := "FindByID"

	query := `
		SELECT id, winner_team_id, eliminated_team_id, result_time_seconds, match_id
		FROM result
		WHERE id = $1
	`

	result, err := st.scanResultRow(ctx, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "id", Value: id})
		}
		return nil, apperr.Wrap(op, "unexpected error selecting result id", err, apperr.Field{Name: "id", Value: id})
	}
	return result, nil
}

func (st *ResultStore) FindByMatchID(ctx context.Context, id domain.MatchID) (*domain.Result, error) {
	op := "FindByMatchID"

	query := `
		SELECT id, winner_team_id, eliminated_team_id, result_time_seconds, match_id
		FROM result
		WHERE match_id = $1
	`

	result, err := st.scanResultRow(ctx, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "match_id", Value: id})
		}
		return nil, apperr.Wrap(op, "unexpected error selecting result match_id", err, apperr.Field{Name: "match_id", Value: id})
	}
	return result, nil
}

func (st *ResultStore) Find(ctx context.Context, q domain.ResultQuery) ([]*domain.Result, error) {
	op := "Find"
	args, query := st.buildQuery(q)

	rows, err := st.store.db.Query(ctx, query, args...)
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error selecting results", err, apperr.Field{Name: "database", Value: q})
	}

	results, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*domain.Result, error) {
		return scanResult(row)
	})
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error collecting results", err, apperr.Field{Name: "collect", Value: q})
	}
	return results, nil
}

func (st *ResultStore) FindAll(ctx context.Context) ([]*domain.Result, error) {
	return st.Find(ctx, domain.ResultQuery{})
}

func (st *ResultStore) scanResultRow(ctx context.Context, query string, args ...any) (*domain.Result, error) {
	return scanResult(st.store.db.QueryRow(ctx, query, args...))
}

func (st *ResultStore) buildQuery(q domain.ResultQuery) ([]any, string) {
	query := `
		SELECT id, winner_team_id, eliminated_team_id, result_time_seconds, match_id
		FROM result
		WHERE 1=1
	`
	var args []any

	if q.Winner != 0 {
		args = append(args, q.Winner)
		query += fmt.Sprintf(" AND winner_team_id = $%d", len(args))
	}
	if q.MatchID != 0 {
		args = append(args, q.MatchID)
		query += fmt.Sprintf(" AND match_id = $%d", len(args))
	}

	return args, query
}

type resultRow interface {
	Scan(dest ...any) error
}

func scanResult(row resultRow) (*domain.Result, error) {
	var id domain.ResultID
	var winner domain.TeamID
	var eliminatedTeamID pgtype.Int4
	var resultTimeSeconds pgtype.Int4
	var matchID domain.MatchID

	if err := row.Scan(&id, &winner, &eliminatedTeamID, &resultTimeSeconds, &matchID); err != nil {
		return nil, err
	}

	var eliminated *domain.TeamID
	if eliminatedTeamID.Valid {
		id := domain.TeamID(eliminatedTeamID.Int32)
		eliminated = &id
	}

	var resultTime *domain.ResultTime
	if resultTimeSeconds.Valid {
		parsedTime, err := domain.NewResultTimeFromSeconds(int(resultTimeSeconds.Int32))
		if err != nil {
			return nil, err
		}
		resultTime = parsedTime
	}

	result, err := domain.NewResult(winner, matchID, eliminated, resultTime)
	if err != nil {
		return nil, err
	}
	result.ID = id
	return result, nil
}
