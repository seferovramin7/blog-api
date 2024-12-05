package models

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Post struct {
	ID      string `json:"id" dynamodbav:"ID"` // DynamoDB primary key
	Title   string `json:"title" dynamodbav:"Title" validate:"required,min=3"`
	Content string `json:"content" dynamodbav:"Content" validate:"required"`
	Author  string `json:"author" dynamodbav:"Author" validate:"required"`
}

func (p *Post) Validate() error {
	err := validate.Struct(p)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := make([]string, 0, len(validationErrors))
		for _, ve := range validationErrors {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed validation: %s", ve.Field(), ve.ActualTag()))
		}
		return errors.New("validation failed: " + fmt.Sprintf("%v", errorMessages))
	}

	//if strings.TrimSpace(p.Title) == "" {
	//	return custom_errors.New("validation failed: Field 'Title' must not be empty or whitespace only")
	//}
	//if strings.TrimSpace(p.Content) == "" {
	//	return custom_errors.New("validation failed: Field 'Content' must not be empty or whitespace only")
	//}
	//if strings.TrimSpace(p.Author) == "" {
	//	return custom_errors.New("validation failed: Field 'Author' must not be empty or whitespace only")
	//}

	return nil
}

func NewPost(title, content, author string) *Post {
	return &Post{
		Title:   title,
		Content: content,
		Author:  author,
	}
}
