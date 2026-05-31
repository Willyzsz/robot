package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"robot/internal/domain"
	"robot/pkg/apperr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type MatchStore struct {
	store *Store
}

var _ domain.MatchRepository = (*MatchStore)(nil)

func NewMatchStore(s *Store) *MatchStore {
	return &MatchStore{
		store: s,
	}
}

func (st *MatchStore) Insert(ctx context.Context, match *domain.Match) (domain.MatchID, error) {
	op := "Insert"

	tx, err := st.store.db.Begin(ctx)
	if err != nil {
		return 0, apperr.Wrap(op, "unexpected error beginning match transaction", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO "match"
		(team_a_id, team_b_id, category_id, bracket_id, bracket_key, bracket_round, bracket_slot, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id domain.MatchID
	err = tx.QueryRow(ctx, query,
		matchTeamID(match.TeamA),
		matchTeamID(match.TeamB),
		match.CategoryID,
		nullableString(match.BracketID),
		nullableString(match.BracketKey),
		nullableInt(match.BracketRound, match.BracketKey),
		nullableInt(match.BracketSlot, match.BracketKey),
		matchStatus(match.Status),
	).Scan(&id)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			switch pgxErr.Code {
			case ForeignKeyViolation:
				return 0, apperr.Wrap(op, "match references a missing team or category", domain.ErrInvalidReference, apperr.Field{Name: "match", Value: match})
			case CheckViolation:
				return 0, apperr.Wrap(op, "teams cannot be the same", domain.ErrInvalid, apperr.Field{Name: "match", Value: match})
			}
		}
		return 0, apperr.Wrap(op, "unexpected error inserting match", err, apperr.Field{Name: "database", Value: match})
	}

	for position, teamID := range match.Queue {
		_, err = tx.Exec(ctx,
			`INSERT INTO match_queue (match_id, team_id, position) VALUES ($1, $2, $3)`,
			id,
			teamID,
			position,
		)
		if err != nil {
			var pgxErr *pgconn.PgError
			if errors.As(err, &pgxErr) && pgxErr.Code == ForeignKeyViolation {
				return 0, apperr.Wrap(op, "queue references a missing team", domain.ErrInvalidReference, apperr.Field{Name: "team_id", Value: teamID})
			}
			return 0, apperr.Wrap(op, "unexpected error inserting match queue", err, apperr.Field{Name: "team_id", Value: teamID})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, apperr.Wrap(op, "unexpected error committing match transaction", err)
	}

	return id, nil
}

func (st *MatchStore) FindByID(ctx context.Context, id domain.MatchID) (*domain.Match, error) {
	op := "FindByID"

	query := matchSelectQuery() + ` WHERE m.id = $1`
	match, err := st.scanMatchRow(ctx, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "id", Value: id})
		}
		return nil, apperr.Wrap(op, "unexpected error selecting match id", err, apperr.Field{Name: "id", Value: id})
	}
	return match, nil
}

func (st *MatchStore) Find(ctx context.Context, q domain.MatchQuery) ([]*domain.Match, error) {
	op := "Find"
	args, query := st.buildQuery(q)

	rows, err := st.store.db.Query(ctx, query, args...)
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error selecting matches", err, apperr.Field{Name: "database", Value: q})
	}
	defer rows.Close()

	var matches []*domain.Match
	for rows.Next() {
		match, err := scanMatch(rows)
		if err != nil {
			return nil, apperr.Wrap(op, "unexpected error scanning match", err, apperr.Field{Name: "scan", Value: q})
		}
		if err := st.hydrateMatch(ctx, match); err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Wrap(op, "unexpected error collecting matches", err, apperr.Field{Name: "collect", Value: q})
	}

	return matches, nil
}

func (st *MatchStore) FindAll(ctx context.Context) ([]*domain.Match, error) {
	return st.Find(ctx, domain.MatchQuery{})
}

func (st *MatchStore) SetStatus(ctx context.Context, id domain.MatchID, status domain.MatchStatus) error {
	op := "SetStatus"
	commandTag, err := st.store.db.Exec(ctx, `
		UPDATE "match"
		SET status = $2
		WHERE id = $1
	`, id, status)
	if err != nil {
		return apperr.Wrap(op, "unexpected error updating match status", err, apperr.Field{Name: "id", Value: id})
	}
	if commandTag.RowsAffected() == 0 {
		return apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "id", Value: id})
	}
	return nil
}

func (st *MatchStore) scanMatchRow(ctx context.Context, query string, args ...any) (*domain.Match, error) {
	match, err := scanMatch(st.store.db.QueryRow(ctx, query, args...))
	if err != nil {
		return nil, err
	}
	if err := st.hydrateMatch(ctx, match); err != nil {
		return nil, err
	}
	return match, nil
}

func (st *MatchStore) hydrateMatch(ctx context.Context, match *domain.Match) error {
	queue, err := st.findQueue(ctx, match.ID)
	if err != nil {
		return err
	}
	match.Queue = queue

	result, err := st.findResult(ctx, match.ID)
	if err != nil {
		return err
	}
	match.Result = result

	return nil
}

