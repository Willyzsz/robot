package service

import (
	"context"
	"robot/internal/domain"
	"strings"
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

func (svc *RobotService) CreateRule(ctx context.Context, description string, ruleType domain.RuleType, categoryID domain.CategoryID) (domain.RuleID, error) {
	rule, err := domain.NewRule(description, ruleType, categoryID)
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

func (svc *RobotService) StartMatchQueue(ctx context.Context, categoryID domain.CategoryID, mode domain.MatchMode) ([]*domain.Match, error) {
	category, err := svc.categoryRepository.FindByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if mode == "" {
		mode = inferMatchMode(category.Name)
	}
	if !mode.Valid() {
		return nil, domain.ErrInvalid
	}

	teams, err := svc.teamRepository.Find(ctx, domain.TeamQuery{CategoryID: categoryID})
	if err != nil {
		return nil, err
	}

	eligibleTeams, err := svc.eligibleTeams(ctx, teams)
	if err != nil {
		return nil, err
	}
	if len(eligibleTeams) == 0 {
		return []*domain.Match{}, nil
	}

	existingMatches, err := svc.matchRepository.Find(ctx, domain.MatchQuery{CategoryID: categoryID})
	if err != nil {
		return nil, err
	}

	var matches []*domain.Match
	switch mode {
	case domain.MatchModePairwise:
		matches, err = svc.startPairwiseMatches(ctx, categoryID, eligibleTeams, existingMatches)
	case domain.MatchModeShared:
		matches, err = svc.startSharedMatch(ctx, categoryID, eligibleTeams, existingMatches)
	}
	if err != nil {
		return nil, err
	}
	return matches, nil
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

func (svc *RobotService) eligibleTeams(ctx context.Context, teams []*domain.Team) ([]domain.Team, error) {
	eligible := make([]domain.Team, 0, len(teams))
	for _, team := range teams {
		robots, err := svc.robotRepository.Find(ctx, domain.RobotQuery{TeamID: team.ID})
		if err != nil {
			return nil, err
		}
		if firstRobotIsValid(robots) {
			eligible = append(eligible, *team)
		}
	}
	return eligible, nil
}

func (svc *RobotService) startPairwiseMatches(ctx context.Context, categoryID domain.CategoryID, teams []domain.Team, existingMatches []*domain.Match) ([]*domain.Match, error) {
	matches := make([]*domain.Match, 0, len(teams)/2)
	for i := 0; i+1 < len(teams); i += 2 {
		if existing := findExistingPairMatch(existingMatches, teams[i].ID, teams[i+1].ID); existing != nil {
			matches = append(matches, existing)
			continue
		}

		match, err := domain.NewPairMatch(teams[i], teams[i+1], categoryID)
		if err != nil {
			return nil, err
		}
		id, err := svc.matchRepository.Insert(ctx, match)
		if err != nil {
			return nil, err
		}
		match.ID = id
		matches = append(matches, match)
	}
	return matches, nil
}

func (svc *RobotService) startSharedMatch(ctx context.Context, categoryID domain.CategoryID, teams []domain.Team, existingMatches []*domain.Match) ([]*domain.Match, error) {
	teamIDs := make([]domain.TeamID, 0, len(teams))
	for _, team := range teams {
		teamIDs = append(teamIDs, team.ID)
	}
	if existing := findExistingSharedMatch(existingMatches, teamIDs); existing != nil {
		return []*domain.Match{existing}, nil
	}

	match, err := domain.NewQueueMatch(categoryID, teams)
	if err != nil {
		return nil, err
	}
	id, err := svc.matchRepository.Insert(ctx, match)
	if err != nil {
		return nil, err
	}
	match.ID = id
	return []*domain.Match{match}, nil
}

func findExistingPairMatch(matches []*domain.Match, teamAID, teamBID domain.TeamID) *domain.Match {
	for _, match := range matches {
		if match.TeamA == nil || match.TeamB == nil {
			continue
		}
		if match.TeamA.ID == teamAID && match.TeamB.ID == teamBID {
			return match
		}
		if match.TeamA.ID == teamBID && match.TeamB.ID == teamAID {
			return match
		}
	}
	return nil
}

func findExistingSharedMatch(matches []*domain.Match, teamIDs []domain.TeamID) *domain.Match {
	for _, match := range matches {
		if match.TeamA != nil || match.TeamB != nil {
			continue
		}
		if sameTeamIDs(match.Queue, teamIDs) {
			return match
		}
	}
	return nil
}

func sameTeamIDs(a, b []domain.TeamID) bool {
	if len(a) != len(b) {
		return false
	}

	counts := make(map[domain.TeamID]int, len(a))
	for _, id := range a {
		counts[id]++
	}
	for _, id := range b {
		counts[id]--
		if counts[id] < 0 {
			return false
		}
	}
	return true
}

func inferMatchMode(categoryName string) domain.MatchMode {
	if strings.Contains(strings.ToLower(categoryName), "sumo") {
		return domain.MatchModePairwise
	}
	return domain.MatchModeShared
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
	robots, err := svc.robotRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return svc.hydrateRobotsWithRules(ctx, robots)
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
	robot, err := svc.robotRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := svc.hydrateRobotWithRules(ctx, robot); err != nil {
		return nil, err
	}
	return robot, nil
}

func (svc *RobotService) GetRobotsByTeamID(ctx context.Context, teamID domain.TeamID) ([]*domain.Robot, error) {
	robots, err := svc.robotRepository.Find(ctx, domain.RobotQuery{TeamID: teamID})
	if err != nil {
		return nil, err
	}
	return svc.hydrateRobotsWithRules(ctx, robots)
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

func (svc *RobotService) hydrateRobotsWithRules(ctx context.Context, robots []*domain.Robot) ([]*domain.Robot, error) {
	for _, robot := range robots {
		if err := svc.hydrateRobotWithRules(ctx, robot); err != nil {
			return nil, err
		}
	}
	return robots, nil
}

func (svc *RobotService) hydrateRobotWithRules(ctx context.Context, robot *domain.Robot) error {
	robot.Rules = []*domain.Rule{}
	for _, ruleID := range robot.ValidRules {
		rule, err := svc.ruleRepository.FindByID(ctx, ruleID)
		if err != nil {
			return err
		}
		robot.Rules = append(robot.Rules, rule)
	}
	return nil
}

func (svc *RobotService) GetTeamsWithMembersAndCategory(ctx context.Context) ([]*domain.Team, error) {
	teams, err := svc.teamRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return svc.hydrateTeamsWithMembersAndCategory(ctx, teams)
}

func (svc *RobotService) GetTeamsWithMembersByCategory(ctx context.Context, categoryID domain.CategoryID) ([]*domain.Team, error) {
	teams, err := svc.teamRepository.Find(ctx, domain.TeamQuery{CategoryID: categoryID})
	if err != nil {
		return nil, err
	}

	return svc.hydrateTeamsWithMembersAndCategory(ctx, teams)
}

func (svc *RobotService) hydrateTeamsWithMembersAndCategory(ctx context.Context, teams []*domain.Team) ([]*domain.Team, error) {
	for _, team := range teams {
		members, err := svc.memberRepository.Find(ctx, domain.MemberQuery{TeamID: team.ID})
		if err != nil {
			return nil, err
		}
		team.Members = members

		robots, err := svc.GetRobotsByTeamID(ctx, team.ID)
		if err != nil {
			return nil, err
		}
		team.RobotValid = firstRobotIsValid(robots)

		category, err := svc.categoryRepository.FindByID(ctx, team.CategoryID)
		if err != nil {
			return nil, err
		}
		team.Category = category.Name
	}

	return teams, nil
}

func firstRobotIsValid(robots []*domain.Robot) bool {
	if len(robots) == 0 {
		return false
	}
	return robots[0].IsValid
}
