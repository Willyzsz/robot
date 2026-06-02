package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"robot/internal/handler"
	"robot/internal/service"
	"robot/internal/store"
)

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *statusResponseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		log.Printf("%s %s -> %d %s", r.Method, r.URL.Path, rw.status, time.Since(start))
	})
}

func main() {
	ctx := context.Background()

	storeDB, err := store.Open(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer storeDB.Close()

	if err := storeDB.Migrate(ctx); err != nil {
		log.Fatal(err)
	}
	// if err := storeDB.Seed(ctx); err != nil {
	// 	log.Fatal(err)
	// }

	userRepo := store.NewUserStore(storeDB)
	categoryRepo := store.NewCategoryStore(storeDB)
	ruleRepo := store.NewRuleStore(storeDB)
	teamRepo := store.NewTeamStore(storeDB)
	memberRepo := store.NewMemberStore(storeDB)
	matchRepo := store.NewMatchStore(storeDB)
	resultRepo := store.NewResultStore(storeDB)
	robotRepo := store.NewRobotStore(storeDB)

	robotService := service.NewRobotService(userRepo, categoryRepo, ruleRepo, teamRepo, memberRepo, matchRepo, resultRepo, robotRepo)
	robotHandler := handler.NewHandler(robotService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", robotHandler.Login)
	mux.HandleFunc("GET /registro", robotHandler.GetAllTeamsWithMembersAndCategory)
	mux.HandleFunc("GET /registro/{category_id}/equipos", robotHandler.GetTeamsWithMembersByCategory)

	mux.HandleFunc("GET /categorias", robotHandler.GetAllCategories)
	mux.HandleFunc("POST /categorias", robotHandler.CreateCategory)
	mux.HandleFunc("GET /categorias/{category_id}/equipos", robotHandler.GetTeamsByCategoryID)
	mux.HandleFunc("GET /categorias/{category_id}/reglas", robotHandler.GetRulesByCategoryID)
	mux.HandleFunc("GET /categorias/{category_id}/bracket", robotHandler.GetCategoryBracket)
	mux.HandleFunc("POST /categorias/{category_id}/partidas/iniciar", robotHandler.StartMatchQueue)
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
	if err := http.ListenAndServe("0.0.0.0:8080", logRequests(mux)); err != nil {
		log.Fatal(err)
	}
}
