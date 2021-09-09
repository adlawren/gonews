package middleware_test // 'middleware_test' instead of 'middleware' to prevent gonews/test <- gonews/middleware <- gonews/test import cycle

import (
	"gonews/auth"
	"gonews/config"
	"gonews/middleware"
	"gonews/test"
	"gonews/user"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	migrationsDir = "../db/migrations"
)

func TestAuthMiddlewareReturnsUnauthorizedWhenCredsMissing(t *testing.T) {
	authMiddlewareHandler, err := middleware.AuthMiddlewareFunc(nil)
	assert.NoError(t, err)

	dbCfg, _ := test.InitDB(t, migrationsDir)
	config.SetDBConfigInst(dbCfg)

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()
	authMiddlewareHandler.ServeHTTP(w, req)

	assertLoginFailed(t, w)
}

func TestAuthMiddlewareReturnsUnauthorizedWhenCredsInvalid(t *testing.T) {
	authMiddlewareHandler, err := middleware.AuthMiddlewareFunc(nil)
	assert.NoError(t, err)

	dbCfg, _ := test.InitDB(t, migrationsDir)
	config.SetDBConfigInst(dbCfg)

	req := httptest.NewRequest("GET", "http://example.com/", nil)

	mockUsername := test.MockUsername()
	mockPassword := test.MockPassword()
	req.SetBasicAuth(mockUsername, mockPassword)

	w := httptest.NewRecorder()
	authMiddlewareHandler.ServeHTTP(w, req)

	assertLoginFailed(t, w)
}

func TestAuthMiddlewareCallsNextHandlerWhenCredsValid(t *testing.T) {
	mockResponseText := "mock response"
	var mockHandlerFunc http.HandlerFunc = func(
		w http.ResponseWriter,
		r *http.Request) {
		w.Write([]byte(mockResponseText))
	}
	authMiddlewareHandler, err := middleware.AuthMiddlewareFunc(mockHandlerFunc)
	assert.NoError(t, err)

	dbCfg, db := test.InitDB(t, migrationsDir)
	config.SetDBConfigInst(dbCfg)

	req := httptest.NewRequest("GET", "http://example.com/", nil)

	mockUsername := test.MockUsername()
	mockPassword := test.MockPassword()
	mockPasswordHash, err := auth.Hash(mockPassword)

	db.Save(&user.User{
		Username:     mockUsername,
		PasswordHash: mockPasswordHash,
	})

	req.SetBasicAuth(mockUsername, mockPassword)

	w := httptest.NewRecorder()
	authMiddlewareHandler.ServeHTTP(w, req)

	resp := w.Result()
	res, err := ioutil.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, mockResponseText, string(res))
}

func assertLoginFailed(t *testing.T, w *httptest.ResponseRecorder) {
	resp := w.Result()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	authHeader := resp.Header.Get("WWW-Authenticate")
	assert.Equal(t, `Basic realm="Login required"`, authHeader)

	res, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Login required", string(res))
}
