package domain

type MemberID int

type Member struct {
	ID		MemberID
	Name 	string
	Email 	string
	IsLeader bool
	TeamID TeamID
}

type MemberQuery struct {
	Name 	string
	Email 	string
	IsLeader *bool
	TeamID 	TeamID
}

func NewMember(name, email string, isLeader bool, teamID TeamID) (*Member, error) {
	if name == "" {
		return nil, NewRobotErr("NewMember", "", "name", name, ErrEmpty, "member name cannot be empty")
	}

	return &Member{
		Name: name,
		Email: email,
		IsLeader: isLeader,
		TeamID: teamID,
	}, nil
}