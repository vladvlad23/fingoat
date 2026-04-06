package handler

import (
	"encoding/json"
	"net/http"

	"github.com/fingoat/api/internal/config"
	"github.com/fingoat/api/internal/db"
	"github.com/fingoat/api/internal/middleware"
	"github.com/fingoat/api/internal/model"
	"github.com/fingoat/api/internal/stores"
)

type UserHandler struct {
	store  stores.UserStore
	config *config.Config
}

func NewUserHandler(q stores.UserStore, cfg *config.Config) *UserHandler {
	return &UserHandler{store: q, config: cfg}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.store.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, model.UserToResponse(user))
}

type updateMeRequest struct {
	MonthlyIncome *string `json:"monthlyIncome"`
	PaymentDay    *int32  `json:"paymentDay"`
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.MonthlyIncome == nil {
		writeError(w, http.StatusBadRequest, "monthlyIncome is required")
		return
	}
	income, ok := parseNumeric(*req.MonthlyIncome)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid monthlyIncome")
		return
	}
	if req.PaymentDay != nil {
		if *req.PaymentDay < 1 || *req.PaymentDay > 28 {
			writeError(w, http.StatusBadRequest, "paymentDay must be between 1 and 28")
			return
		}
	}
	user, err := h.store.UpdateUserSettings(r.Context(), db.UpdateUserSettingsParams{
		ID:            userID,
		MonthlyIncome: income,
		PaymentDay:    req.PaymentDay,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update settings")
		return
	}
	writeJSON(w, http.StatusOK, model.UserToResponse(user))
}
