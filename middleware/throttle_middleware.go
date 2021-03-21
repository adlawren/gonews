package middleware

import (
	"net/http"

	limiter "github.com/ulule/limiter"
	stdlib "github.com/ulule/limiter/drivers/middleware/stdlib"
	memory "github.com/ulule/limiter/drivers/store/memory"
)

func ThrottleMiddlewareFunc(h http.Handler) (http.Handler, error) {
	rate, err := limiter.NewRateFromFormatted("3-S")
	if err != nil {
		return nil, err
	}

	store := memory.NewStore()
	lmt := limiter.New(store, rate)
	m := stdlib.NewMiddleware(lmt)
	return m.Handler(h), nil
}
