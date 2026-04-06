package handler

import (
	"encoding/json"
	"net/http"

	"github.com/fingoat/api/internal/config"
	"github.com/fingoat/api/internal/db"
	"github.com/fingoat/api/internal/middleware"
	"github.com/fingoat/api/internal/model"
	"github.com/fingoat/api/internal/stores"
	"github.com/jackc/pgx/v5/pgtype"
)

type GoalHandler struct {
	store  stores.GoalStore
	config *config.Config
}

func NewGoalHandler(q stores.GoalStore, cfg *config.Config) *GoalHandler {
	return &GoalHandler{store: q, config: cfg}
}

type createGoalRequest struct {
	Title        string  `json:"title"`
	TargetAmount string  `json:"targetAmount"`
	Deadline     *string `json:"deadline,omitempty"`
}

type updateGoalRequest struct {
	Title        string  `json:"title"`
	TargetAmount string  `json:"targetAmount"`
	Deadline     *string `json:"deadline,omitempty"`
	Status       string  `json:"status"`
}

func (h *GoalHandler) ListGoals(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	goals, err := h.store.ListGoalsByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch goals")
		return
	}
	resp := make([]model.GoalResponse, len(goals))
	for i, g := range goals {
		resp[i] = model.GoalToResponse(g)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *GoalHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" || req.TargetAmount == "" {
		writeError(w, http.StatusBadRequest, "title and targetAmount are required")
		return
	}
	target, ok := parseNumeric(req.TargetAmount)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid targetAmount")
		return
	}
	var deadline *pgtype.Date
	if req.Deadline != nil {
		d, ok := parseDate(*req.Deadline)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid deadline format, use YYYY-MM-DD")
			return
		}
		deadline = &d
	}
	goal, err := h.store.CreateGoal(r.Context(), db.CreateGoalParams{
		UserID:       userID,
		Title:        req.Title,
		TargetAmount: target,
		Deadline:     deadline,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create goal")
		return
	}
	writeJSON(w, http.StatusCreated, model.GoalToResponse(goal))
}

func (h *GoalHandler) GetGoal(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid goal id")
		return
	}
	goal, err := h.store.GetGoalByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "goal not found")
		return
	}
	if goal.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	writeJSON(w, http.StatusOK, model.GoalToResponse(goal))
}

func (h *GoalHandler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid goal id")
		return
	}
	existing, err := h.store.GetGoalByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "goal not found")
		return
	}
	if existing.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req updateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		req.Title = existing.Title
	}
	target, ok := parseNumeric(req.TargetAmount)
	if !ok {
		target = existing.TargetAmount
	}
	status := req.Status
	if status == "" {
		status = existing.Status
	}
	var deadline *pgtype.Date
	if req.Deadline != nil {
		d, ok := parseDate(*req.Deadline)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid deadline format")
			return
		}
		deadline = &d
	} else {
		deadline = existing.Deadline
	}
	goal, err := h.store.UpdateGoal(r.Context(), db.UpdateGoalParams{
		ID:           id,
		Title:        req.Title,
		TargetAmount: target,
		Deadline:     deadline,
		Status:       status,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update goal")
		return
	}
	writeJSON(w, http.StatusOK, model.GoalToResponse(goal))
}

func (h *GoalHandler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid goal id")
		return
	}
	existing, err := h.store.GetGoalByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "goal not found")
		return
	}
	if existing.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if err := h.store.DeleteGoal(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete goal")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
