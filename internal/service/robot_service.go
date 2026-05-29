package service

import (
	"context"
	"robot/internal/domain"
	"time"
)

type RobotService struct {
	categoryRepository domain.CategoryRepository
	teamRepository     domain.TeamRepository
	memberRepository   domain.MemberRepository
	matchRepository    domain.MatchRepository
	resultRepository   domain.ResultRepository
}

func NewRobotService(categoryRepo domain.CategoryRepository, teamRepo domain.TeamRepository, memberRepo domain.MemberRepository, matchRepo domain.MatchRepository, resultRepo domain.ResultRepository) *RobotService {
	return &RobotService{
		categoryRepository: categoryRepo,
		teamRepository:     teamRepo,
		memberRepository:   memberRepo,
		matchRepository:    matchRepo,
		resultRepository:   resultRepo,
	}
}

func (svc *RobotService) GetAllCategories(ctx context.Context) ([]*domain.Category, error) {
	return svc.categoryRepository.FindAll(ctx)
}

func (svc *RobotService) GetAllTeams(ctx context.Context) ([]*domain.Team, error) {
	return svc.teamRepository.FindAll(ctx)
}

func (svc *RobotService) GetTeamById(ctx context.Context, id domain.TeamID) (*domain.Team, error) {
	return svc.teamRepository.FindByID(ctx, id)
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
