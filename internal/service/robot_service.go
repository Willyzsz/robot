package service

import (
	"context"
	"fmt"
	"math/rand"
	"robot/internal/domain"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type RobotService struct {
	userRepository     domain.UserRepository
	categoryRepository domain.CategoryRepository
	ruleRepository     domain.RuleRepository
	teamRepository     domain.TeamRepository
	memberRepository   domain.MemberRepository
	matchRepository    domain.MatchRepository
	resultRepository   domain.ResultRepository
	robotRepository    domain.RobotRepository
}

func NewRobotService(userRepo domain.UserRepository, categoryRepo domain.CategoryRepository, ruleRepo domain.RuleRepository, teamRepo domain.TeamRepository, memberRepo domain.MemberRepository, matchRepo domain.MatchRepository, resultRepo domain.ResultRepository, robotRepo domain.RobotRepository) *RobotService {
	return &RobotService{
		userRepository:     userRepo,
		categoryRepository: categoryRepo,
		ruleRepository:     ruleRepo,
		teamRepository:     teamRepo,
		memberRepository:   memberRepo,
		matchRepository:    matchRepo,
		resultRepository:   resultRepo,
		robotRepository:    robotRepo,
	}
}

func (svc *RobotService) Login(ctx context.Context, username, password string) (*domain.User, error) {
	user, err := svc.userRepository.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalid
	}
	return user, nil
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

func (svc *RobotService) CreateTeam(ctx context.Context, name, school, grade, teacher string, isInternal bool, categoryID domain.CategoryID) (domain.TeamID, error) {
	team, err := domain.NewTeam(name, school, grade, teacher, categoryID)
	if err != nil {
		return 0, err
	}
	team.IsInternal = isInternal
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

func (svc *RobotService) GetCategoryBracket(ctx context.Context, categoryID domain.CategoryID) (*domain.Bracket, error) {
	if _, err := svc.categoryRepository.FindByID(ctx, categoryID); err != nil {
		return nil, err
	}

	teams, err := svc.teamRepository.Find(ctx, domain.TeamQuery{CategoryID: categoryID})
	if err != nil {
		return nil, err
	}
	teamIDs := make([]domain.TeamID, 0, len(teams))
	for _, team := range teams {
		teamIDs = append(teamIDs, team.ID)
	}

	matches, err := svc.matchRepository.Find(ctx, domain.MatchQuery{CategoryID: categoryID})
	if err != nil {
		return nil, err
	}

	size := bracketSize(len(teamIDs))
	labels := bracketLabels(size)
	rounds := make([][]domain.BracketMatch, len(labels))
	rounds[0] = firstBracketRound(matches)
	for round := 1; round < len(labels); round++ {
		rounds[round] = nextBracketRound(round, rounds[round-1])
	}

	return &domain.Bracket{
		ID:         fmt.Sprintf("br-%d", categoryID),
		CategoryID: categoryID,
		Size:       size,
		TeamIDs:    teamIDs,
		Labels:     labels,
		Rounds:     rounds,
	}, nil
}

func (svc *RobotService) StartMatchQueue(ctx context.Context, categoryID domain.CategoryID, mode domain.MatchMode) ([]*domain.Match, error) {
	category, err := svc.categoryRepository.FindByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	mode = inferMatchMode(category.Name)
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
	if isSumoCategory(category.Name) {
		eligibleTeams = availableSumoTeamsForQueue(eligibleTeams, existingMatches)
	} else {
		eligibleTeams = availableTeamsForQueue(eligibleTeams, existingMatches)
	}
	if len(eligibleTeams) == 0 {
		return []*domain.Match{}, nil
	}
	shuffleTeams(eligibleTeams)

	return svc.startCategoryMatches(ctx, category, eligibleTeams, existingMatches, mode)
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

func (svc *RobotService) startCategoryMatches(ctx context.Context, category *domain.Category, teams []domain.Team, existingMatches []*domain.Match, mode domain.MatchMode) ([]*domain.Match, error) {
	groups := teamsByInternalStatus(teams)
	matches := []*domain.Match{}

	for _, isInternal := range []bool{true, false} {
		group := groups[isInternal]
		if len(group) == 0 {
			continue
		}

		var groupMatches []*domain.Match
		var err error
		switch {
		case isSumoCategory(category.Name):
			groupMatches, err = svc.startSumoMatches(ctx, category.ID, isInternal, group, existingMatches)
		case isFutbolCategory(category.Name):
			groupMatches, err = svc.startOnePairWithQueue(ctx, category.ID, "main", isInternal, group, existingMatches)
		case isVelocistaCategory(category.Name):
			groupMatches, err = svc.startQueueMatch(ctx, category.ID, "queue", isInternal, group, existingMatches)
		case mode == domain.MatchModePairwise:
			groupMatches, err = svc.startOnePairWithQueue(ctx, category.ID, "main", isInternal, group, existingMatches)
		default:
			groupMatches, err = svc.startQueueMatch(ctx, category.ID, "queue", isInternal, group, existingMatches)
		}
		if err != nil {
			return nil, err
		}
		matches = append(matches, groupMatches...)
	}

	return matches, nil
}

func (svc *RobotService) startOnePairWithQueue(ctx context.Context, categoryID domain.CategoryID, bracketID string, isInternal bool, teams []domain.Team, existingMatches []*domain.Match) ([]*domain.Match, error) {
	bracketID = groupedBracketID(bracketID, isInternal)
	if active := activeMatchForGroup(existingMatches, isInternal, bracketID); active != nil {
		return []*domain.Match{active}, nil
	}
	if len(teams) < 2 {
		return []*domain.Match{}, nil
	}

	round := nextBracketRoundIndex(bracketID, existingMatches)
	match, err := domain.NewPairMatch(teams[0], teams[1], categoryID)
	if err != nil {
		return nil, err
	}
	for _, team := range teams[2:] {
		if err := match.AddToQueue(team); err != nil {
			return nil, err
		}
	}
	match.IsInternal = isInternal
	match.BracketID = bracketID
	match.BracketRound = round
	match.BracketSlot = 0
	match.BracketKey = fmt.Sprintf("r%d-m%d", match.BracketRound, match.BracketSlot)
	match.Status = domain.MatchStatusReady
	id, err := svc.matchRepository.Insert(ctx, match)
	if err != nil {
		return nil, err
	}
	match.ID = id
	return []*domain.Match{match}, nil
}

func (svc *RobotService) startSumoMatches(ctx context.Context, categoryID domain.CategoryID, isInternal bool, teams []domain.Team, existingMatches []*domain.Match) ([]*domain.Match, error) {
	losses := sumoLosses(existingMatches)
	winnerTeams := make([]domain.Team, 0, len(teams))
	loserTeams := make([]domain.Team, 0, len(teams))

	for _, team := range teams {
		switch losses[team.ID] {
		case 0:
			winnerTeams = append(winnerTeams, team)
		case 1:
			loserTeams = append(loserTeams, team)
		}
	}

	var matches []*domain.Match
	winnerMatches, err := svc.startOnePairWithQueue(ctx, categoryID, "winner", isInternal, winnerTeams, existingMatches)
	if err != nil {
		return nil, err
	}
	matches = append(matches, winnerMatches...)

	loserMatches, err := svc.startOnePairWithQueue(ctx, categoryID, "loser", isInternal, loserTeams, existingMatches)
	if err != nil {
		return nil, err
	}
	matches = append(matches, loserMatches...)
	return matches, nil
}

func (svc *RobotService) startQueueMatch(ctx context.Context, categoryID domain.CategoryID, bracketID string, isInternal bool, teams []domain.Team, existingMatches []*domain.Match) ([]*domain.Match, error) {
	bracketID = groupedBracketID(bracketID, isInternal)
	if active := activeMatchForGroup(existingMatches, isInternal, bracketID); active != nil {
		return []*domain.Match{active}, nil
	}

	match, err := domain.NewQueueMatch(categoryID, teams)
	if err != nil {
		return nil, err
	}
	match.IsInternal = isInternal
	match.BracketID = bracketID
	match.BracketRound = nextBracketRoundIndex(bracketID, existingMatches)
	match.BracketSlot = 0
	match.BracketKey = fmt.Sprintf("r%d-m%d", match.BracketRound, match.BracketSlot)
	match.Status = domain.MatchStatusReady
	id, err := svc.matchRepository.Insert(ctx, match)
	if err != nil {
		return nil, err
	}
	match.ID = id
	return []*domain.Match{match}, nil
}

func availableTeamsForQueue(teams []domain.Team, existingMatches []*domain.Match) []domain.Team {
	blocked := make(map[domain.TeamID]struct{})
	for _, match := range existingMatches {
		if match.Result != nil || match.Status == domain.MatchStatusCompleted {
			if match.Result.EliminatedTeamID != nil {
				blocked[*match.Result.EliminatedTeamID] = struct{}{}
			}
			continue
		}

		for _, teamID := range matchTeamIDs(match) {
			blocked[teamID] = struct{}{}
		}
	}

	available := make([]domain.Team, 0, len(teams))
	for _, team := range teams {
		if _, exists := blocked[team.ID]; exists {
			continue
		}
		available = append(available, team)
	}
	return available
}

func availableSumoTeamsForQueue(teams []domain.Team, existingMatches []*domain.Match) []domain.Team {
	blocked := activeMatchTeams(existingMatches)
	losses := sumoLosses(existingMatches)

	available := make([]domain.Team, 0, len(teams))
	for _, team := range teams {
		if _, exists := blocked[team.ID]; exists {
			continue
		}
		if losses[team.ID] >= 2 {
			continue
		}
		available = append(available, team)
	}
	return available
}

func teamsByInternalStatus(teams []domain.Team) map[bool][]domain.Team {
	groups := map[bool][]domain.Team{
		true:  {},
		false: {},
	}
	for _, team := range teams {
		groups[team.IsInternal] = append(groups[team.IsInternal], team)
	}
	return groups
}

func activeMatchForGroup(matches []*domain.Match, isInternal bool, bracketID string) *domain.Match {
	for _, match := range matches {
		if match.IsInternal != isInternal || match.BracketID != bracketID {
			continue
		}
		if match.Result != nil || match.Status == domain.MatchStatusCompleted {
			continue
		}
		return match
	}
	return nil
}

func groupedBracketID(bracketID string, isInternal bool) string {
	if isInternal {
		return bracketID + "-internal"
	}
	return bracketID + "-external"
}

func activeMatchTeams(matches []*domain.Match) map[domain.TeamID]struct{} {
	blocked := make(map[domain.TeamID]struct{})
	for _, match := range matches {
		if match.Result != nil || match.Status == domain.MatchStatusCompleted {
			continue
		}
		for _, teamID := range matchTeamIDs(match) {
			blocked[teamID] = struct{}{}
		}
	}
	return blocked
}

func sumoLosses(matches []*domain.Match) map[domain.TeamID]int {
	losses := make(map[domain.TeamID]int)
	for _, match := range matches {
		if match.Result == nil || match.Result.EliminatedTeamID == nil {
			continue
		}
		losses[*match.Result.EliminatedTeamID]++
	}
	return losses
}

func nextBracketRoundIndex(bracketID string, matches []*domain.Match) int {
	next := 0
	for _, match := range matches {
		if match.BracketID != bracketID {
			continue
		}
		if match.BracketRound >= next {
			next = match.BracketRound + 1
		}
	}
	return next
}

func matchTeamIDs(match *domain.Match) []domain.TeamID {
	var teamIDs []domain.TeamID
	if match.TeamA != nil {
		teamIDs = append(teamIDs, match.TeamA.ID)
	}
	if match.TeamB != nil {
		teamIDs = append(teamIDs, match.TeamB.ID)
	}
	teamIDs = append(teamIDs, match.Queue...)
	return teamIDs
}

func shuffleTeams(teams []domain.Team) {
	rand.Shuffle(len(teams), func(i, j int) {
		teams[i], teams[j] = teams[j], teams[i]
	})
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
	name := normalizeCategoryName(categoryName)
	if strings.Contains(name, "sumo") || strings.Contains(name, "futbol") {
		return domain.MatchModePairwise
	}
	return domain.MatchModeShared
}

func isSumoCategory(categoryName string) bool {
	return strings.Contains(normalizeCategoryName(categoryName), "sumo")
}

func isFutbolCategory(categoryName string) bool {
	return strings.Contains(normalizeCategoryName(categoryName), "futbol")
}

func isVelocistaCategory(categoryName string) bool {
	return strings.Contains(normalizeCategoryName(categoryName), "velocista")
}

func normalizeCategoryName(name string) string {
	name = strings.ToLower(name)
	replacer := strings.NewReplacer(
		"á", "a",
		"é", "e",
		"í", "i",
		"ó", "o",
		"ú", "u",
	)
	return replacer.Replace(name)
}

func bracketSize(teamCount int) int {
	size := 2
	for size < teamCount {
		size *= 2
	}
	return size
}

func bracketLabels(size int) []string {
	roundCount := 0
	for n := size; n > 1; n /= 2 {
		roundCount++
	}

	labels := make([]string, roundCount)
	for i := range labels {
		remaining := roundCount - i
		switch remaining {
		case 1:
			labels[i] = "Final"
		case 2:
			labels[i] = "Semifinal"
		case 3:
			labels[i] = "Quarterfinal"
		default:
			labels[i] = fmt.Sprintf("Round %d", i+1)
		}
	}
	return labels
}

func firstBracketRound(matches []*domain.Match) []domain.BracketMatch {
	round := make([]domain.BracketMatch, 0, len(matches))
	for slot, match := range matches {
		round = append(round, bracketMatchFromMatch(0, slot, match))
	}
	return round
}

func bracketMatchFromMatch(round, slot int, match *domain.Match) domain.BracketMatch {
	var teamA, teamB, winner *domain.TeamID
	var matchID *domain.MatchID
	if match.TeamA != nil {
		id := match.TeamA.ID
		teamA = &id
	}
	if match.TeamB != nil {
		id := match.TeamB.ID
		teamB = &id
	}
	if match.Result != nil {
		id := match.Result.Winner
		winner = &id
	}
	if match.ID != 0 {
		id := match.ID
		matchID = &id
	}

	return domain.BracketMatch{
		Key:     fmt.Sprintf("r%d-m%d", round, slot),
		Round:   round,
		Slot:    slot,
		TeamA:   teamA,
		TeamB:   teamB,
		Winner:  winner,
		MatchID: matchID,
		Status:  bracketMatchStatus(teamA, teamB, winner),
	}
}

func nextBracketRound(round int, previous []domain.BracketMatch) []domain.BracketMatch {
	slots := len(previous) / 2
	if slots < 1 {
		slots = 1
	}

	next := make([]domain.BracketMatch, 0, slots)
	for slot := range slots {
		var teamA, teamB *domain.TeamID
		left := slot * 2
		right := left + 1
		if left < len(previous) {
			teamA = previous[left].Winner
		}
		if right < len(previous) {
			teamB = previous[right].Winner
		}

		next = append(next, domain.BracketMatch{
			Key:    fmt.Sprintf("r%d-m%d", round, slot),
			Round:  round,
			Slot:   slot,
			TeamA:  teamA,
			TeamB:  teamB,
			Status: bracketMatchStatus(teamA, teamB, nil),
		})
	}
	return next
}

func bracketMatchStatus(teamA, teamB, winner *domain.TeamID) string {
	if winner != nil {
		return "completed"
	}
	if teamA != nil && teamB != nil {
		return "ready"
	}
	return "pending"
}

func (svc *RobotService) GetMatchByID(ctx context.Context, id domain.MatchID) (*domain.Match, error) {
	return svc.matchRepository.FindByID(ctx, id)
}

func (svc *RobotService) GetAllResults(ctx context.Context) ([]*domain.Result, error) {
	return svc.resultRepository.FindAll(ctx)
}

func (svc *RobotService) CreateResult(ctx context.Context, matchID domain.MatchID, winner domain.TeamID, eliminatedTeamID *domain.TeamID, resultTime *domain.ResultTime) (domain.ResultID, error) {
	match, err := svc.matchRepository.FindByID(ctx, matchID)
	if err != nil {
		return 0, err
	}

	result, err := domain.NewResultForMatch(match, winner, eliminatedTeamID, resultTime)
	if err != nil {
		return 0, err
	}

	id, err := svc.resultRepository.Insert(ctx, result)
	if err != nil {
		return 0, err
	}
	if err := svc.matchRepository.SetStatus(ctx, matchID, domain.MatchStatusCompleted); err != nil {
		return 0, err
	}
	return id, nil
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
