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