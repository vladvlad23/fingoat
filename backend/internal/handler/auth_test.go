package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fingoat/api/internal/auth"
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

// okRefreshTokenFn is a createRefreshTokenFn that always succeeds.
func okRefreshTokenFn(_ context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error) {
	return db.RefreshToken{
		ID:        1,
		UserID:    arg.UserID,
		TokenHash: arg.TokenHash,
		ExpiresAt: arg.ExpiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// validRefreshCookieFor generates a real raw+hash pair and returns the raw cookie
// value along with a getRefreshTokenByHashFn that returns a live (non-expired) token.
func validRefreshCookieFor(userID int64) (rawToken string, hashFn func(context.Context, string) (db.RefreshToken, error)) {
	raw, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		panic(err)
	}
	rt := db.RefreshToken{
		ID:        42,
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(auth.RefreshTokenTTL),
		CreatedAt: time.Now(),
	}
	return raw, func(_ context.Context, h string) (db.RefreshToken, error) {
		if h == hash {
			return rt, nil
		}
		return db.RefreshToken{}, errNotFound
	}
}

// ---- Register ----

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
		createRefreshTokenFn: okRefreshTokenFn,
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

func TestAuthHandler_Register_SetsRefreshCookie(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		createUserFn: func(_ context.Context, arg db.CreateUserParams) (db.User, error) {
			return db.User{ID: 1, Email: arg.Email, CreatedAt: time.Now()}, nil
		},
		createRefreshTokenFn: okRefreshTokenFn,
	}
	h := NewAuthHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		strings.NewReader(`{"email":"a@b.com","password":"pass"}`))
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rr.Code)
	}

	var found bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == refreshCookieName {
			found = true
			if c.Value == "" {
				t.Error("refresh_token cookie value should not be empty")
			}
			if !c.HttpOnly {
				t.Error("refresh_token cookie should be HttpOnly")
			}
			if c.MaxAge != refreshTokenMaxAge {
				t.Errorf("MaxAge = %d, want %d", c.MaxAge, refreshTokenMaxAge)
			}
		}
	}
	if !found {
		t.Error("expected refresh_token cookie in response")
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

// ---- Login ----

func TestAuthHandler_Login_Success(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	store := &mockStore{
		getUserByEmailFn: func(_ context.Context, email string) (db.User, error) {
			return db.User{
				ID:           1,
				Email:        email,
				PasswordHash: hashForTest(t, "password123"),
				CreatedAt:    time.Now(),
			}, nil
		},
		createRefreshTokenFn: okRefreshTokenFn,
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

func TestAuthHandler_Login_SetsRefreshCookie(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getUserByEmailFn: func(_ context.Context, email string) (db.User, error) {
			return db.User{
				ID:           1,
				Email:        email,
				PasswordHash: hashForTest(t, "pass"),
				CreatedAt:    time.Now(),
			}, nil
		},
		createRefreshTokenFn: okRefreshTokenFn,
	}
	h := NewAuthHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"email":"a@b.com","password":"pass"}`))
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var found bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == refreshCookieName {
			found = true
			if c.Value == "" {
				t.Error("refresh_token cookie value should not be empty")
			}
			if !c.HttpOnly {
				t.Error("refresh_token cookie should be HttpOnly")
			}
		}
	}
	if !found {
		t.Error("expected refresh_token cookie in response")
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

// ---- Refresh ----

func TestAuthHandler_Refresh_ValidToken(t *testing.T) {
	t.Parallel()

	rawToken, lookupFn := validRefreshCookieFor(7)
	store := &mockStore{
		getRefreshTokenByHashFn: lookupFn,
		getUserByIDFn: func(_ context.Context, id int64) (db.User, error) {
			return db.User{ID: id, Email: "u@example.com"}, nil
		},
		deleteRefreshTokenFn: func(_ context.Context, _ int64) error { return nil },
		createRefreshTokenFn: okRefreshTokenFn,
	}
	h := NewAuthHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: rawToken})
	rr := httptest.NewRecorder()

	h.Refresh(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}

	var resp tokenResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty JWT in response")
	}

	// New refresh cookie should be set (rotation).
	var found bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == refreshCookieName && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected rotated refresh_token cookie in response")
	}
}

func TestAuthHandler_Refresh_MissingCookie(t *testing.T) {
	t.Parallel()

	h := NewAuthHandler(&mockStore{}, testConfig())
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	rr := httptest.NewRecorder()

	h.Refresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestAuthHandler_Refresh_TokenNotFound(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		getRefreshTokenByHashFn: func(_ context.Context, _ string) (db.RefreshToken, error) {
			return db.RefreshToken{}, errNotFound
		},
	}
	h := NewAuthHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "deadbeef"})
	rr := httptest.NewRecorder()

	h.Refresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestAuthHandler_Refresh_ExpiredToken(t *testing.T) {
	t.Parallel()

	raw, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	expiredRT := db.RefreshToken{
		ID:        99,
		UserID:    5,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // already expired
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	store := &mockStore{
		getRefreshTokenByHashFn: func(_ context.Context, h string) (db.RefreshToken, error) {
			if h == hash {
				return expiredRT, nil
			}
			return db.RefreshToken{}, errNotFound
		},
		deleteRefreshTokenFn: func(_ context.Context, _ int64) error { return nil },
	}
	h := NewAuthHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: raw})
	rr := httptest.NewRecorder()

	h.Refresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

// ---- Logout ----

func TestAuthHandler_Logout_WithValidCookie(t *testing.T) {
	t.Parallel()

	rawToken, lookupFn := validRefreshCookieFor(3)
	deleted := false
	store := &mockStore{
		getRefreshTokenByHashFn: lookupFn,
		deleteRefreshTokenFn: func(_ context.Context, _ int64) error {
			deleted = true
			return nil
		},
	}
	h := NewAuthHandler(store, testConfig())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: rawToken})
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rr.Code)
	}
	if !deleted {
		t.Error("expected DeleteRefreshToken to be called")
	}

	// Cookie should be cleared.
	var cleared bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == refreshCookieName && c.MaxAge == -1 {
			cleared = true
		}
	}
	if !cleared {
		t.Error("expected refresh_token cookie to be cleared (MaxAge=-1)")
	}
}

func TestAuthHandler_Logout_MissingCookie(t *testing.T) {
	t.Parallel()

	h := NewAuthHandler(&mockStore{}, testConfig())
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rr.Code)
	}
}
