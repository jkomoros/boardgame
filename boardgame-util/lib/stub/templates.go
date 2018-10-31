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

var postProcessReplacements = map[string]string{
	"[[BACKTICK]]": "`",
	"{[[":          "{{",
	"]]}":          "}}",
}

//postProcess to replace hard-to-escape literals moves.With different results.
func postProcess(in []byte) []byte {

	str := string(in)

	for find, replace := range postProcessReplacements {
		str = strings.Replace(str, find, replace, -1)
	}

	return []byte(str)

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

		result[resolvedName] = postProcess(contentBuf.Bytes())

	}

	return result, nil

}

var templateMap = map[string]string{
	"{{.Name}}/main.go":                                          templateContentsMainGo,
	"{{.Name}}/enum.go":                                          templateContentsEnumGo,
	"{{.Name}}/main_test.go":                                     templateContentsMainTestGo,
	"{{.Name}}/player_state.go":                                  templateContentsPlayerStateGo,
	"{{.Name}}/game_state.go":                                    templateContentsGameStateGo,
	"{{.Name}}/moves.go":                                         templateContentsDefaultMovesGo,
	"{{.Name}}/moves_setup.go":                                   templateContentsDefaultMovesGo,
	"{{.Name}}/moves_normal.go":                                  templateContentsNormalMovesGo,
	"{{.Name}}/components.go":                                    templateContentComponentsGo,
	"{{.Name}}/client/boardgame-render-game-{{.Name}}.js":        templateContentsRenderGameJs,
	"{{.Name}}/client/boardgame-render-player-info-{{.Name}}.js": templateContentsRenderPlayerInfoJs,
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
	"github.com/jkomoros/boardgame/base"
	"reflect"
	"strings"
)

