package blackjack

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//+autoreader readsetter
type MoveShuffleDiscardToDraw struct {
	boardgame.BaseMove
}

//+autoreader readsetter
type MoveAdvanceNextPlayer struct {
	boardgame.BaseMove
}

//+autoreader readsetter
type MoveDealInitialCard struct {
	boardgame.BaseMove
	TargetPlayerIndex boardgame.PlayerIndex
	IsHidden          bool
}

//+autoreader readsetter
type MoveRevealHiddenCard struct {
	boardgame.BaseMove
	TargetPlayerIndex boardgame.PlayerIndex
}

//+autoreader readsetter
type MoveCurrentPlayerHit struct {
	boardgame.BaseMove
	TargetPlayerIndex boardgame.PlayerIndex
}

//+autoreader readsetter
type MoveCurrentPlayerStand struct {
	boardgame.BaseMove
	TargetPlayerIndex boardgame.PlayerIndex
}

/**************************************************
 *
 * MoveShuffleDiscardToDraw Implementation
 *
 **************************************************/

var moveShuffleDiscardToDrawConfig = boardgame.MoveTypeConfig{
	Name:     "Shuffle Discard To Draw",
	HelpText: "When the draw deck is empty, shuffles the discard deck into draw deck.",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &MoveShuffleDiscardToDraw{
			boardgame.BaseMove{mType},
		}
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
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &MoveCurrentPlayerHit{
			BaseMove: boardgame.BaseMove{mType},
		}
	},
}

func (m *MoveCurrentPlayerHit) DefaultsForState(state boardgame.State) {
	m.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
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

/**************************************************
 *
 * MoveCurrentPlayerStand Implementation
 *
 **************************************************/

var moveCurrentPlayerStandConfig = boardgame.MoveTypeConfig{
	Name:     "Current Player Stand",
	HelpText: "If the current player no longer wants to draw cards, they can stand.",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &MoveCurrentPlayerStand{
			BaseMove: boardgame.BaseMove{mType},
		}
	},
}

func (m *MoveCurrentPlayerStand) DefaultsForState(state boardgame.State) {
	m.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
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

/**************************************************
 *
 * MoveAdvanceNextPlayer Implementation
 *
 **************************************************/

var moveAdvanceNextPlayerConfig = boardgame.MoveTypeConfig{
	Name:     "Advance Next Player",
	HelpText: "When the current player has either busted or decided to stand, we advance to next player.",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &MoveAdvanceNextPlayer{
			boardgame.BaseMove{mType},
		}
	},
	IsFixUp: true,
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

	game.CurrentPlayer = game.CurrentPlayer.Next(state)

	currentPlayer := players[game.CurrentPlayer]

	currentPlayer.Stood = false

	return nil

}

/**************************************************
 *
 * MoveRevealHiddenCard Implementation
 *
 **************************************************/

var moveRevealHiddenCardConfig = boardgame.MoveTypeConfig{
	Name:     "Reveal Hidden Card",
	HelpText: "Reveals the hidden card in the user's hand",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &MoveRevealHiddenCard{
			BaseMove: boardgame.BaseMove{mType},
		}
	},
	IsFixUp: true,
}

func (m *MoveRevealHiddenCard) DefaultsForState(state boardgame.State) {
	m.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
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

/**************************************************
 *
 * MoveDealInitialHiddenCard Implementation
 *
 **************************************************/

var moveDealInitialCardConfig = boardgame.MoveTypeConfig{
	Name:     "Deal Initial Card",
	HelpText: "Deals a card to the a player who has not gotten their initial deal",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &MoveDealInitialCard{
			BaseMove: boardgame.BaseMove{mType},
		}
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
