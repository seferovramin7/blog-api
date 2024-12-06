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

// GetAll returns a paginated list of posts using a scan.
//
// Note: This approach is not ideal for large datasets due to performance implications.
// For large data sets, consider using key-based queries and LastEvaluatedKey for pagination.
func (r *DynamoPostRepository) GetAll(ctx context.Context, page, limit int) ([]*models.Post, error) {
	if page <= 0 || limit <= 0 {
		return nil, fmt.Errorf("invalid pagination parameters: page=%d, limit=%d", page, limit)
	}

	itemsToSkip := (page - 1) * limit
	var (
		posts            []*models.Post
		lastEvaluatedKey map[string]types.AttributeValue
	)

	for {
		input := &dynamodb.ScanInput{
			TableName:            aws.String(r.TableName),
			ExclusiveStartKey:    lastEvaluatedKey,
			Limit:                aws.Int32(int32(limit)),
			ProjectionExpression: aws.String("ID, Title, Content, Author, CreatedAt"),
		}

		result, err := r.Client.Scan(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan posts: %w", err)
		}

		var batch []*models.Post
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &batch); err != nil {
			return nil, fmt.Errorf("failed to unmarshal posts batch: %w", err)
		}

		posts = append(posts, batch...)
		lastEvaluatedKey = result.LastEvaluatedKey

		// Stop if we have enough items for the requested page or there are no more
		if len(posts) >= itemsToSkip+limit || lastEvaluatedKey == nil {
			break
		}
	}

	// If there aren't enough posts to even reach the requested page, return an empty slice
	if itemsToSkip >= len(posts) {
		return []*models.Post{}, nil
	}

	end := itemsToSkip + limit
	if end > len(posts) {
		end = len(posts)
	}

	return posts[itemsToSkip:end], nil
}

func (r *DynamoPostRepository) GetByID(ctx context.Context, id string) (*models.Post, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: id},
		},
		ProjectionExpression: aws.String("ID, Title, Content, Author, CreatedAt"),
	}

	result, err := r.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by ID=%s: %w", id, err)
	}
	if result.Item == nil {
		return nil, fmt.Errorf("post with ID=%s not found", id)
	}

	var post models.Post
	if err := attributevalue.UnmarshalMap(result.Item, &post); err != nil {
		return nil, fmt.Errorf("failed to unmarshal post with ID=%s: %w", id, err)
	}

	return &post, nil
}

func (r *DynamoPostRepository) Create(ctx context.Context, post *models.Post) (*models.Post, error) {
	if post == nil {
		return nil, errors.New("post cannot be nil")
	}

	if post.ID == "" {
		post.ID = generateUniqueID()
	}

	item, err := attributevalue.MarshalMap(post)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal post: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.TableName),
		Item:      item,
	}

	if _, err := r.Client.PutItem(ctx, input); err != nil {
		return nil, fmt.Errorf("failed to create post with ID=%s: %w", post.ID, err)
	}

	return post, nil
}

func (r *DynamoPostRepository) Update(ctx context.Context, id string, updatedPost *models.Post) (*models.Post, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	if updatedPost == nil {
		return nil, errors.New("updated post cannot be nil")
	}

	// Use UpdateItem to only change the necessary fields.
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("SET Title = :title, Content = :content, Author = :author"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":title":   &types.AttributeValueMemberS{Value: updatedPost.Title},
			":content": &types.AttributeValueMemberS{Value: updatedPost.Content},
			":author":  &types.AttributeValueMemberS{Value: updatedPost.Author},
		},
		ReturnValues: types.ReturnValueAllNew,
	}

	result, err := r.Client.UpdateItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update post with ID=%s: %w", id, err)
	}

	var post models.Post
	if err := attributevalue.UnmarshalMap(result.Attributes, &post); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated post with ID=%s: %w", id, err)
	}

	return &post, nil
}

func (r *DynamoPostRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: id},
		},
	}

	if _, err := r.Client.DeleteItem(ctx, input); err != nil {
		return fmt.Errorf("failed to delete post with ID=%s: %w", id, err)
	}

	return nil
}

func generateUniqueID() string {
	return uuid.New().String()
}
