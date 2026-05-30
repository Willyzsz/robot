package domain

import "robot/pkg/apperr"

type RuleID int
type RuleType string

const (
	RuleTypeCharacteristic RuleType = "characteristic"
	RuleTypeRestriction    RuleType = "restriction"
)

type Rule struct {
	ID          RuleID     `json:"id"`
	Description string     `json:"description"`
	Type        RuleType   `json:"type"`
	CategoryID  CategoryID `json:"category_id"`
}

func NewRule(description string, ruleType RuleType, categoryID CategoryID) (*Rule, error) {
	if description == "" {
		return nil, apperr.Wrap("NewRule", "description cannot be empty", ErrEmpty, apperr.Field{Name: "description", Value: description})
	}
	if !ruleType.Valid() {
		return nil, apperr.Wrap("NewRule", "rule type is invalid", ErrInvalid, apperr.Field{Name: "type", Value: ruleType})
	}

	return &Rule{
		Description: description,
		Type:        ruleType,
		CategoryID:  categoryID,
	}, nil
}

func (rt RuleType) Valid() bool {
	return rt == RuleTypeCharacteristic || rt == RuleTypeRestriction
}

func (r *Rule) ChangeDescription(description string) error {
	if description == "" {
		return apperr.Wrap("ChangeDescription", "description cannot be empty", ErrEmpty, apperr.Field{Name: "description", Value: description})
	}
	r.Description = description
	return nil
}
