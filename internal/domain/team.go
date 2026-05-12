package domain

type TeamID int

type Team struct {
	ID         TeamID
	Name       string
	School     string
	Grade      string
	Members    []*Member
	CategoryID CategoryID
}

func NewTeam(name, school, grade string, members []*Member, categoryID CategoryID) (*Team, error) {
	op := "NewTeam"
	if name == "" || school == "" || grade == "" {
		return nil, ErrEmpty
	}

	if len(members) < 1 {
		return nil, NewRobotErr(op,"", "members", len(members), ErrNotEnough, "members must be greater than 1") 
	}

	leaderCount := 0
	for _, member := range members {
		if member.IsLeader {
			leaderCount++
		}
	}
	switch leaderCount {
	case 0:
		return nil, NewRobotErr(op,"", "leader", leaderCount, ErrNotFound, "team must have one leader")
	case 2:
		return nil, NewRobotErr(op,"", "leader", leaderCount, ErrAlreadyExists, "team cannot have 2 leaders")
	}
	
	return &Team{
		Name:       name,
		School:     school,
		Grade:      grade,
		Members:    members,
		CategoryID: categoryID,
	}, nil
}
