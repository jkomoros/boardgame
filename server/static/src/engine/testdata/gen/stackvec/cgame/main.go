// Package cgame is a game definition that exists only to make real Go emit
// known answers about stack constraints for the TypeScript engine. Every
// constraint kind the default constructors define (max, unique, same,
// maxdistinct) sits on at least one stack, over each kind of property path
// ("Prop", "component.Prop", "dynamic.Prop", a case-folded prefix, a value
// that is an enum, an int, a string, a bool and an int slice, and a deck whose
// components lack the property), on growable stacks, sized stacks, board
// spaces, player stacks and dynamic-value stacks, plus one stack with several
// constraints so the order in which they are checked shows.
//
// It is separate from ../../vecgame so that the vectors vecgame already
// produces stay byte for byte. The generator in .. drives it.
package cgame

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

var Enums = enum.NewSet()

const (
	ColorRed = iota
	ColorGreen
	ColorBlue
)

var ColorEnum = Enums.MustAdd("color", map[enum.EnumKey]string{
	ColorRed:   "Red",
	ColorGreen: "Green",
	ColorBlue:  "Blue",
})

// NumCards and NumChips are the deck sizes.
const (
	NumCards = 18
	NumChips = 6
)

//boardgame:codegen reader
type cardValue struct {
	base.ComponentValues
	Color enum.ImmutableVal `enum:"color"`
	Rank  int
	Suit  string
	Face  bool
	Pips  []int
}

//boardgame:codegen reader
type chipValue struct {
	base.ComponentValues
	Value int
}

//boardgame:codegen
type cardDynamic struct {
	base.SubState
	base.ComponentValues
	Mark int
	Tag  string
	// Nest holds other cards; a constraint on a dynamic-value stack.
	Nest boardgame.Stack `stack:"cards,max(1)"`
}

//boardgame:codegen
type gameState struct {
	base.SubState
	Draw       boardgame.Stack       `stack:"cards"`
	Pool       boardgame.Stack       `stack:"cards"`
	Both       boardgame.MergedStack `concatenate:"Draw,Pool"`
	MaxTwo     boardgame.Stack       `stack:"cards,max(2)"`
	CapSized   boardgame.SizedStack  `sizedstack:"cards,4,max(2)"`
	SameColor  boardgame.Stack       `stack:"cards,same(Color)"`
	UniqueRank boardgame.SizedStack  `sizedstack:"cards,5,unique(Rank)"`
	TwoSuits   boardgame.Stack       `stack:"cards, maxdistinct(Suit, 2)"`
	SameMark   boardgame.Stack       `stack:"cards,same(dynamic.Mark)"`
	FallMark   boardgame.Stack       `stack:"cards,unique(Mark)"`
	SamePips   boardgame.Stack       `stack:"cards,same(Pips)"`
	Combo      boardgame.Stack       `stack:"cards,max(3),unique(component.Color),same(Component.Face)"`
	Spaces     boardgame.Board       `stack:"cards,maxdistinct(Color,1)" board:"3"`
	Chips      boardgame.Stack       `stack:"chips"`
	ChipsColor boardgame.Stack       `stack:"chips,same(Color)"`
	ChipsValue boardgame.Stack       `stack:"chips,maxdistinct(Value,1)"`
}

//boardgame:codegen
type playerState struct {
	base.SubState
	Hand  boardgame.Stack      `stack:"cards,same(Color)"`
	Slots boardgame.SizedStack `sizedstack:"cards,2,unique(Suit)"`
}

type Delegate struct {
	base.GameDelegate
}

func (g *Delegate) Name() string        { return "cgame" }
func (g *Delegate) DisplayName() string { return "Constraint Vector Game" }
func (g *Delegate) MinNumPlayers() int  { return 1 }
func (g *Delegate) MaxNumPlayers() int  { return 4 }
func (g *Delegate) DefaultNumPlayers() int {
	return 2
}

func (g *Delegate) ConfigureEnums() *enum.Set { return Enums }

