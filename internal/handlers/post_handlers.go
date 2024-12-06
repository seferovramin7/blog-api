package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"blog-api/internal/models"
	"github.com/gorilla/mux"
)

type PostService interface {
	GetAllPosts(ctx context.Context, page, limit int) ([]*models.Post, error)
	GetPostByID(ctx context.Context, id string) (*models.Post, error)
	CreatePost(ctx context.Context, post *models.Post) (*models.Post, error)
	UpdatePost(ctx context.Context, id string, post *models.Post) (*models.Post, error)
	DeletePost(ctx context.Context, id string) error
}

type PostHandlerInterface interface {
	GetAllPosts(w http.ResponseWriter, r *http.Request)
	GetPostByID(w http.ResponseWriter, r *http.Request)
	CreatePost(w http.ResponseWriter, r *http.Request)
	UpdatePost(w http.ResponseWriter, r *http.Request)
	PatchPost(w http.ResponseWriter, r *http.Request)
	DeletePost(w http.ResponseWriter, r *http.Request)
}

var _ PostHandlerInterface = (*PostHandler)(nil)

type PostHandler struct {
	service PostService
}

func NewPostHandler(service PostService) *PostHandler {
	return &PostHandler{service: service}
}

func parseID(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		return "", errors.New("id is empty")
	}
	return id, nil
}

func writeJSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	}
}

func handleError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := map[string]string{
		"error":       http.StatusText(status),
		"description": err.Error(),
	}
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		log.Printf("Failed to encode error response: %v", encodeErr)
	}
}

func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	pageStr := query.Get("page")
	limitStr := query.Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	} else if _, err := strconv.ParseFloat(pageStr, 64); err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	} else if _, err := strconv.ParseFloat(limitStr, 64); err != nil {
		limit = 10
	}

	posts, err := h.service.GetAllPosts(ctx, page, limit)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		handleError(w, errors.New("failed to fetch posts"), http.StatusInternalServerError)
		return
	}

	if posts == nil {
		posts = []*models.Post{}
	}

	writeJSONResponse(w, posts, http.StatusOK)
}

func (h *PostHandler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := parseID(r)
	if err != nil {
		handleError(w, errors.New("invalid ID"), http.StatusBadRequest)
		return
	}

	post, err := h.service.GetPostByID(ctx, id)
	if err != nil {
		handleError(w, errors.New("post not found"), http.StatusNotFound)
		return
	}
	writeJSONResponse(w, post, http.StatusOK)
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		handleError(w, errors.New("Content-Type must be application/json"), http.StatusBadRequest)
		return
	}

	createdPost, err := h.service.CreatePost(ctx, &post)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, createdPost, http.StatusCreated)
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := parseID(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		handleError(w, errors.New("Content-Type must be application/json"), http.StatusBadRequest)
		return
	}

	updatedPost, err := h.service.UpdatePost(ctx, id, &post)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, updatedPost, http.StatusOK)
}

func (h *PostHandler) PatchPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := parseID(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		handleError(w, errors.New("Content-Type must be application/json"), http.StatusBadRequest)
		return
	}

	post, err := h.service.GetPostByID(ctx, id)
	if err != nil {
		handleError(w, errors.New("post not found"), http.StatusNotFound)
		return
	}

	if title, ok := updates["title"].(string); ok {
		if title == "" {
			handleError(w, errors.New("title cannot be empty"), http.StatusBadRequest)
			return
		}
		post.Title = title
	}
	if content, ok := updates["content"].(string); ok {
		post.Content = content
	}
	if author, ok := updates["author"].(string); ok {
		post.Author = author
	}

	updatedPost, err := h.service.UpdatePost(ctx, id, post)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	writeJSONResponse(w, updatedPost, http.StatusOK)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := parseID(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.service.DeletePost(ctx, id); err != nil {
		handleError(w, errors.New("post not found"), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
