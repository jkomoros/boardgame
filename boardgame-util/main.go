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

//expandSubcommandObjects will return an expanded list of input, where each
//command that returns non-nil SubcommandObjects will be followed by those in
//the list.
func expandSubcommandObjects(commands []SubcommandObject) []SubcommandObject {

	if len(commands) == 0 {
		return nil
	}

	var result []SubcommandObject

	for _, cmd := range commands {
		result = append(result, cmd)
		//if expandSubcommandObjects is nil, then append will leave result the same
		result = append(result, expandSubcommandObjects(cmd.SubcommandObjects())...)
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

	expandedSubcommands := expandSubcommandObjects(b.SubcommandObjects())

	cmd.Subcommands = makeConfigs(expandedSubcommands)

	cmdNames := make([]string, len(expandedSubcommands))

	for i, obj := range expandedSubcommands {
		cmdNames[i] = obj.Name()
	}

	group := cmd.GroupCommands(cmdNames...)
	group.Header = "General commands:"
	cmd.Help.CommandGroups = append(cmd.Help.CommandGroups, group)

	b.Help.base = cmd

	baseUsage := "Usage: " + b.Name() + " "
	cmd.Help.Usage = baseUsage + b.Usage()

	for _, obj := range expandedSubcommands {
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
