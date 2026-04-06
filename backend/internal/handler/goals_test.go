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

func sampleGoal(userID int64) db.Goal {
	return db.Goal{
		ID:            10,
		UserID:        userID,
		Title:         "Vacation Fund",
		TargetAmount:  makeNumericForHandler("3000.00"),
		CurrentAmount: makeNumericForHandler("0.00"),
		Status:        "active",
		CreatedAt:     time.Now(),
	}
}

// --- ListGoals ---

func TestGoalHandler_ListGoals_Success(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		listGoalsByUserFn: func(_ context.Context, userID int64) ([]db.Goal, error) {
			return []db.Goal{sampleGoal(userID)}, nil
		},
	}
	h := NewGoalHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/goals", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.ListGoals(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []model.GoalResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("len = %d, want 1", len(resp))
	}
}

func TestGoalHandler_ListGoals_Empty(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		listGoalsByUserFn: func(_ context.Context, _ int64) ([]db.Goal, error) {
			return []db.Goal{}, nil
		},
	}
	h := NewGoalHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/goals", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.ListGoals(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestGoalHandler_ListGoals_NoAuth(t *testing.T) {
	t.Parallel()

	h := NewGoalHandler(&mockStore{}, testConfig())
	req := httptest.NewRequest(http.MethodGet, "/api/goals", nil)
	rr := httptest.NewRecorder()

	h.ListGoals(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestGoalHandler_ListGoals_DBError(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		listGoalsByUserFn: func(_ context.Context, _ int64) ([]db.Goal, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewGoalHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/goals", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.ListGoals(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// --- CreateGoal ---

func TestGoalHandler_CreateGoal_Success(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		createGoalFn: func(_ context.Context, arg db.CreateGoalParams) (db.Goal, error) {
			return db.Goal{
				ID:            10,
				UserID:        arg.UserID,
				Title:         arg.Title,
				TargetAmount:  arg.TargetAmount,
				CurrentAmount: makeNumericForHandler("0.00"),
				Deadline:      arg.Deadline,
				Status:        "active",
				CreatedAt:     time.Now(),
			}, nil
		},
	}
	h := NewGoalHandler(store, testConfig())

	body := `{"title":"Save for car","targetAmount":"5000.00","deadline":"2026-12-31"}`
	req := httptest.NewRequest(http.MethodPost, "/api/goals", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.CreateGoal(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp model.GoalResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Title != "Save for car" {
		t.Errorf("Title = %q, want %q", resp.Title, "Save for car")
	}
}

func TestGoalHandler_CreateGoal_NoDeadline(t *testing.T) {
	t.Parallel()

	var receivedDeadline *pgtype.Date
	store := &mockStore{
		createGoalFn: func(_ context.Context, arg db.CreateGoalParams) (db.Goal, error) {
			receivedDeadline = arg.Deadline
			return db.Goal{
				ID:            11,
				UserID:        arg.UserID,
				Title:         arg.Title,
				TargetAmount:  arg.TargetAmount,
				CurrentAmount: makeNumericForHandler("0.00"),
				Status:        "active",
				CreatedAt:     time.Now(),
			}, nil
		},
	}
	h := NewGoalHandler(store, testConfig())

	body := `{"title":"No deadline","targetAmount":"1000.00"}`
	req := httptest.NewRequest(http.MethodPost, "/api/goals", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.CreateGoal(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	if receivedDeadline != nil {
		t.Errorf("deadline should be nil, got %v", receivedDeadline)
	}
}

func TestGoalHandler_CreateGoal_MissingFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{"missing title", `{"targetAmount":"1000"}`},
		{"missing targetAmount", `{"title":"Goal"}`},
		{"empty title", `{"title":"","targetAmount":"1000"}`},
		{"empty targetAmount", `{"title":"Goal","targetAmount":""}`},
	}

	h := NewGoalHandler(&mockStore{}, testConfig())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodPost, "/api/goals", strings.NewReader(tt.body))
			req = requestWithUserID(req, 1)
			rr := httptest.NewRecorder()
			h.CreateGoal(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestGoalHandler_CreateGoal_InvalidTargetAmount(t *testing.T) {
	t.Parallel()

	h := NewGoalHandler(&mockStore{}, testConfig())
	body := `{"title":"Goal","targetAmount":"abc"}`
	req := httptest.NewRequest(http.MethodPost, "/api/goals", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.CreateGoal(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGoalHandler_CreateGoal_InvalidDeadline(t *testing.T) {
	t.Parallel()

	h := NewGoalHandler(&mockStore{}, testConfig())
	body := `{"title":"Goal","targetAmount":"1000","deadline":"not-a-date"}`
	req := httptest.NewRequest(http.MethodPost, "/api/goals", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.CreateGoal(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGoalHandler_CreateGoal_InvalidJSON(t *testing.T) {
	t.Parallel()

	h := NewGoalHandler(&mockStore{}, testConfig())
	req := httptest.NewRequest(http.MethodPost, "/api/goals", strings.NewReader("{bad"))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.CreateGoal(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGoalHandler_CreateGoal_NoAuth(t *testing.T) {
	t.Parallel()

	h := NewGoalHandler(&mockStore{}, testConfig())
	body := `{"title":"Goal","targetAmount":"1000"}`
	req := httptest.NewRequest(http.MethodPost, "/api/goals", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.CreateGoal(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// --- GetGoal ---

func TestGoalHandler_GetGoal_Success(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, id int64) (db.Goal, error) {
			g := sampleGoal(1)
			g.ID = id
			return g, nil
		},
	}
	h := NewGoalHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/goals/10", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "10")
	rr := httptest.NewRecorder()

	h.GetGoal(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestGoalHandler_GetGoal_NotFound(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return db.Goal{}, errNotFound
		},
	}
	h := NewGoalHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/goals/999", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	h.GetGoal(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGoalHandler_GetGoal_Forbidden(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return sampleGoal(99), nil // owned by user 99
		},
	}
	h := NewGoalHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/goals/10", nil)
	req = requestWithUserID(req, 1) // requesting as user 1
	req = requestWithChiURLParam(req, "id", "10")
	rr := httptest.NewRecorder()

	h.GetGoal(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestGoalHandler_GetGoal_InvalidID(t *testing.T) {
	t.Parallel()

	h := NewGoalHandler(&mockStore{}, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/goals/abc", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "abc")
	rr := httptest.NewRecorder()

	h.GetGoal(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// --- UpdateGoal ---

func TestGoalHandler_UpdateGoal_Success(t *testing.T) {
	t.Parallel()

	existing := sampleGoal(1)
	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return existing, nil
		},
		updateGoalFn: func(_ context.Context, arg db.UpdateGoalParams) (db.Goal, error) {
			return db.Goal{
				ID:            arg.ID,
				UserID:        1,
				Title:         arg.Title,
				TargetAmount:  arg.TargetAmount,
				CurrentAmount: existing.CurrentAmount,
				Deadline:      arg.Deadline,
				Status:        arg.Status,
				CreatedAt:     existing.CreatedAt,
			}, nil
		},
	}
	h := NewGoalHandler(store, testConfig())

	body := `{"title":"Updated Goal","targetAmount":"2000","status":"active"}`
	req := httptest.NewRequest(http.MethodPut, "/api/goals/10", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "10")
	rr := httptest.NewRecorder()

	h.UpdateGoal(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp model.GoalResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Title != "Updated Goal" {
		t.Errorf("Title = %q, want %q", resp.Title, "Updated Goal")
	}
}

func TestGoalHandler_UpdateGoal_PartialUpdate(t *testing.T) {
	t.Parallel()

	existing := sampleGoal(1)
	var receivedParams db.UpdateGoalParams
	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return existing, nil
		},
		updateGoalFn: func(_ context.Context, arg db.UpdateGoalParams) (db.Goal, error) {
			receivedParams = arg
			return db.Goal{
				ID:            arg.ID,
				UserID:        1,
				Title:         arg.Title,
				TargetAmount:  arg.TargetAmount,
				CurrentAmount: existing.CurrentAmount,
				Deadline:      arg.Deadline,
				Status:        arg.Status,
				CreatedAt:     existing.CreatedAt,
			}, nil
		},
	}
	h := NewGoalHandler(store, testConfig())

	// Only send title, other fields should default to existing values
	body := `{"title":"New Title"}`
	req := httptest.NewRequest(http.MethodPut, "/api/goals/10", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "10")
	rr := httptest.NewRecorder()

	h.UpdateGoal(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if receivedParams.Title != "New Title" {
		t.Errorf("Title = %q, want %q", receivedParams.Title, "New Title")
	}
	if receivedParams.Status != existing.Status {
		t.Errorf("Status = %q, want %q (existing)", receivedParams.Status, existing.Status)
	}
}

func TestGoalHandler_UpdateGoal_Forbidden(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return sampleGoal(99), nil // owned by user 99
		},
	}
	h := NewGoalHandler(store, testConfig())

	body := `{"title":"Hacked"}`
	req := httptest.NewRequest(http.MethodPut, "/api/goals/10", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "10")
	rr := httptest.NewRecorder()

	h.UpdateGoal(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestGoalHandler_UpdateGoal_NotFound(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return db.Goal{}, errNotFound
		},
	}
	h := NewGoalHandler(store, testConfig())

	body := `{"title":"Anything"}`
	req := httptest.NewRequest(http.MethodPut, "/api/goals/999", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	h.UpdateGoal(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGoalHandler_UpdateGoal_InvalidDeadline(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return sampleGoal(1), nil
		},
	}
	h := NewGoalHandler(store, testConfig())

	body := `{"title":"Goal","deadline":"bad-date"}`
	req := httptest.NewRequest(http.MethodPut, "/api/goals/10", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "10")
	rr := httptest.NewRecorder()

	h.UpdateGoal(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// --- DeleteGoal ---

func TestGoalHandler_DeleteGoal_Success(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return sampleGoal(1), nil
		},
		deleteGoalFn: func(_ context.Context, _ int64) error {
			return nil
		},
	}
	h := NewGoalHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodDelete, "/api/goals/10", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "10")
	rr := httptest.NewRecorder()

	h.DeleteGoal(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}

func TestGoalHandler_DeleteGoal_NotFound(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return db.Goal{}, errNotFound
		},
	}
	h := NewGoalHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodDelete, "/api/goals/999", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	h.DeleteGoal(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGoalHandler_DeleteGoal_Forbidden(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return sampleGoal(99), nil
		},
	}
	h := NewGoalHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodDelete, "/api/goals/10", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "10")
	rr := httptest.NewRecorder()

	h.DeleteGoal(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestGoalHandler_DeleteGoal_DBError(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getGoalByIDFn: func(_ context.Context, _ int64) (db.Goal, error) {
			return sampleGoal(1), nil
		},
		deleteGoalFn: func(_ context.Context, _ int64) error {
			return errors.New("db error")
		},
	}
	h := NewGoalHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodDelete, "/api/goals/10", nil)
	req = requestWithUserID(req, 1)
	req = requestWithChiURLParam(req, "id", "10")
	rr := httptest.NewRecorder()

	h.DeleteGoal(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}
