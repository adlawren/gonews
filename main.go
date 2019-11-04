package main

// TODO:
// - Add support for labels/groups in the config file
//   - And support for use of these in query parameters; filter output accordingly
// - Rename lib package, extract code to separate packages
// - Create config template file, add instructions to readme; it would need to be copied to the data directory
// - Create/use hidden data dir
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
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
	"gonews/lib"
	"gopkg.in/gormigrate.v1"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
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

type gofeedURLParser struct {
	Parser *gofeed.Parser
}

func (gfp *gofeedURLParser) ParseURL(url string) (*gofeed.Feed, error) {
	return gfp.Parser.ParseURL(url)
}

// TODO: Use a singleton instead? How to manage .Close() call.. just put it in main()?
func appDB() lib.DB {
	db, err := gorm.Open("sqlite3", fmt.Sprintf("%v.sqlite3", appConfig().GetString("db_name")))
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &gormDB{db: db}
}

func fetchFeedsMonitor() {
	osfs := &osFS{}
	timestampFile := &lib.TimestampFile{Path: TimestampFilePath}
	feedFetcher := &lib.DefaultFeedFetcher{}
	feedParser := &gofeedURLParser{Parser: gofeed.NewParser()}

	for {
		if err := lib.FetchFeedsAfterDelay(appConfig(), osfs, timestampFile, feedFetcher, feedParser, appDB()); err != nil {
			fmt.Printf("Failed to fetch feeds: %v\n", err)
		}
	}
}

func getOrderedFeedItems(db lib.DB) []*lib.FeedItem {
	dbOrderedByPublished := db.Order("published DESC", true)

	var feedItems []*lib.FeedItem
	dbOrderedByPublished.Find(&feedItems, &lib.FeedItem{})

	return feedItems
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: parse url params, filter items
	//fmt.Println(r.URL.Query())

	db := appDB()
	defer db.Close()

	feedItems := getOrderedFeedItems(db)

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
	if err := loadConfig(appConfig(), ".", "config"); err != nil {
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
