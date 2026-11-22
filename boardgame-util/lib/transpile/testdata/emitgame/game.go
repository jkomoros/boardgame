// Package emitgame is the translator's golden fixture: a small game whose
// functions each exercise one rule of the emitter (types, statements,
// expressions, the API table, the semantics table, the role signatures).
// Every function is reachable from a slot, so the static half counts it.
// It is type-checked, never run.
package emitgame

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

type gameDelegate struct {
	base.GameDelegate
}

// NewDelegate returns the fixture's delegate.
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}

const (
	phaseSetUp = iota
	phasePlay
)

const dim = 3

const waitDuration = 2 * time.Second

//boardgame:codegen
type gameState struct {
	base.SubState
	behaviors.CurrentPlayerBehavior
	behaviors.PhaseBehavior
	Cards    boardgame.Stack      `stack:"cards"`
	Discards boardgame.Stack      `stack:"cards"`
	Board    boardgame.SizedStack `sizedstack:"cards,4"`
	Timer    boardgame.Timer
	Count    int
	Label    string
	Scores   []int
}

//boardgame:codegen
type playerState struct {
	base.SubState
	behaviors.Seat
	behaviors.InactivePlayer
	behaviors.PlayerElimination
	Hand  boardgame.Stack `stack:"cards"`
	Score int
	Done  bool
	Votes []boardgame.PlayerIndex
}

//boardgame:codegen
type cardValue struct {
	base.ComponentValues
	Value int
	Name  string
}

//boardgame:codegen
type moveAll struct {
	moves.CurrentPlayer
	Amount int
}

//boardgame:codegen
type moveFinish struct {
	moves.FinishTurn
}

// Apply: the stock Apply is stockApply.
func (m *moveFinish) Apply(state boardgame.State) error {
	if err := m.FinishTurn.Apply(state); err != nil {
		return err
	}
	return nil
}

func (g *gameDelegate) Name() string {
	return "emitgame"
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return &playerState{}
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := moves.NewAutoConfigurer(g)
	return moves.Add(auto.MustConfig(new(moveAll)), auto.MustConfig(new(moveFinish)))
}

func (g *gameDelegate) ConfigureComputedProperties() []boardgame.ComputedProperty {
	return []boardgame.ComputedProperty{
		boardgame.PlayerComputedInt("Doubled", func(player boardgame.ImmutableSubState) int {
			return player.(*playerState).Score * 2
		}),
	}
}

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)
	players := make([]*playerState, len(state.ImmutablePlayerStates()))
	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}
	return game, players
}

// ----- roles ----------------------------------------------------------------

// Distribute: a Component parameter is { deck, index }; the error result
// throws.
func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game, _ := concreteStates(state)
	if c.Deck().Name() != "cards" {
		return nil, errors.New("unknown deck")
	}
	if c.Values().(*cardValue).Value > 10 {
		return game.Discards, nil
	}
	return game.Cards, nil
}

// BeginSetUp: the variant parameter.
func (g *gameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
	game, _ := concreteStates(state)
	game.Label = variant["label"]
	return nil
}

// FinishSetUp reaches every helper below.
func (g *gameDelegate) FinishSetUp(state boardgame.State) error {
	game, players := concreteStates(state)
	game.Count = folded() + arithmetic(7, 2) + len(strs("a", "b")) + sliceOps(3) + mapOps()
	game.Count += tuples() + control(4) + ranges(state) + named(2) + extremes(1, 2)
	if err := errorsAndFormat(state, players[0]); err != nil {
		return err
	}
	stacksAndComponents(state)
	enumsAndTimers(state)
	behaviorsAndPlayers(state)
	nullableStack(state, true)
	conversions(state)
	return nil
}

// CheckGameFinished: named results, a tuple.
func (g *gameDelegate) CheckGameFinished(state boardgame.ImmutableState) (finished bool, winners []boardgame.PlayerIndex) {
	_, players := concreteStates(state)
	for i, p := range players {
		if p.Score > 100 {
			winners = append(winners, boardgame.PlayerIndex(i))
			finished = true
		}
	}
	return
}

func (p *playerState) TurnDone() error {
	if !p.Done {
		return errors.New("not done")
	}
	return nil
}

func (p *playerState) ResetForTurnStart() error {
	p.Done = false
	return nil
}

func (p *playerState) ResetForTurnEnd() error {
	return nil
}

// DefaultsForState: the stock call first is dropped.
func (m *moveAll) DefaultsForState(state boardgame.ImmutableState) {
	m.CurrentPlayer.DefaultsForState(state)
	m.Amount = 1
}

// Legal: the stock Legal is base().
func (m *moveAll) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}
	if m.Amount < 0 {
		return errors.New("negative")
	}
	return nil
}

// Apply: the delegate through the manager is a direct call.
func (m *moveAll) Apply(state boardgame.State) error {
	if finished, _ := state.Manager().Delegate().CheckGameFinished(state); finished {
		return nil
	}
	game, players := concreteStates(state)
	players[m.TargetPlayerIndex].Score += m.Amount
	game.Count++
	return nil
}

// ----- helpers, one rule each -------------------------------------------------

// folded: constants (iota, dim*dim) are literals.
func folded() int {
	return dim*dim + phasePlay
}

