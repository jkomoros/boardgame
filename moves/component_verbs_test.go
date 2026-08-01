package moves

import (
	"errors"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

// movePlayCardToSlot is the reusable "play a card" verb with nothing added.
//
//boardgame:codegen
type movePlayCardToSlot struct {
	MoveComponentToSlot
	Slot int
}

// movePlayCardWithResidue is the shape every migration target actually has:
// the reusable verb PLUS the game's own declarative precondition and its own
// imperative residue. If embedding the verb cost a game its declarative
// surface, no example game could have adopted it.
//
//boardgame:codegen
type movePlayCardWithResidue struct {
	MoveComponentToSlot
	Slot int
}

func (m *movePlayCardWithResidue) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	game, _ := concreteStates(state)
	if game.Counter != 0 {
		return errors.New("this game says no while the counter is set")
	}
	return nil
}

//boardgame:codegen
type moveDrawOneToPlayer struct {
	DrawToPlayer
}

// moveDrawWithBookkeeping super-calls the reusable Apply and then does its own
// bookkeeping, which is what blackjack's hit move needs.
//
//boardgame:codegen
type moveDrawWithBookkeeping struct {
	DrawToPlayer
}

func (m *moveDrawWithBookkeeping) Apply(state boardgame.State) error {
	if err := m.DrawToPlayer.Apply(state); err != nil {
		return err
	}
	game, _ := concreteStates(state)
	game.Counter++
	return nil
}

//boardgame:codegen
type moveTwoIntFields struct {
	MoveComponentToSlot
	Slot  int
	Other int
}

func playCardInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {
	auto := NewAutoConfigurer(manager.Delegate())
	return AddForPhase(phaseSetUp,
		auto.MustConfig(
			new(movePlayCardToSlot),
			WithMoveName("Play Card To Slot"),
			WithSourceProperty("game.DrawStack"),
			WithDestinationProperty("player.Hand"),
		),
	)
}

func playCardWithResidueInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {
	auto := NewAutoConfigurer(manager.Delegate())
	return AddForPhase(phaseSetUp,
		auto.MustConfig(
			new(movePlayCardWithResidue),
			WithMoveName("Play Card With Residue"),
			WithSourceProperty("game.DrawStack"),
			WithDestinationProperty("player.Hand"),
			WithLegalPreconditions(
				legal.StackNotEmpty("game.DrawStack"),
			),
		),
	)
}

func drawToPlayerInstaller(move AutoConfigurableMove, name string, options ...CustomConfigurationOption) func(*boardgame.GameManager) []boardgame.MoveConfig {
	return func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return AddForPhase(phaseSetUp,
			auto.MustConfig(move, append([]CustomConfigurationOption{WithMoveName(name)}, options...)...),
		)
	}
}

