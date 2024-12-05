package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-api/internal/handlers"
	"blog-api/internal/models"
	"blog-api/internal/repository"
	"blog-api/internal/routes"
	"blog-api/internal/services"
	"github.com/stretchr/testify/assert"
)

func setupTestServer() *httptest.Server {
	repo := repository.NewPostRepository()
	service := services.NewPostService(repo)
	handler := handlers.NewPostHandler(service)

	router := routes.SetupRouter(handler)
	return httptest.NewServer(router)
}

func sendRequest(t *testing.T, client *http.Client, method, url string, body interface{}, headers map[string]string) *http.Response {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		assert.NoError(t, err, "Failed to marshal request body")
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	assert.NoError(t, err, "Failed to create request")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	assert.NoError(t, err, "Failed to send request")
	return resp
}

func TestIntegrationAPI(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := &http.Client{}
	baseURL := server.URL + "/v1/posts"

	t.Run("CreatePost - Success", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":   "Integration Test Post",
			"content": "This is a test post",
			"author":  "Tester",
		}

		resp := sendRequest(t, client, http.MethodPost, baseURL, payload, map[string]string{"Content-Type": "application/json"})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createdPost models.Post
		err := json.NewDecoder(resp.Body).Decode(&createdPost)
		if err != nil {
			return
		}
		assert.Equal(t, "Integration Test Post", createdPost.Title)
	})

	t.Run("GetAllPosts - Empty Query Params", func(t *testing.T) {
		resp := sendRequest(t, client, http.MethodGet, baseURL, nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var posts []models.Post
		err := json.NewDecoder(resp.Body).Decode(&posts)
		if err != nil {
			return
		}
		assert.NotEmpty(t, posts)
	})

	t.Run("GetPostByID - Invalid ID", func(t *testing.T) {
		resp := sendRequest(t, client, http.MethodGet, baseURL+"/abc", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var errorResponse map[string]string
		err := json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			return
		}
		assert.Equal(t, "", errorResponse["description"])
	})
}
