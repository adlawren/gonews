package main

// TODO:
// - Don't use naked errors
// - Lower case error messages, except logs
// - Throw errors in methods if pointers are nil
// - Add support for labels/groups in the config file
//   - And support for use of these in query parameters; filter output accordingly
// - Create a new table for the feed configs from the config file
//   - Replace the 'feed url' field in the feed items with foreign keys, which reference entries in the other table
//   - Each time the application is started, reload the config and update the database, if it has changed (maybe store the hash of the file in the db too)
//     - Create a new monitor thread, to watch the config for changes
// - Rename lib package, extract code to separate packages
// - Update config
//   - Item limit per page
//   - Add params to limit the number of feed items retained in the database
//     - And/or 'max retention time'
// - Display a 'loading' spinner/message when new feed items are being retrieved
// - Add a manual 'refresh' button which prematurely retrieves the feeds (and updates the 'feed-last-retrieved' time in the db)
// - Add support for a http-param-based filter; i.e. <base url>?url=<something>&...
// - Add support for a 'hidden' field in the feed config
//   - Can hide feeds by default, by use a filter query ^ to see them

import (
	"fmt"
	"gonews/config"
	"gonews/db"
	gndb "gonews/db" // TODO: rename
	"gonews/fs"
	"gonews/lib"
	"gonews/list"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mmcdole/gofeed"
)

const (
	//dataDirPath       = "/data/gonews" // TODO: uncomment
	dataDirPath       = ".data/"
	confDirPath       = ".config"
	timestampFilePath = "DB_LAST_UPDATED"
)

func loadConfig(appConfig config.Config, configPath, configName string) error {
	appConfig.SetConfigName(configName)
	appConfig.AddConfigPath(configPath)
	return appConfig.ReadInConfig()
}

var cfg config.Config

func appConfig() config.Config {
	if cfg == nil {
		cfg = config.FromViperConfig()
	}

	return cfg
}

type gofeedURLParser struct {
	Parser *gofeed.Parser
}

func (gfp *gofeedURLParser) ParseURL(url string) (*gofeed.Feed, error) {
	return gfp.Parser.ParseURL(url)
}

// TODO: Use a singleton instead? How to manage .Close() call.. just put it in main()?
func appDB() db.DB {
	gdb, err := gorm.Open("sqlite3", fmt.Sprintf("%v/db.sqlite3", dataDirPath))
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return db.FromGormDB(gdb)
}

func fetchFeedsMonitor() {
	osfs := fs.FromOSFS()
	timestampFile := &lib.TimestampFile{
		Path: fmt.Sprintf("%v/%v", dataDirPath, timestampFilePath),
	}
	feedFetcher := &lib.DefaultFeedFetcher{}
	feedParser := &gofeedURLParser{Parser: gofeed.NewParser()}

	for {
		if err := lib.FetchFeedsAfterDelay(appConfig(), osfs, timestampFile, feedFetcher, feedParser, appDB()); err != nil {
			fmt.Printf("Failed to fetch feeds: %v\n", err)
		}
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var tag string
	if tagList, exists := queryParams["tag"]; exists {
		if len(tagList) > 0 {
			tag = tagList[0]
		}
	}

	db := appDB()
	defer db.Close()

	if t, err := template.ParseFiles("index.html.tmpl"); err != nil {
		log.Fatal(err)
	} else {
		title := appConfig().GetString("homepage_title")
		t.Execute(w, list.New(db, title, tag))
	}
}

func hideHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	db := appDB()
	defer db.Close()

	id, err := strconv.ParseUint(r.PostFormValue("ID"), 10, 64)
	if err == nil {
		var item gndb.Item
		item.ID = uint(id)
		db.Model(&item).Update("Hide", true)
	}

	http.Redirect(w, r, "/", 303)
}

func main() {
	adb := appDB()
	defer adb.Close()

	if err := db.MigrateGormDB(adb); err != nil {
		log.Fatal(err)
	}

	return

	// Create data dir if it doesn't exist
	if _, err := os.Stat(dataDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDirPath, os.ModeDir); err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}

	if err := loadConfig(appConfig(), confDirPath, "config"); err != nil {
		log.Fatal(err)
	}

	// TODO: add a config param to enable logging
	// db.LogMode(true)

	//adb := appDB()
	adb = appDB()
	defer adb.Close()

	if err := db.MigrateGormDB(adb); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hide", hideHandler)

	go fetchFeedsMonitor()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
