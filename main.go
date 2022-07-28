package main

import (
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/commands"
	actioncli "github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/mitchellh/cli"
)

var (
	//go:embed dev/VERSION
	versionCore                         string
	FullVersion, Revision, RevisionTime string
)

func main() {
	status, err := makeCLI(os.Args[1:], versionOutput()).Run()
	if err != nil {
		log.Info("%s", err)
	}
	os.Exit(status)
}

func makeCLI(args []string, version string) *cli.CLI {

	c := cli.NewCLI("actions-go-build", version)

	c.Args = args

	c.Commands = map[string]cli.CommandFactory{
		"test":               makeCommand(commands.Test),
		"build primary":      makeCommand(commands.BuildPrimary),
		"build verification": makeCommand(commands.BuildVerification),
		"build env describe": makeCommand(commands.BuildEnvDescribe),
		"build env dump":     makeCommand(commands.BuildEnvDump),
		"verify":             makeCommand(commands.Verify),
		"config":             makeCommand(commands.Config),
	}

	return c
}

type cmd struct {
	help, synopsis string
	run            func([]string) error
}

func (c *cmd) Help() string     { return c.help }
func (c *cmd) Synopsis() string { return c.synopsis }
func (c *cmd) Run(args []string) int {
	if err := c.run(append([]string{""}, args...)); err != nil {
		log.Info("%s", err)
		return 1
	}
	return 0
}

func makeCommand(command *actioncli.Command) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &cmd{
			help:     command.Help(),
			synopsis: command.Description(),
			run:      command.Execute,
		}, nil
	}
}

func version() string {
	if FullVersion != "" {
		return FullVersion
	}
	versionCore = strings.TrimSpace(versionCore)
	if versionCore == "" {
		versionCore = "0.0.0-unversioned"
	}
	return fmt.Sprintf("%s-local", versionCore)
}

func revision() string {
	if Revision == "" {
		return "(unknown revision)"
	}
	revisionString := fmt.Sprintf("(%s)", Revision[:8])
	if RevisionTime != "" {
		revisionString += fmt.Sprintf(" %s", RevisionTime)
	}
	return revisionString
}

func versionOutput() string {
	return fmt.Sprintf("v%s %s", version(), revision())
}
