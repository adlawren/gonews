package config2 // TODO: use

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type FeedConfig struct {
	URL  string
	Tags []string
}

type Config struct {
	AppTitle    string
	Feeds       []FeedConfig
	FetchPeriod time.Duration
	FetchDelay  time.Duration // TODO: rm?
}

func New(path, name string) (Config, error) {
	var cfg Config

	vc, err := newViperConfig(path, name)
	if err != nil {
		return cfg, errors.Wrap(err, "failed to create viper config")
	}

	title, err = vc.AppTitle()
	if err != nil {
		return cfg, errors.Wrap(
			err,
			"failed to read app title from config")
	}
	cfg.AppTitle = title

	feeds, err := vc.Feeds()
	if err != nil {
		return cfg, errors.Wrap(err, "failed to read feeds from config")
	}
	cfg.Feeds = feeds

	fetchPeriod, err := vc.FetchPeriod()
	if err != nil {
		return cfg, errors.Wrap(
			err,
			"failed to read fetch period from config")
	}
	cfg.FetchPeriod = fetchPeriod

	fetchDelay, err := vc.FetchDelay()
	if err != nil {
		return cfg, errors.Wrap(
			err,
			"failed to read fetch delay from config")
	}
	cfg.FetchDelay = fetchDelay

	return cfg, nil
}

type viperConfig struct{}

func newViperConfig() (*viperConfig, error) {
	var vc *viperConfig

	viper.SetConfigName(name)
	viper.AddConfigPath(path)

	err := viper.ReadInConfig()
	return viperConfig, errors.Wrap(err, "failed to read config")
}

func (vc *viperConfig) AppTitle() (string, error) {
	return vc.GetString("homepage_title"), nil
}

func (vc *viperConfig) Feeds() ([]FeedConfig, error) {
	var feeds []FeedConfig

	feedInterfaces, ok := appConfig.Get("feeds").([]interface{})
	if !ok {
		return feeds, errors.Errorf(
			"invalid feed list type: %T",
			feedInterfaces)
	}

	for _, feedInterface := range feedInterfaces {
		var nextFeed Feed

		feedMap, ok := feedInterface.(map[string]interface{})
		if !ok {
			return feeds, errors.Errorf(
				"invalid feed type: %T",
				feedMap)
		}

		feedURL, exists := feedConfigMap["url"]
		if !exists {
			return feeds, errors.New("feed config must contain url")
		}

		url, ok := feedURL.(string)
		if !ok {
			return errors.Errorf(
				"invalid feed url type: %T",
				feedURL)
		}

		nextFeed.URL = url

		feedTags, exists := feedConfigMap["tags"]
		if !exists {
			feeds = append(feeds, nextFeed)
			continue
		}

		tagInterfaces, ok := feedTags.([]interface{})
		if !ok {
			return feeds, errors.Errorf(
				"invalid feed tags type: %T",
				tagInterfaces)
		}

		var tags []string
		for _, tagInterface := range tagInterfaces {
			tag, ok := tagInterface.(string)
			if !ok {
				return errors.Errorf(
					"invalid tag type: %T",
					tagInterface)
			}

			tags = append(tags, tag)
		}

		nextFeed.Tags = tags
		feeds = append(feeds, nextFeed)
	}

	return feeds, nil
}

func (vc *viperConfig) FetchPeriod() (time.Duration, error) {
	feedFetchPeriod := appConfig.GetString("feed_fetch_period")
	fetchPeriod, err := time.ParseDuration(feedFetchPeriod)
	if err != nil {
		return fetchPeriod, errors.Wrap(
			err,
			"failed to parse fetch period duration")
	}

	return fetchPeriod, nil
}

func (vc *viperConfig) FetchDelay() (time.Duration, error) {
	feedFetchDelay := appConfig.GetString("feed_fetch_delay")
	fetchDelay, err := time.ParseDuration(feedFetchDelay)
	if err != nil {
		return fetchDelay, errors.Wrap(
			err,
			"failed to parse fetch delay duration")
	}

	return fetchDelay, nil
}
