package services

import (
	"blog-api/internal/handlers"
	"fmt"

	"blog-api/internal/models"
)

type Repository interface {
	GetAll(page, limit int) ([]*models.Post, error)
	GetByID(id int) (*models.Post, error)
	Create(post *models.Post) (*models.Post, error)
	Update(id int, updatedPost *models.Post) (*models.Post, error)
	Delete(id int) error
}

var _ handlers.PostService = (*PostService)(nil)

type PostService struct {
	repo Repository
}

func NewPostService(repo Repository) *PostService {
	return &PostService{repo: repo}
}

type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %d not found", e.Resource, e.ID)
}

func (s *PostService) GetAllPosts(page, limit int) ([]*models.Post, error) {
	return s.repo.GetAll(page, limit)
}

func (s *PostService) GetPostByID(id int) (*models.Post, error) {
	post, err := s.repo.GetByID(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "Post", ID: id}
	}
	return post, nil
}

func (s *PostService) CreatePost(post *models.Post) (*models.Post, error) {
	if err := post.Validate(); err != nil {
		return nil, err
	}

	createdPost, err := s.repo.Create(post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	return createdPost, nil
}

func (s *PostService) UpdatePost(id int, updatedPost *models.Post) (*models.Post, error) {
	if err := updatedPost.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetByID(id); err != nil {
		return nil, &NotFoundError{Resource: "Post", ID: id}
	}

	updated, err := s.repo.Update(id, updatedPost)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}
	return updated, nil
}

func (s *PostService) DeletePost(id int) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return &NotFoundError{Resource: "Post", ID: id}
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	return nil
}
