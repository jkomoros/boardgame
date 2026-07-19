package moves

import (
	"errors"
	"strconv"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

// Note: these are also duplicated in server/api/storage.go
const playerToSeatRendevousDataType = "github.com/jkomoros/boardgame/server/api.PlayerToSeat"
const willSeatPlayerRendevousDataType = "github.com/jkomoros/boardgame/server/api.WillSeatPlayer"

// gameWillSeatPlayer returns true if the game will ever potentially call
// moves.SeatPlayer or not.
func gameWillSeatPlayer(state boardgame.ImmutableState) bool {
	willSeatPlayer := state.Manager().Storage().FetchInjectedDataForGame(state.Game().ID(), willSeatPlayerRendevousDataType)
	if willSeatPlayer == nil {
		return false
	}
	willSeatPlayerBool, ok := willSeatPlayer.(bool)
	if !ok {
		return false
	}
	if !willSeatPlayerBool {
		return false
	}
	return true
}

// DefaultRoundSetup returns a serial move progression appropriate for putting at
// the beginning of your RoundSetUp phase, if you have a game that uses
// SeatPlayer. It returns Optional(ActivateInactivePlayer),
// WaitForEnoughPlayers, Optional(InactivateEmptySeat). What this does is
// activate any players who have been seated already but not yet activated, then
// pause until we have enough players to activate the round, and then mark any
// unfilled seats as Inactive so the game logic within the actual round doesn't
// wait for them. This progression is finicky to get right, which is why it's
// provided as a baked function.
//
// If WithManualStart() is passed as an option, the progression will also include
// an Optional(CloseAllSeats) move (displayed as "Confirm Players") before
// WaitForEnoughPlayers. WaitForEnoughPlayers is then configured with
// WithRequireExplicitStart, meaning it will block until all unfilled seats are
// closed. This allows players to wait for more people to join before someone
// explicitly starts the game by proposing the CloseAllSeats move.
//
// In non-server contexts (tests, goldens) the explicit start requirement is
// automatically bypassed, so game logic works the same as without
// WithManualStart. When all seats are filled, the requirement is trivially
// satisfied and WaitForEnoughPlayers auto-fires normally.
func DefaultRoundSetup(auto *AutoConfigurer, options ...CustomConfigurationOption) MoveProgressionGroup {
	config := boardgame.PropertyCollection{}
	for _, opt := range options {
		opt(config)
	}
	manualStart, _ := config[configPropManualStart].(bool)

	if manualStart {
		return Serial(
			Optional(
				auto.MustConfig(
					new(ActivateInactivePlayer),
				),
			),
			Optional(
				auto.MustConfig(
					new(CloseAllSeats),
					WithMoveName("Confirm Players"),
					WithHelpText("Closes empty seats so the game can begin"),
				),
			),
			auto.MustConfig(
				new(WaitForEnoughPlayers),
				WithRequireExplicitStart(),
			),
			Optional(
				auto.MustConfig(
					new(InactivateEmptySeat),
				),
			),
		)
	}

	return Serial(
		Optional(
			auto.MustConfig(
				new(ActivateInactivePlayer),
			),
		),
		auto.MustConfig(
			new(WaitForEnoughPlayers),
		),
		Optional(
			auto.MustConfig(
				new(InactivateEmptySeat),
			),
		),
	)
}

// SeatPlayer is a game that seats a new player into an open seat in the game. It
// is a special interface point for the server library to interact with your game
// logic. The core engine has no notion of whether or not a real user is
// associated with any given player slot. The server package does distinguish
// this, keeping track of which player slots need to be filled by real users. But
// by default your core game logic can't detect which player slots haven't been
// filled, or when they are filled. SeatPlayer, when used in conjunction with
// behaviors.Seat, introduces the notion of a Seat to each player slot. Those
// properties communicate whether the seat is filled with a physical player, and
// whether it is open to having a player sit in it. SeatPlayer is a special type
// of move that will be proposed by the server engine when it has a player
// waiting to be seated. Your core game logic can decide when it should be legal
// based on which phases it is configured to be legal in. If you do not
// explicitly configure SeatPlayer (or a move that derives from it) in your game
// then the server will not alert you when a player has been seated.
//
// You may use this move directly, or embed it in a move of your own that
// overrides some logic, like for example DefaultsForState to override where the
// next player is seated.
//
// If you don't want a seat to have players seated in it, even if it's not yet
// filled, then you can call SetSeatClosed() method on the player state. The move
// CloseEmptySeats will automatically mark all currently unfilled seats as
// closed, so no new players will be accepted.
//
// For more on the concept of seats, see the package doc of boardgame/behaviors
// package.
//
//boardgame:codegen
type SeatPlayer struct {
	FixUp
	TargetPlayerIndex boardgame.PlayerIndex
}

// IsSeatPlayerMove returns true. This is a way for moves to signal to other
// libraries that it's a SeatPlayer move, even if it isn't literally this move
// struct but a subclass of it. Implements interfaces.SeatPlayerMover.
func (s *SeatPlayer) IsSeatPlayerMove() bool {
	return true
}

// the player index for the signaler, if one exists.
func (s *SeatPlayer) playerIndex(state boardgame.ImmutableState) boardgame.PlayerIndex {
	playerToSeatGeneric := state.Manager().Storage().FetchInjectedDataForGame(state.Game().ID(), playerToSeatRendevousDataType)
	if playerToSeatGeneric == nil {
		return boardgame.AdminPlayerIndex
	}
	signaler, ok := playerToSeatGeneric.(interfaces.SeatPlayerSignaler)
	if !ok {
		return boardgame.AdminPlayerIndex
	}
	return signaler.SeatIndex().EnsureValid(state)
}

// DefaultsForState sets TargetPlayerIndex to the PlayerIndex returned by
// SeatPlayerSignaler, or if that doesn't return anything, the next player who is
// neither filled nor closed.
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

// Legal verifies that TargetPlayerIndex is set to a player who is both not
// filled and not closed, and that the proposer is the admin, since only server
// should propose this move.
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

	targetPlayerIndex := s.TargetPlayerIndex.EnsureValid(state)
	if targetPlayerIndex < 0 || int(targetPlayerIndex) >= len(state.ImmutablePlayerStates()) {
		return errors.New("TargetPlayerIndex is invalid")
	}
	seat, ok := state.ImmutablePlayerStates()[targetPlayerIndex].(interfaces.Seater)
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

// Apply sets the targeted player to be Filled. If the player state also
// implements interfaces.Inactiver (for example because it implements
// behaviors.PlayerInactive), then it will also set the player to inactive. This
// is often the behavior you want; if you're in the middle of a round you
// typically don't want a new player to be active in the middle of it. But if you
// do use behaviors.PlayerInactive, remember to implement ActivateInactivePlayer
// at the beginning of rounds to activate any new seated players.
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

	targetPlayerIndex := s.TargetPlayerIndex.EnsureValid(state)
	player := state.ImmutablePlayerStates()[targetPlayerIndex]
	seat, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater")
	}
	seat.SetSeatFilled()
	if inactiver, ok := player.(interfaces.PlayerInactiver); ok {
		inactiver.SetPlayerInactive()
	}
	// Auto-assign game admin to the first seated player if the playerState
	// embeds GameAdministrator and no one is admin yet.
	if adminHolder, ok := player.(behaviors.HasGameAdministrator); ok {
		hasAdmin := false
		for _, ps := range state.ImmutablePlayerStates() {
			if behaviors.PlayerIsAdmin(ps) {
				hasAdmin = true
				break
			}
		}
		if !hasAdmin {
			adminHolder.GetGameAdministrator().SetGameAdmin()
		}
	}
	return nil
}

