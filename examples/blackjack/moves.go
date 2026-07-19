package blackjack

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves"
)

//boardgame:codegen
type moveFinishTurn struct {
	moves.FinishTurn
}

//boardgame:codegen
type moveRevealHiddenCard struct {
	moves.CurrentPlayer
}

//boardgame:codegen
type moveCurrentPlayerHit struct {
	moves.CurrentPlayer
}

//boardgame:codegen
type moveCurrentPlayerStand struct {
	moves.CurrentPlayer
}

// moveStartRoundCleanup transitions to phaseRoundCleanup when all active
// players have either busted or stood.
//
// Declarative migration (design spec §8's second flagship acid test):
// Legal() is deleted; this move opts in via moves.WithLegalPreconditions(
// AllActivePlayers(Any(PlayerBool("Eliminated"), PlayerBool("Stood")))) in
// main.go, matching spec §8 verbatim.
//
// StartPhase embed restored (Task 7, design spec §6 A6 -- seam widened):
// Task 11 originally had to swap this move's base type from
// moves.StartPhase to moves.Default and hand-roll Apply() (see git history
// for that interim shape), because design spec §2's v1 seam only supported
// moves.Default/moves.CurrentPlayer for declarative legality and
// legalUnsupportedMovesBaseType (legal_plan.go) treated ANY other
// moves-package embed -- including moves.StartPhase, even though it has no
// Legal() override of its own -- as an unsupported base type. Task 6
// (design spec §5) widened that seam to include FixUp/FixUpMulti/
// StartPhase -- verified structurally (moves/seam_source_test.go) that none
// of them declares its own Legal() override, so their legality composes
// exactly like Default's -- so this move can now embed moves.StartPhase
// directly again, exactly as design spec §8's literal sample shows. Apply()
// is deleted: moves.StartPhase.Apply (promoted) now does the phase-set via
// PhaseToStart (configured with moves.WithPhaseToStart(phaseRoundCleanup,
// phaseEnum) in main.go, restored alongside this embed) plus
// BeforeLeavePhase/BeforeEnterPhase hook calls that are no-ops here
// (blackjack's gameState implements neither interfaces.BeforeLeavePhaser
// nor interfaces.BeforeEnterPhaser) -- the Task 11 report documented this
// exact equivalence in reverse when justifying the hand-rolled Apply() as a
// faithful stand-in for StartPhase.Apply; restoring the embed is simply
// undoing that stand-in now that it's no longer necessary. Legal() itself
// is unaffected either way: design spec §5 established that StartPhase's
// legality IS Default.Legal verbatim (no override), so the WithLegalPreconditions
// plan composes identically regardless of which of the two base types is
// embedded. Verified behavior-preserving by the pre-existing full-game
// TestGolden (JSON replay) and legal_golden_test.go both staying green. The
// move's registered Name is unaffected: moves.Default.DeriveName derives
// "Start Round Cleanup" from the Go struct name moveStartRoundCleanup by
// reflection, independent of what it embeds.
//
//boardgame:codegen
type moveStartRoundCleanup struct {
	moves.StartPhase
}

// moveAccumulateScores adds each non-busted player's hand value to their
// TotalScore. Fires once at the start of the cleanup phase.
//
//boardgame:codegen
type moveAccumulateScores struct {
	moves.FixUp
}

func (m *moveAccumulateScores) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	// Only legal if at least one player has cards in hand (scores not yet collected)
	for _, p := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		player := p.(*playerState)
		if player.Hand.NumComponents() > 0 {
			return nil
		}
	}
	return errors.New("no active players have cards to score")
}

func (m *moveAccumulateScores) Apply(state boardgame.State) error {
	for _, p := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		player := p.(*playerState)
		if !player.Eliminated {
			player.Score += player.HandValue()
		}
	}
	return nil
}

// moveCollectCards moves all cards from all players' hands back to the
// discard stack.
//
//boardgame:codegen
type moveCollectCards struct {
	moves.FixUp
}

func (m *moveCollectCards) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	for _, p := range state.ImmutablePlayerStates() {
		player := p.(*playerState)
		if player.HiddenHand.NumComponents() > 0 || player.VisibleHand.NumComponents() > 0 {
			return nil
		}
	}
	return errors.New("no player has cards to collect")
}

func (m *moveCollectCards) Apply(state boardgame.State) error {
	game, players := concreteStates(state)
	for _, p := range players {
		if err := p.HiddenHand.MoveAllTo(game.DiscardStack); err != nil {
			return err
		}
		if err := p.VisibleHand.MoveAllTo(game.DiscardStack); err != nil {
			return err
		}
	}
	return nil
}

