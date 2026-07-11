package debuganimations

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

/*
Declarative-legality survey (design spec §8, Task 12): every move type in
this file embeds moves.Default (a SUPPORTED v1-seam base type — design spec
§2), so none is blocked by the base-type seam the way checkers'
movePlaceToken or werewolf's three move types are. Despite that, ALL ELEVEN
move types below stay fully imperative. Two independent catalog gaps are
responsible, and finding them here (a game explicitly built to exercise
every stack-shuffling/animation code path) is this task's single most
concrete piece of design feedback — reported in full in the Task 12 report:

 1. NO STACK-SIZE/COUNT PREDICATE EXISTS. Every one of this file's Legal()
    bodies gates on stack.NumComponents() thresholds or exact counts
    (game.FanStack.NumComponents() > 1, game.AllVisibleStack.NumComponents()
    < 1, etc.) — the direct stack-shaped analog of legal.PropAtLeast/
    legal.PropCompare, which only read INT-typed PROPERTIES, not a stack's
    length. LegalFacetCount (boardgame/legal_types.go) exists in the type
    system specifically for this ("a stack-size check needs only the count
    facet" — design spec §1), but v1's catalog (legal/catalog_*.go) ships
    zero predicates that construct a Read with it; grep confirms
    LegalFacetCount is referenced only in its own declaration and doc
    comments. This single missing primitive is the ENTIRE blocker for
    moveShuffleHidden, moveVisibleShuffleCards, and moveShuffleCards (each
    ONE simple threshold check — textbook PropAtLeast shape, just on a
    stack instead of an int property) and moveStartMoveAllComponentsToHidden
    /moveStartMoveAllComponentsToVisible (each TWO such checks, which
    WithPreconditions' implicit AND already composes correctly) — five
    moves that would become FULLY declarative, Legal()-deleted migrations
    (spec §8's own flagship shape) the moment a legal.StackSizeAtLeast/
    legal.StackSizeCompare constructor exists.

 2. NO NEGATION OR AND-GROUPING-WITHIN-OR COMPOSITOR EXISTS. "any" is v1's
    only compositor (an OR of leaf verdicts — design spec §1's anti-tarpit
    rules); WithPreconditions' own top-level list is an implicit AND, but
    there is no way to express "(A and B) or (C and D)" (a fixed pair of
    known-good toggle states — moveMoveCardBetweenFanStacks,
    moveMoveBetweenHidden, moveMoveToken, moveMoveTokenSanitized) or a
    negation ("NOT both stacks occupied" — moveFlipHiddenCard's XOR
    occupancy check; pig's moveCountDie/moveDoneTurn hit this same gap for
    a negated bool, Task 12's pig commit). These four moves would still be
    hard-custom even with gap #1 fixed.

Two moves are blocked by neither gap, but by something the catalog was
simply never built to express at all: moveMoveCardBetweenShortStacks and
moveMoveCardBetweenDrawAndDiscardStacks choose WHICH stack is source and
which is destination based on a move-field bool (m.FromFirst/m.FromDraw) —
the catalog's path grammar (boardgame/legal_path.go's parseLegalPath) has no
conditional/indirect path resolution, only fixed "kind.Property" literals
known at WithPreconditions-authoring time.

Every move below is therefore left byte-for-byte unchanged from
pre-Task-12; no golden test file is added since no Legal() is touched.
*/

//boardgame:codegen
type moveMoveCardBetweenShortStacks struct {
	moves.Default
	FromFirst bool
}

//boardgame:codegen
type moveMoveCardBetweenDrawAndDiscardStacks struct {
	moves.Default
	FromDraw bool
}

//boardgame:codegen
type moveFlipHiddenCard struct {
	moves.Default
}

//boardgame:codegen
type moveMoveCardBetweenFanStacks struct {
	moves.Default
}

//boardgame:codegen
type moveVisibleShuffleCards struct {
	moves.Default
}

//boardgame:codegen
type moveShuffleCards struct {
	moves.Default
}

//boardgame:codegen
type moveMoveBetweenHidden struct {
	moves.Default
}

//boardgame:codegen
type moveMoveToken struct {
	moves.Default
}

//boardgame:codegen
type moveMoveTokenSanitized struct {
	moves.Default
}

//boardgame:codegen
type moveStartMoveAllComponentsToHidden struct {
	moves.Default
}

//boardgame:codegen
type moveStartMoveAllComponentsToVisible struct {
	moves.Default
}

//boardgame:codegen
type moveShuffleHidden struct {
	moves.Default
}

/**************************************************
 *
 * moveMoveCardBetweenShortStacks Implementation
 *
 **************************************************/

func (m *moveMoveCardBetweenShortStacks) HelpText() string {
	return "Moves a card between two short stacks"
}

func (m *moveMoveCardBetweenShortStacks) DefaultsForState(state boardgame.ImmutableState) {
	gameState, _ := concreteStates(state)

	if gameState.FirstShortStack.NumComponents() < 1 {
		m.FromFirst = false
	} else {
		m.FromFirst = true
	}
}

