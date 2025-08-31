package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const (
	idKey contextKey = "id"
)

func WithID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), idKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetIDFromContext(ctx context.Context) int {
	return ctx.Value(idKey).(int)
}
