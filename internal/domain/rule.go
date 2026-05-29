package domain

import "robot/pkg/apperr"

type RuleID int

type Rule struct {
	ID          RuleID     `json:"id"`
	Description string     `json:"description"`
	CategoryID  CategoryID `json:"category_id"`
}

func NewRule(description string, categoryID CategoryID) (*Rule, error) {
	if description == "" {
		return nil, apperr.Wrap("NewRule", "description cannot be empty", ErrEmpty, apperr.Field{Name: "description", Value: description})
	}

	return &Rule{
		Description: description,
		CategoryID:  categoryID,
	}, nil
}

func (r *Rule) ChangeDescription(description string) error {
	if description == "" {
		return apperr.Wrap("ChangeDescription", "description cannot be empty", ErrEmpty, apperr.Field{Name: "description", Value: description})
	}
	r.Description = description
	return nil
}
