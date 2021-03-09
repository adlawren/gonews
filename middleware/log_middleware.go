package middleware

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func LogMiddlewareFunc(h http.Handler) (http.Handler, error) {
	var handlerFunc http.HandlerFunc = func(
		w http.ResponseWriter,
		r *http.Request) {
		log.Info().
			Int64("ContentLength", r.ContentLength).
			Str("Method", r.Method).
			Str("RemoteAddr", r.RemoteAddr).
			Stringer("URL", r.URL).
			Msg("Incoming request")
		h.ServeHTTP(w, r)
	}

	return handlerFunc, nil
}
