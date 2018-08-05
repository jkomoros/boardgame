/*

	boardgame-util is a comprehensive CLI tool to help administer projects
	built with boardgame.

*/
package main

import (
	"github.com/bobziuchkovski/writ"
	"os"
)

func main() {
	mainImpl(os.Args)
}

func mainImpl(args []string) {
	boardgameUtil := &BoardgameUtil{}

	cmd := writ.New(cmdBase, boardgameUtil)

	boardgameUtil.Help.base = cmd

	baseUsage := "Usage: " + cmdBase + " "

	cmd.Help.Usage = baseUsage + "COMMAND [OPTION]... [ARG]..."
	cmd.Subcommand("help").Help.Usage = "help SUBCOMMAND"

	path, positional, err := cmd.Decode(args[1:])

	if err != nil {
		path.Last().ExitHelp(err)
	}

	switch path.String() {
	case cmdBase:
		boardgameUtil.Run(path, positional)
	case cmdBase + " " + cmdHelp:
		boardgameUtil.Help.Run(path, positional)
	default:
		panic("BUG: new subcomand that wasn't added to dispatch table yet")
	}

}
