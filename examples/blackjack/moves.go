package blackjack

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

type MoveShuffleDiscardToDraw struct{}

type MoveAdvanceNextPlayer struct{}

type MoveDealInitialCard struct {
	TargetPlayerIndex boardgame.PlayerIndex
	IsHidden          bool
}

type MoveRevealHiddenCard struct {
	TargetPlayerIndex boardgame.PlayerIndex
}

type MoveCurrentPlayerHit struct {
	TargetPlayerIndex boardgame.PlayerIndex
}

type MoveCurrentPlayerStand struct {
	TargetPlayerIndex boardgame.PlayerIndex
}

/**************************************************
 *
 * MoveShuffleDiscardToDraw Implementation
 *
 **************************************************/

func MoveShuffleDiscardToDrawFactory(state boardgame.State) boardgame.Move {
	return &MoveShuffleDiscardToDraw{}
}

func (m *MoveShuffleDiscardToDraw) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, _ := concreteStates(state)

	if game.DrawStack.Len() > 0 {
		return errors.New("The draw stack is not yet empty")
	}

	return nil
}

func (m *MoveShuffleDiscardToDraw) Apply(state boardgame.MutableState) error {
	game, _ := concreteStates(state)

	game.DiscardStack.MoveAllTo(game.DrawStack)
	game.DrawStack.Shuffle()

	return nil
}

func (m *MoveShuffleDiscardToDraw) Name() string {
	return "Shuffle Discard To Draw"
}

func (m *MoveShuffleDiscardToDraw) Description() string {
	return "When the draw deck is empty, shuffles the discard deck into draw deck."
}

func (m *MoveShuffleDiscardToDraw) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

func (t *MoveShuffleDiscardToDraw) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}

/**************************************************
 *
 * MoveCurrentPlayerHit Implementation
 *
 **************************************************/

func MoveCurrentPlayerHitFactory(state boardgame.State) boardgame.Move {
	result := &MoveCurrentPlayerHit{}

	if state != nil {
		result.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
	}

	return result
}

func (m *MoveCurrentPlayerHit) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	game, players := concreteStates(state)

	if !m.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("The proposing player is not who the move is acting on behalf of.")
	}

	if game.CurrentPlayer != m.TargetPlayerIndex {
		return errors.New("The specified player is not the current player.")
	}

	currentPlayer := players[game.CurrentPlayer]

	if currentPlayer.Busted {
		return errors.New("Current player is busted")
	}

	handValue, _ := state.Computed().Player(currentPlayer.PlayerIndex()).Reader().IntProp("HandValue")

	if handValue >= targetScore {
		return errors.New("Current player is already at target scores")
	}

	return nil
}

func (m *MoveCurrentPlayerHit) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	game.DrawStack.MoveComponent(boardgame.FirstComponentIndex, currentPlayer.VisibleHand, boardgame.FirstSlotIndex)

	handValue, _ := state.Computed().Player(currentPlayer.PlayerIndex()).Reader().IntProp("HandValue")

	if handValue > targetScore {
		currentPlayer.Busted = true
	}

	if handValue == targetScore {
		currentPlayer.Stood = true
	}

	return nil
}

func (m *MoveCurrentPlayerHit) Name() string {
	return "Current Player Hit"
}

func (m *MoveCurrentPlayerHit) Description() string {
	return "The current player hits, drawing a card."
}

func (m *MoveCurrentPlayerHit) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

func (t *MoveCurrentPlayerHit) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}

/**************************************************
 *
 * MoveCurrentPlayerStand Implementation
 *
 **************************************************/

func MoveCurrentPlayerStandFactory(state boardgame.State) boardgame.Move {
	result := &MoveCurrentPlayerStand{}

	if state != nil {
		result.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
	}

	return result
}

func (m *MoveCurrentPlayerStand) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	game, players := concreteStates(state)

	if !m.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("The proposing player is not who the move is on behalf of.")
	}

	if game.CurrentPlayer != m.TargetPlayerIndex {
		return errors.New("The specified player is not the current player.")
	}

	currentPlayer := players[game.CurrentPlayer]

	if currentPlayer.Busted {
		return errors.New("The current player has already busted.")
	}

	if currentPlayer.Stood {
		return errors.New("The current player already stood.")
	}

	return nil

}

func (m *MoveCurrentPlayerStand) Apply(state boardgame.MutableState) error {

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	currentPlayer.Stood = true

	return nil
}

func (m *MoveCurrentPlayerStand) Name() string {
	return "Current Player Stand"
}

func (m *MoveCurrentPlayerStand) Description() string {
	return "If the current player no longer wants to draw cards, they can stand."
}

func (m *MoveCurrentPlayerStand) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

func (t *MoveCurrentPlayerStand) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}