func TestMoveComponentToSlot(t *testing.T) {

	t.Run("moves the first component into the chosen slot", func(t *testing.T) {
		manager, err := newGameManager(playCardInstaller)
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		before, _ := concreteStates(game.CurrentState())
		expected := before.DrawStack.First().DeckIndex()

		move := game.MoveByName("Play Card To Slot")
		if err := <-game.ProposeMove(move, 0); err != nil {
			t.Fatalf("propose: %v", err)
		}

		after, players := concreteStates(game.CurrentState())
		if got := players[0].Hand.NumComponents(); got != 1 {
			t.Fatalf("hand = %d, want 1", got)
		}
		if players[0].Hand.ComponentAt(0).DeckIndex() != expected {
			t.Fatal("the component that moved was not the first one in the source")
		}
		if got := after.DrawStack.NumComponents(); got != 51 {
			t.Fatalf("draw stack = %d, want 51", got)
		}
	})

	t.Run("an occupied or out-of-range slot is illegal", func(t *testing.T) {
		manager, err := newGameManager(playCardInstaller)
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		move := game.MoveByName("Play Card To Slot").(*movePlayCardToSlot)
		//Hand is empty, so the only valid insertion point is 0.
		move.Slot = 4
		if err := move.Legal(game.CurrentState(), 0); err == nil {
			t.Fatal("expected an out-of-range slot to be illegal")
		}
	})

	t.Run("an empty source is illegal", func(t *testing.T) {
		manager, err := newGameManager(playCardInstaller)
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		gameState, players := concreteStates(game.CurrentState())
		if err := gameState.DrawStack.MoveAllTo(players[1].OtherHand); err != nil {
			t.Fatalf("emptying source: %v", err)
		}
		move := game.MoveByName("Play Card To Slot")
		if err := move.Legal(game.CurrentState(), 0); err == nil {
			t.Fatal("expected an empty source stack to be illegal")
		}
	})

	t.Run("the embedding game keeps its declarative surface", func(t *testing.T) {
		manager, err := newGameManager(playCardWithResidueInstaller)
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		move := game.MoveByName("Play Card With Residue").(*movePlayCardWithResidue)
		if err := move.Legal(game.CurrentState(), 0); err != nil {
			t.Fatalf("the move should be legal: %v", err)
		}
		//The verb's own contributed atom still runs: slot 1 is out of range
		//for an empty destination.
		move.Slot = 1
		if err := move.Legal(game.CurrentState(), 0); err == nil {
			t.Fatal("the verb's contributed MayMoveFirstToSlot atom did not run")
		}
		//And so does the game's own LegalCustom residue.
		move.Slot = 0
		gameState, _ := concreteStates(game.CurrentState())
		gameState.Counter = 1
		if err := move.Legal(game.CurrentState(), 0); err == nil || !strings.Contains(err.Error(), "counter is set") {
			t.Fatalf("the game's own LegalCustom did not run: %v", err)
		}
	})

	t.Run("an ambiguous slot field is a boot error naming the candidates", func(t *testing.T) {
		installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
			auto := NewAutoConfigurer(manager.Delegate())
			return AddForPhase(phaseSetUp,
				auto.MustConfig(
					new(moveTwoIntFields),
					WithMoveName("Two Int Fields"),
					WithSourceProperty("game.DrawStack"),
					WithDestinationProperty("player.Hand"),
				),
			)
		}
		_, err := newGameManager(installer)
		if err == nil {
			t.Fatal("expected an ambiguous slot field to be a boot error")
		}
		if !strings.Contains(err.Error(), "WithSlotField") || !strings.Contains(err.Error(), "Other") {
			t.Fatalf("boot error did not name the candidates and the fix: %v", err)
		}
	})

	t.Run("naming the slot field resolves the ambiguity", func(t *testing.T) {
		installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
			auto := NewAutoConfigurer(manager.Delegate())
			return AddForPhase(phaseSetUp,
				auto.MustConfig(
					new(moveTwoIntFields),
					WithMoveName("Two Int Fields"),
					WithSourceProperty("game.DrawStack"),
					WithDestinationProperty("player.Hand"),
					WithSlotField("Slot"),
				),
			)
		}
		if _, err := newGameManager(installer); err != nil {
			t.Fatalf("WithSlotField should have resolved it: %v", err)
		}
	})
}

func TestDrawToPlayer(t *testing.T) {

	t.Run("draws from the discovered draw stack into the named player stack", func(t *testing.T) {
		manager, err := newGameManager(drawToPlayerInstaller(new(moveDrawOneToPlayer), "Draw One", WithPlayerProperty("Hand")))
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		if err := <-game.ProposeMove(game.MoveByName("Draw One"), 0); err != nil {
			t.Fatalf("propose: %v", err)
		}
		gameState, players := concreteStates(game.CurrentState())
		if got := players[0].Hand.NumComponents(); got != 1 {
			t.Fatalf("hand = %d, want 1", got)
		}
		if got := gameState.DrawStack.NumComponents(); got != 51 {
			t.Fatalf("draw stack = %d, want 51", got)
		}
		if got := players[1].Hand.NumComponents(); got != 0 {
			t.Fatalf("a non-current player drew: %d", got)
		}
	})

	t.Run("an ambiguous player stack is a boot error naming the candidates", func(t *testing.T) {
		_, err := newGameManager(drawToPlayerInstaller(new(moveDrawOneToPlayer), "Draw One"))
		if err == nil {
			t.Fatal("expected an ambiguous player stack to be a boot error")
		}
		if !strings.Contains(err.Error(), "WithPlayerProperty") {
			t.Fatalf("boot error did not name the fix: %v", err)
		}
	})

	t.Run("an empty draw stack is illegal", func(t *testing.T) {
		manager, err := newGameManager(drawToPlayerInstaller(new(moveDrawOneToPlayer), "Draw One", WithPlayerProperty("Hand")))
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		gameState, _ := concreteStates(game.CurrentState())
		if err := gameState.DrawStack.MoveAllTo(gameState.DiscardStack); err != nil {
			t.Fatalf("emptying draw stack: %v", err)
		}
		if err := game.MoveByName("Draw One").Legal(game.CurrentState(), 0); err == nil {
			t.Fatal("expected an empty draw stack to be illegal")
		}
	})

	t.Run("an embedding move can super-call Apply", func(t *testing.T) {
		manager, err := newGameManager(drawToPlayerInstaller(new(moveDrawWithBookkeeping), "Draw With Bookkeeping", WithPlayerProperty("Hand")))
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		if err := <-game.ProposeMove(game.MoveByName("Draw With Bookkeeping"), 0); err != nil {
			t.Fatalf("propose: %v", err)
		}
		gameState, players := concreteStates(game.CurrentState())
		if got := players[0].Hand.NumComponents(); got != 1 {
			t.Fatalf("hand = %d, want 1", got)
		}
		if gameState.Counter != 1 {
			t.Fatalf("the embedding move's own bookkeeping did not run: Counter = %d", gameState.Counter)
		}
	})
}