// arithmetic: idiv, imod, imul, compound assignment.
func arithmetic(a, b int) int {
	x := a / b
	x += a % b
	x *= b
	x -= 1
	x++
	return x
}

// strs: concatenation, ordered comparison by code point, byte length.
func strs(a, b string) []string {
	s := a + b
	if a < b && len(s) > 1 {
		return []string{s, strings.ToLower(s)}
	}
	return nil
}

// sliceOps: nil-aware slices, append, make, index, copy.
func sliceOps(n int) int {
	var xs []int
	xs = append(xs, n, n+1)
	ys := make([]int, n)
	copy(ys, xs)
	ys[0] = xs[1]
	zs := []int{}
	return len(xs) + len(ys) + len(zs) + ys[0]
}

// mapOps: make, read with the zero value, write, ++, comma-ok, delete.
func mapOps() int {
	counts := make(map[string]int)
	counts["a"]++
	counts["b"] = 2
	v, ok := counts["c"]
	if !ok {
		delete(counts, "b")
	}
	return counts["a"] + v
}

func pair(x int) (int, string) {
	return x, strconv.Itoa(x)
}

// tuples: multiple results, := with a new and an old name.
func tuples() int {
	a, s := pair(1)
	b, s := pair(a + 1)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return a + b + n
}

// control: if with an initializer, else-if, the three for forms, labels,
// switch with a tag, several values and default, and without a tag.
func control(n int) int {
	total := 0
	if x := n * 2; x > 6 {
		total = x
	} else if x > 2 {
		total = 1
	} else {
		total = 2
	}
	for i := 0; i < n; i++ {
		total += i
	}
	for total > 100 {
		total -= 10
	}
	for {
		if total > 0 {
			break
		}
		total++
	}
outer:
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if j > i {
				continue outer
			}
		}
	}
	switch n {
	case 1, 2:
		total++
	case 3:
		return total
	default:
		total--
	}
	switch {
	case total > 10:
		total = 10
	}
	return total
}

// ranges: a slice with key and value, an integer, a component snapshot,
// the player states.
func ranges(state boardgame.State) int {
	game, _ := concreteStates(state)
	total := 0
	for i, s := range game.Scores {
		total += i + s
	}
	for i := range 3 {
		total += i
	}
	for i, c := range game.Cards.Components() {
		if c != nil {
			total += i
		}
	}
	for _, p := range state.PlayerStates() {
		if behaviors.PlayerIsInactive(p) {
			total++
		}
	}
	return total
}

// named: named results and a bare return.
func named(n int) (doubled int) {
	doubled = n * 2
	return
}

// extremes: min and max.
func extremes(a, b int) int {
	return min(a, b) + max(a, b)
}

// errorsAndFormat: errors.New, fmt.Errorf, err.Error(), the discarded error
// of a stack operation.
func errorsAndFormat(state boardgame.State, p *playerState) error {
	game, _ := concreteStates(state)
	if err := game.Cards.MoveAllTo(p.Hand); err != nil {
		return errors.New("could not move: " + err.Error())
	}
	game.Discards.Shuffle()
	if p.Score > 3 {
		return fmt.Errorf("score %d is too high for %s", p.Score, game.Label)
	}
	return nil
}

// stacksAndComponents: stack queries, component handles, values, moves.
func stacksAndComponents(state boardgame.State) {
	game, players := concreteStates(state)
	first := game.Cards.First()
	if first == nil {
		return
	}
	if first.Values().(*cardValue).Name == "joker" {
		first.MoveTo(game.Board, game.Board.SlotsRemaining()-1)
	}
	for i := 0; i < game.Discards.Len(); i++ {
		c := game.Discards.ComponentAt(i)
		if c.MayMoveTo(players[0].Hand) == nil {
			c.MoveToNextSlot(players[0].Hand)
		}
	}
}

// enumsAndTimers: the phase enum, a timer started with a whole number of
// milliseconds.
func enumsAndTimers(state boardgame.State) {
	game, _ := concreteStates(state)
	if game.Phase.Value() == phaseSetUp {
		game.SetCurrentPhase(phasePlay)
	}
	if !game.Timer.Active() {
		game.Timer.Start(waitDuration, state.Game().MoveByName("Move All"))
	} else {
		game.Timer.Cancel()
	}
}

// behaviorsAndPlayers: behavior setters, player indexes, bounds-checked
// player access, random draws.
func behaviorsAndPlayers(state boardgame.State) {
	game, players := concreteStates(state)
	next := game.CurrentPlayer.Next(state)
	players[next].SetEliminated()
	pick := state.Rand().Intn(len(players))
	players[pick].Votes = append(players[pick].Votes, next)
	players[pick].Votes = nil
}

// nullableStack: a stack variable that may be nil.
func nullableStack(state boardgame.State, fromCards bool) {
	game, _ := concreteStates(state)
	var from boardgame.Stack
	if fromCards {
		from = game.Cards
	} else {
		from = game.Discards
	}
	from.First().MoveToFirstSlot(game.Discards)
}

// conversions: among int, PlayerIndex and EnumKey.
func conversions(state boardgame.State) boardgame.PlayerIndex {
	game, _ := concreteStates(state)
	return boardgame.PlayerIndex(int(game.CurrentPlayer) + 1)
}
