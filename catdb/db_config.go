package catdb

import (
	"fmt"
	"time"
)

type ConfigPool struct {
	LogConnectOperations  bool
	MaxConnLifetime       *time.Duration
	MaxConnLifetimeJitter *time.Duration
	MaxConnIdleTime       *time.Duration
	MaxConns              *int32
	MinConns              *int32
	HealthCheckPeriod     *time.Duration
}

type ConfigDB struct {
	// prefer URL for coonection
	DSN        string
	Host       string
	Port       uint16
	Name       string
	User       string
	Pass       string
	LogQueries bool

	ConfigPool `mapstructure:"Pool"`

	// TLS *tls.Config
}

// TODO: add TLS
func (cfg ConfigDB) GetDSN() string {
	if cfg.DSN != "" {
		return cfg.DSN
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.User,
		cfg.Pass,
	)
}
