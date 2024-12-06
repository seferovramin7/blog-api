package repository

import (
	"blog-api/internal/models"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type DynamoPostRepository struct {
	Client    *dynamodb.Client
	TableName string
}

func NewDynamoPostRepository(client *dynamodb.Client, tableName string) *DynamoPostRepository {
	return &DynamoPostRepository{
		Client:    client,
		TableName: tableName,
	}
}

func (r *DynamoPostRepository) GetAll(page, limit int) ([]*models.Post, error) {
	// Validate inputs
	if page <= 0 || limit <= 0 {
		return nil, fmt.Errorf("invalid page (%d) or limit (%d)", page, limit)
	}

	// Calculate how many items to skip for the requested page
	itemsToSkip := (page - 1) * limit

	var posts []*models.Post
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		// Set up scan input
		input := &dynamodb.ScanInput{
			TableName:            &r.TableName,
			Limit:                aws.Int32(int32(limit)),
			ExclusiveStartKey:    lastEvaluatedKey,
			ProjectionExpression: aws.String("ID, Title, Content, CreatedAt"), // Fetch only necessary fields
		}

		// Execute the scan
		result, err := r.Client.Scan(context.TODO(), input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan posts: %w", err)
		}

		// Unmarshal items in a batch
		var batch []*models.Post
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &batch); err != nil {
			return nil, fmt.Errorf("failed to unmarshal posts: %w", err)
		}

		// Append to the result set
		posts = append(posts, batch...)

		// Check if we need to continue scanning
		lastEvaluatedKey = result.LastEvaluatedKey
		if lastEvaluatedKey == nil || len(posts) > itemsToSkip+limit {
			break
		}
	}

	// Trim results to the requested page
	start := itemsToSkip
	if start >= len(posts) {
		return []*models.Post{}, nil
	}

	end := start + limit
	if end > len(posts) {
		end = len(posts)
	}

	return posts[start:end], nil
}

func (r *DynamoPostRepository) GetByID(id string) (*models.Post, error) {
	input := &dynamodb.GetItemInput{
		TableName: &r.TableName,
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: id},
		},
	}

	result, err := r.Client.GetItem(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	if result.Item == nil {
		return nil, errors.New("post not found")
	}

	post := new(models.Post)
	if err := attributevalue.UnmarshalMap(result.Item, post); err != nil {
		return nil, fmt.Errorf("failed to unmarshal post: %w", err)
	}

	return post, nil
}

func (r *DynamoPostRepository) Create(post *models.Post) (*models.Post, error) {
	if post == nil {
		return nil, errors.New("post cannot be nil")
	}

	// Generate a unique ID for the post
	post.ID = generateUniqueID()

	// Marshal the post into a DynamoDB attribute map
	item, err := attributevalue.MarshalMap(post)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal post: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: &r.TableName,
		Item:      item,
	}

	_, err = r.Client.PutItem(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

func generateUniqueID() string {
	return uuid.New().String()
}

func (r *DynamoPostRepository) Update(id string, updatedPost *models.Post) (*models.Post, error) {
	existingPost, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	existingPost.Title = updatedPost.Title
	existingPost.Content = updatedPost.Content
	existingPost.Author = updatedPost.Author

	return r.Create(existingPost)
}

func (r *DynamoPostRepository) Delete(id string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: &r.TableName,
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberN{Value: id},
		},
	}

	_, err := r.Client.DeleteItem(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

// Helper function to convert an int to a pointer
func int32Ptr(i int) *int32 {
	val := int32(i)
	return &val
}
