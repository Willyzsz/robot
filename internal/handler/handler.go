package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"robot/internal/domain"
	"robot/internal/service"
	"robot/pkg/apperr"
	"strconv"
	"time"
)

type Handler struct {
	robotService *service.RobotService
}

type createCategoryRequest struct {
	Name string `json:"name"`
}

type createRuleRequest struct {
	Description string            `json:"description"`
	Type        domain.RuleType   `json:"type"`
	CategoryID  domain.CategoryID `json:"category_id"`
}

type createTeamRequest struct {
	Name       string            `json:"name"`
	School     string            `json:"school"`
	Grade      string            `json:"grade"`
	Teacher    string            `json:"teacher"`
	CategoryID domain.CategoryID `json:"category_id"`
}

type createMemberRequest struct {
	Name     string        `json:"name"`
	Email    string        `json:"email"`
	IsLeader bool          `json:"is_leader"`
	TeamID   domain.TeamID `json:"team_id"`
}

type createMatchRequest struct {
	TeamAID domain.TeamID   `json:"team_a_id"`
	TeamBID domain.TeamID   `json:"team_b_id"`
	Queue   []domain.TeamID `json:"queue"`
}

type startMatchQueueRequest struct {
	Mode domain.MatchMode `json:"mode"`
}

type createResultRequest struct {
	Winner domain.TeamID `json:"winner"`
	Time   *time.Time    `json:"time,omitempty"`
}

type createRobotRequest struct {
	TeamID     domain.TeamID   `json:"team_id"`
	ValidRules []domain.RuleID `json:"valid_rules"`
}

type verifyRobotRequest struct {
	RuleID domain.RuleID `json:"rule_id"`
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

func (h *Handler) GetTeamsWithMembersByCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.PathValue("category_id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}

	teams, err := h.robotService.GetTeamsWithMembersByCategory(r.Context(), domain.CategoryID(id))
	if err != nil {
		http.Error(w, "Failed to retrieve teams", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, teams)
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

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid category payload", http.StatusBadRequest)
		return
	}

	id, err := h.robotService.CreateCategory(r.Context(), req.Name)
	if err != nil {
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, createResponse{ID: id})
}

func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid rule payload", http.StatusBadRequest)
		return
	}

	id, err := h.robotService.CreateRule(r.Context(), req.Description, req.Type, req.CategoryID)
	if err != nil {
		handleError(w, r, "Failed to create rule", err, req)
		return
	}
	writeJSON(w, http.StatusCreated, createResponse{ID: id})
}

func (h *Handler) GetRulesByCategoryID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.PathValue("category_id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}

	rules, err := h.robotService.GetRulesByCategoryID(r.Context(), domain.CategoryID(id))
	if err != nil {
		http.Error(w, "Failed to retrieve rules", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid team payload", http.StatusBadRequest)
		return
	}

	id, err := h.robotService.CreateTeam(r.Context(), req.Name, req.School, req.Grade, req.Teacher, req.CategoryID)
	if err != nil {
		http.Error(w, "Failed to create team", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, createResponse{ID: id})
}

func (h *Handler) GetTeamsByCategoryID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.PathValue("category_id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}

	teams, err := h.robotService.GetTeamsByCategoryID(r.Context(), domain.CategoryID(id))
	if err != nil {
		http.Error(w, "Failed to retrieve teams", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, teams)
}

func (h *Handler) CreateMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid member payload", http.StatusBadRequest)
		return
	}

	id, err := h.robotService.CreateMember(r.Context(), req.Name, req.Email, req.IsLeader, req.TeamID)
	if err != nil {
		http.Error(w, "Failed to create member", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, createResponse{ID: id})
}

func (h *Handler) GetMembersByTeamID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.PathValue("team_id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid team id", http.StatusBadRequest)
		return
	}

	members, err := h.robotService.GetMembersByTeamID(r.Context(), domain.TeamID(id))
	if err != nil {
		http.Error(w, "Failed to retrieve members", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, members)
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

func (h *Handler) StartMatchQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.PathValue("category_id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}

	var req startMatchQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, "invalid match queue payload", http.StatusBadRequest)
		return
	}

	matches, err := h.robotService.StartMatchQueue(r.Context(), domain.CategoryID(id), req.Mode)
	if err != nil {
		handleError(w, r, "Failed to start match queue", err, req)
		return
	}
	writeJSON(w, http.StatusCreated, matches)
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

func (h *Handler) GetAllRobots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	robots, err := h.robotService.GetAllRobots(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve robots", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, robots)
}

func (h *Handler) CreateRobot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createRobotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid robot payload", http.StatusBadRequest)
		return
	}

	id, err := h.robotService.CreateRobot(r.Context(), req.TeamID, req.ValidRules)
	if err != nil {
		http.Error(w, "Failed to create robot", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, createResponse{ID: id})
}

func (h *Handler) GetRobotByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid robot id", http.StatusBadRequest)
		return
	}

	robot, err := h.robotService.GetRobotByID(r.Context(), domain.RobotID(id))
	if err != nil {
		http.Error(w, "Failed to retrieve robot", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, robot)
}

func (h *Handler) GetRobotsByTeamID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.PathValue("team_id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid team id", http.StatusBadRequest)
		return
	}

	robots, err := h.robotService.GetRobotsByTeamID(r.Context(), domain.TeamID(id))
	if err != nil {
		http.Error(w, "Failed to retrieve robots", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, robots)
}

func (h *Handler) VerifyRobot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid robot id", http.StatusBadRequest)
		return
	}

	var req verifyRobotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid verification payload", http.StatusBadRequest)
		return
	}

	if err := h.robotService.VerifyRobot(r.Context(), domain.RobotID(id), req.RuleID); err != nil {
		http.Error(w, "Failed to verify robot", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func handleError(w http.ResponseWriter, r *http.Request, msg string, err error, payload any) {
	status := statusForError(err)
	formatted := apperr.Format(err)
	log.Printf("%s %s failed -> %d payload=%+v err=%s", r.Method, r.URL.Path, status, payload, formatted)
	http.Error(w, msg+": "+formatted, status)
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, domain.ErrEmpty), errors.Is(err, domain.ErrInvalid), errors.Is(err, domain.ErrInvalidReference), errors.Is(err, domain.ErrNotEnough):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
