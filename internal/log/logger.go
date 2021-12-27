package log

import (
	"io"

	"github.com/kataras/pio"
)

type Logger struct {
	Level   Level
	Printer *pio.Printer
	Output  io.Writer
}

func NewLogger(output io.Writer) *Logger {
	return &Logger{
		Level:   DefaultLevel,
		Printer: pio.NewPrinter("line", output).Marshal(newMarshaler(pio.SupportColors(output))),
		Output:  output,
	}
}

func (l *Logger) Print(message *Message) {
	if message.Level <= l.Level {
		l.Printer.Print(message)
	}
}

func (l *Logger) Println(message *Message) {
	if message.Level <= l.Level {
		l.Printer.Println(message)
	}
}
