package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPostService struct {
	mock.Mock
}

func (m *MockPostService) GetAllPosts(page, limit int) ([]*models.Post, error) {
	args := m.Called(page, limit)
	var posts []*models.Post
	if args.Get(0) != nil {
		posts = args.Get(0).([]*models.Post)
	}
	return posts, args.Error(1)
}

func (m *MockPostService) GetPostByID(id int) (*models.Post, error) {
	args := m.Called(id)
	var post *models.Post
	if args.Get(0) != nil {
		post = args.Get(0).(*models.Post)
	}
	return post, args.Error(1)
}

func (m *MockPostService) CreatePost(post *models.Post) (*models.Post, error) {
	args := m.Called(post)
	var createdPost *models.Post
	if args.Get(0) != nil {
		createdPost = args.Get(0).(*models.Post)
	}
	return createdPost, args.Error(1)
}

func (m *MockPostService) UpdatePost(id int, post *models.Post) (*models.Post, error) {
	args := m.Called(id, post)
	var updatedPost *models.Post
	if args.Get(0) != nil {
		updatedPost = args.Get(0).(*models.Post)
	}
	return updatedPost, args.Error(1)
}

func (m *MockPostService) DeletePost(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestPostHandlers(t *testing.T) {
	t.Run("GetAllPosts - Success", func(t *testing.T) {
		mockService := new(MockPostService)
		handler := NewPostHandler(mockService)

		posts := []*models.Post{
			{ID: 1, Title: "Post 1", Content: "Content 1", Author: "Author 1"},
			{ID: 2, Title: "Post 2", Content: "Content 2", Author: "Author 2"},
		}
		mockService.On("GetAllPosts", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(posts, nil)

		req := httptest.NewRequest("GET", "/posts", nil)
		rec := httptest.NewRecorder()

		handler.GetAllPosts(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var gotPosts []*models.Post
		err := json.Unmarshal(rec.Body.Bytes(), &gotPosts)
		assert.NoError(t, err)
		assert.Equal(t, posts, gotPosts)

		mockService.AssertExpectations(t)
	})

	t.Run("GetAllPosts - Service Error", func(t *testing.T) {
		mockService := new(MockPostService)
		handler := NewPostHandler(mockService)

		mockService.On("GetAllPosts", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("internal server error"))

		req := httptest.NewRequest("GET", "/posts", nil)
		rec := httptest.NewRecorder()

		handler.GetAllPosts(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code, "Expected 500 Internal Server Error")

		var errResponse map[string]string
		err := json.Unmarshal(rec.Body.Bytes(), &errResponse)
		assert.NoError(t, err, "Expected a valid JSON error response")
		assert.Equal(t, "Internal Server Error", errResponse["error"], "Expected error field to be 'Internal Server Error'")
		assert.Equal(t, "failed to fetch posts", errResponse["description"], "Expected description field to match the error")

		mockService.AssertExpectations(t)
	})

	t.Run("GetPostByID - Success", func(t *testing.T) {
		mockService := new(MockPostService)
		handler := NewPostHandler(mockService)

		post := &models.Post{ID: 1, Title: "Post 1", Content: "Content 1", Author: "Author 1"}
		mockService.On("GetPostByID", 1).Return(post, nil)

		req := httptest.NewRequest("GET", "/posts/1", nil)
		req = muxSetVars(req, map[string]string{"id": "1"})
		rec := httptest.NewRecorder()

		handler.GetPostByID(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var gotPost models.Post
		err := json.Unmarshal(rec.Body.Bytes(), &gotPost)
		assert.NoError(t, err)
		assert.Equal(t, *post, gotPost)

		mockService.AssertExpectations(t)
	})

	t.Run("GetPostByID - Not Found", func(t *testing.T) {
		mockService := new(MockPostService)
		handler := NewPostHandler(mockService)

		mockService.On("GetPostByID", 1).Return(nil, errors.New("post not found"))

		req := httptest.NewRequest("GET", "/posts/1", nil)
		req = muxSetVars(req, map[string]string{"id": "1"})
		rec := httptest.NewRecorder()

		handler.GetPostByID(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var errResponse map[string]string
		err := json.Unmarshal(rec.Body.Bytes(), &errResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Not Found", errResponse["error"])
		assert.Equal(t, "post not found", errResponse["description"])

		mockService.AssertExpectations(t)
	})

	t.Run("CreatePost - Success", func(t *testing.T) {
		mockService := new(MockPostService)
		handler := NewPostHandler(mockService)

		post := &models.Post{Title: "New Post", Content: "New Content", Author: "Author"}
		createdPost := &models.Post{ID: 1, Title: "New Post", Content: "New Content", Author: "Author"}
		mockService.On("CreatePost", post).Return(createdPost, nil)

		body, _ := json.Marshal(post)
		req := httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.CreatePost(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var gotPost models.Post
		err := json.Unmarshal(rec.Body.Bytes(), &gotPost)
		assert.NoError(t, err)
		assert.Equal(t, *createdPost, gotPost)

		mockService.AssertExpectations(t)
	})

	t.Run("CreatePost - Invalid JSON", func(t *testing.T) {
		mockService := new(MockPostService)
		handler := NewPostHandler(mockService)

		req := httptest.NewRequest("POST", "/posts", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.CreatePost(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errResponse map[string]string
		err := json.Unmarshal(rec.Body.Bytes(), &errResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Bad Request", errResponse["error"])
		assert.Equal(t, "Content-Type must be application/json", errResponse["description"])
	})

	t.Run("UpdatePost - Success", func(t *testing.T) {
		mockService := new(MockPostService)
		handler := NewPostHandler(mockService)

		post := &models.Post{Title: "Updated Post", Content: "Updated Content", Author: "Author"}
		updatedPost := &models.Post{ID: 1, Title: "Updated Post", Content: "Updated Content", Author: "Author"}
		mockService.On("UpdatePost", 1, post).Return(updatedPost, nil)

		body, _ := json.Marshal(post)
		req := httptest.NewRequest("PUT", "/posts/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = muxSetVars(req, map[string]string{"id": "1"})
		rec := httptest.NewRecorder()

		handler.UpdatePost(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var gotPost models.Post
		err := json.Unmarshal(rec.Body.Bytes(), &gotPost)
		assert.NoError(t, err)
		assert.Equal(t, *updatedPost, gotPost)

		mockService.AssertExpectations(t)
	})

	t.Run("UpdatePost - Invalid ID", func(t *testing.T) {
		mockService := new(MockPostService)
		handler := NewPostHandler(mockService)

		req := httptest.NewRequest("PUT", "/posts/abc", nil)
		req = muxSetVars(req, map[string]string{"id": "abc"})
		rec := httptest.NewRecorder()

		handler.UpdatePost(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errResponse map[string]string
		err := json.Unmarshal(rec.Body.Bytes(), &errResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Bad Request", errResponse["error"])
		assert.Equal(t, "invalid ID", errResponse["description"])
	})

	t.Run("DeletePost - Success", func(t *testing.T) {
		mockService := new(MockPostService)
		handler := NewPostHandler(mockService)

		mockService.On("DeletePost", 1).Return(nil)

		req := httptest.NewRequest("DELETE", "/posts/1", nil)
		req = muxSetVars(req, map[string]string{"id": "1"})
		rec := httptest.NewRecorder()

		handler.DeletePost(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("DeletePost - Not Found", func(t *testing.T) {
		mockService := new(MockPostService)
		handler := NewPostHandler(mockService)

		mockService.On("DeletePost", 1).Return(errors.New("post not found"))

		req := httptest.NewRequest("DELETE", "/posts/1", nil)
		req = muxSetVars(req, map[string]string{"id": "1"})
		rec := httptest.NewRecorder()

		handler.DeletePost(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var errResponse map[string]string
		err := json.Unmarshal(rec.Body.Bytes(), &errResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Not Found", errResponse["error"])
		assert.Equal(t, "post not found", errResponse["description"])

		mockService.AssertExpectations(t)
	})
}

func muxSetVars(r *http.Request, vars map[string]string) *http.Request {
	return mux.SetURLVars(r, vars)
}
