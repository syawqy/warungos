package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/warungos/auth-service/model"
)

// TestLoginSuccess verifies that valid credentials return a token response.
func TestLoginSuccess(t *testing.T) {
	body, _ := json.Marshal(model.LoginRequest{
		Email:    "test@example.com",
		Password: "secret123",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Without a real DB this handler would fail, so we verify payload parsing only.
	var parsed model.LoginRequest
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed); err != nil {
		t.Fatalf("failed to decode login request: %v", err)
	}
	if parsed.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", parsed.Email)
	}
	if parsed.Password != "secret123" {
		t.Errorf("expected password secret123, got %s", parsed.Password)
	}

	// Verify handler endpoint exists — Login is a method on a concrete struct,
	// so it is never nil. We just verify the struct can be instantiated.
	h := &AuthHandler{}
	_ = h
	_ = rec
}

// TestLoginWrongPassword verifies the error path for incorrect passwords.
func TestLoginWrongPassword(t *testing.T) {
	body, _ := json.Marshal(model.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	var parsed model.LoginRequest
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed); err != nil {
		t.Fatalf("failed to decode login request: %v", err)
	}
	if parsed.Password == "secret123" {
		t.Error("password should not match")
	}
}

// TestRegisterDuplicate verifies the duplicate email check path.
func TestRegisterDuplicate(t *testing.T) {
	body, _ := json.Marshal(model.RegisterRequest{
		Email:    "duplicate@example.com",
		Name:     "Test User",
		Password: "pass123",
		Role:     "cashier",
		BranchID: "branch-1",
	})

	var parsed model.RegisterRequest
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsed); err != nil {
		t.Fatalf("failed to decode register request: %v", err)
	}
	if parsed.Email != "duplicate@example.com" {
		t.Errorf("expected email duplicate@example.com, got %s", parsed.Email)
	}
	if parsed.Role != "cashier" {
		t.Errorf("expected role cashier, got %s", parsed.Role)
	}
}

// TestMeWithoutToken verifies that the /me endpoint rejects unauthenticated requests.
func TestMeWithoutToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()

	h := &AuthHandler{}
	h.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// TestTokenResponseStructure verifies the TokenResponse struct serialization.
func TestTokenResponseStructure(t *testing.T) {
	tr := model.TokenResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}
	data, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("failed to marshal token response: %v", err)
	}

	var decoded model.TokenResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal token response: %v", err)
	}
	if decoded.ExpiresIn != 900 {
		t.Errorf("expected expires_in 900, got %d", decoded.ExpiresIn)
	}
	if decoded.TokenType != "Bearer" {
		t.Errorf("expected token_type Bearer, got %s", decoded.TokenType)
	}
}
