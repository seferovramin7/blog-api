package routes

import (
	"encoding/json"
	"log"
	"net/http"
)

type JSONErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
}

func validationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if r.Header.Get("Content-Type") != "application/json" {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				json.NewEncoder(w).Encode(JSONErrorResponse{
					Error:       "Unsupported Media Type",
					Description: "Content-Type must be application/json",
				})
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func errorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Recovered from panic: %v", rec)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(JSONErrorResponse{
					Error:       "Internal Server Error",
					Description: "A server error occurred.",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
