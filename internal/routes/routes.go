package routes

import (
	"net/http"

	"blog-api/internal/handlers"
	"github.com/gorilla/mux"
)

const (
	APIPrefix  = ""
	PostsBase  = "/posts"
	PostWithID = "/posts/{id:[0-9]+}"
)

func SetupRouter(postHandler handlers.PostHandlerInterface) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.SkipClean(true)

	allowedOrigins := []string{"*"} // We can replace "*" with specific origins for production
	allowedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	allowedHeaders := []string{"Content-Type", "Authorization"}

	router.Use(loggingMiddleware)
	router.Use(corsMiddleware(allowedOrigins, allowedMethods, allowedHeaders))
	router.Use(errorHandlingMiddleware)

	api := router.PathPrefix(APIPrefix).Subrouter()

	api.HandleFunc(PostsBase, postHandler.GetAllPosts).Methods(http.MethodGet)
	api.HandleFunc(PostWithID, postHandler.GetPostByID).Methods(http.MethodGet)
	api.HandleFunc(PostsBase, postHandler.CreatePost).Methods(http.MethodPost)
	api.HandleFunc(PostWithID, postHandler.UpdatePost).Methods(http.MethodPut)
	api.HandleFunc(PostWithID, postHandler.PatchPost).Methods(http.MethodPatch)
	api.HandleFunc(PostWithID, postHandler.DeletePost).Methods(http.MethodDelete)

	return router
}
