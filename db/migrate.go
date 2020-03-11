package db

import (
	"gonews/legacy"

	"github.com/jinzhu/gorm"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"gopkg.in/gormigrate.v1"
)

type TestA struct {
	gorm.Model
	Str    string
	TestBs []TestB
}

type TestB struct {
	gorm.Model
	Str     string
	TestAID uint
}

func MigrateGormDB(db DB) error {
	gdb, ok := db.(*gormDB)
	if !ok {
		return errors.Errorf("db type: (%T), expected gorm.DB", db)
	}

	gdb.db.AutoMigrate(&TestA{})
	gdb.db.AutoMigrate(&TestB{})

	a := &TestA{
		Str: "TestA",
	}
	var existingA TestA
	gdb.FirstOrCreate(&existingA, a)

	b := &TestB{
		Str: "Test 1",
		//TestAID: existingA.ID,
	}
	var existingB TestB
	gdb.FirstOrCreate(&existingB, b)

	b2 := &TestB{
		Str: "Test 2",
		//TestAID: existingA.ID,
	}
	var existingB2 TestB
	gdb.FirstOrCreate(&existingB2, b2)

	b3 := &TestB{
		Str: "Test 3",
		//TestAID: existingA.ID,
	}
	var existingB3 TestB
	gdb.FirstOrCreate(&existingB3, b3)

	var existingTestA2 TestA
	gdb.FirstOrCreate(&existingTestA2, &TestA{
		Str: "TestA",
	})

	return nil // TODO

	var migrations []*gormigrate.Migration
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
