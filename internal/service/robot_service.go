package service

import (
	"context"
	"robot/internal/domain"
	"time"
)

type RobotService struct {
	categoryRepository domain.CategoryRepository
	ruleRepository     domain.RuleRepository
	teamRepository     domain.TeamRepository
	memberRepository   domain.MemberRepository
	matchRepository    domain.MatchRepository
	resultRepository   domain.ResultRepository
	robotRepository    domain.RobotRepository
}

func NewRobotService(categoryRepo domain.CategoryRepository, ruleRepo domain.RuleRepository, teamRepo domain.TeamRepository, memberRepo domain.MemberRepository, matchRepo domain.MatchRepository, resultRepo domain.ResultRepository, robotRepo domain.RobotRepository) *RobotService {
	return &RobotService{
		categoryRepository: categoryRepo,
		ruleRepository:     ruleRepo,
		teamRepository:     teamRepo,
		memberRepository:   memberRepo,
		matchRepository:    matchRepo,
		resultRepository:   resultRepo,
		robotRepository:    robotRepo,
	}
}

func (svc *RobotService) GetAllCategories(ctx context.Context) ([]*domain.Category, error) {
	return svc.categoryRepository.FindAll(ctx)
}

func (svc *RobotService) CreateCategory(ctx context.Context, name string) (domain.CategoryID, error) {
	category, err := domain.NewCategory(name)
	if err != nil {
		return 0, err
	}
	return svc.categoryRepository.Insert(ctx, category)
}

func (svc *RobotService) CreateRule(ctx context.Context, description string, categoryID domain.CategoryID) (domain.RuleID, error) {
	rule, err := domain.NewRule(description, categoryID)
	if err != nil {
		return 0, err
	}
	return svc.ruleRepository.Insert(ctx, rule)
}

func (svc *RobotService) GetRulesByCategoryID(ctx context.Context, categoryID domain.CategoryID) ([]*domain.Rule, error) {
	return svc.ruleRepository.FindByCategoryID(ctx, categoryID)
}

func (svc *RobotService) GetAllTeams(ctx context.Context) ([]*domain.Team, error) {
	return svc.teamRepository.FindAll(ctx)
}

func (svc *RobotService) CreateTeam(ctx context.Context, name, school, grade, teacher string, categoryID domain.CategoryID) (domain.TeamID, error) {
	team, err := domain.NewTeam(name, school, grade, teacher, categoryID)
	if err != nil {
		return 0, err
	}
	return svc.teamRepository.Insert(ctx, team)
}

func (svc *RobotService) GetTeamById(ctx context.Context, id domain.TeamID) (*domain.Team, error) {
	return svc.teamRepository.FindByID(ctx, id)
}

func (svc *RobotService) GetTeamsByCategoryID(ctx context.Context, categoryID domain.CategoryID) ([]*domain.Team, error) {
	return svc.teamRepository.Find(ctx, domain.TeamQuery{CategoryID: categoryID})
}

func (svc *RobotService) CreateMember(ctx context.Context, name, email string, isLeader bool, teamID domain.TeamID) (domain.MemberID, error) {
	member, err := domain.NewMember(name, email, isLeader, teamID)
	if err != nil {
		return 0, err
	}
	return svc.memberRepository.Insert(ctx, member, teamID)
}

func (svc *RobotService) GetMembersByTeamID(ctx context.Context, teamID domain.TeamID) ([]*domain.Member, error) {
	return svc.memberRepository.Find(ctx, domain.MemberQuery{TeamID: teamID})
}

func (svc *RobotService) GetAllMatches(ctx context.Context) ([]*domain.Match, error) {
	return svc.matchRepository.FindAll(ctx)
}

func (svc *RobotService) CreateMatch(ctx context.Context, teamAID, teamBID domain.TeamID, queueIDs []domain.TeamID) (domain.MatchID, error) {
	teamA, err := svc.teamRepository.FindByID(ctx, teamAID)
	if err != nil {
		return 0, err
	}

	teamB, err := svc.teamRepository.FindByID(ctx, teamBID)
	if err != nil {
		return 0, err
	}

	match, err := domain.NewMatch(*teamA, *teamB, teamA.CategoryID)
	if err != nil {
		return 0, err
	}

	for _, teamID := range queueIDs {
		team, err := svc.teamRepository.FindByID(ctx, teamID)
		if err != nil {
			return 0, err
		}
		if err := match.AddToQueue(*team); err != nil {
			return 0, err
		}
	}

	return svc.matchRepository.Insert(ctx, match)
}

