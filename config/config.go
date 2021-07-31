package config

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	dbConfigInst *DBConfig
)

// Config contains the values parsed from the config file
type Config struct {
	AppTitle    string `mapstructure:"homepage_title"`
	Feeds       []*FeedConfig
	FetchPeriod time.Duration `mapstructure:"feed_fetch_period"`
}

// FeedConfig contains the values associated with each feed, parsed from the
// config file
type FeedConfig struct {
	URL               string
	Tags              []string
	FetchLimit        uint          `mapstructure:"fetch_limit"`
	AutoDismissPeriod time.Duration `mapstructure:"auto_dismiss_period"`
}

// DBConfig contains the values needed to connect to the database
// Note: currently not parsed from the config file
type DBConfig struct {
	DSN string
}

// SetDBConfigInst assigns the global database configuration instance to the
// given config
func SetDBConfigInst(dbCfg *DBConfig) {
	dbConfigInst = dbCfg
}

// DBConfigInst returns the global database configuration instance
func DBConfigInst() *DBConfig {
	return dbConfigInst
}

// New creates an instance of Config by parsing the given config file
func New(path, name string) (*Config, error) {
	viper.SetConfigName(name)
	viper.AddConfigPath(path)
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}

	var c Config
	err = viper.Unmarshal(&c)
	if err != nil {
		return &c, errors.Wrap(err, "failed to unmarshal config")
	}

	return &c, nil
}

func (c Config) String() string {
	return fmt.Sprintf(
		"App Title: %s, Feeds: %s, Fetch Period: %s",
		c.AppTitle,
		c.Feeds,
		c.FetchPeriod)
}

func (fc FeedConfig) String() string {
	return fmt.Sprintf(
		"URL: %s, Tags: %s, Fetch Limit: %s, AutoDismissPeriod: %s",
		fc.URL,
		fc.Tags,
		fc.FetchLimit,
		fc.AutoDismissPeriod)
}
