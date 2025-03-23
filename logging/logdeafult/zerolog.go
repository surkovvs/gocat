package logdeafult

import (
	"os"

	"github.com/surkovvs/gocat/logging"

	"github.com/rs/zerolog"
)

func NewZerologDefault(cfg logging.Configurer) *zerolog.Logger {
	if lvl := cfg.GetLogLvl(); lvl != nil {
		switch *lvl {
		case logging.LevelDebug:
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case logging.LevelInfo:
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case logging.LevelWarn:
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case logging.LevelError:
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		}
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &logger
}
