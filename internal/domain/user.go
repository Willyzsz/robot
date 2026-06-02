package domain

import "robot/pkg/apperr"

type UserID int
type UserRole string

const (
	UserRoleJuez      UserRole = "juez"
	UserRoleVisitante UserRole = "visitante"
	UserRoleArbitro   UserRole = "arbitro"
	UserRoleAdmin     UserRole = "admin"
	UserRoleDev       UserRole = "dev"
)

type User struct {
	ID           UserID       `json:"id"`
	Username     string       `json:"username"`
	Name         string       `json:"name"`
	Role         UserRole     `json:"role"`
	CategoryIDs  []CategoryID `json:"category_ids,omitempty"`
	PasswordHash string       `json:"-"`
}

func NewUser(username, name string, role UserRole, passwordHash string) (*User, error) {
	if username == "" {
		return nil, apperr.Wrap("NewUser", "username cannot be empty", ErrEmpty, apperr.Field{Name: "username", Value: username})
	}
	if !role.Valid() {
		return nil, apperr.Wrap("NewUser", "role is invalid", ErrInvalid, apperr.Field{Name: "role", Value: role})
	}
	if passwordHash == "" {
		return nil, apperr.Wrap("NewUser", "password_hash cannot be empty", ErrEmpty)
	}
	return &User{
		Username:     username,
		Name:         name,
		Role:         role,
		CategoryIDs:  []CategoryID{},
		PasswordHash: passwordHash,
	}, nil
}

func (r UserRole) Valid() bool {
	return r == UserRoleJuez ||
		r == UserRoleVisitante ||
		r == UserRoleArbitro ||
		r == UserRoleAdmin ||
		r == UserRoleDev
}
