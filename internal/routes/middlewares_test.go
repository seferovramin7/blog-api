package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationMiddleware(t *testing.T) {
	t.Run("Valid JSON Content-Type", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/posts", bytes.NewReader([]byte(`{"title":"test","content":"test content","author":"test author"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		validationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Invalid Content-Type", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/posts", bytes.NewReader([]byte(`{"title":"test"}`)))
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()

		validationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
	})
}

func TestLoggingMiddleware(t *testing.T) {
	t.Run("Log Request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/posts", nil)
		rec := httptest.NewRecorder()

		loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestErrorHandlingMiddleware(t *testing.T) {
	t.Run("Error Handling with Panic", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/posts", nil)
		rec := httptest.NewRecorder()

		errorHandlingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("unexpected error")
		})).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		var errorResponse JSONErrorResponse
		err := json.NewDecoder(rec.Body).Decode(&errorResponse)
		if err != nil {
			return
		}
		assert.Equal(t, "Internal Server Error", errorResponse.Error)
		assert.Contains(t, errorResponse.Description, "server error occurred")
	})
}

func TestCORSMiddleware(t *testing.T) {
	allowedOrigins := []string{"http://example.com", "http://localhost"}
	allowedMethods := []string{"GET", "POST", "PUT", "DELETE"}
	allowedHeaders := []string{"Content-Type", "Authorization"}

	t.Run("Valid CORS Preflight Request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/v1/posts", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()

		corsMiddleware(allowedOrigins, allowedMethods, allowedHeaders)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "http://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, strings.Join(allowedMethods, ", "), rec.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, strings.Join(allowedHeaders, ", "), rec.Header().Get("Access-Control-Allow-Headers"))
	})

	t.Run("Invalid Origin", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/v1/posts", nil)
		req.Header.Set("Origin", "http://unauthorized.com")
		rec := httptest.NewRecorder()

		corsMiddleware(allowedOrigins, allowedMethods, allowedHeaders)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		var errorResponse JSONErrorResponse
		err := json.NewDecoder(rec.Body).Decode(&errorResponse)
		if err != nil {
			return
		}
		assert.Equal(t, "Forbidden", errorResponse.Error)
		assert.Contains(t, errorResponse.Description, "Origin not allowed")
	})

	t.Run("No Origin in Request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/posts", nil)
		rec := httptest.NewRecorder()

		corsMiddleware(allowedOrigins, allowedMethods, allowedHeaders)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}
