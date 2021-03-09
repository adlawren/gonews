package middleware

import (
	"net/http"

	"github.com/pkg/errors"
)

type MiddlewareFunc func(http.Handler) (http.Handler, error)

func Wrap(handler http.Handler, middleware ...MiddlewareFunc) (http.Handler, error) {
	if len(middleware) == 0 {
		return handler, nil
	}

	h, err := Wrap(handler, middleware[1:]...)
	if err != nil {
		return h, errors.Wrap(err, "failed to wrap handler")
	}

	return middleware[0](h)
}
