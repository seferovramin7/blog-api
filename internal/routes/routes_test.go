package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPostHandler struct {
	mock.Mock
}

func (m *MockPostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("GetAllPosts"))
	if err != nil {
		return
	}
}

func (m *MockPostHandler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("GetPostByID"))
	if err != nil {
		return
	}
}

func (m *MockPostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusCreated)
	_, err := w.Write([]byte("CreatePost"))
	if err != nil {
		return
	}
}

func (m *MockPostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("UpdatePost"))
	if err != nil {
		return
	}
}

func (m *MockPostHandler) PatchPost(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("PatchPost"))
	if err != nil {
		return
	}
}

func (m *MockPostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func TestRoutes(t *testing.T) {
	mockHandler := new(MockPostHandler)
	router := SetupRouter(mockHandler)

	t.Run("Route GetAllPosts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/posts", nil)
		rec := httptest.NewRecorder()
		mockHandler.On("GetAllPosts", mock.Anything, mock.Anything).Return().Once()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "GetAllPosts", rec.Body.String())
		mockHandler.AssertExpectations(t)
	})

	t.Run("Route GetPostByID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/posts/1", nil)
		rec := httptest.NewRecorder()
		mockHandler.On("GetPostByID", mock.Anything, mock.Anything).Return().Once()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "GetPostByID", rec.Body.String())
		mockHandler.AssertExpectations(t)
	})

	t.Run("Route CreatePost", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/posts", nil)
		rec := httptest.NewRecorder()
		mockHandler.On("CreatePost", mock.Anything, mock.Anything).Return().Once()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "CreatePost", rec.Body.String())
		mockHandler.AssertExpectations(t)
	})

	t.Run("Route UpdatePost", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/v1/posts/1", nil)
		rec := httptest.NewRecorder()
		mockHandler.On("UpdatePost", mock.Anything, mock.Anything).Return().Once()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "UpdatePost", rec.Body.String())
		mockHandler.AssertExpectations(t)
	})

	t.Run("Route DeletePost", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/posts/1", nil)
		rec := httptest.NewRecorder()
		mockHandler.On("DeletePost", mock.Anything, mock.Anything).Return().Once()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockHandler.AssertExpectations(t)
	})
}
