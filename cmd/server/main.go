package main

import (
	"context"
	"log"
	"net/http"

	"robot/internal/handler"
	"robot/internal/service"
	"robot/internal/store"
)

func main() {
	ctx := context.Background()

	storeDB, err := store.Open(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer storeDB.Close()

	categoryRepo := store.NewCategoryStore(storeDB)
	ruleRepo := store.NewRuleStore(storeDB)
	teamRepo := store.NewTeamStore(storeDB)
	memberRepo := store.NewMemberStore(storeDB)
	matchRepo := store.NewMatchStore(storeDB)
	resultRepo := store.NewResultStore(storeDB)
	robotRepo := store.NewRobotStore(storeDB)

	robotService := service.NewRobotService(categoryRepo, ruleRepo, teamRepo, memberRepo, matchRepo, resultRepo, robotRepo)
	robotHandler := handler.NewHandler(robotService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /registro", robotHandler.GetAllTeamsWithMembersAndCategory)

	mux.HandleFunc("GET /categorias", robotHandler.GetAllCategories)
	mux.HandleFunc("POST /categorias", robotHandler.CreateCategory)
	mux.HandleFunc("GET /categorias/{category_id}/equipos", robotHandler.GetTeamsByCategoryID)
	mux.HandleFunc("GET /categorias/{category_id}/reglas", robotHandler.GetRulesByCategoryID)
	mux.HandleFunc("POST /reglas", robotHandler.CreateRule)
	mux.HandleFunc("POST /equipos", robotHandler.CreateTeam)
	mux.HandleFunc("GET /equipos/{team_id}/miembros", robotHandler.GetMembersByTeamID)
	mux.HandleFunc("POST /miembros", robotHandler.CreateMember)
	mux.HandleFunc("GET /partidas", robotHandler.GetAllMatches)
	mux.HandleFunc("POST /partidas", robotHandler.CreateMatch)
	mux.HandleFunc("GET /partidas/{id}", robotHandler.GetMatchByID)
	mux.HandleFunc("GET /resultados", robotHandler.GetAllResults)
	mux.HandleFunc("POST /partidas/{match_id}/resultado", robotHandler.CreateResult)
	mux.HandleFunc("GET /partidas/{match_id}/resultado", robotHandler.GetResultByMatchID)
	mux.HandleFunc("GET /robots", robotHandler.GetAllRobots)
	mux.HandleFunc("POST /robots", robotHandler.CreateRobot)
	mux.HandleFunc("GET /robots/{id}", robotHandler.GetRobotByID)
	mux.HandleFunc("POST /robots/{id}/verificar", robotHandler.VerifyRobot)
	mux.HandleFunc("GET /equipos/{team_id}/robots", robotHandler.GetRobotsByTeamID)
	log.Println("server listening on :8080")
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		log.Fatal(err)
	}
}
