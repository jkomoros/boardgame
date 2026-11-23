package moves

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
	"github.com/workfit/tester/assert"
)

/*
Golden-equivalence fence for "player.X" as the move's player
(docs/superpowers/specs/2026-11-22-legality-actor-paths-design.md).

moveActorGolden is the shape every simultaneous-phase migration has: a
moves.CurrentPlayer move whose rules read the acting player's state. Its
declarative plan spells that player "player.X" — the same spelling a
turn-based move uses — and legacyActorGoldenLegal is the imperative oracle
such a game wrote before migrating: CurrentPlayer's checks, then
players[m.TargetPlayerIndex]. The fence sweeps turn-based and simultaneous
(AnyPlayerIndex) current players, every target, and every proposer including
AdminPlayerIndex acting on a player's behalf, and requires identical
legality and the same failing check.
*/

//boardgame:codegen
type moveActorGolden struct {
	CurrentPlayer
}

func (m *moveActorGolden) Apply(state boardgame.State) error { return nil }

func legacyActorGoldenLegal(m *moveActorGolden, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	target := m.TargetPlayerIndex.EnsureValid(state)
	if !target.Valid(state) || target < 0 {
		return errors.New("proposer: The specified target player is not valid")
	}
	if !target.Equivalent(state.CurrentPlayerIndex()) || !target.Equivalent(proposer) {
		return errors.New("proposer: it's not your turn")
	}
	_, players := concreteStates(state)
	player := players[m.TargetPlayerIndex]
	if player.Counter < 1 {
		return errors.New("counter")
	}
	if player.Hand.NumComponents() == 0 {
		return errors.New("hand")
	}
	return nil
}

// actorGoldenCategory maps a declarative failure onto the oracle's check
// names so the fence compares which check failed, not just nil-ness.
func actorGoldenCategory(err error) string {
	if err == nil {
		return "legal"
	}
	msg := err.Error()
	switch {
	case strings.HasPrefix(msg, "proposer: "):
		return msg
	case msg == "it's not your turn" || msg == "The specified target player is not valid":
		return "proposer: " + msg
	case strings.HasPrefix(msg, "requires at least"):
		return "counter"
	case strings.HasPrefix(msg, "requires the stack to not be empty"):
		return "hand"
	}
	return msg
}

func TestGoldenMovePlayerPathsCurrentPlayerMove(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveActorGolden),
				WithMoveName("Actor Golden"),
				WithLegalPreconditions(
					legal.PropAtLeast("player.Counter", 1),
					legal.StackNotEmpty("player.Hand"),
				),
			),
		)
	}
	manager, err := newGameManager(installer)
	assert.For(t, "boots").ThatActual(err).IsNil()
	if manager == nil {
		return
	}

	players := []boardgame.PlayerIndex{0, 1, 2, 3}
	specials := []boardgame.PlayerIndex{boardgame.AdminPlayerIndex, boardgame.ObserverPlayerIndex, boardgame.AnyPlayerIndex}
	currents := append([]boardgame.PlayerIndex{boardgame.AnyPlayerIndex, boardgame.ObserverPlayerIndex}, players...)
	targets := append(append([]boardgame.PlayerIndex{}, players...), specials...)
	proposers := append(append([]boardgame.PlayerIndex{}, players...), boardgame.AdminPlayerIndex, boardgame.ObserverPlayerIndex)

	// Player states: 0 passes everything, 1 has no counter, 2 has an empty
	// hand, 3 passes too. Distinct failures per player prove the plan reads
	// the move's player, not merely some player.
	setUp := func(game *boardgame.Game, current boardgame.PlayerIndex) boardgame.ImmutableState {
		state := game.CurrentState()
		gameState, playerStates := concreteStates(state)
		gameState.CurrentPlayer = current
		for i, p := range playerStates {
			p.Counter = 1
			if i == 1 {
				p.Counter = 0
			}
			if i == 2 {
				for p.Hand.NumComponents() > 0 {
					assert.For(t).ThatActual(p.Hand.First().MoveToLastSlot(gameState.DiscardStack)).IsNil()
				}
			} else if p.Hand.NumComponents() == 0 {
				assert.For(t).ThatActual(gameState.DrawStack.First().MoveToFirstSlot(p.Hand)).IsNil()
			}
		}
		return state
	}

	sawSimultaneousLegal := false
	sawAdminOnBehalfFailure := false
	for _, current := range currents {
		game, err := manager.NewDefaultGame()
		assert.For(t).ThatActual(err).IsNil()
		state := setUp(game, current)
		for _, target := range targets {
			for _, proposer := range proposers {
				move := game.MoveByName("Actor Golden").(*moveActorGolden)
				move.TargetPlayerIndex = target
				want := actorGoldenCategory(legacyActorGoldenLegal(move, state, proposer))
				got := actorGoldenCategory(move.Legal(state, proposer))
				if want != got {
					t.Errorf("current=%v target=%v proposer=%v: legacy %q, declarative %q", current, target, proposer, want, got)
				}
				if current == boardgame.AnyPlayerIndex && want == "legal" {
					sawSimultaneousLegal = true
				}
				if current == boardgame.AnyPlayerIndex && proposer == boardgame.AdminPlayerIndex && (want == "counter" || want == "hand") {
					sawAdminOnBehalfFailure = true
				}
			}
		}
	}
	assert.For(t, "fence exercises a legal simultaneous move").ThatActual(sawSimultaneousLegal).IsTrue()
	assert.For(t, "fence exercises admin acting for a player who fails a rule").ThatActual(sawAdminOnBehalfFailure).IsTrue()
}

