package base

import (
	"math"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves/interfaces"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"github.com/sirupsen/logrus"
)

//GameDelegate is a struct that implements stubs for all of GameDelegate's
//methods. This makes it easy to override just one or two methods by creating
//your own struct that anonymously embeds this one. Name,
//GameStateConstructor, PlayerStateConstructor, and ConfigureMoves are not
//implemented, since those almost certainly must be overridden for your
//particular game.
type GameDelegate struct {
	manager *boardgame.GameManager
	//the names of properties on playerStates that should be used in
	//GroupMembership.
	cachedGroupMembershipProperties []string
}

//Diagram returns the string "This should be overriden to render a reasonable state here"
func (g *GameDelegate) Diagram(state boardgame.ImmutableState) string {
	return "This should be overriden to render a reasonable state here"
}

//DisplayName by default just returns the title-case of Name() that is
//returned from the delegate in use.
func (g *GameDelegate) DisplayName() string {
	return strings.Title(g.Manager().Delegate().Name())
}

//Description defaults to "" if not overriden.
func (g *GameDelegate) Description() string {
	return ""
}

//Manager returns the manager object that was provided to SetManager.
func (g *GameDelegate) Manager() *boardgame.GameManager {
	return g.manager
}

//SetManager keeps a reference to the passed manager, and returns it when
//Manager() is called.
func (g *GameDelegate) SetManager(manager *boardgame.GameManager) {
	g.manager = manager
}

