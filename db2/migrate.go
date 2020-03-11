package db2

import (
	"gonews/legacy"

	"github.com/jinzhu/gorm"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"gopkg.in/gormigrate.v1"
)

func MigrateGormDB(db DB) error {
	gdb, ok := db.(*gormDB)
	if !ok {
		return errors.Errorf("db type: (%T), expected gorm.DB", db)
	}

	migrations = append(migrations, &gormigrate.Migration{
		ID: "201908072046",
		Migrate: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(&legacy.FeedItem{}).Error
			if err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			err := tx.DropTable("feed_items").Error
			if err != nil {
				return err
			}

			return nil
		},
	})

	migrations = append(migrations, &gormigrate.Migration{
		ID: "202003062117",
		Migrate: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(&Feed{}).Error
			if err != nil {
				return err
			}

			err = tx.AutoMigrate(&Tag{}).Error
			if err != nil {
				return err
			}

			err = tx.AutoMigrate(&Item{}).Error
			if err != nil {
				return err
			}

			var feedItems []*legacy.FeedItem
			err = tx.Find(&feedItems, &legacy.FeedItem{}).Error
			if err != nil {
				return err
			}

			for _, f := range feedItems {
				var existingFeed Feed
				err = tx.FirstOrCreate(
					&existingFeed,
					&Feed{URL: f.Url}).Error
				if err != nil {
					return err
				}

				var existingItem Item
				err = tx.FirstOrCreate(
					&existingItem,
					&Item{
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

			err = tx.DropTable("feed_items").Error
			if err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(&legacy.FeedItem{}).Error
			if err != nil {
				return err
			}

			var feeds []*Feed
			err = tx.Find(&feeds, &Feed{}).Error
			if err != nil {
				return err
			}

			for _, f := range feeds {
				var relatedItems []*Item
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
						Url:         f.URL,
						Published:   i.Published,
						Hide:        i.Hide,
						AuthorName:  i.Name,
						AuthorEmail: i.Email,
					}
					err = tx.FirstOrCreate(
						&existingFeedItem,
						fi).Error
					if err != nil {
						return err
					}
				}
			}

			err = tx.DropTable("feeds").Error
			if err != nil {
				return err
			}

			err = tx.DropTable("items").Error
			if err != nil {
				return err
			}

			return nil
		},
	})

	m := gormigrate.New(gdb.db, gormigrate.DefaultOptions, migrations)
	return errors.Wrap(m.Migrate(), "failed to migrate gorm DB")
}
