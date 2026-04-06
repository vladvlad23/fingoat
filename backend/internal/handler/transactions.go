package handler

import (
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/fingoat/api/internal/config"
	"github.com/fingoat/api/internal/db"
	"github.com/fingoat/api/internal/middleware"
	"github.com/fingoat/api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TransactionHandler struct {
	store  Store
	pool   TxBeginner
	config *config.Config
}

func NewTransactionHandler(q Store, pool TxBeginner, cfg *config.Config) *TransactionHandler {
	return &TransactionHandler{store: q, pool: pool, config: cfg}
}

type createTransactionRequest struct {
	GoalID   *int64  `json:"goalId,omitempty"`
	Title    string  `json:"title"`
	Amount   string  `json:"amount"`
	Type     string  `json:"type"`
	Category *string `json:"category,omitempty"`
	Date     string  `json:"date"`
}

type updateTransactionRequest struct {
	GoalID   *int64  `json:"goalId,omitempty"`
	Title    string  `json:"title"`
	Amount   string  `json:"amount"`
	Type     string  `json:"type"`
	Category *string `json:"category,omitempty"`
	Date     string  `json:"date"`
}

func negateNumeric(n pgtype.Numeric) pgtype.Numeric {
	if !n.Valid || n.Int == nil {
		return n
	}
	return pgtype.Numeric{
		Int:              new(big.Int).Neg(n.Int),
		Exp:              n.Exp,
		Valid:            n.Valid,
		NaN:              n.NaN,
		InfinityModifier: n.InfinityModifier,
	}
}

func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	params := db.ListTransactionsParams{UserID: userID}
	if v := r.URL.Query().Get("goalId"); v != "" {
		id, err := parseInt64(v)
		if err == nil {
			params.GoalID = &id
		}
	}
	if v := r.URL.Query().Get("type"); v != "" {
		params.Type = &v
	}
	if v := r.URL.Query().Get("from"); v != "" {
		d, ok := parseDate(v)
		if ok {
			params.From = &d
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		d, ok := parseDate(v)
		if ok {
			params.To = &d
		}
	}
	txs, err := h.store.ListTransactions(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}
	resp := make([]model.TransactionResponse, len(txs))
	for i, t := range txs {
		resp[i] = model.TransactionToResponse(t)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" || req.Amount == "" || req.Type == "" || req.Date == "" {
		writeError(w, http.StatusBadRequest, "title, amount, type, and date are required")
		return
	}
	amount, ok := parseNumeric(req.Amount)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid amount")
		return
	}
	date, ok := parseDate(req.Date)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	if req.GoalID != nil {
		pgTx, err := h.pool.BeginTx(r.Context(), pgx.TxOptions{})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to begin transaction")
			return
		}
		defer pgTx.Rollback(r.Context())

		qtx := h.store.WithTx(pgTx)
		// Verify goal ownership
		goal, err := qtx.GetGoalByID(r.Context(), *req.GoalID)
		if err != nil || goal.UserID != userID {
			writeError(w, http.StatusForbidden, "goal not found or forbidden")
			return
		}
		t, err := qtx.CreateTransaction(r.Context(), db.CreateTransactionParams{
			UserID:   userID,
			GoalID:   req.GoalID,
			Title:    req.Title,
			Amount:   amount,
			Type:     req.Type,
			Category: req.Category,
			Date:     date,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create transaction")
			return
		}
		if _, err := qtx.UpdateGoalCurrentAmount(r.Context(), *req.GoalID, amount); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update goal amount")
			return
		}
		if err := pgTx.Commit(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to commit transaction")
			return
		}
		writeJSON(w, http.StatusCreated, model.TransactionToResponse(t))
		return
	}

	t, err := h.store.CreateTransaction(r.Context(), db.CreateTransactionParams{
		UserID:   userID,
		GoalID:   nil,
		Title:    req.Title,
		Amount:   amount,
		Type:     req.Type,
		Category: req.Category,
		Date:     date,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create transaction")
		return
	}
	writeJSON(w, http.StatusCreated, model.TransactionToResponse(t))
}

func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	t, err := h.store.GetTransactionByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}
	if t.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	writeJSON(w, http.StatusOK, model.TransactionToResponse(t))
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	existing, err := h.store.GetTransactionByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}
	if existing.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req updateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	newAmount, ok := parseNumeric(req.Amount)
	if !ok {
		newAmount = existing.Amount
	}
	newDate, ok := parseDate(req.Date)
	if !ok {
		newDate = existing.Date
	}
	title := req.Title
	if title == "" {
		title = existing.Title
	}
	txType := req.Type
	if txType == "" {
		txType = existing.Type
	}

	oldGoalID := existing.GoalID
	newGoalID := req.GoalID

	needsGoalUpdate := oldGoalID != nil || newGoalID != nil
	if needsGoalUpdate {
		pgTx, err := h.pool.BeginTx(r.Context(), pgx.TxOptions{})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to begin transaction")
			return
		}
		defer pgTx.Rollback(r.Context())
		qtx := h.store.WithTx(pgTx)

		// Reverse old goal delta
		if oldGoalID != nil {
			negOld := negateNumeric(existing.Amount)
			if _, err := qtx.UpdateGoalCurrentAmount(r.Context(), *oldGoalID, negOld); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to update old goal")
				return
			}
		}
		// Apply new goal delta
		if newGoalID != nil {
			goal, err := qtx.GetGoalByID(r.Context(), *newGoalID)
			if err != nil || goal.UserID != userID {
				writeError(w, http.StatusForbidden, "new goal not found or forbidden")
				return
			}
			if _, err := qtx.UpdateGoalCurrentAmount(r.Context(), *newGoalID, newAmount); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to update new goal")
				return
			}
		}
		t, err := qtx.UpdateTransaction(r.Context(), db.UpdateTransactionParams{
			ID: id, Title: title, Amount: newAmount, Type: txType,
			Category: req.Category, Date: newDate, GoalID: newGoalID,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update transaction")
			return
		}
		if err := pgTx.Commit(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to commit")
			return
		}
		writeJSON(w, http.StatusOK, model.TransactionToResponse(t))
		return
	}

	t, err := h.store.UpdateTransaction(r.Context(), db.UpdateTransactionParams{
		ID: id, Title: title, Amount: newAmount, Type: txType,
		Category: req.Category, Date: newDate, GoalID: newGoalID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update transaction")
		return
	}
	writeJSON(w, http.StatusOK, model.TransactionToResponse(t))
}

func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	existing, err := h.store.GetTransactionByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}
	if existing.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if existing.GoalID != nil {
		pgTx, err := h.pool.BeginTx(r.Context(), pgx.TxOptions{})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to begin transaction")
			return
		}
		defer pgTx.Rollback(r.Context())
		qtx := h.store.WithTx(pgTx)
		negAmount := negateNumeric(existing.Amount)
		if _, err := qtx.UpdateGoalCurrentAmount(r.Context(), *existing.GoalID, negAmount); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update goal amount")
			return
		}
		if err := qtx.DeleteTransaction(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete transaction")
			return
		}
		if err := pgTx.Commit(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to commit")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := h.store.DeleteTransaction(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete transaction")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
