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

//boardgame:codegen
type moveRevealCard struct {
	moves.CurrentPlayer
	CardIndex int
}

func (m *moveRevealCard) DefaultsForState(state boardgame.ImmutableState) {

	m.CurrentPlayer.DefaultsForState(state)

	game, _ := concreteStates(state)

	for i, c := range game.HiddenCards.Components() {
		if c != nil {
			m.CardIndex = i
			break
		}
	}
}

// Legal() is deliberately absent: this move opted into declarative legality
// (design spec §8's flagship migration) via the moves.WithPreconditions call
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

func (m *moveRevealCard) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer.EnsureValid(state)]

	p.CardsLeftToReveal--
	game.HiddenCards.ComponentAt(m.CardIndex).MoveTo(game.VisibleCards, m.CardIndex)

	//If the cards are the same, the FixUpMove CaptureCards will fire after this.

	return nil
}

/**************************************************
 *
 * moveStartHideCardsTimer Implementation
 *
 **************************************************/

// moveStartHideCardsTimer stays fully imperative (spec §8's "hard-custom"
// survey, and the brief's "if NO declarative gates apply naturally, leave
// the move fully imperative/opaque and document why; do not force it"):
// none of its three checks has a catalog builder — there is no stack-COUNT
// predicate (VisibleCards.NumComponents() != 2), no Timer-state predicate
// (HideCardsTimer.Active()), and no component-VALUE-compare predicate
// (cardOneType == cardTwoType, the card-type match check). With zero
// authored WithPreconditions candidates there is nothing to opt in with —
// moves.WithPreconditions() with no natural specs would not satisfy
// boardgame's "declaring is implementing" opt-in rule anyway (an empty
// authored list is treated as not-opted-in, so CustomLegaler would never be
// consulted even if implemented). Separately, this move also embeds
// moves.FixUp, an unsupported v1-seam base type (design spec §2: only
// moves.Default and moves.CurrentPlayer support declarative legality), so it
// could not opt in even if a natural gate existed. Left byte-for-byte
// unchanged from pre-Task-11.
//
//boardgame:codegen
type moveStartHideCardsTimer struct {
	moves.FixUp
}

func (m *moveStartHideCardsTimer) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.VisibleCards.NumComponents() != 2 {
		return errors.New("there aren't two cards showing")
	}

	if game.HideCardsTimer.Active() {
		return errors.New("the timer is already active")
	}

	var revealedCards []boardgame.Component

	for _, c := range game.VisibleCards.Components() {
		if c != nil {
			revealedCards = append(revealedCards, c)
		}
	}

	cardOneType := revealedCards[0].Values().(*cardValue).Type
	cardTwoType := revealedCards[1].Values().(*cardValue).Type

	if cardOneType == cardTwoType {
		return errors.New("The two revealed cards are of the same type")
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

// moveCaptureCards is memory's other hard-custom card-type comparison (spec
// §8) — same rationale as moveStartHideCardsTimer just above: no catalog
// stack-count or component-value-compare predicate exists, so there is no
// natural WithPreconditions candidate to opt in with, and it embeds
// moves.FixUp (unsupported v1 seam base type) regardless. Left byte-for-byte
// unchanged from pre-Task-11.
//
//boardgame:codegen
type moveCaptureCards struct {
	moves.FixUp
}

func (m *moveCaptureCards) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.VisibleCards.NumComponents() != 2 {
		return errors.New("there aren't two cards showing")
	}

	var revealedCards []boardgame.Component

	for _, c := range game.VisibleCards.Components() {
		if c != nil {
			revealedCards = append(revealedCards, c)
		}
	}

	cardOneType := revealedCards[0].Values().(*cardValue).Type
	cardTwoType := revealedCards[1].Values().(*cardValue).Type

	if cardOneType != cardTwoType {
		return errors.New("The two revealed cards are not of the same type")
	}

	return nil
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
// legal.StackNotEmpty("game.VisibleCards"), added via WithPreconditions in
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
