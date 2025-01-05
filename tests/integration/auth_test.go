package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fowergram/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestAuthFlow(t *testing.T) {
	app := setupTestApp()

	// Test Registration
	registerReq := domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "Test123!",
	}
	reqBody, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test Login
	loginReq := domain.LoginRequest{
		Email:    "test@example.com",
		Password: "Test123!",
	}
	reqBody, _ = json.Marshal(loginReq)
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp domain.AuthResponse
	json.NewDecoder(resp.Body).Decode(&loginResp)
	assert.NotEmpty(t, loginResp.Token)
}

func TestAuthFlow_InvalidCredentials(t *testing.T) {
	app := setupTestApp()

	loginReq := domain.LoginRequest{
		Email:    "wrong@example.com",
		Password: "wrongpass",
	}
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthFlow_TokenValidation(t *testing.T) {
	app := setupTestApp()

	// Register and login to get token
	registerAndLogin := func() string {
		registerReq := domain.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "Test123!",
		}
		reqBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		loginReq := domain.LoginRequest{
			Email:    "test@example.com",
			Password: "Test123!",
		}
		reqBody, _ = json.Marshal(loginReq)
		req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, _ = app.Test(req, -1)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var loginResp domain.AuthResponse
		json.NewDecoder(resp.Body).Decode(&loginResp)
		return loginResp.Token
	}

	t.Run("valid_token", func(t *testing.T) {
		token := registerAndLogin()
		req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := app.Test(req, -1)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("invalid_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")
		resp, _ := app.Test(req, -1)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("missing_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
		resp, _ := app.Test(req, -1)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAuthFlow_RateLimiting(t *testing.T) {
	app := setupTestApp()

	// Try multiple rapid login attempts
	for i := 0; i < 35; i++ {
		loginReq := domain.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpass",
		}
		reqBody, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)

		if i >= 30 {
			assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		}
	}
}

func TestAuthFlow_AccountLocking(t *testing.T) {
	app := setupTestApp()

	// Register a user first
	registerReq := domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "Test123!",
	}
	reqBody, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Try multiple failed login attempts
	for i := 0; i < 6; i++ {
		loginReq := domain.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpass",
		}
		reqBody, _ = json.Marshal(loginReq)
		req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, _ = app.Test(req, -1)

		if i >= 5 {
			// Account should be locked after 5 failed attempts
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			assert.Contains(t, response["error"], "locked")
		}
	}

	// Try logging in with correct password while account is locked
	loginReq := domain.LoginRequest{
		Email:    "test@example.com",
		Password: "Test123!",
	}
	reqBody, _ = json.Marshal(loginReq)
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req, -1)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
