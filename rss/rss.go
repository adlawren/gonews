package rss

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

func Serve(ctx context.Context, xmlPath string, port int) error {
	_, err := os.Stat(xmlPath)
	if err != nil {
		return fmt.Errorf("failed to stat XML file: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, xmlPath)
	})

	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", port),
		Handler:     mux,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	go func() {
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Failed to listen and serve")
		}
	}()

	select {
	case <-ctx.Done():
		err = server.Shutdown(context.Background())
		if err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	return nil
}
