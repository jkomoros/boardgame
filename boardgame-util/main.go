/*

	boardgame-util is a comprehensive CLI tool to help administer projects
	built with boardgame.

*/
package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
	"os"
)

const cmdBase = "boardgame-util"
const cmdHelp = "help"

type BoardgameUtil struct {
	Help Help `command:"help" description:"Prints help for a specific subcommand"`
}

type Help struct {
}

func (b *BoardgameUtil) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New("COMMAND is required"))
}

func (h *Help) Run(p writ.Path, positional []string) {
	if len(positional) != 1 {
		p.Last().ExitHelp(errors.New(cmdHelp + " requires one argument SUBCOMMAND"))
	}

	p.Last().ExitHelp(errors.New("No subcommands yet"))

}

func main() {
	mainImpl(os.Args)
}

func mainImpl(args []string) {
	boardgameUtil := &BoardgameUtil{}

	cmd := writ.New(cmdBase, boardgameUtil)

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