//DynamicComponentValuesConstructor returns nil, as not all games have
//DynamicComponentValues. Override this if your game does require
//DynamicComponentValues.
func (g *GameDelegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.ConfigurableSubState {
	return nil
}

//ProposeFixUpMove runs through all moves in Moves, in order, and returns the
//first one that returns true from IsFixUp and is legal at the current state. In
//many cases, this behavior should be suficient and need not be overwritten. Be
//extra sure that your FixUpMoves have a conservative Legal function, otherwise
//you could get a panic from applying too many FixUp moves. Wil emit debug
//information about why certain fixup moves didn't apply if the Manager's log
//level is Debug or higher.
func (g *GameDelegate) ProposeFixUpMove(state boardgame.ImmutableState) boardgame.Move {

	isDebug := g.Manager().Logger().Level >= logrus.DebugLevel

	var logEntry *logrus.Entry

	if isDebug {
		logEntry = g.Manager().Logger().WithFields(logrus.Fields{
			"game":    state.Game().ID(),
			"version": state.Version(),
		})
		logEntry.Debug("***** ProposeFixUpMove called *****")
	}

	for _, move := range state.Game().Moves() {

		var entry *logrus.Entry
		if isDebug {
			entry = logEntry.WithField("movetype", move.Info().Name())
		}

		if !IsFixUp(move) {
			//Not a fix up move
			continue
		}

		err := move.Legal(state, boardgame.AdminPlayerIndex)
		if err == nil {
			if isDebug {
				entry.Debug(move.Info().Name() + " : MATCH")
			}
			//Found it!
			return move
		}

		if isDebug {
			entry.Debug(move.Info().Name() + " : " + err.Error())
		}

	}
	if isDebug {
		logEntry.Debug("NO MATCH")
	}
	//No moves apply now.
	return nil
}

//CurrentPlayerIndex returns gameState.CurrentPlayer, if that is a PlayerIndex
//property. If not, returns ObserverPlayerIndex. If you use
//behaviors.CurrentPlayerBehavior it works well with this. Will use EnsureValid.
func (g *GameDelegate) CurrentPlayerIndex(state boardgame.ImmutableState) boardgame.PlayerIndex {
	index, err := state.ImmutableGameState().Reader().PlayerIndexProp("CurrentPlayer")

	if err != nil {
		//Guess that's not where they store CurrentPlayer.
		return boardgame.ObserverPlayerIndex
	}

	return index.EnsureValid(state)
}

//CurrentPhase by default with return the value of gameState.Phase, if it is
//an enum. If it is not, it will return -1 instead, to make it more clear that
//it's an invalid CurrentPhase (phase 0 is often valid).
func (g *GameDelegate) CurrentPhase(state boardgame.ImmutableState) int {

	phaseEnum, err := state.ImmutableGameState().Reader().ImmutableEnumProp("Phase")

	if err != nil {
		//Guess it wasn't there
		return -1
	}

	return phaseEnum.Value()

}

//PhaseEnum defaults to the enum named "phase" (or "Phase", if that doesn't
//exist) which is the convention for the name of the Phase enum. moves.Default
//will handle cases where that isn't a valid enum gracefully.
func (g *GameDelegate) PhaseEnum() enum.Enum {
	result := g.Manager().Chest().Enums().Enum("phase")
	if result != nil {
		return result
	}
	return g.Manager().Chest().Enums().Enum("Phase")
}

const defaultGroupsName = "group"

//GroupEnum will return the enum named 'group', if it exists, otherwise nil.
//'group' is the name of the special combine group that codegen treats specially
//and combines with boardgame.BaseGroupEnum.
func (g *GameDelegate) GroupEnum() enum.Enum {
	return g.Manager().Chest().Enums().Enum(defaultGroupsName)
}

//DistributeComponentToStarterStack does nothing any returns an error. If your
//game has components, it should override this to tell the engine where to stash
//the components to start. If your game doesn't have any components, then this
//won't be called on GameManager boot up, and this stub will have prevented you
//from needing to define a no-op.
func (g *GameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	//The stub returns an error, because if this is called that means there
	//was a component in the deck. And if we didn't store it in a stack, then
	//we are in violation of the invariant.
	return nil, errors.New("DistributeComponentToStarterStack was called, but the component was not stored in a stack")
}

//GroupMembership will look for any Enum properties on playerState, and if any
//of them are part of GroupEnum(), will return true for the values that they
//are. This handles many common cases correctly. For example, if you use
//behaviors.Color, and your color enum is combined into the enum called 'group',
//then this will automatically report that membership for the player.
func (g *GameDelegate) GroupMembership(playerState boardgame.ImmutableSubState) map[int]bool {

	//Calculating which properties to include is expensive, so only do it once.
	if playerState != nil && g.cachedGroupMembershipProperties == nil {
		//use manager.delegate to ensure we're getting any structs that embed usoupEnum := g.Manager().Delegate().GroupEnum()
		groupEnum := g.Manager().Delegate().GroupEnum()
		if groupEnum == nil {
			return nil
		}
		//Don't start as nil, so in the common case where they aren't any props,
		//we still won't regenerate this every time.
		props := make([]string, 0)
		for propName, propType := range playerState.Reader().Props() {
			if propType != boardgame.TypeEnum {
				continue
			}
			enumVal, err := playerState.Reader().ImmutableEnumProp(propName)
			if err != nil {
				continue
			}
			if enumVal.Enum().SubsetOf(groupEnum) {
				props = append(props, propName)
			}
		}
		g.cachedGroupMembershipProperties = props
	}

	if len(g.cachedGroupMembershipProperties) == 0 {
		return nil
	}
	result := make(map[int]bool, len(g.cachedGroupMembershipProperties))
	for _, propName := range g.cachedGroupMembershipProperties {
		enumVal, err := playerState.Reader().ImmutableEnumProp(propName)
		if err != nil {
			continue
		}
		result[enumVal.Value()] = true
	}
	return result
}

const computedGroupNameDelimiter = "-"

//TODO: also support 'overlapping' and 'nonoverlapping' (note the latter cant
//have a dash as that's the delimiter)
const computedGroupNameFunctionSame = "same"
const computedGroupNameFunctionsDifferent = "different"

var legalComputedGroupNameFunctions = map[string]bool{
	computedGroupNameFunctionSame:       true,
	computedGroupNameFunctionsDifferent: true,
}

//fun will be one of legalComputedGroupNameFunctions. e will not be nil, and will be known to be a subset of GroupEnum.
func doComputedGroupMembership(fun string, e enum.Enum, playerMembership, viewingAsPlayerMembership map[int]bool) bool {
	for _, key := range e.Values() {
		p := playerMembership[key]
		v := viewingAsPlayerMembership[key]
		switch fun {
		case computedGroupNameFunctionSame:
			//all of p and v must be the same
			if p != v {
				return false
			}
		case computedGroupNameFunctionsDifferent:
			//need any one key to be different
			if p != v {
				return true
			}
		}
	}
	//Default values for each
	switch fun {
	case computedGroupNameFunctionSame:
		return true
	case computedGroupNameFunctionsDifferent:
		return false
	}
	return false
}

/*
ComputedPlayerGroupMembership is the override point where advanced groups like
'same-ENUMNAME' are supported. Typically you leave this as-is without
overriding. If you override this, always fall back in the base case to returning
the value from this implementation, so you don't lose the ability to have the
special group names it provides.

The special names it supports are of the form 'TYPE-ENUMMNAME'. ENUMNAME must be
a named enum in the game's chest that is also a subset of delegate.GroupEnum.
TYPE must be one of the following types:

'same' returns true if all of the keys for that enum in playerMembership and
viewingAsPlayerMembership are the same.

'different' returns true if any of the keys for that enum in playerMembership
and viewingAsPlayerMembership are different.

Example: 'same-color': true if the two players are precisely the same color as
returned by GroupMembership.

*/
func (g *GameDelegate) ComputedPlayerGroupMembership(groupName string, playerMembership, viewingAsPlayerMembership map[int]bool) (bool, error) {

	parts := strings.Split(groupName, computedGroupNameDelimiter)

	if len(parts) == 2 {

		fun := strings.ToLower(parts[0])
		if _, ok := legalComputedGroupNameFunctions[fun]; !ok {
			return false, errors.New(parts[0] + " was used as a computed group name function but it's not a known one")
		}

		e := g.Manager().Chest().Enums().Enum(parts[1])
		if e == nil {
			return false, errors.New(parts[1] + " was used as a computed group name enum but it's not a legal enum")
		}

		//use manager.delegate to make sure we use any overriden funtions
		groupEnum := g.Manager().Delegate().GroupEnum()

		if groupEnum == nil {
			return false, errors.New("A computed group name used an enum, but there is no Group enum")
		}

		if !e.SubsetOf(groupEnum) {
			return false, errors.New(parts[1] + " enum is not a subset of GroupEnum")
		}

		return doComputedGroupMembership(fun, e, playerMembership, viewingAsPlayerMembership), nil
	}

	return false, errors.New("Unsupported group name: " + groupName)
}

//SanitizationPolicy uses struct tags to identify the right policy to apply
//(see the package doc on SanitizationPolicy for how to configure those tags).
//It sees which policies apply given the provided group membership, and then
//returns the LEAST restrictive policy that applies. This behavior is almost
//always what you want; it is rare to need to override this method.
func (g *GameDelegate) SanitizationPolicy(prop boardgame.StatePropertyRef, groupMembership map[string]bool) boardgame.Policy {

	manager := g.Manager()

	inflater := manager.Internals().StructInflater(prop)

	if inflater == nil {
		return boardgame.PolicyInvalid
	}

	policyMap := inflater.PropertySanitizationPolicy(prop.PropName)

	var applicablePolicies []int

	for groupName, isMember := range groupMembership {

		//The only ones that are in the map should be `true` but sanity check
		//just in case.
		if !isMember {
			continue
		}

		//Only if the policy is actually in the map should we use it
		if policy, ok := policyMap[groupName]; ok {
			applicablePolicies = append(applicablePolicies, int(policy))
		}
	}

	if len(applicablePolicies) == 0 {
		return boardgame.PolicyVisible
	}

	sort.Ints(applicablePolicies)

	return boardgame.Policy(applicablePolicies[0])

}

//ComputedGlobalProperties returns nil.
func (g *GameDelegate) ComputedGlobalProperties(state boardgame.ImmutableState) boardgame.PropertyCollection {
	return nil
}

//ComputedPlayerProperties returns nil.
func (g *GameDelegate) ComputedPlayerProperties(player boardgame.ImmutableSubState) boardgame.PropertyCollection {
	return nil
}

//BeginSetUp does not do anything and returns nil.
func (g *GameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
	//Don't need to do anything by default
	return nil
}

//FinishSetUp doesn't do anything and returns nil.
func (g *GameDelegate) FinishSetUp(state boardgame.State) error {
	//Don't need to do anything by default
	return nil
}

//defaultCheckGameFinishedDelegate can be private because
//DefaultGameFinished implements the methods by default.
type defaultCheckGameFinishedDelegate interface {
	GameEndConditionMet(state boardgame.ImmutableState) bool
	PlayerScore(pState boardgame.ImmutableSubState) int
	LowScoreWins() bool
}

//PlayerGameScorer is an optional interface that can be implemented by
//PlayerSubStates. If it is implemented, base.GameDelegate's default
//PlayerScore() method will return it.
type PlayerGameScorer interface {
	//Score returns the overall score for the game for the player at this
	//point in time.
	GameScore() int
}

//CheckGameFinished by default checks delegate.GameEndConditionMet(). If true,
//then it fetches delegate.PlayerScore() for each player and returns all players
//who have the highest score as winners. (If delegate.LowScoreWins() is true,
//instead of highest score, it does lowest score.) It skips any players who are
//Inactive (according to behaviors.PlayerIsInactive). To use this implementation
//simply implement those methods. This is sufficient for many games, but not
//all, so sometimes needs to be overriden.
func (g *GameDelegate) CheckGameFinished(state boardgame.ImmutableState) (finished bool, winners []boardgame.PlayerIndex) {

	if g.Manager() == nil {
		return false, nil
	}

	//Have to reach up to the manager's delegate to get the thing that embeds
	//us. Don't use the comma-ok pattern because we want to panic with
	//descriptive error if not met.
	checkGameFinished := g.Manager().Delegate().(defaultCheckGameFinishedDelegate)

	if !checkGameFinished.GameEndConditionMet(state) {
		return false, nil
	}

	lowScoreWins := checkGameFinished.LowScoreWins()

	//Game is over. What's the most extreme (max or min, depending on
	//LowScoreWins) score?
	extremeScore := 0

	if lowScoreWins {
		extremeScore = math.MaxInt32
	}

	for _, player := range state.ImmutablePlayerStates() {

		if behaviors.PlayerIsInactive(player) {
			continue
		}

		score := checkGameFinished.PlayerScore(player)

		if lowScoreWins {
			if score < extremeScore {
				extremeScore = score
			}
		} else {
			if score > extremeScore {
				extremeScore = score
			}
		}
	}

	//Who has the most extreme score score?
	for i, player := range state.ImmutablePlayerStates() {

		if behaviors.PlayerIsInactive(player) {
			continue
		}

		score := checkGameFinished.PlayerScore(player)

		if score == extremeScore {
			winners = append(winners, boardgame.PlayerIndex(i))
		}
	}

	return true, winners

}

//LowScoreWins is used in base.GameDelegate's CheckGameFinished. If false
//(default) higher scores are better. If true, however, then lower scores win
//(similar to golf), and all of the players with the lowest score win.
func (g *GameDelegate) LowScoreWins() bool {
	return false
}

//GameEndConditionMet is used in the default CheckGameFinished implementation.
//It should return true when the game is over and ready for scoring.
//CheckGameFinished uses this by default; if you override CheckGameFinished
//you don't need to override this. The default implementation of this simply
//returns false.
func (g *GameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	return false
}

//PlayerScore is used in the default CheckGameFinished implementation. It
//should return the score for the given player. CheckGameFinished uses this by
//default; if you override CheckGameFinished you don't need to override this.
//The default implementation returns pState.GameScore() (if pState implements
//the PlayerGameScorer interface), or 0 otherwise.
func (g *GameDelegate) PlayerScore(pState boardgame.ImmutableSubState) int {
	if scorer, ok := pState.(PlayerGameScorer); ok {
		return scorer.GameScore()
	}
	return 0
}

//NumSeatedActivePlayers returns the number of players who are both seated and
//active. This is typically the number you want to decide how many 'real'
//players there are at the moment. See boardgame/behaviors package doc for more.
func (g *GameDelegate) NumSeatedActivePlayers(state boardgame.ImmutableState) int {
	count := 0
	for _, p := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		if seater, ok := p.(interfaces.Seater); ok {
			if !seater.SeatIsFilled() {
				continue
			}
		}
		count++
	}
	return count
}

