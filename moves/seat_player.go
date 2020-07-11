package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

//Note: this is also duplicated in server/api/storage.go
const playerToSeatRendevousDataType = "github.com/jkomoros/boardgame/server/api.PlayerToSeat"

//SeatPlayer is a game that seats a new player into an open seat in the game. It
//is a special interface point for the server library to interact with your game
//logic. The core engine has no notion of whether or not a real user is
//associated with any given player slot. The server package does distinguish
//this, keeping track of which player slots need to be filled by real users. But
//by default your core game logic can't detect which player slots haven't been
//filled, or when they are filled. SeatPlayer, when used in conjunction with
//behaviors.Seat, introduces the notion of a Seat to each player slot. Those
//properties communicate whether the seat is filled with a physical player, and
//whether it is open to having a player sit in it. SeatPlayer is a special type
//of move that will be proposed by the server engine when it has a player
//waiting to be seated. Your core game logic can decide when it should be legal
//based on which phases it is configured to be legal in. If you do not
//explicitly configure SeatPlayer (or a move that derives from it) in your game
//then the server will not alert you when a player has been seated.
//
//You may use this move directly, or embed it in a move of your own that
//overrides some logic, like for example DefaultsForState to override where the
//next player is seated.
//
//If you don't want a seat to have players seated in it, even if it's not yet
//filled, then you can call SetSeatClosed() method on the player state. The move
//CloseEmptySeats will automatically mark all currently unfilled seats as
//closed, so no new players will be accepted.
//
//For more on the concept of seats, see the package doc of boardgame/behaviors
//package.
//
//boardgame:codegen
type SeatPlayer struct {
	FixUp
	TargetPlayerIndex boardgame.PlayerIndex
}

//IsSeatPlayerMove returns true. This is a way for moves to signal to other
//libraries that it's a SeatPlayer move, even if it isn't literally this move
//struct but a subclass of it. Implements interfaces.SeatPlayerMover.
func (s *SeatPlayer) IsSeatPlayerMove() bool {
	return true
}

//the player index for the signaler, if one exists.
func (s *SeatPlayer) playerIndex(state boardgame.ImmutableState) boardgame.PlayerIndex {
	playerToSeatGeneric := state.Manager().Storage().FetchInjectedDataForGame(state.Game().ID(), playerToSeatRendevousDataType)
	if playerToSeatGeneric == nil {
		return boardgame.AdminPlayerIndex
	}
	signaler, ok := playerToSeatGeneric.(interfaces.SeatPlayerSignaler)
	if !ok {
		return boardgame.AdminPlayerIndex
	}
	return signaler.SeatIndex()
}

//DefaultsForState sets TargetPlayerIndex to the PlayerIndex returned by
//SeatPlayerSignaler, or if that doesn't return anything, the next player who is
//neither filled nor closed.
func (s *SeatPlayer) DefaultsForState(state boardgame.ImmutableState) {

	index := s.playerIndex(state)

	if index >= 0 && int(index) < len(state.ImmutablePlayerStates()) {
		s.TargetPlayerIndex = index
		return
	}

	for i, p := range state.ImmutablePlayerStates() {
		if seat, ok := p.(interfaces.Seater); ok {
			if seat.SeatIsClosed() || seat.SeatIsFilled() {
				continue
			}
			s.TargetPlayerIndex = boardgame.PlayerIndex(i)
			return
		}
	}
}

//Legal verifies that TargetPlayerIndex is set to a player who is both not
//filled and not closed, and that the proposer is the admin, since only server
//should propose this move.
func (s *SeatPlayer) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := s.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	playerToSeatGeneric := state.Manager().Storage().FetchInjectedDataForGame(state.Game().ID(), playerToSeatRendevousDataType)
	if playerToSeatGeneric == nil {
		return errors.New("No player to seat")
	}
	_, ok := playerToSeatGeneric.(interfaces.SeatPlayerSignaler)
	if !ok {
		return errors.New("PlayerToSeat was not a SeatPlayerSignaler as expected")
	}
	if proposer != boardgame.AdminPlayerIndex {
		return errors.New("This move may only be proposed by an admin")
	}
	if s.TargetPlayerIndex < 0 || int(s.TargetPlayerIndex) >= len(state.ImmutablePlayerStates()) {
		return errors.New("TargetPlayerIndex is invalid")
	}
	seat, ok := state.ImmutablePlayerStates()[s.TargetPlayerIndex].(interfaces.Seater)
	if !ok {
		return errors.New("The selected player did not implement interfaces.Seater")
	}
	if seat.SeatIsClosed() {
		return errors.New("The selected seat was closed")
	}
	if seat.SeatIsFilled() {
		return errors.New("The selected seat was already filled")
	}
	return nil
}

