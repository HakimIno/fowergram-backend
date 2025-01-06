package integration

import (
	"encoding/json"
	"fowergram/tests/helpers"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoints(t *testing.T) {
	app := helpers.SetupTestApp()

	t.Run("Ping Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, "ok", result["status"])
		assert.NotNil(t, result["time"])
	})

	t.Run("Health Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, "ok", result["status"])
		assert.NotNil(t, result["time"])

		services, ok := result["services"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "up", services["api"])
		assert.Equal(t, "up", services["db"])
		assert.Equal(t, "up", services["redis"])
	})
}
