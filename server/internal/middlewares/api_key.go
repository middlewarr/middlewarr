package middlewares

import (
	"net/http"

	"github.com/middlewarr/server/internal/tools"
)

const (
	apiKeyHeader string = "X-Api-Key"
)

func WithAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := tools.GetSettings()

		apiKey := r.Header.Get(apiKeyHeader)
		serviceApiKey := s.String("apiKey")

		if apiKey != serviceApiKey {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
