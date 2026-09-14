package memory

import (
	"errors"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

const hideCardsDuration = 4 * time.Second

/**************************************************
 *
 * moveRevealCard Implementation
 *
 **************************************************/

// moveRevealCard is moves.MoveComponentToSlot -- "move the card at CardIndex
// out of HiddenCards and into the same slot of VisibleCards" -- plus this
// game's own bookkeeping. main.go's ConfigureMoves names both stacks and
// declares the mirrored layout; the verb discovers CardIndex as the slot field
// because it is the move's only int property.
//
//boardgame:codegen
type moveRevealCard struct {
	moves.MoveComponentToSlot
	CardIndex int
}

func (m *moveRevealCard) DefaultsForState(state boardgame.ImmutableState) {

	m.MoveComponentToSlot.DefaultsForState(state)

	game, _ := concreteStates(state)

	for i, c := range game.HiddenCards.Components() {
		if c != nil {
			m.CardIndex = i
			break
		}
	}
}

// Legal() is deliberately absent: this move opted into declarative legality
// (design spec §8's flagship migration) via the moves.WithLegalPreconditions call
// in main.go's ConfigureMoves. moves.CurrentPlayer.Legal (promoted, since
// this type no longer overrides it) calls moves.Default.Legal, which detects
// the assembled plan and evaluates it instead of running the old imperative
// chain below (kept only as legacyLegalMoveRevealCard, a private copy in
// legal_golden_test.go, for golden-equivalence testing):
//
//	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
//		return err
//	}
//	game, players := concreteStates(state)
//	p := players[game.CurrentPlayer.EnsureValid(state)]
//	if p.CardsLeftToReveal < 1 {
//		return errors.New("You have no cards left to reveal this turn")
//	}
//	c := game.HiddenCards.ImmutableComponentAt(m.CardIndex)
//	if c == nil {
//		if game.VisibleCards.ImmutableComponentAt(m.CardIndex) == nil {
//			return errors.New("there is no card at that index")
//		}
//		return errors.New("that card has already been revealed")
//	}
//	return c.MayMoveToSlot(game.VisibleCards, m.CardIndex)

// Apply super-calls moves.MoveComponentToSlot.Apply for the transfer itself,
// then does this game's own bookkeeping.
func (m *moveRevealCard) Apply(state boardgame.State) error {

	if err := m.MoveComponentToSlot.Apply(state); err != nil {
		return err
	}

	game, players := concreteStates(state)

	players[game.CurrentPlayer.EnsureValid(state)].CardsLeftToReveal--

	//If the cards are the same, the FixUpMove CaptureCards will fire after this.

	return nil
}

/**************************************************
 *
 * moveStartHideCardsTimer Implementation
 *
 **************************************************/

// VisibleCards==2 and the static Type mismatch are declarative. Only the
// timer's runtime state remains in LegalCustom.
//
//boardgame:codegen
type moveStartHideCardsTimer struct {
	moves.FixUp
}

func (m *moveStartHideCardsTimer) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	game, _ := concreteStates(state)

	if game.HideCardsTimer.Active() {
		return errors.New("the timer is already active")
	}

	return nil
}

func (m *moveStartHideCardsTimer) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)

	move := state.Game().MoveByName(hideCardMoveName)

	game.HideCardsTimer.Start(hideCardsDuration, move)

	return nil
}

/**************************************************
 *
 * moveCaptureCards Implementation
 *
 **************************************************/

// VisibleCards==2 and the two static Type values matching are declarative.
//
//boardgame:codegen
type moveCaptureCards struct {
	moves.FixUp
}

func (m *moveCaptureCards) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	for i, c := range game.VisibleCards.Components() {
		if c != nil {
			game.VisibleCards.ComponentAt(i).MoveToNextSlot(p.WonCards)
		}
	}

	return nil
}

/**************************************************
 *
 * moveHideCards Implementation
 *
 **************************************************/

//boardgame:codegen
type moveHideCards struct {
	moves.CurrentPlayer
}

// Legal() is deliberately absent: this move opted into declarative legality.
// The original stale comment (and TUTORIAL.md) used moveHideCards as the
// canonical "logic the declarative catalog can't express" example, on the
// theory that its two gates — CardsLeftToReveal must be non-positive, and at
// least one card must be showing — had no catalog builders. The completeness
// round falsified that: the gates are now
// legal.PropCompare("player.CardsLeftToReveal", "<=", 0) and
// legal.StackNotEmpty("game.VisibleCards"), added via WithLegalPreconditions in
// main.go's ConfigureMoves (Workstream 9 re-migration); the
// proposer/current-player check is contributed base-first by
// moves.CurrentPlayer (this is a normal player move — no fixup memo, every
// proposer cell recomputes fresh). The original imperative body (kept only as
// legacyLegalMoveHideCards, a private copy in legal_golden_test.go, for
// golden-equivalence testing) read:
//
//	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
//		return err
//	}
//	game, players := concreteStates(state)
//	p := players[game.CurrentPlayer.EnsureValid(state)]
//	if p.CardsLeftToReveal > 0 {
//		return errors.New("You still have to reveal more cards before your turn is over")
//	}
//	if game.VisibleCards.NumComponents() < 1 {
//		return errors.New("no cards left to hide")
//	}
//	return nil
func (m *moveHideCards) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)

	//Cancel a timer in case it was still going.
	game.HideCardsTimer.Cancel()

	for i, c := range game.VisibleCards.Components() {
		if c != nil {
			if err := c.MoveTo(game.HiddenCards, i); err != nil {
				return errors.New("Couldn't move component: " + err.Error())
			}
		}
	}

	return nil
}
