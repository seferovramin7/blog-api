package routes

import (
	"blog-api/internal/handlers"
	"github.com/gorilla/mux"
)

func SetupRouter(postHandler handlers.PostHandlerInterface) *mux.Router {
	router := mux.NewRouter()

	router.Use(validationMiddleware)
	router.Use(corsMiddleware([]string{"*"}, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, []string{"Content-Type", "Authorization"}))
	router.Use(errorHandlingMiddleware)

	api := router.PathPrefix("/v1").Subrouter()
	api.HandleFunc("/posts", postHandler.GetAllPosts).Methods("GET")
	api.HandleFunc("/posts/{id:[0-9]+}", postHandler.GetPostByID).Methods("GET")
	api.HandleFunc("/posts", postHandler.CreatePost).Methods("POST")
	api.HandleFunc("/posts/{id:[0-9]+}", postHandler.UpdatePost).Methods("PUT")
	api.HandleFunc("/posts/{id:[0-9]+}", postHandler.DeletePost).Methods("DELETE")

	return router
}
