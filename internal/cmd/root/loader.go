package root

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/manager"
	"github.com/g2a-com/klio/internal/log"
	"github.com/g2a-com/klio/internal/schema"
	"github.com/spf13/cobra"
)

const fiveSeconds = 5 * time.Second

func loadExternalCommand(ctx context.CLIContext, rootCmd *cobra.Command, dep dependency.DependenciesIndexEntry, global bool) {
	if cmd, _, _ := rootCmd.Find([]string{dep.Alias}); cmd != rootCmd {
		log.Spamf("cannot register already registered command '%s'", dep.Alias)
		return
	}

	cmdConfig, err := schema.LoadCommandConfig(filepath.Join(dep.Path, "command.yaml"))
	if err != nil {
		log.Warnf("Cannot load command: %s", err)
		return
	}

	cmd := &cobra.Command{
		Use:                dep.Alias,
		Short:              cmdConfig.Description,
		Long:               "",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			externalCmdPath := filepath.Join(dep.Path, cmdConfig.BinPath)
			var externalCmd *exec.Cmd
			if runtime.GOOS == "windows" {
				args = append([]string{"/c", externalCmdPath}, args...)
				externalCmdPath = "cmd"
			}
			externalCmd = exec.Command(externalCmdPath, args...)
			externalCmd.Stdin = os.Stdin

			var wg sync.WaitGroup

			switch cmdConfig.APIVersion {
			case "g2a-cli/v1beta1":
				externalCmd.Stdout = os.Stdout
				externalCmd.Stderr = os.Stderr
			case "klio/v1":
				setupLogProcessor(externalCmd, &wg)
			default:
				log.Warnf(
					"Cannot load command %s since it requires an unsupported API Version to run (%s). Try to update the %s and try again.",
					dep.Alias,
					cmdConfig.APIVersion,
					ctx.Config.CommandName,
				)
				return
			}

			updateMsgChannel := make(chan string, 1)
			go getUpdateMessage(ctx, dep, global, updateMsgChannel)

			timeoutChannel := make(chan bool, 1)
			go func() {
				time.Sleep(fiveSeconds)
				timeoutChannel <- true
			}()

			log.Debugf(`Running %s "%s"`, externalCmdPath, strings.Join(args, `" "`))

			if err := externalCmd.Start(); err != nil {
				log.Fatal(err)
			}

			wg.Wait()

			err = externalCmd.Wait()

			select {
			case msg := <-updateMsgChannel:
				if msg != "" {
					for _, line := range strings.Split(msg, "\n") {
						log.ErrorLogger.Println(&log.Message{Level: log.WarnLevel, Text: line})
					}
				}
			case <-timeoutChannel:
				break
			}

			if err != nil {
				switch e := err.(type) {
				case *exec.ExitError:
					os.Exit(e.ExitCode())
				default:
					log.Fatal(err)
				}
			}
		},
		Version: fmt.Sprintf("%s (registry: %s, arch: %s, os: %s, checksum: %s)", dep.Version, dep.Registry, dep.Arch, dep.OS, dep.Checksum),
	}
	rootCmd.AddCommand(cmd)
}

func getUpdateMessage(ctx context.CLIContext, dep dependency.DependenciesIndexEntry, global bool, msg chan<- string) {
	depMgr := manager.NewManager(ctx)

	getInstallCmd := func(ver string) string {
		cmd := fmt.Sprintf("%s get", ctx.Config.CommandName)

		if global {
			cmd += " -g"
		}
		cmd += fmt.Sprintf(" %s --version %s --from %s", dep.Name, ver, dep.Registry)
		if dep.Name != dep.Alias {
			cmd += fmt.Sprintf(" --as %s", dep.Alias)
		}

		return cmd
	}

	// Check for new version
	update, err := depMgr.GetUpdateFor(dependency.Dependency{Registry: dep.Registry, Name: dep.Name, Version: dep.Version})
	if err != nil {
		log.Warn(err)
	}

	// Message
	if update.NonBreaking == "" && update.Breaking == "" {
		msg <- ""
	} else if update.NonBreaking != "" {
		msg <- fmt.Sprintf("New version of this command is available, please update it using:\n    %s", getInstallCmd(update.NonBreaking))
	} else {
		msg <- fmt.Sprintf("New version of this command is available, but it may introduce some BREAKING CHANGES. Please consider updating it using:\n    %s", getInstallCmd(update.Breaking))
	}
}

func setupLogProcessor(cmd *exec.Cmd, wg *sync.WaitGroup) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	stdoutLogProcessor := log.NewLogProcessor()
	stdoutLogProcessor.DefaultLevel = log.InfoLevel
	stdoutLogProcessor.Input = stdoutPipe
	stdoutLogProcessor.Logger = log.DefaultLogger
	go func() {
		stdoutLogProcessor.Process()
		wg.Done()
	}()

	wg.Add(1)
	stderrLogProcessor := log.NewLogProcessor()
	stderrLogProcessor.DefaultLevel = log.ErrorLevel
	stderrLogProcessor.Input = stderrPipe
	stderrLogProcessor.Logger = log.ErrorLogger
	go func() {
		stderrLogProcessor.Process()
		wg.Done()
	}()
}
