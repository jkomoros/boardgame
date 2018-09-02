package stub

import (
	"errors"
	"text/template"
)

//TemplateSet is a collection of templates that can create a derived and
//expanded FileContents when given an Options struct.
type TemplateSet map[string]*template.Template

//DefaultTemplateSet returns the default template set for this stub.
func DefaultTemplateSet() (TemplateSet, error) {
	result := make(TemplateSet, len(templateMap))

	for name, contents := range templateMap {
		tmpl := template.New(name)
		tmpl, err := tmpl.Parse(contents)
		if err != nil {
			return nil, errors.New(name + " could not be parsed: " + err.Error())
		}
		result[name] = tmpl
	}
	return result, nil
}

//Generate generates FileContents based on this TemplateSet, using those
//options to expand. Names of files will also be run through templates and
//expanded.
func (t TemplateSet) Generate(opt *Options) (FileContents, error) {
	return nil, errors.New("Not yet implemented")
}

var templateMap = map[string]string{
	"{{.Name}}/main.go": templateContentsMainGo,
}

const templateContentsMainGo = `package {{.Name}}

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/moves/with"
)

/*

Call the code generation for readers and enums here, so "go generate" will generate code correctly.

*/
//go:generate boardgame-util codegen

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return "{{.Name}}"
}


func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return nil
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {
	return nil
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {

	return nil, errors.New("Not yet implemented")

}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	return nil

}

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}

`