/**************************************************
 *
 * MoveAdvanceNextPlayer Implementation
 *
 **************************************************/

func MoveAdvanceNextPlayerFactory(state boardgame.State) boardgame.Move {
	return &MoveAdvanceNextPlayer{}
}

func (m *MoveAdvanceNextPlayer) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	if currentPlayer.Busted || currentPlayer.Stood {
		return nil
	}

	return errors.New("The current player has neither busted nor decided to stand.")
}

func (m *MoveAdvanceNextPlayer) Apply(state boardgame.MutableState) error {

	game, players := concreteStates(state)

	game.CurrentPlayer.Next(state)

	currentPlayer := players[game.CurrentPlayer]

	currentPlayer.Stood = false

	return nil

}

func (m *MoveAdvanceNextPlayer) Name() string {
	return "Advance Next Player"
}

func (m *MoveAdvanceNextPlayer) Description() string {
	return "When the current player has either busted or decided to stand, we advance to next player."
}

func (m *MoveAdvanceNextPlayer) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

func (t *MoveAdvanceNextPlayer) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}

/**************************************************
 *
 * MoveRevealHiddenCard Implementation
 *
 **************************************************/

func MoveRevealHiddenCardFactory(state boardgame.State) boardgame.Move {
	result := &MoveRevealHiddenCard{}

	if state != nil {
		result.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
	}

	return result
}

func (m *MoveRevealHiddenCard) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	game, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	if !m.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("The proposing player is not the player the move is on behalf of.")
	}

	if !m.TargetPlayerIndex.Equivalent(game.CurrentPlayer) {
		return errors.New("The target player is not the current player.")
	}

	if p.HiddenHand.NumComponents() < 1 {
		return errors.New("Target player has no cards to reveal")
	}

	return nil
}

func (m *MoveRevealHiddenCard) Apply(state boardgame.MutableState) error {
	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	p.HiddenHand.MoveComponent(boardgame.FirstComponentIndex, p.VisibleHand, boardgame.FirstSlotIndex)

	return nil
}

func (m *MoveRevealHiddenCard) Name() string {
	return "Reveal Hidden Card"
}

func (m *MoveRevealHiddenCard) Description() string {
	return "Reveals the hidden card in the user's hand"
}

func (m *MoveRevealHiddenCard) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

func (t *MoveRevealHiddenCard) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}

/**************************************************
 *
 * MoveDealInitialHiddenCard Implementation
 *
 **************************************************/

func MoveDealInitialCardFactory(state boardgame.State) boardgame.Move {
	result := &MoveDealInitialCard{}

	if state != nil {
		_, players := concreteStates(state)

		//First look for the first player with no hidden card dealt
		for i := 0; i < len(players); i++ {
			p := players[i]
			if p.HiddenHand.NumComponents() == 0 {
				result.TargetPlayerIndex = boardgame.PlayerIndex(i)
				result.IsHidden = true
				return result
			}
		}
		//OK, hidden hands were full. Anyone who hasn't had the other card now gets it.
		for i := 0; i < len(players); i++ {
			p := players[i]
			if !p.GotInitialDeal {
				result.TargetPlayerIndex = boardgame.PlayerIndex(i)
				result.IsHidden = false
				return result
			}
		}
	}

	return result
}

func (m *MoveDealInitialCard) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	_, players := concreteStates(state)

	if !m.TargetPlayerIndex.Valid(state) {
		return errors.New("Invalid target player index")
	}

	p := players[m.TargetPlayerIndex]

	if p.GotInitialDeal {
		return errors.New("The target player already got their initial deal")
	}

	if p.HiddenHand.NumComponents() == 1 && m.IsHidden {
		return errors.New("We were supposed to deal the hidden card, but the hidden hand was already dealt")
	}

	if p.HiddenHand.NumComponents() == 0 && !m.IsHidden {
		return errors.New("We were told to deal to the non-hidden card even though hidden hand was empty")
	}

	return nil

}

func (m *MoveDealInitialCard) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	if m.IsHidden {

		if err := game.DrawStack.MoveComponent(boardgame.FirstComponentIndex, p.HiddenHand, boardgame.NextSlotIndex); err != nil {
			return err
		}

	} else {

		if err := game.DrawStack.MoveComponent(boardgame.FirstComponentIndex, p.VisibleHand, boardgame.NextSlotIndex); err != nil {
			return err
		}

		//This completes their initial deal
		p.GotInitialDeal = true
	}

	return nil

}

func (m *MoveDealInitialCard) Name() string {
	return "Deal Initial Card"
}

func (m *MoveDealInitialCard) Description() string {
	return "Deals a card to the a player who has not gotten their initial deal"
}

func (m *MoveDealInitialCard) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

func (t *MoveDealInitialCard) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}
