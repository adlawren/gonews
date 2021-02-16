package config

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Config contains the values parsed from the config file
type Config struct {
	AppTitle    string
	Feeds       []*FeedConfig
	FetchPeriod time.Duration
}

// FeedConfig contains the values associated with each feed, parsed from the
// config file
type FeedConfig struct {
	URL        string
	Tags       []string
	FetchLimit uint
}

// DBConfig contains the values needed to connect to the database
// Note: currently not parsed from the config file
type DBConfig struct {
	DSN string
}

// New creates an instance of Config by parsing the given config file
func New(path, name string) (*Config, error) {
	var cfg Config

	vc, err := newViperConfig(path, name)
	if err != nil {
		return &cfg, errors.Wrap(err, "failed to create viper config")
	}

	title, err := vc.AppTitle()
	if err != nil {
		return &cfg, errors.Wrap(
			err,
			"failed to read app title from config")
	}
	cfg.AppTitle = title

	feeds, err := vc.Feeds()
	if err != nil {
		return &cfg, errors.Wrap(err, "failed to read feeds from config")
	}
	cfg.Feeds = feeds

	fetchPeriod, err := vc.FetchPeriod()
	if err != nil {
		return &cfg, errors.Wrap(
			err,
			"failed to read fetch period from config")
	}
	cfg.FetchPeriod = fetchPeriod

	return &cfg, nil
}

type viperConfig struct{}

func newViperConfig(path, name string) (*viperConfig, error) {
	var vc viperConfig

	viper.SetConfigName(name)
	viper.AddConfigPath(path)

	err := viper.ReadInConfig()
	return &vc, errors.Wrap(err, "failed to read config")
}

func (vc *viperConfig) AppTitle() (string, error) {
	return viper.GetString("homepage_title"), nil
}

func (vc *viperConfig) Feeds() ([]*FeedConfig, error) {
	var feeds []*FeedConfig

	feedInterfaces, ok := viper.Get("feeds").([]interface{})
	if !ok {
		return feeds, errors.Errorf(
			"invalid feed list type: %T",
			viper.Get("feeds"))
	}

	for _, feedInterface := range feedInterfaces {
		var nextFeed FeedConfig

		feedMap, ok := feedInterface.(map[string]interface{})
		if !ok {
			return feeds, errors.Errorf(
				"invalid feed type: %T",
				feedInterface)
		}

		feedURL, exists := feedMap["url"]
		if !exists {
			return feeds, errors.New("feed config must contain url")
		}

		url, ok := feedURL.(string)
		if !ok {
			return feeds, errors.Errorf(
				"invalid feed url type: %T",
				feedURL)
		}

		nextFeed.URL = url

		feedTags, exists := feedMap["tags"]
		if !exists {
			feeds = append(feeds, &nextFeed)
			continue
		}

		tagInterfaces, ok := feedTags.([]interface{})
		if !ok {
			return feeds, errors.Errorf(
				"invalid feed tags type: %T",
				feedTags)
		}

		var tags []string
		for _, tagInterface := range tagInterfaces {
			tag, ok := tagInterface.(string)
			if !ok {
				return feeds, errors.Errorf(
					"invalid tag type: %T",
					tagInterface)
			}

			tags = append(tags, tag)
		}

		nextFeed.Tags = tags

		feedFetchLimit, exists := feedMap["fetch_limit"]
		if !exists {
			feeds = append(feeds, &nextFeed)
			continue
		}

		fetchLimit, ok := feedFetchLimit.(int64)
		if !ok {
			return feeds, errors.Errorf(
				"invalid feed fetch limit type: %T",
				feedFetchLimit)
		}

		nextFeed.FetchLimit = uint(fetchLimit)

		feeds = append(feeds, &nextFeed)
	}

	return feeds, nil
}

func (vc *viperConfig) FetchPeriod() (time.Duration, error) {
	feedFetchPeriod := viper.GetString("feed_fetch_period")
	fetchPeriod, err := time.ParseDuration(feedFetchPeriod)
	return fetchPeriod, errors.Wrap(
		err,
		"failed to parse fetch period duration")
}
