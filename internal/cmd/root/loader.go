package root

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/g2a-com/klio/internal/cmd"
	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/manager"
	"github.com/g2a-com/klio/internal/env"
	"github.com/g2a-com/klio/internal/log"
	"github.com/g2a-com/klio/internal/project"
	"github.com/g2a-com/klio/internal/scope"
	"github.com/spf13/cobra"
)

const (
	updateTimeout         = 5 * time.Second
	commandConfigFileName = "command.yaml"
)

func loadExternalCommand(ctx context.CLIContext, rootCmd *cobra.Command, dep dependency.DependenciesIndexEntry) {
	if c, _, _ := rootCmd.Find([]string{dep.Alias}); c != rootCmd {
		log.Spamf("cannot register already registered command '%s'", dep.Alias)
		return
	}

	cmdConfig, err := cmd.LoadConfig(filepath.Join(dep.Path, commandConfigFileName))
	if err != nil {
		log.Warnf("Cannot load command: %s", err)
		return
	}

	updatedDep, err := autoDownloadCommand(&ctx, dep)
	if err != nil {
		log.Warnf("Cannot auto update command: %s", err)
	} else {
		dep = *updatedDep
	}

	newCmd := &cobra.Command{
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
			case "g2a-cli/v1beta1", "g2a-cli/v1beta2", "g2a-cli/v1beta3", "g2a-cli/v1beta4", "klio/v1":
				externalCmd.Stdout = os.Stdout
				externalCmd.Stderr = os.Stderr
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
			timeoutChannel := make(chan bool, 1)
			skipUpdates := false

			if skipUpdatesStr, exists := os.LookupEnv(env.KLIO_SKIP_UPDATE_CHECK); exists {
				if v, err := strconv.ParseBool(skipUpdatesStr); err != nil {
					log.Warnf("Could not parse boolean value of %s, err: %s", env.KLIO_SKIP_UPDATE_CHECK, err.Error())
				} else {
					skipUpdates = v
				}
			}

			if !skipUpdates {
				go getUpdateMessage(ctx, dep, updateMsgChannel)
				go func() {
					time.Sleep(updateTimeout)
					timeoutChannel <- true
				}()
			}

			log.Debugf(`Running %s "%s"`, externalCmdPath, strings.Join(args, `" "`))

			if err := externalCmd.Start(); err != nil {
				log.Fatal(err)
			}

			wg.Wait()
			err = externalCmd.Wait()

			if !skipUpdates {
				select {
				case msg := <-updateMsgChannel:
					if msg != "" {
						for _, line := range strings.Split(msg, "\n") {
							log.ErrorLogger.Warn(line)
						}
					}
				case <-timeoutChannel:
					break
				}
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
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Parses the completion info provided by cobra.Command. This should be formatted similar to:
			//   help	Help about any command
			//   :4
			//   Completion ended with directive: ShellCompDirectiveNoFileComp
			var buffer bytes.Buffer
			var externalCmd *exec.Cmd
			externalCmdPath := filepath.Join(dep.Path, cmdConfig.BinPath)
			completionArgs := []string{"__complete"}
			if runtime.GOOS == "windows" {
				args = append([]string{"/c", externalCmdPath}, args...)
				externalCmdPath = "cmd"
			}
			completionArgs = append(completionArgs, args...)
			completionArgs = append(completionArgs, toComplete)
			externalCmd = exec.Command(externalCmdPath, completionArgs...)
			externalCmd.Stdin = os.Stdin
			externalCmd.Stdout = &buffer
			externalCmd.Env = os.Environ()
			externalCmd.Env = append(externalCmd.Env, fmt.Sprintf("%s=%t", env.KLIO_SKIP_UPDATE_CHECK, true))

			if err := externalCmd.Start(); err != nil {
				log.Fatal(err)
			}

			if err = externalCmd.Wait(); err != nil {
				switch e := err.(type) {
				case *exec.ExitError:
					os.Exit(e.ExitCode())
				default:
					log.Fatal(err)
				}
			}

			output := buffer.String()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			lines := strings.Split(strings.Trim(output, "\n"), "\n")
			var results []string
			for _, line := range lines {
				if strings.HasPrefix(line, ":") {
					// Special marker in output to indicate the end
					directive, err := strconv.Atoi(line[1:])
					if err != nil {
						return results, cobra.ShellCompDirectiveError
					}
					return results, cobra.ShellCompDirective(directive)
				}
				results = append(results, line)
			}
			return []string{}, cobra.ShellCompDirectiveError
		},
	}
	rootCmd.AddCommand(newCmd)
}

