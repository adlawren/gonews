package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gonews/config"
	"gonews/db"
	"gonews/feed"
	"gonews/lib"
	"gonews/middleware"
	"net/http"
	"os"
	"path"
	"strconv"
	"text/template"

	"github.com/justinas/nosurf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	envAuth     = "GONEWS_AUTH"
	envDebug    = "GONEWS_DEBUG"
	envTLS      = "GONEWS_TLS"
	secretsPath = "/var/run/secrets" // #nosec G101
)

var (
	certPath = path.Join(secretsPath, "tls_cert")
	keyPath  = path.Join(secretsPath, "tls_key")
)

var cfg *config.Config
var dbCfg *config.DBConfig

func indexHandlerFunc(w http.ResponseWriter, r *http.Request) {
	tmplParams := make(map[string]string)
	tmplParams["title"] = cfg.AppTitle
	tmplParams["token"] = nosurf.Token(r)

	t, err := template.ParseFiles("assets/index.html.tmpl")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse html template")
		return
	}

	err = t.Execute(w, tmplParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render html template")
	}
}

func itemsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var tagName string
	tagList, exists := queryParams["tag_name"]
	if exists {
		if len(tagList) > 0 {
			tagName = tagList[0]
		}
	}

	db, err := db.New(dbCfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create db client")
		return
	}

	defer db.Close()

	var items []*feed.Item
	if tagName == "" {
		items, err = db.Items()
	} else {
		items, err = db.ItemsFromTag(&feed.Tag{Name: tagName})
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items")
		return
	}

	text, err := json.Marshal(&items)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(text)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render json")
		return
	}
}

func hideHandlerFunc(w http.ResponseWriter, r *http.Request) {
	db, err := db.New(dbCfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create db client")
		return
	}

	defer db.Close()

	err = r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse form")
		return
	}

	id, err := strconv.ParseUint(r.PostFormValue("ID"), 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Failed parse form ID")
		return
	}

	item, err := db.Item(uint(id))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get item from ID")
		return
	}

	item.Hide = true

	err = db.SaveItem(item)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update item")
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

func main() {
	authEnabled := flag.Bool("auth", false, "enable user authentication")
	debugEnabled := flag.Bool("debug", false, "enable debug logging")
	tlsEnabled := flag.Bool("tls", false, "enable TLS")
	confDir := flag.String("conf-dir", ".config", "config directory path")
	dataDir := flag.String("data-dir", "/data/gonews", "data directory path")
	migrationsDir := flag.String("migrations-dir", "db/migrations", "DB migrations directory path")

	flag.Parse()

	if *debugEnabled || os.Getenv(envDebug) == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Create data dir if it doesn't exist
	_, err := os.Stat(*dataDir)
	if err != nil && !os.IsNotExist(err) {
		log.Error().Err(err).Msg("Failed to stat data directory")
		return
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(*dataDir, os.ModeDir)
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to create data directory")
		return
	}

	cfg, err = config.New(*confDir, "config")
	if err != nil {
		log.Error().Err(err).Msg("Failed to load config")
		return
	}

	dbCfg = &config.DBConfig{
		DSN: fmt.Sprintf("file:%s/db.sqlite3", *dataDir),
	}
	config.SetDBConfigInst(dbCfg)

	adb, err := db.New(dbCfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create db client")
		return
	}

	defer adb.Close()

	err = adb.Migrate(*migrationsDir)
	if err != nil {
		log.Error().Err(err).Msg("Failed to migrate db")
		return
	}

	err = lib.InsertMissingFeeds(cfg, adb)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert new feeds")
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/", nosurf.New(http.HandlerFunc(indexHandlerFunc)))
	mux.Handle("/hide", http.HandlerFunc(hideHandlerFunc))
	mux.Handle("/api/v1/items", http.HandlerFunc(itemsHandlerFunc))

	go func() {
		for {
			err := lib.WatchFeeds(cfg, dbCfg)
			log.Error().Err(err).Msg("Failed to watch feeds")
		}
	}()

	middlewareFuncs := []middleware.MiddlewareFunc{
		middleware.LogMiddlewareFunc,
		middleware.ThrottleMiddlewareFunc,
	}
	if *authEnabled || os.Getenv(envAuth) == "true" {
		middlewareFuncs = append(
			middlewareFuncs,
			middleware.AuthMiddlewareFunc)
	}

	wrappedHandler, err := middleware.Wrap(mux, middlewareFuncs...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to inject middleware")
		return
	}

	if *tlsEnabled || os.Getenv(envTLS) == "true" {
		err = http.ListenAndServeTLS(
			":8080",
			certPath,
			keyPath,
			wrappedHandler)
	} else {
		err = http.ListenAndServe(
			":8080",
			wrappedHandler)
	}
	log.Error().Err(err).Msg("Server failed")
}
