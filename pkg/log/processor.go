package log

import (
	"fmt"
	"bufio"
	"strings"
	"io"
	"encoding/json"
)

type LogProcessor struct {
	DefaultLevel Level
	Input io.Reader
	Logger *Logger
}

func NewLogProcessor() *LogProcessor {
	return &LogProcessor{
		DefaultLevel: InfoLevel,
	}
}

func (lp *LogProcessor) Process() {
	scanner := bufio.NewScanner(lp.Input)
	scanner.Split(scanLinesAndKlioEscCodes)

	level := lp.DefaultLevel
	tags := []string{}
	line := ""

	flush := func () {
		if line != "" {
			lp.Logger.Println(&Message{
				Level: level,
				Tags: tags,
				Text: line,
			})
			line = ""
		}
	}

	for scanner.Scan() {
		chunk := scanner.Text()

		switch {
			case isEscCode(chunk):
				cmd, args, err := parseEscCode(chunk)
				if err != nil {
					Spamf("Failed to parse esc sequence while processing logs: %s", err)
					continue
				}

				switch cmd {
				case "klio_log_level":
					newLevel, ok := LevelsByName[args[0]]
					if ok {
						if newLevel != level {
							flush()
							level = newLevel
						}
					} else {
						Spamf("Failed to parse esc sequence while processing logs: invalid log level: %s", args[0])
					}
				case "klio_tags":
					newTags := args
					if !equalStringSlices(newTags, tags) {
						flush()
						tags = newTags
					}
				case "klio_reset":
					if len(tags) == 0 || level != lp.DefaultLevel {
						flush()
						level = lp.DefaultLevel
						tags = []string{}
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
	//Spamf("scan-chunk (atEOF %t): '%s'", atEOF, data)

	if len(data) == 0 {
		return 0, nil, nil
	}

	if data[0] == '\n' {
		return 1, data[0:1], nil
	}

	if data[0] == '\033' {
		if len(data) == 1 {
			if atEOF {
				return 1, data[0: 1], nil
			} else {
				return 0, nil, nil
			}
		}

		if data[1] != '_' {
			return 1, data[0:1], nil
		}

		for idx := 1; idx < len(data); idx++ {
			if data[idx-1] == '\033' && data[idx] == '\\'{
				return idx + 1, data[0:idx+1], nil
			}
		}

		return 0, nil, nil
	}

	// Return all data until next ESC or \n
	for idx, chr := range data {
		if chr == '\033' || chr == '\n' {
		  return idx, data[0:idx], nil
		}
	}

	return len(data), data, nil
}

func parseEscCode (code string) (cmd string, args []string, err error) {
	parts := strings.SplitN(code[2:len(code)-2], " ", 2)
	cmd = parts[0]

	switch cmd {
	case "klio_log_level":
		if len(parts) < 2 {
			return cmd, args, fmt.Errorf("%s requires an argument", cmd)
		}

		var arg string
		if err := json.Unmarshal([]byte(parts[1]), &arg); err != nil {
			return cmd, args, err
		}

		return cmd, []string{ arg }, nil
	case "klio_tags":
		if len(parts) < 2 {
			return cmd, args, fmt.Errorf("%s requires an argument", cmd)
		}

		if err := json.Unmarshal([]byte(parts[1]), &args); err != nil {
			return cmd, args, fmt.Errorf("%s", err)
		}

		return cmd, args, nil
	case "klio_reset":
		if len(parts) > 1 {
			return cmd, args, fmt.Errorf("%s doesn't accept arguments", cmd)
		}

		return cmd, args, nil
	}

	return cmd, args, fmt.Errorf("%s is not supported", cmd)
}

func isEscCode (code string) bool {
	return strings.HasPrefix(code, "\033_klio") && strings.HasSuffix(code, "\033\\")
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
