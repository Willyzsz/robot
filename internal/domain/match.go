package domain

import (
	"robot/pkg/apperr"
	"slices"
)

type MatchID int
type MatchMode string
type MatchStatus string

const (
	MatchModePairwise MatchMode = "pairwise"
	MatchModeShared   MatchMode = "shared"

	MatchStatusPending   MatchStatus = "pending"
	MatchStatusReady     MatchStatus = "ready"
	MatchStatusCompleted MatchStatus = "completed"
	MatchStatusBye       MatchStatus = "bye"
)

type Match struct {
	ID           MatchID     `json:"id"`
	TeamA        *Team       `json:"team_a,omitempty"`
	TeamB        *Team       `json:"team_b,omitempty"`
	Queue        []TeamID    `json:"queue,omitempty"`
	CategoryID   CategoryID  `json:"category_id"`
	IsInternal   bool        `json:"is_internal"`
	BracketID    string      `json:"bracket_id,omitempty"`
	BracketKey   string      `json:"bracket_key,omitempty"`
	BracketRound int         `json:"bracket_round,omitempty"`
	BracketSlot  int         `json:"bracket_slot,omitempty"`
	Status       MatchStatus `json:"status,omitempty"`
	Result       *Result     `json:"result,omitempty"`
}

type MatchQuery struct {
	TeamAID    TeamID     `json:"team_a_id"`
	TeamBID    TeamID     `json:"team_b_id"`
	CategoryID CategoryID `json:"category_id"`
}

func NewMatch(teamA, teamB Team, categoryID CategoryID) (*Match, error) {
	return NewPairMatch(teamA, teamB, categoryID)
}

func NewPairMatch(teamA, teamB Team, categoryID CategoryID) (*Match, error) {
	op := "NewPairMatch"
	if teamA.ID == 0 {
		return nil, apperr.Wrap(op, "team_a cannot be empty", ErrEmpty, apperr.Field{Name: "team_a", Value: teamA.ID})
	}
	if teamB.ID == 0 {
		return nil, apperr.Wrap(op, "team_b cannot be empty", ErrEmpty, apperr.Field{Name: "team_b", Value: teamB.ID})
	}
	if categoryID == 0 {
		return nil, apperr.Wrap(op, "category_id cannot be empty", ErrEmpty, apperr.Field{Name: "category_id", Value: categoryID})
	}
	if teamA.ID == teamB.ID {
		return nil, apperr.Wrap(op, "teams cannot be the same", ErrInvalid, apperr.Field{Name: "team_a", Value: teamA.ID}, apperr.Field{Name: "team_b", Value: teamB.ID})
	}
	if teamA.CategoryID != categoryID || teamB.CategoryID != categoryID {
		return nil, apperr.Wrap(op, "teams must belong to the match category", ErrInvalid, apperr.Field{Name: "category_id", Value: categoryID})
	}

	return &Match{
		TeamA:      &teamA,
		TeamB:      &teamB,
		Queue:      []TeamID{},
		CategoryID: categoryID,
		Status:     MatchStatusReady,
	}, nil
}

func NewQueueMatch(categoryID CategoryID, teams []Team) (*Match, error) {
	op := "NewQueueMatch"
	if categoryID == 0 {
		return nil, apperr.Wrap(op, "category_id cannot be empty", ErrEmpty, apperr.Field{Name: "category_id", Value: categoryID})
	}
	if len(teams) == 0 {
		return nil, apperr.Wrap(op, "queue must have at least one team", ErrNotEnough)
	}

	match := &Match{
		Queue:      []TeamID{},
		CategoryID: categoryID,
	}
	for _, team := range teams {
		if team.ID == 0 {
			return nil, apperr.Wrap(op, "team cannot be empty", ErrEmpty, apperr.Field{Name: "team", Value: team.ID})
		}
		if team.CategoryID != categoryID {
			return nil, apperr.Wrap(op, "team must belong to the match category", ErrInvalid, apperr.Field{Name: "team", Value: team.ID})
		}
		if slices.Contains(match.Queue, team.ID) {
			return nil, apperr.Wrap(op, "team is already in the queue", ErrAlreadyExists, apperr.Field{Name: "team", Value: team.ID})
		}
		match.Queue = append(match.Queue, team.ID)
	}
	return match, nil
}