//Apply sets the targeted player to be Filled. If the player state also
//implements interfaces.Inactiver (for example because it implements
//behaviors.PlayerInactive), then it will also set the player to inactive. This
//is often the behavior you want; if you're in the middle of a round you
//typically don't want a new player to be active in the middle of it. But if you
//do use behaviors.PlayerInactive, remember to implement ActivateInactivePlayer
//at the beginning of rounds to activate any new seated players.
func (s *SeatPlayer) Apply(state boardgame.State) error {

	//Make sure server will get a signal when the player is seated.
	playerToSeatGeneric := state.Manager().Storage().FetchInjectedDataForGame(state.Game().ID(), playerToSeatRendevousDataType)
	if playerToSeatGeneric == nil {
		return errors.New("No player to seat")
	}
	playerSeater, ok := playerToSeatGeneric.(interfaces.SeatPlayerSignaler)
	if !ok {
		return errors.New("PlayerToSeat was not a SeatPlayerSignaler as expected")
	}
	state.Manager().Internals().AddCommittedCallback(state, playerSeater.Committed)

	player := state.ImmutablePlayerStates()[s.TargetPlayerIndex]
	seat, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater")
	}
	seat.SetSeatFilled()
	if inactiver, ok := player.(interfaces.PlayerInactiver); ok {
		inactiver.SetPlayerInactive()
	}
	return nil
}

//ValidConfiguration checks that player states implement interfaces.Seater
func (s *SeatPlayer) ValidConfiguration(exampleState boardgame.State) error {
	player := exampleState.ImmutablePlayerStates()[0]
	_, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater. behaviors.Seat implements it for free")
	}
	return nil
}

//FallbackHelpText returns "Marks the next available seat as seated, which when
//done will mean the next player is part of the game"
func (s *SeatPlayer) FallbackHelpText() string {
	return "Marks the next available seat as seated, which when done will mean the next player is part of the game"
}

//FallbackName returns "Seat Player"
func (s *SeatPlayer) FallbackName(m *boardgame.GameManager) string {
	return "Seat Player"
}

//CloseEmptySeat is a move that will go through and repeatedly apply itself to
//close any seat that is not filled. Typically you put this at the end of a
//SetUp phase, once all of the players are there who you care to wait for, and
//want to tell the game to not try to seat any more people in them. For more on
//the notion of empty seats, see the package doc of boardgames/behaviors.
//
//boardgame:codegen
type CloseEmptySeat struct {
	FixUpMulti
	TargetPlayerIndex boardgame.PlayerIndex
}

//DefaultsForState sets TargetPlayerIndex to the next player who is currently
//marked as empty, according to interfaces.Seater.
func (c *CloseEmptySeat) DefaultsForState(state boardgame.ImmutableState) {
	for i, p := range state.ImmutablePlayerStates() {
		if seat, ok := p.(interfaces.Seater); ok {
			if !seat.SeatIsFilled() && !seat.SeatIsClosed() {
				c.TargetPlayerIndex = boardgame.PlayerIndex(i)
				return
			}
		}
	}
}

//Legal verifies that TargetPlayerIndex is set to a player that is currently
//empty and not currently closed.
func (c *CloseEmptySeat) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := c.FixUpMulti.Legal(state, proposer); err != nil {
		return err
	}
	if c.TargetPlayerIndex < 0 || int(c.TargetPlayerIndex) >= len(state.ImmutablePlayerStates()) {
		return errors.New("Invalid TargetPlayerIndex")
	}
	player := state.ImmutablePlayerStates()[c.TargetPlayerIndex]
	seat, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater")
	}
	if seat.SeatIsFilled() {
		return errors.New("The selected player seat is already filled, not empty. There must not be any seats left to apply to")
	}
	if seat.SeatIsClosed() {
		return errors.New("The selected player seat is already closed. There must not be any seats left to apply to")
	}
	return nil
}

