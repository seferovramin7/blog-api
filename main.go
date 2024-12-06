package main

import (
	"blog-api/internal/handlers"
	"blog-api/internal/repository"
	"blog-api/internal/routes"
	"blog-api/internal/services"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// appConfig holds application-level configuration for DynamoDB.
type appConfig struct {
	DynamoDBEndpoint string
	DynamoDBRegion   string
	DynamoDBTable    string
}

func main() {
	// Set up application
	appCfg, err := loadAppConfig()
	if err != nil {
		log.Fatalf("Failed to load app configuration: %v", err)
	}

	dynamoClient, err := newDynamoDBClient(appCfg)
	if err != nil {
		log.Fatalf("Failed to create DynamoDB client: %v", err)
	}

	// Initialize repository, service, and handler
	repo := repository.NewDynamoPostRepository(dynamoClient, appCfg.DynamoDBTable)
	postService := services.NewPostService(repo)
	postHandler := handlers.NewPostHandler(postService)

	// Set up the HTTP router (using the project's internal routes)
	router := routes.SetupRouter(postHandler)

	// Wrap the router using lambda httpadapter
	adapter := httpadapter.New(router)

	// Start the Lambda function
	lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return adapter.ProxyWithContext(ctx, req)
	})
}

// loadAppConfig loads and validates configuration from environment variables.
func loadAppConfig() (appConfig, error) {
	cfg := appConfig{
		DynamoDBEndpoint: getEnv("DYNAMODB_ENDPOINT", "http://host.docker.internal:8000"),
		DynamoDBRegion:   getEnv("DYNAMODB_REGION", "us-east-1"),
		DynamoDBTable:    getEnv("DYNAMODB_TABLE", "TestTable"),
	}

	// Add validation if needed. For example:
	// if cfg.DynamoDBTable == "" {
	// 	return cfg, fmt.Errorf("DynamoDB table name cannot be empty")
	// }

	return cfg, nil
}

// newDynamoDBClient sets up a new DynamoDB client using a custom endpoint resolver.
func newDynamoDBClient(cfg appConfig) (*dynamodb.Client, error) {
	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service == dynamodb.ServiceID {
			return aws.Endpoint{
				URL:           cfg.DynamoDBEndpoint,
				SigningRegion: cfg.DynamoDBRegion,
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown service: %s", service)
	})

	awsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolver(customResolver))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return dynamodb.NewFromConfig(awsCfg), nil
}

// getEnv retrieves an environment variable with a fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsDuration retrieves an environment variable as a time.Duration,
// falling back if invalid or not present. (Retained for future use if needed.)
func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Invalid duration format for %s: %s; using fallback %s", key, value, fallback)
	}
	return fallback
}
