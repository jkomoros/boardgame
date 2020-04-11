package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

//ActivateInactivePlayer is a fixup move that is designed to be used to activate
//any players who are currently inactive, running repeatedly until all of them
//are activated. Typically you explicitly mark this move as legal whenever you
//want players to be able to actually "join" the logic of a game. Designed to be
//used with behaviors.PlayerInactive. If that behavior is used, then when a new
//player is "seated" once a game starts, by default they'll be marked as
//Inactive. That means that the game will act as though they don't exist until
//they're explicitly marked as active. Typically if a player is seated in the
//middle of a round, they won't actually be included in play until the round is
//over. If you use that behavior, you should explicitly include this move as
//legal whenever you want players to be reactivated. Typically you would have it
//either ALWAYS be legal (which means, as soon as a player is seated they should
//be actively included in play), or have a specific phase when this move
//activates, for example at the end of the SetUpForNextRound phase in your game.
//Designed to be used on its own directly. For more on inactive players, see the
//package doc of boardgame/behaviors.
//
//boardgame:codegen
type ActivateInactivePlayer struct {
	FixUpMulti
	TargetPlayerIndex boardgame.PlayerIndex
}

//DefaultsForState sets TargetPlayerIndex to the next player who is currently
//marked as inactive, according to interfaces.PlayerInactiver.
func (a *ActivateInactivePlayer) DefaultsForState(state boardgame.ImmutableState) {
	for i, p := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsInactive(p) {
			a.TargetPlayerIndex = boardgame.PlayerIndex(i)
			return
		}
	}
}

//Legal verifies that TargetPlayerIndex is set to a player whose InActive returns true.
func (a *ActivateInactivePlayer) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := a.FixUpMulti.Legal(state, proposer); err != nil {
		return err
	}
	if a.TargetPlayerIndex < 0 || int(a.TargetPlayerIndex) >= len(state.ImmutablePlayerStates()) {
		return errors.New("Invalid TargetPlayerIndex")
	}
	player := state.ImmutablePlayerStates()[a.TargetPlayerIndex]
	if !behaviors.PlayerIsInactive(player) {
		return errors.New("The selected player is not inactive; there must be no inactive players to activate")
	}
	return nil
}

//Apply sets the TargetPlayerIndex to be active via SetPlayerActive.
func (a *ActivateInactivePlayer) Apply(state boardgame.State) error {
	player := state.ImmutablePlayerStates()[a.TargetPlayerIndex]
	inactiver, ok := player.(interfaces.PlayerInactiver)
	if !ok {
		return errors.New("Player state didn't implement interfaces.PlayerInactiver")
	}
	inactiver.SetPlayerActive()
	return nil
}

//ValidConfiguration checks that player states implement interfaces.PlayerInactiver
func (a *ActivateInactivePlayer) ValidConfiguration(exampleState boardgame.State) error {
	player := exampleState.ImmutablePlayerStates()[0]
	_, ok := player.(interfaces.PlayerInactiver)
	if !ok {
		return errors.New("Player state didn't implement interfaces.PlayerInactiver. behaviors.PlayerInactive implements it for free")
	}
	return nil
}

//FallbackHelpText returns "Activates any players who are not currently active."
func (a *ActivateInactivePlayer) FallbackHelpText() string {
	return "Activates any players who are not currently active."
}

//FallbackName returns "Activate Inactive Players"
func (a *ActivateInactivePlayer) FallbackName() string {
	return "Activate Inactive Players"
}
