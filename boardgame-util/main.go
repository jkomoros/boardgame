/*
boardgame-util is a comprehensive CLI tool to help administer projects built
with boardgame. All of its substantive functionality is implemented in
sub-libraries in lib/, which can be used directly if necessary.

The canonical help documentation is provided by `boardgame-util help`.
*/
package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bobziuchkovski/writ"
)

func main() {
	os.Exit(mainImpl(os.Args))
}

var (
	commandStdout io.Writer = os.Stdout
	commandStderr io.Writer = os.Stderr
)

type commandExit struct{ code int }

func quit(code int) {
	panic(commandExit{code: code})
}

func exitHelp(command *writ.Command, err error) {
	writer := commandStdout
	code := 0
	if err != nil {
		writer = commandStderr
		code = 1
	}
	if writeErr := command.WriteHelp(writer); writeErr != nil {
		fmt.Fprintf(commandStderr, "Couldn't render help: %v\n", writeErr)
		quit(1)
	}
	if err != nil {
		fmt.Fprintf(commandStderr, "\nError: %s\n", err)
	}
	quit(code)
}

func mainImpl(args []string) (exitCode int) {
	b := &boardgameUtil{}

	setupParents(b, nil, nil)

	defer b.Cleanup()
	defer func() {
		if recovered := recover(); recovered != nil {
			if exit, ok := recovered.(commandExit); ok {
				exitCode = exit.code
				return
			}
			panic(recovered)
		}
	}()

	//Make sure that even if we get exited early we still clean up.
	c := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(c)
	defer close(done)

	go func() {
		select {
		case <-c:
			b.Cleanup()
			os.Exit(1)
		case <-done:
			return
		}
	}()

	cmd := b.WritCommand()

	path, positional, err := cmd.Decode(args[1:])

	if err != nil {
		exitHelp(path.Last(), err)
	}

	subcommandObj := selectSubcommandObject(b, strings.Split(path.String(), " "))

	if subcommandObj == nil {
		panic("BUG: one of the subcommands didn't enumerate all subcommands")
	}

	subcommandObj.Run(path, positional)
	return 0
}