func (m *moveMoveCardBetweenShortStacks) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	var from boardgame.Stack
	var to boardgame.Stack
	if m.FromFirst {
		from = game.FirstShortStack
		to = game.SecondShortStack
	} else {
		from = game.SecondShortStack
		to = game.FirstShortStack
	}

	first := from.ImmutableFirst()
	if first == nil {
		return errors.New("source stack has no cards to move")
	}

	return first.MayMoveTo(to)
}

func (m *moveMoveCardBetweenShortStacks) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	from := game.SecondShortStack
	to := game.FirstShortStack

	if m.FromFirst {
		from = game.FirstShortStack
		to = game.SecondShortStack
	}

	if err := from.First().MoveToFirstSlot(to); err != nil {
		return err
	}

	return nil
}

/**************************************************
 *
 * moveMoveCardBetweenDrawAndDiscardStacks Implementation
 *
 **************************************************/

func (m *moveMoveCardBetweenDrawAndDiscardStacks) HelpText() string {
	return "Moves a card between draw and discard stacks"
}

func (m *moveMoveCardBetweenDrawAndDiscardStacks) DefaultsForState(state boardgame.ImmutableState) {
	gameState, _ := concreteStates(state)

	if gameState.DiscardStack.NumComponents() < 3 {
		m.FromDraw = true
	} else {
		m.FromDraw = false
	}
}

func (m *moveMoveCardBetweenDrawAndDiscardStacks) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	var from boardgame.Stack
	var to boardgame.Stack
	if m.FromDraw {
		from = game.DrawStack
		to = game.DiscardStack
	} else {
		from = game.DiscardStack
		to = game.DrawStack
	}

	first := from.ImmutableFirst()
	if first == nil {
		return errors.New("source stack has no cards to move")
	}

	return first.MayMoveTo(to)
}

func (m *moveMoveCardBetweenDrawAndDiscardStacks) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	from := game.DiscardStack
	to := game.DrawStack

	if m.FromDraw {
		from = game.DrawStack
		to = game.DiscardStack
	}

	if err := from.First().MoveToFirstSlot(to); err != nil {
		return err
	}

	return nil
}

/**************************************************
 *
 * moveFlipHiddenCard Implementation
 *
 **************************************************/

func (m *moveFlipHiddenCard) HelpText() string {
	return "Flips the card between hidden and revealed"
}

func (m *moveFlipHiddenCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.HiddenCard.NumComponents() < 1 && game.VisibleCard.NumComponents() < 1 {
		return errors.New("Neither the HiddenCard nor RevealedCard is set")
	}

	if game.HiddenCard.NumComponents() > 0 && game.VisibleCard.NumComponents() > 0 {
		return errors.New("both hidden and revealed are full")
	}

	return nil
}

func (m *moveFlipHiddenCard) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	from := game.VisibleCard
	to := game.HiddenCard

	if game.HiddenCard.NumComponents() > 0 {
		from = game.HiddenCard
		to = game.VisibleCard
	}

	if err := from.First().MoveToFirstSlot(to); err != nil {
		return err
	}

	return nil
}

/**************************************************
 *
 * moveMoveCardBetweenFanStacks Implementation
 *
 **************************************************/

func (m *moveMoveCardBetweenFanStacks) HelpText() string {
	return "Moves a card from or to Fan and Fan Discard"
}

func (m *moveMoveCardBetweenFanStacks) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() == 6 && game.FanDiscard.NumComponents() == 3 {
		return nil
	}

	if game.FanStack.NumComponents() == 5 && game.FanDiscard.NumComponents() == 4 {
		return nil
	}

	return errors.New("Fan stacks aren't in known toggle state")
}

func (m *moveMoveCardBetweenFanStacks) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() < 6 {
		return game.FanDiscard.First().MoveTo(game.FanStack, 2)
	}

	return game.FanStack.ComponentAt(2).MoveToFirstSlot(game.FanDiscard)
}

/**************************************************
 *
 * moveVisibleShuffleCards Implementation
 *
 **************************************************/

func (m *moveVisibleShuffleCards) HelpText() string {
	return "Performs a visible shuffle"
}

func (m *moveVisibleShuffleCards) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() > 1 {
		return nil
	}

	return errors.New("Aren't enough cards to shuffle")
}

func (m *moveVisibleShuffleCards) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	return game.FanStack.PublicShuffle()

}

/**************************************************
 *
 * moveShuffleCards Implementation
 *
 **************************************************/

func (m *moveShuffleCards) HelpText() string {
	return "Performs a secret shuffle"
}

func (m *moveShuffleCards) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() > 1 {
		return nil
	}

	return errors.New("Aren't enough cards to shuffle")
}

func (m *moveShuffleCards) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	return game.FanStack.Shuffle()

}

