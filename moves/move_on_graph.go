package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

// LocationProvider should be implemented by moves that embed MoveOnGraph. It
// returns the LocationBehavior for a given player state. This interface is
// defined here (rather than in moves/interfaces) to avoid an import cycle with
// the behaviors package.
type LocationProvider interface {
	PlayerLocationBehavior(playerState boardgame.ImmutableSubState) *behaviors.LocationBehavior
}

/*
MoveOnGraph is a player-facing move for spatial games. The player specifies a
TargetLocation, and the framework computes the shortest path via the graph
associated with the player's LocationBehavior. The path is stored on the
behavior's LocRemainingPath for the HopAlongPath FixUp to execute hop-by-hop.

The embedding move must implement LocationProvider to tell MoveOnGraph how to
find the player's LocationBehavior.

Optionally, the embedding move may implement:
  - interfaces.SpaceValidator: validate each space in the path
  - interfaces.MovementBudgeter: check and consume a movement budget
  - interfaces.FreeMovePredicate: allow teleportation for certain targets
  - interfaces.FreeMoveApplier: game-specific cleanup after a free move

boardgame:codegen
*/
type MoveOnGraph struct {
	CurrentPlayer
	TargetLocation int
}

// Legal validates the move: checks turn, destination validity, path existence,
// space legality, and movement budget.
func (m *MoveOnGraph) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	locProvider, ok := m.Info().ConcreteMove().(LocationProvider)
	if !ok {
		return errors.New("MoveOnGraph: embedding move must implement LocationProvider")
	}

	playerState := state.ImmutablePlayerStates()[m.TargetPlayerIndex]
	behavior := locProvider.PlayerLocationBehavior(playerState)

	if behavior == nil {
		return errors.New("MoveOnGraph: PlayerLocationBehavior returned nil")
	}

	locationEnum := behavior.LocationEnum()
	if locationEnum == nil {
		return errors.New("MoveOnGraph: LocationBehavior has no graph connected")
	}

	currentIndex := behavior.LocationIndex()
	targetKey := enum.EnumKey(m.TargetLocation)
	targetVal, err := locationEnum.NewImmutableVal(targetKey)
	if err != nil {
		return errors.New("MoveOnGraph: invalid target location: " + err.Error())
	}

	if currentIndex != nil && currentIndex.Value() == targetKey {
		return errors.New("already at the target location")
	}

	// Check for free move (teleport)
	if freePred, ok := m.Info().ConcreteMove().(interfaces.FreeMovePredicate); ok {
		if freePred.IsFreeMove(playerState, targetVal) {
			if err := behavior.MayMoveTo(m.TargetLocation); err != nil {
				return errors.New("cannot move to target location: " + err.Error())
			}
			return nil
		}
	}

	// Compute shortest path
	path, err := behavior.ShortestPathTo(targetVal)
	if err != nil {
		return errors.New("no valid path to target: " + err.Error())
	}
	for _, spaceVal := range path[1:] {
		if err := behavior.MayMoveTo(spaceVal.Value().Int()); err != nil {
			return errors.New("path contains an unreachable stack slot: " + err.Error())
		}
	}

	// Validate each space in the path (skip start)
	if validator, ok := m.Info().ConcreteMove().(interfaces.SpaceValidator); ok {
		for _, spaceVal := range path[1:] {
			if err := validator.SpaceIsLegal(playerState, spaceVal); err != nil {
				return err
			}
		}
	}

	// Check movement budget
	if budgeter, ok := m.Info().ConcreteMove().(interfaces.MovementBudgeter); ok {
		hops := len(path) - 1
		remaining := budgeter.MovesRemaining(playerState)
		if hops > remaining {
			return errors.New("not enough moves remaining")
		}
	}

	return nil
}

// Apply stores the path on the LocationBehavior for HopAlongPath to execute,
// and consumes the movement budget.
func (m *MoveOnGraph) Apply(state boardgame.State) error {

	locProvider, ok := m.Info().ConcreteMove().(LocationProvider)
	if !ok {
		return errors.New("MoveOnGraph: embedding move must implement LocationProvider")
	}

	playerState := state.PlayerStates()[m.TargetPlayerIndex]
	// Get behavior from the immutable view (same underlying struct)
	behavior := locProvider.PlayerLocationBehavior(playerState.(boardgame.ImmutableSubState))

	if behavior == nil {
		return errors.New("MoveOnGraph: PlayerLocationBehavior returned nil")
	}

	locationEnum := behavior.LocationEnum()
	if locationEnum == nil {
		return errors.New("MoveOnGraph: LocationBehavior has no graph connected")
	}

	targetKey := enum.EnumKey(m.TargetLocation)
	targetVal, err := locationEnum.NewImmutableVal(targetKey)
	if err != nil {
		return errors.New("MoveOnGraph: invalid target location: " + err.Error())
	}

	// Check for free move
	if freePred, ok := m.Info().ConcreteMove().(interfaces.FreeMovePredicate); ok {
		if freePred.IsFreeMove(playerState.(boardgame.ImmutableSubState), targetVal) {
			// Move directly
			if err := behavior.MoveTo(m.TargetLocation); err != nil {
				return err
			}
			// Handle game-specific free move cleanup
			if applier, ok := m.Info().ConcreteMove().(interfaces.FreeMoveApplier); ok {
				return applier.ApplyFreeMove(playerState, targetVal)
			}
			return nil
		}
	}

	// Compute path
	path, err := behavior.ShortestPathTo(targetVal)
	if err != nil {
		return errors.New("no valid path to target: " + err.Error())
	}

	// Store path for HopAlongPath (convert []ImmutableVal to []int)
	intPath := make([]int, len(path))
	for i, v := range path {
		intPath[i] = v.Value().Int()
	}
	behavior.LocRemainingPath = intPath

	// Consume movement budget
	if budgeter, ok := m.Info().ConcreteMove().(interfaces.MovementBudgeter); ok {
		hops := len(path) - 1
		if err := budgeter.ConsumeMovement(playerState, hops); err != nil {
			return err
		}
	}

	return nil
}

// FallbackName returns "Move On Graph"
func (m *MoveOnGraph) FallbackName(mgr *boardgame.GameManager) string {
	return "Move On Graph"
}

// FallbackHelpText returns a description of the move.
func (m *MoveOnGraph) FallbackHelpText() string {
	return "Move a token to a target location along the shortest path."
}
