package lib

import (
	"context"
	"errors"
	"fmt"
	"gonews/config"
	"gonews/db"
	"gonews/db/orm/query"
	"gonews/db/orm/query/clause"
	"gonews/feed"
	"gonews/parser"
	"gonews/timestamp"
	"time"

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
		err := db.Find(&existingFeed, clause.Where("url = ?", f.URL))
		if err != nil && !errors.Is(err, query.ErrModelNotFound) {
			return fmt.Errorf("failed to get matching feed: %w", err)
		}

		err = db.Save(f)
		if err != nil {
			return fmt.Errorf("failed to save feed: %w", err)
		}

		log.Debug().Msgf("inserted: %s", f)

		for _, cfgTagName := range cfgFeed.Tags {
			t := &feed.Tag{
				Name:   cfgTagName,
				FeedID: f.ID,
			}

			var existingTag feed.Tag
			err = db.Find(&existingTag, clause.Where("name = ?", t.Name))
			if err != nil && !errors.Is(err, query.ErrModelNotFound) {
				return fmt.Errorf("failed to get matching tag: %w", err)
			}

			err = db.Save(t)
			if err != nil {
				return fmt.Errorf("failed to save tag: %w", err)
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
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	for _, f := range feeds {
		items, err := p.ParseURL(f.URL)
		if err != nil {
			return fmt.Errorf("failed to parse feed: %w", err)
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
			err = db.Find(&existingItem, clause.Where("link = ?", item.Link))
			if err != nil && !errors.Is(err, query.ErrModelNotFound) {
				return fmt.Errorf("failed to get matching item: %w", err)
			} else if err == nil {
				log.Info().Msgf("skipping: %s", item)
				continue
			}

			item.FeedID = f.ID

			err = db.Save(item)
			if err != nil {
				return fmt.Errorf("failed to save item: %w", err)
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
		return fmt.Errorf("failed to create db client: %w", err)
	}

	defer db.Close()

	parser, err := parser.New()
	if err != nil {
		return fmt.Errorf("failed to create feed parser: %w", err)
	}

	fetchPeriod := cfg.FetchPeriod
	var lastFetched timestamp.Timestamp
	err = db.Find(&lastFetched, clause.Where("name = 'feeds_fetched_at'"))
	if errors.Is(err, query.ErrModelNotFound) {
		lastFetched = timestamp.Timestamp{Name: "feeds_fetched_at"}

	} else if err != nil {
		return fmt.Errorf("failed to get matching timestamp: %w", err)
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
			return fmt.Errorf("failed to fetch feeds: %w", err)
		}

		lastFetched.T = time.Now()
		err = db.Save(&lastFetched)
		if err != nil {
			return fmt.Errorf("failed to update timestamp: %w", err)
		}
	}

	return nil
}

// Periodically hide items older than the configured duration
func AutoDismissItems(ctx context.Context, cfg *config.Config, dbCfg *config.DBConfig) error {
	db, err := db.New(dbCfg)
	if err != nil {
		return fmt.Errorf("failed to create db client: %w", err)
	}

	defer db.Close()

	autoDismissPeriod := cfg.AutoDismissPeriod
	var lastAutoDismissed timestamp.Timestamp
	err = db.Find(&lastAutoDismissed, clause.Where("name = 'auto_dismissed_at'"))
	if errors.Is(err, query.ErrModelNotFound) {
		lastAutoDismissed = timestamp.Timestamp{Name: "auto_dismissed_at"}
	} else if err != nil {
		return fmt.Errorf("failed to get matching timestamp: %w", err)
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
			err = db.Find(&f, clause.Where("url = ?", feedCfg.URL))
			if err != nil {
				return fmt.Errorf("failed to get matching feed: %w", err)
			}

			var items []*feed.Item
			err = db.FindAll(&items, clause.Where("feed_id = ?", f.ID))
			if err != nil {
				return fmt.Errorf("failed to get items from feed: %w", err)
			}

			for _, item := range items {
				if time.Now().Before(item.CreatedAt.Add(feedCfg.AutoDismissAfter)) {
					continue
				}

				item.Hide = true
				err := db.Save(item)
				if err != nil {
					return fmt.Errorf("failed to save item: %w", err)
				}
			}
		}

		lastAutoDismissed.T = time.Now()
		err = db.Save(&lastAutoDismissed)
		if err != nil {
			return fmt.Errorf("failed to update timestamp: %w", err)
		}
	}

	return nil
}
