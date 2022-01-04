package flagx

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
)

func New(args ...string) *App {
	if len(args) == 0 {
		args = os.Args
	}
	return &App{name: args[0], args: args[1:], commands: map[string]Command{}}
}

type App struct {
	args        []string
	name        string
	description string
	commands    map[string]Command
}

func (app *App) AddCommand(name string, command Command) *App {
	app.commands[name] = command
	return app
}

func (app *App) Usage() {
	if app.description != "" {
		fmt.Fprintf(os.Stderr, "%s\n", app.description)
	}
	fmt.Fprintf(os.Stderr, "命令格式: %s [命令] [...参数]\n", app.name)
	fmt.Fprintf(os.Stderr, "命令说明:\n")

	max := 0
	for name, command := range app.commands {
		if !isHidden(command) {
			if l := len(name); l > max {
				max = l
			}
		}
	}

	for name, command := range app.commands {
		if !isHidden(command) {
			fmt.Fprintf(os.Stderr, "  %*s  %s\n", -max, name, getUsage(command))
		}
	}

	fmt.Fprintln(os.Stderr)
}

func (app *App) Run(ctx context.Context) {
	if len(app.args) == 1 && app.args[0] == "" {
		app.Usage()
		os.Exit(1)
	}

	if len(app.args) == 0 || strings.HasPrefix(app.args[0], "-") {
		app.Usage()
		os.Exit(1)
	}

	commandName := app.args[0]
	command, find := app.commands[commandName]
	if !find {
		app.Usage()
		os.Exit(1)
	}

	args := app.args
	flagOpt := func(flagSet *flag.FlagSet) { flagSet.Init(commandName, flag.ContinueOnError) }
	var err error
	if args, err = Parse(command, args[1:], flagOpt); err != nil {
		os.Exit(1)
	}

	if err := command.Run(ctx, args); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

type Command interface {
	Run(ctx context.Context, args []string) error
}

func isHidden(c Command) bool {
	if x, ok := c.(hiddenable); ok {
		return x.Hidden()
	}
	return false
}

func getUsage(c Command) string {
	if x, ok := c.(usageable); ok {
		return x.Usage()
	}
	return ""
}

type (
	usageable  interface{ Usage() string }
	hiddenable interface{ Hidden() bool }
)
