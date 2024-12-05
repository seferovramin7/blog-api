package main

import (
	"blog-api/internal/handlers"
	"blog-api/internal/repository"
	"blog-api/internal/routes"
	"blog-api/internal/services"
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func loadDynamoDBClient() *dynamodb.Client {
	// Define a custom endpoint resolver
	customResolver := aws.EndpointResolverFunc(
		func(service, region string) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{
					URL:           getEnv("DYNAMODB_ENDPOINT", "http://localhost:8000"),
					SigningRegion: getEnv("DYNAMODB_REGION", "us-west-2"),
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown service: %s", service)
		},
	)

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolver(customResolver),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	log.Println("AWS configuration successfully loaded")
	return dynamodb.NewFromConfig(cfg)
}

func lambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Lambda handler invoked")

	// Log incoming request details
	log.Printf("Request Details: %+v", request)

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Validate HTTPMethod and Path
	if request.HTTPMethod == "" || request.Path == "" {
		log.Println("Invalid request: Missing HTTP method or path")
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request: Missing HTTP method or path",
		}, nil
	}

	client := loadDynamoDBClient()
	tableName := getEnv("DYNAMODB_TABLE", "TestTable")
	repo := repository.NewDynamoPostRepository(client, tableName)
	log.Println("Using DynamoDB repository")

	service := services.NewPostService(repo)
	postHandler := handlers.NewPostHandler(service)

	router := routes.SetupRouter(postHandler)

	// Construct full URL safely
	host := request.Headers["Host"]
	if host == "" {
		host = "localhost" // Fallback for local testing
	}
	fullURL := fmt.Sprintf("https://%s%s", host, request.Path)
	log.Printf("Full request URL: %s", fullURL)

	// Create HTTP request
	httpRequest, err := http.NewRequest(request.HTTPMethod, fullURL, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Add query parameters to the HTTP request
	q := httpRequest.URL.Query()
	for key, value := range request.QueryStringParameters {
		q.Add(key, value)
	}
	httpRequest.URL.RawQuery = q.Encode()

	// Log query parameters
	log.Printf("Query Parameters: %v", httpRequest.URL.RawQuery)

	// Custom response recorder to capture the response from the router
	responseRecorder := &responseRecorder{}
	router.ServeHTTP(responseRecorder, httpRequest)

	// Log response details
	log.Printf("Response Status Code: %d", responseRecorder.statusCode)
	log.Printf("Response Body: %s", responseRecorder.body)

	return events.APIGatewayProxyResponse{
		StatusCode: responseRecorder.statusCode,
		Body:       responseRecorder.body,
	}, nil
}

type responseRecorder struct {
	statusCode int
	body       string
}

func (r *responseRecorder) Header() http.Header         { return http.Header{} }
func (r *responseRecorder) Write(b []byte) (int, error) { r.body = string(b); return len(b), nil }
func (r *responseRecorder) WriteHeader(statusCode int)  { r.statusCode = statusCode }

func main() {
	log.Println("Starting Lambda function")
	lambda.Start(lambdaHandler)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
