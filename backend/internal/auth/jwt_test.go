package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	t.Parallel()

	secret := "test-secret-key"
	userID := int64(42)

	token, err := GenerateToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken() returned empty token")
	}

	got, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken() error: %v", err)
	}
	if got != userID {
		t.Errorf("ValidateToken() = %d, want %d", got, userID)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	t.Parallel()

	token, err := GenerateToken(1, "correct-secret")
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	_, err = ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Error("ValidateToken() with wrong secret should return error")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	t.Parallel()

	secret := "test-secret"
	claims := Claims{
		UserID: 99,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("signing expired token: %v", err)
	}

	_, err = ValidateToken(tokenStr, secret)
	if err == nil {
		t.Error("ValidateToken() with expired token should return error")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"garbage", "not-a-jwt-token"},
		{"partial jwt", "eyJhbGciOiJIUzI1NiJ9."},
		{"three dots no content", "a.b.c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ValidateToken(tt.token, "secret")
			if err == nil {
				t.Errorf("ValidateToken(%q) should return error", tt.token)
			}
		})
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	t.Parallel()

	// Sign with a non-HMAC method (none)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, &Claims{
		UserID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("signing with none method: %v", err)
	}

	_, err = ValidateToken(tokenStr, "secret")
	if err == nil {
		t.Error("ValidateToken() with 'none' signing method should return error")
	}
}

func TestGenerateToken_DifferentUserIDs(t *testing.T) {
	t.Parallel()

	secret := "test-secret"
	for _, userID := range []int64{0, 1, -1, 999999} {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			token, err := GenerateToken(userID, secret)
			if err != nil {
				t.Fatalf("GenerateToken(%d) error: %v", userID, err)
			}
			got, err := ValidateToken(token, secret)
			if err != nil {
				t.Fatalf("ValidateToken() error: %v", err)
			}
			if got != userID {
				t.Errorf("round-trip userID = %d, want %d", got, userID)
			}
		})
	}
}
