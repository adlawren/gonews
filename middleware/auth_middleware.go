package middleware

import (
	"gonews/auth"
	"gonews/config"
	"gonews/db"
	"net/http"

	"github.com/rs/zerolog/log"
)

func AuthMiddlewareFunc(h http.Handler) http.Handler {
	var handlerFunc http.HandlerFunc = func(
		w http.ResponseWriter,
		r *http.Request) {
		dbCfg := config.DBConfigInst()

		db, err := db.New(dbCfg)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create db client")
			return
		}

		defer db.Close()

		// Advance declaration for goto
		var isValid bool

		username, password, ok := r.BasicAuth()
		if !ok {
			goto loginFailed
		}

		isValid, err = auth.IsValid(username, password, db)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate login")
			goto loginFailed
		}

		if isValid {
			h.ServeHTTP(w, r)
			return
		}

	loginFailed:
		w.Header().Add(
			"WWW-Authenticate",
			`Basic realm="Login required"`)
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write([]byte("Login required"))
		if err != nil {
			log.Error().Err(err).Msg("Failed to write response")
		}
	}

	return handlerFunc
}
