package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fingoat/api/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-jwt-secret"

func TestAuth_ValidToken(t *testing.T) {
	t.Parallel()

	userID := int64(42)
	token, err := auth.GenerateToken(userID, testSecret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	var gotUserID int64
	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := GetUserID(r.Context())
		if err != nil {
			t.Errorf("GetUserID() error: %v", err)
			return
		}
		gotUserID = id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if gotUserID != userID {
		t.Errorf("userID = %d, want %d", gotUserID, userID)
	}
}

func TestAuth_MissingAuthorizationHeader(t *testing.T) {
	t.Parallel()

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuth_MissingBearerPrefix(t *testing.T) {
	t.Parallel()

	token, _ := auth.GenerateToken(1, testSecret, time.Hour)

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", token) // no "Bearer " prefix
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	t.Parallel()

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	t.Parallel()

	claims := auth.Claims{
		UserID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(testSecret))

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestGetUserID_NoValueInContext(t *testing.T) {
	t.Parallel()

	_, err := GetUserID(context.Background())
	if err == nil {
		t.Error("GetUserID() with empty context should return error")
	}
}

func TestGetUserID_WrongType(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), userIDKey, "not-an-int64")
	_, err := GetUserID(ctx)
	if err == nil {
		t.Error("GetUserID() with wrong type should return error")
	}
}
