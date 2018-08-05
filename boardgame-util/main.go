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

	//TODO: actually expand the list inline if any of the commands return a
	//non-zero-length slice from SubcommandObjects.

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
	cmd.Help.Usage = baseUsage + b.Usage()

	for _, obj := range b.SubcommandObjects() {
		cmd.Subcommand(obj.Name()).Help.Usage = obj.Name() + " " + obj.Usage()
	}

	path, positional, err := cmd.Decode(args[1:])

	if err != nil {
		path.Last().ExitHelp(err)
	}

	subcommandObj := selectSubcommandObject(b, path)

	if subcommandObj == nil {
		panic("BUG: one of the subcommands didn't enumerate all subcommands")
	}

	subcommandObj.Run(path, positional)

}
