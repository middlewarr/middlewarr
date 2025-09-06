package proxy

import (
	"fmt"
	"net/http"
	"time"

	"github.com/middlewarr/server/internal/tools"
	"github.com/rs/zerolog/hlog"
)

type Middleware func(http.Handler) http.Handler

func chainMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

// Log Request
func middlewareLogRequest() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxyID := r.Header.Get("X-Proxy-Id")

			l := tools.GetLoggerFile(proxyID)

			h := hlog.NewHandler(*l)

			accessHandler := hlog.AccessHandler(func(req *http.Request, status, size int, duration time.Duration) {
				hlog.FromRequest(req).Info().
					Str("method", req.Method).
					Str("url", tools.SanitizeURI(req)).
					Int("status_code", status).
					Int("response_size_bytes", size).
					Dur("elapsed_ms", duration).
					Str("proxy_app", r.Header.Get("X-Proxy-App")).
					Str("proxy_service", r.Header.Get("X-Proxy-Service")).
					Msg("Request")
			})

			userAgentHandler := hlog.UserAgentHandler("user_agent")

			h(accessHandler(userAgentHandler(next))).ServeHTTP(w, r)
		})
	}
}

// Validate Requert
func validateRequest(r *http.Request, endpoints ProxyEndpoints) bool {
	method := r.Method
	path := r.URL.Path

	for _, endpoint := range endpoints {
		if method == endpoint.Method && endpoint.PathRegex.MatchString(path) {
			return true
		}
	}

	return false
}

func middlewareValidateRequest(endpoints ProxyEndpoints) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ok := validateRequest(r, endpoints)
			if !ok {
				http.Error(w, fmt.Sprintf("Forbidden, %s %s not allowed", r.Method, r.URL.Path), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
