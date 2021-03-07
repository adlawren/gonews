package middleware

import "net/http"

type MiddlewareFunc func(http.Handler) http.Handler

func Wrap(handler http.Handler, middleware ...MiddlewareFunc) http.Handler {
	if len(middleware) == 0 {
		return handler
	}

	return middleware[0](Wrap(handler, middleware[1:]...))
}
