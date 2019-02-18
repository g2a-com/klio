package runner

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/g2a-com/klio/pkg/log"
)

type writer struct {
	LogPrefix      string
	Level          log.Level
	DecorateOutput bool
	Text           *string
}

func (writer *writer) Write(data []byte) (int, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		text := scanner.Text()
		*writer.Text += text

		if !writer.DecorateOutput {
			log.Println(text)
		} else if writer.LogPrefix != "" {
			log.Logf(writer.Level, "[%s] %s", writer.LogPrefix, text)
		} else {
			log.Logf(writer.Level, "%s", text)
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return len(data), nil
}

type Command struct {
	Command        string
	Args           []string
	DecorateOutput bool
	StdoutLevel    log.Level
	StderrLevel    log.Level
	StdoutText     string
	StderrText     string
	LogPrefix      string
}

func NewCommand(cmd string, args ...string) *Command {
	return &Command{
		Command:        cmd,
		Args:           args,
		DecorateOutput: true,
		StdoutLevel:    log.DefaultLevel,
		StderrLevel:    log.DefaultErrorLevel,
		LogPrefix:      strings.ToUpper(path.Base(cmd)),
	}
}

func (cmd *Command) Run() error {
	log.Debugf(`running %s "%s"`, cmd.Command, strings.Join(cmd.Args, `" "`))
	externalCmd := exec.Command(cmd.Command, cmd.Args...)
	externalCmd.Stdin = os.Stdin
	externalCmd.Stdout = &writer{
		Level:          cmd.StdoutLevel,
		LogPrefix:      cmd.LogPrefix,
		DecorateOutput: cmd.DecorateOutput,
		Text:           &cmd.StdoutText,
	}
	externalCmd.Stderr = &writer{
		Level:          cmd.StderrLevel,
		LogPrefix:      cmd.LogPrefix,
		DecorateOutput: cmd.DecorateOutput,
		Text:           &cmd.StderrText,
	}
	return externalCmd.Run()
}
