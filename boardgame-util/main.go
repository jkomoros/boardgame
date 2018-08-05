/*

	boardgame-util is a comprehensive CLI tool to help administer projects
	built with boardgame. All of its substantive functionality is implemented
	in sub-libraries in lib/, which can be used directly if necessary.

	The canonical help documentation is provided by `boardgame-util help`.

*/
package main

import (
	"os"
	"strings"
)

func main() {
	mainImpl(os.Args)
}

func mainImpl(args []string) {
	b := &BoardgameUtil{}

	setupParents(b, nil, nil)

	cmd := b.WritCommand()

	path, positional, err := cmd.Decode(args[1:])

	if err != nil {
		path.Last().ExitHelp(err)
	}

	subcommandObj := selectSubcommandObject(b, strings.Split(path.String(), " "))

	if subcommandObj == nil {
		panic("BUG: one of the subcommands didn't enumerate all subcommands")
	}

	subcommandObj.Run(path, positional)

}
