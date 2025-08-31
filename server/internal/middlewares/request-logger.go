package middlewares

import (
	"net/http"
	"time"

	"github.com/middlewarr/server/internal/tools"
	"github.com/rs/zerolog/hlog"
)

func RequestLogger(next http.Handler) http.Handler {
	l := tools.GetLogger()

	h := hlog.NewHandler(l)

	accessHandler := hlog.AccessHandler(
		func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Str("url", tools.SanitizeURI(r)).
				Int("status_code", status).
				Int("response_size_bytes", size).
				Dur("elapsed_ms", duration).
				Msg("Request")
		},
	)

	userAgentHandler := hlog.UserAgentHandler("user_agent")

	return h(accessHandler(userAgentHandler(next)))
}
