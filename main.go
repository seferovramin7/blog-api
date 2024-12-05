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
	"log"
	"net/http"
	"os"
)

func loadDynamoDBClient() *dynamodb.Client {
	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{
					URL:           getEnv("DYNAMODB_ENDPOINT", "http://localhost:8000"),
					SigningRegion: getEnv("DYNAMODB_REGION", "us-west-2"),
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown service: %s", service)
		},
	)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	return dynamodb.NewFromConfig(cfg)
}

func lambdaHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	client := loadDynamoDBClient()
	tableName := getEnv("DYNAMODB_TABLE", "TestTable")
	repo := repository.NewDynamoPostRepository(client, tableName)
	service := services.NewPostService(repo)
	postHandler := handlers.NewPostHandler(service)

	router := routes.SetupRouter(postHandler)

	// Convert the Lambda event to an HTTP request
	httpRequest, err := createHTTPRequestFromAPIGatewayEvent(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("Failed to create HTTP request: %v", err),
		}, nil
	}

	// Custom response recorder to capture the response from the router
	responseRecorder := &responseRecorder{}
	router.ServeHTTP(responseRecorder, httpRequest)

	return events.APIGatewayProxyResponse{
		StatusCode: responseRecorder.statusCode,
		Body:       responseRecorder.body,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

type responseRecorder struct {
	statusCode int
	body       string
	header     http.Header
}

func (r *responseRecorder) Header() http.Header         { return http.Header{} }
func (r *responseRecorder) Write(b []byte) (int, error) { r.body = string(b); return len(b), nil }
func (r *responseRecorder) WriteHeader(statusCode int)  { r.statusCode = statusCode }

func createHTTPRequestFromAPIGatewayEvent(event events.APIGatewayProxyRequest) (*http.Request, error) {
	url := fmt.Sprintf("https://%s%s", event.Headers["Host"], event.Path)
	req, err := http.NewRequest(event.HTTPMethod, url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range event.Headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

func main() {
	lambda.Start(lambdaHandler)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
