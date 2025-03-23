package conninit

import (
	"crypto/tls"
)

type Config struct {
	// prefer URL for coonection
	URL    string
	Host   string
	Port   uint16
	DBName string
	User   string
	Pass   string

	TLS    *tls.Config
	Logger Logger
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}
