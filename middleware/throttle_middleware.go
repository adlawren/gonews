package middleware

import (
	"net/http"

	"github.com/didip/tollbooth"
)

func ThrottleMiddlewareFunc(h http.Handler) http.Handler {
	var maxRequests float64 = 1.0
	var seconds float64 = 5.0
	limiter := tollbooth.NewLimiter(maxRequests/seconds, nil)
	return tollbooth.LimitFuncHandler(
		limiter,
		func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		})
}
