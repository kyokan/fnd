package log

import "github.com/sirupsen/logrus"

type logrusLogger struct {
	backend logrus.FieldLogger
}

var _ Logger = (*logrusLogger)(nil)

func (l *logrusLogger) Trace(msg string, fields ...interface{}) {
	if l.isEnabled(LevelTrace) {
		l.parseFields(fields).Debug(msg)
	}
}

func (l *logrusLogger) Debug(msg string, fields ...interface{}) {
	if l.isEnabled(LevelDebug) {
		l.parseFields(fields).Debug(msg)
	}
}

func (l *logrusLogger) Info(msg string, fields ...interface{}) {
	if l.isEnabled(LevelInfo) {
		l.parseFields(fields).Info(msg)
	}
}

func (l *logrusLogger) Warn(msg string, fields ...interface{}) {
	if l.isEnabled(LevelWarn) {
		l.parseFields(fields).Warn(msg)
	}
}

func (l *logrusLogger) Error(msg string, fields ...interface{}) {
	if l.isEnabled(LevelError) {
		l.parseFields(fields).Error(msg)
	}
}

func (l *logrusLogger) Fatal(msg string, fields ...interface{}) {
	if l.isEnabled(LevelFatal) {
		l.parseFields(fields).Fatal(msg)
	}
}

func (l *logrusLogger) Sub(fields ...interface{}) Logger {
	return &logrusLogger{
		backend: l.parseFields(fields),
	}
}

func (l *logrusLogger) isEnabled(level Level) bool {
	return level >= currLevel
}

func (l *logrusLogger) parseFields(fields []interface{}) logrus.FieldLogger {
	argLen := len(fields)
	if argLen == 0 {
		return l.backend
	}
	if argLen%2 != 0 {
		panic("must specify arguments as tuples")
	}

	lFields := make(logrus.Fields)
	for i := 0; i < argLen; i += 2 {
		k := fields[i]
		v := fields[i+1]

		kStr, ok := k.(string)
		if !ok {
			panic("argument keys must be strings")
		}

		lFields[kStr] = v
	}
	return l.backend.WithFields(lFields)
}
