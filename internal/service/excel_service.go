package service

import (
	"context"
	"errors"
	"robot/internal/domain"
	"robot/internal/excel"
	"robot/pkg/apperr"
	"strings"
)

const pendingExcelValue = "pendiente"

type ExcelService struct {
	categoryRepository domain.CategoryRepository
	ruleRepository     domain.RuleRepository
	teamRepository     domain.TeamRepository
	memberRepository   domain.MemberRepository
}

func NewExcelService(categoryRepo domain.CategoryRepository, ruleRepo domain.RuleRepository, teamRepo domain.TeamRepository, memberRepo domain.MemberRepository) *ExcelService {
	return &ExcelService{
		categoryRepository: categoryRepo,
		ruleRepository:     ruleRepo,
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

func (svc *ExcelService) CreateRules(ctx context.Context, rows []excel.RuleRow) error {
	op := "CreateRules"
	categoriesIDs := make(map[string]domain.CategoryID)
	existingRules := make(map[domain.CategoryID]map[string]struct{})

	for _, row := range rows {
		description := ruleDescription(row)
		if strings.TrimSpace(row.Category) == "" || description == "" {
			continue
		}

		categoryID, err := svc.getOrCreateCategory(ctx, row.Category, categoriesIDs)
		if err != nil {
			return svc.err(op, err)
		}

		ruleType := domain.RuleTypeCharacteristic
		if row.Type == excel.RuleSheetTypeRestriction {
			ruleType = domain.RuleTypeRestriction
		}

		key := ruleKey(description, ruleType)
		rules, ok, err := svc.rulesForCategory(ctx, categoryID, existingRules)
		if err != nil {
			return svc.err(op, err)
		}
		if ok {
			if _, exists := rules[key]; exists {
				continue
			}
		}

		rule, err := domain.NewRule(description, ruleType, categoryID)
		if err != nil {
			return svc.err(op, err)
		}
		if _, err := svc.ruleRepository.Insert(ctx, rule); err != nil {
			return svc.err(op, err)
		}
		rules[key] = struct{}{}
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

func ruleDescription(row excel.RuleRow) string {
	switch row.Type {
	case excel.RuleSheetTypeRegistration:
		characteristic := strings.TrimSpace(row.Characteristic)
		specification := strings.TrimSpace(row.Specification)
		switch {
		case characteristic == "":
			return specification
		case specification == "":
			return characteristic
		default:
			return characteristic + ": " + specification
		}
	case excel.RuleSheetTypeRestriction:
		return strings.TrimSpace(row.Restriction)
	default:
		return ""
	}
}

func ruleKey(description string, ruleType domain.RuleType) string {
	return string(ruleType) + ":" + strings.ToLower(strings.Join(strings.Fields(description), " "))
}

func (svc *ExcelService) rulesForCategory(ctx context.Context, categoryID domain.CategoryID, cache map[domain.CategoryID]map[string]struct{}) (map[string]struct{}, bool, error) {
	if rules, exists := cache[categoryID]; exists {
		return rules, true, nil
	}

	foundRules, err := svc.ruleRepository.FindByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, false, err
	}

	rules := make(map[string]struct{}, len(foundRules))
	for _, rule := range foundRules {
		rules[ruleKey(rule.Description, rule.Type)] = struct{}{}
	}
	cache[categoryID] = rules
	return rules, true, nil
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
	return apperr.Wrap(op, "", err)
}
