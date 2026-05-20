package domain

type MemberID int

type Member struct {
	ID       MemberID `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	IsLeader bool     `json:"is_leader"`
	TeamID   TeamID   `json:"team_id"`
}

type MemberQuery struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsLeader *bool  `json:"is_leader"`
	TeamID   TeamID `json:"team_id"`
}

func NewMember(name, email string, isLeader bool, teamID TeamID) (*Member, error) {
	if name == "" {
		return nil, NewRobotErr("NewMember", "", "name", name, ErrEmpty, "member name cannot be empty")
	}

	return &Member{
		Name:     name,
		Email:    email,
		IsLeader: isLeader,
		TeamID:   teamID,
	}, nil
}
