package main

import (
	"encoding/json"
	"fmt"
	"gonews/config"
	"gonews/db"
	"gonews/feed"
	"gonews/page"
	"gonews/parser"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	envDebug = "GONEWS_DEBUG"
	dataDir  = "/data/gonews"
	confDir  = ".config"
)

var cfg *config.Config

func appConfig() (*config.Config, error) {
	var err error
	if cfg == nil {
		cfg, err = config.New(confDir, "config")
	}

	return cfg, errors.Wrap(err, "failed to load config")
}

func dbConfig() *config.DBConfig {
	return &config.DBConfig{
		Path: fmt.Sprintf("%v/db.sqlite3", dataDir),
	}
}

func fetchFeeds(cfg *config.Config, gp parser.GofeedParser, db db.DB) error {
	for _, cfgFeed := range cfg.Feeds {
		f := &feed.Feed{
			URL: cfgFeed.URL,
		}
		err := db.SaveFeed(f)
		if err != nil {
			return errors.Wrap(err, "failed to save feed")
		}

		for _, cfgTagName := range cfgFeed.Tags {
			t := &feed.Tag{
				Name: cfgTagName,
			}
			err := db.SaveTagToFeed(t, f)
			if err != nil {
				return errors.Wrap(err, "failed to save tag")
			}
		}

		gfeed, err := gp.ParseURL(cfgFeed.URL)
		if err != nil {
			return errors.Wrap(err, "failed to parse feed")
		}

		if len(gfeed.Items) == 0 {
			log.Warn().Msgf("%s feed is empty", cfgFeed.URL)
			continue
		}

		for _, gitem := range gfeed.Items {
			var i feed.Item
			err := i.FromGofeedItem(gitem)
			if err != nil {
				return errors.Wrap(err, "failed to initialize item")
			}

			// Don't insert item if there's an existing item with
			// the same author, title & link
			existingItem, err := db.MatchingItem(
				&feed.Item{
					Person: gofeed.Person{
						Name: i.Name,
					},
					Title: i.Title,
					Link:  i.Link,
				})
			if err != nil {
				return errors.Wrap(
					err,
					"failed to get matching item")
			}
			if existingItem.Title == i.Title {
				log.Info().Msgf("skipping: %s", i)
				continue
			}

			err = db.SaveItemToFeed(&i, f)
			if err != nil {
				return errors.Wrap(err, "failed to save item")
			}

			log.Debug().Msgf("inserted: %s", i)
		}
	}

	return nil
}

func watchFeeds(cfg *config.Config, dbCfg *config.DBConfig) {
	db, err := db.New(dbCfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create db client")
		return
	}

	defer db.Close()

	parser, err := parser.New()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create feed parser")
		return
	}

	fetchPeriod := cfg.FetchPeriod
	lastFetched, err := db.Timestamp()
	if err != nil || lastFetched == nil {
		log.Error().Err(err).Msg("Failed to get timestamp")
		return
	}

	for {
		wait := fetchPeriod - time.Since(*lastFetched)
		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
				break
			}
		}

		err := fetchFeeds(cfg, parser, db)
		if err != nil {
			log.Error().Err(err).Msg("Failed to fetch feeds")
		}

		now := time.Now()
		lastFetched = &now
		err = db.UpdateTimestamp(lastFetched)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update timestamp")
		}
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var tag string
	tagList, exists := queryParams["tag"]
	if exists {
		if len(tagList) > 0 {
			tag = tagList[0]
		}
	}

	db, err := db.New(dbConfig())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create db client")
		return
	}

	defer db.Close()

	t, err := template.ParseFiles("assets/index.html.tmpl")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse html template")
		return
	}

	cfg, err := appConfig()
	if err != nil {

		log.Error().Err(err).Msg("Failed to create config client")
		return
	}

	p, err := page.New(db, cfg.AppTitle, tag)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create page")
		return
	}

	err = t.Execute(w, p)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render html template")
	}
}

func hideHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New(dbConfig())
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

	var item feed.Item
	item.ID = uint(id)
	item.Hide = true

	// Using UpdateItem instead of SaveItem b/c it appears gorm.DB.Save
	// deletes the item for some reason..
	err = db.UpdateItem(&item)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update item")
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	s := struct {
		Test string
	}{
		Test: "TestString",
	}

	t, err := template.ParseFiles("assets/index2.html.tmpl")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse html template")
		return
	}

	err = t.Execute(w, s)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render html template")
	}
}

type S struct {
	Test string `json:"test"`
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New(dbConfig())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create db client")
		return
	}

	defer db.Close()

	items, err := db.AllItems()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items")
		return
	}

	// s := &S{
	// 	Test: "val",
	// }

	//text, err := json.Marshal(s)
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

func main() {
	if os.Getenv(envDebug) == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Create data dir if it doesn't exist
	_, err := os.Stat(dataDir)
	if err != nil && !os.IsNotExist(err) {
		log.Error().Err(err).Msg("Failed to stat data directory")
		return
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(dataDir, os.ModeDir)
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to create data directory")
		return
	}

	cfg, err := appConfig()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create config client")
		return
	}

	adb, err := db.New(dbConfig())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create db client")
		return
	}

	defer adb.Close()

	err = db.Migrate(adb)
	if err != nil {
		log.Error().Err(err).Msg("Failed to migrate db")
		return
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hide", hideHandler)
	http.HandleFunc("/test", testHandler)
	//http.HandleFunc("/api/v1/item", itemHandler) // TODO: replace /hide with this?
	http.HandleFunc("/api/v1/items", itemsHandler)

	go watchFeeds(cfg, dbConfig())

	err = http.ListenAndServe(":8080", nil)
	log.Error().Err(err).Msg("Server failed")
}
