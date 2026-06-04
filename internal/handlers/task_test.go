package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateTask_Validation(t *testing.T) {
	// Set Gin to test mode to suppress debug logs in the console
	gin.SetMode(gin.TestMode)

	// We don't need a real DB or channel to test JSON payload validation (400 errors).
	// Therefore, we pass nil for the database connection and the task channel.
	apiDeps := &API{
		DB:       nil,
		TaskChan: nil,
	}

	// Configure the test router
	router := gin.Default()
	router.POST("/api/v1/tasks", apiDeps.CreateTask)

	// Table of negative test cases
	tests := []struct {
		name         string
		payload      string
		expectedCode int
	}{
		{
			name:         "Empty city",
			payload:      `{"city": "", "max_price": 600, "email": "test@mail.com"}`,
			expectedCode: http.StatusBadRequest, // Expecting 400
		},
		{
			name:         "Negative price",
			payload:      `{"city": "Neuss", "max_price": -100, "email": "test@mail.com"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Invalid email format",
			payload:      `{"city": "Neuss", "max_price": 600, "email": "not-an-email"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty JSON payload",
			payload:      `{}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Create a fake HTTP request with the JSON payload
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(tc.payload))
			req.Header.Set("Content-Type", "application/json")

			// 2. Create a response recorder to capture the server's response
			w := httptest.NewRecorder()

			// 3. Pass the request to the router (simulate a real network call)
			router.ServeHTTP(w, req)

			// 4. Verify that the API rejected the request correctly
			if w.Code != tc.expectedCode {
				t.Errorf("Case '%s': expected status %d, got %d. Body: %s", tc.name, tc.expectedCode, w.Code, w.Body.String())
			}
		})
	}
}