//Apply sets the TargetPlayerIndex to be closed via interfaces.Seater
func (c *CloseEmptySeat) Apply(state boardgame.State) error {
	player := state.ImmutablePlayerStates()[c.TargetPlayerIndex]
	seat, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater")
	}
	seat.SetSeatClosed()
	return nil
}

//ValidConfiguration checks that player states implement interfaces.Seater
func (c *CloseEmptySeat) ValidConfiguration(exampleState boardgame.State) error {
	player := exampleState.ImmutablePlayerStates()[0]
	_, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater. behaviors.Seat implements it for free")
	}
	return nil
}

//FallbackHelpText returns "Marks any empty seats as being not open for more people to be seated."
func (c *CloseEmptySeat) FallbackHelpText() string {
	return "Marks any empty seats as being not open for more people to be seated."
}

//FallbackName returns "Close Empty Seat"
func (c *CloseEmptySeat) FallbackName(m *boardgame.GameManager) string {
	return "Close Empty Seat"
}

//InactivateEmptySeat is a move that will go through and repeatedly apply itself
//to mark as closed any seat that is not filled. Typically you put this at the
//end of a SetUp phase, once all of the players are there who you care to wait
//for, and want to signal to your own game logic to not block on them being
//seated, and act like those seats aren't even there. For more on the notion of
//seats and inactive players, see the package doc of boardagme/behaviors.
//
//boardgame:codegen
type InactivateEmptySeat struct {
	FixUpMulti
	TargetPlayerIndex boardgame.PlayerIndex
}

//DefaultsForState sets TargetPlayerIndex to the next player who is currently
//marked as inactive and also empty, according to interfaces.Seater and
//interfaces.PlayerInactiver.
func (i *InactivateEmptySeat) DefaultsForState(state boardgame.ImmutableState) {
	for j, p := range state.ImmutablePlayerStates() {
		if seat, ok := p.(interfaces.Seater); ok {
			if seat.SeatIsFilled() {
				continue
			}
			if behaviors.PlayerIsInactive(p) {
				continue
			}
			i.TargetPlayerIndex = boardgame.PlayerIndex(j)
			return
		}
	}
}

//Legal verifies that TargetPlayerIndex is set to a player that is currently
//empty and not currently inactive.
func (i *InactivateEmptySeat) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := i.FixUpMulti.Legal(state, proposer); err != nil {
		return err
	}
	if i.TargetPlayerIndex < 0 || int(i.TargetPlayerIndex) >= len(state.ImmutablePlayerStates()) {
		return errors.New("Invalid TargetPlayerIndex")
	}
	player := state.ImmutablePlayerStates()[i.TargetPlayerIndex]
	seat, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater")
	}
	if seat.SeatIsFilled() {
		return errors.New("The selected player seat is already filled, not empty. There must not be any seats left to apply to")
	}
	if behaviors.PlayerIsInactive(player) {
		return errors.New("Player is already inactive. There must not be any any seats left to apply to")
	}
	return nil
}

//Apply sets the TargetPlayerIndex to be inactive via interfaces.PlayerInactiver.
func (i *InactivateEmptySeat) Apply(state boardgame.State) error {
	player := state.ImmutablePlayerStates()[i.TargetPlayerIndex]
	inactiver, ok := player.(interfaces.PlayerInactiver)
	if !ok {
		return errors.New("Player state didn't implement interfaces.PlayerInactiver")
	}
	inactiver.SetPlayerInactive()
	return nil
}

//ValidConfiguration checks that player states implement interfaces.Seater and
//interfaces.PlayerInactiver.
func (i *InactivateEmptySeat) ValidConfiguration(exampleState boardgame.State) error {
	player := exampleState.ImmutablePlayerStates()[0]
	_, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater. behaviors.Seat implements it for free")
	}
	_, ok = player.(interfaces.PlayerInactiver)
	if !ok {
		return errors.New("Player state didn't implement interfaces.PlayerInactiver. behaviors.PlayerInactiveBehavior implements it for free")
	}
	return nil
}

//FallbackHelpText returns "Marks any empty seats as being inactive, so other game logic will skip them"
func (i *InactivateEmptySeat) FallbackHelpText() string {
	return "Marks any empty seats as being inactive, so other game logic will skip them"
}

//FallbackName returns "Inactivate Empty Seat"
func (i *InactivateEmptySeat) FallbackName(m *boardgame.GameManager) string {
	return "Inactivate Empty Seat"
}
