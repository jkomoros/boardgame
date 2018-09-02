package stub

import (
	"bytes"
	"errors"
	"strings"
	"text/template"
)

//TemplateSet is a collection of templates that can create a derived and
//expanded FileContents when given an Options struct.
type TemplateSet map[string]*template.Template

//lowercaseFirst ensures first character is lower case
func lowercaseFirst(in string) string {
	if len(in) == 0 {
		return in
	}
	return strings.ToLower(in[0:1]) + in[1:]
}

//uppercastFirst ensures first characer is upper case
func uppercaseFirst(in string) string {
	if len(in) == 0 {
		return in
	}
	return strings.ToUpper(in[0:1]) + in[1:]
}

//DefaultTemplateSet returns the default template set for this stub.
func DefaultTemplateSet(opt *Options) (TemplateSet, error) {
	result := make(TemplateSet, len(templateMap))

	for name, contents := range templateMap {

		if opt.SuppressTest && strings.Contains(name, "main_test.go") {
			continue
		}

		tmpl := template.New(name)
		tmpl.Funcs(template.FuncMap{
			"lowercaseFirst": lowercaseFirst,
			"uppercaseFirst": uppercaseFirst,
		})
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

	result := make(FileContents)

	for name, tmpl := range t {

		nameTmpl := template.New("pass")
		nameTmpl, err := nameTmpl.Parse(name)

		if err != nil {
			return nil, errors.New(name + " could not be intpreted as a template: " + err.Error())
		}

		buf := new(bytes.Buffer)

		if err := nameTmpl.Execute(buf, opt); err != nil {
			return nil, errors.New(name + " name template could not be executed: " + err.Error())
		}

		resolvedName := buf.String()

		contentBuf := new(bytes.Buffer)

		if err := tmpl.Execute(contentBuf, opt); err != nil {
			return nil, errors.New(name + " template could not be executed: " + err.Error())
		}

		result[resolvedName] = contentBuf.Bytes()

	}

	return result, nil

}

var templateMap = map[string]string{
	"{{.Name}}/main.go":      templateContentsMainGo,
	"{{.Name}}/main_test.go": templateContentsMainTestGo,
	"{{.Name}}/state.go":     templateContentsStateGo,
}

const templateContentsMainGo = `{{if .Description -}}
/*

	{{.Name}} is {{lowercaseFirst .Description}}

*/
{{- end}}
package {{.Name}}

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

{{if .DisplayName -}}
func (g *gameDelegate) DisplayName() string {
	return "{{.DisplayName}}"
}
{{- end}}

{{if .Description -}}
func (g *gameDelegate) Description() string {
	return "{{.Description}}"
}
{{- end}}

{{if .MinNumPlayers -}}
func (g *gameDelegate) MinNumPlayers() int {
	return {{.MinNumPlayers}}
}
{{- end}}

{{if .MaxNumPlayers -}}
func (g *gameDelegate) MaxNumPlayers() int {
	return {{.MaxNumPlayers}}
}
{{- end}}

{{if .DefaultNumPlayers -}}
func (g *gameDelegate) DefaultNumPlayers() int {
	return {{.DefaultNumPlayers}}
}
{{- end}}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {
	return &playerState{
		playerIndex: index,
	}
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

const templateContentsStateGo = `package {{.Name}}

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/roundrobinhelpers"
)

//boardgame:codegen
type gameState struct {
	//Use roundrobinhelpers so roundrobin moves can be used without any changes
	roundrobinhelpers.BaseGameState
	{{if not .SuppressCurrentPlayer -}}
	CurrentPlayer boardgame.PlayerIndex
	{{- end}}
}

//boardgame:codegen
type playerState struct {
	boardgame.BaseSubState
	playerIndex         boardgame.PlayerIndex
}

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)

	players := make([]*playerState, len(state.ImmutablePlayerStates()))

	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}
`

const templateContentsMainTestGo = `package {{.Name}}

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())

	assert.For(t).ThatActual(manager).IsNotNil()

	assert.For(t).ThatActual(err).IsNil()

}


`
