package blackjack

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//+autoreader readsetter
type MoveShuffleDiscardToDraw struct {
	moves.Base
}

//+autoreader readsetter
type MoveFinishTurn struct {
	moves.FinishTurn
}

//+autoreader readsetter
type MoveDealInitialCard struct {
	moves.Base
	TargetPlayerIndex boardgame.PlayerIndex
	IsHidden          bool
}

//+autoreader readsetter
type MoveRevealHiddenCard struct {
	moves.CurrentPlayer
}

//+autoreader readsetter
type MoveCurrentPlayerHit struct {
	moves.CurrentPlayer
}

//+autoreader readsetter
type MoveCurrentPlayerStand struct {
	moves.CurrentPlayer
}

/**************************************************
 *
 * MoveShuffleDiscardToDraw Implementation
 *
 **************************************************/

var moveShuffleDiscardToDrawConfig = boardgame.MoveTypeConfig{
	Name:     "Shuffle Discard To Draw",
	HelpText: "When the draw deck is empty, shuffles the discard deck into draw deck.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveShuffleDiscardToDraw)
	},
	IsFixUp: true,
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

/**************************************************
 *
 * MoveCurrentPlayerHit Implementation
 *
 **************************************************/

var moveCurrentPlayerHitConfig = boardgame.MoveTypeConfig{
	Name:     "Current Player Hit",
	HelpText: "The current player hits, drawing a card.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveCurrentPlayerHit)
	},
}

func (m *MoveCurrentPlayerHit) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

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

func (m *MoveCurrentPlayerHit) Apply(state boardgame.MutableState) error {
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

/**************************************************
 *
 * MoveCurrentPlayerStand Implementation
 *
 **************************************************/

var moveCurrentPlayerStandConfig = boardgame.MoveTypeConfig{
	Name:     "Current Player Stand",
	HelpText: "If the current player no longer wants to draw cards, they can stand.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveCurrentPlayerStand)
	},
}

func (m *MoveCurrentPlayerStand) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

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

/**************************************************
 *
 * MoveAdvanceNextPlayer Implementation
 *
 **************************************************/

var moveFinishTurnConfig = boardgame.MoveTypeConfig{
	Name:     "Finish Turn",
	HelpText: "When the current player has either busted or decided to stand, we advance to next player.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveFinishTurn)
	},
	IsFixUp: true,
}

/**************************************************
 *
 * MoveRevealHiddenCard Implementation
 *
 **************************************************/

var moveRevealHiddenCardConfig = boardgame.MoveTypeConfig{
	Name:     "Reveal Hidden Card",
	HelpText: "Reveals the hidden card in the user's hand",
	MoveConstructor: func() boardgame.Move {
		return new(MoveRevealHiddenCard)
	},
	IsFixUp: true,
}

func (m *MoveRevealHiddenCard) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

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

func (m *MoveRevealHiddenCard) Apply(state boardgame.MutableState) error {
	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	p.HiddenHand.MoveComponent(boardgame.FirstComponentIndex, p.VisibleHand, boardgame.FirstSlotIndex)

	return nil
}

/**************************************************
 *
 * MoveDealInitialHiddenCard Implementation
 *
 **************************************************/

var moveDealInitialCardConfig = boardgame.MoveTypeConfig{
	Name:     "Deal Initial Card",
	HelpText: "Deals a card to the a player who has not gotten their initial deal",
	MoveConstructor: func() boardgame.Move {
		return new(MoveDealInitialCard)
	},
	IsFixUp: true,
}

func (m *MoveDealInitialCard) DefaultsForState(state boardgame.State) {
	_, players := concreteStates(state)

	//First look for the first player with no hidden card dealt
	for i := 0; i < len(players); i++ {
		p := players[i]
		if p.HiddenHand.NumComponents() == 0 {
			m.TargetPlayerIndex = boardgame.PlayerIndex(i)
			m.IsHidden = true
			return
		}
	}
	//OK, hidden hands were full. Anyone who hasn't had the other card now gets it.
	for i := 0; i < len(players); i++ {
		p := players[i]
		if !p.GotInitialDeal {
			m.TargetPlayerIndex = boardgame.PlayerIndex(i)
			m.IsHidden = false
			return
		}
	}
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
