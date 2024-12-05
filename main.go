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
	"time"
)

func loadDynamoDBClient() *dynamodb.Client {
	startTime := time.Now()
	log.Println("[INFO] Initializing DynamoDB client")

	customResolver := aws.EndpointResolverFunc(
		func(service, region string) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				log.Printf("[DEBUG] Resolving endpoint for service: %s in region: %s", service, region)
				return aws.Endpoint{
					URL:           getEnv("DYNAMODB_ENDPOINT", "http://localhost:8000"),
					SigningRegion: getEnv("DYNAMODB_REGION", "us-west-2"),
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown service: %s", service)
		},
	)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolver(customResolver),
	)
	if err != nil {
		log.Fatalf("[ERROR] Failed to load AWS config: %v", err)
	}

	log.Printf("[INFO] DynamoDB client initialized successfully (duration: %v)", time.Since(startTime))
	return dynamodb.NewFromConfig(cfg)
}

func lambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("[INFO] Lambda handler invoked with method: %s, path: %s", request.HTTPMethod, request.Path)

	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] No .env file found, using system environment variables")
	} else {
		log.Println("[INFO] .env file loaded successfully")
	}

	client := loadDynamoDBClient()
	tableName := getEnv("DYNAMODB_TABLE", "TestTable")
	log.Printf("[INFO] Using DynamoDB table: %s", tableName)

	repo := repository.NewDynamoPostRepository(client, tableName)
	log.Println("[INFO] DynamoDB repository initialized")

	service := services.NewPostService(repo)
	log.Println("[INFO] Post service initialized")

	postHandler := handlers.NewPostHandler(service)
	log.Println("[INFO] Post handler initialized")

	router := routes.SetupRouter(postHandler)
	log.Println("[INFO] Router setup completed")

	httpRequest, err := http.NewRequest(request.HTTPMethod, request.Path, nil)
	if err != nil {
		log.Printf("[ERROR] Failed to create HTTP request: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	responseRecorder := &responseRecorder{}
	log.Println("[DEBUG] Invoking router with constructed HTTP request")
	router.ServeHTTP(responseRecorder, httpRequest)

	log.Printf("[INFO] Request processed with status: %d", responseRecorder.statusCode)
	return events.APIGatewayProxyResponse{
		StatusCode: responseRecorder.statusCode,
		Body:       responseRecorder.body,
	}, nil
}

type responseRecorder struct {
	statusCode int
	body       string
}

func (r *responseRecorder) Header() http.Header {
	log.Println("[DEBUG] Accessing response recorder headers")
	return http.Header{}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = string(b)
	log.Printf("[DEBUG] Writing response body: %s", r.body)
	return len(b), nil
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	log.Printf("[DEBUG] Writing response status code: %d", r.statusCode)
}

func main() {
	log.Println("[INFO] Starting Lambda function")
	lambda.Start(lambdaHandler)
	log.Println("[INFO] Lambda function execution completed")
}

func getEnv(key, fallback string) string {
	log.Printf("[DEBUG] Fetching environment variable: %s", key)
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("[INFO] Environment variable %s found with value: %s", key, value)
		return value
	}
	log.Printf("[WARN] Environment variable %s not found, using fallback: %s", key, fallback)
	return fallback
}
