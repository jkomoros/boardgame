// Package vecgame is a game definition that exists only to make real Go emit
// known-answer vectors for the TypeScript engine. It has one of every kind of
// property the engine has to reproduce: growable, capped and sized stacks,
// both kinds of merged stack, a board, dynamic component values, enums,
// slices and a timer, plus player groups for the computed sanitization
// groups.
//
// It is never registered with a server. The generators under ../statevec
// drive it.
package vecgame

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

var GroupEnum = Enums.MustCombine("group", ColorEnum)

// NumCards and NumTokens are the deck sizes.
const (
	NumCards  = 24
	NumTokens = 6
)

//boardgame:codegen reader
type cardValue struct {
	base.ComponentValues
	Rank int
	Suit string
}

//boardgame:codegen reader
type tokenValue struct {
	base.ComponentValues
	Kind string
}

//boardgame:codegen
type cardDynamic struct {
	base.SubState
	base.ComponentValues
	Counter int
	Held    boardgame.SizedStack `sizedstack:"tokens,1"`
	// Nest holds other cards, so visibility of dynamic values has to be
	// followed transitively.
	Nest boardgame.Stack `stack:"cards"`
	Mark enum.Val        `enum:"color"`
}

//boardgame:codegen
type gameState struct {
	base.SubState
	Draw    boardgame.Stack       `stack:"cards"`
	Limited boardgame.Stack       `stack:"cards,3"`
	Row     boardgame.SizedStack  `sizedstack:"cards,4"`
	Under   boardgame.SizedStack  `sizedstack:"cards,4"`
	Overlap boardgame.MergedStack `overlap:"Row,Under"`
	All     boardgame.MergedStack `concatenate:"Draw,Limited"`
	Tokens  boardgame.Stack       `stack:"tokens"`
	Spaces  boardgame.Board       `stack:"tokens" board:"3"`
	Count   int
	Name    string
	Flag    bool
	Who     boardgame.PlayerIndex
	Ints    []int
	Bools   []bool
	Strings []string
	Whos    []boardgame.PlayerIndex
	Color   enum.Val       `enum:"color"`
	Colors  enum.EnumSlice `enum:"color"`
	Clock   boardgame.Timer
	// The remaining properties carry real struct tags, so the vectors also
	// cover Go's own tag parsing and default groups.
	TagHidden   int             `sanitize:"hidden"`
	TagLen      boardgame.Stack `stack:"cards" sanitize:"len"`
	TagOrder    boardgame.Stack `stack:"cards" sanitize:"order"`
	TagNonEmpty boardgame.Stack `stack:"cards" sanitize:"nonempty"`
	TagRedOnly  string          `sanitize:"all:hidden,Red:visible"`
	TagSpaced   int             `sanitize:" Hidden "`
}

//boardgame:codegen
type playerState struct {
	base.SubState
	Hand  boardgame.Stack      `stack:"cards" sanitize:"len"`
	Slots boardgame.SizedStack `sizedstack:"cards,2" sanitize:"self:visible,other:order"`
	Color enum.Val             `enum:"color"`
	Score int
	// Secret omits the group name, which means "other" on a player state.
	Secret     int    `sanitize:"hidden"`
	TeamSecret string `sanitize:"all:hidden,same-color:visible"`
	RivalNote  string `sanitize:"all:hidden,different-color:visible"`
	OwnOnly    []int  `sanitize:"all:nonempty,self:visible"`
	Dyn        int
}

// PolicyOverride, when non-nil, replaces the struct-tag policy for one
// property. It is how the generator exercises every policy on every property
// type without a struct per combination. The resolution that follows
// (least-restrictive-wins over the viewer's groups) is still Go's own.
type PolicyOverride func(ref boardgame.StatePropertyRef) (map[string]boardgame.Policy, bool)

type Delegate struct {
	base.GameDelegate
	Override PolicyOverride
}

func (g *Delegate) Name() string        { return "vecgame" }
func (g *Delegate) DisplayName() string { return "Vector Game" }
func (g *Delegate) MinNumPlayers() int  { return 1 }
func (g *Delegate) MaxNumPlayers() int  { return 6 }
func (g *Delegate) DefaultNumPlayers() int {
	return 3
}

func (g *Delegate) ConfigureEnums() *enum.Set { return Enums }

func (g *Delegate) ConfigureDecks() map[string]*boardgame.Deck {
	cards := boardgame.NewDeck()
	suits := []string{"clubs", "hearts", "spades"}
	for i := 0; i < NumCards; i++ {
		cards.AddComponent(&cardValue{Rank: i%4 + 1, Suit: suits[i%3]})
	}
	tokens := boardgame.NewDeck()
	for i := 0; i < NumTokens; i++ {
		tokens.AddComponent(&tokenValue{Kind: "token"})
	}
	return map[string]*boardgame.Deck{"cards": cards, "tokens": tokens}
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

func (g *Delegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game := state.ImmutableGameState().(*gameState)
	switch c.Deck().Name() {
	case "cards":
		return game.Draw, nil
	case "tokens":
		return game.Tokens, nil
	}
	return nil, errors.New("unknown deck")
}

// moveNever exists because a manager refuses to boot with no moves. It is
// never legal, so it never applies.
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

func (g *Delegate) SanitizationPolicy(ref boardgame.StatePropertyRef, groupMembership map[string]bool) boardgame.Policy {
	if g.Override != nil {
		if policyMap, ok := g.Override(ref); ok {
			return boardgame.ResolveSanitizationPolicy(policyMap, groupMembership, boardgame.PolicyVisible)
		}
	}
	return g.GameDelegate.SanitizationPolicy(ref, groupMembership)
}

// Game returns the concrete game state so a generator can reach its stacks.
func Game(state boardgame.ImmutableState) *gameState {
	return state.ImmutableGameState().(*gameState)
}

// Players returns the concrete player states.
func Players(state boardgame.ImmutableState) []*playerState {
	in := state.ImmutablePlayerStates()
	out := make([]*playerState, len(in))
	for i, p := range in {
		out[i] = p.(*playerState)
	}
	return out
}

// CardDynamic returns the concrete dynamic values for card i.
func CardDynamic(state boardgame.ImmutableState, i int) *cardDynamic {
	return state.ImmutableDynamicComponentValues()["cards"][i].(*cardDynamic)
}
