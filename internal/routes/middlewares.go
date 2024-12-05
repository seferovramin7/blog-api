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
	for _, validType := range validTypes {
		if strings.Contains(contentType, validType) {
			return true
		}
	}
	return false
}

func validationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == http.MethodPost || r.Method == http.MethodPut) &&
			!validateContentType(r, []string{"application/json"}) {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			if err := json.NewEncoder(w).Encode(JSONErrorResponse{
				Error:       "Unsupported Media Type",
				Description: "Content-Type must be application/json",
			}); err != nil {
				log.Printf("Error encoding response: %v", err)
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

func errorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var requestBody bytes.Buffer
		if r.Body != nil {
			tee := io.TeeReader(r.Body, &requestBody)
			r.Body = io.NopCloser(&requestBody)
			_, err := io.ReadAll(tee)
			if err != nil {
				return
			}
		}

		log.Printf("Request: Method: %s, URL: %s, Headers: %v, Body: %s",
			r.Method, r.URL.Path, r.Header, requestBody.String())

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		log.Printf("Response: Method: %s, URL: %s, Status: %d, Time: %v",
			r.Method, r.URL.Path, lrw.statusCode, time.Since(start))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	lrw.body.Write(data)
	return lrw.ResponseWriter.Write(data)
}

func corsMiddleware(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
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
				w.WriteHeader(http.StatusForbidden)
				if err := json.NewEncoder(w).Encode(JSONErrorResponse{
					Error:       "Forbidden",
					Description: "Origin not allowed",
				}); err != nil {
					log.Printf("Error encoding response: %v", err)
				}
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
