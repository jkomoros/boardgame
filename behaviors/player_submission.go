package behaviors

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
PlayerSubmission is a struct designed to be embedded anonymously in your
PlayerStates. It tracks whether a player has submitted a selection during a
simultaneous play phase. Used in conjunction with [moves.AllPlayersSubmitted]
and [moves.ResetAllPlayerSubmissions] to implement the common pattern of
simultaneous secret selection followed by a reveal.

PlayerSubmission is not a [Connectable] behavior; simply embed it and the
framework handles the rest. The PlayerSubmitted bool is automatically included
in the generated PropertyReader by boardgame-util codegen.

Three common usage patterns for submissions:

  - Locked: The player-facing move's Legal() checks !HasSubmitted(), preventing changes after submission.
  - Always changeable: The move ignores HasSubmitted() and overwrites the selection freely.
  - Changeable with re-confirm: A "change" move calls ResetSubmission() first, then the player re-submits.

The framework does not enforce any of these patterns; the game's player-facing
move controls when the flag transitions. [moves.AllPlayersSubmitted] simply
checks whether all active players' flags are currently true.
*/
type PlayerSubmission struct {
	PlayerSubmitted bool
}

// HasSubmitted returns whether the player has submitted their selection.
// Satisfies the PlayerSubmitter interface in moves/interfaces.
func (p *PlayerSubmission) HasSubmitted() bool {
	return p.PlayerSubmitted
}

// SetPlayerSubmitted marks the player as having submitted their selection.
// Satisfies the PlayerSubmitter interface in moves/interfaces.
func (p *PlayerSubmission) SetPlayerSubmitted() {
	p.PlayerSubmitted = true
}

// ResetSubmission clears the player's submitted flag, typically done at the
// start of a new simultaneous selection phase. Satisfies the PlayerSubmitter
// interface in moves/interfaces.
func (p *PlayerSubmission) ResetSubmission() {
	p.PlayerSubmitted = false
}

// PlayerHasSubmitted is a convenience method that does the cast to
// interfaces.PlayerSubmitter, so you don't have to. You can pass any
// playerState to it and it will return true if the player state implements
// interfaces.PlayerSubmitter and HasSubmitted returns true, false otherwise.
func PlayerHasSubmitted(playerState boardgame.ImmutableSubState) bool {
	submitter, ok := playerState.(interfaces.PlayerSubmitter)
	if !ok {
		return false
	}
	return submitter.HasSubmitted()
}
