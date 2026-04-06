package handler

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fingoat/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- helpers ---

func makeDateForSummary(s string) pgtype.Date {
	var d pgtype.Date
	_ = d.Scan(s)
	return d
}

// makeNumericFromInt builds a pgtype.Numeric with value n * 10^exp.
func makeNumericFromInt(n int64, exp int32) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   big.NewInt(n),
		Exp:   exp,
		Valid: true,
	}
}

// sampleSummaryRows returns rows simulating two expense rows and one income row.
func sampleSummaryRows() []db.TransactionSummaryRow {
	groceries := "Groceries"
	return []db.TransactionSummaryRow{
		{Type: "expense", Category: groceries, Total: makeNumericFromInt(32000, -2)}, // 320.00
		{Type: "expense", Category: "Rent", Total: makeNumericFromInt(113000, -2)},   // 1130.00
		{Type: "income", Category: "Salary", Total: makeNumericFromInt(320000, -2)},  // 3200.00
	}
}

func sampleGoalsForSummary(userID int64) []db.Goal {
	return []db.Goal{
		{
			ID:            1,
			UserID:        userID,
			Title:         "Emergency Fund",
			TargetAmount:  makeNumericFromInt(500000, -2), // 5000.00
			CurrentAmount: makeNumericFromInt(120000, -2), // 1200.00
			Status:        "active",
			CreatedAt:     time.Now(),
		},
	}
}

// --- tests ---

func TestSummaryHandler_GetSummary_Success(t *testing.T) {
	t.Parallel()

	const userID int64 = 1
	store := &mockStore{
		getTransactionSummaryFn: func(_ context.Context, arg db.GetTransactionSummaryParams) ([]db.TransactionSummaryRow, error) {
			if arg.UserID != userID {
				t.Errorf("userID = %d, want %d", arg.UserID, userID)
			}
			return sampleSummaryRows(), nil
		},
		listGoalsByUserFn: func(_ context.Context, uid int64) ([]db.Goal, error) {
			return sampleGoalsForSummary(uid), nil
		},
	}
	h := NewSummaryHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/summary", nil)
	req = requestWithUserID(req, userID)
	rr := httptest.NewRecorder()

	h.GetSummary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp summaryResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Total income = 3200.00
	if resp.TotalIncome != "3200.00" {
		t.Errorf("TotalIncome = %q, want %q", resp.TotalIncome, "3200.00")
	}
	// Total expenses = 320.00 + 1130.00 = 1450.00
	if resp.TotalExpenses != "1450.00" {
		t.Errorf("TotalExpenses = %q, want %q", resp.TotalExpenses, "1450.00")
	}
	// Net = 3200.00 - 1450.00 = 1750.00
	if resp.Net != "1750.00" {
		t.Errorf("Net = %q, want %q", resp.Net, "1750.00")
	}

	// byCategory: Groceries (expense 320.00), Rent (expense 1130.00), Salary (income 3200.00)
	if len(resp.ByCategory) != 3 {
		t.Fatalf("len(byCategory) = %d, want 3", len(resp.ByCategory))
	}
	groceries := resp.ByCategory[0]
	if groceries.Category != "Groceries" {
		t.Errorf("byCategory[0].category = %q, want %q", groceries.Category, "Groceries")
	}
	if groceries.TotalExpenses != "320.00" {
		t.Errorf("byCategory[0].totalExpenses = %q, want %q", groceries.TotalExpenses, "320.00")
	}
	if groceries.TotalIncome != "0.00" {
		t.Errorf("byCategory[0].totalIncome = %q, want %q", groceries.TotalIncome, "0.00")
	}

	// goalProgress: 1200.00 / 5000.00 = 24%
	if len(resp.GoalProgress) != 1 {
		t.Fatalf("len(goalProgress) = %d, want 1", len(resp.GoalProgress))
	}
	gp := resp.GoalProgress[0]
	if gp.ID != 1 {
		t.Errorf("goalProgress[0].id = %d, want 1", gp.ID)
	}
	if gp.TargetAmount != "5000.00" {
		t.Errorf("goalProgress[0].targetAmount = %q, want %q", gp.TargetAmount, "5000.00")
	}
	if gp.ProgressPct != 24.0 {
		t.Errorf("goalProgress[0].progressPct = %v, want 24.0", gp.ProgressPct)
	}
}

func TestSummaryHandler_GetSummary_DefaultDateRange(t *testing.T) {
	t.Parallel()

	// Capture the params that arrive at the store to confirm defaults are set.
	var capturedParams db.GetTransactionSummaryParams
	store := &mockStore{
		getTransactionSummaryFn: func(_ context.Context, arg db.GetTransactionSummaryParams) ([]db.TransactionSummaryRow, error) {
			capturedParams = arg
			return nil, nil
		},
		listGoalsByUserFn: func(_ context.Context, _ int64) ([]db.Goal, error) {
			return nil, nil
		},
	}
	h := NewSummaryHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/summary", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.GetSummary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	if capturedParams.From == nil || !capturedParams.From.Valid {
		t.Error("expected non-nil From date with default range")
	}
	if capturedParams.To == nil || !capturedParams.To.Valid {
		t.Error("expected non-nil To date with default range")
	}

	// The From day should be the 1st of the current month.
	now := time.Now().UTC()
	wantFrom := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if !capturedParams.From.Time.Equal(wantFrom) {
		t.Errorf("From = %v, want %v", capturedParams.From.Time, wantFrom)
	}
}

