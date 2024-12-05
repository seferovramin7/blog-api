package main

import (
	"blog-api/internal/handlers"
	"blog-api/internal/repository"
	"blog-api/internal/routes"
	"blog-api/internal/services"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
)

var (
	router         http.Handler
	configurations appConfig
)

type appConfig struct {
	DynamoDBEndpoint string
	DynamoDBRegion   string
	DynamoDBTable    string
}

// init runs once during cold starts to initialize dependencies
func init() {
	log.Println("Initializing application...")

	if err := loadEnv(); err != nil {
		log.Println("No .env file found, falling back to system environment variables")
	}

	// Load configurations
	configurations = loadAppConfig()

	// Set up dependencies
	client := newDynamoDBClient(configurations)
	repo := repository.NewDynamoPostRepository(client, configurations.DynamoDBTable)
	service := services.NewPostService(repo)
	postHandler := handlers.NewPostHandler(service)

	// Initialize the router
	router = routes.SetupRouter(postHandler)
}

// loadEnv loads environment variables from a .env file
func loadEnv() error {
	return godotenv.Load()
}

// loadAppConfig loads and validates configuration from environment variables
func loadAppConfig() appConfig {
	return appConfig{
		DynamoDBEndpoint: getEnv("DYNAMODB_ENDPOINT", "http://host.docker.internal:8000"),
		DynamoDBRegion:   getEnv("DYNAMODB_REGION", "us-east-1"),
		DynamoDBTable:    getEnv("DYNAMODB_TABLE", "TestTable"),
	}
}

// newDynamoDBClient initializes a DynamoDB client
func newDynamoDBClient(cfg appConfig) *dynamodb.Client {
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
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	return dynamodb.NewFromConfig(awsCfg)
}

// lambdaHandler processes API Gateway events
func lambdaHandler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	httpReq, err := convertAPIGatewayRequestToHTTPRequest(req)
	if err != nil {
		log.Printf("Error converting request: %v", err)
		return errorResponse(http.StatusInternalServerError, "Failed to parse request", err), nil
	}

	respWriter := newLambdaResponseWriter()
	router.ServeHTTP(respWriter, httpReq)

	return respWriter.buildAPIGatewayProxyResponse(), nil
}

// main starts the Lambda function
func main() {
	log.Println("Starting Lambda handler...")
	lambda.Start(lambdaHandler)
}

// Utility Functions

// getEnv retrieves an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// convertAPIGatewayRequestToHTTPRequest converts an API Gateway request to an HTTP request
func convertAPIGatewayRequestToHTTPRequest(req events.APIGatewayProxyRequest) (*http.Request, error) {
	body := req.Body
	if req.IsBase64Encoded {
		decodedBody, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 body: %w", err)
		}
		body = string(decodedBody)
	}

	httpReq, err := http.NewRequest(req.HTTPMethod, req.Path, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	query := httpReq.URL.Query()
	for k, v := range req.QueryStringParameters {
		query.Add(k, v)
	}
	httpReq.URL.RawQuery = query.Encode()

	return httpReq, nil
}

// errorResponse creates a standardized error response
func errorResponse(statusCode int, message string, err error) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       fmt.Sprintf(`{"error": "%s", "details": "%v"}`, message, err),
	}
}

// lambdaResponseWriter is a custom HTTP response writer for Lambda
type lambdaResponseWriter struct {
	statusCode int
	headers    http.Header
	body       []byte
}

func newLambdaResponseWriter() *lambdaResponseWriter {
	return &lambdaResponseWriter{
		headers: http.Header{},
	}
}

func (rw *lambdaResponseWriter) Header() http.Header {
	return rw.headers
}

func (rw *lambdaResponseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return len(b), nil
}

func (rw *lambdaResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

func (rw *lambdaResponseWriter) buildAPIGatewayProxyResponse() events.APIGatewayProxyResponse {
	headers := make(map[string]string)
	for k, v := range rw.headers {
		headers[k] = v[0]
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      rw.statusCode,
		Headers:         headers,
		Body:            string(rw.body),
		IsBase64Encoded: false,
	}
}
