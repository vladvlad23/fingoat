package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fingoat/api/internal/config"
	"github.com/fingoat/api/internal/db"
	"github.com/jackc/pgx/v5/pgconn"
)

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret: "test-secret-key-for-handlers",
		Port:      "8080",
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		createUserFn: func(_ context.Context, arg db.CreateUserParams) (db.User, error) {
			return db.User{
				ID:        1,
				Email:     arg.Email,
				CreatedAt: time.Now(),
			}, nil
		},
	}
	h := NewAuthHandler(store, testConfig())

	body := `{"email":"test@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusCreated)
	}

	var resp tokenResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token == "" {
		t.Error("token should not be empty")
	}
}

func TestAuthHandler_Register_MissingFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{"missing email", `{"password":"secret"}`},
		{"missing password", `{"email":"test@example.com"}`},
		{"empty email", `{"email":"","password":"secret"}`},
		{"empty password", `{"email":"test@example.com","password":""}`},
		{"both empty", `{"email":"","password":""}`},
	}

	h := NewAuthHandler(&mockStore{}, testConfig())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			h.Register(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	t.Parallel()

	h := NewAuthHandler(&mockStore{}, testConfig())
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader("not json"))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		createUserFn: func(_ context.Context, _ db.CreateUserParams) (db.User, error) {
			return db.User{}, &pgconn.PgError{Code: "23505"}
		},
	}
	h := NewAuthHandler(store, testConfig())

	body := `{"email":"taken@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusConflict)
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	t.Parallel()

	// Pre-hash a known password
	cfg := testConfig()
	store := &mockStore{
		getUserByEmailFn: func(_ context.Context, email string) (db.User, error) {
			// bcrypt hash of "password123"
			return db.User{
				ID:           1,
				Email:        email,
				PasswordHash: hashForTest(t, "password123"),
				CreatedAt:    time.Now(),
			}, nil
		},
	}
	h := NewAuthHandler(store, cfg)

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp tokenResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token == "" {
		t.Error("token should not be empty")
	}
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{
				ID:           1,
				PasswordHash: hashForTest(t, "correct-password"),
			}, nil
		},
	}
	h := NewAuthHandler(store, testConfig())

	body := `{"email":"test@example.com","password":"wrong-password"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, errNotFound
		},
	}
	h := NewAuthHandler(store, testConfig())

	body := `{"email":"noone@example.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	t.Parallel()

	h := NewAuthHandler(&mockStore{}, testConfig())
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("{invalid"))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
