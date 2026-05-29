package domain

import "context"

type CategoryRepository interface {
	// Insert adds a new category to the repository and returns its ID.
	// Returns an error wrapping ErrAlreadyExists if a category with the same name already exists.
	Insert(ctx context.Context, category *Category) (CategoryID, error)

	// FindByID retrieves a category by its ID.
	// Returns an error wrapping ErrNotFound if no category with the given ID exists.
	FindByID(ctx context.Context, id CategoryID) (*Category, error)

	// FindByName retrieves a category by its name.
	// Returns an error wrapping ErrNotFound if no category with the given name exists.
	FindByName(ctx context.Context, name string) (*Category, error)

	// FindAll retrieves all categories from the repository.
	FindAll(ctx context.Context) ([]*Category, error)
}

type RuleRepository interface {
	// Insert adds a new rule to the repository and returns its ID.
	// Returns an error wrapping ErrInvalidReference if category_id does not reference an existing category.
	Insert(ctx context.Context, rule *Rule) (RuleID, error)

	// FindByID retrieves a rule by its ID.
	// Returns an error wrapping ErrNotFound if no rule with the given ID exists.
	FindByID(ctx context.Context, id RuleID) (*Rule, error)

	// FindByCategoryID retrieves all rules for the given category.
	FindByCategoryID(ctx context.Context, id CategoryID) ([]*Rule, error)

	// FindAll retrieves all rules from the repository.
	FindAll(ctx context.Context) ([]*Rule, error)
}

type TeamRepository interface {
	// Insert adds a new team to the repository and returns its ID.
	// Returns an error wrapping ErrAlreadyExists if a team with the same name already exists.
	// Returns an error wrapping ErrInvalidReference if the category_id does not reference an existing category.
	Insert(ctx context.Context, team *Team) (TeamID, error)

	// FindByID retrieves a team by its ID.
	// Returns an error wrapping ErrNotFound if no team with the given ID exists.
	FindByID(ctx context.Context, id TeamID) (*Team, error)

	// FindByName retrieves a team by its name.
	// Returns an error wrapping ErrNotFound if no team with the given name exists.
	FindByName(ctx context.Context, name string) (*Team, error)

	// Find retrieves all teams from the given query.
	// Returns an error wrapping ErrNotFound if no teams matching the query exist.
	Find(ctx context.Context, t TeamQuery) ([]*Team, error)

	// FindAll retrieves all teams from the repository.
	FindAll(ctx context.Context) ([]*Team, error)
}

type MemberRepository interface {
	// Insert adds a new member to the repository and returns its ID.
	// Returns an error wrapping ErrAlreadyExists if a member with the same email already exists.
	// Returns an error wrapping ErrInvalidReference if the team_id does not reference an existing team.
	Insert(ctx context.Context, member *Member, teamID TeamID) (MemberID, error)

	// FindByID retrieves a member by its ID.
	// Returns an error wrapping ErrNotFound if no member with the given ID exists.
	FindByID(ctx context.Context, id MemberID) (*Member, error)

	// Find retrieves all members from the given query.
	// Returns an error wrapping ErrNotFound if no members matching the query exist.
	Find(ctx context.Context, m MemberQuery) ([]*Member, error)
}

type MatchRepository interface {
	// Insert adds a new match to the repository and returns its ID.
	// Returns an error wrapping ErrInvalidReference if any team_id or category_id does not reference an existing row.
	Insert(ctx context.Context, match *Match) (MatchID, error)

	// FindByID retrieves a match by its ID.
	// Returns an error wrapping ErrNotFound if no match with the given ID exists.
	FindByID(ctx context.Context, id MatchID) (*Match, error)

	// Find retrieves all matches from the given query.
	Find(ctx context.Context, q MatchQuery) ([]*Match, error)

	// FindAll retrieves all matches from the repository.
	FindAll(ctx context.Context) ([]*Match, error)
}

type ResultRepository interface {
	// Insert adds a new result to the repository and returns its ID.
	// Returns an error wrapping ErrAlreadyExists if the match already has a result.
	// Returns an error wrapping ErrInvalidReference if the match_id or winner does not reference an existing row.
	Insert(ctx context.Context, result *Result) (ResultID, error)

	// FindByID retrieves a result by its ID.
	// Returns an error wrapping ErrNotFound if no result with the given ID exists.
	FindByID(ctx context.Context, id ResultID) (*Result, error)

	// FindByMatchID retrieves a result by its match ID.
	// Returns an error wrapping ErrNotFound if no result with the given match ID exists.
	FindByMatchID(ctx context.Context, id MatchID) (*Result, error)

	// Find retrieves all results from the given query.
	Find(ctx context.Context, q ResultQuery) ([]*Result, error)

	// FindAll retrieves all results from the repository.
	FindAll(ctx context.Context) ([]*Result, error)
}

type RobotRepository interface {
	// Insert adds a new robot to the repository and returns its ID.
	// Returns an error wrapping ErrInvalidReference if team_id or any valid rule_id does not reference an existing row.
	Insert(ctx context.Context, robot *Robot) (RobotID, error)

	// FindByID retrieves a robot by its ID.
	// Returns an error wrapping ErrNotFound if no robot with the given ID exists.
	FindByID(ctx context.Context, id RobotID) (*Robot, error)

	// Find retrieves all robots from the given query.
	Find(ctx context.Context, q RobotQuery) ([]*Robot, error)

	// FindAll retrieves all robots from the repository.
	FindAll(ctx context.Context) ([]*Robot, error)

	// AddValidRule records that a robot was valid under a rule.
	AddValidRule(ctx context.Context, robotID RobotID, ruleID RuleID) error

	// SetValidity updates the final robot validity.
	SetValidity(ctx context.Context, robotID RobotID, isValid bool) error
}
