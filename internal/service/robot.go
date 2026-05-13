package service

import (
	"context"
	"errors"
	"robot/internal/domain"
	"robot/internal/excel"
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

func (svc *RobotService) CreateData(ctx context.Context, rows []excel.FormRow) error {
	categoriesIDs := make(map[string]domain.CategoryID)
	teamsIDs := make(map[string]domain.TeamID)

	for _, row := range rows {
		categoryID, err := svc.getOrCreateCategory(ctx, row.Category, categoriesIDs)
		if err != nil {
			return err
		}

		if _, ok := teamsIDs[row.NameTeam]; ok {
			continue
		}

		team, err := svc.buildTeamFromRow(row, categoryID)
		if err != nil {
			return err
		}

		teamID, created, err := svc.getOrCreateTeam(ctx, team, teamsIDs)
		if err != nil {
			return err
		}
		if !created {
			continue
		}

		if err := svc.insertMembers(ctx, team, teamID); err != nil {
			return err
		}
		teamsIDs[row.NameTeam] = teamID
	}
	return nil
}

func (svc *RobotService) getOrCreateCategory(ctx context.Context, name string, cache map[string]domain.CategoryID) (domain.CategoryID, error) {
	if id, exists := cache[name]; exists {
		return id, nil
	}

	category, err := domain.NewCategory(name)
	if err != nil {
		return 0, err
	}

	id, err := svc.categoryRepository.Insert(ctx, category)
	if errors.Is(err, domain.ErrAlreadyExists) {
		found, findErr := svc.categoryRepository.FindByName(ctx, name)
		if findErr != nil {
			return 0, findErr
		}
		id = found.ID
	} else if err != nil {
		return 0, err
	}

	cache[name] = id
	return id, nil
}

func (svc *RobotService) buildTeamFromRow(row excel.FormRow, categoryID domain.CategoryID) (*domain.Team, error) {
	team, err := domain.NewTeam(row.NameTeam, row.School, row.Grade, row.Teacher, categoryID)
	if err != nil {
		return nil, err
	}

	leader, err := domain.NewMember(row.NameLeader, row.EmailLeader, true, team.ID)
	if err != nil {
		return nil, err
	}
	if err := team.AddMember(leader); err != nil {
		return nil, err
	}

	for i := 0; i < len(row.Members); i++ {
		email := ""
		if i < len(row.EmailMembers) {
			email = row.EmailMembers[i]
		}

		member, err := domain.NewMember(row.Members[i], email, false, team.ID)
		if err != nil {
			return nil, err
		}
		if err := team.AddMember(member); err != nil {
			return nil, err
		}
	}

	if err := team.ValidateMembers(); err != nil {
		return nil, err
	}

	return team, nil
}

func (svc *RobotService) getOrCreateTeam(ctx context.Context, team *domain.Team, cache map[string]domain.TeamID) (domain.TeamID, bool, error) {
	if id, exists := cache[team.Name]; exists {
		return id, false, nil
	}

	id, err := svc.teamRepository.Insert(ctx, team)
	if errors.Is(err, domain.ErrAlreadyExists) {
		teamFound, findError := svc.teamRepository.FindByName(ctx, team.Name)
		if findError != nil {
			return 0, false, findError
		}

		id = teamFound.ID
		cache[team.Name] = id

		return id, false, nil
	} else if err != nil {
		return 0, false, err
	}

	cache[team.Name] = id
	return id, true, nil
}

func (svc *RobotService) insertMembers(ctx context.Context, team *domain.Team, teamID domain.TeamID) error {
	for _, member := range team.Members {
		member.TeamID = teamID
		if _, err := svc.memberRepository.Insert(ctx, member, teamID); err != nil {
			return err
		}
	}
	return nil
}
