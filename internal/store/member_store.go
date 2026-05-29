package store

import (
	"context"
	"errors"
	"fmt"
	"robot/internal/domain"
	"robot/pkg/apperr"

	"github.com/jackc/pgx/v5"
)

type MemberStore struct {
	store *Store
}

var _ domain.MemberRepository = (*MemberStore)(nil)

func NewMemberStore(s *Store) *MemberStore {
	return &MemberStore{
		store: s,
	}
}

func (st *MemberStore) Insert(ctx context.Context, member *domain.Member, teamID domain.TeamID) (domain.MemberID, error) {
	query := `
		INSERT INTO member
		(name, email, is_leader, team_id)
		VALUES
		($1, $2, $3, $4)
		RETURNING id
	`

	var id domain.MemberID
	err := st.store.db.QueryRow(ctx, query,
		member.Name,
		member.Email,
		member.IsLeader,
		teamID,
	).Scan(&id)
	if err != nil {
		return 0, apperr.Wrap("Insert", "unexpected error inserting member", err, apperr.Field{Name: "database", Value: member.Name})
	}

	return id, nil
}

func (st *MemberStore) FindByID(ctx context.Context, id domain.MemberID) (*domain.Member, error) {
	op := "FindByID"

	query := `
		SELECT id, name, email, is_leader, team_id
		FROM member
		WHERE id = $1
	`

	var foundID domain.MemberID
	var memberName, memberEmail string
	var isLeader bool
	var teamID domain.TeamID

	err := st.store.db.QueryRow(ctx, query, id).Scan(
		&foundID,
		&memberName,
		&memberEmail,
		&isLeader,
		&teamID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Wrap(op, "", domain.ErrNotFound, apperr.Field{Name: "id", Value: id})
		}
		return nil, apperr.Wrap(op, "unexpected error selecting id", err, apperr.Field{Name: "database", Value: id})
	}

	member, err := domain.NewMember(memberName, memberEmail, isLeader, teamID)
	if err != nil {
		return nil, err
	}
	member.ID = id
	return member, nil
}

func (st *MemberStore) Find(ctx context.Context, q domain.MemberQuery) ([]*domain.Member, error) {
	op := "Find"

	args, query := st.buildQuery(q)

	rows, err := st.store.db.Query(ctx, query, args...)
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error finding members", err, apperr.Field{Name: "database", Value: ""})
	}

	members, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*domain.Member, error) {
		var foundID domain.MemberID
		var memberName, memberEmail string
		var isLeader bool
		var teamID domain.TeamID

		err := row.Scan(
			&foundID,
			&memberName,
			&memberEmail,
			&isLeader,
			&teamID,
		)
		if err != nil {
			return nil, apperr.Wrap(op, "unexpected error scanning member", err, apperr.Field{Name: "scan", Value: row})
		}

		member, err := domain.NewMember(memberName, memberEmail, isLeader, teamID)
		if err != nil {
			return nil, err
		}
		member.ID = foundID
		return member, nil
	})
	if err != nil {
		return nil, apperr.Wrap(op, "unexpected error collecting members", err, apperr.Field{Name: "database", Value: ""})
	}

	return members, nil
}

func (st *MemberStore) buildQuery(q domain.MemberQuery) ([]any, string) {
	query := `
		SELECT id, name, email, is_leader, team_id
		FROM member
		WHERE 1=1
	`
	var args []any
	if q.Name != "" {
		args = append(args, q.Name)
		query += fmt.Sprintf(" AND name = $%d", len(args))
	}
	if q.Email != "" {
		args = append(args, q.Email)
		query += fmt.Sprintf(" AND email = $%d", len(args))
	}
	if q.IsLeader != nil {
		args = append(args, q.IsLeader)
		query += fmt.Sprintf(" AND is_leader = $%d", len(args))
	}
	if q.TeamID != 0 {
		args = append(args, q.TeamID)
		query += fmt.Sprintf(" AND team_id = $%d", len(args))
	}
	return args, query
}
