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
// Legal() is deleted; this move opts in via moves.WithPreconditions(
// AllActivePlayers(Any(PlayerBool("Eliminated"), PlayerBool("Stood")))) in
// main.go, matching spec §8 verbatim.
//
// Base-type deviation from the literal spec sample (documented): spec §8's
// code shows this move still embedding moves.StartPhase. It cannot: design
// spec §2's v1 seam only supports moves.Default/moves.CurrentPlayer for
// declarative legality, and legalUnsupportedMovesBaseType (legal_plan.go)
// treats ANY other moves-package embed -- including moves.StartPhase, even
// though it has no Legal() override of its own -- as an unsupported base
// type; this is locked in by an existing, passing test
// (moves.TestUnsupportedBaseTypeStartPhaseIsBootError). Rather than leave
// this move un-migrated (spec §8's own genuinely-natural AllActivePlayers
// gate would then go undelivered) or weaken that framework-level seam
// restriction (out of Task 11's scope -- design spec §10 defers "seam
// expansion" to later work), this move now embeds moves.Default directly
// and hand-rolls the one piece of behavior moves.StartPhase provided that
// this move actually used: setting the game's current phase in Apply (see
// below). moves.StartPhase's other machinery -- generic PhaseToStart()
// config lookup, BeforeLeavePhase/BeforeEnterPhase hooks -- is unused here
// (blackjack's gameState implements neither hook interface), so hand-
// rolling costs nothing. moves.WithIsFixUp(true) is now required explicitly
// in main.go (moves.Default defaults IsFixUp to false; moves.FixUp/
// moves.StartPhase default it to true, which this move relied on
// implicitly before). The move's registered Name is unaffected: moves.
// Default.DeriveName derives "Start Round Cleanup" from the Go struct name
// moveStartRoundCleanup by reflection, independent of what it embeds.
//
//boardgame:codegen
type moveStartRoundCleanup struct {
	moves.Default
}

// Apply sets the game's current phase to phaseRoundCleanup -- the one
// behavior this move used from moves.StartPhase.Apply before the base-type
// migration documented above. blackjack's gameState implements neither
// interfaces.BeforeLeavePhaser nor interfaces.BeforeEnterPhaser, so
// StartPhase.Apply's hook-calling logic had no effect for this move and is
// not replicated.
func (m *moveStartRoundCleanup) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)
	game.SetCurrentPhase(phaseRoundCleanup)
	return nil
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
		p.HiddenHand.MoveAllTo(game.DiscardStack)
		p.VisibleHand.MoveAllTo(game.DiscardStack)
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
 * Stays fully imperative (spec §8's "hand-value arithmetic... stay in
 * LegalCustom" survey, applied honestly per the Task 11 brief's "if NO
 * declarative gates apply naturally, leave the move fully imperative/opaque
 * and document why; do not force it"): moves.CurrentPlayer is a supported
 * v1-seam base, but none of this move's OWN checks has a catalog builder --
 * "Eliminated must be false" and the HandValue() comparison are both
 * negations/computed-value checks the catalog has no primitive for (no
 * "NOT playerBool", and HandValue() is a delegate-computed player property,
 * not a stored field reachable via the player.* path grammar), and the
 * DrawStack-has-a-card-at-index-0 / MayMoveTo checks need a move.* idxField
 * naming a literal 0, which no field on this move provides without adding
 * one purely to satisfy the catalog API (forcing it). With zero natural
 * WithPreconditions candidates, opting in just to reach LegalCustom would
 * violate boardgame's "declaring is implementing" rule for nothing (an
 * empty authored spec list is treated as not-opted-in regardless). Left
 * byte-for-byte unchanged from pre-Task-11.
 *
 **************************************************/

func (m *moveCurrentPlayerHit) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer.EnsureValid(state)]

	if currentPlayer.Eliminated {
		return errors.New("Current player is busted")
	}

	if currentPlayer.HandValue() >= targetScore {
		return errors.New("Current player is already at target scores")
	}

	first := game.DrawStack.ImmutableFirst()
	if first == nil {
		return errors.New("No cards left in draw stack")
	}

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
 * Stays fully imperative, same rationale as moveCurrentPlayerHit above: both
 * checks are "this bool field must be FALSE", and the catalog has no
 * negation primitive (legal.PlayerBool only passes when true). No natural
 * WithPreconditions candidate; not forced.
 *
 **************************************************/

func (m *moveCurrentPlayerStand) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer.EnsureValid(state)]

	if currentPlayer.Eliminated {
		return errors.New("the current player has already busted")
	}

	if currentPlayer.Stood {
		return errors.New("the current player already stood")
	}

	return nil

}

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
 * Stays fully imperative, same rationale family as moveCurrentPlayerHit/
 * Stand above: both checks need a component-presence/MayMoveTo predicate
 * anchored at index 0 ("first"), which the catalog only supports via a
 * move.* idxField -- and this move has no field naming a literal 0 without
 * adding one purely to satisfy the catalog API. No natural WithPreconditions
 * candidate; not forced.
 *
 **************************************************/

func (m *moveRevealHiddenCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	first := p.HiddenHand.ImmutableFirst()
	if first == nil {
		return errors.New("Target player has no cards to reveal")
	}

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
