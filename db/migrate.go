package db

import (
	"fmt"
	"gonews/feed"
	"gonews/legacy"
	"gonews/timestamp"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"gopkg.in/gormigrate.v1"
)

// Migrate applies migrations to the given database
func Migrate(db DB) error {
	gdb, ok := db.(*gormDB)
	if !ok {
		return errors.Errorf("db type: (%T), expected gorm.DB", db)
	}

	var migrations []*gormigrate.Migration
	migrations = append(migrations, &gormigrate.Migration{
		ID: "201908072046",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&legacy.FeedItem{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("feed_items").Error
		},
	})

	migrations = append(migrations, &gormigrate.Migration{
		ID: "201911051805",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&legacy.FeedItem{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			// Can't drop "tags" column using SQlite
			// See https://stackoverflow.com/questions/8442147/how-to-delete-or-add-column-in-sqlite/8442173#8442173
			// return tx.Model(
			// 	&legacy.FeedItem{}).DropColumn("tags").Error
			return nil
		},
	})

	migrations = append(migrations, &gormigrate.Migration{
		ID: "202003062117",
		Migrate: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(&timestamp.Timestamp{}).Error
			if err != nil {
				return err
			}

			err = tx.AutoMigrate(&feed.Feed{}).Error
			if err != nil {
				return err
			}

			err = tx.AutoMigrate(&feed.Tag{}).Error
			if err != nil {
				return err
			}

			err = tx.AutoMigrate(&feed.Item{}).Error
			if err != nil {
				return err
			}

			var feedItems []*legacy.FeedItem
			err = tx.Find(&feedItems, &legacy.FeedItem{}).Error
			if err != nil {
				return err
			}

			for _, f := range feedItems {
				var existingFeed feed.Feed
				err = tx.FirstOrCreate(
					&existingFeed,
					&feed.Feed{URL: f.URL}).Error
				if err != nil {
					return err
				}

				tags := strings.Split(f.Tags, ",")

				for _, t := range tags {
					t = strings.TrimRight(t, ">")
					t = strings.TrimLeft(t, "<")

					var existingTag feed.Tag
					err = tx.FirstOrCreate(
						&existingTag,
						&feed.Tag{
							Name:   t,
							FeedID: f.ID,
						}).Error
					if err != nil {
						return err
					}
				}

				var existingItem feed.Item
				err = tx.FirstOrCreate(
					&existingItem,
					&feed.Item{
						Title:       f.Title,
						Description: f.Description,
						Link:        f.Link,
						Published:   f.Published,
						Hide:        f.Hide,
						Person: gofeed.Person{
							Name:  f.AuthorName,
							Email: f.AuthorEmail,
						},
						FeedID: existingFeed.ID,
					}).Error
				if err != nil {
					return err
				}
			}

			return tx.DropTable("feed_items").Error
		},
		Rollback: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(&legacy.FeedItem{}).Error
			if err != nil {
				return err
			}

			var feeds []*feed.Feed
			err = tx.Find(&feeds, &feed.Feed{}).Error
			if err != nil {
				return err
			}

			for _, f := range feeds {
				var relatedTags []*feed.Tag
				err = tx.Model(f).Related(&relatedTags).Error
				if err != nil {
					return err
				}

				tagNames := make(
					[]string,
					len(relatedTags),
					len(relatedTags))
				for i := 0; i < len(relatedTags); i++ {
					tagNames[i] = fmt.Sprintf(
						"<%s>",
						relatedTags[i].Name)
				}
				tags := strings.Join(tagNames, ",")

				var relatedItems []*feed.Item
				err = tx.Model(f).Related(&relatedItems).Error
				if err != nil {
					return err
				}

				for _, i := range relatedItems {
					var existingFeedItem legacy.FeedItem
					fi := &legacy.FeedItem{
						Title:       i.Title,
						Description: i.Description,
						Link:        i.Link,
						URL:         f.URL,
						Published:   i.Published,
						Hide:        i.Hide,
						AuthorName:  i.Name,
						AuthorEmail: i.Email,
						Tags:        tags,
					}
					err = tx.FirstOrCreate(
						&existingFeedItem,
						fi).Error
					if err != nil {
						return err
					}
				}
			}

			err = tx.DropTable("timestamps").Error
			if err != nil {
				return err
			}

			err = tx.DropTable("feeds").Error
			if err != nil {
				return err
			}

			err = tx.DropTable("tags").Error
			if err != nil {
				return err
			}

			return tx.DropTable("items").Error
		},
	})

	m := gormigrate.New(gdb.db, gormigrate.DefaultOptions, migrations)
	return errors.Wrap(m.Migrate(), "failed to migrate gorm DB")
}
