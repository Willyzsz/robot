package domain

import "slices"

type CategoryID int

type Category struct {
	ID		CategoryID
	Name 	string
	Rules 	[]*Rule
}


func NewCategory(name string) (*Category, error) {
	if name == "" {
		return nil, NewRobotErr("NewCategory", "", "name", name, ErrEmpty, "category name cannot be emtpy")
	}

	return &Category{
		Name: name,
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