// TestMovePlayerSelectorMatchesPlayerPath pins that the typed selector and
// the wire path are one concept: legal.Player() spells "player.", and the
// deprecated legal.CurrentPlayer() is the same selector.
func TestMovePlayerSelectorMatchesPlayerPath(t *testing.T) {
	spec := legal.PlayerBoolAt(legal.Player(), "Eliminated", false)
	assert.For(t).ThatActual(spec.Args[0]).Equals(fmt.Sprintf("player.%s", "Eliminated"))
	assert.For(t, "deprecated alias").ThatActual(legal.CurrentPlayer()).Equals(legal.Player())
}

// TestStackPathPlayerAgreesWithLegalityDuringSimultaneousPlay pins that a
// "player." stack path and a "player.X" legal path name the same player. In a
// simultaneous phase, MoveComponentToSlot's declarative legality resolves
// "player.Hand" to the move's player, so Apply must move the card into that
// same player's hand rather than failing for lack of a concrete current
// player.
func TestStackPathPlayerAgreesWithLegalityDuringSimultaneousPlay(t *testing.T) {
	manager, err := newGameManager(playCardInstaller)
	assert.For(t, "boots").ThatActual(err).IsNil()
	if manager == nil {
		return
	}
	for _, target := range []boardgame.PlayerIndex{0, 1, 2, 3} {
		for _, proposer := range []boardgame.PlayerIndex{target, boardgame.AdminPlayerIndex} {
			game, err := manager.NewDefaultGame()
			assert.For(t).ThatActual(err).IsNil()
			state, ok := game.CurrentState().(boardgame.State)
			if !ok {
				t.Fatal("current state is not mutable")
			}
			gameState, players := concreteStates(state)
			gameState.CurrentPlayer = boardgame.AnyPlayerIndex

			move := game.MoveByName("Play Card To Slot").(*movePlayCardToSlot)
			move.TargetPlayerIndex = target
			if err := move.Legal(state, proposer); err != nil {
				t.Fatalf("target=%v proposer=%v: expected legal, got %v", target, proposer, err)
			}
			before := players[target].Hand.NumComponents()
			if err := move.Apply(state); err != nil {
				t.Fatalf("target=%v proposer=%v: legal move failed to apply: %v", target, proposer, err)
			}
			if got := players[target].Hand.NumComponents(); got != before+1 {
				t.Fatalf("target=%v proposer=%v: the move's player's hand = %d, want %d", target, proposer, got, before+1)
			}
		}
	}

	// A move that names no player of its own still finds nobody to act for.
	game, err := manager.NewDefaultGame()
	assert.For(t).ThatActual(err).IsNil()
	state := game.CurrentState().(boardgame.State)
	gameState, _ := concreteStates(state)
	gameState.CurrentPlayer = boardgame.AnyPlayerIndex
	_, err = ResolveStackPath(nil, state, "player.Hand")
	assert.For(t, "no on-behalf player").ThatActual(err).IsNotNil()
}

