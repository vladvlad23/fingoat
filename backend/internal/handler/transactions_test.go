package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fingoat/api/internal/db"
	"github.com/fingoat/api/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

func sampleTransaction(userID int64) db.Transaction {
	return db.Transaction{
		ID:        100,
		UserID:    userID,
		Title:     "Coffee",
		Amount:    makeNumericForHandler("5.50"),
		Type:      "expense",
		Date:      pgtype.Date{Time: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), Valid: true},
		CreatedAt: time.Now(),
	}
}

func newTxHandler(store *mockStore) *TransactionHandler {
	return NewTransactionHandler(store, &mockTxBeginner{}, testConfig())
}

// --- ListTransactions ---

func TestTransactionHandler_ListTransactions_Success(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		listTransactionsFn: func(_ context.Context, arg db.ListTransactionsParams) ([]db.Transaction, error) {
			return []db.Transaction{sampleTransaction(arg.UserID)}, nil
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.ListTransactions(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []model.TransactionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("len = %d, want 1", len(resp))
	}
}

func TestTransactionHandler_ListTransactions_Empty(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		listTransactionsFn: func(_ context.Context, _ db.ListTransactionsParams) ([]db.Transaction, error) {
			return []db.Transaction{}, nil
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.ListTransactions(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestTransactionHandler_ListTransactions_WithFilters(t *testing.T) {
	t.Parallel()

	var capturedParams db.ListTransactionsParams
	store := &mockStore{
		listTransactionsFn: func(_ context.Context, arg db.ListTransactionsParams) ([]db.Transaction, error) {
			capturedParams = arg
			return nil, nil
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/transactions?type=income&goalId=5&from=2025-01-01&to=2025-12-31", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.ListTransactions(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if capturedParams.Type == nil || *capturedParams.Type != "income" {
		t.Errorf("Type = %v, want %q", capturedParams.Type, "income")
	}
	if capturedParams.GoalID == nil || *capturedParams.GoalID != 5 {
		t.Errorf("GoalID = %v, want 5", capturedParams.GoalID)
	}
	if capturedParams.From == nil {
		t.Error("From should be set")
	}
	if capturedParams.To == nil {
		t.Error("To should be set")
	}
}

func TestTransactionHandler_ListTransactions_NoAuth(t *testing.T) {
	t.Parallel()

	h := newTxHandler(&mockStore{})
	req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
	rr := httptest.NewRecorder()

	h.ListTransactions(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestTransactionHandler_ListTransactions_DBError(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		listTransactionsFn: func(_ context.Context, _ db.ListTransactionsParams) ([]db.Transaction, error) {
			return nil, errors.New("db error")
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.ListTransactions(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// --- CreateTransaction (no goal) ---

func TestTransactionHandler_CreateTransaction_Success_NoGoal(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		createTransactionFn: func(_ context.Context, arg db.CreateTransactionParams) (db.Transaction, error) {
			return db.Transaction{
				ID:        101,
				UserID:    arg.UserID,
				Title:     arg.Title,
				Amount:    arg.Amount,
				Type:      arg.Type,
				Date:      arg.Date,
				CreatedAt: time.Now(),
			}, nil
		},
	}
	h := newTxHandler(store)

	body := `{"title":"Lunch","amount":"12.50","type":"expense","date":"2025-06-15"}`
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.CreateTransaction(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp model.TransactionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Title != "Lunch" {
		t.Errorf("Title = %q, want %q", resp.Title, "Lunch")
	}
}

// --- CreateTransaction (with goal — uses pgx transaction) ---

func TestTransactionHandler_CreateTransaction_Success_WithGoal(t *testing.T) {
	t.Parallel()

	goalID := int64(10)
	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, id int64) (db.Goal, error) {
			return sampleGoal(1), nil // owned by user 1
		},
		createTransactionFn: func(_ context.Context, arg db.CreateTransactionParams) (db.Transaction, error) {
			return db.Transaction{
				ID:        102,
				UserID:    arg.UserID,
				GoalID:    arg.GoalID,
				Title:     arg.Title,
				Amount:    arg.Amount,
				Type:      arg.Type,
				Date:      arg.Date,
				CreatedAt: time.Now(),
			}, nil
		},
		updateGoalCurrentAmountFn: func(_ context.Context, id int64, delta pgtype.Numeric) (db.Goal, error) {
			return sampleGoal(1), nil
		},
	}
	h := newTxHandler(store)

	body := `{"title":"Savings","amount":"100.00","type":"income","date":"2025-06-15","goalId":10}`
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.CreateTransaction(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	_ = goalID
}

func TestTransactionHandler_CreateTransaction_WithGoal_Forbidden(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return sampleGoal(99), nil // owned by user 99, not user 1
		},
	}
	h := newTxHandler(store)

	body := `{"title":"Savings","amount":"100.00","type":"income","date":"2025-06-15","goalId":10}`
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.CreateTransaction(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestTransactionHandler_CreateTransaction_MissingFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{"missing title", `{"amount":"10","type":"expense","date":"2025-01-01"}`},
		{"missing amount", `{"title":"T","type":"expense","date":"2025-01-01"}`},
		{"missing type", `{"title":"T","amount":"10","date":"2025-01-01"}`},
		{"missing date", `{"title":"T","amount":"10","type":"expense"}`},
		{"empty title", `{"title":"","amount":"10","type":"expense","date":"2025-01-01"}`},
	}

	h := newTxHandler(&mockStore{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(tt.body))
			req = requestWithUserID(req, 1)
			rr := httptest.NewRecorder()
			h.CreateTransaction(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
			}
		})
	}
}

func TestTransactionHandler_CreateTransaction_InvalidAmount(t *testing.T) {
	t.Parallel()

	h := newTxHandler(&mockStore{})
	body := `{"title":"T","amount":"not-a-number","type":"expense","date":"2025-01-01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.CreateTransaction(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestTransactionHandler_CreateTransaction_InvalidDate(t *testing.T) {
	t.Parallel()

	h := newTxHandler(&mockStore{})
	body := `{"title":"T","amount":"10","type":"expense","date":"not-a-date"}`
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.CreateTransaction(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestTransactionHandler_CreateTransaction_InvalidJSON(t *testing.T) {
	t.Parallel()

	h := newTxHandler(&mockStore{})
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader("{bad"))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.CreateTransaction(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestTransactionHandler_CreateTransaction_NoAuth(t *testing.T) {
	t.Parallel()

	h := newTxHandler(&mockStore{})
	body := `{"title":"T","amount":"10","type":"expense","date":"2025-01-01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.CreateTransaction(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// --- GetTransaction ---

func TestTransactionHandler_GetTransaction_Success(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, id int64) (db.Transaction, error) {
			tx := sampleTransaction(1)
			tx.ID = id
			return tx, nil
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/transactions/100", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "100")
	rr := httptest.NewRecorder()

	h.GetTransaction(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestTransactionHandler_GetTransaction_NotFound(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return db.Transaction{}, errNotFound
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/transactions/999", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	h.GetTransaction(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestTransactionHandler_GetTransaction_Forbidden(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return sampleTransaction(99), nil // owned by user 99
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/transactions/100", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "100")
	rr := httptest.NewRecorder()

	h.GetTransaction(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestTransactionHandler_GetTransaction_InvalidID(t *testing.T) {
	t.Parallel()

	h := newTxHandler(&mockStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/transactions/abc", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "abc")
	rr := httptest.NewRecorder()

	h.GetTransaction(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// --- UpdateTransaction (no goal changes) ---

func TestTransactionHandler_UpdateTransaction_Success_NoGoal(t *testing.T) {
	t.Parallel()

	existing := sampleTransaction(1)
	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return existing, nil
		},
		updateTransactionFn: func(_ context.Context, arg db.UpdateTransactionParams) (db.Transaction, error) {
			return db.Transaction{
				ID:        arg.ID,
				UserID:    1,
				Title:     arg.Title,
				Amount:    arg.Amount,
				Type:      arg.Type,
				Date:      arg.Date,
				CreatedAt: existing.CreatedAt,
			}, nil
		},
	}
	h := newTxHandler(store)

	body := `{"title":"Updated Coffee","amount":"6.00","type":"expense","date":"2025-06-20"}`
	req := httptest.NewRequest(http.MethodPut, "/api/transactions/100", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "100")
	rr := httptest.NewRecorder()

	h.UpdateTransaction(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp model.TransactionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Title != "Updated Coffee" {
		t.Errorf("Title = %q, want %q", resp.Title, "Updated Coffee")
	}
}

func TestTransactionHandler_UpdateTransaction_PartialUpdate(t *testing.T) {
	t.Parallel()

	existing := sampleTransaction(1)
	var capturedParams db.UpdateTransactionParams
	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return existing, nil
		},
		updateTransactionFn: func(_ context.Context, arg db.UpdateTransactionParams) (db.Transaction, error) {
			capturedParams = arg
			return db.Transaction{
				ID:        arg.ID,
				UserID:    1,
				Title:     arg.Title,
				Amount:    arg.Amount,
				Type:      arg.Type,
				Date:      arg.Date,
				CreatedAt: existing.CreatedAt,
			}, nil
		},
	}
	h := newTxHandler(store)

	// Only title — amount/type/date should fall back to existing
	body := `{"title":"New Name"}`
	req := httptest.NewRequest(http.MethodPut, "/api/transactions/100", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "100")
	rr := httptest.NewRecorder()

	h.UpdateTransaction(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if capturedParams.Title != "New Name" {
		t.Errorf("Title = %q, want %q", capturedParams.Title, "New Name")
	}
	if capturedParams.Type != existing.Type {
		t.Errorf("Type = %q, want existing %q", capturedParams.Type, existing.Type)
	}
}

func TestTransactionHandler_UpdateTransaction_WithNewGoal(t *testing.T) {
	t.Parallel()

	newGoalID := int64(10)
	existing := sampleTransaction(1) // no existing goal

	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return existing, nil
		},
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return sampleGoal(1), nil
		},
		updateGoalCurrentAmountFn: func(_ context.Context, _ int64, _ pgtype.Numeric) (db.Goal, error) {
			return sampleGoal(1), nil
		},
		updateTransactionFn: func(_ context.Context, arg db.UpdateTransactionParams) (db.Transaction, error) {
			return db.Transaction{
				ID:        arg.ID,
				UserID:    1,
				GoalID:    arg.GoalID,
				Title:     arg.Title,
				Amount:    arg.Amount,
				Type:      arg.Type,
				Date:      arg.Date,
				CreatedAt: existing.CreatedAt,
			}, nil
		},
	}
	h := newTxHandler(store)

	body := `{"title":"Coffee","amount":"5.50","type":"expense","date":"2025-06-15","goalId":10}`
	req := httptest.NewRequest(http.MethodPut, "/api/transactions/100", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "100")
	rr := httptest.NewRecorder()

	h.UpdateTransaction(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	_ = newGoalID
}

func TestTransactionHandler_UpdateTransaction_NotFound(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return db.Transaction{}, errNotFound
		},
	}
	h := newTxHandler(store)

	body := `{"title":"T"}`
	req := httptest.NewRequest(http.MethodPut, "/api/transactions/999", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	h.UpdateTransaction(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestTransactionHandler_UpdateTransaction_Forbidden(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return sampleTransaction(99), nil // owned by user 99
		},
	}
	h := newTxHandler(store)

	body := `{"title":"T"}`
	req := httptest.NewRequest(http.MethodPut, "/api/transactions/100", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "100")
	rr := httptest.NewRecorder()

	h.UpdateTransaction(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

// --- DeleteTransaction ---

func TestTransactionHandler_DeleteTransaction_Success_NoGoal(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return sampleTransaction(1), nil // no GoalID
		},
		deleteTransactionFn: func(_ context.Context, _ int64) error {
			return nil
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/transactions/100", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "100")
	rr := httptest.NewRecorder()

	h.DeleteTransaction(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}

func TestTransactionHandler_DeleteTransaction_Success_WithGoal(t *testing.T) {
	t.Parallel()

	goalID := int64(10)
	tx := sampleTransaction(1)
	tx.GoalID = &goalID

	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return tx, nil
		},
		updateGoalCurrentAmountFn: func(_ context.Context, _ int64, _ pgtype.Numeric) (db.Goal, error) {
			return sampleGoal(1), nil
		},
		deleteTransactionFn: func(_ context.Context, _ int64) error {
			return nil
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/transactions/100", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "100")
	rr := httptest.NewRecorder()

	h.DeleteTransaction(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}

func TestTransactionHandler_DeleteTransaction_NotFound(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return db.Transaction{}, errNotFound
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/transactions/999", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	h.DeleteTransaction(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestTransactionHandler_DeleteTransaction_Forbidden(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return sampleTransaction(99), nil // owned by user 99
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/transactions/100", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "100")
	rr := httptest.NewRecorder()

	h.DeleteTransaction(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestTransactionHandler_DeleteTransaction_DBError(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionByIDFn: func(_ context.Context, _ int64) (db.Transaction, error) {
			return sampleTransaction(1), nil
		},
		deleteTransactionFn: func(_ context.Context, _ int64) error {
			return errors.New("db error")
		},
	}
	h := newTxHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/transactions/100", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "100")
	rr := httptest.NewRecorder()

	h.DeleteTransaction(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestTransactionHandler_DeleteTransaction_NoAuth(t *testing.T) {
	t.Parallel()

	h := newTxHandler(&mockStore{})
	req := httptest.NewRequest(http.MethodDelete, "/api/transactions/100", nil)
	req = requestWithChiURLParam(req, "id", "100")
	rr := httptest.NewRecorder()

	h.DeleteTransaction(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}
