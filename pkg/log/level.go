package log

import (
	"github.com/kataras/golog"
)

// Level is a number which defines the log level
type Level uint32

const (
	// DisableLevel disables all leveled logs (some output, such as usage, will be
	// still printed).
	DisableLevel = iota
	// FatalLevel logs indicates serious errors, which make program unable to
	// running.
	FatalLevel
	// ErrorLevel logs indicates errors which prevent program to perform some
	// function.
	ErrorLevel
	// WarnLevel logs indicates some less serious errors, which doesn't affect
	// program execution.
	WarnLevel
	// InfoLevel logs confirmates that things are working as expected.
	InfoLevel
	// VerboseLevel logs contain detailed information that should be
	// understandable to experienced users to provide insight in the software’s
	// behavior.
	VerboseLevel
	// DebugLevel logs contain detailed information, typically of interest only
	// when diagnosing problems.
	DebugLevel
	// SpamLevel logs are too verbose for regular debbuging.
	SpamLevel
)

// DefaultLevel represents default level for non-error logs
const DefaultLevel = InfoLevel

// DefaultErrorLevel represents default level for error logs
const DefaultErrorLevel = ErrorLevel

// LevelNames contains all supported names of logging levels
var LevelNames = []string{"disable", "fatal", "error", "warn", "info", "verbose", "debug", "spam"}

var levels = map[golog.Level]*golog.LevelMetadata{
	golog.Level(DisableLevel): {
		Name:         "disable",
		RawText:      "",
		ColorfulText: "",
	},
	golog.Level(FatalLevel): {
		Name:         "fatal",
		RawText:      "[FATA]",
		ColorfulText: "\x1b[41m[FATA]\x1b[0m",
	},
	golog.Level(ErrorLevel): {
		Name:         "error",
		RawText:      "[ERRO]",
		ColorfulText: "\x1b[31m[ERRO]\x1b[0m",
	},
	golog.Level(WarnLevel): {
		Name:         "warn",
		RawText:      "[WARN]",
		ColorfulText: "\x1b[33m[WARN]\x1b[0m",
	},
	golog.Level(InfoLevel): {
		Name:         "info",
		RawText:      "[INFO]",
		ColorfulText: "\x1b[36m[INFO]\x1b[0m",
	},
	golog.Level(VerboseLevel): {
		Name:         "verbose",
		RawText:      "[VERB]",
		ColorfulText: "\x1b[90m[VERB]\x1b[0m",
	},
	golog.Level(DebugLevel): {
		Name:         "debug",
		RawText:      "[DEBU]",
		ColorfulText: "\x1b[90m[DEBU]\x1b[0m",
	},
	golog.Level(SpamLevel): {
		Name:         "spam",
		RawText:      "[SPAM]",
		ColorfulText: "\x1b[90m[SPAM]\x1b[0m",
	},
}

func init() {
	golog.Levels = levels
}
