package log

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kataras/pio"
)

type message struct {
	Level Level
	Tags  []string
	Text  string
}

type messageMarshaler struct {
	SupportColors bool
}

func newMessageMarshaler(supportColors bool) messageMarshaler {
	return messageMarshaler{
		SupportColors: supportColors,
	}
}

func (m messageMarshaler) Marshal(data interface{}) ([]byte, error) {
	msg, ok := data.(*message)

	if !ok {
		return []byte{}, errors.New("not a *message")
	}

	level := levels[msg.Level]
	text := ""

	if msg.Level != DisableLevel {
		if m.SupportColors {
			text += pio.Rich(fmt.Sprintf("[%s]", level.DisplayText), level.Color)
		} else {
			text += fmt.Sprintf("[%s]", level.DisplayText)
		}
	}

	if len(msg.Tags) > 0 {
		text += fmt.Sprintf("[%s]", strings.ToUpper(strings.Join(msg.Tags, "][")))
	}

	if len(text) > 0 && len(msg.Text) > 0 {
		text += " "
	}

	if len(msg.Text) > 0 {
		text += msg.Text
	}

	return []byte(text), nil
}
