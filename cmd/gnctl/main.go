package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"gonews/auth"
	"gonews/config"
	"gonews/db"
	"gonews/feed"
	"gonews/parser"
	"gonews/user"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func printUsers(users ...*user.User) error {
	for _, user := range users {
		userBytes, err := json.Marshal(user)
		if err != nil {
			return errors.Wrap(err, "failed to marshal json")
		}
		fmt.Println(string(userBytes[:]))
	}

	return nil
}

func printFeeds(feeds ...*feed.Feed) error {
	for _, feed := range feeds {
		feedBytes, err := json.Marshal(feed)
		if err != nil {
			return errors.Wrap(err, "failed to marshal json")
		}
		fmt.Println(string(feedBytes[:]))
	}

	return nil
}

func printTags(tags ...*feed.Tag) error {
	for _, tag := range tags {
		tagBytes, err := json.Marshal(tag)
		if err != nil {
			return errors.Wrap(err, "failed to marshal json")
		}
		fmt.Println(string(tagBytes[:]))
	}

	return nil
}

func printItems(items ...*feed.Item) error {
	for _, item := range items {
		itemBytes, err := json.Marshal(item)
		if err != nil {
			return errors.Wrap(err, "failed to marshal json")
		}
		fmt.Println(string(itemBytes[:]))
	}

	return nil
}

func scanLines() []string {
	var lines []string

	s := bufio.NewScanner(os.Stdin)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	return lines
}

func scanUsers() ([]*user.User, error) {
	lines := scanLines()

	var users []*user.User
	for _, line := range lines {
		var u user.User

		err := json.Unmarshal([]byte(line), &u)
		if err != nil {
			return users, errors.Wrap(err, "failed to unmarshal user")
		}

		users = append(users, &u)
	}

	return users, nil
}

func scanFeeds() ([]*feed.Feed, error) {
	lines := scanLines()

	var feeds []*feed.Feed
	for _, line := range lines {
		var f feed.Feed

		err := json.Unmarshal([]byte(line), &f)
		if err != nil {
			return feeds, errors.Wrap(err, "failed to unmarshal feed")
		}

		feeds = append(feeds, &f)
	}

	return feeds, nil
}

func scanTags() ([]*feed.Tag, error) {
	lines := scanLines()

	var tags []*feed.Tag
	for _, line := range lines {
		var t feed.Tag

		err := json.Unmarshal([]byte(line), &t)
		if err != nil {
			return tags, errors.Wrap(err, "failed to unmarshal tag")
		}

		tags = append(tags, &t)
	}

	return tags, nil
}

func scanItems() ([]*feed.Item, error) {
	lines := scanLines()

	var items []*feed.Item
	for _, line := range lines {
		var i feed.Item

		err := json.Unmarshal([]byte(line), &i)
		if err != nil {
			return items, errors.Wrap(err, "failed to unmarshal item")
		}

		items = append(items, &i)
	}

	return items, nil
}

func saveUsers(db db.DB, users []*user.User) error {
	for _, user := range users {
		err := db.SaveUser(user)
		if err != nil {
			return errors.Wrap(err, "failed to save user")
		}
	}

	return nil
}

func saveFeeds(db db.DB, feeds []*feed.Feed) error {
	for _, feed := range feeds {
		err := db.SaveFeed(feed)
		if err != nil {
			return errors.Wrap(err, "failed to save feed")
		}
	}

	return nil
}

func saveTags(db db.DB, tags []*feed.Tag) error {
	for _, tag := range tags {
		err := db.SaveTag(tag)
		if err != nil {
			return errors.Wrap(err, "failed to save tag")
		}
	}

	return nil
}

func saveItems(db db.DB, items []*feed.Item) error {
	for _, item := range items {
		err := db.SaveItem(item)
		if err != nil {
			return errors.Wrap(err, "failed to save item")
		}
	}

	return nil
}

