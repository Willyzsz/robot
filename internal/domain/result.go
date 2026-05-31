package domain

import "robot/pkg/apperr"

type ResultID int

type Result struct {
	ID               ResultID    `json:"id"`
	Winner           TeamID      `json:"team_id"`
	EliminatedTeamID *TeamID     `json:"eliminated_team_id,omitempty"`
	Time             *ResultTime `json:"time,omitempty"`
	MatchID          MatchID     `json:"match_id"`
}

type ResultTime struct {
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

type ResultQuery struct {
	Winner  TeamID  `json:"winner"`
	MatchID MatchID `json:"match_id"`
}

func NewResultTime(minutes, seconds int) (*ResultTime, error) {
	op := "NewResultTime"
	if minutes < 0 {
		return nil, apperr.Wrap(op, "minutes cannot be negative", ErrInvalid, apperr.Field{Name: "minutes", Value: minutes})
	}
	if seconds < 0 || seconds > 59 {
		return nil, apperr.Wrap(op, "seconds must be between 0 and 59", ErrInvalid, apperr.Field{Name: "seconds", Value: seconds})
	}
	return &ResultTime{Minutes: minutes, Seconds: seconds}, nil
}

func NewResultTimeFromSeconds(totalSeconds int) (*ResultTime, error) {
	if totalSeconds < 0 {
		return nil, apperr.Wrap("NewResultTimeFromSeconds", "time cannot be negative", ErrInvalid, apperr.Field{Name: "time", Value: totalSeconds})
	}
	return NewResultTime(totalSeconds/60, totalSeconds%60)
}

func (t ResultTime) TotalSeconds() int {
	return t.Minutes*60 + t.Seconds
}

func NewResult(winner TeamID, matchID MatchID, eliminatedTeamID *TeamID, resultTime *ResultTime) (*Result, error) {
	op := "NewResult"
	if winner == 0 {
		return nil, apperr.Wrap(op, "team_id cannot be empty", ErrEmpty, apperr.Field{Name: "team_id", Value: winner})
	}
	if matchID == 0 {
		return nil, apperr.Wrap(op, "match_id cannot be empty", ErrEmpty, apperr.Field{Name: "match_id", Value: matchID})
	}
	if eliminatedTeamID != nil {
		if *eliminatedTeamID == 0 {
			return nil, apperr.Wrap(op, "eliminated_team_id cannot be empty", ErrEmpty, apperr.Field{Name: "eliminated_team_id", Value: *eliminatedTeamID})
		}
		if *eliminatedTeamID == winner {
			return nil, apperr.Wrap(op, "winner cannot also be eliminated", ErrInvalid, apperr.Field{Name: "team_id", Value: winner})
		}
	}
	if resultTime != nil {
		if _, err := NewResultTime(resultTime.Minutes, resultTime.Seconds); err != nil {
			return nil, err
		}
	}

	return &Result{
		Winner:           winner,
		EliminatedTeamID: eliminatedTeamID,
		Time:             resultTime,
		MatchID:          matchID,
	}, nil
}

func NewResultForMatch(match *Match, winner TeamID, eliminatedTeamID *TeamID, resultTime *ResultTime) (*Result, error) {
	op := "NewResultForMatch"
	if match == nil {
		return nil, apperr.Wrap(op, "match cannot be nil", ErrInvalid)
	}
	if !match.hasTeam(winner) {
		return nil, apperr.Wrap(op, "team_id must be one of the match teams", ErrInvalid, apperr.Field{Name: "team_id", Value: winner})
	}
	if eliminatedTeamID == nil {
		eliminatedTeamID = match.eliminatedTeamFor(winner)
	}
	if eliminatedTeamID != nil && !match.hasTeam(*eliminatedTeamID) {
		return nil, apperr.Wrap(op, "eliminated_team_id must be one of the match teams", ErrInvalid, apperr.Field{Name: "eliminated_team_id", Value: *eliminatedTeamID})
	}

	return NewResult(winner, match.ID, eliminatedTeamID, resultTime)
}
