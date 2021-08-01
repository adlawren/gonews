package main

import (
	"context"
	"flag"
	"gonews/rss"

	"github.com/rs/zerolog/log"
)

func main() {
	xmlPath := flag.String("xml-path", "", "path to xml file")
	port := flag.Int("port", 8081, "port")

	flag.Parse()

	if len(*xmlPath) > 0 {
		err := rss.Serve(context.Background(), *xmlPath, *port)
		if err != nil {
			log.Error().Err(err).Msg("Failed to serve RSS")
			return
		}
	}
}
