package catlog

import (
	"go.uber.org/zap"
)

type zapAdapter struct {
	*zap.SugaredLogger
}

func NewZapAdapter(logger *zap.Logger) zapAdapter {
	return zapAdapter{logger.Sugar()}
}

func (za zapAdapter) Debug(msg string, args ...any) {
	za.SugaredLogger.Debugw(msg, args...)
}

func (za zapAdapter) Info(msg string, args ...any) {
	za.SugaredLogger.Infow(msg, args...)
}

func (za zapAdapter) Warn(msg string, args ...any) {
	za.SugaredLogger.Warnw(msg, args...)
}

func (za zapAdapter) Error(msg string, args ...any) {
	za.SugaredLogger.Errorw(msg, args...)
}