func (svc *RobotService) GetMatchByID(ctx context.Context, id domain.MatchID) (*domain.Match, error) {
	return svc.matchRepository.FindByID(ctx, id)
}

func (svc *RobotService) GetAllResults(ctx context.Context) ([]*domain.Result, error) {
	return svc.resultRepository.FindAll(ctx)
}

func (svc *RobotService) CreateResult(ctx context.Context, matchID domain.MatchID, winner domain.TeamID, resultTime *time.Time) (domain.ResultID, error) {
	match, err := svc.matchRepository.FindByID(ctx, matchID)
	if err != nil {
		return 0, err
	}

	result, err := domain.NewResultForMatch(match, winner, resultTime)
	if err != nil {
		return 0, err
	}

	return svc.resultRepository.Insert(ctx, result)
}

func (svc *RobotService) GetResultByMatchID(ctx context.Context, id domain.MatchID) (*domain.Result, error) {
	return svc.resultRepository.FindByMatchID(ctx, id)
}

func (svc *RobotService) GetAllRobots(ctx context.Context) ([]*domain.Robot, error) {
	return svc.robotRepository.FindAll(ctx)
}

func (svc *RobotService) CreateRobot(ctx context.Context, teamID domain.TeamID, validRules []domain.RuleID) (domain.RobotID, error) {
	team, err := svc.teamRepository.FindByID(ctx, teamID)
	if err != nil {
		return 0, err
	}

	robot, err := svc.buildRobotForTeam(ctx, team, validRules)
	if err != nil {
		return 0, err
	}

	return svc.robotRepository.Insert(ctx, robot)
}

func (svc *RobotService) GetRobotByID(ctx context.Context, id domain.RobotID) (*domain.Robot, error) {
	return svc.robotRepository.FindByID(ctx, id)
}

func (svc *RobotService) GetRobotsByTeamID(ctx context.Context, teamID domain.TeamID) ([]*domain.Robot, error) {
	return svc.robotRepository.Find(ctx, domain.RobotQuery{TeamID: teamID})
}

func (svc *RobotService) VerifyRobot(ctx context.Context, robotID domain.RobotID, ruleID domain.RuleID) error {
	robot, err := svc.robotRepository.FindByID(ctx, robotID)
	if err != nil {
		return err
	}

	team, err := svc.teamRepository.FindByID(ctx, robot.TeamID)
	if err != nil {
		return err
	}

	rule, err := svc.ruleRepository.FindByID(ctx, ruleID)
	if err != nil {
		return err
	}
	if rule.CategoryID != team.CategoryID {
		return domain.ErrInvalid
	}

	if err := robot.AddValidRule(ruleID); err != nil {
		return err
	}

	categoryRules, err := svc.ruleRepository.FindByCategoryID(ctx, team.CategoryID)
	if err != nil {
		return err
	}
	robot.SetFinalValidity(len(categoryRules))

	if err := svc.robotRepository.AddValidRule(ctx, robotID, ruleID); err != nil {
		return err
	}
	return svc.robotRepository.SetValidity(ctx, robotID, robot.IsValid)
}

func (svc *RobotService) buildRobotForTeam(ctx context.Context, team *domain.Team, validRules []domain.RuleID) (*domain.Robot, error) {
	robot, err := domain.NewRobot(team.ID, validRules)
	if err != nil {
		return nil, err
	}

	categoryRules, err := svc.ruleRepository.FindByCategoryID(ctx, team.CategoryID)
	if err != nil {
		return nil, err
	}

	categoryRuleIDs := make(map[domain.RuleID]struct{}, len(categoryRules))
	for _, rule := range categoryRules {
		categoryRuleIDs[rule.ID] = struct{}{}
	}

	for _, ruleID := range robot.ValidRules {
		if _, ok := categoryRuleIDs[ruleID]; !ok {
			return nil, domain.ErrInvalid
		}
	}

	robot.SetFinalValidity(len(categoryRules))
	return robot, nil
}

func (svc *RobotService) GetTeamsWithMembersAndCategory(ctx context.Context) ([]*domain.Team, error) {
	teams, err := svc.teamRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, team := range teams {
		members, err := svc.memberRepository.Find(ctx, domain.MemberQuery{TeamID: team.ID})
		if err != nil {
			return nil, err
		}
		team.Members = members

		category, err := svc.categoryRepository.FindByID(ctx, team.CategoryID)
		if err != nil {
			return nil, err
		}
		team.Category = category.Name
	}

	return teams, nil
}
