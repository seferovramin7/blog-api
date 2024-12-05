package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-api/internal/handlers"
	"blog-api/internal/repository"
	"blog-api/internal/routes"
	"blog-api/internal/services"
	"github.com/stretchr/testify/assert"
)

func setupFuzzTestServer() *httptest.Server {
	repo := repository.NewPostRepository()
	service := services.NewPostService(repo)
	handler := handlers.NewPostHandler(service)

	router := routes.SetupRouter(handler)
	return httptest.NewServer(router)
}

func sendFuzzRequest(t *testing.T, client *http.Client, method, url, payload string, headers map[string]string) *http.Response {
	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(payload)))
	assert.NoError(t, err)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	assert.NoError(t, err)

	return resp
}

func parseErrorResponse(t *testing.T, body *http.Response) map[string]string {
	var errorResponse map[string]string
	err := json.NewDecoder(body.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	return errorResponse
}

func TestFuzzingAPI(t *testing.T) {
	server := setupFuzzTestServer()
	defer server.Close()

	client := &http.Client{}
	baseURL := server.URL + "/v1/posts"

	fuzzCases := []struct {
		name           string
		payload        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid Title Type",
			payload:        `{"title":123,"content":"Test Content","author":"Tester"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Content-Type must be application/json",
		},
		{
			name:           "Empty JSON Body",
			payload:        `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title cannot be empty",
		},
		{
			name:           "Missing Content Field",
			payload:        `{"title":"Valid Title","author":"Tester"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "content cannot be empty",
		},
		{
			name:           "Malformed JSON",
			payload:        `{"title":`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Content-Type must be application/json",
		},
		{
			name:           "Extra Fields",
			payload:        `{"title":"Valid Title","content":"Valid Content","author":"Tester","extra":"field"}`,
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name:           "Non-JSON Content Type",
			payload:        `<xml><title>Test</title></xml>`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Content-Type must be application/json",
		},
	}

	for _, tc := range fuzzCases {
		t.Run(tc.name, func(t *testing.T) {
			headers := map[string]string{"Content-Type": "application/json"}
			if tc.name == "Non-JSON Content Type" {
				headers["Content-Type"] = "text/xml"
			}

			resp := sendFuzzRequest(t, client, http.MethodPost, baseURL, tc.payload, headers)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedError != "" {
				errorResponse := parseErrorResponse(t, resp)
				assert.Contains(t, errorResponse["description"], tc.expectedError)
			}
		})
	}

}
