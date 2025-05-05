package catcfg

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/spf13/viper"
	"github.com/surkovvs/gocat/catdb"
	"github.com/surkovvs/gocat/catlog"
)

type Config struct {
	Logger           catlog.Logger `mapstructure:"-"`
	catlog.ConfigLog `mapstructure:"Logger"`
	catdb.ConfigDB   `mapstructure:"Database"`
	Kafka            map[string]kafka.ConfigMap `mapstructure:"Kafka"` // `mapstructure:"-"`
}

func ParseFile(path string) (*Config, error) {
	var cfg Config
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf(`read config: %w`, err)
	}

	viper.SetOptions(viper.KeyDelimiter("|"))

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf(`unmarshal config: %w`, err)
	}
	return &cfg, nil
}

func (cfg *Config) SetLogger(logger catlog.Logger) {
	cfg.Logger = logger
}
