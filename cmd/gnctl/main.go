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
	"gonews/timestamp"
	"gonews/user"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func printModels(i interface{}) error {
	iKind := reflect.TypeOf(i).Kind()
	if iKind != reflect.Array && iKind != reflect.Slice {
		return fmt.Errorf("array or slice required")
	}

	iVal := reflect.ValueOf(i)
	for i := 0; i < iVal.Len(); i++ {
		printModel(iVal.Index(i).Interface())
	}

	return nil
}

func printModel(model interface{}) error {
	modelBytes, err := json.Marshal(model)
	if err != nil {
		return errors.Wrap(err, "failed to marshal json")
	}
	fmt.Println(string(modelBytes[:]))

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

func scanTimestamps() ([]*timestamp.Timestamp, error) {
	lines := scanLines()

	var timestamps []*timestamp.Timestamp
	for _, line := range lines {
		var t timestamp.Timestamp

		err := json.Unmarshal([]byte(line), &t)
		if err != nil {
			return timestamps, errors.Wrap(err, "failed to unmarshal timestamp")
		}

		timestamps = append(timestamps, &t)
	}

	return timestamps, nil
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

func saveTimestamps(db db.DB, timestamps []*timestamp.Timestamp) error {
	for _, timestamp := range timestamps {
		err := db.SaveTimestamp(timestamp)
		if err != nil {
			return errors.Wrap(err, "failed to save timestamp")
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
	matchingTimestamp := flag.String("matching-timestamp", "", "show matching timestamp, given serialized timestamp fields")
	matchingUser := flag.String("matching-user", "", "show matching user, given serialized user fields")
	migrateDB := flag.Bool("migrate-db", false, "apply DB migrations")
	migrationsDir := flag.String("migrations-dir", "db/migrations", "database migrations directory")
	pingDB := flag.Bool("ping-db", false, "ping DB")
	showFeeds := flag.Bool("feeds", false, "show feeds")
	showItems := flag.Bool("items", false, "show items")
	showTags := flag.Bool("tags", false, "show tags")
	showUsers := flag.Bool("users", false, "show users")
	tagName := flag.String("items-from-tag", "", "show items from tag name")
	testAuth := flag.String("test-auth", "", "validate the given authentication credentials; ex. 'some_user:some_password'")
	upsertFeeds := flag.Bool("upsert-feeds", false, "upsert the given serialized feeds read from stdin, one per line")
	upsertItems := flag.Bool("upsert-items", false, "upsert the given serialized items read from stdin, one per line")
	upsertTags := flag.Bool("upsert-tags", false, "upsert the given serialized tags read from stdin, one per line")
	upsertTimestamps := flag.Bool("upsert-timestamps", false, "upsert the given serialized timestamps read from stdin, one per line")
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

	if *showUsers {
		users, err := adb.Users()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get users")
			return
		}

		err = printModels(users)
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

		err = printModels(feeds)
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

		err = printModels(tags)
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

		err = printModels(items)
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

		err = printModels(items)
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

		err = printModels(items)
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

		err = printModel(item)
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

	if len(*matchingTimestamp) > 0 {
		var timestamp timestamp.Timestamp
		err := json.Unmarshal([]byte(*matchingTimestamp), &timestamp)
		if err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal timestamp")
			return
		}

		match, err := adb.MatchingTimestamp(&timestamp)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get matching timestamp")
			return
		}

		err = printModel(match)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print timestamp")
			return
		}
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

		err = printModel(match)
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

		err = printModel(match)
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

		err = printModel(match)
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

		err = printModel(match)
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

		err = printModels(items)
		if err != nil {
			log.Error().Err(err).Msg("Failed to print items")
			return
		}
	}

	if *upsertTimestamps {
		timestamps, err := scanTimestamps()
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan timestamps")
			return
		}

		err = saveTimestamps(adb, timestamps)
		if err != nil {
			log.Error().Err(err).Msg("Failed to upsert timestamps")
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