/*

Call the code generation for readers and enums here, so "go generate" will generate code correctly.

*/
//go` +

	//Split this here so that running go gen with the whole module won't generate code here

	`:generate boardgame-util codegen

type gameDelegate struct {
	base.GameDelegate
}

var memoizedDelegateName string

func (g *gameDelegate) Name() string {

	//If our package name and delegate.Name() don't match, NewGameManager will
	//fail with an error. Given they have to be the same, we might as well
	//just ensure they are actually the same, via a one-time reflection.

	if memoizedDelegateName == "" {
		pkgPath := reflect.ValueOf(g).Elem().Type().PkgPath()
		pathPieces := strings.Split(pkgPath, "/")
		memoizedDelegateName = pathPieces[len(pathPieces)-1]
	}
	return memoizedDelegateName
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
func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		{{if .SuppressPhase }}
		moves.Add(
			auto.MustConfig(new(moves.NoOp),
				moves.WithMoveName("Example No Op Move"),
				moves.WithHelpText("This move is an example that is always legal and does nothing. It exists to show how to return moves and make sure 'go test' works from the beginning, but you should remove it."),
			),
		),
		{{ else -}}
		moves.AddOrderedForPhase(
			PhaseSetUp,
			{{if .EnableExampleDeck -}}
			//This move will keep on applying itself in round robin fashion
			//until all of the cards are dealt.
			auto.MustConfig(new(moves.DealComponentsUntilPlayerCountReached),
				moves.WithGameProperty("DrawStack"),
				moves.WithPlayerProperty("Hand"),
				moves.WithTargetCount(2),
			),
			{{- end }}
			//Because we used AddOrderedForPhase, this next move won't apply
			//until the move before it is done applying.
			auto.MustConfig(new(moves.StartPhase),
				moves.WithPhaseToStart(PhaseNormal, PhaseEnum),
				moves.WithHelpText("Move to the normal play phase."),
			),
		),
		{{if .EnableExampleMoves }}
		moves.AddForPhase(
			PhaseNormal,
			auto.MustConfig(new(moveDrawCard),
				moves.WithHelpText("Draw a card from the deck when it's your turn"),
			),
			//FinishTurn will advance to the next player automatically, when
			//playerState.TurnDone() is true.
			auto.MustConfig(new(moves.FinishTurn)),
		),
		{{- end}}
		{{- end }}
	)

}

{{if .EnableExampleDeck }}
func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		exampleCardDeckName: newExampleCardDeck(),
	}
}

{{end}}
{{if .EnableExampleConstants }}
func (g *gameDelegate) ConfigureConstants() boardgame.PropertyCollection {

	//ConfigureConstants isn't needed very often. It's useful to ensure a
	//constant value is available client-side, or if you want to use the value
	//in a struct tag.

	return boardgame.PropertyCollection{
		"numCards": numCards,
	}
}

{{end}}
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
		return game.DrawStack, nil
	}
	return nil, errors.New("Unknown deck: " + c.Deck().Name())
	{{- else -}}
	return nil, errors.New("Not yet implemented")
	{{- end}}

}

{{if .EnableExampleVariants}}
func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {

	//This is the only time that config is passed in, so we need to interpret
	//it now and set it as a property in GameState.
	targetCardsLeftVal := variant[variantKeyTargetCardsLeft]

	var targetCardsLeft int

	switch targetCardsLeftVal {
	case variantTargetCardsLeftShort:
		targetCardsLeft = 2
	case variantTargetCardsLeftDefault:
		targetCardsLeft = 0
	default:
		//This shouldn't happen because NewGame checks that given configs are
		//legal before passing to this method.
		return errors.New("Unknown value for " + variantKeyTargetCardsLeft + ": " + targetCardsLeftVal)
	}

	game := state.GameState().(*gameState)
	game.TargetCardsLeft = targetCardsLeft

	return nil

}

{{end}}

{{if .EnableExampleDeck }}
func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game := state.GameState().(*gameState)
	return game.DrawStack.Shuffle()
}

{{end}}
{{if .EnableExampleEndState }}
func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	//base.GameDelegate's CheckGameFinished checks this method and if true
	//looks at the score to see who won.

	//In this example, the game is over once all of the cards are gone.
	return state.ImmutableGameState().(*gameState).CardsDone()
}

{{end}}
{{if .DefaultNumPlayers -}}
func (g *gameDelegate) DefaultNumPlayers() int {
	return {{.DefaultNumPlayers}}
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
{{if .EnableExampleVariants}}

//values for the variant setup
const (
	variantKeyTargetCardsLeft = "target_cards_left"
)

const (
	variantTargetCardsLeftDefault = "default"
	variantTargetCardsLeftShort = "short"
)

func (g *gameDelegate) Variants() boardgame.VariantConfig {

	//variants are the legal configuration options that will be show in the
	//new game dialog.

	return boardgame.VariantConfig{
		variantKeyTargetCardsLeft: {
			VariantDisplayInfo: boardgame.VariantDisplayInfo{
				//Can skip DisplayName because that will be set automatically correctly
				Description: "Whether or not the target cards left is the default",
			},
			Default: variantTargetCardsLeftDefault,
			Values: map[string]*boardgame.VariantDisplayInfo{
				variantTargetCardsLeftShort: {
					Description:  "A short game that ends when 2 cards are left",
				},
				variantTargetCardsLeftDefault: {
					Description: "A normal-length game that ends when no cards are left",
				},
			},
		},
	}
}

{{end}}
{{if .EnableExampleComputedProperties}}
func (g *gameDelegate) ComputedGlobalProperties(state boardgame.ImmutableState) boardgame.PropertyCollection {
	
	//ComputedProperties are mostly useful when a given state object's
	//computed property is useful clientside, too.

	game := state.ImmutableGameState().(*gameState)

	return boardgame.PropertyCollection{
		"CardsDone": game.CardsDone(),
	}
}

func (g *gameDelegate) ComputedPlayerProperties(player boardgame.ImmutablePlayerState) boardgame.PropertyCollection {

	//ComputedProperties are mostly useful when a given state object's
	//computed property is useful clientside, too.

	p := player.(*playerState)

	return boardgame.PropertyCollection{
		"GameScore": p.GameScore(),
	}
}

{{end}}
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
	{{if .EnableExampleMoves}}
	"errors"
	{{end}}
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
)

//boardgame:codegen
type playerState struct {
	base.SubState
	playerIndex         boardgame.PlayerIndex
	{{if .EnableExampleDeck -}}
	Hand boardgame.Stack ` + "`stack:\"examplecards\" sanitize:\"len\"`" + `
	{{- end}}
	{{if .EnableExampleMoves -}}
	HasDrawnCardThisTurn bool
	{{- end}}
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}

{{if or .EnableExampleEndState .EnableExampleComputedProperties}}
func (p *playerState) GameScore() int {
	//base.GameDelegate's PlayerScore will use the GameScore() method on
	//playerState automatically if it exists.

	{{if .EnableExampleComputedProperties }}
	//This method is exported as a computed property which means this method
	//will be called on created states, including ones that are sanitized.
	//Because Hand, as configured in the struct tag, will be sanitized 'len',
	//sometimes the values we need to sum will be generic placeholder
	//components. However, because newExampleCardDeck used SetGenericValues,
	//we'll always have a *exampleCard, never nil, to cast to.
	{{end}}

	var sum int
	for _, c := range p.Hand.Components() {
		card := c.Values().(*exampleCard)
		sum += card.Value
	}
	return sum
}

{{end}}
{{if .EnableExampleMoves}}
//TurnDone returns true when this player's turn is done. moves.FinishTurn expects it.
func (p *playerState) TurnDone() error {
	if p.HasDrawnCardThisTurn {
		return nil
	}
	return errors.New("Player hasn't drawn their card yet.")
}

//ResetForTurnStart is called by moves.FinishTurn when the player's turn has
//just started,
func (p *playerState) ResetForTurnStart() error {
	p.HasDrawnCardThisTurn = false
	return nil
}

//ResetForTurnEnd is called by moves.FinishTurn when the player's turn has
//just finished,
func (p *playerState) ResetForTurnEnd() error {
	return nil
}

{{end}}
`