func main() {
	configPath := flag.String("parse-config", "", "parse the application configuration file")
	dbDSN := flag.String("db-dsn", "file:/data/gonews/db.sqlite3", "database DSN")
	feedID := flag.Uint("items-from-feed", 0, "show items from feed ID")
	feedURL := flag.String("parse-url", "", "parse items from URL")
	hashPassword := flag.String("hash-password", "", "print the hash of the given password")
	itemID := flag.Uint("item", 0, "show item with given ID")
	matchingFeed := flag.String("matching-feed", "", "show matching feed, given serialized feed fields")
	matchingItem := flag.String("matching-item", "", "show matching item, given serialized item fields")
	matchingTag := flag.String("matching-tag", "", "show matching tag, given serialized tag fields")
	matchingUser := flag.String("matching-user", "", "show matching user, given serialized user fields")
	migrateDB := flag.Bool("migrate-db", false, "apply DB migrations")
	migrationsDir := flag.String("migrations-dir", "db/migrations", "database migrations directory")
	pingDB := flag.Bool("ping-db", false, "ping DB")
	showFeeds := flag.Bool("feeds", false, "show feeds")
	showItems := flag.Bool("items", false, "show items")
	showTags := flag.Bool("tags", false, "show tags")
	showTimestamp := flag.Bool("timestamp", false, "show timestamp")
	showUsers := flag.Bool("users", false, "show users")
	tagName := flag.String("items-from-tag", "", "show items from tag name")
	testAuth := flag.String("test-auth", "", "validate the given authentication credentials; ex. 'some_user:some_password'")
	upsertFeeds := flag.Bool("upsert-feeds", false, "upsert the given serialized feeds read from stdin, one per line")
	upsertItems := flag.Bool("upsert-items", false, "upsert the given serialized items read from stdin, one per line")
	upsertTags := flag.Bool("upsert-tags", false, "upsert the given serialized tags read from stdin, one per line")
	upsertTimestamp := flag.String("upsert-timestamp", "", "upsert the timestamp using the given time")
	upsertUsers := flag.Bool("upsert-users", false, "upsert the given serialized users read from stdin, one per line")

	flag.Parse()

	adb, err := db.New(&config.DBConfig{DSN: *dbDSN})
	defer adb.Close()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create DB client")
		return
	}

	if len(*configPath) > 0 {
		dir := path.Dir(*configPath)
		base := path.Base(*configPath)
		name := strings.Replace(base, path.Ext(base), "", 1)
		parsedConfig, err := config.New(dir, name)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse application configuration file")
			return
		}
		fmt.Println(parsedConfig)
	}

	if *pingDB {
		err = adb.Ping()
		if err != nil {
			log.Error().Err(err).Msg("Ping failed")
			return
		}
		fmt.Println("Ping succeeded")
	}

	if *migrateDB {
		err = adb.Migrate(*migrationsDir)
		if err != nil {
			log.Error().Err(err).Msg("Failed to migrate DB")
			return
		}
		fmt.Println("Migrations succeeded")
	}

	if *showTimestamp {
		t, err := adb.Timestamp()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get timestamp")
			return
		}

		fmt.Println(t.Format(time.RFC3339))
	}

	if *showUsers {
		users, err := adb.Users()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get users")
			return
		}

		err = printUsers(users...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print users")
			return
		}
	}

	if *showFeeds {
		feeds, err := adb.Feeds()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get feeds")
			return
		}

		err = printFeeds(feeds...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print feeds")
			return
		}
	}

	if *showTags {
		tags, err := adb.Tags()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get tags")
			return
		}

		err = printTags(tags...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print tags")
			return
		}
	}

	if *showItems {
		items, err := adb.Items()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get items")
			return
		}

		err = printItems(items...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print items")
			return
		}
	}

	if len(*tagName) > 0 {
		items, err := adb.ItemsFromTag(&feed.Tag{Name: *tagName})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get items")
			return
		}

		err = printItems(items...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print items")
			return
		}
	}

	if *feedID != 0 {
		items, err := adb.ItemsFromFeed(&feed.Feed{ID: *feedID})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get items")
			return
		}

		err = printItems(items...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print items")
			return
		}
	}

	if *itemID != 0 {
		item, err := adb.Item(*itemID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get item")
			return
		}

		err = printItems(item)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print item")
			return
		}
	}

	if len(*hashPassword) > 0 {
		hash, err := auth.Hash(*hashPassword)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash password")
			return
		}

		fmt.Println(hash)
	}

	if len(*testAuth) > 0 {
		res := strings.Split(*testAuth, ":")
		username := res[0]
		password := res[1]
		isValid, err := auth.IsValid(username, password, adb)

		// TODO: fix so that it can distinguish between an actual error and an 'invalid creds' error
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate creds")
			return
		}

		fmt.Println(isValid)
	}

	if len(*matchingUser) > 0 {
		var user user.User
		err := json.Unmarshal([]byte(*matchingUser), &user)
		if err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal user")
			return
		}

		match, err := adb.MatchingUser(&user)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get matching user")
			return
		}

		err = printUsers(match)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print user")
			return
		}
	}

	if len(*matchingFeed) > 0 {
		var feed feed.Feed
		err := json.Unmarshal([]byte(*matchingFeed), &feed)
		if err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal feed")
			return
		}

		match, err := adb.MatchingFeed(&feed)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get matching feed")
			return
		}

		err = printFeeds(match)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print feed")
			return
		}
	}

	if len(*matchingTag) > 0 {
		var tag feed.Tag
		err := json.Unmarshal([]byte(*matchingTag), &tag)
		if err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal tag")
			return
		}

		match, err := adb.MatchingTag(&tag)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get matching tag")
			return
		}

		err = printTags(match)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print tag")
			return
		}
	}

	if len(*matchingItem) > 0 {
		var item feed.Item
		err := json.Unmarshal([]byte(*matchingItem), &item)
		if err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal item")
			return
		}

		match, err := adb.MatchingItem(&item)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get matching item")
			return
		}

		err = printItems(match)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print item")
			return
		}
	}

	if len(*feedURL) > 0 {
		p, err := parser.New()
		if err != nil {
			log.Error().Err(err).Msg("Failed to create parser")
			return
		}

		items, err := p.ParseURL(*feedURL)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse feed")
			return
		}

		err = printItems(items...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print items")
			return
		}
	}

	if len(*upsertTimestamp) > 0 {
		t, err := time.Parse(time.RFC3339, *upsertTimestamp)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse given time")
			return
		}

		err = adb.SaveTimestamp(&t)
		if err != nil {
			log.Error().Err(err).Msg("Failed to save timestamp")
			return
		}
	}

	if *upsertUsers {
		users, err := scanUsers()
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan users")
			return
		}

		err = saveUsers(adb, users)
		if err != nil {
			log.Error().Err(err).Msg("Failed to upsert users")
			return
		}
	}

	if *upsertFeeds {
		feeds, err := scanFeeds()
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan feeds")
			return
		}

		err = saveFeeds(adb, feeds)
		if err != nil {
			log.Error().Err(err).Msg("Failed to upsert feeds")
			return
		}
	}

	if *upsertTags {
		tags, err := scanTags()
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan tags")
			return
		}

		err = saveTags(adb, tags)
		if err != nil {
			log.Error().Err(err).Msg("Failed to upsert tags")
			return
		}
	}

	if *upsertItems {
		items, err := scanItems()
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan items")
			return
		}

		err = saveItems(adb, items)
		if err != nil {
			log.Error().Err(err).Msg("Failed to upsert item")
			return
		}
	}
}