//NumActivePlayers returns the number of players who are active (whether or not
//they are seated). See also NumSeatedActivePlayers, which is typically what you
//want. See boardgame/behaviors package doc for more.
func (g *GameDelegate) NumActivePlayers(state boardgame.ImmutableState) int {
	count := 0
	for _, p := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		count++
	}
	return count
}

//NumSeatedPlayers returns the number of players who are seated (whether or not
//they are active.) See also NumSeatedActivePlayers, which is typically what you
//want. See boardgame/behaviors package doc for more.
func (g *GameDelegate) NumSeatedPlayers(state boardgame.ImmutableState) int {
	count := 0
	for _, p := range state.ImmutablePlayerStates() {
		if seater, ok := p.(interfaces.Seater); ok {
			if !seater.SeatIsFilled() {
				continue
			}
		}
		count++
	}
	return count
}

//DefaultNumPlayers returns 2.
func (g *GameDelegate) DefaultNumPlayers() int {
	return 2
}

//MinNumPlayers returns 1
func (g *GameDelegate) MinNumPlayers() int {
	return 1
}

//MaxNumPlayers returns 16
func (g *GameDelegate) MaxNumPlayers() int {
	return 16
}

//LegalNumPlayers checks that the number of players is between MinNumPlayers
//and MaxNumPlayers, inclusive. You'd only want to override this if some
//player numbers in that range are not legal, for example a game where only
//even numbers of players may play.
func (g *GameDelegate) LegalNumPlayers(numPlayers int) bool {

	min := g.Manager().Delegate().MinNumPlayers()
	max := g.Manager().Delegate().MaxNumPlayers()

	return numPlayers >= min && numPlayers <= max

}

