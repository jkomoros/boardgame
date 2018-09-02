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

	if err := opt.Validate(); err != nil {
		return nil, errors.New("Options didn't validate: " + err.Error())
	}

	result := make(TemplateSet, len(templateMap))

	for name, contents := range templateMap {

		if opt.SuppressTest && strings.Contains(name, "main_test.go") {
			continue
		}

		if opt.SuppressPhase && strings.Contains(name, "enum.go") {
			continue
		}

		if opt.SuppressClientRenderGame && strings.Contains(name, "boardgame-render-game") {
			continue
		}

		if opt.SuppressClientRenderPlayerInfo && strings.Contains(name, "boardgame-render-player-info") {
			continue
		}

		if opt.SuppressComponentsStubs && strings.Contains(name, "components.go") {
			continue
		}

		if opt.SuppressMovesStubs && strings.Contains(name, "moves") {
			continue
		}

		//There are three entries for moves.go; in every case we skip either
		//one or two of them.
		if opt.SuppressPhase {
			if strings.Contains(name, "moves_") {
				continue
			}
		} else {
			if strings.Contains(name, "moves.go") {
				continue
			}
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

	if err := opt.Validate(); err != nil {
		return nil, errors.New("Options didn't validate: " + err.Error())
	}

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
	"{{.Name}}/main.go":         templateContentsMainGo,
	"{{.Name}}/enum.go":         templateContentsEnumGo,
	"{{.Name}}/main_test.go":    templateContentsMainTestGo,
	"{{.Name}}/player_state.go": templateContentsPlayerStateGo,
	"{{.Name}}/game_state.go":   templateContentsGameStateGo,
	"{{.Name}}/moves.go":        templateContentsMovesGo,
	"{{.Name}}/moves_setup.go":  templateContentsMovesGo,
	"{{.Name}}/moves_normal.go": templateContentsMovesGo,
	"{{.Name}}/components.go":   templateContentComponentsGo,
	"{{.Name}}/client/{{.Name}}/boardgame-render-game-{{.Name}}.html":        templateContentsRenderGameHtml,
	"{{.Name}}/client/{{.Name}}/boardgame-render-player-info-{{.Name}}.html": templateContentsRenderPlayerInfoHtml,
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

{{if .EnableExampleDynamicComponentValues }}
func (g *gameDelegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.ConfigurableSubState {
	if deck.Name() == exampleCardDeckName {
		return new(exampleCardDynamicValues)
	}
	return nil
}

{{end}}
func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {

	{{if .EnableExampleDeck -}}
	game := state.ImmutableGameState().(*gameState)
	if c.Deck().Name() == exampleCardDeckName {
		return game.DrawDeck, nil
	}
	return nil, errors.New("Unknown deck: " + c.Deck().Name())
	{{- else -}}
	return nil, errors.New("Not yet implemented")
	{{- end}}

}

{{if .EnableExampleDeck }}
func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game := state.GameState().(*gameState)
	return game.DrawDeck.Shuffle()
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		exampleCardDeckName: newExampleCardDeck(),
	}
}

{{end}}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		moves.Add(
			auto.MustConfig(new(moves.NoOp),
				with.MoveName("Example No Op Move"),
				with.HelpText("This move is an example that is always legal and does nothing. It exists to show how to return moves and make sure 'go test' works from the beginning, but you should remove it."),
			),
		),
	)

}

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}

`

const templateContentsEnumGo = `package {{.Name}}

//boardgame:codegen
const(
	//Because the naked Phase exists, this will be a TreeEnum. See package doc for "boardgame/enum" for more.
	Phase = iota
	PhaseSetUp
	PhaseNormal
)

`

const templateContentsPlayerStateGo = `package {{.Name}}

import (
	"github.com/jkomoros/boardgame"
)