// ValidConfiguration checks that player states implement interfaces.Seater
func (s *SeatPlayer) ValidConfiguration(exampleState boardgame.State) error {
	player := exampleState.ImmutablePlayerStates()[0]
	_, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater. behaviors.Seat implements it for free")
	}
	return nil
}

// FallbackHelpText returns "Marks the next available seat as seated, which when
// done will mean the next player is part of the game"
func (s *SeatPlayer) FallbackHelpText() string {
	return "Marks the next available seat as seated, which when done will mean the next player is part of the game"
}

// FallbackName returns "Seat Player"
func (s *SeatPlayer) FallbackName(m *boardgame.GameManager) string {
	return "Seat Player"
}

// CloseEmptySeat is a move that will go through and repeatedly apply itself to
// close any seat that is not filled. Typically you put this at the end of a
// SetUp phase, once all of the players are there who you care to wait for, and
// want to tell the game to not try to seat any more people in them. For more on
// the notion of empty seats, see the package doc of boardgames/behaviors.
//
//boardgame:codegen
type CloseEmptySeat struct {
	FixUpMulti
	TargetPlayerIndex boardgame.PlayerIndex
}

// DefaultsForState sets TargetPlayerIndex to the next player who is currently
// marked as empty, according to interfaces.Seater.
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

