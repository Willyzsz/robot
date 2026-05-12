package domain

type RuleID int

type Rule struct {
	ID 			RuleID
	Description string
	CategoryID  CategoryID
}

func NewRule(description string, categoryID CategoryID) (*Rule, error) {
	if description == "" {
		return nil, NewRobotErr("newRule", "", "description", description, ErrEmpty, "description cannot be empty")
	}

	return &Rule{
		Description: description,
		CategoryID: categoryID,
	}, nil
}