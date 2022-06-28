package log

import (
	"github.com/g2a-com/klio/internal/log"
)

func setLevel(level string) {
	log.SetLevel(level)
}

func setLevelFromEnv(level string) {
	log.SetLevelFromEnv()
}
func IncreaseLevel(levels int) {
	log.IncreaseLevel(levels)
}
func Warn(v ...interface{}) {
	log.Warn(v)
}
func Warnf(format string, v ...interface{}) {
	log.Warn(format, v)
}
func Debug(v ...interface{}) {
	log.Debug(v)
}
func Debugf(format string, v ...interface{}) {
	log.Debugf(format, v)
}

func Spam(v ...interface{}) {
	log.Spam(v)
}

func Spamf(format string, v ...interface{}) {
	log.Spamf(format, v)
}
func Verbose(v ...interface{}) {
	log.Spam(v)
}
func Verbosef(format string, v ...interface{}) {
	log.Spamf(format, v)
}
func Error(v ...interface{}) {
	log.Error(v)
}
func Errorf(format string, v ...interface{}) {
	log.Errorf(format, v)
}
func Log(level log.Level, v ...interface{}) {
	log.Log(level, v)
}
func Logf(level log.Level, format string, v ...interface{}) {
	log.Logf(level, format, v)
}
func Info(v ...interface{}) {
	log.Info(v)
}
func Infof(format string, v ...interface{}) {
	log.Infof(format, v)
}
