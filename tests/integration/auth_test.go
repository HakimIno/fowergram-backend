package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper functions
func registerTestUser(t *testing.T, app *fiber.App, username, email string) *http.Response {
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(fmt.Sprintf(`{
		"username": "%s",
		"email": "%s",
		"password": "Test123!"
	}`, username, email)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	return resp
}

func TestAuthFlow(t *testing.T) {
	app := setupTestApp()
	cleanupTestDB(testDB)

	// Register a new user
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(`{
		"username": "testuser1",
		"email": "test1@example.com",
		"password": "Password123!"
	}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Login with the registered user
	req = httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{
		"email": "test1@example.com",
		"password": "Password123!"
	}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
}

func TestAuthFlow_InvalidCredentials(t *testing.T) {
	app := setupTestApp()
	cleanupTestDB(testDB)

	// Try to login with wrong credentials
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{
		"email": "wrong@example.com",
		"password": "WrongPassword123!"
	}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 401, resp.StatusCode)
}

func TestAuthFlow_TokenValidation(t *testing.T) {
	app := setupTestApp()
	cleanupTestDB(testDB)

	// Register a new user
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(`{
		"username": "testuser2",
		"email": "test2@example.com",
		"password": "Password123!"
	}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Login to get a token
	req = httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{
		"email": "test2@example.com",
		"password": "Password123!"
	}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Read token from response
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var loginResp map[string]interface{}
	err = json.Unmarshal(body, &loginResp)
	require.NoError(t, err)
	token := loginResp["token"].(string)

	// Validate token
	req = httptest.NewRequest("GET", "/api/v1/auth/validate", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
}

func TestAuthFlow_RateLimiting(t *testing.T) {
	app := setupTestApp()
	cleanupTestDB(testDB)

	// Try multiple failed login attempts
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{
			"email": "nonexistent@example.com",
			"password": "wrongpassword"
		}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)

		if i < 5 {
			assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		} else {
			assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
		}
	}
}

func TestAuthFlow_AccountLocking(t *testing.T) {
	app := setupTestApp()
	cleanupTestDB(testDB)

	// Register a test user first
	registerResp := registerTestUser(t, app, "testuser3", "test3@example.com")
	assert.Equal(t, fiber.StatusOK, registerResp.StatusCode)

	// Try multiple failed login attempts
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{
			"email": "test3@example.com",
			"password": "wrongpassword"
		}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	}

	// Next attempt should fail due to account being locked
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{
		"email": "test3@example.com",
		"password": "Test123!"
	}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Contains(t, result["error"], "Account is locked")
}
