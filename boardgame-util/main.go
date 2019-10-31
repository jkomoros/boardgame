/*

boardgame-util is a comprehensive CLI tool to help administer projects built
with boardgame. All of its substantive functionality is implemented in
sub-libraries in lib/, which can be used directly if necessary.

The canonical help documentation is provided by `boardgame-util help`.

*/
package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	mainImpl(os.Args)
}

func mainImpl(args []string) {
	b := &boardgameUtil{}

	setupParents(b, nil, nil)

	defer b.Cleanup()

	//Make sure that even if we get exited early we still clean up.
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		<-c
		b.Cleanup()
		os.Exit(1)
	}()

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
