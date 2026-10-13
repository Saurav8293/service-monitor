package auth_test

import (
	"service-monitor/internal/monitor/auth"
	"testing"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret-key-1234567890"
	tm := auth.NewTokenManager(secret)

	token, err := tm.GenerateToken(42, "developer")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token string")
	}

	claims, err := tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != 42 {
		t.Errorf("expected userID 42, got %d", claims.UserID)
	}
	if claims.Username != "developer" {
		t.Errorf("expected username 'developer', got %s", claims.Username)
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	tm1 := auth.NewTokenManager("secret-one")
	tm2 := auth.NewTokenManager("secret-two")

	token, err := tm1.GenerateToken(1, "alice")
	if err != nil {
		t.Fatalf("generate token error: %v", err)
	}

	_, err = tm2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected validation error with mismatched secret, got nil")
	}
}

func TestPasswordHashing(t *testing.T) {
	password := "my-secret-pass-2026"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password error: %v", err)
	}

	if !auth.CheckPasswordHash(password, hash) {
		t.Fatal("expected password to match hash")
	}

	if auth.CheckPasswordHash("wrong-password", hash) {
		t.Fatal("expected wrong password to fail check")
	}
}
