package configuration

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/surkovvs/gocat/database/conninit"
	"github.com/surkovvs/gocat/logging"
)

type Config struct {
	Logger   logging.Config
	Database conninit.Config
}

func ParseFile(path string) (*Config, error) {
	var cfg Config
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf(`read config: %w`, err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf(`unmarshal config: %w`, err)
	}
	return &cfg, nil
}
