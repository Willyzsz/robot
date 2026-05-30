package domain

import "robot/pkg/apperr"

type RobotID int

type Robot struct {
	ID         RobotID  `json:"id"`
	TeamID     TeamID   `json:"team_id"`
	IsValid    bool     `json:"is_valid"`
	ValidRules []RuleID `json:"valid_rules,omitempty"`
	Rules      []*Rule  `json:"rules,omitempty"`
}

type RobotQuery struct {
	TeamID  TeamID `json:"team_id"`
	IsValid *bool  `json:"is_valid"`
}

func NewRobot(teamID TeamID, validRules []RuleID) (*Robot, error) {
	if teamID == 0 {
		return nil, apperr.Wrap("NewRobot", "team_id cannot be empty", ErrEmpty, apperr.Field{Name: "team_id", Value: teamID})
	}

	robot := &Robot{
		TeamID:     teamID,
		ValidRules: []RuleID{},
	}

	for _, ruleID := range validRules {
		if err := robot.AddValidRule(ruleID); err != nil {
			return nil, err
		}
	}

	return robot, nil
}

func (r *Robot) AddValidRule(ruleID RuleID) error {
	if ruleID == 0 {
		return apperr.Wrap("AddValidRule", "rule_id cannot be empty", ErrEmpty, apperr.Field{Name: "rule_id", Value: ruleID})
	}

	for _, existing := range r.ValidRules {
		if existing == ruleID {
			return apperr.Wrap("AddValidRule", "rule already exists for robot", ErrAlreadyExists, apperr.Field{Name: "rule_id", Value: ruleID})
		}
	}

	r.ValidRules = append(r.ValidRules, ruleID)
	return nil
}

func (r *Robot) SetFinalValidity(requiredRuleCount int) {
	r.IsValid = requiredRuleCount > 0 && len(r.ValidRules) == requiredRuleCount
}
