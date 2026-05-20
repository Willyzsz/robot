package domain

type TeamID int

type Team struct {
	ID         TeamID     `json:"id"`
	Name       string     `json:"name"`
	School     string     `json:"school"`
	Grade      string     `json:"grade"`
	Teacher    string     `json:"teacher"`
	Members    []*Member  `json:"members"`
	CategoryID CategoryID `json:"category_id"`
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
		return nil, NewRobotErr(op, "", "name", name, ErrEmpty, "name cannot be empty")
	}

	if school == "" {
		return nil, NewRobotErr(op, "", "school", school, ErrEmpty, "school cannot be empty")
	}

	if grade == "" {
		return nil, NewRobotErr(op, "", "grade", grade, ErrEmpty, "grade cannot be empty")
	}

	if teacher == "" {
		return nil, NewRobotErr(op, "", "teacher", teacher, ErrEmpty, "teacher cannot be empty")
	}

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
		return NewRobotErr(op, "", "members", len(t.Members), ErrNotEnough, "team must have at least 1 member")
	}

	leaderCount := 0
	for _, member := range t.Members {
		if member.IsLeader {
			leaderCount++
		}
	}
	if leaderCount != 1 {
		return NewRobotErr(op, "", "members", leaderCount, ErrInvalid, "there must be exactly one leader")
	}

	return nil
}
