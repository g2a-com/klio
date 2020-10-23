package log

import (
	"fmt"
	"os"
)

// Following functions are based on golog default logging methods:
// https://github.com/kataras/golog/blob/master/golog.go

var DefaultLogger = NewLogger(os.Stdout)
var ErrorLogger = NewLogger(os.Stderr)

// SetLevel sets minimum level for logs, logs with level above specified value will not be printed
func SetLevel(levelName string) {
	level, ok := LevelsByName[levelName]
	if ok {
		DefaultLogger.Level = level
		ErrorLogger.Level = level
	} else {
		DefaultLogger.Level = DefaultLevel
		ErrorLogger.Level = level
	}
}

// // SetOutput
// func SetOutput(w io.Writer) {
// 	logger.SetOutput(w)
// }

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
	return levels[DefaultLogger.Level].Name
}

// IncreaseLevel changes current level by specified number
func IncreaseLevel(levels int) {
	if DefaultLogger.Level + Level(levels) > MaxLevel {
		DefaultLogger.Level = MaxLevel
	} else {
		DefaultLogger.Level = DefaultLogger.Level + Level(levels)
	}

	if ErrorLogger.Level + Level(levels) > MaxLevel {
		ErrorLogger.Level = MaxLevel
	} else {
		ErrorLogger.Level = ErrorLogger.Level + Level(levels)
	}
}

// Print prints a log message without levels and colors.
func Print(v ...interface{}) {
	DefaultLogger.Print(&Message{
		Text: fmt.Sprint(v...),
	})
}

// Println prints a log message without levels and colors.
// It adds a new line at the end.
func Println(v ...interface{}) {
	DefaultLogger.Println(&Message{
		Text: fmt.Sprint(v...),
	})
}

// Fatal `os.Exit(1)` exit no matter the level of the logger.
// If the logger's level is fatal, error, warn, info, verbose debug or spam
// then it will print the log message too.
func Fatal(v ...interface{}) {
	Log(FatalLevel, v...)
	// TODO: flush?
	os.Exit(1)
}

// Fatalf will `os.Exit(1)` no matter the level of the logger.
// If the logger's level is fatal, error, warn, info, verbose debug or spam
// then it will print the log message too.
func Fatalf(format string, args ...interface{}) {
	Logf(FatalLevel, format, args...)
	// TODO: flush?
	os.Exit(1)
}

// Error will print only when logger's Level is error, warn, info, verbose, debug or spam.
func Error(v ...interface{}) {
	Log(ErrorLevel, v...)
}

// Errorf will print only when logger's Level is error, warn, info, verbose, debug or spam.
func Errorf(format string, args ...interface{}) {
	Logf(ErrorLevel, format, args...)
}

// Warn will print only when logger's Level is warn, info, verbose, debug or spam.
func Warn(v ...interface{}) {
	Log(WarnLevel, v...)
}

// Warnf will print only when logger's Level is warn, info, verbose, debug or spam.
func Warnf(format string, args ...interface{}) {
	Logf(WarnLevel, format, args...)
}

// Info will print only when logger's Level is info, verbose, debug or spam.
func Info(v ...interface{}) {
	Log(InfoLevel, v...)
}

// Infof will print only when logger's Level is info, verbose, debug or spam.
func Infof(format string, args ...interface{}) {
	Logf(InfoLevel, format, args...)
}

// Verbose will print only when logger's Level is verbose, debug or spam.
func Verbose(v ...interface{}) {
	Log(VerboseLevel, v...)
}

// Verbosef will print only when logger's Level is verbose, debug or spam.
func Verbosef(format string, args ...interface{}) {
	Logf(VerboseLevel, format, args...)
}

// Debug will print only when logger's Level is debug or spam.
func Debug(v ...interface{}) {
	Log(DebugLevel, v...)
}

// Debugf will print only when logger's Level is debug or spam.
func Debugf(format string, args ...interface{}) {
	Logf(DebugLevel, format, args...)
}

// Spam will print when logger's Level is spam.
func Spam(v ...interface{}) {
	Log(SpamLevel, v...)
}

// Spamf will print when logger's Level is spam.
func Spamf(format string, args ...interface{}) {
	Logf(SpamLevel, format, args...)
}

// Log prints message with specified level
func Log(level Level, v ...interface{}) {
	DefaultLogger.Println(&Message{
		Level: level,
		Text: fmt.Sprint(v...),
	})
}

// Logf prints message with specified level
func Logf(level Level, format string, args ...interface{}) {
	DefaultLogger.Println(&Message{
		Level: level,
		Text: fmt.Sprintf(format, args...),
	})
}

// LogAndExit prints message with specified level and calls os.Exit(1)
func LogAndExit(level Level, v ...interface{}) {
	Log(level, v...)
	// TODO: flush?
	os.Exit(1)
}

// LogfAndExit prints message with specified level and calls os.Exit(1)
func LogfAndExit(level Level, format string, args ...interface{}) {
	Logf(level, format, args...)
	// TODO: flush?
	os.Exit(1)
}
