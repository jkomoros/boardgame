package main

import (
	"errors"
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/stub"
)

type Stub struct {
	baseSubCommand

	Dir string
}

func (s *Stub) Run(p writ.Path, positional []string) {

	b := s.Base()

	if len(positional) != 1 {
		p.Last().ExitHelp(errors.New("No gamename provided"))
	}

	opt := stub.InteractiveOptions(nil, nil, positional[0])

	files, err := stub.Generate(opt)

	if err != nil {
		b.errAndQuit("Couldn't generate stubs: " + err.Error())
	}

	if err := files.Save(s.Dir, false); err != nil {
		b.errAndQuit("Couldn't save generated files: " + err.Error())
	}

	//TODO: run go generate on the output.

	fmt.Println("Generated game in " + positional[0])

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

GAMENAME is the base name of the game. It should be short, unique, and not have spaces, for example "checkers", "tic-tac-toe".`

}

func (s *Stub) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"dir", "d"},
			Description: "The directory to save the generated game folder in. Defaults to '.'",
			Decoder:     writ.NewOptionDecoder(&s.Dir),
		},
	}
}
