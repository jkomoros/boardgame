package blackjack

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves"
)

//boardgame:codegen
type moveShuffleDiscardToDraw struct {
	moves.FixUp
}

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
//boardgame:codegen
type moveStartRoundCleanup struct {
	moves.StartPhase
}

func (m *moveStartRoundCleanup) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.StartPhase.Legal(state, proposer); err != nil {
		return err
	}
	for _, p := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		player := p.(*playerState)
		if !player.Busted && !player.Stood {
			return errors.New("not all active players have finished their turn")
		}
	}
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
		if !player.Busted {
			player.TotalScore += player.HandValue()
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
		if player.Busted || player.Stood {
			return nil
		}
	}
	return errors.New("no active players need resetting")
}

func (m *moveResetPlayerForNewRound) Apply(state boardgame.State) error {
	_, players := concreteStates(state)
	for _, p := range players {
		p.Busted = false
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
 * moveShuffleDiscardToDraw Implementation
 *
 **************************************************/

func (m *moveShuffleDiscardToDraw) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.DrawStack.Len() > 0 {
		return errors.New("The draw stack is not yet empty")
	}

	return nil
}

func (m *moveShuffleDiscardToDraw) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)

	game.DiscardStack.MoveAllTo(game.DrawStack)
	game.DrawStack.Shuffle()

	return nil
}

/**************************************************
 *
 * moveCurrentPlayerHit Implementation
 *
 **************************************************/

func (m *moveCurrentPlayerHit) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer.EnsureValid(state)]

	if currentPlayer.Busted {
		return errors.New("Current player is busted")
	}

	if currentPlayer.HandValue() >= targetScore {
		return errors.New("Current player is already at target scores")
	}

	return nil
}

func (m *moveCurrentPlayerHit) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer.EnsureValid(state)]

	game.DrawStack.First().MoveToFirstSlot(currentPlayer.VisibleHand)

	handValue := currentPlayer.HandValue()

	if handValue > targetScore {
		currentPlayer.Busted = true
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

func (m *moveCurrentPlayerStand) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer.EnsureValid(state)]

	if currentPlayer.Busted {
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
 **************************************************/

func (m *moveRevealHiddenCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	if p.HiddenHand.NumComponents() < 1 {
		return errors.New("Target player has no cards to reveal")
	}

	return nil
}

func (m *moveRevealHiddenCard) Apply(state boardgame.State) error {
	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	p.HiddenHand.First().MoveToFirstSlot(p.VisibleHand)

	return nil
}
