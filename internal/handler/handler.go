package handler

import (
	"encoding/json"
	"net/http"
	"robot/internal/service"
)

type Handler struct {
	robotService *service.RobotService
}

func NewHandler(robotService *service.RobotService) *Handler {
	return &Handler{
		robotService: robotService,
	}
}

func (h *Handler) GetAllTeams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
  	}
	
	ctx := r.Context()

	teams, err := h.robotService.GetAllTeams(ctx)
	if err != nil {
		http.Error(w, "Failed to retrieve teams", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(teams); err != nil {
		return
	}
}