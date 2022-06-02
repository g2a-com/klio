package log

import (
	"github.com/kataras/pio"
)

// Level is a number which defines the log level.
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

// DefaultLevel represents default level for non-error logs.
const DefaultLevel = InfoLevel

// DefaultErrorLevel represents default level for error logs.
const DefaultErrorLevel = ErrorLevel

// MaxLevel represents maximum level for logs.
const MaxLevel = SpamLevel

// LevelNames contains all supported names of logging levels.
var LevelNames = []string{"disable", "fatal", "error", "warn", "info", "verbose", "debug", "spam"}

// levelsByName maps level names to golog.level type.
var levelsByName = map[string]Level{
	"disable": DisableLevel,
	"fatal":   FatalLevel,
	"error":   ErrorLevel,
	"warn":    WarnLevel,
	"info":    InfoLevel,
	"verbose": VerboseLevel,
	"debug":   DebugLevel,
	"spam":    SpamLevel,
}

type LevelConfig struct {
	Name        string
	DisplayText string
	Color       int
}

var levels = map[Level]LevelConfig{
	DisableLevel: {
		Name:        "disable",
		DisplayText: "",
	},
	FatalLevel: {
		Name:        "fatal",
		DisplayText: "FATA",
		Color:       pio.Red,
	},
	ErrorLevel: {
		Name:        "error",
		DisplayText: "ERRO",
		Color:       pio.Red,
	},
	WarnLevel: {
		Name:        "error",
		DisplayText: "WARN",
		Color:       pio.Yellow,
	},
	InfoLevel: {
		Name:        "info",
		DisplayText: "INFO",
		Color:       pio.Cyan,
	},
	VerboseLevel: {
		Name:        "verbose",
		DisplayText: "VERB",
		Color:       pio.Gray,
	},
	DebugLevel: {
		Name:        "debug",
		DisplayText: "DEBU",
		Color:       pio.Gray,
	},
	SpamLevel: {
		Name:        "spam",
		DisplayText: "SPAM",
		Color:       pio.Gray,
	},
}