// Legal verifies that TargetPlayerIndex is set to a player that is currently
// empty and not currently closed.
func (c *CloseEmptySeat) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := c.FixUpMulti.Legal(state, proposer); err != nil {
		return err
	}
	targetPlayerIndex := c.TargetPlayerIndex.EnsureValid(state)
	if targetPlayerIndex < 0 || int(targetPlayerIndex) >= len(state.ImmutablePlayerStates()) {
		return errors.New("Invalid TargetPlayerIndex")
	}
	player := state.ImmutablePlayerStates()[targetPlayerIndex]
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

// Apply sets the TargetPlayerIndex to be closed via interfaces.Seater
func (c *CloseEmptySeat) Apply(state boardgame.State) error {
	targetPlayerIndex := c.TargetPlayerIndex.EnsureValid(state)
	player := state.ImmutablePlayerStates()[targetPlayerIndex]
	seat, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater")
	}
	seat.SetSeatClosed()
	return nil
}

// ValidConfiguration checks that player states implement interfaces.Seater
func (c *CloseEmptySeat) ValidConfiguration(exampleState boardgame.State) error {
	player := exampleState.ImmutablePlayerStates()[0]
	_, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater. behaviors.Seat implements it for free")
	}
	return nil
}

// FallbackHelpText returns "Marks any empty seats as being not open for more people to be seated."
func (c *CloseEmptySeat) FallbackHelpText() string {
	return "Marks any empty seats as being not open for more people to be seated."
}

// FallbackName returns "Close Empty Seat"
func (c *CloseEmptySeat) FallbackName(m *boardgame.GameManager) string {
	return "Close Empty Seat"
}

// InactivateEmptySeat is a move that will go through and repeatedly apply itself
// to mark as closed any seat that is not filled. Typically you put this at the
// end of a SetUp phase, once all of the players are there who you care to wait
// for, and want to signal to your own game logic to not block on them being
// seated, and act like those seats aren't even there. For more on the notion of
// seats and inactive players, see the package doc of boardagme/behaviors.
//
//boardgame:codegen
type InactivateEmptySeat struct {
	FixUpMulti
	TargetPlayerIndex boardgame.PlayerIndex
}

// DefaultsForState sets TargetPlayerIndex to the next player who is currently
// marked as inactive and also empty, according to interfaces.Seater and
// interfaces.PlayerInactiver.
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

// Legal verifies that TargetPlayerIndex is set to a player that is currently
// empty and not currently inactive. If the game is running in a context where
// moves.SeatPlayer will never be called, then it will not activate for any seat
// (because it would activate for ALL seats).
func (i *InactivateEmptySeat) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := i.FixUpMulti.Legal(state, proposer); err != nil {
		return err
	}

	if !gameWillSeatPlayer(state) {
		return errors.New("Game will never seat players, so we shouldn't inactivate any, or we'd inactivate all of them")
	}

	targetPlayerIndex := i.TargetPlayerIndex.EnsureValid(state)
	if targetPlayerIndex < 0 || int(targetPlayerIndex) >= len(state.ImmutablePlayerStates()) {
		return errors.New("Invalid TargetPlayerIndex")
	}
	player := state.ImmutablePlayerStates()[targetPlayerIndex]
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

// Apply sets the TargetPlayerIndex to be inactive via interfaces.PlayerInactiver.
func (i *InactivateEmptySeat) Apply(state boardgame.State) error {
	targetPlayerIndex := i.TargetPlayerIndex.EnsureValid(state)
	player := state.ImmutablePlayerStates()[targetPlayerIndex]
	inactiver, ok := player.(interfaces.PlayerInactiver)
	if !ok {
		return errors.New("Player state didn't implement interfaces.PlayerInactiver")
	}
	inactiver.SetPlayerInactive()
	return nil
}

// ValidConfiguration checks that player states implement interfaces.Seater and
// interfaces.PlayerInactiver.
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

// FallbackHelpText returns "Marks any empty seats as being inactive, so other game logic will skip them"
func (i *InactivateEmptySeat) FallbackHelpText() string {
	return "Marks any empty seats as being inactive, so other game logic will skip them"
}

// FallbackName returns "Inactivate Empty Seat"
func (i *InactivateEmptySeat) FallbackName(m *boardgame.GameManager) string {
	return "Inactivate Empty Seat"
}

