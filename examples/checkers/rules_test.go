package checkers

import (
	"sort"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// The rules tests pin the fixed game: a checkerboard, two players with
// distinct colors, men moving forward on the diagonals, single jumps that
// capture the jumped token, multi-jump continuation, crowning on the far row
// and crowned tokens moving both ways. Positions are built on a real state
// after set-up and each move goes through its real Legal and Apply.

// The players' colors, as BeginSetUp assigns them.
const (
	redPlayer   boardgame.PlayerIndex = 0
	blackPlayer boardgame.PlayerIndex = 1
)

// space returns the Spaces index of (row, col).
func space(row, col int) int {
	return row*boardWidth + col
}

type placed struct {
	space   int
	color   enum.EnumKey
	crowned bool
}

// buildPosition returns a copy of a fresh game's state after set-up with
// every token taken off the board and the given ones put on it, current
// player current, and every player's FinishedTurn false.
func buildPosition(t *testing.T, tokens []placed, current boardgame.PlayerIndex) (*boardgame.Game, boardgame.State) {
	t.Helper()
	game, base := newMoveMoveTokenGame(t)
	copied, err := base.Copy(false)
	if err != nil {
		t.Fatalf("copying state: %v", err)
	}
	state := copied.(boardgame.State)
	g := state.GameState().(*gameState)
	for i := 0; i < g.Spaces.Len(); i++ {
		if c := g.Spaces.ComponentAt(i); c != nil {
			if err := c.MoveToLastSlot(g.UnusedTokens); err != nil {
				t.Fatalf("clearing space %d: %v", i, err)
			}
		}
	}
	for _, p := range tokens {
		var found boardgame.ComponentInstance
		for _, c := range g.UnusedTokens.Components() {
			if c != nil && c.Values().(*token).Color.Value() == p.color {
				found = c
				break
			}
		}
		if found == nil {
			t.Fatalf("no unused token of color %d left", p.color)
		}
		if err := found.MoveTo(g.Spaces, p.space); err != nil {
			t.Fatalf("placing token at %d: %v", p.space, err)
		}
		found.DynamicValues().(*tokenDynamic).Crowned = p.crowned
	}
	g.CurrentPlayer = current
	for _, p := range state.PlayerStates() {
		p.(*playerState).FinishedTurn = false
	}
	return game, state
}

// legalMoves lists every (token, space) Move Token pair Legal accepts for
// proposer, as "token>space".
func legalMoves(t *testing.T, game *boardgame.Game, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) []string {
	t.Helper()
	var out []string
	for tok := 0; tok < boardSize; tok++ {
		for sp := 0; sp < boardSize; sp++ {
			if tokenMove(t, game, state, tok, sp).Legal(state, proposer) == nil {
				out = append(out, pair(tok, sp))
			}
		}
	}
	sort.Strings(out)
	return out
}

// tokenMove is a Move Token from tok to sp with its defaults (the target
// player) taken from state.
func tokenMove(t *testing.T, game *boardgame.Game, state boardgame.ImmutableState, tok, sp int) *moveMoveToken {
	t.Helper()
	m, ok := game.MoveByNameForState("Move Token", state).(*moveMoveToken)
	if !ok {
		t.Fatal("no Move Token move")
	}
	if err := m.TokenIndexToMove.SetValue(enum.EnumKey(tok)); err != nil {
		t.Fatal(err)
	}
	if err := m.SpaceIndex.SetValue(enum.EnumKey(sp)); err != nil {
		t.Fatal(err)
	}
	return m
}

func pair(tok, sp int) string {
	return spacesEnum.String(enum.EnumKey(tok)) + ">" + spacesEnum.String(enum.EnumKey(sp))
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// moveToken checks Move Token (tok to sp) is legal for the current player and
// applies it to state.
func moveToken(t *testing.T, game *boardgame.Game, state boardgame.State, tok, sp int) {
	t.Helper()
	m := tokenMove(t, game, state, tok, sp)
	if err := m.Legal(state, state.CurrentPlayerIndex()); err != nil {
		t.Fatalf("Move Token %s not legal: %v", pair(tok, sp), err)
	}
	if err := m.Apply(state); err != nil {
		t.Fatalf("Move Token %s: Apply failed after Legal passed: %v", pair(tok, sp), err)
	}
}

func colorAt(state boardgame.ImmutableState, i int) (enum.EnumKey, bool) {
	c := state.ImmutableGameState().(*gameState).Spaces.ImmutableComponentAt(i)
	if c == nil {
		return 0, false
	}
	return c.Values().(*token).Color.Value(), true
}

func TestStartingPosition(t *testing.T) {
	game, state := newMoveMoveTokenGame(t)
	g, players := concreteStates(state)

	if g.Phase.Value() != phasePlaying {
		t.Fatalf("phase is %d after set-up, want playing", g.Phase.Value())
	}

	// A checkerboard: (0,0) is black and colors alternate along rows and
	// columns.
	for row := 0; row < boardWidth; row++ {
		for col := 0; col < boardWidth; col++ {
			want := (row+col)%2 == 0
			if got := spaceIsBlack(space(row, col)); got != want {
				t.Errorf("spaceIsBlack(%d,%d) = %v, want %v", row, col, got, want)
			}
		}
	}

	// Red fills the black spaces of rows 0-2, black those of rows 5-7.
	for row := 0; row < boardWidth; row++ {
		for col := 0; col < boardWidth; col++ {
			color, occupied := colorAt(state, space(row, col))
			var wantOccupied bool
			var wantColor enum.EnumKey
			if spaceIsBlack(space(row, col)) && row <= 2 {
				wantOccupied, wantColor = true, colorRed
			}
			if spaceIsBlack(space(row, col)) && row >= 5 {
				wantOccupied, wantColor = true, colorBlack
			}
			if occupied != wantOccupied || (occupied && color != wantColor) {
				t.Errorf("space (%d,%d): occupied %v color %d, want occupied %v color %d", row, col, occupied, color, wantOccupied, wantColor)
			}
		}
	}

	// Each player has a distinct color and owns 12 tokens of it.
	if len(players) != 2 {
		t.Fatalf("%d players, want 2", len(players))
	}
	if players[redPlayer].Color.Value() != colorRed || players[blackPlayer].Color.Value() != colorBlack {
		t.Errorf("player colors are %s and %s, want Red and Black", players[0].Color, players[1].Color)
	}
	for i, p := range players {
		owned := 0
		for s := 0; s < g.Spaces.Len(); s++ {
			if color, ok := colorAt(state, s); ok && color == p.Color.Value() {
				owned++
			}
		}
		if owned != numTokens {
			t.Errorf("player %d owns %d tokens on the board, want %d", i, owned, numTokens)
		}
	}

	if state.CurrentPlayerIndex() != 0 {
		t.Fatalf("current player is %d, want 0", state.CurrentPlayerIndex())
	}

	// Red's opening moves: each row-2 token one step down a diagonal.
	want := []string{
		pair(space(2, 0), space(3, 1)),
		pair(space(2, 2), space(3, 1)), pair(space(2, 2), space(3, 3)),
		pair(space(2, 4), space(3, 3)), pair(space(2, 4), space(3, 5)),
		pair(space(2, 6), space(3, 5)), pair(space(2, 6), space(3, 7)),
	}
	sort.Strings(want)
	if got := legalMoves(t, game, state, 0); !sameStrings(got, want) {
		t.Errorf("player 0's legal opening moves are %v, want %v", got, want)
	}
	if got := legalMoves(t, game, state, 1); len(got) != 0 {
		t.Errorf("player 1 has legal moves %v when it is not their turn", got)
	}
}

// TestSimpleMove plays a step through the real game: the token moves, nothing
// is captured, Finish Turn ends the turn and black may then answer.
func TestSimpleMove(t *testing.T) {
	game, _ := newMoveMoveTokenGame(t)
	before := game.Version()

	from, to := space(2, 2), space(3, 3)
	if err := <-game.ProposeMove(moveMoveTokenMove(t, game, from, to), 0); err != nil {
		t.Fatalf("proposing %s: %v", pair(from, to), err)
	}
	state := game.CurrentState()
	if _, occupied := colorAt(state, from); occupied {
		t.Errorf("%s still occupied after the move", pair(from, to))
	}
	if color, occupied := colorAt(state, to); !occupied || color != colorRed {
		t.Errorf("destination holds %d (occupied %v), want red", color, occupied)
	}
	_, players := concreteStates(state)
	for i, p := range players {
		if p.CapturedTokens.NumComponents() != 0 {
			t.Errorf("player %d captured %d tokens on a step", i, p.CapturedTokens.NumComponents())
		}
	}
	if state.CurrentPlayerIndex() != 1 {
		t.Errorf("current player is %d after the step, want 1", state.CurrentPlayerIndex())
	}
	// The step, then Finish Turn.
	if game.Version() != before+2 {
		t.Errorf("version is %d after the step, want %d", game.Version(), before+2)
	}

	black, blackTo := space(5, 1), space(4, 2)
	if err := <-game.ProposeMove(moveMoveTokenMove(t, game, black, blackTo), 1); err != nil {
		t.Fatalf("proposing black's answer %s: %v", pair(black, blackTo), err)
	}
	if color, occupied := colorAt(game.CurrentState(), blackTo); !occupied || color != colorBlack {
		t.Errorf("black's destination holds %d (occupied %v), want black", color, occupied)
	}
}

func TestCapture(t *testing.T) {
	game, state := buildPosition(t, []placed{
		{space(5, 1), colorBlack, false},
		{space(4, 2), colorRed, false},
		{space(1, 1), colorRed, false},
	}, blackPlayer)

	moveToken(t, game, state, space(5, 1), space(3, 3))

	if _, occupied := colorAt(state, space(4, 2)); occupied {
		t.Errorf("the jumped space is still occupied")
	}
	if color, occupied := colorAt(state, space(3, 3)); !occupied || color != colorBlack {
		t.Errorf("landing space holds %d (occupied %v), want black", color, occupied)
	}
	if color, occupied := colorAt(state, space(1, 1)); !occupied || color != colorRed {
		t.Errorf("an unrelated red token was disturbed")
	}
	_, players := concreteStates(state)
	captured := players[blackPlayer].CapturedTokens
	if captured.NumComponents() != 1 {
		t.Fatalf("black captured %d tokens, want 1", captured.NumComponents())
	}
	if captured.ImmutableComponentAt(0).Values().(*token).Color.Value() != colorRed {
		t.Errorf("black captured a token that is not red")
	}
	if players[redPlayer].CapturedTokens.NumComponents() != 0 {
		t.Errorf("red captured a token on black's jump")
	}
	if !players[blackPlayer].FinishedTurn {
		t.Errorf("the turn is not finished after a jump with no follow-up capture")
	}
	if game.Manager().Delegate().(*gameDelegate).GameEndConditionMet(state) {
		t.Errorf("game over after one capture")
	}
}

// TestJumpOverOwnToken: a token cannot jump its own color.
func TestJumpOverOwnToken(t *testing.T) {
	game, state := buildPosition(t, []placed{
		{space(5, 1), colorBlack, false},
		{space(4, 2), colorBlack, false},
	}, blackPlayer)
	if err := tokenMove(t, game, state, space(5, 1), space(3, 3)).Legal(state, blackPlayer); err == nil {
		t.Errorf("jumping one's own token is legal")
	}
}

func TestMultiJump(t *testing.T) {
	game, state := buildPosition(t, []placed{
		{space(5, 1), colorBlack, false},
		{space(4, 2), colorRed, false},
		{space(2, 4), colorRed, false},
	}, blackPlayer)

	moveToken(t, game, state, space(5, 1), space(3, 3))
	_, players := concreteStates(state)
	if players[blackPlayer].FinishedTurn {
		t.Fatalf("the turn finished with a second jump available")
	}
	if err := players[blackPlayer].TurnDone(); err == nil {
		t.Errorf("TurnDone is nil mid-jump")
	}

	moveToken(t, game, state, space(3, 3), space(1, 5))
	if !players[blackPlayer].FinishedTurn {
		t.Errorf("the turn did not finish after the last jump")
	}
	if players[blackPlayer].CapturedTokens.NumComponents() != 2 {
		t.Errorf("black captured %d tokens, want 2", players[blackPlayer].CapturedTokens.NumComponents())
	}
	for _, s := range []int{space(4, 2), space(2, 4), space(3, 3), space(5, 1)} {
		if _, occupied := colorAt(state, s); occupied {
			t.Errorf("space %s still occupied", spacesEnum.String(enum.EnumKey(s)))
		}
	}
}

func crownMove(t *testing.T, game *boardgame.Game, state boardgame.ImmutableState) *moveCrownToken {
	t.Helper()
	m, ok := game.MoveByNameForState("Crown Token", state).(*moveCrownToken)
	if !ok {
		t.Fatal("no Crown Token move")
	}
	return m
}

func TestCrowningAndBackwardMoves(t *testing.T) {
	game, state := buildPosition(t, []placed{
		{space(1, 1), colorBlack, false},
		{space(6, 6), colorRed, false},
	}, blackPlayer)

	// Nothing is crownable yet.
	if err := crownMove(t, game, state).Legal(state, boardgame.AdminPlayerIndex); err == nil {
		t.Fatalf("Crown Token legal before any token reached the far row")
	}

	moveToken(t, game, state, space(1, 1), space(0, 2))

	crown := crownMove(t, game, state)
	if crown.ComponentIndex != space(0, 2) {
		t.Errorf("Crown Token defaults to %d, want %d", crown.ComponentIndex, space(0, 2))
	}
	if err := crown.Legal(state, boardgame.AdminPlayerIndex); err != nil {
		t.Fatalf("Crown Token not legal for a black token on row 0: %v", err)
	}
	if err := crown.Apply(state); err != nil {
		t.Fatalf("Crown Token Apply: %v", err)
	}
	g := state.GameState().(*gameState)
	if !g.Spaces.ComponentAt(space(0, 2)).DynamicValues().(*tokenDynamic).Crowned {
		t.Fatalf("token not crowned")
	}
	if err := crownMove(t, game, state).Legal(state, boardgame.AdminPlayerIndex); err == nil {
		t.Errorf("Crown Token still legal after crowning")
	}

	// The crowned black token now moves down the board too.
	state.GameState().(*gameState).CurrentPlayer = blackPlayer
	state.PlayerStates()[blackPlayer].(*playerState).FinishedTurn = false
	want := []string{pair(space(0, 2), space(1, 1)), pair(space(0, 2), space(1, 3))}
	sort.Strings(want)
	if got := legalMoves(t, game, state, blackPlayer); !sameStrings(got, want) {
		t.Errorf("crowned token's moves are %v, want %v", got, want)
	}
	moveToken(t, game, state, space(0, 2), space(1, 3))
}

// TestCrownedBothWays: an uncrowned token only moves forward; a crowned one
// steps and jumps backward as well.
func TestCrownedBothWays(t *testing.T) {
	positions := []bool{false, true}
	for _, crowned := range positions {
		game, state := buildPosition(t, []placed{
			{space(3, 3), colorRed, crowned},
			{space(2, 4), colorBlack, false},
		}, redPlayer)
		got := legalMoves(t, game, state, redPlayer)
		want := []string{pair(space(3, 3), space(4, 2)), pair(space(3, 3), space(4, 4))}
		if crowned {
			want = append(want, pair(space(3, 3), space(2, 2)), pair(space(3, 3), space(1, 5)))
		}
		sort.Strings(want)
		if !sameStrings(got, want) {
			t.Errorf("crowned %v: red's moves are %v, want %v", crowned, got, want)
		}
		if crowned {
			moveToken(t, game, state, space(3, 3), space(1, 5))
			_, players := concreteStates(state)
			if players[redPlayer].CapturedTokens.NumComponents() != 1 {
				t.Errorf("backward jump captured %d tokens, want 1", players[redPlayer].CapturedTokens.NumComponents())
			}
		}
	}
}

func TestGameEnds(t *testing.T) {
	game, state := buildPosition(t, []placed{
		{space(5, 1), colorBlack, false},
		{space(4, 2), colorRed, false},
	}, blackPlayer)
	g := state.GameState().(*gameState)
	_, players := concreteStates(state)
	for i := 0; i < numTokens-1; i++ {
		var red boardgame.ComponentInstance
		for _, c := range g.UnusedTokens.Components() {
			if c != nil && c.Values().(*token).Color.Value() == colorRed {
				red = c
				break
			}
		}
		if err := red.MoveToLastSlot(players[blackPlayer].CapturedTokens); err != nil {
			t.Fatalf("pre-capturing: %v", err)
		}
	}
	delegate := game.Manager().Delegate().(*gameDelegate)
	if delegate.GameEndConditionMet(state) {
		t.Fatalf("game over with 11 captured")
	}
	moveToken(t, game, state, space(5, 1), space(3, 3))
	if !delegate.GameEndConditionMet(state) {
		t.Errorf("game not over after the twelfth capture")
	}
	if delegate.PlayerScore(players[blackPlayer]) != numTokens {
		t.Errorf("black's score is %d, want %d", delegate.PlayerScore(players[blackPlayer]), numTokens)
	}
}

// TestLegalMovesApply applies every legal Move Token from a set of positions,
// each on a fresh copy: none may fail in Apply.
func TestLegalMovesApply(t *testing.T) {
	positions := [][]placed{
		nil, // filled below with the real starting position
		{
			{space(3, 3), colorRed, true}, {space(2, 2), colorBlack, false}, {space(2, 4), colorBlack, true},
			{space(4, 2), colorBlack, false}, {space(4, 4), colorBlack, false}, {space(6, 0), colorRed, false},
			{space(7, 1), colorBlack, false}, {space(0, 6), colorBlack, false},
		},
	}
	for i, tokens := range positions {
		var game *boardgame.Game
		var state boardgame.State
		if tokens == nil {
			game, state = newMoveMoveTokenGame(t)
		} else {
			game, state = buildPosition(t, tokens, 1)
		}
		for _, proposer := range []boardgame.PlayerIndex{0, 1} {
			state.GameState().(*gameState).CurrentPlayer = proposer
			for tok := 0; tok < boardSize; tok++ {
				for sp := 0; sp < boardSize; sp++ {
					m := tokenMove(t, game, state, tok, sp)
					if m.Legal(state, proposer) != nil {
						continue
					}
					copied, err := state.Copy(false)
					if err != nil {
						t.Fatal(err)
					}
					if err := m.Apply(copied.(boardgame.State)); err != nil {
						t.Errorf("position %d, player %d, %s: Apply failed after Legal passed: %v", i, proposer, pair(tok, sp), err)
					}
				}
			}
		}
	}
}
