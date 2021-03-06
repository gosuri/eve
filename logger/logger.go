package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/logrusorgru/aurora"
)

var defaultLogger *Logger

// Level is the level of logging.
type Level int

const (
	LevelDebug Level = iota // Debug level.
	LevelInfo               // Info level.
	LevelWarn               // Warn level.
	LevelError              // Error level.
)

// Logger is a simple logger that logs to a writer.
type Logger struct {
	out   io.Writer
	level Level
}

func init() {
	defaultLogger = FromEnv(os.Stderr)
}

func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}

func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

func Warn(v ...interface{}) {
	defaultLogger.Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

func Error(v ...interface{}) {
	defaultLogger.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

// FromEnv creates a logger from the environment variable LOG_LEVEL.
func FromEnv(out io.Writer) *Logger {
	return &Logger{
		out:   out,
		level: levelFromEnv(),
	}
}

func levelFromEnv() Level {
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		lvl = "warn"
	}

	switch strings.ToLower(lvl) {
	default:
		return LevelInfo
	case "debug":
		return LevelDebug
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	}
}

func (l *Logger) debug(v ...interface{}) {
	if str, ok := v[0].(string); ok {
		byteString := []byte(str)
		if json.Valid(byteString) {
			var prettyJSON bytes.Buffer
			err := json.Indent(&prettyJSON, byteString, "", "  ")
			if err == nil {
				quick.Highlight(l.out, prettyJSON.String()+"\n", "json", "terminal", "monokai")
			}
		}
	}
	fmt.Fprintln(l.out, aurora.Faint("DEBUG"), fmt.Sprint(v...))
}

func (l *Logger) Debug(v ...interface{}) {
	if l.level <= LevelDebug {
		l.debug(v...)
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level <= LevelDebug {
		l.debug(fmt.Sprintf(format, v...))
	}
}

func (l *Logger) info(v ...interface{}) {
	fmt.Fprintln(l.out, aurora.Faint("INFO"), fmt.Sprint(v...))
}

func (l *Logger) Info(v ...interface{}) {
	if l.level <= LevelInfo {
		l.info(v...)
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level <= LevelInfo {
		l.info(fmt.Sprintf(format, v...))
	}
}

func (l *Logger) warn(v ...interface{}) {
	fmt.Fprintln(l.out, aurora.Yellow("WARN"), fmt.Sprint(v...))
}

func (l *Logger) Warn(v ...interface{}) {
	if l.level <= LevelWarn {
		l.warn(v...)
	}
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level <= LevelWarn {
		l.warn(fmt.Sprintf(format, v...))
	}
}

func (l *Logger) error(v ...interface{}) {
	fmt.Fprintln(l.out, aurora.Red("ERROR"), fmt.Sprint(v...))
}

func (l *Logger) Error(v ...interface{}) {
	if l.level <= LevelError {
		l.error(v...)
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level <= LevelError {
		l.error(fmt.Sprintf(format, v...))
	}
}