// TestDrawToPlayerAgreesWithLegalityDuringSimultaneousPlay is the
// DrawToPlayer counterpart: its contributed "player.<Hand>" precondition and
// its Apply must name the same player in an AnyPlayerIndex phase.
func TestDrawToPlayerAgreesWithLegalityDuringSimultaneousPlay(t *testing.T) {
	manager, err := newGameManager(drawToPlayerInstaller(new(moveDrawOneToPlayer), "Draw One", WithPlayerProperty("Hand")))
	assert.For(t, "boots").ThatActual(err).IsNil()
	if manager == nil {
		return
	}
	for _, target := range []boardgame.PlayerIndex{0, 1, 2, 3} {
		for _, proposer := range []boardgame.PlayerIndex{target, boardgame.AdminPlayerIndex} {
			game, err := manager.NewDefaultGame()
			assert.For(t).ThatActual(err).IsNil()
			state := game.CurrentState().(boardgame.State)
			gameState, players := concreteStates(state)
			gameState.CurrentPlayer = boardgame.AnyPlayerIndex

			move := game.MoveByName("Draw One").(*moveDrawOneToPlayer)
			move.TargetPlayerIndex = target
			if err := move.Legal(state, proposer); err != nil {
				t.Fatalf("target=%v proposer=%v: expected legal, got %v", target, proposer, err)
			}
			before := players[target].Hand.NumComponents()
			if err := move.Apply(state); err != nil {
				t.Fatalf("target=%v proposer=%v: legal move failed to apply: %v", target, proposer, err)
			}
			if got := players[target].Hand.NumComponents(); got != before+1 {
				t.Fatalf("target=%v proposer=%v: hand = %d, want %d", target, proposer, got, before+1)
			}
		}
	}
}

//boardgame:codegen
type moveAnyPlayerActor struct {
	AnyPlayer
}

func (m *moveAnyPlayerActor) Apply(state boardgame.State) error { return nil }

// TestGoldenAnyPlayerMovePlayerIgnoresCurrentPlayer fences the rule for a
// move made for its own player that nothing ties to the turn
// (moves.AnyPlayer): "player.X" and the "player." stack path must name the
// player the imperative oracle would read, players[m.TargetPlayerIndex] once
// AnyPlayer.Legal has passed (and filled an unset target in from the
// proposer), whatever the current player is. That covers a turn-based phase
// played out of turn and an ObserverPlayerIndex gathering-style phase, the
// two cases where reading the current player would pick the wrong player or
// nobody.
func TestGoldenAnyPlayerMovePlayerIgnoresCurrentPlayer(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(auto.MustConfig(new(moveAnyPlayerActor), WithMoveName("Any Player Actor")))
	}
	manager, err := newGameManager(installer)
	assert.For(t, "boots").ThatActual(err).IsNil()
	if manager == nil {
		return
	}

	players := []boardgame.PlayerIndex{0, 1, 2, 3}
	currents := append([]boardgame.PlayerIndex{boardgame.AnyPlayerIndex, boardgame.ObserverPlayerIndex}, players...)
	targets := append([]boardgame.PlayerIndex{boardgame.ObserverPlayerIndex}, players...)
	proposers := append([]boardgame.PlayerIndex{boardgame.AdminPlayerIndex}, players...)

	checked, outOfTurn := 0, 0
	for _, current := range currents {
		game, err := manager.NewDefaultGame()
		assert.For(t).ThatActual(err).IsNil()
		state := game.CurrentState().(boardgame.State)
		gameState, playerStates := concreteStates(state)
		gameState.CurrentPlayer = current
		for i, p := range playerStates {
			p.Counter = 100 + i
		}
		for _, target := range targets {
			for _, proposer := range proposers {
				move := game.MoveByName("Any Player Actor").(*moveAnyPlayerActor)
				move.TargetPlayerIndex = target
				if move.Legal(state, proposer) != nil {
					continue
				}
				want := playerStates[move.TargetPlayerIndex]

				ctx := boardgame.LegalContext{State: state, Move: move, ProposerPlayerIndex: proposer}
				got, _, err := ctx.ResolvePath("player.Counter")
				if err != nil || got != want.Counter {
					t.Errorf("current=%v target=%v proposer=%v: player.Counter = %v (%v), want %v", current, target, proposer, got, err, want.Counter)
				}
				stack, err := ResolveStackPath(move, state, "player.Hand")
				if err != nil || stack != boardgame.Stack(want.Hand) {
					t.Errorf("current=%v target=%v proposer=%v: player.Hand resolved to the wrong player (%v)", current, target, proposer, err)
				}
				checked++
				if current >= 0 && current != move.TargetPlayerIndex {
					outOfTurn++
				}
			}
		}
	}
	assert.For(t, "fence is not vacuous").ThatActual(checked > 0).IsTrue()
	assert.For(t, "fence covers out-of-turn moves").ThatActual(outOfTurn > 0).IsTrue()
}
