package main

// TODO:
// - Add support for a http-param-based filter; i.e. <base url>?date=<something>&...
//   - Ideally, you support parameters for each field in the item struct..
// - Add CSS; Bootstrap

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	// "html"
	"github.com/spf13/viper"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type FeedItem struct {
	Author      *gofeed.Person
	Title       string
	Description string
	Link        string
	Published   *time.Time
	Url         *url.URL
}

func convertToFeedItem(feedUrl string, gfi *gofeed.Item) *FeedItem {
	feedItemUrl, _ := url.Parse(feedUrl)
	return &FeedItem{
		Author:      gfi.Author,
		Title:       gfi.Title,
		Description: gfi.Description,
		Link:        gfi.Link,
		Published:   gfi.PublishedParsed,
		Url:         feedItemUrl,
	}
}

func convertToFeedItems(feedUrl string, goFeedItems []*gofeed.Item) []*FeedItem {
	var feedItems []*FeedItem
	for _, gfi := range goFeedItems {
		feedItems = append(feedItems, convertToFeedItem(feedUrl, gfi))
	}

	return feedItems
}

type FeedItemSlice []*FeedItem

func (fis FeedItemSlice) Len() int {
	return len(fis)
}

func (fis FeedItemSlice) Less(i, j int) bool {
	iTime := fis[i].Published
	jTime := fis[j].Published

	return iTime.After(*jTime)
}

func (fis FeedItemSlice) Swap(i, j int) {
	fis[i], fis[j] = fis[j], fis[i]
}

type IndexTemplateParams struct {
	Title    string
	ItemList FeedItemSlice
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error: failed to read config file")
		return
	}

	var allFeedItems FeedItemSlice
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

		var feedItemLimit int64
		if !exists {
			feedItemLimit = viper.Get("item_limit").(int64)
		} else {
			feedItemLimit = feedItemLimitInterface.(int64)
		}

		if nextFeed, err := fp.ParseURL(feedUrl); err != nil {
			fmt.Printf("Warning: could not retrieve feed from %v\n", feedUrl)
		} else {
			allFeedItems = append(
				allFeedItems,
				convertToFeedItems(
					feedUrl,
					nextFeed.Items[:feedItemLimit])...)

		}
	}

	sort.Sort(allFeedItems)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: parse url params, filter items
		//fmt.Println(r.URL.Query())

		if t, err := template.ParseFiles("index.html"); err != nil {
			fmt.Println("Error: failed to parse template")
		} else {
			t.Execute(w, &IndexTemplateParams{
				Title:    viper.GetString("homepage_title"),
				ItemList: allFeedItems,
			})
		}
	})

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
