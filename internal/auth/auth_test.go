package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

const secretKey = "somesecretkey"

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, secretKey, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	validatedID, err := ValidateJWT(token, secretKey)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if validatedID != userID {
		t.Fatalf("Expected userID %v, got %v", userID, validatedID)
	}
}

func TestExpiredJWT(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, secretKey, -1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create expired JWT: %v", err)
	}

	_, err = ValidateJWT(token, secretKey)
	if err == nil {
		t.Fatalf("Expected error for expired JWT, but got none")
	}
}

func TestInvalidJWTSecret(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, secretKey, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create  JWT: %v", err)
	}

	_, err = ValidateJWT(token, "wrongsecretkey")
	if err == nil {
		t.Fatalf("Expected error for wrong secret, but got none")
	}
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer testtoken")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err != nil {
		t.Fatalf("Expected token 'testtoken', got '%s'", token)
	}
}

func TestGetBearerToken_MissingHeader(t *testing.T) {
	headers := http.Header{}

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatalf("Expected error for missing Authorization header, got none")
	}
}

func TestGetBearerToken_IndalidFormat(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "InvalidTokenFormat")

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatalf("Expected error for invalid Authorization header format, got none")
	}
}
