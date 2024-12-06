package services

import (
	"blog-api/internal/handlers"
	"blog-api/internal/models"
	"context"
	"errors"
	"fmt"
)

type Repository interface {
	GetAll(ctx context.Context, page, limit int) ([]*models.Post, error)
	GetByID(ctx context.Context, id string) (*models.Post, error)
	Create(ctx context.Context, post *models.Post) (*models.Post, error)
	Update(ctx context.Context, id string, updatedPost *models.Post) (*models.Post, error)
	Delete(ctx context.Context, id string) error
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
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}

func (s *PostService) GetAllPosts(ctx context.Context, page, limit int) ([]*models.Post, error) {
	posts, err := s.repo.GetAll(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get all posts: %w", err)
	}
	return posts, nil
}

func (s *PostService) GetPostByID(ctx context.Context, id string) (*models.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		// Check if error is a not found scenario. If repository returns a generic error,
		// you may need a custom error check here.
		return nil, &NotFoundError{Resource: "Post", ID: id}
	}
	return post, nil
}

func (s *PostService) CreatePost(ctx context.Context, post *models.Post) (*models.Post, error) {
	if err := post.Validate(); err != nil {
		return nil, fmt.Errorf("post validation failed: %w", err)
	}

	createdPost, err := s.repo.Create(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	return createdPost, nil
}

func (s *PostService) UpdatePost(ctx context.Context, id string, updatedPost *models.Post) (*models.Post, error) {
	if err := updatedPost.Validate(); err != nil {
		return nil, fmt.Errorf("updated post validation failed: %w", err)
	}

	// Check if the post exists before attempting the update
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return nil, &NotFoundError{Resource: "Post", ID: id}
	}

	updated, err := s.repo.Update(ctx, id, updatedPost)
	if err != nil {
		return nil, fmt.Errorf("failed to update post with ID=%s: %w", id, err)
	}
	return updated, nil
}

func (s *PostService) DeletePost(ctx context.Context, id string) error {
	// Check if the post exists before deleting
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return &NotFoundError{Resource: "Post", ID: id}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete post with ID=%s: %w", id, err)
	}
	return nil
}

// If needed, you can implement an IsNotFound function to differentiate between not found and other errors
func IsNotFound(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}
