package logging

import (
	"github.com/rs/zerolog"
)

type zerologAdapter struct {
	*zerolog.Logger
}

func NewZerologAdapter(logger *zerolog.Logger) zerologAdapter {
	return zerologAdapter{logger}
}

func (zla zerologAdapter) Debug(msg string, args ...any) {
	zla.Logger.Debug().Fields(args).Msg(msg)
}

func (zla zerologAdapter) Info(msg string, args ...any) {
	zla.Logger.Info().Fields(args).Msg(msg)
}

func (zla zerologAdapter) Warn(msg string, args ...any) {
	zla.Logger.Warn().Fields(args).Msg(msg)
}

func (zla zerologAdapter) Error(msg string, args ...any) {
	zla.Logger.Error().Fields(args).Msg(msg)
}
