package blackjack

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

type MoveShuffleDiscardToDraw struct{}

type MoveAdvanceNextPlayer struct{}

type MoveDealInitialCard struct {
	TargetPlayerIndex int
	IsHidden          bool
}

type MoveRevealHiddenCard struct {
	TargetPlayerIndex int
}

type MoveCurrentPlayerHit struct {
	TargetPlayerIndex int
}

type MoveCurrentPlayerStand struct {
	TargetPlayerIndex int
}

/**************************************************
 *
 * MoveShuffleDiscardToDraw Implementation
 *
 **************************************************/

func (m *MoveShuffleDiscardToDraw) Legal(state *boardgame.State) error {
	game, _ := concreteStates(state)

	if game.DrawStack.Len() > 0 {
		return errors.New("The draw stack is not yet empty")
	}

	return nil
}

func (m *MoveShuffleDiscardToDraw) Apply(state *boardgame.State) error {
	game, _ := concreteStates(state)

	game.DiscardStack.MoveAllTo(game.DrawStack)
	game.DrawStack.Shuffle()

	return nil
}

func (m *MoveShuffleDiscardToDraw) Copy() boardgame.Move {
	var result MoveShuffleDiscardToDraw
	result = *m
	return &result
}

func (m *MoveShuffleDiscardToDraw) DefaultsForState(state *boardgame.State) {
	//Nothing to do
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

/**************************************************
 *
 * MoveCurrentPlayerHit Implementation
 *
 **************************************************/

func (m *MoveCurrentPlayerHit) Legal(state *boardgame.State) error {

	game, players := concreteStates(state)

	if game.CurrentPlayer != m.TargetPlayerIndex {
		return errors.New("The specified player is not the current player.")
	}

	currentPlayer := players[game.CurrentPlayer]

	if currentPlayer.Busted {
		return errors.New("Current player is busted")
	}

	handValue, _ := state.Computed().Player(currentPlayer.PlayerIndex()).IntProp("HandValue")

	if handValue >= targetScore {
		return errors.New("Current player is already at target scores")
	}

	return nil
}

func (m *MoveCurrentPlayerHit) Apply(state *boardgame.State) error {
	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	game.DrawStack.MoveComponent(boardgame.FirstComponentIndex, currentPlayer.VisibleHand, boardgame.FirstSlotIndex)

	handValue, _ := state.Computed().Player(currentPlayer.PlayerIndex()).IntProp("HandValue")

	if handValue > targetScore {
		currentPlayer.Busted = true
	}

	if handValue == targetScore {
		currentPlayer.Stood = true
	}

	return nil
}

func (m *MoveCurrentPlayerHit) Copy() boardgame.Move {
	var result MoveCurrentPlayerHit
	result = *m
	return &result
}

func (m *MoveCurrentPlayerHit) DefaultsForState(state *boardgame.State) {
	game, _ := concreteStates(state)

	m.TargetPlayerIndex = game.CurrentPlayer
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

/**************************************************
 *
 * MoveCurrentPlayerStand Implementation
 *
 **************************************************/

func (m *MoveCurrentPlayerStand) Legal(state *boardgame.State) error {

	game, players := concreteStates(state)

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

func (m *MoveCurrentPlayerStand) Apply(state *boardgame.State) error {

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	currentPlayer.Stood = true

	return nil
}

func (m *MoveCurrentPlayerStand) Copy() boardgame.Move {
	var result MoveCurrentPlayerStand
	result = *m
	return &result
}

func (m *MoveCurrentPlayerStand) DefaultsForState(state *boardgame.State) {
	game, _ := concreteStates(state)
	m.TargetPlayerIndex = game.CurrentPlayer
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

/**************************************************
 *
 * MoveAdvanceNextPlayer Implementation
 *
 **************************************************/

func (m *MoveAdvanceNextPlayer) Legal(state *boardgame.State) error {
	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	if currentPlayer.Busted || currentPlayer.Stood {
		return nil
	}

	return errors.New("The current player has neither busted nor decided to stand.")
}

func (m *MoveAdvanceNextPlayer) Apply(state *boardgame.State) error {

	game, players := concreteStates(state)

	game.CurrentPlayer++
	if game.CurrentPlayer >= len(players) {
		game.CurrentPlayer = 0
	}

	currentPlayer := players[game.CurrentPlayer]

	currentPlayer.Stood = false

	return nil

}

func (m *MoveAdvanceNextPlayer) Copy() boardgame.Move {
	var result MoveAdvanceNextPlayer
	result = *m
	return &result
}

func (m *MoveAdvanceNextPlayer) DefaultsForState(state *boardgame.State) {
	//TODO: implement
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

/**************************************************
 *
 * MoveRevealHiddenCard Implementation
 *
 **************************************************/

func (m *MoveRevealHiddenCard) Legal(state *boardgame.State) error {

	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	if p.HiddenHand.NumComponents() < 1 {
		return errors.New("Target player has no cards to reveal")
	}

	return nil
}

func (m *MoveRevealHiddenCard) Apply(state *boardgame.State) error {
	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	p.HiddenHand.MoveComponent(boardgame.FirstComponentIndex, p.VisibleHand, boardgame.FirstSlotIndex)

	return nil
}

func (m *MoveRevealHiddenCard) Copy() boardgame.Move {
	var result MoveRevealHiddenCard
	result = *m
	return &result
}

func (m *MoveRevealHiddenCard) DefaultsForState(state *boardgame.State) {
	game, _ := concreteStates(state)

	m.TargetPlayerIndex = game.CurrentPlayer

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

/**************************************************
 *
 * MoveDealInitialHiddenCard Implementation
 *
 **************************************************/

func (m *MoveDealInitialCard) Legal(state *boardgame.State) error {
	_, players := concreteStates(state)

	if m.TargetPlayerIndex < 0 || m.TargetPlayerIndex >= len(players) {
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

func (m *MoveDealInitialCard) Apply(state *boardgame.State) error {
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

func (m *MoveDealInitialCard) Copy() boardgame.Move {
	var result MoveDealInitialCard
	result = *m
	return &result
}

func (m *MoveDealInitialCard) DefaultsForState(state *boardgame.State) {

	//The default game delegate will cycle around calling this, so
	//DefaultsForState should pick the next one each time.

	_, players := concreteStates(state)

	//First look for the first player with no hidden card dealt
	for i := 0; i < len(players); i++ {
		p := players[i]
		if p.HiddenHand.NumComponents() == 0 {
			m.TargetPlayerIndex = i
			m.IsHidden = true
			return
		}
	}
	//OK, hidden hands were full. Anyone who hasn't had the other card now gets it.
	for i := 0; i < len(players); i++ {
		p := players[i]
		if !p.GotInitialDeal {
			m.TargetPlayerIndex = i
			m.IsHidden = false
			return
		}
	}

	return
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
