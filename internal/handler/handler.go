package handler

import (
	"encoding/json"
	"net/http"
	"robot/internal/domain"
	"robot/internal/service"
	"strconv"
	"time"
)

type Handler struct {
	robotService *service.RobotService
}

type createMatchRequest struct {
	TeamAID domain.TeamID   `json:"team_a_id"`
	TeamBID domain.TeamID   `json:"team_b_id"`
	Queue   []domain.TeamID `json:"queue"`
}

type createResultRequest struct {
	Winner domain.TeamID `json:"winner"`
	Time   *time.Time    `json:"time,omitempty"`
}

type createResponse struct {
	ID any `json:"id"`
}

func NewHandler(robotService *service.RobotService) *Handler {
	return &Handler{
		robotService: robotService,
	}
}

func (h *Handler) GetAllTeamsWithMembersAndCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	teams, err := h.robotService.GetTeamsWithMembersAndCategory(ctx)
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

func (h *Handler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	categories, err := h.robotService.GetAllCategories(ctx)
	if err != nil {
		http.Error(w, "Failed to retrieve categories", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(categories); err != nil {
		return
	}
}

func (h *Handler) GetAllMatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	matches, err := h.robotService.GetAllMatches(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve matches", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, matches)
}

func (h *Handler) CreateMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid match payload", http.StatusBadRequest)
		return
	}

	id, err := h.robotService.CreateMatch(r.Context(), req.TeamAID, req.TeamBID, req.Queue)
	if err != nil {
		http.Error(w, "Failed to create match", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, createResponse{ID: id})
}

func (h *Handler) GetMatchByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid match id", http.StatusBadRequest)
		return
	}

	match, err := h.robotService.GetMatchByID(r.Context(), domain.MatchID(id))
	if err != nil {
		http.Error(w, "Failed to retrieve match", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, match)
}

func (h *Handler) GetAllResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	results, err := h.robotService.GetAllResults(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve results", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *Handler) CreateResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	matchID, err := strconv.Atoi(r.PathValue("match_id"))
	if err != nil || matchID <= 0 {
		http.Error(w, "invalid match id", http.StatusBadRequest)
		return
	}

	var req createResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid result payload", http.StatusBadRequest)
		return
	}

	id, err := h.robotService.CreateResult(r.Context(), domain.MatchID(matchID), req.Winner, req.Time)
	if err != nil {
		http.Error(w, "Failed to create result", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, createResponse{ID: id})
}

func (h *Handler) GetResultByMatchID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.PathValue("match_id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid match id", http.StatusBadRequest)
		return
	}

	result, err := h.robotService.GetResultByMatchID(r.Context(), domain.MatchID(id))
	if err != nil {
		http.Error(w, "Failed to retrieve result", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
