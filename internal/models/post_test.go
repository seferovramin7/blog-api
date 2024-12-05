package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostValidation(t *testing.T) {
	tests := []struct {
		name    string
		post    *Post
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Post",
			post: &Post{
				Title:   "A Valid Title",
				Content: "This is some valid content.",
				Author:  "John Doe",
			},
			wantErr: false,
		},
		{
			name: "Minimal Valid Title Length",
			post: &Post{
				Title:   "ABC",
				Content: "Valid Content",
				Author:  "Author Name",
			},
			wantErr: false,
		},

		{
			name: "Missing All Fields",
			post: &Post{
				Title:   "",
				Content: "",
				Author:  "",
			},
			wantErr: true,
			errMsg:  "Field 'Title' failed validation: required",
		},
		{
			name: "Title Too Short",
			post: &Post{
				Title:   "Hi",
				Content: "Valid Content",
				Author:  "Author Name",
			},
			wantErr: true,
			errMsg:  "Field 'Title' failed validation: min",
		},
		{
			name: "Whitespace Title",
			post: &Post{
				Title:   "   ",
				Content: "Valid Content",
				Author:  "Author Name",
			},
			wantErr: true,
			errMsg:  "Field 'Title' must not be empty or whitespace only",
		},
		{
			name: "Whitespace Author",
			post: &Post{
				Title:   "Valid Title",
				Content: "Valid Content",
				Author:  "    ",
			},
			wantErr: true,
			errMsg:  "Field 'Author' must not be empty or whitespace only",
		},
		{
			name: "Long Title",
			post: &Post{
				Title:   strings.Repeat("A", 1001),
				Content: "Valid Content",
				Author:  "Author Name",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.post.Validate()

			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got nil")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should match expected substring")
				}
			} else {
				assert.NoError(t, err, "Did not expect an error but got one")
			}
		})
	}
}

func TestNewPost(t *testing.T) {
	t.Run("Valid New Post Creation", func(t *testing.T) {
		post := NewPost("Sample Title", "Sample Content", "Author Name")

		assert.Equal(t, "Sample Title", post.Title, "Title should match")
		assert.Equal(t, "Sample Content", post.Content, "Content should match")
		assert.Equal(t, "Author Name", post.Author, "Author should match")
		assert.Equal(t, 0, post.ID, "ID should be zero for a new post")
	})

	t.Run("Empty New Post", func(t *testing.T) {
		post := NewPost("", "", "")

		assert.Equal(t, "", post.Title, "Title should be empty")
		assert.Equal(t, "", post.Content, "Content should be empty")
		assert.Equal(t, "", post.Author, "Author should be empty")
	})
}
