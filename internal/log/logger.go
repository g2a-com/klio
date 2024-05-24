package log

import (
	"fmt"
	"io"
	"os"

	"github.com/kataras/pio"
)

type Logger struct {
	level   Level
	printer *pio.Printer
	output  io.Writer
}

func NewLogger(output io.Writer) *Logger {
	return &Logger{
		level:   DefaultLevel,
		printer: pio.NewPrinter("line", output).Marshal(newMessageMarshaler(pio.SupportColors(output))),
		output:  output,
	}
}

// WithDefaultLevel allows to set default level for logger
func (l *Logger) WithDefaultLevel(level Level) *Logger {
	l.level = level
	return l
}

// WithOutput allows to set custom output for logger
func (l *Logger) WithOutput(output io.Writer) *Logger {
	l.output = output
	return l
}

// Fatal `os.Exit(1)` exit no matter the level of the logger.
// If the logger's level is fatal, error, warn, info, verbose debug or spam
// then it will print the log message too.
func (l *Logger) Fatal(v ...interface{}) {
	l.log(FatalLevel, v...)
	os.Exit(1)
}

// Fatalf will `os.Exit(1)` no matter the level of the logger.
// If the logger's level is fatal, error, warn, info, verbose debug or spam
// then it will print the log message too.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logf(FatalLevel, format, args...)
	os.Exit(1)
}

// Error will print only when logger's Level is error, warn, info, verbose, debug or spam.
func (l *Logger) Error(v ...interface{}) {
	l.log(ErrorLevel, v...)
}

// Errorf will print only when logger's Level is error, warn, info, verbose, debug or spam.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf(ErrorLevel, format, args...)
}

// Warn will print only when logger's Level is warn, info, verbose, debug or spam.
func (l *Logger) Warn(v ...interface{}) {
	l.log(WarnLevel, v...)
}

// Warnf will print only when logger's Level is warn, info, verbose, debug or spam.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logf(WarnLevel, format, args...)
}

// Info will print only when logger's Level is info, verbose, debug or spam.
func (l *Logger) Info(v ...interface{}) {
	l.log(InfoLevel, v...)
}

// Infof will print only when logger's Level is info, verbose, debug or spam.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf(InfoLevel, format, args...)
}

// Verbose will print only when logger's Level is verbose, debug or spam.
func (l *Logger) Verbose(v ...interface{}) {
	l.log(VerboseLevel, v...)
}

// Verbosef will print only when logger's Level is verbose, debug or spam.
func (l *Logger) Verbosef(format string, args ...interface{}) {
	l.logf(VerboseLevel, format, args...)
}

// Debug will print only when logger's Level is debug or spam.
func (l *Logger) Debug(v ...interface{}) {
	l.log(DebugLevel, v...)
}

// Debugf will print only when logger's Level is debug or spam.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logf(DebugLevel, format, args...)
}

// Spam will print when logger's Level is spam.
func (l *Logger) Spam(v ...interface{}) {
	l.log(SpamLevel, v...)
}

// Spamf will print when logger's Level is spam.
func (l *Logger) Spamf(format string, args ...interface{}) {
	l.logf(SpamLevel, format, args...)
}

// Println prints a log message without levels and colors.
func (l *Logger) Println(v ...interface{}) {
	l.log(DisableLevel, v...)
}

// Printf formats according to the specified format without levels and colors.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.logf(DisableLevel, format, args...)
}

// log prints message with specified level.
func (l *Logger) log(level Level, v ...interface{}) {
	if level <= l.level {
		_, _ = l.printer.Println(&message{Level: level, Text: fmt.Sprint(v...)})
	}
}

// logf prints message with specified level.
func (l *Logger) logf(level Level, format string, args ...interface{}) {
	if level <= l.level {
		_, _ = l.printer.Println(&message{Level: level, Text: fmt.Sprintf(format, args...)})
	}
}
