package log

import (
	"os"

	"github.com/kataras/golog"
)

// Following functions are based on golog default logging methods:
// https://github.com/kataras/golog/blob/master/golog.go

var logger = newLogger()

// SetLevel sets minimum level for logs, logs with level above specified value will not be printed
func SetLevel(levelName string) {
	for level, meta := range levels {
		if meta.Name == levelName {
			logger.Level = level
			return
		}
	}
	logger.Level = DefaultLevel
	return
}

// GetDefaultLevel returns default logging level name
func GetDefaultLevel() string {
	return levels[DefaultLevel].Name
}

// SetLevelFromEnv sets minimum level for logs based on environment variables
func SetLevelFromEnv() {
	SetLevel(os.Getenv("G2A_CLI_LOG_LEVEL"))
}

// GetLevel returns current logging level name
func GetLevel() string {
	return levels[logger.Level].Name
}

// IncreaseLevel changes current level by specified number
func IncreaseLevel(levels int) {
	logger.Level += golog.Level(levels)
}

// Print prints a log message without levels and colors.
func Print(v ...interface{}) {
	logger.Print(v...)
}

// Println prints a log message without levels and colors.
// It adds a new line at the end.
func Println(v ...interface{}) {
	logger.Println(v...)
}

// Fatal `os.Exit(1)` exit no matter the level of the logger.
// If the logger's level is fatal, error, warn, info, verbose debug or spam
// then it will print the log message too.
func Fatal(v ...interface{}) {
	logger.Log(FatalLevel, v...)
}

// Fatalf will `os.Exit(1)` no matter the level of the logger.
// If the logger's level is fatal, error, warn, info, verbose debug or spam
// then it will print the log message too.
func Fatalf(format string, args ...interface{}) {
	logger.Logf(FatalLevel, format, args...)
}

// Error will print only when logger's Level is error, warn, info, verbose, debug or spam.
func Error(v ...interface{}) {
	logger.Log(ErrorLevel, v...)
}

// Errorf will print only when logger's Level is error, warn, info, verbose, debug or spam.
func Errorf(format string, args ...interface{}) {
	logger.Logf(ErrorLevel, format, args...)
}

// Warn will print only when logger's Level is warn, info, verbose, debug or spam.
func Warn(v ...interface{}) {
	logger.Log(WarnLevel, v...)
}

// Warnf will print only when logger's Level is warn, info, verbose, debug or spam.
func Warnf(format string, args ...interface{}) {
	logger.Logf(WarnLevel, format, args...)
}

// Info will print only when logger's Level is info, verbose, debug or spam.
func Info(v ...interface{}) {
	logger.Log(InfoLevel, v...)
}

// Infof will print only when logger's Level is info, verbose, debug or spam.
func Infof(format string, args ...interface{}) {
	logger.Logf(InfoLevel, format, args...)
}

// Verbose will print only when logger's Level is verbose, debug or spam.
func Verbose(v ...interface{}) {
	logger.Log(VerboseLevel, v...)
}

// Verbosef will print only when logger's Level is verbose, debug or spam.
func Verbosef(format string, args ...interface{}) {
	logger.Logf(VerboseLevel, format, args...)
}

// Debug will print only when logger's Level is debug or spam.
func Debug(v ...interface{}) {
	logger.Log(DebugLevel, v...)
}

// Debugf will print only when logger's Level is debug or spam.
func Debugf(format string, args ...interface{}) {
	logger.Logf(DebugLevel, format, args...)
}

// Spam will print when logger's Level is spam.
func Spam(v ...interface{}) {
	logger.Log(SpamLevel, v...)
}

// Spamf will print when logger's Level is spam.
func Spamf(format string, args ...interface{}) {
	logger.Logf(SpamLevel, format, args...)
}

// Log prints message with specified level
func Log(level Level, v ...interface{}) {
	logger.Log(golog.Level(level), v...)
}

// Logf prints message with specified level
func Logf(level Level, format string, args ...interface{}) {
	logger.Logf(golog.Level(level), format, args...)
}
