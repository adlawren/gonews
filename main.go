package main

// TODO:
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
	"gonews/lib"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
	"gopkg.in/gormigrate.v1"
)

const (
	DataDirPath       = "/data/gonews"
	ConfDirPath       = ".config"
	TimestampFilePath = "DB_LAST_UPDATED"
)

type IndexTemplateParams struct {
	Title     string
	FeedItems []*lib.FeedItem
}

func migrateGormDB(gdb *gormDB) error {
	m := gormigrate.New(gdb.db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201908072046",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(&lib.FeedItem{}).Error; err != nil {
					return err
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				if err := tx.DropTable("feed_items").Error; err != nil {
					return err
				}

				return nil
			},
		},
		{
			ID: "201911051805",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(&lib.FeedItem{}).Error; err != nil {
					return err
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				if err := tx.Table("feed_items").DropColumn("tags").Error; err != nil {
					return err
				}

				return nil
			},
		},
	})

	return m.Migrate()
}

func loadConfig(appConfig lib.AppConfig, configPath, configName string) error {
	appConfig.SetConfigName(configName)
	appConfig.AddConfigPath(configPath)
	return appConfig.ReadInConfig()
}

type osFS struct{}

func (osfs *osFS) Stat(path string) (lib.FileInfo, error) {
	return os.Stat(path)
}

func (osfs *osFS) Open(path string) (lib.File, error) {
	return os.Open(path)
}

func (osfs *osFS) OpenFile(path string, flag int, perm os.FileMode) (lib.File, error) {
	return os.OpenFile(path, flag, perm)
}

type viperAppConfig struct{}

func (vac *viperAppConfig) SetConfigName(configName string) {
	viper.SetConfigName(configName)
}

func (vac *viperAppConfig) AddConfigPath(configPath string) {
	viper.AddConfigPath(configPath)
}

func (vac *viperAppConfig) ReadInConfig() error {
	return viper.ReadInConfig()
}

func (vac *viperAppConfig) Get(property string) interface{} {
	return viper.Get(property)
}

func (vac *viperAppConfig) GetString(property string) string {
	return viper.GetString(property)
}

var vac *viperAppConfig

func appConfig() *viperAppConfig {
	if vac == nil {
		vac = &viperAppConfig{}
	}

	return vac
}

type gormDB struct {
	db *gorm.DB
}

func (gdb *gormDB) FirstOrCreate(out interface{}, where ...interface{}) lib.DB {
	return &gormDB{db: gdb.db.FirstOrCreate(out, where...)}
}

func (gdb *gormDB) Find(out interface{}, where ...interface{}) lib.DB {
	return &gormDB{db: gdb.db.Find(out, where...)}
}

func (gdb *gormDB) Order(value interface{}, reorder ...bool) lib.DB {
	return &gormDB{db: gdb.db.Order(value, reorder...)}
}

func (gdb *gormDB) Close() error {
	return gdb.db.Close()
}

func (gdb *gormDB) Model(value interface{}) lib.DB {
	return &gormDB{db: gdb.db.Model(value)}
}

func (gdb *gormDB) Update(attrs ...interface{}) lib.DB {
	return &gormDB{db: gdb.db.Update(attrs...)}
}

func (gdb *gormDB) Where(query interface{}, args ...interface{}) lib.DB {
	return &gormDB{db: gdb.db.Where(query, args...)}
}

type gofeedURLParser struct {
	Parser *gofeed.Parser
}

func (gfp *gofeedURLParser) ParseURL(url string) (*gofeed.Feed, error) {
	return gfp.Parser.ParseURL(url)
}

// TODO: Use a singleton instead? How to manage .Close() call.. just put it in main()?
func appDB() lib.DB {
	db, err := gorm.Open("sqlite3", fmt.Sprintf("%v/db.sqlite3", DataDirPath))
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &gormDB{db: db}
}

func fetchFeedsMonitor() {
	osfs := &osFS{}
	timestampFile := &lib.TimestampFile{Path: fmt.Sprintf("%v/%v", DataDirPath, TimestampFilePath)}
	feedFetcher := &lib.DefaultFeedFetcher{}
	feedParser := &gofeedURLParser{Parser: gofeed.NewParser()}

	for {
		if err := lib.FetchFeedsAfterDelay(appConfig(), osfs, timestampFile, feedFetcher, feedParser, appDB()); err != nil {
			fmt.Printf("Failed to fetch feeds: %v\n", err)
		}
	}
}

func getOrderedFeedItems(db lib.DB, tag string) []*lib.FeedItem {
	dbOrderedByPublished := db.Order("published DESC", true)

	var feedItems []*lib.FeedItem
	if tag != "" {
		likeString := fmt.Sprintf("%%<%v>%%", tag)
		dbOrderedByPublished.Where("tags LIKE ?", likeString).Find(&feedItems, &lib.FeedItem{})
	} else {
		dbOrderedByPublished.Find(&feedItems, &lib.FeedItem{})
	}

	return feedItems
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

	feedItems := getOrderedFeedItems(db, tag)

	if t, err := template.ParseFiles("index.html.tmpl"); err != nil {
		log.Fatal(err)
	} else {
		t.Execute(w, &IndexTemplateParams{
			Title:     appConfig().GetString("homepage_title"),
			FeedItems: feedItems,
		})
	}
}

func hideHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	db := appDB()
	defer db.Close()

	if id, err := strconv.ParseUint(r.PostFormValue("ID"), 10, 64); err == nil {
		var feedItem lib.FeedItem
		feedItem.ID = uint(id)
		db.Model(&feedItem).Update("Hide", true)
	}

	http.Redirect(w, r, "/", 303)
}

func main() {
	// Create data dir if it doesn't exist
	if _, err := os.Stat(DataDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(DataDirPath, os.ModeDir); err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}

	if err := loadConfig(appConfig(), ConfDirPath, "config"); err != nil {
		log.Fatal(err)
	}

	// TODO: add a config param to enable logging
	// db.LogMode(true)

	db := appDB()
	defer db.Close()

	if err := migrateGormDB(db.(*gormDB)); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hide", hideHandler)

	go fetchFeedsMonitor()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
