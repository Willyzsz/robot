package service

import (
	"context"
	"robot/internal/domain"
)

type RobotService struct {
	categoryRepository domain.CategoryRepository
	teamRepository     domain.TeamRepository
	memberRepository   domain.MemberRepository
}

func NewRobotService(categoryRepo domain.CategoryRepository, teamRepo domain.TeamRepository, memberRepo domain.MemberRepository) *RobotService {
	return &RobotService{
		categoryRepository: categoryRepo,
		teamRepository:     teamRepo,
		memberRepository:   memberRepo,
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
