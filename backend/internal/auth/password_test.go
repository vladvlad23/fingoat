package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	t.Parallel()

	plain := "mysecretpassword"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}
	if hash == plain {
		t.Fatal("HashPassword() returned plaintext")
	}
	if !CheckPassword(plain, hash) {
		t.Error("CheckPassword() should return true for correct password")
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("correct")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	if CheckPassword("wrong", hash) {
		t.Error("CheckPassword() should return false for wrong password")
	}
}

func TestHashPassword_UniqueHashes(t *testing.T) {
	t.Parallel()

	h1, _ := HashPassword("same")
	h2, _ := HashPassword("same")
	if h1 == h2 {
		t.Error("HashPassword() should produce different hashes for same input (bcrypt salt)")
	}
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("notempty")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	if CheckPassword("", hash) {
		t.Error("CheckPassword() with empty password should return false")
	}
}
