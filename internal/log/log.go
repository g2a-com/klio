package log

import (
	"io"
	"os"

	"github.com/g2a-com/klio/internal/env"
)

// Following functions are based on golog default logging methods:
// https://github.com/kataras/golog/blob/master/golog.go

var (
	DefaultLogger = NewLogger(os.Stdout)
	ErrorLogger   = NewLogger(os.Stderr)
)

// SetOutput allows for setting output to which logger writes messages
func SetOutput(o io.Writer) {
	DefaultLogger.SetOutput(o)
}

// SetLevel sets minimum level for logs, logs with level above specified value will not be printed.
func SetLevel(levelName string) {
	level, ok := levelsByName[levelName]
	if ok {
		DefaultLogger.level = level
		ErrorLogger.level = level
	} else {
		DefaultLogger.level = DefaultLevel
		ErrorLogger.level = DefaultErrorLevel
	}
}

// SetLevelFromEnv sets minimum level for logs based on environment variables
func SetLevelFromEnv() {
	SetLevel(os.Getenv(env.KLIO_LOG_LEVEL))
}

// GetDefaultLevel returns default logging level name.
func GetDefaultLevel() string {
	return levels[DefaultLevel].Name
}

// GetDefaultErrorLevel returns default error logging level name.
func GetDefaultErrorLevel() string {
	return levels[DefaultErrorLevel].Name
}

// GetLevel returns current logging level name.
func GetLevel() string {
	return levels[DefaultLogger.level].Name
}

// GetSupportedLevels returns supported logging levels.
func GetSupportedLevels() []string {
	return LevelNames
}

// IncreaseLevel changes current level by specified number.
func IncreaseLevel(levels int) {
	if DefaultLogger.level+Level(levels) > MaxLevel {
		DefaultLogger.level = MaxLevel
	} else {
		DefaultLogger.level = DefaultLogger.level + Level(levels)
	}

	if ErrorLogger.level+Level(levels) > MaxLevel {
		ErrorLogger.level = MaxLevel
	} else {
		ErrorLogger.level = ErrorLogger.level + Level(levels)
	}
}

// Println prints a log message without levels and colors.
func Println(v ...interface{}) {
	DefaultLogger.Println(v...)
}

// Printf prints a log message without levels and colors.
func Printf(format string, args ...interface{}) {
	DefaultLogger.Printf(format, args...)
}

// Fatal `os.Exit(1)` exit no matter the level of the logger.
// If the logger's level is fatal, error, warn, info, verbose debug or spam
// then it will print the log message too.
func Fatal(v ...interface{}) {
	DefaultLogger.Fatal(v...)
	os.Exit(1)
}

// Fatalf will `os.Exit(1)` no matter the level of the logger.
// If the logger's level is fatal, error, warn, info, verbose debug or spam
// then it will print the log message too.
func Fatalf(format string, args ...interface{}) {
	DefaultLogger.Fatalf(format, args...)
	os.Exit(1)
}

// Error will print only when logger's Level is error, warn, info, verbose, debug or spam.
func Error(v ...interface{}) {
	DefaultLogger.Error(v...)
}

// Errorf will print only when logger's Level is error, warn, info, verbose, debug or spam.
func Errorf(format string, args ...interface{}) {
	DefaultLogger.Errorf(format, args...)
}

// Warn will print only when logger's Level is warn, info, verbose, debug or spam.
func Warn(v ...interface{}) {
	DefaultLogger.Warn(v...)
}

// Warnf will print only when logger's Level is warn, info, verbose, debug or spam.
func Warnf(format string, args ...interface{}) {
	DefaultLogger.Warnf(format, args...)
}

// Info will print only when logger's Level is info, verbose, debug or spam.
func Info(v ...interface{}) {
	DefaultLogger.Info(v...)
}

// Infof will print only when logger's Level is info, verbose, debug or spam.
func Infof(format string, args ...interface{}) {
	DefaultLogger.Infof(format, args...)
}

// Verbose will print only when logger's Level is verbose, debug or spam.
func Verbose(v ...interface{}) {
	DefaultLogger.Verbose(v...)
}

// Verbosef will print only when logger's Level is verbose, debug or spam.
func Verbosef(format string, args ...interface{}) {
	DefaultLogger.Verbosef(format, args...)
}

// Debug will print only when logger's Level is debug or spam.
func Debug(v ...interface{}) {
	DefaultLogger.Debug(v...)
}

// Debugf will print only when logger's Level is debug or spam.
func Debugf(format string, args ...interface{}) {
	DefaultLogger.Debugf(format, args...)
}

// Spam will print when logger's Level is spam.
func Spam(v ...interface{}) {
	DefaultLogger.Spam(v...)
}

// Spamf will print when logger's Level is spam.
func Spamf(format string, args ...interface{}) {
	DefaultLogger.Spamf(format, args...)
}

// Log prints message with specified level.
func Log(level Level, v ...interface{}) {
	DefaultLogger.log(level, v...)
}

// Logf prints message with specified level.
func Logf(level Level, format string, args ...interface{}) {
	DefaultLogger.logf(level, format, args...)
}