// ActivateEmptySeat is a FixUpMulti move that re-activates empty (unfilled)
// seats that were previously marked inactive by InactivateEmptySeat. This is
// the complement of InactivateEmptySeat. Unlike ActivateInactivePlayer, which
// activates ALL inactive players (both filled and empty), ActivateEmptySeat
// ONLY activates inactive players in unfilled seats. This is useful when you
// want to re-open empty seats for new players between rounds without activating
// seated-but-deliberately-inactive players.
//
//boardgame:codegen
type ActivateEmptySeat struct {
	FixUpMulti
	TargetPlayerIndex boardgame.PlayerIndex
}

// DefaultsForState sets TargetPlayerIndex to the next player who is currently
// not filled and also inactive.
func (a *ActivateEmptySeat) DefaultsForState(state boardgame.ImmutableState) {
	for i, p := range state.ImmutablePlayerStates() {
		if seat, ok := p.(interfaces.Seater); ok {
			if seat.SeatIsFilled() {
				continue
			}
			if !behaviors.PlayerIsInactive(p) {
				continue
			}
			a.TargetPlayerIndex = boardgame.PlayerIndex(i)
			return
		}
	}
}

// Legal verifies that TargetPlayerIndex is set to a player that is currently
// unfilled and inactive.
func (a *ActivateEmptySeat) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := a.FixUpMulti.Legal(state, proposer); err != nil {
		return err
	}
	targetPlayerIndex := a.TargetPlayerIndex.EnsureValid(state)
	if targetPlayerIndex < 0 || int(targetPlayerIndex) >= len(state.ImmutablePlayerStates()) {
		return errors.New("Invalid TargetPlayerIndex")
	}
	player := state.ImmutablePlayerStates()[targetPlayerIndex]
	seat, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater")
	}
	if seat.SeatIsFilled() {
		return errors.New("The selected player seat is filled; only empty seats should be activated by this move")
	}
	if !behaviors.PlayerIsInactive(player) {
		return errors.New("Player is already active. There must not be any empty inactive seats left to apply to")
	}
	return nil
}

// Apply sets the TargetPlayerIndex to be active via interfaces.PlayerInactiver.
func (a *ActivateEmptySeat) Apply(state boardgame.State) error {
	targetPlayerIndex := a.TargetPlayerIndex.EnsureValid(state)
	player := state.ImmutablePlayerStates()[targetPlayerIndex]
	inactiver, ok := player.(interfaces.PlayerInactiver)
	if !ok {
		return errors.New("Player state didn't implement interfaces.PlayerInactiver")
	}
	inactiver.SetPlayerActive()
	return nil
}

// ValidConfiguration checks that player states implement interfaces.Seater and
// interfaces.PlayerInactiver.
func (a *ActivateEmptySeat) ValidConfiguration(exampleState boardgame.State) error {
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

// FallbackHelpText returns a description of the move.
func (a *ActivateEmptySeat) FallbackHelpText() string {
	return "Re-activates any empty seats that were previously marked inactive"
}

// FallbackName returns "Activate Empty Seat"
func (a *ActivateEmptySeat) FallbackName(m *boardgame.GameManager) string {
	return "Activate Empty Seat"
}

// CloseAllSeats is a player move that closes all currently unfilled, unclosed
// seats in one action. This is useful for allowing a player to say "let's get
// started" before all seats are filled, without waiting for more players. It is
// only legal if there is at least one unfilled, unclosed seat, and if enough
// players are already seated (at least TargetCount, which defaults to
// MinNumPlayers). Unlike CloseEmptySeat (which is a FixUpMulti that closes one
// seat per application), this move closes all empty seats at once.
//
// When used via DefaultRoundSetup(auto, WithManualStart()), this move is
// configured with the name "Confirm Players" and placed in the round setup
// progression so that WaitForEnoughPlayers blocks until a player proposes it.
//
//boardgame:codegen
type CloseAllSeats struct {
	Default
}

// IsGatheringStartMove returns true, satisfying [interfaces.GatheringStartMover].
// This allows the server to include IsGatheringStart in the move form so the
// client can render a "Start Game" button without name string matching.
func (c *CloseAllSeats) IsGatheringStartMove() bool {
	return true
}

// Legal verifies that there is at least one unfilled, unclosed seat, and that
// the number of seated active players is at least TargetCount (which defaults to
// MinNumPlayers if not configured via WithTargetCount).
func (c *CloseAllSeats) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := c.Default.Legal(state, proposer); err != nil {
		return err
	}

	hasOpenSeat := false
	for _, p := range state.ImmutablePlayerStates() {
		if seat, ok := p.(interfaces.Seater); ok {
			if !seat.SeatIsFilled() && !seat.SeatIsClosed() {
				hasOpenSeat = true
				break
			}
		}
	}
	if !hasOpenSeat {
		return errors.New("No unfilled, unclosed seats to close")
	}

	targetCounter, ok := c.Info().ConcreteMove().(interfaces.TargetCounter)
	if !ok {
		return errors.New("Concrete move unexpectedly didn't implement TargetCounter")
	}

	activePlayerser, ok := state.Manager().Delegate().(numSeatedActivePlayerser)
	if !ok {
		return errors.New("Game delegate didn't implement NumSeatedActivePlayers")
	}

	targetCount := targetCounter.TargetCount(state)
	seatedPlayers := activePlayerser.NumSeatedActivePlayers(state)

	if seatedPlayers < targetCount {
		return errors.New("Only " + strconv.Itoa(seatedPlayers) + " players are seated, but at least " + strconv.Itoa(targetCount) + " are required")
	}

	// Admin check first — non-admins should see "only admin can start" rather
	// than a ReadyToStart configuration error they can't act on.
	if err := checkRequireAdmin(c.CustomConfiguration(), state, proposer); err != nil {
		return err
	}

	// Check delegate readiness to prevent the deadlock where seats close
	// before configuration is validated. This error surfaces in the client
	// via LegalForPlayerError on the move form (since CloseAllSeats is a
	// player move, not a FixUp).
	if err := state.Manager().Delegate().ReadyToStart(state); err != nil {
		return errors.New("not ready to start: " + err.Error())
	}

	return nil
}

