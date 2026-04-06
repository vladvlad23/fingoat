package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fingoat/api/internal/db"
	"github.com/fingoat/api/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestUserHandler_GetMe_Success(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getUserByIDFn: func(_ context.Context, id int64) (db.User, error) {
			return db.User{
				ID:        id,
				Email:     "user@example.com",
				CreatedAt: time.Now(),
			}, nil
		},
	}
	h := NewUserHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp model.UserResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != 1 {
		t.Errorf("ID = %d, want 1", resp.ID)
	}
	if resp.Email != "user@example.com" {
		t.Errorf("Email = %q, want %q", resp.Email, "user@example.com")
	}
}

func TestUserHandler_GetMe_NoUserInContext(t *testing.T) {
	t.Parallel()

	h := NewUserHandler(&mockStore{}, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestUserHandler_GetMe_UserNotFound(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getUserByIDFn: func(_ context.Context, _ int64) (db.User, error) {
			return db.User{}, errNotFound
		},
	}
	h := NewUserHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req = requestWithUserID(req, 999)
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestUserHandler_UpdateMe_Success(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		updateUserSettingsFn: func(_ context.Context, arg db.UpdateUserSettingsParams) (db.User, error) {
			return db.User{
				ID:            arg.ID,
				Email:         "user@example.com",
				MonthlyIncome: arg.MonthlyIncome,
				PaymentDay:    arg.PaymentDay,
				CreatedAt:     time.Now(),
			}, nil
		},
	}
	h := NewUserHandler(store, testConfig())

	body := `{"monthlyIncome":"5000.00","paymentDay":15}`
	req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.UpdateMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestUserHandler_UpdateMe_NoUserInContext(t *testing.T) {
	t.Parallel()

	h := NewUserHandler(&mockStore{}, testConfig())

	body := `{"monthlyIncome":"5000.00"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.UpdateMe(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestUserHandler_UpdateMe_InvalidJSON(t *testing.T) {
	t.Parallel()

	h := NewUserHandler(&mockStore{}, testConfig())

	req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader("{invalid"))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.UpdateMe(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_UpdateMe_MissingIncome(t *testing.T) {
	t.Parallel()

	h := NewUserHandler(&mockStore{}, testConfig())

	body := `{"paymentDay":15}`
	req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.UpdateMe(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_UpdateMe_InvalidIncome(t *testing.T) {
	t.Parallel()

	h := NewUserHandler(&mockStore{}, testConfig())

	body := `{"monthlyIncome":"not-a-number"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.UpdateMe(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_UpdateMe_InvalidPaymentDay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{"too low", `{"monthlyIncome":"5000","paymentDay":0}`},
		{"too high", `{"monthlyIncome":"5000","paymentDay":29}`},
		{"negative", `{"monthlyIncome":"5000","paymentDay":-1}`},
	}

	h := NewUserHandler(&mockStore{}, testConfig())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader(tt.body))
			req = requestWithUserID(req, 1)
			rr := httptest.NewRecorder()
			h.UpdateMe(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestUserHandler_UpdateMe_WithoutPaymentDay(t *testing.T) {
	t.Parallel()

	var receivedParams db.UpdateUserSettingsParams
	store := &mockStore{
		updateUserSettingsFn: func(_ context.Context, arg db.UpdateUserSettingsParams) (db.User, error) {
			receivedParams = arg
			return db.User{
				ID:            arg.ID,
				Email:         "user@example.com",
				MonthlyIncome: arg.MonthlyIncome,
				CreatedAt:     time.Now(),
			}, nil
		},
	}
	h := NewUserHandler(store, testConfig())

	body := `{"monthlyIncome":"3000.00"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.UpdateMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if receivedParams.PaymentDay != nil {
		t.Errorf("PaymentDay = %v, want nil", receivedParams.PaymentDay)
	}
}

func TestUserHandler_UpdateMe_DBError(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		updateUserSettingsFn: func(_ context.Context, _ db.UpdateUserSettingsParams) (db.User, error) {
			return db.User{}, errNotFound
		},
	}
	h := NewUserHandler(store, testConfig())

	body := `{"monthlyIncome":"5000.00"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader(body))
	req = requestWithUserID(req, 1)
	rr := httptest.NewRecorder()

	h.UpdateMe(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func makeNumericForHandler(s string) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(s)
	return n
}
