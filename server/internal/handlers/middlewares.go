package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/middlewarr/server/internal/tools"
)

// With API Key
const (
	apiKeyHeader string = "X-Api-Key"
)

func withAPIKey(next http.Handler) http.Handler {
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

// With ID
type contextKey string

const (
	idContextKey contextKey = "id"
)

func withID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), idContextKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getIDFromContext(ctx context.Context) int {
	return ctx.Value(idContextKey).(int)
}
