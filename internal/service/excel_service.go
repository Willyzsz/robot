package service

import (
	"context"
	"errors"
	"robot/internal/domain"
	"robot/internal/excel"
	"strings"
)

const pendingExcelValue = "pendiente"

type ExcelService struct {
	categoryRepository domain.CategoryRepository
	teamRepository     domain.TeamRepository
	memberRepository   domain.MemberRepository
}

func NewExcelService(categoryRepo domain.CategoryRepository, teamRepo domain.TeamRepository, memberRepo domain.MemberRepository) *ExcelService {
	return &ExcelService{
		categoryRepository: categoryRepo,
		teamRepository:     teamRepo,
		memberRepository:   memberRepo,
	}
}

func (svc *ExcelService) CreateData(ctx context.Context, rows []excel.FormRow) error {
	op := "CreateData"
	categoriesIDs := make(map[string]domain.CategoryID)
	teamsIDs := make(map[string]domain.TeamID)

	for _, row := range rows {
		if shouldSkipExcelRow(row) {
			continue
		}

		categoryID, err := svc.getOrCreateCategory(ctx, row.Category, categoriesIDs)
		if err != nil {
			return svc.err(op, err)
		}

		if _, ok := teamsIDs[row.NameTeam]; ok {
			continue
		}

		team, err := svc.buildTeamFromRow(row, categoryID)
		if err != nil {
			return svc.err(op, err)
		}

		teamID, created, err := svc.getOrCreateTeam(ctx, team, teamsIDs)
		if err != nil {
			return svc.err(op, err)
		}
		if !created {
			continue
		}

		if err := svc.insertMembers(ctx, team, teamID); err != nil {
			return svc.err(op, err)
		}
		teamsIDs[row.NameTeam] = teamID
	}
	return nil
}

func (svc *ExcelService) getOrCreateCategory(ctx context.Context, name string, cache map[string]domain.CategoryID) (domain.CategoryID, error) {
	if id, exists := cache[name]; exists {
		return id, nil
	}

	found, err := svc.categoryRepository.FindByName(ctx, name)
	if err == nil {
		cache[name] = found.ID
		return found.ID, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return 0, err
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

func (svc *ExcelService) buildTeamFromRow(row excel.FormRow, categoryID domain.CategoryID) (*domain.Team, error) {
	teacher := valueOrPending(row.Teacher)
	leaderName := valueOrPending(row.NameLeader)

	team, err := domain.NewTeam(row.NameTeam, row.School, row.Grade, teacher, categoryID)
	if err != nil {
		return nil, err
	}

	leader, err := domain.NewMember(leaderName, row.EmailLeader, true, team.ID)
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
			if errors.Is(err, domain.ErrAlreadyExists) {
				continue
			}
			return nil, err
		}
	}

	if err := team.ValidateMembers(); err != nil {
		return nil, err
	}

	return team, nil
}

func valueOrPending(value string) string {
	if strings.TrimSpace(value) == "" {
		return pendingExcelValue
	}
	return value
}

func shouldSkipExcelRow(row excel.FormRow) bool {
	return strings.TrimSpace(row.NameTeam) == "" ||
		strings.TrimSpace(row.Category) == "" ||
		strings.TrimSpace(row.School) == "" ||
		strings.TrimSpace(row.Grade) == ""
}

func (svc *ExcelService) getOrCreateTeam(ctx context.Context, team *domain.Team, cache map[string]domain.TeamID) (domain.TeamID, bool, error) {
	if id, exists := cache[team.Name]; exists {
		return id, false, nil
	}

	teamFound, err := svc.teamRepository.FindByName(ctx, team.Name)
	if err == nil {
		cache[team.Name] = teamFound.ID
		return teamFound.ID, false, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return 0, false, err
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

func (svc *ExcelService) insertMembers(ctx context.Context, team *domain.Team, teamID domain.TeamID) error {
	for _, member := range team.Members {
		member.TeamID = teamID
		if _, err := svc.memberRepository.Insert(ctx, member, teamID); err != nil {
			return err
		}
	}
	return nil
}

func (svc *ExcelService) err(op string, err error) error {
	if err == nil {
		return nil
	}
	var robErr *domain.RobotError
	if errors.As(err, &robErr) {
		return domain.NewRobotErr(op, robErr.Op, robErr.Field, robErr.Value, robErr.Err, robErr.Msg)
	}
	return err
}