/**************************************************
 *
 * moveMoveBetweenHidden Implementation
 *
 **************************************************/

func (m *moveMoveBetweenHidden) HelpText() string {
	return "Moves between hidden and visible stacks"
}

func (m *moveMoveBetweenHidden) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.VisibleStack.NumComponents() == 5 && game.HiddenStack.NumComponents() == 4 {
		return nil
	}

	if game.VisibleStack.NumComponents() == 4 && game.HiddenStack.NumComponents() == 5 {
		return nil
	}

	return errors.New("Cards aren't in known position")
}

func (m *moveMoveBetweenHidden) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.VisibleStack.NumComponents() < 5 {
		return game.HiddenStack.First().MoveTo(game.VisibleStack, 2)
	}

	return game.VisibleStack.ComponentAt(2).MoveToFirstSlot(game.HiddenStack)

}

/**************************************************
 *
 * moveMoveToken Implementation
 *
 **************************************************/

func (m *moveMoveToken) HelpText() string {
	return "Moves tokens"
}

func (m *moveMoveToken) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.TokensFrom.NumComponents() == 10 && game.TokensTo.NumComponents() == 9 {
		return nil
	}

	if game.TokensFrom.NumComponents() == 9 && game.TokensTo.NumComponents() == 10 {
		return nil
	}

	return errors.New("tokens aren't in known position")
}

func (m *moveMoveToken) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.TokensFrom.NumComponents() < 10 {
		return game.TokensTo.First().MoveTo(game.TokensFrom, 2)
	}

	return game.TokensFrom.ComponentAt(2).MoveToFirstSlot(game.TokensTo)

}

/**************************************************
 *
 * moveMoveTokenSanitized Implementation
 *
 **************************************************/

func (m *moveMoveTokenSanitized) HelpText() string {
	return "Moves tokens"
}

func (m *moveMoveTokenSanitized) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.SanitizedTokensFrom.NumComponents() == 10 && game.SanitizedTokensTo.NumComponents() == 9 {
		return nil
	}

	if game.SanitizedTokensFrom.NumComponents() == 9 && game.SanitizedTokensTo.NumComponents() == 10 {
		return nil
	}

	return errors.New("tokens aren't in known position")
}

func (m *moveMoveTokenSanitized) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.SanitizedTokensFrom.NumComponents() < 10 {
		return game.SanitizedTokensTo.First().MoveTo(game.SanitizedTokensFrom, 2)
	}

	return game.SanitizedTokensFrom.ComponentAt(2).MoveToFirstSlot(game.SanitizedTokensTo)

}

/**************************************************
 *
 * moveStartMoveAllComponentsToHidden Implementation
 *
 **************************************************/

func (m *moveStartMoveAllComponentsToHidden) HelpText() string {
	return "Moves all components from visible to hidden"
}

func (m *moveStartMoveAllComponentsToHidden) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game := state.ImmutableGameState().(*gameState)

	if game.AllVisibleStack.NumComponents() < 1 {
		return errors.New("No components in visible stack to move")
	}

	if game.AllHiddenStack.NumComponents() > 0 {
		return errors.New("The hidden stack already has items. Use the 'To Visible' move")
	}

	return nil
}

func (m *moveStartMoveAllComponentsToHidden) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	for game.AllVisibleStack.NumComponents() > 0 {
		if err := game.AllVisibleStack.First().MoveToNextSlot(game.AllHiddenStack); err != nil {
			return err
		}
	}

	return nil
}

/**************************************************
 *
 * moveStartMoveAllComponentsToVisible Implementation
 *
 **************************************************/

func (m *moveStartMoveAllComponentsToVisible) HelpText() string {
	return "Moves all components from hidden to visible"
}

func (m *moveStartMoveAllComponentsToVisible) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game := state.ImmutableGameState().(*gameState)

	if game.AllHiddenStack.NumComponents() < 1 {
		return errors.New("No components in hidden stack to move")
	}

	if game.AllVisibleStack.NumComponents() > 0 {
		return errors.New("The visible stack already has items. Use the 'To Hidden' move")
	}

	return nil
}

func (m *moveStartMoveAllComponentsToVisible) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	for game.AllHiddenStack.NumComponents() > 0 {
		if err := game.AllHiddenStack.First().MoveToNextSlot(game.AllVisibleStack); err != nil {
			return err
		}
	}

	return nil
}

/**************************************************
 *
 * moveShuffleHidden Implementation
 *
 **************************************************/

func (m *moveShuffleHidden) HelpText() string {
	return "Shuffles the fan discard pile and increments the shuffle count"
}

func (m *moveShuffleHidden) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game := state.ImmutableGameState().(*gameState)

	if game.FanDiscard.NumComponents() < 1 {
		return errors.New("FanDiscard has no cards to shuffle")
	}

	return nil
}

func (m *moveShuffleHidden) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if err := game.FanDiscard.Shuffle(); err != nil {
		return err
	}

	game.FanShuffleCount++

	return nil
}
