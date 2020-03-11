package config

import "github.com/spf13/viper"

type FeedConfig struct {
	URL  string
	Tags string
}

type Config interface {
	SetConfigName(string)
	AddConfigPath(string)
	ReadInConfig() error
	Get(string) interface{}
	GetString(string) string
}

type viperConfig struct{}

func FromViperConfig() Config {
	return &viperConfig{}
}

func (vc *viperConfig) SetConfigName(configName string) {
	viper.SetConfigName(configName)
}

func (vc *viperConfig) AddConfigPath(configPath string) {
	viper.AddConfigPath(configPath)
}

func (vc *viperConfig) ReadInConfig() error {
	return viper.ReadInConfig()
}

func (vc *viperConfig) Get(property string) interface{} {
	return viper.Get(property)
}

func (vc *viperConfig) GetString(property string) string {
	return viper.GetString(property)
}
