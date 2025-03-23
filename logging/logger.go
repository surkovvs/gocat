package logging

import (
	"strings"

	"github.com/AlekSi/pointer"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type Configurer interface {
	GetLogLvl() *Level
	IsJSONEncoder() bool
}

type Config struct {
	Level          any
	ConsoleEncoder bool // if not - json encoder will be used
}

type Level int

const (
	LevelDebug Level = iota + 1
	LevelInfo
	LevelWarn
	LevelError

	LevelDebugStr string = "debug"
	LevelInfoStr  string = "info"
	LevelWarnStr  string = "warn"
	LevelErrorStr string = "error"
)

func (c Config) GetLogLvl() *Level {
	switch toParse := c.Level.(type) {
	case int:
		res := Level(toParse)
		return &res
	case string:
		switch {
		case strings.EqualFold(LevelDebugStr, toParse):
			return pointer.To(Level(LevelDebug))
		case strings.EqualFold(LevelInfoStr, toParse):
			return pointer.To(Level(LevelInfo))
		case strings.EqualFold(LevelWarnStr, toParse):
			return pointer.To(Level(LevelWarn))
		case strings.EqualFold(LevelErrorStr, toParse):
			return pointer.To(Level(LevelError))
		}
	}
	return nil
}

func (c Config) IsJSONEncoder() bool {
	return !c.ConsoleEncoder
}
