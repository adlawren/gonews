package middleware

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapReturnsHandlerWhenNoMiddlewareProvided(t *testing.T) {
	responseText := "mock response"
	handler := mockHandler(responseText)

	wrappedHandler, err := Wrap(handler)
	assert.NoError(t, err)
	assert.NotNil(t, wrappedHandler)

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	res, err := ioutil.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, responseText, string(res))
}

func TestWrapReturnsWrappedHandlerWhenMiddlewareProvided(t *testing.T) {
	handlerResponseText := "handler"
	handler := mockHandler(handlerResponseText)

	middleware1ResponseText := "middleware1"
	middleware1 := mockMiddleware(middleware1ResponseText)

	middleware2ResponseText := "middleware2"
	middleware2 := mockMiddleware(middleware2ResponseText)

	middleware3ResponseText := "middleware3"
	middleware3 := mockMiddleware(middleware3ResponseText)

	wrappedHandler, err := Wrap(handler, middleware1, middleware2, middleware3)
	assert.NoError(t, err)
	assert.NotNil(t, wrappedHandler)

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	res, err := ioutil.ReadAll(resp.Body)

	assert.NoError(t, err)
	expectedResponseText := fmt.Sprintf(
		"%s%s%s%s",
		middleware1ResponseText,
		middleware2ResponseText,
		middleware3ResponseText,
		handlerResponseText,
	)
	assert.Equal(t, expectedResponseText, string(res))
}

// TODO
func TestWrapReturnsErrorWhenMiddlewareReturnsError(t *testing.T) {
	handlerResponseText := "handler"
	handler := mockHandler(handlerResponseText)

	middleware1ResponseText := "middleware1"
	middleware1 := mockMiddleware(middleware1ResponseText)

	mockErr := errors.New("mock error")
	var middleware2 MiddlewareFunc = func(h http.Handler) (http.Handler, error) {
		return nil, mockErr
	}

	middleware3ResponseText := "middleware3"
	middleware3 := mockMiddleware(middleware3ResponseText)

	_, err := Wrap(handler, middleware1, middleware2, middleware3)
	expectedErrMsg := fmt.Sprintf("failed to wrap handler: %v", mockErr.Error())
	assert.EqualError(t, err, expectedErrMsg)
}

func mockHandler(responseText string) http.Handler {
	var handlerFunc http.HandlerFunc = func(
		w http.ResponseWriter,
		r *http.Request) {
		w.Write([]byte(responseText))
	}

	return handlerFunc
}

func mockMiddleware(responseText string) MiddlewareFunc {
	var middlewareFunc MiddlewareFunc = func(h http.Handler) (http.Handler, error) {
		var handlerFunc http.HandlerFunc = func(
			w http.ResponseWriter,
			r *http.Request) {
			w.Write([]byte(responseText))
			h.ServeHTTP(w, r)
		}
		return handlerFunc, nil
	}

	return middlewareFunc
}
