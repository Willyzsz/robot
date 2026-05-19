package domain

import "context"

type CategoryRepository interface {
	// Insert adds a new category to the repository and returns its ID.
	// Returns an RobotErr wrapping ErrAlreadyExists if a category with the same name already exists.
	Insert(ctx context.Context, category *Category) (CategoryID, error)
	
	// FindByID retrieves a category by its ID.
	// Returns an RobotErr wrapping ErrNotFound if no category with the given ID exists.
	FindByID(ctx context.Context, id CategoryID) (*Category, error)
	
	// FindByName retrieves a category by its name.
	// Returns an RobotErr wrapping ErrNotFound if no category with the given name exists.
	FindByName(ctx context.Context, name string) (*Category, error)
	
	// FindAll retrieves all categories from the repository.
	FindAll(ctx context.Context) ([]*Category, error)
}

type TeamRepository interface {
	// Insert adds a new team to the repository and returns its ID.
	// Returns an RobotErr wrapping ErrAlreadyExists if a team with the same name already exists.
	// Returns an RobotErr wrapping ErrInvalidReference if the category_id does not reference an existing category.
	Insert(ctx context.Context, team *Team) (TeamID, error)
	
	// FindByID retrieves a team by its ID.
	// Returns an RobotErr wrapping ErrNotFound if no team with the given ID exists.
	FindByID(ctx context.Context, id TeamID) (*Team, error)
	
	// FindByName retrieves a team by its name.
	// Returns an RobotErr wrapping ErrNotFound if no team with the given name exists.
	FindByName(ctx context.Context, name string) (*Team, error)
	
	// Find retrieves all teams from the given query.
	// Returns an RobotErr wrapping ErrNotFound if no teams matching the query exist.
	Find(ctx context.Context, t TeamQuery) ([]*Team, error)

	// FindAll retrieves all teams from the repository.
	FindAll(ctx context.Context) ([]*Team, error)
}

type MemberRepository interface {
	// Insert adds a new member to the repository and returns its ID.
	// Returns an RobotErr wrapping ErrAlreadyExists if a member with the same email already exists.
	// Returns an RobotErr wrapping ErrInvalidReference if the team_id does not reference an existing team.
	Insert(ctx context.Context, member *Member, teamID TeamID) (MemberID, error)
	
	// FindByID retrieves a member by its ID.
	// Returns an RobotErr wrapping ErrNotFound if no member with the given ID exists.
	FindByID(ctx context.Context, id MemberID) (*Member, error)
	
	// Find retrieves all members from the given query.
	// Returns an RobotErr wrapping ErrNotFound if no members matching the query exist.
	Find(ctx context.Context, m MemberQuery) ([]*Member, error)
}