// movePlayCardReordered is the suppress-then-reauthor recipe: the verb's own
// defining atom is suppressed so that a gate the GAME authored runs first and
// owns the message, then re-authored behind it so the check itself survives.
//
//boardgame:codegen
type movePlayCardReordered struct {
	MoveComponentToSlot
	Slot int
}

// movePlayCardSuppressedOnly suppresses the defining atom and never puts it
// back, which the boot guard must refuse.
//
//boardgame:codegen
type movePlayCardSuppressedOnly struct {
	MoveComponentToSlot
	Slot int
}

// moveDrawSuppressedOnly is the DrawToPlayer flavor of the same mistake.
//
//boardgame:codegen
type moveDrawSuppressedOnly struct {
	DrawToPlayer
}

// reorderedGateMessage is legal.StackNotEmpty's default rendering
// (legal/templates.go's TemplateStackNotEmpty), which is what the game's own
// gate says when it wins the message.
const reorderedGateMessage = "requires the stack to not be empty"

func playCardReorderedInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {
	auto := NewAutoConfigurer(manager.Delegate())
	return AddForPhase(phaseSetUp,
		auto.MustConfig(
			new(movePlayCardReordered),
			WithMoveName("Play Card Reordered"),
			WithSourceProperty("game.DrawStack"),
			WithDestinationProperty("player.Hand"),
			//Suppress the verb's defining atom so it no longer runs first...
			WithoutLegalPrecondition(PreconditionMayMoveFirstToSlot),
			WithLegalPreconditions(
				//...let this game's own gate own the message...
				legal.StackNotEmpty("game.DiscardStack"),
				//...and put the verb's check back, behind it.
				legal.MayMoveFirstToSlot("game.DrawStack", "player.Hand", "move.Slot"),
			),
		),
	)
}

// TestSuppressThenReauthorPreservesTheCheck pins the documented remedy for the
// base-first ordering rule (TUTORIAL.md's "reordering the verb's own check"
// recipe, moves/doc.go's same paragraph). Before this existed the library's
// own documentation told creators to keep their bespoke move instead, so this
// test is what makes the advice it replaced safe to delete: the recipe must
// both hand the message to the game's gate AND keep the verb's check running.
func TestSuppressThenReauthorPreservesTheCheck(t *testing.T) {

	newReorderedGame := func(t *testing.T) *boardgame.Game {
		t.Helper()
		manager, err := newGameManager(playCardReorderedInstaller)
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		return game
	}

	t.Run("the game's own gate now owns the message", func(t *testing.T) {
		game := newReorderedGame(t)
		move := game.MoveByName("Play Card Reordered").(*movePlayCardReordered)
		//Slot 3 is out of range for an empty hand, so before the suppression
		//the verb's own atom would have failed FIRST and its message would
		//have won. DiscardStack is empty too, so the game's gate also fails.
		move.Slot = 3
		err := move.Legal(game.CurrentState(), 0)
		if err == nil {
			t.Fatal("expected the move to be illegal")
		}
		if !strings.Contains(err.Error(), reorderedGateMessage) {
			t.Fatalf("the game's own gate did not win the message: %v", err)
		}
	})

	t.Run("the re-authored check still runs", func(t *testing.T) {
		game := newReorderedGame(t)
		gameState, _ := concreteStates(game.CurrentState())
		//Satisfy the game's own gate, so only the re-authored atom can object.
		if err := gameState.DrawStack.First().MoveToFirstSlot(gameState.DiscardStack); err != nil {
			t.Fatalf("filling discard stack: %v", err)
		}
		move := game.MoveByName("Play Card Reordered").(*movePlayCardReordered)
		move.Slot = 3
		err := move.Legal(game.CurrentState(), 0)
		if err == nil {
			t.Fatal("suppress-then-reauthor lost the verb's own check entirely")
		}
		if strings.Contains(err.Error(), reorderedGateMessage) {
			t.Fatalf("the game's gate should be satisfied now: %v", err)
		}
	})

	t.Run("and a legal move is still legal", func(t *testing.T) {
		game := newReorderedGame(t)
		gameState, _ := concreteStates(game.CurrentState())
		if err := gameState.DrawStack.First().MoveToFirstSlot(gameState.DiscardStack); err != nil {
			t.Fatalf("filling discard stack: %v", err)
		}
		move := game.MoveByName("Play Card Reordered").(*movePlayCardReordered)
		move.Slot = 0
		if err := move.Legal(game.CurrentState(), 0); err != nil {
			t.Fatalf("the move should be legal: %v", err)
		}
	})
}

