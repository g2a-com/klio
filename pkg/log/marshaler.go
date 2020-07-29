package log

import (
	"fmt"
	"strings"
	"errors"
	"github.com/kataras/pio"
)

type messageMarshaler struct {
	SupportColors bool
}

func newMarshaler (supportColors bool) messageMarshaler {
	return messageMarshaler{
		SupportColors: supportColors,
	}
}

func(m messageMarshaler) Marshal (data interface{}) ([]byte, error) {
	msg, ok := data.(*Message)

	if (!ok) {
		return []byte{}, errors.New("not a *Message")
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

