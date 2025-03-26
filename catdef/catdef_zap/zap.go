package catdefzap

import (
	"os"

	"github.com/surkovvs/gocat/catlog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapDefault(cfg catlog.Configurer) *zap.Logger {
	var encoder zapcore.Encoder
	if cfg.IsJSONEncoder() {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	} else {
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}

	zapLvl := zapcore.WarnLevel
	if lvl := cfg.GetLogLvl(); lvl != nil {
		switch *lvl {
		case catlog.LevelDebug:
			zapLvl = zapcore.DebugLevel
		case catlog.LevelInfo:
			zapLvl = zapcore.InfoLevel
		case catlog.LevelWarn:
			zapLvl = zapcore.WarnLevel
		case catlog.LevelError:
			zapLvl = zapcore.ErrorLevel
		}
	}

	return zap.New(zapcore.NewCore(encoder, os.Stdout, zapLvl))
}