func getUpdateMessage(ctx context.CLIContext, dep dependency.DependenciesIndexEntry, msg chan<- string) {
	depMgr := manager.NewManager()

	getInstallCmd := func(ver string) string {
		installMsg := fmt.Sprintf("%s get", ctx.Config.CommandName)

		if ctx.Paths.IsGlobal(dep.Path) {
			installMsg += " -g"
		}
		installMsg += fmt.Sprintf(" %s --version %s --from %s", dep.Name, ver, dep.Registry)
		if dep.Name != dep.Alias {
			installMsg += fmt.Sprintf(" --as %s", dep.Alias)
		}

		return installMsg
	}

	// Check for new version
	update, err := depMgr.GetUpdateFor(dependency.Dependency{Registry: dep.Registry, Name: dep.Name, Version: dep.Version})
	if err != nil {
		log.Warn(err)
	}

	// message
	if update.NonBreaking == "" && update.Breaking == "" {
		msg <- ""
	} else if update.NonBreaking != "" {
		msg <- fmt.Sprintf("New version of this command is available, please update it using:\n    %s", getInstallCmd(update.NonBreaking))
	} else {
		msg <- fmt.Sprintf("New version of this command is available, but it may introduce some BREAKING CHANGES. Please consider updating it using:\n    %s", getInstallCmd(update.Breaking))
	}
}

func autoDownloadCommand(ctx *context.CLIContext, dep dependency.DependenciesIndexEntry) (*dependency.DependenciesIndexEntry, error) {
	args := os.Args[1:]
	if len(args) == 0 {
		return &dep, nil
	} else {
		command := args[0]
		if command != dep.Alias {
			return &dep, nil
		}
	}

	skipAutoDownload := false
	if skipAutoDownloadStr, exists := os.LookupEnv(env.KLIO_SKIP_PROJECT_COMMAND_AUTO_DOWNLOAD); exists {
		if v, err := strconv.ParseBool(skipAutoDownloadStr); err != nil {
			return nil, fmt.Errorf("could not parse boolean value of %s, err: %s", env.KLIO_SKIP_PROJECT_COMMAND_AUTO_DOWNLOAD, err.Error())
		} else {
			skipAutoDownload = v
		}
	}

	updatedDep := dep
	if isProjectScope := ctx.Paths.IsProject(dep.Path); isProjectScope && !skipAutoDownload {
		projectConfig, err := project.LoadProjectConfig(ctx.Paths.ProjectConfigFile)
		if err != nil {
			return nil, fmt.Errorf("cannot load project config: %s", err)
		}
		projectDependency := projectConfig.GetDependency(dep.Alias)

		if projectDependency != nil && projectDependency.Version != "" && projectDependency.Version != dep.Version {
			localScope, err := scope.NewLocal(ctx, false, false)
			if err != nil {
				return nil, fmt.Errorf("cannot initialize local scope: %s", err)
			}

			log.Infof(`Installed %[1]s@%[2]s but project config defines %[1]s@%[3]s, downloading %[1]s@%[3]s now...`, dep.Alias, dep.Version, projectDependency.Version)

			_, installedDep, err := localScope.InstallDependencies([]dependency.Dependency{
				{
					Name:     dep.Name,
					Registry: dep.Registry,
					Version:  projectDependency.Version,
					Alias:    dep.Alias,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("cannot install dependency: %s", err)
			} else {
				updatedDep = installedDep[0]
				updatedDep.Path = filepath.Join(ctx.Paths.ProjectInstallDir, updatedDep.Path)
			}
		}
	}
	return &updatedDep, nil
}
