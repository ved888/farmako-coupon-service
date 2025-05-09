package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

// corsOptions setting up routes for cors
func corsOptions() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"*", "GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*", "Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Access-Token", "importDate", "X-Client-Version", "Cache-Control", "Pragma", "x-started-at", "x-api-key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
	})
}

func CORSMiddleware() func(http.Handler) http.Handler {
	return corsOptions().Handler
}
