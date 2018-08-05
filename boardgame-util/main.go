/*

	boardgame-util is a comprehensive CLI tool to help administer projects
	built with boardgame. All of its substantive functionality is implemented
	in sub-libraries in lib/, which can be used directly if necessary.

	The canonical help documentation is provided by `boardgame-util help`.

*/
package main

import (
	"github.com/bobziuchkovski/writ"
	"os"
	"strings"
)

func makeConfigs(commands []SubcommandObject) []*writ.Command {

	result := make([]*writ.Command, len(commands))

	for i, cmd := range commands {
		result[i] = &writ.Command{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Aliases:     cmd.Aliases(),
			Subcommands: makeConfigs(cmd.SubcommandObjects()),
		}
	}

	return result
}

func setupHelp(cmdNames []string, cmd *writ.Command, obj SubcommandObject) {

	cmdNames = append(cmdNames, obj.Name())

	baseSubCommands := obj.SubcommandObjects()

	if len(baseSubCommands) > 0 {

		subCmdNames := make([]string, len(baseSubCommands))

		for i, obj := range baseSubCommands {
			subCmdNames[i] = obj.Name()

			subCmd := cmd.Subcommand(obj.Name())
			setupHelp(cmdNames, subCmd, obj)
		}

		group := cmd.GroupCommands(subCmdNames...)
		group.Header = "Subcommands:"
		cmd.Help.CommandGroups = append(cmd.Help.CommandGroups, group)

	}

	cmd.Help.Usage = "Usage: " + strings.Join(cmdNames, " ")

}

func main() {
	mainImpl(os.Args)
}

func mainImpl(args []string) {
	b := &BoardgameUtil{}

	setupParents(b, nil)

	cmd := &writ.Command{
		Name: b.Name(),
	}

	baseSubCommands := b.SubcommandObjects()

	cmd.Subcommands = makeConfigs(baseSubCommands)

	setupHelp(nil, cmd, b)

	b.Help.base = cmd

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
