package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
)

// ResolvedSlotField returns the move property a MoveComponentToSlot-shaped
// move reads its chosen slot from: the one named by [WithSlotField], or, when
// that option is absent, the move's only int property, exactly as the move
// resolves it at boot and on every Apply. It is read only, for tools (the
// game-definition exporter) that must report what Go inferred; it changes
// nothing about the move.
//
// move is the configured move (for example one from
// GameManager.ExampleMoves); it must expose its custom configuration, as
// every move built by auto.Config does.
func ResolvedSlotField(move boardgame.Move) (string, error) {
	infoer, ok := move.(moveInfoer)
	if !ok {
		return "", errors.New("move did not expose its custom configuration")
	}
	return resolveSlotField(move, infoer)
}