// TestSuppressingTheDefiningAtomWithoutReauthoringIsABootError pins the hazard
// half of the recipe. Neither verb's Apply re-checks its own defining atom --
// both lean on ComponentInstance.MoveTo's internal net -- so a suppression
// with nothing put back publishes a legality verdict Apply can then refuse.
func TestSuppressingTheDefiningAtomWithoutReauthoringIsABootError(t *testing.T) {

	t.Run("MoveComponentToSlot", func(t *testing.T) {
		installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
			auto := NewAutoConfigurer(manager.Delegate())
			return AddForPhase(phaseSetUp,
				auto.MustConfig(
					new(movePlayCardSuppressedOnly),
					WithMoveName("Play Card Suppressed Only"),
					WithSourceProperty("game.DrawStack"),
					WithDestinationProperty("player.Hand"),
					WithoutLegalPrecondition(PreconditionMayMoveFirstToSlot),
				),
			)
		}
		_, err := newGameManager(installer)
		if err == nil {
			t.Fatal("expected suppressing the defining atom with nothing put back to be a boot error")
		}
		if !strings.Contains(err.Error(), "mayMoveFirstToSlot") || !strings.Contains(err.Error(), "MayMoveFirstToSlot") {
			t.Fatalf("boot error did not name the suppressed atom and the constructor that restores it: %v", err)
		}
	})

	t.Run("DrawToPlayer", func(t *testing.T) {
		_, err := newGameManager(drawToPlayerInstaller(
			new(moveDrawSuppressedOnly),
			"Draw Suppressed Only",
			WithPlayerProperty("Hand"),
			WithoutLegalPrecondition(PreconditionMayMoveFirstTo),
		))
		if err == nil {
			t.Fatal("expected suppressing the defining atom with nothing put back to be a boot error")
		}
		if !strings.Contains(err.Error(), "mayMoveFirstTo") {
			t.Fatalf("boot error did not name the suppressed atom: %v", err)
		}
	})
}

// TestMoveComponentToSlotDoesNotContributeStackConstraints pins the reason
// examples/memory can name two gameState stacks and still keep its own
// endgame messages: Default derives a generic "stackConstraints" atom from
// WithSourceProperty+WithDestinationProperty, which asks whether the FIRST
// component would be accepted ANYWHERE in the destination. That is neither the
// component nor the slot this move is about, and it would run ahead of both
// the proposer check and everything the game authored.
func TestMoveComponentToSlotDoesNotContributeStackConstraints(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return AddForPhase(phaseSetUp,
			auto.MustConfig(
				new(movePlayCardToSlot),
				WithMoveName("Play Card Between Game Stacks"),
				WithSourceProperty("game.DrawStack"),
				WithDestinationProperty("game.DiscardStack"),
			),
		)
	}
	manager, err := newGameManager(installer)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("new game: %v", err)
	}
	move := game.MoveByName("Play Card Between Game Stacks").(*movePlayCardToSlot)

	provider, ok := boardgame.Move(move).(PreconditionsProvider)
	if !ok {
		t.Fatal("the move did not implement PreconditionsProvider")
	}
	for _, spec := range provider.ContributedPreconditions() {
		if spec.Name == string(PreconditionStackConstraints) {
			t.Fatalf("MoveComponentToSlot contributed the generic %q atom alongside its own slot-aware one", spec.Name)
		}
	}
}
