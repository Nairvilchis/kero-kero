package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORS configura el middleware de CORS
func CORS(allowedOrigins, allowedMethods, allowedHeaders []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   allowedMethods,
		AllowedHeaders:   allowedHeaders,
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
