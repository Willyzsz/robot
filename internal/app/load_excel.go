package app

import (
	"context"
	"robot/internal/excel"
	"robot/internal/service"
	"robot/internal/store"
)

func LoadFromExcel(path string) error {
	parsedRows, err := excel.ParseRows(path)
	if err != nil {
		return err
	}

	ctx := context.Background()
	storeDB, err := store.Open(ctx)
	if err != nil {
		return err
	}
	defer storeDB.Close()

	if err := storeDB.Migrate(ctx); err != nil {
		return err
	}

	categoryRepo := store.NewCategoryStore(storeDB)
	teamRepo := store.NewTeamStore(storeDB)
	memberRepo := store.NewMemberStore(storeDB)

	service := service.NewRobotService(categoryRepo, teamRepo,memberRepo)

	if err := service.CreateData(ctx, parsedRows); err != nil {
		return err
	}
	return nil
}