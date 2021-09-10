package middleware

import (
	"fmt"
	"net/http"
)

type MiddlewareFunc func(http.Handler) (http.Handler, error)

func Wrap(handler http.Handler, middleware ...MiddlewareFunc) (http.Handler, error) {
	if len(middleware) == 0 {
		return handler, nil
	}

	h, err := Wrap(handler, middleware[1:]...)
	if err != nil {
		return h, fmt.Errorf("failed to wrap handler: %w", err)
	}

	return middleware[0](h)
}
