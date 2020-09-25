package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gonews/config"
	"gonews/db"
	"gonews/feed"
	"gonews/lib"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/justinas/nosurf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	envDebug = "GONEWS_DEBUG"
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

	r.ParseForm()
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
	debug := flag.Bool("debug", false, "enable debug logging")
	confDir := flag.String("conf-dir", ".config", "config directory path")
	dataDir := flag.String("data-dir", "/data/gonews", "data directory path")
	migrationsDir := flag.String("migrations-dir", "db/migrations", "DB migrations directory path")

	flag.Parse()

	if *debug || os.Getenv(envDebug) == "true" {
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

	http.Handle("/", nosurf.New(http.HandlerFunc(indexHandlerFunc)))
	http.Handle("/hide", http.HandlerFunc(hideHandlerFunc))
	http.Handle("/api/v1/items", http.HandlerFunc(itemsHandlerFunc))

	go func() {
		for {
			err := lib.WatchFeeds(cfg, dbCfg)
			log.Error().Err(err).Msg("Failed to watch feeds")
		}
	}()

	err = http.ListenAndServe(":8080", nil)
	log.Error().Err(err).Msg("Server failed")
}