//PlayerMayBeActive returns true for all players, unless they implement
//moves/interfaces.PlayerInactiverer, in which case IsInactive is consulted, and
//if it's true then this returns false. Designed to work well with behaviors.InactivePlayer
func (g *GameDelegate) PlayerMayBeActive(player boardgame.ImmutableSubState) bool {
	return !behaviors.PlayerIsInactive(player)
}

//Variants returns a VariantConfig with no entries.
func (g *GameDelegate) Variants() boardgame.VariantConfig {
	return boardgame.VariantConfig{}
}

//ConfigureAgents by default returns nil. If you want agents in your game,
//override this.
func (g *GameDelegate) ConfigureAgents() []boardgame.Agent {
	return nil
}

//ConfigureEnums simply returns nil. In general you want to override this with
//a body of `return Enums`, if you're using `boardgame-util config` to
//generate your enum set.
func (g *GameDelegate) ConfigureEnums() *enum.Set {
	return nil
}

//ConfigureDecks returns a zero-entry map. You want to override this if you
//have any components in your game (which the vast majority of games do)
func (g *GameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return make(map[string]*boardgame.Deck)
}

//ConfigureConstants returns a zero-entry map. If you have any constants you
//wa8nt to use client-side or in tag-based struct auto-inflaters, you will want
//to override this.
func (g *GameDelegate) ConfigureConstants() boardgame.PropertyCollection {
	return nil
}
