package main

import (
	"errors"
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/stub"
	"os"
	"os/exec"
)

type Stub struct {
	baseSubCommand

	Dir  string
	Fast bool
}

func (s *Stub) Run(p writ.Path, positional []string) {

	b := s.Base()

	if len(positional) != 1 {
		p.Last().ExitHelp(errors.New("No gamename provided"))
	}

	gameName := positional[0]

	opt := &stub.Options{
		Name: gameName,
	}

	if !s.Fast {
		opt = stub.InteractiveOptions(nil, nil, gameName)
	}

	files, err := stub.Generate(opt)

	if err != nil {
		b.errAndQuit("Couldn't generate stubs: " + err.Error())
	}

	if err := files.Save(s.Dir, false); err != nil {
		b.errAndQuit("Couldn't save generated files: " + err.Error())
	}

	if _, err := os.Stat(gameName); os.IsNotExist(err) {
		b.errAndQuit("Unexpected error: game directory didn't exist after saving")
	}

	cmd := exec.Command("go", "generate")
	cmd.Dir = gameName
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		b.errAndQuit("Go generate failed: " + err.Error())
	}

	fmt.Println("Generated game in " + gameName)

}

func (s *Stub) Name() string {
	return "stub"
}

func (s *Stub) Description() string {
	return "Generate starter game with reasonable boilerplate"
}

func (s *Stub) Usage() string {
	return "GAMENAME"
}

func (s *Stub) HelpText() string {
	return s.Name() +

		` generates a starter game stub based on options provided interactively at the command prompt.

GAMENAME is the base name of the game. It should be short, unique, and be a valid go package name, for example "checkers", "tictactoe".`

}

func (s *Stub) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"dir", "d"},
			Description: "The directory to save the generated game folder in. Defaults to '.'",
			Decoder:     writ.NewOptionDecoder(&s.Dir),
		},
		{
			Names:       []string{"fast", "f"},
			Description: "If provided, skips the interactive prompts and just uses the defaults.",
			Decoder:     writ.NewFlagDecoder(&s.Fast),
			Flag:        true,
		},
	}
}
