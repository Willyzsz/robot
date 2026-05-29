package domain

import (
	"robot/pkg/apperr"
	"slices"
)

type CategoryID int

type Category struct {
	ID    CategoryID `json:"id"`
	Name  string     `json:"name"`
	Rules []*Rule    `json:"rules"`
}

func NewCategory(name string) (*Category, error) {
	if name == "" {
		return nil, apperr.Wrap("NewCategory", "category name cannot be empty", ErrEmpty, apperr.Field{Name: "name", Value: name})
	}

	return &Category{
		Name:  name,
		Rules: []*Rule{},
	}, nil
}

func (c *Category) AddRule(rule *Rule) error {
	if rule == nil {
		return ErrInvalid
	}

	if rule.CategoryID != c.ID {
		return ErrInvalid
	}

	if slices.ContainsFunc(c.Rules, func(r *Rule) bool {
		return r.ID == rule.ID
	}) {
		return ErrAlreadyExists
	}

	c.Rules = append(c.Rules, rule)
	return nil
}