func TestSummaryHandler_GetSummary_ExplicitDateRange(t *testing.T) {
	t.Parallel()

	var capturedParams db.GetTransactionSummaryParams
	store := &mockStore{
		getTransactionSummaryFn: func(_ context.Context, arg db.GetTransactionSummaryParams) ([]db.TransactionSummaryRow, error) {
			capturedParams = arg
			return nil, nil
		},
		listGoalsByUserFn: func(_ context.Context, _ int64) ([]db.Goal, error) {
			return nil, nil
		},
	}
	h := NewSummaryHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/summary?from=2026-01-01&to=2026-03-31", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.GetSummary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	wantFrom := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	wantTo := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	if !capturedParams.From.Time.Equal(wantFrom) {
		t.Errorf("From = %v, want %v", capturedParams.From.Time, wantFrom)
	}
	if !capturedParams.To.Time.Equal(wantTo) {
		t.Errorf("To = %v, want %v", capturedParams.To.Time, wantTo)
	}
}

func TestSummaryHandler_GetSummary_EmptyResults(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionSummaryFn: func(_ context.Context, _ db.GetTransactionSummaryParams) ([]db.TransactionSummaryRow, error) {
			return nil, nil
		},
		listGoalsByUserFn: func(_ context.Context, _ int64) ([]db.Goal, error) {
			return nil, nil
		},
	}
	h := NewSummaryHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/summary", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.GetSummary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp summaryResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.TotalIncome != "0.00" {
		t.Errorf("TotalIncome = %q, want %q", resp.TotalIncome, "0.00")
	}
	if resp.TotalExpenses != "0.00" {
		t.Errorf("TotalExpenses = %q, want %q", resp.TotalExpenses, "0.00")
	}
	if resp.Net != "0.00" {
		t.Errorf("Net = %q, want %q", resp.Net, "0.00")
	}
	if len(resp.ByCategory) != 0 {
		t.Errorf("len(byCategory) = %d, want 0", len(resp.ByCategory))
	}
	if len(resp.GoalProgress) != 0 {
		t.Errorf("len(goalProgress) = %d, want 0", len(resp.GoalProgress))
	}
}

func TestSummaryHandler_GetSummary_NoAuth(t *testing.T) {
	t.Parallel()

	h := NewSummaryHandler(&mockStore{}, testConfig())
	req := httptest.NewRequest(http.MethodGet, "/api/summary", nil)
	rr := httptest.NewRecorder()

	h.GetSummary(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestSummaryHandler_GetSummary_InvalidFromDate(t *testing.T) {
	t.Parallel()

	h := NewSummaryHandler(&mockStore{}, testConfig())
	req := httptest.NewRequest(http.MethodGet, "/api/summary?from=not-a-date", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.GetSummary(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSummaryHandler_GetSummary_InvalidToDate(t *testing.T) {
	t.Parallel()

	h := NewSummaryHandler(&mockStore{}, testConfig())
	req := httptest.NewRequest(http.MethodGet, "/api/summary?to=bad", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.GetSummary(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSummaryHandler_GetSummary_TransactionSummaryDBError(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionSummaryFn: func(_ context.Context, _ db.GetTransactionSummaryParams) ([]db.TransactionSummaryRow, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewSummaryHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/summary", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.GetSummary(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestSummaryHandler_GetSummary_GoalsDBError(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getTransactionSummaryFn: func(_ context.Context, _ db.GetTransactionSummaryParams) ([]db.TransactionSummaryRow, error) {
			return nil, nil
		},
		listGoalsByUserFn: func(_ context.Context, _ int64) ([]db.Goal, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewSummaryHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/summary", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.GetSummary(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestSummaryHandler_GetSummary_ZeroTargetGoal(t *testing.T) {
	t.Parallel()

	// A goal with zero target_amount should yield progressPct = 0 without panic.
	store := &mockStore{
		getTransactionSummaryFn: func(_ context.Context, _ db.GetTransactionSummaryParams) ([]db.TransactionSummaryRow, error) {
			return nil, nil
		},
		listGoalsByUserFn: func(_ context.Context, uid int64) ([]db.Goal, error) {
			return []db.Goal{
				{
					ID:            5,
					UserID:        uid,
					Title:         "Zero Target",
					TargetAmount:  makeNumericFromInt(0, 0),
					CurrentAmount: makeNumericFromInt(10000, -2),
					Status:        "active",
					CreatedAt:     time.Now(),
				},
			}, nil
		},
	}
	h := NewSummaryHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/summary", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.GetSummary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp summaryResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.GoalProgress) != 1 {
		t.Fatalf("len(goalProgress) = %d, want 1", len(resp.GoalProgress))
	}
	if resp.GoalProgress[0].ProgressPct != 0.0 {
		t.Errorf("progressPct = %v, want 0.0", resp.GoalProgress[0].ProgressPct)
	}
}

func TestBuildSummaryResponse_PeriodDates(t *testing.T) {
	t.Parallel()

	from := makeDateForSummary("2026-04-01")
	to := makeDateForSummary("2026-04-30")

	resp := buildSummaryResponse(from, to, nil, nil)

	if resp.Period.From != "2026-04-01" {
		t.Errorf("period.from = %q, want %q", resp.Period.From, "2026-04-01")
	}
	if resp.Period.To != "2026-04-30" {
		t.Errorf("period.to = %q, want %q", resp.Period.To, "2026-04-30")
	}
}
