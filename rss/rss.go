package rss

import (
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

func Serve(xmlPath string, port int) error {
	_, err := os.Stat(xmlPath)
	if err != nil {
		return errors.Wrap(err, "failed to stat XML file")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, xmlPath)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
