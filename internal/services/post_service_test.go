package services

import (
	"errors"
	"testing"

	"blog-api/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetAll(page, limit int) ([]*models.Post, error) {
	args := m.Called(page, limit)
	return args.Get(0).([]*models.Post), args.Error(1)
}

func (m *MockRepository) GetByID(id int) (*models.Post, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockRepository) Create(post *models.Post) (*models.Post, error) {
	args := m.Called(post)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockRepository) Update(id int, updatedPost *models.Post) (*models.Post, error) {
	args := m.Called(id, updatedPost)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestPostService(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewPostService(mockRepo)

	t.Run("CreatePost - Validation Error", func(t *testing.T) {
		invalidPost := &models.Post{Title: "", Content: "Content", Author: "Author"}
		post, err := service.CreatePost(invalidPost)
		assert.Nil(t, post, "Expected no post to be created")
		assert.Error(t, err, "Expected a validation error")
	})

	t.Run("CreatePost - Success", func(t *testing.T) {
		validPost := &models.Post{Title: "Title", Content: "Content", Author: "Author"}
		mockRepo.On("Create", validPost).Return(validPost, nil)

		post, err := service.CreatePost(validPost)
		assert.NoError(t, err, "Expected no error on CreatePost")
		assert.Equal(t, validPost, post, "Created post mismatch")

		mockRepo.AssertExpectations(t)
	})

	t.Run("GetPostByID - Not Found", func(t *testing.T) {
		mockRepo.On("GetByID", 99).Return(nil, errors.New("not found"))

		post, err := service.GetPostByID(99)
		assert.Nil(t, post, "Expected no post to be returned")
		assert.Error(t, err, "Expected an error on GetPostByID")
		assert.IsType(t, &NotFoundError{}, err, "Error type mismatch")

		mockRepo.AssertExpectations(t)
	})

	t.Run("GetPostByID - Success", func(t *testing.T) {
		expectedPost := &models.Post{ID: 1, Title: "Post Title", Content: "Content", Author: "Author"}
		mockRepo.On("GetByID", 1).Return(expectedPost, nil)

		post, err := service.GetPostByID(1)
		assert.NoError(t, err, "Expected no error on GetPostByID")
		assert.Equal(t, expectedPost, post, "Fetched post mismatch")

		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdatePost - Success", func(t *testing.T) {
		validPost := &models.Post{Title: "Updated", Content: "Updated Content", Author: "Author"}
		mockRepo.On("GetByID", 1).Return(validPost, nil)
		mockRepo.On("Update", 1, validPost).Return(validPost, nil)

		post, err := service.UpdatePost(1, validPost)
		assert.NoError(t, err, "Expected no error on UpdatePost")
		assert.Equal(t, validPost, post, "Updated post mismatch")

		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdatePost - Not Found", func(t *testing.T) {
		updatedPost := &models.Post{Title: "Updated", Content: "Updated Content", Author: "Author"}
		mockRepo.On("GetByID", 99).Return(nil, errors.New("not found"))

		post, err := service.UpdatePost(99, updatedPost)
		assert.Nil(t, post, "Expected no post to be updated")
		assert.Error(t, err, "Expected an error on UpdatePost")
		assert.IsType(t, &NotFoundError{}, err, "Error type mismatch")

		mockRepo.AssertExpectations(t)
	})

	t.Run("DeletePost - Success", func(t *testing.T) {
		mockRepo.On("GetByID", 1).Return(&models.Post{ID: 1}, nil)
		mockRepo.On("Delete", 1).Return(nil)

		err := service.DeletePost(1)
		assert.NoError(t, err, "Expected no error on DeletePost")

		mockRepo.AssertExpectations(t)
	})

	t.Run("DeletePost - Not Found", func(t *testing.T) {
		mockRepo.On("GetByID", 99).Return(nil, errors.New("not found"))

		err := service.DeletePost(99)
		assert.Error(t, err, "Expected an error on DeletePost")
		assert.IsType(t, &NotFoundError{}, err, "Error type mismatch")

		mockRepo.AssertExpectations(t)
	})
}
