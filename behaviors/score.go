package behaviors

import "github.com/jkomoros/boardgame"

/*
ScoreBehavior is a struct designed to be embedded anonymously in your
PlayerStates. It tracks the player's game score as a simple int, and provides a
GameScore() method that satisfies [base.PlayerGameScorer]. This means that
[base.GameDelegate]'s [base.GameDelegate.CheckGameFinished] will automatically
use this score to determine winners, with no delegate overrides needed.

It's named ScoreBehavior and not Score because otherwise it would conflict with
the internal property name when accessing it from your SubState. (Same
convention as [CurrentPlayerBehavior] and [PhaseBehavior].)

Example:

	type playerState struct {
	    base.SubState
	    behaviors.ScoreBehavior
	}

	// In a move's Apply:
	p.Score += pointsEarned

For games where score is derived from other state (e.g. counting cards in a
stack), do not use ScoreBehavior. Instead, implement GameScore() int directly on
your playerState to satisfy [base.PlayerGameScorer].
*/
type ScoreBehavior struct {
	Score int
}

// GameScore returns the current Score value. This satisfies the
// [base.PlayerGameScorer] interface, which means [base.GameDelegate]'s default
// PlayerScore() method will automatically return this value. As a result, the
// default CheckGameFinished will use this score to determine winners without
// any delegate overrides.
func (s *ScoreBehavior) GameScore() int {
	return s.Score
}

// PlayerGameScore is a convenience function that returns the GameScore for the
// given player state, if it implements GameScore() int. Returns (0, false) if
// the player state does not implement the interface. This parallels
// [PlayerIsInactive] and PlayerHasSubmitted.
func PlayerGameScore(playerState boardgame.ImmutableSubState) (int, bool) {
	type gameScorer interface {
		GameScore() int
	}
	if scorer, ok := playerState.(gameScorer); ok {
		return scorer.GameScore(), true
	}
	return 0, false
}
