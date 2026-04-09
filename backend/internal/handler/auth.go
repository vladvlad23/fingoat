package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/fingoat/api/internal/auth"
	"github.com/fingoat/api/internal/config"
	"github.com/fingoat/api/internal/db"
	"github.com/fingoat/api/internal/stores"
	"github.com/jackc/pgx/v5/pgconn"
)

const refreshCookieName = "refresh_token"

type ErrEmailTaken struct{}

func (e *ErrEmailTaken) Error() string { return "email already registered" }

func classifyCreateUserErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return &ErrEmailTaken{}
	}
	return err
}

type AuthHandler struct {
	store  stores.AuthStore
	config *config.Config
}

func NewAuthHandler(q stores.AuthStore, cfg *config.Config) *AuthHandler {
	return &AuthHandler{store: q, config: cfg}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

// issueRefreshCookie generates a refresh token, persists its hash, and sets the httpOnly cookie.
func (h *AuthHandler) issueRefreshCookie(w http.ResponseWriter, r *http.Request, userID int64) error {
	raw, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		return err
	}
	if _, err := h.store.CreateRefreshToken(r.Context(), db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(h.config.RefreshTokenTTL),
	}); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    raw,
		Path:     "/api/auth",
		MaxAge:   int(h.config.RefreshTokenTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if h.config.DisableRegister {
		writeError(w, http.StatusForbidden, "registration is disabled")
		return
	}
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	user, err := h.store.CreateUser(r.Context(), db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: hash,
	})
	if err != nil {
		var emailTaken *ErrEmailTaken
		if errors.As(classifyCreateUserErr(err), &emailTaken) {
			writeError(w, http.StatusConflict, emailTaken.Error())
			return
		}
		log.Printf("register: unexpected error: %v\n%s", err, debug.Stack())
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	token, err := auth.GenerateToken(user.ID, h.config.JWTSecret, h.config.JWTTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	if err := h.issueRefreshCookie(w, r, user.ID); err != nil {
		log.Printf("register: issue refresh token: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, tokenResponse{Token: token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := auth.GenerateToken(user.ID, h.config.JWTSecret, h.config.JWTTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	if err := h.issueRefreshCookie(w, r, user.ID); err != nil {
		log.Printf("login: issue refresh token: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, tokenResponse{Token: token})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	hash := auth.HashRefreshToken(cookie.Value)
	rt, err := h.store.GetRefreshTokenByHash(r.Context(), hash)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	if time.Now().After(rt.ExpiresAt) {
		// Expired — clean up the stale row and reject.
		if delErr := h.store.DeleteRefreshToken(r.Context(), rt.ID); delErr != nil {
			log.Printf("refresh: delete expired token id=%d: %v", rt.ID, delErr)
		}
		writeError(w, http.StatusUnauthorized, "refresh token expired")
		return
	}

	user, err := h.store.GetUserByID(r.Context(), rt.UserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	accessToken, err := auth.GenerateToken(user.ID, h.config.JWTSecret, h.config.JWTTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Rotate: delete old token first, then issue a new one.
	if err := h.store.DeleteRefreshToken(r.Context(), rt.ID); err != nil {
		log.Printf("refresh: delete old token id=%d: %v", rt.ID, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if err := h.issueRefreshCookie(w, r, user.ID); err != nil {
		log.Printf("refresh: issue new refresh token: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{Token: accessToken})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		// No cookie present — nothing to do.
		w.WriteHeader(http.StatusNoContent)
		return
	}

	hash := auth.HashRefreshToken(cookie.Value)
	rt, err := h.store.GetRefreshTokenByHash(r.Context(), hash)
	if err == nil {
		if delErr := h.store.DeleteRefreshToken(r.Context(), rt.ID); delErr != nil {
			log.Printf("logout: delete token id=%d: %v", rt.ID, delErr)
		}
	}

	// Clear the cookie regardless of whether we found a DB row.
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}
