package logdeafult

import (
	"os"

	"github.com/surkovvs/gocat/logging"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapDefault(cfg logging.Configurer) *zap.Logger {
	var encoder zapcore.Encoder
	if cfg.IsJSONEncoder() {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	} else {
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}

	zapLvl := zapcore.WarnLevel
	if lvl := cfg.GetLogLvl(); lvl != nil {
		switch *lvl {
		case logging.LevelDebug:
			zapLvl = zapcore.DebugLevel
		case logging.LevelInfo:
			zapLvl = zapcore.InfoLevel
		case logging.LevelWarn:
			zapLvl = zapcore.WarnLevel
		case logging.LevelError:
			zapLvl = zapcore.ErrorLevel
		}
	}

	return zap.New(zapcore.NewCore(encoder, os.Stdout, zapLvl))
}
