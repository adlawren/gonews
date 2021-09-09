package lib

import (
	"context"
	"gonews/config"
	"gonews/db"
	"gonews/db/orm/query"
	"gonews/feed"
	"gonews/parser"
	"gonews/timestamp"
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

		if feed.FetchLimit != 0 && uint(len(items)) > feed.FetchLimit {
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
func WatchFeeds(ctx context.Context, cfg *config.Config, dbCfg *config.DBConfig) error {
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
	var lastFetched timestamp.Timestamp
	err = db.Find(&lastFetched, query.NewClause("where name = 'feeds_fetched_at'"))
	if errors.Is(err, query.ErrModelNotFound) {
		lastFetched = timestamp.Timestamp{Name: "feeds_fetched_at"}

	} else if err != nil {
		return errors.Wrap(err, "failed to get matching timestamp")
	}

	for {
		wait := fetchPeriod - time.Since(lastFetched.T)
		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
				break
			case <-ctx.Done():
				return nil
			}
		}

		err := fetchFeeds(db, parser)
		if err != nil {
			return errors.Wrap(err, "failed to fetch feeds")
		}

		lastFetched.T = time.Now()
		err = db.SaveTimestamp(&lastFetched)
		if err != nil {
			return errors.Wrap(err, "failed to update timestamp")
		}
	}

	return nil
}

// Periodically hide items older than the configured duration
func AutoDismissItems(ctx context.Context, cfg *config.Config, dbCfg *config.DBConfig) error {
	db, err := db.New(dbCfg)
	if err != nil {
		return errors.Wrap(err, "failed to create db client")
	}

	defer db.Close()

	autoDismissPeriod := cfg.AutoDismissPeriod
	var lastAutoDismissed timestamp.Timestamp
	err = db.Find(&lastAutoDismissed, query.NewClause("where name = 'auto_dismissed_at'"))
	if errors.Is(err, query.ErrModelNotFound) {
		lastAutoDismissed = timestamp.Timestamp{Name: "auto_dismissed_at"}
	} else if err != nil {
		return errors.Wrap(err, "failed to get matching timestamp")
	}

	for {
		wait := autoDismissPeriod - time.Since(lastAutoDismissed.T)
		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
				break
			case <-ctx.Done():
				return nil
			}
		}

		for _, feedCfg := range cfg.Feeds {
			feed, err := db.MatchingFeed(&feed.Feed{URL: feedCfg.URL})
			if err != nil {
				return errors.Wrap(err, "failed to get matching feed")
			}

			items, err := db.ItemsFromFeed(feed)
			if err != nil {
				return errors.Wrap(err, "failed to get items from feed")
			}

			for _, item := range items {
				if time.Now().Before(item.CreatedAt.Add(feedCfg.AutoDismissAfter)) {
					continue
				}

				item.Hide = true
				err := db.SaveItem(item)
				if err != nil {
					return errors.Wrap(err, "failed to save item")
				}
			}
		}

		lastAutoDismissed.T = time.Now()
		err = db.SaveTimestamp(&lastAutoDismissed)
		if err != nil {
			return errors.Wrap(err, "failed to update timestamp")
		}
	}

	return nil
}
