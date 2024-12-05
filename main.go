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
	log.Println("Initializing DynamoDB client")

	// Define a custom endpoint resolver
	customResolver := aws.EndpointResolverFunc(
		func(service, region string) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				endpoint := getEnv("DYNAMODB_ENDPOINT", "http://localhost:8000")
				region := getEnv("DYNAMODB_REGION", "us-west-2")
				log.Printf("Custom resolver: Service=%s, Endpoint=%s, Region=%s", service, endpoint, region)
				return aws.Endpoint{
					URL:           endpoint,
					SigningRegion: region,
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown service: %s", service)
		},
	)

	// Load AWS configuration
	log.Println("Loading AWS configuration")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolver(customResolver),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	log.Println("AWS configuration loaded successfully")
	return dynamodb.NewFromConfig(cfg)
}

func lambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Lambda handler invoked")

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	} else {
		log.Println(".env file loaded successfully")
	}

	client := loadDynamoDBClient()
	tableName := getEnv("DYNAMODB_TABLE", "TestTable")
	log.Printf("Using DynamoDB table: %s", tableName)

	repo := repository.NewDynamoPostRepository(client, tableName)
	log.Println("DynamoDB repository initialized")

	service := services.NewPostService(repo)
	log.Println("Post service initialized")

	postHandler := handlers.NewPostHandler(service)
	log.Println("Post handler initialized")

	router := routes.SetupRouter(postHandler)
	log.Println("Router setup complete")

	// Convert API Gateway request to HTTP request
	log.Printf("Converting API Gateway request to HTTP request: Method=%s, Path=%s", request.HTTPMethod, request.Path)
	httpRequest, err := http.NewRequest(request.HTTPMethod, request.Path, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Custom response recorder to capture the response from the router
	responseRecorder := &responseRecorder{}
	log.Println("Serving HTTP request through router")
	router.ServeHTTP(responseRecorder, httpRequest)

	log.Printf("HTTP response generated: StatusCode=%d, Body=%s", responseRecorder.statusCode, responseRecorder.body)
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
	log.Println("responseRecorder: Header invoked")
	return http.Header{}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = string(b)
	log.Printf("responseRecorder: Write invoked, Body=%s", r.body)
	return len(b), nil
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	log.Printf("responseRecorder: WriteHeader invoked, StatusCode=%d", r.statusCode)
}

func main() {
	log.Println("Starting Lambda function")
	lambda.Start(lambdaHandler)
}

func getEnv(key, fallback string) string {
	log.Printf("Fetching environment variable: Key=%s", key)
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("Environment variable found: Key=%s, Value=%s", key, value)
		return value
	}
	log.Printf("Environment variable not found, using fallback: Key=%s, Fallback=%s", key, fallback)
	return fallback
}
