package routes

import (
	"log"
	"net/http"

	"blog-api/internal/handlers"
	"github.com/gorilla/mux"
)

const (
	APIPrefix  = "/v1"
	PostsBase  = "/posts"
	PostWithID = "/posts/{id:[0-9]+}"
)

func SetupRouter(postHandler handlers.PostHandlerInterface) *mux.Router {
	log.Println("SetupRouter: Initializing router setup...")

	router := mux.NewRouter().StrictSlash(true)
	log.Println("SetupRouter: Router initialized with StrictSlash(true).")

	router.SkipClean(true)
	log.Println("SetupRouter: Router SkipClean set to true.")

	allowedOrigins := []string{"*"} // We can replace "*" with specific origins for production
	log.Printf("SetupRouter: Allowed origins set to: %v\n", allowedOrigins)

	allowedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	log.Printf("SetupRouter: Allowed methods set to: %v\n", allowedMethods)

	allowedHeaders := []string{"Content-Type", "Authorization"}
	log.Printf("SetupRouter: Allowed headers set to: %v\n", allowedHeaders)

	router.Use(loggingMiddleware)
	log.Println("SetupRouter: Logging middleware added.")

	router.Use(corsMiddleware(allowedOrigins, allowedMethods, allowedHeaders))
	log.Println("SetupRouter: CORS middleware added with specified origins, methods, and headers.")

	router.Use(errorHandlingMiddleware)
	log.Println("SetupRouter: Error handling middleware added.")

	api := router.PathPrefix(APIPrefix).Subrouter()
	log.Printf("SetupRouter: API subrouter created with prefix: %s\n", APIPrefix)

	log.Println("SetupRouter: Setting up API routes...")
	api.HandleFunc(PostsBase, postHandler.GetAllPosts).Methods(http.MethodGet)
	log.Printf("SetupRouter: Route added - GET %s (Handler: GetAllPosts)\n", PostsBase)

	api.HandleFunc(PostWithID, postHandler.GetPostByID).Methods(http.MethodGet)
	log.Printf("SetupRouter: Route added - GET %s (Handler: GetPostByID)\n", PostWithID)

	api.HandleFunc(PostsBase, postHandler.CreatePost).Methods(http.MethodPost)
	log.Printf("SetupRouter: Route added - POST %s (Handler: CreatePost)\n", PostsBase)

	api.HandleFunc(PostWithID, postHandler.UpdatePost).Methods(http.MethodPut)
	log.Printf("SetupRouter: Route added - PUT %s (Handler: UpdatePost)\n", PostWithID)

	api.HandleFunc(PostWithID, postHandler.PatchPost).Methods(http.MethodPatch)
	log.Printf("SetupRouter: Route added - PATCH %s (Handler: PatchPost)\n", PostWithID)

	api.HandleFunc(PostWithID, postHandler.DeletePost).Methods(http.MethodDelete)
	log.Printf("SetupRouter: Route added - DELETE %s (Handler: DeletePost)\n", PostWithID)

	log.Println("SetupRouter: Router setup completed successfully.")
	return router
}
