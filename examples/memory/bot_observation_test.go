package memory

import (
	"sync"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/bots"
	storageMemory "github.com/jkomoros/boardgame/storage/memory"
)

type blockingDiagramDelegate struct {
	boardgame.GameDelegate
	once    sync.Once
	entered chan struct{}
	release chan struct{}
}

func (d *blockingDiagramDelegate) ConfigureLegalTemplates() map[string]string {
	return d.GameDelegate.(interface{ ConfigureLegalTemplates() map[string]string }).ConfigureLegalTemplates()
}

func (d *blockingDiagramDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	return d.GameDelegate.(interface {
		GameEndConditionMet(boardgame.ImmutableState) bool
	}).GameEndConditionMet(state)
}

func (d *blockingDiagramDelegate) PlayerScore(player boardgame.ImmutableSubState) int {
	return d.GameDelegate.(interface {
		PlayerScore(boardgame.ImmutableSubState) int
	}).PlayerScore(player)
}

func (d *blockingDiagramDelegate) LowScoreWins() bool {
	return d.GameDelegate.(interface{ LowScoreWins() bool }).LowScoreWins()
}

func (d *blockingDiagramDelegate) Diagram(state boardgame.ImmutableState) string {
	d.once.Do(func() {
		close(d.entered)
		<-d.release
	})
	return d.GameDelegate.Diagram(state)
}

func TestBotObservationRejectsAdvanceDuringViewerSerialization(t *testing.T) {
	delegate := &blockingDiagramDelegate{
		GameDelegate: NewDelegate(),
		entered:      make(chan struct{}),
		release:      make(chan struct{}),
	}
	manager, err := boardgame.NewGameManager(delegate, storageMemory.NewStorageManager())
	if err != nil {
		t.Fatal(err)
	}
	manager.Internals().UseManualTimers()
	t.Cleanup(manager.Internals().Close)
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	player := game.CurrentState().CurrentPlayerIndex()
	observed := make(chan error, 1)
	go func() {
		_, err := bots.Observe(game, player)
		observed <- err
	}()
	<-delegate.entered

	move, err := game.MoveWithInput("Reveal Card", map[string]interface{}{"CardIndex": 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := <-game.ProposeMove(move, player); err != nil {
		t.Fatal(err)
	}
	close(delegate.release)
	if err := <-observed; err == nil {
		t.Fatal("observation spanning a committed move was accepted")
	}
	observation, err := bots.Observe(game, player)
	if err != nil {
		t.Fatalf("settled observation after the commit: %v", err)
	}
	if observation.Version != game.Version() || observation.Version == 0 {
		t.Fatalf("expected the nonzero committed version, got %d", observation.Version)
	}
}
