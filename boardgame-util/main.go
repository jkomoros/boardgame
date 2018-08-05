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
)

func makeHelp(cmd *writ.Command, obj SubcommandObject) writ.Help {

	var result writ.Help

	result.Header = obj.HelpText()

	baseSubCommands := obj.SubcommandObjects()

	if len(baseSubCommands) > 0 {

		subCmdNames := make([]string, len(baseSubCommands))
		for i, obj := range baseSubCommands {
			subCmdNames[i] = obj.Name()
		}

		group := cmd.GroupCommands(subCmdNames...)
		group.Header = "Subcommands:"
		result.CommandGroups = append(result.CommandGroups, group)

	}

	result.Usage = "Usage: " + FullName(obj) + obj.Usage()

	return result
}

func makeConfig(obj SubcommandObject) *writ.Command {

	cmd := &writ.Command{
		Name:        obj.Name(),
		Description: obj.Description(),
		Aliases:     obj.Aliases(),
		Subcommands: makeConfigs(obj.SubcommandObjects()),
	}

	cmd.Help = makeHelp(cmd, obj)

	return cmd
}

func makeConfigs(commands []SubcommandObject) []*writ.Command {
	result := make([]*writ.Command, len(commands))
	for i, obj := range commands {
		result[i] = makeConfig(obj)
	}
	return result
}

func main() {
	mainImpl(os.Args)
}

func mainImpl(args []string) {
	b := &BoardgameUtil{}

	setupParents(b, nil)

	cmd := makeConfig(b)

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