func (m *Match) AddToQueue(team Team) error {
	if team.CategoryID != m.CategoryID {
		return apperr.Wrap("AddToQueue", "team must belong to the match category", ErrInvalid, apperr.Field{Name: "team", Value: team.ID})
	}
	if m.TeamA != nil && team.ID == m.TeamA.ID {
		return apperr.Wrap("AddToQueue", "team is already in the match", ErrAlreadyExists, apperr.Field{Name: "team", Value: team.ID})
	}
	if m.TeamB != nil && team.ID == m.TeamB.ID {
		return apperr.Wrap("AddToQueue", "team is already in the match", ErrAlreadyExists, apperr.Field{Name: "team", Value: team.ID})
	}
	if slices.Contains(m.Queue, team.ID) {
		return apperr.Wrap("AddToQueue", "team is already in the queue", ErrAlreadyExists, apperr.Field{Name: "team", Value: team.ID})
	}

	m.Queue = append(m.Queue, team.ID)
	return nil
}

func (m *Match) RemoveFromQueue(team Team) error {
	if m.TeamA != nil && team.ID == m.TeamA.ID {
		return apperr.Wrap("RemoveFromQueue", "team is currently in the match", ErrInvalid, apperr.Field{Name: "team", Value: team.ID})
	}
	if m.TeamB != nil && team.ID == m.TeamB.ID {
		return apperr.Wrap("RemoveFromQueue", "team is currently in the match", ErrInvalid, apperr.Field{Name: "team", Value: team.ID})
	}

	before := len(m.Queue)
	m.Queue = slices.DeleteFunc(m.Queue, func(id TeamID) bool {
		return id == team.ID
	})
	if len(m.Queue) == before {
		return apperr.Wrap("RemoveFromQueue", "team is not in the queue", ErrNotFound, apperr.Field{Name: "team", Value: team.ID})
	}
	return nil
}

func (m *Match) StartMatch() error {
	if m.TeamA == nil || m.TeamB == nil {
		if len(m.Queue) == 0 {
			return apperr.Wrap("StartMatch", "match requires queued teams", ErrNotEnough)
		}
		return nil
	}
	if m.TeamA.ID == 0 || m.TeamB.ID == 0 {
		return apperr.Wrap("StartMatch", "match requires two teams", ErrNotEnough, apperr.Field{Name: "team_a", Value: m.TeamA.ID}, apperr.Field{Name: "team_b", Value: m.TeamB.ID})
	}
	if m.TeamA.ID == m.TeamB.ID {
		return apperr.Wrap("StartMatch", "teams cannot be the same", ErrInvalid, apperr.Field{Name: "team_a", Value: m.TeamA.ID}, apperr.Field{Name: "team_b", Value: m.TeamB.ID})
	}
	return nil
}

func (m *Match) SetResult(result *Result) error {
	if result == nil {
		return apperr.Wrap("SetResult", "result cannot be nil", ErrInvalid)
	}
	if m.ID != 0 && result.MatchID != m.ID {
		return apperr.Wrap("SetResult", "result does not belong to match", ErrInvalid, apperr.Field{Name: "match_id", Value: result.MatchID})
	}
	if !m.hasTeam(result.Winner) {
		return apperr.Wrap("SetResult", "winner must be one of the match teams", ErrInvalid, apperr.Field{Name: "winner", Value: result.Winner})
	}
	if result.EliminatedTeamID != nil {
		if !m.hasTeam(*result.EliminatedTeamID) {
			return apperr.Wrap("SetResult", "eliminated team must be one of the match teams", ErrInvalid, apperr.Field{Name: "eliminated_team_id", Value: *result.EliminatedTeamID})
		}
		if *result.EliminatedTeamID == result.Winner {
			return apperr.Wrap("SetResult", "winner cannot also be eliminated", ErrInvalid, apperr.Field{Name: "team_id", Value: result.Winner})
		}
	}
	m.Result = result
	return nil
}

func (m *Match) hasTeam(teamID TeamID) bool {
	if m.TeamA != nil && m.TeamA.ID == teamID {
		return true
	}
	if m.TeamB != nil && m.TeamB.ID == teamID {
		return true
	}
	return slices.Contains(m.Queue, teamID)
}

func (m *Match) eliminatedTeamFor(winner TeamID) *TeamID {
	if m.TeamA == nil || m.TeamB == nil {
		return nil
	}
	if m.TeamA.ID == winner {
		eliminated := m.TeamB.ID
		return &eliminated
	}
	if m.TeamB.ID == winner {
		eliminated := m.TeamA.ID
		return &eliminated
	}
	return nil
}

func (mode MatchMode) Valid() bool {
	return mode == MatchModePairwise || mode == MatchModeShared
}
