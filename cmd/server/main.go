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

	robotService := service.NewRobotService(categoryRepo, teamRepo, memberRepo)
	robotHandler := handler.NewHandler(robotService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /registro", robotHandler.GetAllTeamsWithMembersAndCategory)

	mux.HandleFunc("GET /categorias", robotHandler.GetAllCategories)
	log.Println("server listening on :8080")
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		log.Fatal(err)
	}
}