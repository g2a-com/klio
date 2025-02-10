package log

import (
	"io"

	"github.com/g2a-com/klio/internal/log"
)

type Level = log.Level

const (
	// DisableLevel disables all leveled logs (some output, such as usage, will be
	// still printed).
	DisableLevel = log.DisableLevel
	// FatalLevel logs indicates serious errors, which make program unable to
	// running.
	FatalLevel = log.FatalLevel
	// ErrorLevel logs indicates errors which prevent program to perform some
	// function.
	ErrorLevel = log.ErrorLevel
	// WarnLevel logs indicates some less serious errors, which doesn't affect
	// program execution.
	WarnLevel = log.WarnLevel
	// InfoLevel logs confirmates that things are working as expected.
	InfoLevel = log.InfoLevel
	// VerboseLevel logs contain detailed information that should be
	// understandable to experienced users to provide insight in the software’s
	// behavior.
	VerboseLevel = log.VerboseLevel
	// DebugLevel logs contain detailed information, typically of interest only
	// when diagnosing problems.
	DebugLevel = log.DebugLevel
	// SpamLevel logs are too verbose for regular debbuging.
	SpamLevel = log.SpamLevel
	// DefaultLevel represents default level for non-error logs.
	DefaultLevel = log.DefaultLevel
	// DefaultErrorLevel represents default level for error logs.
	DefaultErrorLevel = log.DefaultErrorLevel
	// MaxLevel represents maximum level for logs.
	MaxLevel = log.MaxLevel
)

// NewLogger returns a new logger instance with defined output and default level.
func NewLogger(out io.Writer, defaultLevel Level) *log.Logger {
	return log.NewLogger(out).WithDefaultLevel(defaultLevel)
}

// SetLevel sets minimum level for logs, logs with level above specified value will not be printed.
func SetLevel(level string) {
	log.SetLevel(level)
}

// SetLevelFromEnv sets minimum level for logs based on environment variables.
func SetLevelFromEnv() {
	log.SetLevelFromEnv()
}

// IncreaseLevel changes current level by specified number.
func IncreaseLevel(levels int) {
	log.IncreaseLevel(levels)
}

// GetDefaultLevel returns default logging level name.
func GetDefaultLevel() string {
	return log.GetDefaultLevel()
}

// GetDefaultErrorLevel returns default error logging level name.
func GetDefaultErrorLevel() string {
	return log.GetDefaultErrorLevel()
}

// GetLevel returns current logging level name.
func GetLevel() string {
	return log.GetLevel()
}

// GetSupportedLevels returns supported logging levels.
func GetSupportedLevels() []string {
	return log.GetSupportedLevels()
}

// Println prints a log message without levels and colors.
func Println(v ...interface{}) {
	log.Println(v...)
}

// Printf prints a log message without levels and colors.
func Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// Fatal `os.Exit(1)` exit no matter the level of the logger.
// If the logger's level is fatal, error, warn, info, verbose debug or spam
// then it will print the log message too.
func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

// Fatalf will `os.Exit(1)` no matter the level of the logger.
// If the logger's level is fatal, error, warn, info, verbose debug or spam
// then it will print the log message too.
func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

// Error will print only when logger's Level is error, warn, info, verbose, debug or spam.
func Error(v ...interface{}) {
	log.Error(v...)
}

// Errorf will print only when logger's Level is error, warn, info, verbose, debug or spam.
func Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

// Warn will print only when logger's Level is warn, info, verbose, debug or spam.
func Warn(v ...interface{}) {
	log.Warn(v...)
}

// Warnf will print only when logger's Level is warn, info, verbose, debug or spam.
func Warnf(format string, v ...interface{}) {
	log.Warnf(format, v...)
}

// Info will print only when logger's Level is info, verbose, debug or spam.
func Info(v ...interface{}) {
	log.Info(v...)
}

// Infof will print only when logger's Level is info, verbose, debug or spam.
func Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

// Debug will print only when logger's Level is debug or spam.
func Debug(v ...interface{}) {
	log.Debug(v...)
}

// Debugf will print only when logger's Level is debug or spam.
func Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

// Spam will print when logger's Level is spam.
func Spam(v ...interface{}) {
	log.Spam(v...)
}

// Spamf will print when logger's Level is spam.
func Spamf(format string, v ...interface{}) {
	log.Spamf(format, v...)
}

// Verbose will print only when logger's Level is verbose, debug or spam.
func Verbose(v ...interface{}) {
	log.Spam(v...)
}

// Verbosef will print only when logger's Level is verbose, debug or spam.
func Verbosef(format string, v ...interface{}) {
	log.Spamf(format, v...)
}

// Log prints message with specified level.
func Log(level Level, v ...interface{}) {
	log.Log(level, v...)
}

// Logf prints message with specified level.
func Logf(level Level, format string, args ...interface{}) {
	log.Logf(level, format, args...)
}
