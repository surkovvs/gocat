package catdefzerolog

import (
	"os"

	"github.com/surkovvs/gocat/catlog"

	"github.com/rs/zerolog"
)

func NewZerologDefault(cfg catlog.Configurer) *zerolog.Logger {
	if lvl := cfg.GetLogLvl(); lvl != nil {
		switch *lvl {
		case catlog.LevelDebug:
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case catlog.LevelInfo:
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case catlog.LevelWarn:
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case catlog.LevelError:
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		}
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &logger
}
