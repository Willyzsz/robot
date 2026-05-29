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
	teamRepo := store.NewTeamStore(storeDB)
	memberRepo := store.NewMemberStore(storeDB)
	matchRepo := store.NewMatchStore(storeDB)
	resultRepo := store.NewResultStore(storeDB)

	robotService := service.NewRobotService(categoryRepo, teamRepo, memberRepo, matchRepo, resultRepo)
	robotHandler := handler.NewHandler(robotService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /registro", robotHandler.GetAllTeamsWithMembersAndCategory)

	mux.HandleFunc("GET /categorias", robotHandler.GetAllCategories)
	mux.HandleFunc("GET /partidas", robotHandler.GetAllMatches)
	mux.HandleFunc("POST /partidas", robotHandler.CreateMatch)
	mux.HandleFunc("GET /partidas/{id}", robotHandler.GetMatchByID)
	mux.HandleFunc("GET /resultados", robotHandler.GetAllResults)
	mux.HandleFunc("POST /partidas/{match_id}/resultado", robotHandler.CreateResult)
	mux.HandleFunc("GET /partidas/{match_id}/resultado", robotHandler.GetResultByMatchID)
	log.Println("server listening on :8080")
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		log.Fatal(err)
	}
}
