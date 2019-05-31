package log

import (
	"github.com/kataras/golog"
)

func newLogger() *golog.Logger {
	logger := golog.New()
	logger.SetTimeFormat("")
	return logger
}