const templateContentsGameStateGo = `package {{.Name}}

import (
	"github.com/jkomoros/boardgame"{{if not .SuppressPhase}}
	"github.com/jkomoros/boardgame/enum"{{- end}}
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/base"
)

//boardgame:codegen
type gameState struct {
	base.SubState
	//Use RoundRobinGameStateProperties so roundrobin moves can be used without any changes
	moves.RoundRobinGameStateProperties
	{{if not .SuppressCurrentPlayer -}}
	//base.GameDelegate will automatically return this from CurrentPlayerIndex
	CurrentPlayer boardgame.PlayerIndex
	{{- end}}
	{{if not .SuppressPhase -}}
	//base.GameDelegate will automatically return this from PhaseEnum, CurrentPhase.
	Phase enum.Val ` + "`enum:\"Phase\"`" + `
	{{- end}}
	{{if .EnableExampleDeck -}}
	DrawStack boardgame.Stack ` + "`stack:\"examplecards\" sanitize:\"len\"`" + `
	{{- end}}
	{{if .EnableExampleVariants -}}
	//This is where the example config is stored in BeginSetup. We use it in
	//gameState.CardsDone().
	TargetCardsLeft int
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

{{if not .SuppressPhase }}
func (g *gameState) SetCurrentPhase(phase int) {
	g.Phase.SetValue(phase)
}

{{end}}
{{if not .SuppressCurrentPlayer}}
func (g *gameState) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
	//Having this setter allows us to work moves.With moves.TurnDone
	g.CurrentPlayer = currentPlayer
}

{{end}}
{{if or .EnableExampleEndState .EnableExampleComputedProperties}}
func (g *gameState) CardsDone() bool {
	//It's common to hang computed properties and methods off of gameState and
	//playerState to use in logic elsewhere.

	return g.DrawStack.Len() == {{if .EnableExampleVariants}}g.TargetCardsLeft{{else}}0{{end}}
}

{{end}}

`

const templateContentsDefaultMovesGo = `package {{.Name}}

//TODO: define your move structs here. Don't forget the 'boardgame:codegen'
//magic comment, and don't forget to return them from
//delegate.ConfigureMoves().
{{if not .SuppressPhase}}

//Typically you create a separate file for moves of each major phase, and put
//the moves for that phase in it.
{{- end}}

`

const templateContentsNormalMovesGo = `package {{.Name}}

{{if .EnableExampleMoves}}
import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)
{{end}}

//TODO: define your move structs here. Don't forget the 'boardgame:codegen'
//magic comment, and don't forget to return them from
//delegate.ConfigureMoves().
{{if not .SuppressPhase}}

//Typically you create a separate file for moves of each major phase, and put
//the moves for that phase in it.
{{- end}}

{{if .EnableExampleMoves}}
//boardgame:codegen
type moveDrawCard struct {
	moves.CurrentPlayer
}

func (m *moveDrawCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	//There's important logic up the chain of move types; it's always
	//important to call your parent's Legal. CurrentPlayer will ensure that
	//it's the proposer's turn.
	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game := state.ImmutableGameState().(*gameState)

	if game.DrawStack.Len() < 1 {
		return errors.New("No cards left to draw!")
	}

	return nil

}

func (m *moveDrawCard) Apply(state boardgame.State) error {
	game := state.GameState().(*gameState)
	player := state.CurrentPlayer().(*playerState)

	if err := game.DrawStack.First().MoveToLastSlot(player.Hand); err != nil {
		return err
	}

	player.HasDrawnCardThisTurn = true
	return nil
}

{{end}}
`

