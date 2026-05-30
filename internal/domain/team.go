package domain

import "robot/pkg/apperr"

type TeamID int

type Team struct {
	ID         TeamID     `json:"id"`
	Name       string     `json:"name"`
	School     string     `json:"school"`
	Grade      string     `json:"grade"`
	Teacher    string     `json:"teacher"`
	Members    []*Member  `json:"members"`
	RobotValid bool       `json:"robot_valid"`
	CategoryID CategoryID `json:"category_id"`
	Category   string     `json:"category,omitempty"`
}

type TeamQuery struct {
	Name       string     `json:"name"`
	School     string     `json:"school"`
	Grade      string     `json:"grade"`
	Teacher    string     `json:"teacher"`
	CategoryID CategoryID `json:"category_id"`
}

func NewTeam(name, school, grade, teacher string, categoryID CategoryID) (*Team, error) {
	op := "NewTeam"
	if name == "" {
		return nil, apperr.Wrap(op, "name cannot be empty", ErrEmpty, apperr.Field{Name: "name", Value: name})
	}

	// if school == "" {
	// 	return nil, apperr.Wrap(op, "school cannot be empty", ErrEmpty, apperr.Field{Name: "school", Value: school})
	// }

	// if grade == "" {
	// 	return nil, apperr.Wrap(op, "grade cannot be empty", ErrEmpty, apperr.Field{Name: "grade", Value: grade})
	// }

	// if teacher == "" {
	// 	return nil, apperr.Wrap(op, "teacher cannot be empty", ErrEmpty, apperr.Field{Name: "teacher", Value: teacher})
	// }

	return &Team{
		Name:       name,
		School:     school,
		Grade:      grade,
		Teacher:    teacher,
		Members:    []*Member{},
		CategoryID: categoryID,
	}, nil
}

func (t *Team) AddMember(member *Member) error {
	if member == nil {
		return ErrInvalid
	}

	if t.ID != member.TeamID {
		return ErrInvalid
	}

	for _, existing := range t.Members {
		if existing.ID != 0 && member.ID != 0 && existing.ID == member.ID {
			return ErrAlreadyExists
		}
		if existing.Email != "" && existing.Email == member.Email {
			return ErrAlreadyExists
		}
	}

	t.Members = append(t.Members, member)
	return nil
}

func (t *Team) ValidateMembers() error {
	op := "ValidateMembers"

	if len(t.Members) < 1 {
		return apperr.Wrap(op, "team must have at least 1 member", ErrNotEnough, apperr.Field{Name: "members", Value: len(t.Members)})
	}

	leaderCount := 0
	for _, member := range t.Members {
		if member.IsLeader {
			leaderCount++
		}
	}
	if leaderCount != 1 {
		return apperr.Wrap(op, "there must be exactly one leader", ErrInvalid, apperr.Field{Name: "members", Value: leaderCount})
	}

	return nil
}