// TargetCount returns the value that was provided via WithTargetCount. If none
// was provided, it returns your game delegate's MinNumPlayers.
func (c *CloseAllSeats) TargetCount(state boardgame.ImmutableState) int {
	config := c.CustomConfiguration()

	val, ok := config[configPropTargetCount]

	if !ok {
		if state == nil {
			return 0
		}
		return state.Manager().Delegate().MinNumPlayers()
	}

	intVal, ok := val.(int)

	if !ok {
		return -1
	}

	return intVal
}

// Apply closes all unfilled, unclosed seats.
func (c *CloseAllSeats) Apply(state boardgame.State) error {
	for _, p := range state.ImmutablePlayerStates() {
		if seat, ok := p.(interfaces.Seater); ok {
			if !seat.SeatIsFilled() && !seat.SeatIsClosed() {
				seat.SetSeatClosed()
			}
		}
	}
	return nil
}

// ValidConfiguration checks that player states implement interfaces.Seater and
// that the move implements TargetCounter.
func (c *CloseAllSeats) ValidConfiguration(exampleState boardgame.State) error {
	if err := c.Default.ValidConfiguration(exampleState); err != nil {
		return err
	}
	player := exampleState.ImmutablePlayerStates()[0]
	_, ok := player.(interfaces.Seater)
	if !ok {
		return errors.New("Player state didn't implement interfaces.Seater. behaviors.Seat implements it for free")
	}
	_, ok = c.Info().ConcreteMove().(interfaces.TargetCounter)
	if !ok {
		return errors.New("Concrete move unexpectedly didn't implement TargetCounter")
	}
	_, ok = exampleState.Manager().Delegate().(numSeatedActivePlayerser)
	if !ok {
		return errors.New("Game delegate didn't implement NumSeatedActivePlayers")
	}
	return nil
}

// FallbackHelpText returns a description of the move.
func (c *CloseAllSeats) FallbackHelpText() string {
	return "Closes all unfilled seats so no more players can join and the game can begin"
}

// FallbackName returns "Close All Seats"
func (c *CloseAllSeats) FallbackName(m *boardgame.GameManager) string {
	return "Close All Seats"
}

type numSeatedActivePlayerser interface {
	NumSeatedActivePlayers(state boardgame.ImmutableState) int
}

// WaitForEnoughPlayers is a move that is useful to include in your phase
// progressions where you want to wait until there are enough players to start a
// round. Typically the logic in your SetUpRound game phase will have an
// Optional(ActivateInactivePlayers) (if your game includes
// behaviors.InactivePlayer), then a non-optional call to this move, and then the
// rest of the logic to set up the round. This move will apply as a no-op as long
// as GameDelegate.NumSeatedActivePlayers is greater than its TargetCount. By
// default, TargetCount is your game delegate's MinNumPlayers. This move will
// auto-apply itself in contexts where SeatPlayer won't ever be called (for
// example, if you're running your game logic outside of an instance of server).
//
// If configured with WithRequireExplicitStart (typically via
// DefaultRoundSetup(auto, WithManualStart())), this move will additionally
// require that all unfilled seats are closed before it fires. This allows
// players to use CloseAllSeats to explicitly start the game rather than
// auto-starting as soon as MinNumPlayers are seated.
//
//boardgame:codegen
type WaitForEnoughPlayers struct {
	FixUp
}