const templateContentComponentsGo = `package {{.Name}}

{{if .EnableExampleDeck }}
import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
)

const numCards = 10
const exampleCardDeckName = "examplecards"

//boardgame:codegen
type exampleCard struct {
	base.ComponentValues
	Value int
}

{{if .EnableExampleDynamicComponentValues}}
//boardgame:codegen
type exampleCardDynamicValues struct {
	base.SubState
	base.ComponentValues
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

	//Set the value to return whenever the stack is sanitized. If we didn't
	//set this then sometimes the ComponentValues in a stack would be nil when
	//they are sanitized, which is error-prone for methods. It's always best
	//to set a reasonable generic value so that methods can always assume non-
	//nil ComponentValues.
	deck.SetGenericValues(&exampleCard{
		Value:0,
	})

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
	//that simply testing that we don't get an error moves.With our delegate is a
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

const templateContentsRenderGameJs = `import { BoardgameBaseGameRenderer } from '../../src/boardgame-base-game-renderer.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';
{{- if .EnableExampleClient }}
import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/iron-flex-layout/iron-flex-layout.js';
import '../../src/boardgame-component-stack.js';
import '../../src/boardgame-card.js';
import '../../src/boardgame-deck-defaults.js';
import '../../src/boardgame-fading-text.js';
{{- end}}

class BoardgameRenderGame{{uppercaseFirst .Name}} extends BoardgameBaseGameRenderer {

  static get template() {
  	{{if .EnableExampleClient }}
  	return html[[BACKTICK]]<style>
      #players {
        @apply --layout-horizontal;
        @apply --layout-center;
      }
      .flex {
        @apply --layout-flex;
      }
      .player {
        @apply --layout-vertical;
      }
    </style>
    <boardgame-deck-defaults>
      <template deck="examplecards">
        <boardgame-card rank="{[[item.Values.Value]]}"></boardgame-card>
      </template>
    </boardgame-deck-defaults>
    <boardgame-component-stack stack="{[[state.Game.DrawStack]]}" layout="stack" messy {{if .EnableExampleMoves}} component-propose-move="Draw Card"{{end}}></boardgame-component-stack>
    <div id="players">
      <template is="dom-repeat" items="{[[state.Players]]}">
      	<div class="player flex">
		    <strong>Player {[[index]]}</strong>
		    <boardgame-component-stack stack="{[[item.Hand]]}" layout="fan" messy component-rotated>
		    	<boardgame-fading-text trigger="{[[item.Computed.GameScore]]}" auto-message="diff-up"></boardgame-fading-text>
		    </boardgame-component-stack>
	    </div>
      </template>
    </div>
    <boardgame-fading-text trigger="{[[isCurrentPlayer]]}" message="Your Turn" suppress="falsey"></boardgame-fading-text>
[[BACKTICK]];
{{else}}
return html[[BACKTICK]]This is where you game should render itself. See boardgame/server/README.md for more on the components you can use, or check out the examples in boardgame/examples.[[BACKTICK]];
{{end}}
  }

  static get is() {
    return "boardgame-render-game-{{.Name}}"
  }

  //We don't need to compute any properties that BoardgameBaseGamErenderer
  //doesn't have.

}

customElements.define(BoardgameRenderGame{{uppercaseFirst .Name}}.is, BoardgameRenderGame{{uppercaseFirst .Name}});

`

const templateContentsRenderPlayerInfoJs = `import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';
{{if .EnableExampleClient}}
import '../../src/boardgame-status-text.js';
{{end}}

class BoardgameRenderPlayerInfo{{uppercaseFirst .Name}} extends PolymerElement {

  static get template() {
  	{{if .EnableExampleClient}}
		return html[[BACKTICK]]Number of Cards <boardgame-status-text>{[[playerState.Hand.Indexes.length]]}</boardgame-status-text>[[BACKTICK]];
	{{else}}
		return html[[BACKTICK]]This is where you render info on player, typically using &lt;boardgame-status-text&gt;.[[BACKTICK]];
	{{end}}

  }

  static get is() {
    return "boardgame-render-player-info-{{.Name}}"
  }

  {{if .EnableExampleClient}}
  static get properties() {
    return {
      state: Object,
      playerIndex: Number,
      playerState: Object,
    }
  }
  {{end}}

}

customElements.define(BoardgameRenderPlayerInfo{{uppercaseFirst .Name}}.is, BoardgameRenderPlayerInfo{{uppercaseFirst .Name}});
`
