package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type JSONErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
}

func validateContentType(r *http.Request, validTypes []string) bool {
	contentType := r.Header.Get("Content-Type")
	log.Printf("Validating Content-Type. Found: %s, Valid Types: %v", contentType, validTypes)
	for _, validType := range validTypes {
		if strings.Contains(contentType, validType) {
			log.Printf("Content-Type validation succeeded for: %s", validType)
			return true
		}
	}
	log.Printf("Content-Type validation failed for: %s", contentType)
	return false
}

func validationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Entering validationMiddleware. Method: %s, URL: %s, Headers: %v", r.Method, r.URL.Path, r.Header)
		if (r.Method == http.MethodPost || r.Method == http.MethodPut) &&
			!validateContentType(r, []string{"application/json"}) {
			log.Printf("Request validation failed. Method: %s, URL: %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			if err := json.NewEncoder(w).Encode(JSONErrorResponse{
				Error:       "Unsupported Media Type",
				Description: "Content-Type must be application/json",
			}); err != nil {
				log.Printf("Error encoding response: %v", err)
			}
			return
		}
		log.Printf("Request validation succeeded. Proceeding to next handler.")
		next.ServeHTTP(w, r)
	})
}

func errorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Entering errorHandlingMiddleware. Method: %s, URL: %s", r.Method, r.URL.Path)
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Recovered from panic: %v. Method: %s, URL: %s, Headers: %v",
					rec, r.Method, r.URL.Path, r.Header)
				w.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(JSONErrorResponse{
					Error:       "Internal Server Error",
					Description: "A server error occurred. Please contact support.",
				}); err != nil {
					log.Printf("Error encoding response: %v", err)
				}
			}
		}()
		log.Printf("Calling next handler in errorHandlingMiddleware.")
		next.ServeHTTP(w, r)
		log.Printf("Exiting errorHandlingMiddleware. Method: %s, URL: %s", r.Method, r.URL.Path)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Entering loggingMiddleware. Method: %s, URL: %s, Headers: %v", r.Method, r.URL.Path, r.Header)

		var requestBody bytes.Buffer
		if r.Body != nil {
			tee := io.TeeReader(r.Body, &requestBody)
			r.Body = io.NopCloser(&requestBody)
			bodyContent, err := io.ReadAll(tee)
			if err != nil {
				log.Printf("Error reading request body: %v", err)
				return
			}
			log.Printf("Request Body: %s", string(bodyContent))
		}

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		log.Printf("Calling next handler in loggingMiddleware.")
		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf("Exiting loggingMiddleware. Method: %s, URL: %s, Status: %d, Duration: %v",
			r.Method, r.URL.Path, lrw.statusCode, duration)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	log.Printf("Writing response header. Status Code: %d", code)
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	log.Printf("Writing response body. Data: %s", string(data))
	lrw.body.Write(data)
	return lrw.ResponseWriter.Write(data)
}

func corsMiddleware(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Entering corsMiddleware. Method: %s, URL: %s, Origin: %s", r.Method, r.URL.Path, r.Header.Get("Origin"))

			origin := r.Header.Get("Origin")
			if origin == "" {
				log.Printf("No Origin header found. Proceeding to next handler.")
				next.ServeHTTP(w, r)
				return
			}

			allowed := false
			for _, ao := range allowedOrigins {
				if ao == "*" || ao == origin {
					allowed = true
					break
				}
			}

			if !allowed {
				log.Printf("Origin not allowed: %s", origin)
				w.WriteHeader(http.StatusForbidden)
				if err := json.NewEncoder(w).Encode(JSONErrorResponse{
					Error:       "Forbidden",
					Description: "Origin not allowed",
				}); err != nil {
					log.Printf("Error encoding response: %v", err)
				}
				return
			}

			log.Printf("Setting CORS headers. Allowed Origin: %s", origin)
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))

			if r.Method == http.MethodOptions {
				log.Printf("Preflight request detected. Returning OK status.")
				w.WriteHeader(http.StatusOK)
				return
			}

			log.Printf("Proceeding to next handler in corsMiddleware.")
			next.ServeHTTP(w, r)
			log.Printf("Exiting corsMiddleware. Method: %s, URL: %s", r.Method, r.URL.Path)
		})
	}
}
