package lib

import (
	"gonews/config"
	"gonews/db"
	"gonews/feed"
	"gonews/parser"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Insert any nonexistent feeds & tags from the config into the database
func InsertMissingFeeds(cfg *config.Config, db db.DB) error {
	for _, cfgFeed := range cfg.Feeds {
		f := &feed.Feed{
			URL:        cfgFeed.URL,
			FetchLimit: cfgFeed.FetchLimit,
		}

		existingFeed, err := db.MatchingFeed(f)
		if err != nil {
			return errors.Wrap(err, "failed to get matching feed")
		}
		if existingFeed != nil {
			continue
		}

		err = db.SaveFeed(f)
		if err != nil {
			return errors.Wrap(err, "failed to save feed")
		}

		log.Debug().Msgf("inserted: %s", f)

		for _, cfgTagName := range cfgFeed.Tags {
			t := &feed.Tag{
				Name:   cfgTagName,
				FeedID: f.ID,
			}

			existingTag, err := db.MatchingTag(t)
			if err != nil {
				return errors.Wrap(
					err,
					"failed to get matching tag")
			}
			if existingTag != nil {
				continue
			}

			err = db.SaveTag(t)
			if err != nil {
				return errors.Wrap(err, "failed to save tag")
			}

			log.Debug().Msgf("inserted: %s", t)
		}
	}

	return nil
}

func fetchFeeds(db db.DB, p parser.Parser) error {
	feeds, err := db.Feeds()
	if err != nil {
		return errors.Wrap(err, "failed to get feeds")
	}

	for _, feed := range feeds {
		items, err := p.ParseURL(feed.URL)
		if err != nil {
			return errors.Wrap(err, "failed to parse feed")
		}

		if len(items) == 0 {
			log.Warn().Msgf("%s feed is empty", feed.URL)
			continue
		}

		if feed.FetchLimit != 0 {
			items = items[:feed.FetchLimit]
		}

		for _, item := range items {
			// Don't insert item if there's an existing item with
			// the same author, title & link
			existingItem, err := db.MatchingItem(item)
			if err != nil {
				return errors.Wrap(
					err,
					"failed to get matching item")
			}

			if existingItem != nil {
				log.Info().Msgf("skipping: %s", item)
				continue
			}

			item.FeedID = feed.ID

			err = db.SaveItem(item)
			if err != nil {
				return errors.Wrap(err, "failed to save item")
			}

			log.Debug().Msgf("inserted: %s", item)
		}
	}

	return nil
}

// Periodically parse feeds from the DB and insert any nonexistent items,
// subject to the fetch period from the given config
func WatchFeeds(cfg *config.Config, dbCfg *config.DBConfig) error {
	db, err := db.New(dbCfg)
	if err != nil {
		return errors.Wrap(err, "failed to create db client")
	}

	defer db.Close()

	parser, err := parser.New()
	if err != nil {
		return errors.Wrap(err, "failed to create feed parser")
	}

	fetchPeriod := cfg.FetchPeriod
	lastFetched, err := db.Timestamp()
	if err != nil || lastFetched == nil {
		return errors.Wrap(err, "failed to get timestamp")
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

		err := fetchFeeds(db, parser)
		if err != nil {
			return errors.Wrap(err, "failed to fetch feeds")
		}

		now := time.Now()
		lastFetched = &now
		err = db.SaveTimestamp(lastFetched)
		if err != nil {
			return errors.Wrap(err, "failed to update timestamp")
		}
	}

	return nil
}
