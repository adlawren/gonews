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

		var existingFeed feed.Feed
		err := db.Find(&existingFeed, query.NewClause("where url = ?", f.URL))
		if err != nil && !errors.Is(err, query.ErrModelNotFound) {
			return errors.Wrap(err, "failed to get matching feed")
		}

		err = db.Save(f)
		if err != nil {
			return errors.Wrap(err, "failed to save feed")
		}

		log.Debug().Msgf("inserted: %s", f)

		for _, cfgTagName := range cfgFeed.Tags {
			t := &feed.Tag{
				Name:   cfgTagName,
				FeedID: f.ID,
			}

			var existingTag feed.Tag
			err = db.Find(&existingTag, query.NewClause("where name = ?", t.Name))
			if err != nil && !errors.Is(err, query.ErrModelNotFound) {
				return errors.Wrap(
					err,
					"failed to get matching tag")
			}

			err = db.Save(t)
			if err != nil {
				return errors.Wrap(err, "failed to save tag")
			}

			log.Debug().Msgf("inserted: %s", t)
		}
	}

	return nil
}

func fetchFeeds(db db.DB, p parser.Parser) error {
	var feeds []*feed.Feed
	err := db.All(&feeds)
	if err != nil {
		return errors.Wrap(err, "failed to get feeds")
	}

	for _, f := range feeds {
		items, err := p.ParseURL(f.URL)
		if err != nil {
			return errors.Wrap(err, "failed to parse feed")
		}

		if len(items) == 0 {
			log.Warn().Msgf("%s feed is empty", f.URL)
			continue
		}

		if f.FetchLimit != 0 && uint(len(items)) > f.FetchLimit {
			items = items[:f.FetchLimit]
		}

		for _, item := range items {
			// Don't insert item if there's an existing item with
			// the same author, title & link
			var existingItem feed.Item
			err = db.Find(&existingItem, query.NewClause("where link = ?", item.Link))
			if err != nil && !errors.Is(err, query.ErrModelNotFound) {
				return errors.Wrap(
					err,
					"failed to get matching item")
			} else if err == nil {
				log.Info().Msgf("skipping: %s", item)
				continue
			}

			item.FeedID = f.ID

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
		err = db.Save(&lastFetched)
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
			var f feed.Feed
			err = db.Find(&f, query.NewClause("where url = ?", feedCfg.URL))
			if err != nil {
				return errors.Wrap(err, "failed to get matching feed")
			}

			var items []*feed.Item
			err = db.FindAll(&items, query.NewClause("where feed_id = ?", f.ID))
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
		err = db.Save(&lastAutoDismissed)
		if err != nil {
			return errors.Wrap(err, "failed to update timestamp")
		}
	}

	return nil
}
