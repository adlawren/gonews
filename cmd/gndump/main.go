package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gonews/config"
	"gonews/db"
	"gonews/feed"

	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog/log"
)

func main() {
	pingDB := flag.Bool("ping-db", false, "ping DB")
	showFeeds := flag.Bool("feeds", false, "show feeds")
	showTags := flag.Bool("tags", false, "show tags")
	showItems := flag.Bool("items", false, "show items")
	showFeedsFromTag := flag.String("feeds-from-tag", "", "show feeds from tag name")
	showItemsFromFeed := flag.Uint("items-from-feed", 0, "show items from feed ID")

	flag.Parse()

	db, err := db.New(&config.DBConfig{DSN: "file:/data/gonews/db.sqlite3"})
	defer db.Close()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create DB client")
		return
	}

	if *pingDB {
		err = db.Ping()
		if err != nil {
			log.Error().Err(err).Msg("Ping failed")
			return
		}
		fmt.Println("Ping succeeded")
	}

	if *showFeeds {
		feeds, err := db.AllFeeds()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get feeds")
			return
		}

		for _, feed := range feeds {
			feedBytes, err := json.Marshal(feed)
			if err != nil {
				break
			}
			fmt.Println(string(feedBytes[:]))
		}
	}

	if *showTags {
		tags, err := db.AllTags()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get tags")
			return
		}

		for _, tag := range tags {
			tagBytes, err := json.Marshal(tag)
			if err != nil {
				break
			}
			fmt.Println(string(tagBytes[:]))
		}
	}

	if *showItems {
		items, err := db.AllItems()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get items")
			return
		}

		for _, item := range items {
			itemBytes, err := json.Marshal(item)
			if err != nil {
				break
			}
			fmt.Println(string(itemBytes[:]))
		}
	}

	if len(*showFeedsFromTag) > 0 {
		tagName := *showFeedsFromTag
		feeds, err := db.FeedsFromTag(&feed.Tag{Name: tagName})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get feeds")
			return
		}

		for _, feed := range feeds {
			feedBytes, err := json.Marshal(feed)
			if err != nil {
				break
			}
			fmt.Println(string(feedBytes[:]))
		}
	}

	if *showItemsFromFeed != 0 {
		feedID := *showItemsFromFeed
		items, err := db.ItemsFromFeed(&feed.Feed{
			Model: gorm.Model{ID: feedID},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get items")
			return
		}

		for _, item := range items {
			itemBytes, err := json.Marshal(item)
			if err != nil {
				break
			}
			fmt.Println(string(itemBytes[:]))
		}
	}
}