//boardgame:codegen
type playerState struct {
	boardgame.BaseSubState
	playerIndex         boardgame.PlayerIndex
	{{if .EnableExampleDeck -}}
	Hand boardgame.Stack ` + "`stack:\"examplecards\" sanitize:\"len\"`" + `
	{{- end}}
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}
`

const templateContentsGameStateGo = `package {{.Name}}

import (
	"github.com/jkomoros/boardgame"{{if not .SuppressPhase}}
	"github.com/jkomoros/boardgame/enum"{{- end}}
	"github.com/jkomoros/boardgame/moves/roundrobinhelpers"
)

//boardgame:codegen
type gameState struct {
	//Use roundrobinhelpers so roundrobin moves can be used without any changes
	roundrobinhelpers.BaseGameState
	{{if not .SuppressCurrentPlayer -}}
	//DefaultGameDelegate will automatically return this from CurrentPlayerIndex
	CurrentPlayer boardgame.PlayerIndex
	{{- end}}
	{{if not .SuppressPhase -}}
	//DefaultGameDelegate will automatically return this from PhaseEnum, CurrentPhase.
	Phase enum.Val ` + "`enum:\"Phase\"`" + `
	{{- end}}
	{{if .EnableExampleDeck -}}
	DrawDeck boardgame.Stack ` + "`stack:\"examplecards\"`" + `
	{{- end}}
}

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)

	players := make([]*playerState, len(state.ImmutablePlayerStates()))

	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}
`

const templateContentsMovesGo = `package {{.Name}}

//TODO: define your move structs here. Don't forget the 'boardgame:codegen'
//magic comment, and don't forget to return them from
//delegate.ConfigureMoves().
{{if not .SuppressPhase}}

//Typically you create a separate file for moves of each major phase, and put
//the moves for that phase in it.
{{- end}}

`

const templateContentComponentsGo = `package {{.Name}}

{{if .EnableExampleDeck }}
import (
	"github.com/jkomoros/boardgame"
)

const numCards = 10
const exampleCardDeckName = "examplecards"

//boardgame:codegen
type exampleCard struct {
	boardgame.BaseComponentValues
	Value int
}

{{if .EnableExampleDynamicComponentValues}}
//boardgame:codegen
type exampleCardDynamicValues struct {
	boardgame.BaseSubState
	boardgame.BaseComponentValues
	DynamicValue int
}

{{end}}
//newExampleCardDeck returns a new deck for examplecards.
func newExampleCardDeck() *boardgame.Deck {
	deck := boardgame.NewDeck()

	for i := 0; i < numCards; i++ {
		deck.AddComponent(&exampleCard{
			Value: i + 1,
		})
	}

	return deck
}
{{else}}
//components.go is where you generally define your component structs and deck
//constructors.
{{end}}

`

const templateContentsMainTestGo = `package {{.Name}}

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestNewManager(t *testing.T) {

	//A lot of validation goes on in boardgame.NewGameManager, which means
	//that simply testing that we don't get an error with our delegate is a
	//useful test. However, this is not a very robust test because it doesn't
	//verify that moves are legal when they should be or do the right things,
	//among other things. Typically you should also create golden game
	//examples to verify the behavior of your game matches expectations. See
	//TUTORIAL.md for more on goldens.

	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())

	assert.For(t).ThatActual(manager).IsNotNil()

	assert.For(t).ThatActual(err).IsNil()

}


`

const templateContentsRenderGameHtml = `<link rel="import" href="../../bower_components/polymer/polymer-element.html">
<link rel="import" href="../../src/boardgame-base-game-renderer.html">

<dom-module id="boardgame-render-game-{{.Name}}">
  <template>
  This is where you game should render itself. See boardgame/server/README.md for more on the components you can use, or check out the examples in boardgame/examples.
  </template>

  <script>

    class BoardgameRenderGame{{uppercaseFirst .Name}} extends BoardgameBaseGameRenderer {

      static get is() {
        return "boardgame-render-game-{{.Name}}"
      }

      //We don't need to compute any properties that BoardgameBaseGamErenderer
      //doesn't have.

    }

    customElements.define(BoardgameRenderGame{{uppercaseFirst .Name}}.is, BoardgameRenderGame{{uppercaseFirst .Name}};

  </script>
</dom-module>
`

const templateContentsRenderPlayerInfoHtml = `<link rel="import" href="../../bower_components/polymer/polymer-element.html">

<dom-module id="boardgame-render-player-info-{{.Name}}">
  <template>
    This is where you render info on player, typically using &gt;boardgame-status-text&lt;.
  </template>

  <script>

    class BoardgameRenderPlayerInfo{{uppercaseFirst .Name}} extends Polymer.Element {

      static get is() {
        return "boardgame-render-player-info-{{.Name}}"
      }
    }

    customElements.define(BoardgameRenderPlayerInfo{{uppercaseFirst .Name}}.is, BoardgameRenderPlayerInfo{{uppercaseFirst .Name}});

  </script>
</dom-module>
`