// Legal verifies that the GameDelegate's NumSeatedActivePlayers is at least
// TargetCount. If configured with WithRequireExplicitStart, it additionally
// verifies that all unfilled seats are closed.
func (w *WaitForEnoughPlayers) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := w.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	if !gameWillSeatPlayer(state) {
		//We're in a context that won't ever SeatPlayer, so we should just
		//auto trigger now. The requireExplicitStart check is also skipped
		//because in non-server contexts there is no player to propose
		//CloseAllSeats.
		return nil
	}

	targetCounter, ok := w.Info().ConcreteMove().(interfaces.TargetCounter)
	if !ok {
		return errors.New("Concrete move unexpectedly didn't implement targetCounter")
	}

	activePlayerser, ok := state.Manager().Delegate().(numSeatedActivePlayerser)
	if !ok {
		return errors.New("Game delegate didn't implement NumSeatedActivePlayers")
	}

	targetCount := targetCounter.TargetCount(state)
	seatedPlayers := activePlayerser.NumSeatedActivePlayers(state)

	if seatedPlayers < targetCount {
		return errors.New("Only " + strconv.Itoa(seatedPlayers) + " are seated, but move requires at least " + strconv.Itoa(targetCount))
	}

	if w.requireExplicitStart() {
		for _, p := range state.ImmutablePlayerStates() {
			if seat, ok := p.(interfaces.Seater); ok {
				if !seat.SeatIsFilled() && !seat.SeatIsClosed() {
					return errors.New("there are still open seats; a player must confirm players before the game can start")
				}
			}
		}
	}

	// Check delegate readiness (e.g., team balance, role assignment)
	if err := state.Manager().Delegate().ReadyToStart(state); err != nil {
		return errors.New("not ready to start: " + err.Error())
	}

	return nil
}

// requireExplicitStart returns true if WithRequireExplicitStart was configured.
func (w *WaitForEnoughPlayers) requireExplicitStart() bool {
	config := w.CustomConfiguration()
	val, ok := config[configPropRequireExplicitStart]
	if !ok {
		return false
	}
	boolVal, ok := val.(bool)
	return ok && boolVal
}

// TargetCount returns the value tha twas provided via WithTargetCount. If none
// was provided, it returns your game delegate's MinNumPlayers, which is nearly
// always a reasonable default for the minimum number of players for a round.
func (w *WaitForEnoughPlayers) TargetCount(state boardgame.ImmutableState) int {
	config := w.CustomConfiguration()

	val, ok := config[configPropTargetCount]

	if !ok {
		//No configuration provided, just return default, which is MinNumPlayers
		if state == nil {
			return 0
		}
		return state.Manager().Delegate().MinNumPlayers()
	}

	intVal, ok := val.(int)

	if !ok {
		//signal error
		return -1
	}

	return intVal
}

// Apply does nothing. The main purpose of this move is to block a move
// progression from proceeding when it is not yet legal.
func (w *WaitForEnoughPlayers) Apply(state boardgame.State) error {
	return nil
}

// ValidConfiguration checks that player states implement interfaces.Seater and
// interfaces.PlayerInactiver.
func (w *WaitForEnoughPlayers) ValidConfiguration(exampleState boardgame.State) error {
	_, ok := w.Info().ConcreteMove().(interfaces.TargetCounter)
	if !ok {
		return errors.New("Concrete move doesn't implement targetCounter")
	}

	_, ok = exampleState.Manager().Delegate().(numSeatedActivePlayerser)
	if !ok {
		return errors.New("Game delegate didn't implement NumSeatedActivePlayers")
	}
	return nil
}

// FallbackHelpText returns "Waits until at least target count players are active and seated before applying itself"
func (w *WaitForEnoughPlayers) FallbackHelpText() string {
	//TODO: live target count
	return "Waits until at least target count players are active and seated before applying itself"
}

// FallbackName returns "Wait For Enough Players"
func (w *WaitForEnoughPlayers) FallbackName(m *boardgame.GameManager) string {
	return "Wait For Enough Players"
}
