package domain

import (
	"robot/pkg/apperr"
	"time"
)

type ResultID int

type Result struct {
	ID      ResultID   `json:"id"`
	Winner  TeamID     `json:"winner"`
	Time    *time.Time `json:"time,omitempty"`
	MatchID MatchID    `json:"match_id"`
}

type ResultQuery struct {
	Winner  TeamID  `json:"winner"`
	MatchID MatchID `json:"match_id"`
}

func NewResult(winner TeamID, matchID MatchID, resultTime *time.Time) (*Result, error) {
	op := "NewResult"
	if winner == 0 {
		return nil, apperr.Wrap(op, "winner cannot be empty", ErrEmpty, apperr.Field{Name: "winner", Value: winner})
	}
	if matchID == 0 {
		return nil, apperr.Wrap(op, "match_id cannot be empty", ErrEmpty, apperr.Field{Name: "match_id", Value: matchID})
	}

	return &Result{
		Winner:  winner,
		Time:    resultTime,
		MatchID: matchID,
	}, nil
}

func NewResultForMatch(match *Match, winner TeamID, resultTime *time.Time) (*Result, error) {
	op := "NewResultForMatch"
	if match == nil {
		return nil, apperr.Wrap(op, "match cannot be nil", ErrInvalid)
	}
	if winner != match.TeamA.ID && winner != match.TeamB.ID {
		return nil, apperr.Wrap(op, "winner must be one of the match teams", ErrInvalid, apperr.Field{Name: "winner", Value: winner})
	}

	return NewResult(winner, match.ID, resultTime)
}