func (g *Delegate) ConfigureDecks() map[string]*boardgame.Deck {
	cards := boardgame.NewDeck()
	suits := []string{"clubs", "hearts", "spades", "stars"}
	for i := 0; i < NumCards; i++ {
		cards.AddComponent(&cardValue{
			Color: ColorEnum.MustNewImmutableVal(enum.EnumKey(i % 3)),
			Rank:  i%5 + 1,
			Suit:  suits[i%4],
			Face:  i%2 == 0,
			Pips:  []int{i % 2, i % 3},
		})
	}
	chips := boardgame.NewDeck()
	for i := 0; i < NumChips; i++ {
		chips.AddComponent(&chipValue{Value: i % 2})
	}
	return map[string]*boardgame.Deck{"cards": cards, "chips": chips}
}

func (g *Delegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *Delegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(playerState)
}

func (g *Delegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.ConfigurableSubState {
	if deck.Name() == "cards" {
		return new(cardDynamic)
	}
	return nil
}

// DistributeComponentToStarterStack puts cards 0 and 1 (Red and Green) into
// SameColor, which its constraint would refuse: Go does not check
// constraints during set-up. A few more start in constrained stacks so that
// refusals are common; the rest go to Draw, and every chip to Chips.
func (g *Delegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game := state.ImmutableGameState().(*gameState)
	switch c.Deck().Name() {
	case "cards":
		switch c.DeckIndex() {
		case 0, 1:
			return game.SameColor, nil
		case 2, 3:
			return game.UniqueRank, nil
		case 4:
			return game.FallMark, nil
		case 5, 6:
			return game.TwoSuits, nil
		case 7:
			return game.MaxTwo, nil
		}
		return game.Draw, nil
	case "chips":
		return game.Chips, nil
	}
	return nil, errors.New("unknown deck")
}

// FinishSetUp gives each card's dynamic values something to constrain on.
func (g *Delegate) FinishSetUp(state boardgame.State) error {
	for i, values := range state.DynamicComponentValues()["cards"] {
		d := values.(*cardDynamic)
		d.Mark = i % 3
		d.Tag = []string{"a", "b"}[i%2]
	}
	return nil
}

// moveNever exists because a manager refuses to boot with no moves.
//
//boardgame:codegen
type moveNever struct {
	moves.Default
}

func (m *moveNever) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	return errors.New("never legal")
}

func (m *moveNever) Apply(state boardgame.State) error {
	return errors.New("never applies")
}

func (g *Delegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := moves.NewAutoConfigurer(g)
	return moves.Add(auto.MustConfig(new(moveNever), moves.WithMoveName("Never")))
}

// Each badN struct carries one struct tag Go refuses. The generator feeds
// them to boardgame.NewStructInflater to record Go's error text for each.

//boardgame:codegen
type bad1 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,max()"`
}

//boardgame:codegen
type bad2 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,max(1,2)"`
}

//boardgame:codegen
type bad3 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,max(two)"`
}

//boardgame:codegen
type bad4 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,same(Colour)"`
}

//boardgame:codegen
type bad5 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,unique()"`
}

//boardgame:codegen
type bad6 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,maxdistinct(Color)"`
}

//boardgame:codegen
type bad7 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,maxdistinct(Color,x)"`
}

//boardgame:codegen
type bad8 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,reject()"`
}

//boardgame:codegen
type bad9 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,same(dynamic.Rank)"`
}

//boardgame:codegen
type bad10 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,same(component.Mark)"`
}

//boardgame:codegen
type bad11 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,unique(Rank,Suit)"`
}

//boardgame:codegen
type bad12 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,max(2.5)"`
}

//boardgame:codegen
type bad13 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,maxdistinct(Nope,1)"`
}

//boardgame:codegen
type bad14 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,max(99999999999999999999)"`
}

//boardgame:codegen
type bad15 struct {
	base.SubState
	S boardgame.Stack `stack:"cards,max(1),bogus"`
}

// BadStates returns one fresh instance of each bad struct, in order.
func BadStates() []boardgame.ConfigurableSubState {
	return []boardgame.ConfigurableSubState{
		new(bad1), new(bad2), new(bad3), new(bad4), new(bad5), new(bad6), new(bad7),
		new(bad8), new(bad9), new(bad10), new(bad11), new(bad12), new(bad13), new(bad14),
		new(bad15),
	}
}