func (st *MatchStore) findQueue(ctx context.Context, id domain.MatchID) ([]domain.TeamID, error) {
	op := "findQueue"

	rows, err := st.store.db.Query(ctx, `
		SELECT team_id
		FROM match_queue
		WHERE match_id = $1
		ORDER BY position
	`, id)
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error selecting match queue", err, apperr.Field{Name: "match_id", Value: id})
	}

	queue, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.TeamID, error) {
		var teamID domain.TeamID
		if err := row.Scan(&teamID); err != nil {
			return 0, err
		}
		return teamID, nil
	})
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error collecting match queue", err, apperr.Field{Name: "match_id", Value: id})
	}
	return queue, nil
}

func (st *MatchStore) findResult(ctx context.Context, id domain.MatchID) (*domain.Result, error) {
	result, err := scanResult(st.store.db.QueryRow(ctx, `
		SELECT id, winner_team_id, eliminated_team_id, result_time_seconds, match_id
		FROM result
		WHERE match_id = $1
	`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, apperr.Wrap("findResult", "unexpected error selecting match result", err, apperr.Field{Name: "match_id", Value: id})
	}
	return result, nil
}

func (st *MatchStore) buildQuery(q domain.MatchQuery) ([]any, string) {
	query := matchSelectQuery() + ` WHERE 1=1`
	var args []any

	if q.TeamAID != 0 {
		args = append(args, q.TeamAID)
		query += fmt.Sprintf(" AND m.team_a_id = $%d", len(args))
	}
	if q.TeamBID != 0 {
		args = append(args, q.TeamBID)
		query += fmt.Sprintf(" AND m.team_b_id = $%d", len(args))
	}
	if q.CategoryID != 0 {
		args = append(args, q.CategoryID)
		query += fmt.Sprintf(" AND m.category_id = $%d", len(args))
	}

	query += " ORDER BY m.id"
	return args, query
}

type matchRow interface {
	Scan(dest ...any) error
}

func scanMatch(row matchRow) (*domain.Match, error) {
	var id domain.MatchID
	var categoryID domain.CategoryID
	var bracketID sql.NullString
	var bracketKey sql.NullString
	var bracketRound sql.NullInt64
	var bracketSlot sql.NullInt64
	var status string
	var teamAID sql.NullInt64
	var teamAName sql.NullString
	var teamASchool sql.NullString
	var teamAGrade sql.NullString
	var teamATeacher sql.NullString
	var teamACategoryID sql.NullInt64
	var teamBID sql.NullInt64
	var teamBName sql.NullString
	var teamBSchool sql.NullString
	var teamBGrade sql.NullString
	var teamBTeacher sql.NullString
	var teamBCategoryID sql.NullInt64

	err := row.Scan(
		&id,
		&categoryID,
		&bracketID,
		&bracketKey,
		&bracketRound,
		&bracketSlot,
		&status,
		&teamAID,
		&teamAName,
		&teamASchool,
		&teamAGrade,
		&teamATeacher,
		&teamACategoryID,
		&teamBID,
		&teamBName,
		&teamBSchool,
		&teamBGrade,
		&teamBTeacher,
		&teamBCategoryID,
	)
	if err != nil {
		return nil, err
	}

	match := &domain.Match{
		ID:           id,
		Queue:        []domain.TeamID{},
		CategoryID:   categoryID,
		BracketID:    bracketID.String,
		BracketKey:   bracketKey.String,
		BracketRound: int(bracketRound.Int64),
		BracketSlot:  int(bracketSlot.Int64),
		Status:       domain.MatchStatus(status),
	}

	if teamAID.Valid {
		match.TeamA = &domain.Team{
			ID:         domain.TeamID(teamAID.Int64),
			Name:       teamAName.String,
			School:     teamASchool.String,
			Grade:      teamAGrade.String,
			Teacher:    teamATeacher.String,
			CategoryID: domain.CategoryID(teamACategoryID.Int64),
		}
	}
	if teamBID.Valid {
		match.TeamB = &domain.Team{
			ID:         domain.TeamID(teamBID.Int64),
			Name:       teamBName.String,
			School:     teamBSchool.String,
			Grade:      teamBGrade.String,
			Teacher:    teamBTeacher.String,
			CategoryID: domain.CategoryID(teamBCategoryID.Int64),
		}
	}

	return match, nil
}

func matchSelectQuery() string {
	return `
		SELECT
			m.id,
			m.category_id,
			m.bracket_id,
			m.bracket_key,
			m.bracket_round,
			m.bracket_slot,
			m.status,
			ta.id,
			ta.name,
			ta.school,
			ta.grade,
			ta.teacher,
			ta.category_id,
			tb.id,
			tb.name,
			tb.school,
			tb.grade,
			tb.teacher,
			tb.category_id
		FROM "match" m
		LEFT JOIN team ta ON ta.id = m.team_a_id
		LEFT JOIN team tb ON tb.id = m.team_b_id
	`
}

func matchTeamID(team *domain.Team) any {
	if team == nil {
		return nil
	}
	return team.ID
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableInt(value int, key string) any {
	if key == "" {
		return nil
	}
	return value
}

func matchStatus(status domain.MatchStatus) domain.MatchStatus {
	if status == "" {
		return domain.MatchStatusReady
	}
	return status
}
