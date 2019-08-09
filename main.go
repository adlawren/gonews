package main

// TODO:
// - Don't insert duplicate items
// - Update config
//   - Add params to limit the number of feed items retained in the database
//     - And/or 'max retention time'
//   - 'feed retrieval frequency'
// - Add database entries to track feed retrieval times
// - Display a 'loading' spinner/message when new feed items are being retrieved
// - Add a manual 'refresh' button which prematurely retrieves the feeds (and updates the 'feed-last-retrieved' time in the db)
// - Implement the feed retrieval function as a goroutine; run it without blocking the main app
// - Add support for a http-param-based filter; i.e. <base url>?date=<something>&...
//   - Ideally, you support parameters for each field in the item struct..
// - Add support for a 'hidden' field in the feed config
//   - Can hide feeds by default, by use a filter query ^ to see them
// - Add CSS; Bootstrap
// - Create/use hidden data dir

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
	"gopkg.in/gormigrate.v1"
	"html/template"
	"log"
	"net/http"
	"time"
)

type FeedItem struct {
	gorm.Model
	Title       string
	Description string
	Link        string
	Published   *time.Time
	Url         string
	AuthorName  string
	AuthorEmail string
}

func convertToFeedItem(feedUrl string, goFeedItem *gofeed.Item) *FeedItem {
	var goFeedAuthorName string
	var goFeedAuthorEmail string
	if goFeedItem.Author != nil {
		goFeedAuthorName = goFeedItem.Author.Name
		goFeedAuthorEmail = goFeedItem.Author.Email
	}

	return &FeedItem{
		Title:       goFeedItem.Title,
		Description: goFeedItem.Description,
		Link:        goFeedItem.Link,
		Published:   goFeedItem.PublishedParsed,
		Url:         feedUrl,
		AuthorName:  goFeedAuthorName,
		AuthorEmail: goFeedAuthorEmail,
	}
}

type IndexTemplateParams struct {
	Title     string
	FeedItems []*FeedItem
}

func migrateDatabase(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201908072046",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(&FeedItem{}).Error; err != nil {
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

func loadConfig(configPath, configName string) error {
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)
	return viper.ReadInConfig()
}

// TODO: see if there's a way to avoid the redundancy - the feedUrl is in the feedConfigMap too
func processGoFeedItem(db *gorm.DB, goFeedItem *gofeed.Item, feedUrl string, feedConfigMap map[string]interface{}) {
	fi := convertToFeedItem(feedUrl, goFeedItem)

	db.Create(fi)
}

func main() {
	if err := loadConfig(".", "config"); err != nil {
		log.Fatal(err)
	}

	// TODO: add a config param to enable logging
	// db.LogMode(true)

	db, err := gorm.Open("sqlite3", fmt.Sprintf("%v.sqlite3", viper.GetString("db_name")))
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err := migrateDatabase(db); err != nil {
		log.Fatal(err)
	}

	fp := gofeed.NewParser()

	feedConfigs := viper.Get("feeds").([]interface{})
	for _, feedConfig := range feedConfigs {
		feedConfigMap := feedConfig.(map[string]interface{})
		feedUrl, exists := feedConfigMap["url"].(string)
		if !exists {
			fmt.Println("Error: feed config must contain url")
			return
		}

		feedItemLimitInterface, exists := feedConfigMap["item_limit"]

		var feedItemLimit int
		if !exists {
			feedItemLimit = int(viper.Get("item_limit").(int64))
		} else {
			feedItemLimit = int(feedItemLimitInterface.(int64))
		}

		if nextFeed, err := fp.ParseURL(feedUrl); err != nil {
			fmt.Printf("Warning: could not retrieve feed from %v\n", feedUrl)
		} else if len(nextFeed.Items) < 1 {
			fmt.Printf("Warning: %v feed is empty\n", feedUrl)
		} else {
			if len(nextFeed.Items) < feedItemLimit {
				fmt.Printf("Warning: truncating item_limit; not enough items in %v feed\n", feedUrl)
				feedItemLimit = len(nextFeed.Items)
			}

			for _, goFeedItem := range nextFeed.Items[:feedItemLimit] {
				processGoFeedItem(db, goFeedItem, feedUrl, feedConfigMap)
			}
		}
	}

	dbOrderedByPublished := db.Order("published DESC", true)

	var feedItems []*FeedItem
	dbOrderedByPublished.Find(&feedItems, &FeedItem{})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: parse url params, filter items
		//fmt.Println(r.URL.Query())

		if t, err := template.ParseFiles("index.html"); err != nil {
			log.Fatal(err)
		} else {
			t.Execute(w, &IndexTemplateParams{
				Title:     viper.GetString("homepage_title"),
				FeedItems: feedItems,
			})
		}
	})

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
