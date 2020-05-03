package log

import (
	"errors"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

type Level int

const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

func NewLevel(l string) (Level, error) {
	switch l {
	case LevelTrace.String():
		return LevelTrace, nil
	case LevelDebug.String():
		return LevelDebug, nil
	case LevelInfo.String():
		return LevelInfo, nil
	case LevelWarn.String():
		return LevelWarn, nil
	case LevelError.String():
		return LevelError, nil
	case LevelFatal.String():
		return LevelFatal, nil
	default:
		return LevelTrace, errors.New("invalid log level")
	}
}

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "trace"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	default:
		panic("invalid level")
	}
}

var currLevel = LevelInfo

var rootLogger = &logrusLogger{
	backend: logrus.New(),
}

type Logger interface {
	Trace(string, ...interface{})
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Fatal(string, ...interface{})
	Sub(...interface{}) Logger
}

func SetLevel(level Level) {
	currLevel = level

	var logrusLevel logrus.Level
	switch level {
	case LevelTrace:
		logrusLevel = logrus.TraceLevel
	case LevelDebug:
		logrusLevel = logrus.DebugLevel
	case LevelInfo:
		logrusLevel = logrus.InfoLevel
	case LevelWarn:
		logrusLevel = logrus.WarnLevel
	case LevelError:
		logrusLevel = logrus.ErrorLevel
	case LevelFatal:
		logrusLevel = logrus.PanicLevel
	}
	rootLogger.backend.(*logrus.Logger).SetLevel(logrusLevel)
}

func WithModule(name string) Logger {
	return rootLogger.Sub("module", name)
}

func init() {
	// set log level to trace by default in test
	if strings.HasSuffix(os.Args[0], ".test") {
		SetLevel(LevelTrace)
	}
}
