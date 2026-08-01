package moves

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
)

// playerScopedTransferInstaller configures a MoveCountComponents whose
// destination is a stack on a PLAYER state rather than on gameState. Before the
// stack-path grammar existed this was simply not expressible: SourceStack and
// DestinationStack both read state.GameState() only, so a player-scoped name
// resolved to nil and the move failed ValidConfiguration at boot.
func playerScopedTransferInstaller(source, destination string, target int) func(*boardgame.GameManager) []boardgame.MoveConfig {
	return func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return AddForPhase(phaseSetUp,
			auto.MustConfig(
				new(MoveCountComponents),
				WithMoveName("Move To Player Stack"),
				WithSourceProperty(source),
				WithDestinationProperty(destination),
				WithTargetCount(target),
			),
		)
	}
}

func TestStackPathReachesPlayerState(t *testing.T) {

	t.Run("destination on the current player", func(t *testing.T) {
		manager, err := newGameManager(playerScopedTransferInstaller("game.DrawStack", "player.Hand", 2))
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		gameState, players := concreteStates(game.CurrentState())
		if got := players[0].Hand.NumComponents(); got != 2 {
			t.Fatalf("current player hand = %d, want 2", got)
		}
		if got := gameState.DrawStack.NumComponents(); got != 50 {
			t.Fatalf("draw stack = %d, want 50", got)
		}
		//No other player should have received anything: this is not a deal.
		if got := players[1].Hand.NumComponents(); got != 0 {
			t.Fatalf("non-current player hand = %d, want 0", got)
		}
	})

	t.Run("source on the current player", func(t *testing.T) {
		manager, err := newGameManager(playerScopedTransferInstaller("player.Hand", "game.DiscardStack", 1))
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		//The player's hand starts empty, so the move can never be legal; the
		//point of this case is that the configuration BOOTS, which it could not
		//before, and that a player-scoped SOURCE resolves.
		move := game.MoveByName("Move To Player Stack")
		if move == nil {
			t.Fatal("move was not installed")
		}
		if err := move.Legal(game.CurrentState(), boardgame.AdminPlayerIndex); err == nil {
			t.Fatal("expected an empty player hand to be an illegal source")
		}
	})

	t.Run("unqualified names still mean gameState", func(t *testing.T) {
		manager, err := newGameManager(playerScopedTransferInstaller("DrawStack", "DiscardStack", 3))
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		gameState, _ := concreteStates(game.CurrentState())
		if got := gameState.DiscardStack.NumComponents(); got != 3 {
			t.Fatalf("discard stack = %d, want 3", got)
		}
	})

	t.Run("a bad path kind is a boot error", func(t *testing.T) {
		_, err := newGameManager(playerScopedTransferInstaller("game.DrawStack", "nonsense.Hand", 1))
		if err == nil {
			t.Fatal("expected a boot error for an unknown path kind")
		}
		if !strings.Contains(err.Error(), "nonsense.Hand") {
			t.Fatalf("boot error did not name the bad path: %v", err)
		}
	})

	t.Run("a misspelled property is a boot error", func(t *testing.T) {
		_, err := newGameManager(playerScopedTransferInstaller("game.DrawStack", "player.NoSuchStack", 1))
		if err == nil {
			t.Fatal("expected a boot error for an unknown player property")
		}
	})
}

// playCardFromTargetPlayerInstaller configures the reusable "play a card" verb
// with a players[move.<Field>] SOURCE, which is the one path kind whose player
// is named by the move rather than by the state.
func playCardFromTargetPlayerInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {
	auto := NewAutoConfigurer(manager.Delegate())
	return AddForPhase(phaseSetUp,
		auto.MustConfig(
			new(movePlayCardToSlot),
			WithMoveName("Play Card From Target Player"),
			WithSourceProperty("players[move.TargetPlayerIndex].Hand"),
			WithDestinationProperty("game.DiscardStack"),
		),
	)
}

// TestStackPathMoveFieldPlayerIsNotRetargeted pins the fourth instance of the
// EnsureValid class this branch has now found four times (SeatPlayer,
// ActivateEmptySeat, CloseEmptySeat, and here).
//
// A players[move.<Field>] path is resolved TWICE per move, by two different
// resolvers: at Legal() time by core's own path grammar
// (resolveLegalPlayerReader in legal_path.go, which MoveComponentToSlot's
// contributed legal.MayMoveToSlot atom goes through) and at Apply() time by
// this package's parsedStackPath.resolve. They must agree on WHICH PLAYER the
// field names, or Legal() judges one player's stacks and Apply() moves another
// player's components. They did not: core rejects an index that is not a
// concrete in-bounds player, while this package ran it through EnsureValid
// first, which silently advances to the next player who may be active.
func TestStackPathMoveFieldPlayerIsNotRetargeted(t *testing.T) {

	manager, err := newGameManager(playCardFromTargetPlayerInstaller)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("new game: %v", err)
	}

	state := game.CurrentState().(boardgame.State)

	stackFor := func(target boardgame.PlayerIndex) boardgame.Stack {
		move := game.MoveByName("Play Card From Target Player")
		if move == nil {
			t.Fatal("move was not installed")
		}
		concrete, ok := move.(*movePlayCardToSlot)
		if !ok {
			t.Fatalf("move was a %T, not the configured type", move)
		}
		concrete.TargetPlayerIndex = target
		return concrete.SourceStack(state)
	}

	t.Run("a concrete player resolves to that player", func(t *testing.T) {
		_, players := concreteStates(state)
		got := stackFor(2)
		if got == nil {
			t.Fatal("a valid target player resolved to no stack")
		}
		if got != players[2].Hand {
			t.Fatal("the path resolved to some player other than the one the field named")
		}
	})

	//The retarget case: 9 is not a player at all in a four-player game, and
	//EnsureValid's Next() wraps it around to player 0 -- a player the move
	//never named, whose hand Apply would then have emptied.
	t.Run("an out-of-range player resolves to nothing", func(t *testing.T) {
		if got := stackFor(9); got != nil {
			t.Fatal("an out-of-range target player silently retargeted onto a real player's stack")
		}
	})

	t.Run("Admin resolves to nothing", func(t *testing.T) {
		if got := stackFor(boardgame.AdminPlayerIndex); got != nil {
			t.Fatal("AdminPlayerIndex resolved to a player's stack")
		}
	})

	t.Run("Observer resolves to nothing", func(t *testing.T) {
		if got := stackFor(boardgame.ObserverPlayerIndex); got != nil {
			t.Fatal("ObserverPlayerIndex resolved to a player's stack")
		}
	})
}

// TestUnconsumedStackPropertyOptionIsBootError covers the config-validation
// hole this task found: WithSourceProperty and WithDestinationProperty were the
// only stack-shaped options with no entry in validateCustomConfiguration, so
// passing them to a move that cannot consume them was silently ignored.
func TestUnconsumedStackPropertyOptionIsBootError(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return AddForPhase(phaseSetUp,
			auto.MustConfig(
				new(NoOp),
				WithMoveName("No Op With Stray Source"),
				WithSourceProperty("DrawStack"),
			),
		)
	}
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected MustConfig to reject an unconsumed WithSourceProperty")
		}
		if !strings.Contains(strings.ToLower(err2string(recovered)), "withsourceproperty") {
			t.Fatalf("panic did not name the option: %v", recovered)
		}
	}()
	newGameManager(installer)
}

func err2string(v interface{}) string {
	switch value := v.(type) {
	case error:
		return value.Error()
	case string:
		return value
	}
	return ""
}