// moveResetPlayerForNewRound resets Busted and Stood flags for all players.
//
//boardgame:codegen
type moveResetPlayerForNewRound struct {
	moves.FixUp
}

func (m *moveResetPlayerForNewRound) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	for _, p := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		player := p.(*playerState)
		if player.Eliminated || player.Stood {
			return nil
		}
	}
	return errors.New("no active players need resetting")
}

func (m *moveResetPlayerForNewRound) Apply(state boardgame.State) error {
	_, players := concreteStates(state)
	for _, p := range players {
		p.Eliminated = false
		p.Stood = false
	}
	return nil
}

// moveIncrementRoundsCompleted increments the RoundsCompleted counter.
//
//boardgame:codegen
type moveIncrementRoundsCompleted struct {
	moves.FixUp
}

func (m *moveIncrementRoundsCompleted) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	game, _ := concreteStates(state)
	// Legal only if no player has cards and flags are reset (cleanup already done)
	for _, p := range state.ImmutablePlayerStates() {
		player := p.(*playerState)
		if player.HiddenHand.NumComponents() > 0 || player.VisibleHand.NumComponents() > 0 {
			return errors.New("cards haven't been collected yet")
		}
	}
	// Check that we haven't already incremented (by looking at whether we're
	// still in the cleanup phase — the StartPhase move after us will change it)
	if game.Phase.Value() != phaseRoundCleanup {
		return errors.New("not in cleanup phase")
	}
	return nil
}

func (m *moveIncrementRoundsCompleted) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)
	game.RoundsCompleted++
	return nil
}

/**************************************************
 *
 * moveCurrentPlayerHit Implementation
 *
 * The persisted gates (not eliminated and a non-empty draw stack) are
 * declarative. HandValue arithmetic and the component MayMoveTo check remain
 * in LegalCustom because they are computed/value-level rules.
 *
 **************************************************/

func (m *moveCurrentPlayerHit) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer.EnsureValid(state)]

	if currentPlayer.HandValue() >= targetScore {
		return errors.New("Current player is already at target scores")
	}

	first := game.DrawStack.ImmutableFirst()
	return first.MayMoveTo(currentPlayer.VisibleHand)
}

func (m *moveCurrentPlayerHit) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer.EnsureValid(state)]

	game.DrawStack.First().MoveToFirstSlot(currentPlayer.VisibleHand)

	handValue := currentPlayer.HandValue()

	if handValue > targetScore {
		currentPlayer.Eliminated = true
	}

	if handValue == targetScore {
		currentPlayer.Stood = true
	}

	return nil
}

/**************************************************
 *
 * moveCurrentPlayerStand Implementation
 *
 **************************************************/

// Legal() is deliberately absent: this move opted into declarative legality.
// The original stale comment claimed moveCurrentPlayerStand could not migrate
// because both its gates are "this bool field must be FALSE" and v1's catalog
// had no negation primitive (legal.PlayerBool only passes when true). The
// completeness round shipped legal.PlayerBoolIs(prop, want), so the gates are
// now legal.PlayerBoolIs("Eliminated", false) and legal.PlayerBoolIs("Stood",
// false), added via WithLegalPreconditions in main.go's ConfigureMoves; the
// proposer/current-player check is contributed base-first by
// moves.CurrentPlayer. The declaration order is preserved within the
// field-independent bucket, so Eliminated's message wins if both are somehow
// true, matching the legacy body's top-to-bottom order. The original
// imperative body (kept only as legacyLegalMoveCurrentPlayerStand, a private
// copy in legal_golden_test.go, for golden-equivalence testing) read:
//
//	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
//		return err
//	}
//	game, players := concreteStates(state)
//	currentPlayer := players[game.CurrentPlayer.EnsureValid(state)]
//	if currentPlayer.Eliminated {
//		return errors.New("the current player has already busted")
//	}
//	if currentPlayer.Stood {
//		return errors.New("the current player already stood")
//	}
//	return nil
func (m *moveCurrentPlayerStand) Apply(state boardgame.State) error {

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer.EnsureValid(state)]

	currentPlayer.Stood = true

	return nil
}

/**************************************************
 *
 * moveRevealHiddenCard Implementation
 *
 * HiddenHand non-empty is declarative. The first component's MayMoveTo check
 * remains in LegalCustom because it examines component values at a fixed
 * position rather than a move-selected index.
 *
 **************************************************/

func (m *moveRevealHiddenCard) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	first := p.HiddenHand.ImmutableFirst()
	if err := first.MayMoveTo(p.VisibleHand); err != nil {
		return err
	}

	return nil
}

func (m *moveRevealHiddenCard) Apply(state boardgame.State) error {
	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	p.HiddenHand.First().MoveToFirstSlot(p.VisibleHand)

	return nil
}
