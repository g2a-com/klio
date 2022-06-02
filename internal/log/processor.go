package log

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Mode string

const (
	lineMode Mode = "line"
	rawMode  Mode = "raw"

	controlRune = '\033'

	EscapeMarkerMode     = "klio_mode"
	EscapeMarkerLogLevel = "klio_log_level"
	EscapeMarkerTags     = "klio_tags"
	EscapeMarkerReset    = "klio_reset"
)

type Processor struct {
	defaultLevel Level
	defaultMode  Mode
	input        io.Reader
	logger       *Logger
}

func NewProcessor(DefaultLevel Level, Logger *Logger, Input io.Reader) *Processor {
	return &Processor{
		defaultLevel: DefaultLevel,
		logger:       Logger,
		input:        Input,
		defaultMode:  lineMode,
	}
}

func (lp *Processor) Process() {
	scanner := bufio.NewScanner(lp.input)
	scanner.Split(scanLinesAndKlioEscCodes)

	level := lp.defaultLevel
	var tags []string
	line := ""
	mode := lp.defaultMode

	flush := func() {
		if line != "" {
			_, _ = lp.logger.printer.Print(&message{
				Level: level,
				Tags:  tags,
				Text:  line,
			})
			line = ""
		}
	}

	for scanner.Scan() {
		chunk := scanner.Text()

		if mode == rawMode && !isEscCode(chunk) {
			_, _ = lp.logger.output.Write([]byte(chunk))
			continue
		}

		switch {
		case isEscCode(chunk):
			cmd, args, err := parseEscCode(chunk)
			if err != nil {
				Spamf("Failed to parse esc sequence while processing logs: %s", err)
				continue
			}

			// When in raw mode, ignore all commands except mode change

			if mode == rawMode && cmd != EscapeMarkerMode {
				continue
			}

			switch cmd {
			case EscapeMarkerLogLevel:
				newLevel, ok := levelsByName[args[0]]
				if ok {
					if newLevel != level {
						flush()
						level = newLevel
					}
				} else {
					Spamf("Failed to parse esc sequence while processing logs: invalid log level: %s", args[0])
				}
			case EscapeMarkerTags:
				newTags := args
				if !equalStringSlices(newTags, tags) {
					flush()
					tags = newTags
				}
			case EscapeMarkerReset:
				if len(tags) == 0 || level != lp.defaultLevel {
					flush()
					level = lp.defaultLevel
					tags = []string{}
				}
			case EscapeMarkerMode:
				newMode := Mode(args[0])
				if newMode == lineMode || newMode == rawMode {
					if newMode != mode {
						flush()
					}
					mode = newMode
				} else {
					Spamf("Failed to parse esc sequence while processing logs: invalid log mode: %s", args[0])
				}
			}
		case chunk == "\n":
			flush()
		default:
			line += chunk
		}
	}

	flush()

	if err := scanner.Err(); err != nil {
		Errorf("Error while processing logs: %s", err)
	}
}

func scanLinesAndKlioEscCodes(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil
	}

	if data[0] == '\n' {
		return 1, data[0:1], nil
	}

	if data[0] == controlRune {
		if len(data) == 1 {
			if atEOF {
				return 1, data[0:1], nil
			} else {
				return 0, nil, nil
			}
		}

		if data[1] != '_' {
			return 1, data[0:1], nil
		}

		for idx := 1; idx < len(data); idx++ {
			if data[idx-1] == controlRune && data[idx] == '\\' {
				return idx + 1, data[0 : idx+1], nil
			}
		}

		return 0, nil, nil
	}

	// Return all data until next ESC or \n
	for idx, chr := range data {
		if chr == controlRune || chr == '\n' {
			return idx, data[0:idx], nil
		}
	}

	return len(data), data, nil
}

func parseEscCode(code string) (cmd string, args []string, err error) {
	parts := strings.SplitN(code[2:len(code)-2], " ", 2)
	cmd = parts[0]

	switch cmd {
	case EscapeMarkerLogLevel:
		if len(parts) < 2 {
			return cmd, args, fmt.Errorf("%s requires an argument", cmd)
		}

		var arg string
		if err := json.Unmarshal([]byte(parts[1]), &arg); err != nil {
			return cmd, args, err
		}

		return cmd, []string{arg}, nil
	case EscapeMarkerTags:
		if len(parts) < 2 {
			return cmd, args, fmt.Errorf("%s requires an argument", cmd)
		}

		if err := json.Unmarshal([]byte(parts[1]), &args); err != nil {
			return cmd, args, fmt.Errorf("%s", err)
		}

		return cmd, args, nil
	case EscapeMarkerReset:
		if len(parts) > 1 {
			return cmd, args, fmt.Errorf("%s doesn't accept arguments", cmd)
		}

		return cmd, args, nil
	case EscapeMarkerMode:
		if len(parts) < 2 {
			return cmd, args, fmt.Errorf("%s requires an argument", cmd)
		}

		var arg string
		if err := json.Unmarshal([]byte(parts[1]), &arg); err != nil {
			return cmd, args, err
		}

		return cmd, []string{arg}, nil
	}

	return cmd, args, fmt.Errorf("%s is not supported", cmd)
}

func isEscCode(code string) bool {
	return strings.HasPrefix(code, fmt.Sprintf("%c_klio", controlRune)) &&
		strings.HasSuffix(code, fmt.Sprintf("%c\\", controlRune))
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
