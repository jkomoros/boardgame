package werewolf

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
)

func TestNewGameManager(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())
	if err != nil {
		t.Fatalf("NewGameManager: %v", err)
	}
	if manager == nil {
		t.Fatal("NewGameManager returned nil")
	}

	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("NewDefaultGame: %v", err)
	}
	if game == nil {
		t.Fatal("NewDefaultGame returned nil")
	}
}

// TestRoleHiddenFromObserver pins the core companion-mode privacy contract:
// the Table surface connects as ObserverPlayerIndex, so the framework's
// sanitization with behaviors.PlayerRole's default sanitize:"other:hidden"
// tag hides every player's role from the projector. If this test fails,
// the projector would display everyone's secret role on the shared screen.
func TestRoleHiddenFromObserver(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())
	if err != nil {
		t.Fatalf("NewGameManager: %v", err)
	}

	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("NewDefaultGame: %v", err)
	}

	state := game.CurrentState()

	// Verify roles were assigned (at least one werewolf exists)
	_, players := concreteStates(state)
	hasWerewolf := false
	for _, p := range players {
		if p.Role.Value() == roleWerewolf {
			hasWerewolf = true
			break
		}
	}
	if !hasWerewolf {
		t.Fatal("Expected at least one werewolf after game setup")
	}

	// Sanitize for ObserverPlayerIndex (the projector's perspective).
	// With sanitize:"other:hidden" on PlayerRole.Role, the observer
	// should NOT be able to see any player's role.
	sanitized, err := state.SanitizedForPlayer(boardgame.ObserverPlayerIndex)
	if err != nil {
		t.Fatalf("SanitizedForPlayer(Observer): %v", err)
	}

	for i, ps := range sanitized.ImmutablePlayerStates() {
		reader := ps.Reader()
		roleProp, err := reader.ImmutableEnumProp("Role")
		if err != nil {
			t.Fatalf("Player %d: reading Role prop: %v", i, err)
		}
		if roleProp != nil && roleProp.Value() != 0 {
			t.Errorf("Player %d: observer can see Role=%d; expected hidden (0)", i, roleProp.Value())
		}
	}
}

// TestRoleVisibleToSelf pins the other half of the companion-mode contract:
// each player's phone (Hand surface) connects as PlayerIndex(n), and
// sanitize:"other:hidden" only hides OTHER players' roles — you can always
// see your own. If this test fails, phones would show "Unknown" for the
// player's own role.
func TestRoleVisibleToSelf(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())
	if err != nil {
		t.Fatalf("NewGameManager: %v", err)
	}

	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("NewDefaultGame: %v", err)
	}

	state := game.CurrentState()

	// Find a player who is a werewolf.
	_, players := concreteStates(state)
	werewolfIdx := -1
	for i, p := range players {
		if p.Role.Value() == roleWerewolf {
			werewolfIdx = i
			break
		}
	}
	if werewolfIdx < 0 {
		t.Fatal("No werewolf found")
	}

	// Sanitize for that player's own perspective — they should see
	// their own role (sanitize:"other:hidden" hides OTHER players'
	// roles, not your own).
	sanitized, err := state.SanitizedForPlayer(boardgame.PlayerIndex(werewolfIdx))
	if err != nil {
		t.Fatalf("SanitizedForPlayer(%d): %v", werewolfIdx, err)
	}

	selfState := sanitized.ImmutablePlayerStates()[werewolfIdx]
	reader := selfState.Reader()
	roleProp, err := reader.ImmutableEnumProp("Role")
	if err != nil {
		t.Fatalf("Reading own Role: %v", err)
	}
	if roleProp == nil || roleProp.Value() != roleWerewolf {
		t.Errorf("Player %d cannot see their own werewolf role; got %v", werewolfIdx, roleProp)
	}
}
