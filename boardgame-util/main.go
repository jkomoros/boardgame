/*

	boardgame-util is a comprehensive CLI tool to help administer projects
	built with boardgame.

*/
package main

import (
	"github.com/bobziuchkovski/writ"
	"os"
)

func makeConfigs(commands []SubcommandObject) []*writ.Command {
	result := make([]*writ.Command, len(commands))

	for i, cmd := range commands {
		result[i] = &writ.Command{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Aliases:     cmd.Aliases(),
		}
	}

	return result
}

func main() {
	mainImpl(os.Args)
}

func mainImpl(args []string) {
	b := &BoardgameUtil{}

	cmd := &writ.Command{
		Name: b.Name(),
	}

	cmd.Subcommands = makeConfigs(b.SubcommandObjects())

	b.Help.base = cmd

	baseUsage := "Usage: " + b.Name() + " "

	cmd.Help.Usage = baseUsage + "COMMAND [OPTION]... [ARG]..."
	cmd.Subcommand(b.Help.Name()).Help.Usage = b.Help.Name() + " SUBCOMMAND"
	cmd.Subcommand(b.Db.Name()).Help.Usage = b.Db.Name() + " SUBCOMMAND"

	path, positional, err := cmd.Decode(args[1:])

	if err != nil {
		path.Last().ExitHelp(err)
	}

	//TODO: this dispatch table should go straight to b.Dispatch, which
	//returns a subcommand, which is thne called Run().
	switch path.String() {
	case b.Name():
		b.Run(path, positional)
	case b.Name() + " " + b.Help.Name():
		b.Help.Run(path, positional)
	case b.Name() + " " + b.Db.Name():
		b.Db.Run(path, positional)
	default:
		panic("BUG: new subcomand that wasn't added to dispatch table yet")
	}

}
