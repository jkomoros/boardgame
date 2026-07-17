package base

import (
	"math"
	"strings"

	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/constraints"
	"github.com/jkomoros/boardgame/moves/interfaces"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"github.com/sirupsen/logrus"
)

// GameDelegate is a struct that implements stubs for all of GameDelegate's
// methods. This makes it easy to override just one or two methods by creating
// your own struct that anonymously embeds this one. Name,
// GameStateConstructor, PlayerStateConstructor, and ConfigureMoves are not
// implemented, since those almost certainly must be overridden for your
// particular game.
type GameDelegate struct {
	manager *boardgame.GameManager
	//the names of properties on playerStates that should be used in
	//GroupMembership.
	cachedGroupMembershipProperties []string
}

// Diagram returns the string "This should be overriden to render a reasonable state here"
func (g *GameDelegate) Diagram(state boardgame.ImmutableState) string {
	return "This should be overriden to render a reasonable state here"
}

// DisplayName by default just returns the title-case of Name() that is
// returned from the delegate in use.
func (g *GameDelegate) DisplayName() string {
	return strings.Title(g.Manager().Delegate().Name())
}

// Description defaults to "" if not overriden.
func (g *GameDelegate) Description() string {
	return ""
}

// Manager returns the manager object that was provided to SetManager.
func (g *GameDelegate) Manager() *boardgame.GameManager {
	return g.manager
}

// SetManager keeps a reference to the passed manager, and returns it when
// Manager() is called.
func (g *GameDelegate) SetManager(manager *boardgame.GameManager) {
	g.manager = manager
}

// DynamicComponentValuesConstructor returns nil, as not all games have
// DynamicComponentValues. Override this if your game does require
// DynamicComponentValues.
func (g *GameDelegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.ConfigurableSubState {
	return nil
}

// ProposeFixUpMove runs through all moves in Moves, in order, and returns the
// first one that returns true from IsFixUp and is legal at the current state. In
// many cases, this behavior should be suficient and need not be overwritten. Be
// extra sure that your FixUpMoves have a conservative Legal function, otherwise
// you could get a panic from applying too many FixUp moves. Wil emit debug
// information about why certain fixup moves didn't apply if the Manager's log
// level is Debug or higher.
//
// Iterates GameManager.CandidateMoves(state) rather than every configured
// move type (design spec §5's "Phase bucketing" engine win, #640): moves
// declaratively impossible in state's current phase are skipped with zero
// Legal() evaluations. This is the v1 integration point for that index —
// the loop itself, and its first-match-wins semantics, are unchanged; only
// the candidate list feeding it is now pre-filtered. An opaque (non-opted-in)
// move is always a candidate (the superset property), so this is a
// zero-behavior-change for any move that hasn't adopted declarative
// legality.
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

	for _, move := range g.Manager().CandidateMoves(state) {

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

// CurrentPlayerIndex returns gameState.CurrentPlayer, if that is a PlayerIndex
// property. If not, returns ObserverPlayerIndex. If you use
// behaviors.CurrentPlayerBehavior it works well with this. Will use EnsureValid.
func (g *GameDelegate) CurrentPlayerIndex(state boardgame.ImmutableState) boardgame.PlayerIndex {
	index, err := state.ImmutableGameState().Reader().PlayerIndexProp("CurrentPlayer")

	if err != nil {
		//Guess that's not where they store CurrentPlayer.
		return boardgame.ObserverPlayerIndex
	}

	return index.EnsureValid(state)
}

// CurrentPhase by default returns the ImmutableVal for gameState.Phase. If the
// Phase property doesn't exist or isn't an enum, it returns nil.
func (g *GameDelegate) CurrentPhase(state boardgame.ImmutableState) enum.ImmutableVal {

	phaseVal, err := state.ImmutableGameState().Reader().ImmutableEnumProp("Phase")

	if err != nil {
		//Guess it wasn't there
		return nil
	}

	return phaseVal

}

// PhaseEnum defaults to the enum named "phase" (or "Phase", if that doesn't
// exist) which is the convention for the name of the Phase enum. moves.Default
// will handle cases where that isn't a valid enum gracefully.
func (g *GameDelegate) PhaseEnum() enum.Enum {
	result := g.Manager().Chest().Enums().Enum("phase")
	if result != nil {
		return result
	}
	return g.Manager().Chest().Enums().Enum("Phase")
}

const defaultGroupsName = "group"

// GroupEnum will return the enum named 'group', if it exists, otherwise nil.
// 'group' is the name of the special combine group that codegen treats specially
// and combines with boardgame.BaseGroupEnum.
func (g *GameDelegate) GroupEnum() enum.Enum {
	return g.Manager().Chest().Enums().Enum(defaultGroupsName)
}

// DistributeComponentToStarterStack does nothing any returns an error. If your
// game has components, it should override this to tell the engine where to stash
// the components to start. If your game doesn't have any components, then this
// won't be called on GameManager boot up, and this stub will have prevented you
// from needing to define a no-op.
func (g *GameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	//The stub returns an error, because if this is called that means there
	//was a component in the deck. And if we didn't store it in a stack, then
	//we are in violation of the invariant.
	return nil, errors.New("DistributeComponentToStarterStack was called, but the component was not stored in a stack")
}

// GroupMembership will look for any Enum properties on playerState, and if any
// of them are part of GroupEnum(), will return true for the values that they
// are. This handles many common cases correctly. For example, if you use
// behaviors.Color, and your color enum is combined into the enum called 'group',
// then this will automatically report that membership for the player.
func (g *GameDelegate) GroupMembership(playerState boardgame.ImmutableSubState) enum.ImmutableMembershipSet {

	//Calculating which properties to include is expensive, so only do it once.
	if playerState != nil && g.cachedGroupMembershipProperties == nil {
		//use manager.delegate to ensure we're getting any structs that embed us
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
	groupEnum := g.Manager().Delegate().GroupEnum()
	members := make([]enum.EnumKey, 0, len(g.cachedGroupMembershipProperties))
	for _, propName := range g.cachedGroupMembershipProperties {
		enumVal, err := playerState.Reader().ImmutableEnumProp(propName)
		if err != nil {
			continue
		}
		members = append(members, enumVal.Value())
	}
	return groupEnum.NewMembershipSet(members...)
}

const computedGroupNameDelimiter = "-"

// TODO: also support 'overlapping' and 'nonoverlapping' (note the latter cant
// have a dash as that's the delimiter)
const computedGroupNameFunctionSame = "same"
const computedGroupNameFunctionsDifferent = "different"

var legalComputedGroupNameFunctions = map[string]bool{
	computedGroupNameFunctionSame:       true,
	computedGroupNameFunctionsDifferent: true,
}

// fun will be one of legalComputedGroupNameFunctions. e will not be nil, and will be known to be a subset of GroupEnum.
func doComputedGroupMembership(fun string, e enum.Enum, playerMembership, viewingAsPlayerMembership enum.ImmutableMembershipSet) bool {
	for _, key := range e.Values() {
		p := playerMembership != nil && playerMembership.Contains(key)
		v := viewingAsPlayerMembership != nil && viewingAsPlayerMembership.Contains(key)
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
func (g *GameDelegate) ComputedPlayerGroupMembership(groupName string, playerMembership, viewingAsPlayerMembership enum.ImmutableMembershipSet) (bool, error) {

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

// SanitizationPolicy uses struct tags to identify the right policy to apply
// (see the package doc on SanitizationPolicy for how to configure those tags).
// It sees which policies apply given the provided group membership, and then
// returns the LEAST restrictive policy that applies. This behavior is almost
// always what you want; it is rare to need to override this method.
func (g *GameDelegate) SanitizationPolicy(prop boardgame.StatePropertyRef, groupMembership map[string]bool) boardgame.Policy {

	manager := g.Manager()

	inflater := manager.Internals().StructInflater(prop)

	if inflater == nil {
		return boardgame.PolicyInvalid
	}

	policyMap := inflater.PropertySanitizationPolicy(prop.PropName)

	return boardgame.ResolveSanitizationPolicy(policyMap, groupMembership, boardgame.PolicyVisible)

}

// CustomPlayerOrder returns the custom player order from the gameState if it
// implements moves/interfaces.PlayerOrderer (e.g. by embedding
// behaviors.PlayerOrderBehavior). Returns nil otherwise, meaning default
// sequential order.
func (g *GameDelegate) CustomPlayerOrder(state boardgame.ImmutableState) []boardgame.PlayerIndex {
	if orderer, ok := state.ImmutableGameState().(interfaces.PlayerOrderer); ok {
		return orderer.PlayerOrder()
	}
	return nil
}

// FrameworkComputedGlobalProperties returns framework-owned defaults. It
// "PlayerOrder" ([]int) when a PlayerOrderBehavior is embedded in GameState.
// Game authors add values through ConfigureComputedProperties instead of
// overriding this method.
func (g *GameDelegate) FrameworkComputedGlobalProperties(state boardgame.ImmutableState) boardgame.PropertyCollection {
	result := boardgame.PropertyCollection{}

	if order := g.Manager().Delegate().CustomPlayerOrder(state); order != nil {
		intOrder := make([]int, len(order))
		for i, idx := range order {
			intOrder[i] = int(idx)
		}
		result["PlayerOrder"] = intOrder
	}

	// Gathering: available enum values for client pickers.
	// Only include values for enums whose corresponding behavior is actually
	// embedded in the player state. This avoids injecting AvailableColors
	// into games like checkers that have a "color" enum for component
	// ownership but don't use the gathering selection system.
	examplePlayer := state.ImmutablePlayerStates()[0]
	chest := g.Manager().Chest()
	if _, ok := examplePlayer.(behaviors.HasPlayerTeam); ok {
		if teamEnum := chest.Enums().Enum("team"); teamEnum != nil {
			result["AvailableTeams"] = enumValuesForClient(teamEnum)
		}
	}
	if _, ok := examplePlayer.(behaviors.HasPlayerRole); ok {
		if roleEnum := chest.Enums().Enum("role"); roleEnum != nil {
			result["AvailableRoles"] = enumValuesForClient(roleEnum)
		}
	}
	if _, ok := examplePlayer.(behaviors.HasPlayerColor); ok {
		if colorEnum := chest.Enums().Enum("color"); colorEnum != nil {
			result["AvailableColors"] = colorEnumValuesForClient(colorEnum)
		}
	}

	// Gathering: readiness error for client display. FixUp move errors are
	// invisible to the client, so this is the primary delivery path for
	// ReadyToStart error messages. Only computed when the delegate's
	// ReadyToStart is overridden (returns non-nil for some states), to avoid
	// calling it on every state change during normal gameplay where the
	// default (nil) adds no value.
	if err := g.Manager().Delegate().ReadyToStart(state); err != nil {
		result["ReadyToStartError"] = err.Error()
	}

	return result
}

// enumValuesForClient returns a list of {Key, Name} objects for all values in
// the enum, suitable for serialization to the client.
func enumValuesForClient(e enum.Enum) []map[string]interface{} {
	var result []map[string]interface{}
	for _, key := range e.Values() {
		result = append(result, map[string]interface{}{
			"Key":  int(key),
			"Name": e.String(key),
		})
	}
	return result
}

// colorEnumValuesForClient is like enumValuesForClient but also includes CSS
// color strings from behaviors.CSSColorForKey, so the client picker can render
// accurate color swatches without a hardcoded mapping.
func colorEnumValuesForClient(e enum.Enum) []map[string]interface{} {
	var result []map[string]interface{}
	for _, key := range e.Values() {
		entry := map[string]interface{}{
			"Key":  int(key),
			"Name": e.String(key),
		}
		if css, ok := behaviors.CSSColorForKey[key]; ok {
			entry["CSSColor"] = css
		}
		result = append(result, entry)
	}
	return result
}

// FrameworkComputedPlayerProperties returns framework defaults: "Color" (CSS color
// string from the player's Color enum or palette fallback) and "MayBeActive".
// Game authors add or type-safely replace values through
// ConfigureComputedProperties instead of overriding this method.
func (g *GameDelegate) FrameworkComputedPlayerProperties(player boardgame.ImmutableSubState) boardgame.PropertyCollection {
	result := boardgame.PropertyCollection{
		"Color":       behaviors.CSSColorForPlayer(player),
		"MayBeActive": g.Manager().Delegate().PlayerMayBeActive(player),
	}
	if score, ok := behaviors.PlayerGameScore(player); ok {
		result["GameScore"] = score
	}
	// Gathering: current team/role/color selections
	if th, ok := player.(behaviors.HasPlayerTeam); ok {
		result["TeamValue"] = th.GetPlayerTeam().Team.String()
	}
	if rh, ok := player.(behaviors.HasPlayerRole); ok {
		result["RoleValue"] = rh.GetPlayerRole().Role.String()
	}
	if ch, ok := player.(behaviors.HasPlayerColor); ok {
		result["ColorValue"] = ch.GetPlayerColor().Color.String()
	}
	if behaviors.PlayerIsAdmin(player) {
		result["IsGameAdmin"] = true
	}
	return result
}

// BeginSetUp does not do anything and returns nil.
func (g *GameDelegate) BeginSetUp(state boardgame.State, variant boardgame.Variant) error {
	//Don't need to do anything by default
	return nil
}

// FinishSetUp doesn't do anything and returns nil.
func (g *GameDelegate) FinishSetUp(state boardgame.State) error {
	//Don't need to do anything by default
	return nil
}

// defaultCheckGameFinishedDelegate can be private because
// DefaultGameFinished implements the methods by default.
type defaultCheckGameFinishedDelegate interface {
	GameEndConditionMet(state boardgame.ImmutableState) bool
	PlayerScore(pState boardgame.ImmutableSubState) int
	LowScoreWins() bool
}

// PlayerGameScorer is an optional interface that can be implemented by
// PlayerSubStates. If it is implemented, base.GameDelegate's default
// PlayerScore() method will return it.
type PlayerGameScorer interface {
	//Score returns the overall score for the game for the player at this
	//point in time.
	GameScore() int
}

// CheckGameFinished by default checks delegate.GameEndConditionMet(). If true,
// then it fetches delegate.PlayerScore() for each player and returns all players
// who have the highest score as winners. (If delegate.LowScoreWins() is true,
// instead of highest score, it does lowest score.) It skips any players who are
// Inactive (according to behaviors.PlayerIsInactive). To use this implementation
// simply implement those methods. This is sufficient for many games, but not
// all, so sometimes needs to be overriden.
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

// LowScoreWins is used in base.GameDelegate's CheckGameFinished. If false
// (default) higher scores are better. If true, however, then lower scores win
// (similar to golf), and all of the players with the lowest score win.
func (g *GameDelegate) LowScoreWins() bool {
	return false
}

// GameEndConditionMet is used in the default CheckGameFinished implementation.
// It should return true when the game is over and ready for scoring.
// CheckGameFinished uses this by default; if you override CheckGameFinished
// you don't need to override this. The default implementation of this simply
// returns false.
func (g *GameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	return false
}

// PlayerScore is used in the default CheckGameFinished implementation. It
// should return the score for the given player. CheckGameFinished uses this by
// default; if you override CheckGameFinished you don't need to override this.
// The default implementation returns pState.GameScore() (if pState implements
// the PlayerGameScorer interface), or 0 otherwise.
func (g *GameDelegate) PlayerScore(pState boardgame.ImmutableSubState) int {
	if scorer, ok := pState.(PlayerGameScorer); ok {
		return scorer.GameScore()
	}
	return 0
}

// NumSeatedActivePlayers returns the number of players who are both seated and
// active. This is typically the number you want to decide how many 'real'
// players there are at the moment. See boardgame/behaviors package doc for more.
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

// NumActivePlayers returns the number of players who are active (whether or not
// they are seated). See also NumSeatedActivePlayers, which is typically what you
// want. See boardgame/behaviors package doc for more.
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

// NumSeatedPlayers returns the number of players who are seated (whether or not
// they are active.) See also NumSeatedActivePlayers, which is typically what you
// want. See boardgame/behaviors package doc for more.
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

// DefaultNumPlayers returns 2.
func (g *GameDelegate) DefaultNumPlayers() int {
	return 2
}

// MinNumPlayers returns 1
func (g *GameDelegate) MinNumPlayers() int {
	return 1
}

// MaxNumPlayers returns 16
func (g *GameDelegate) MaxNumPlayers() int {
	return 16
}

// LegalNumPlayers checks that the number of players is between MinNumPlayers
// and MaxNumPlayers, inclusive. You'd only want to override this if some
// player numbers in that range are not legal, for example a game where only
// even numbers of players may play.
func (g *GameDelegate) LegalNumPlayers(numPlayers int) bool {

	min := g.Manager().Delegate().MinNumPlayers()
	max := g.Manager().Delegate().MaxNumPlayers()

	return numPlayers >= min && numPlayers <= max

}

// ReadyToStart returns nil, indicating the game is always ready to start once
// enough players are seated. Override this in your delegate to add custom
// validation (e.g., team balance, role assignment).
func (g *GameDelegate) ReadyToStart(state boardgame.ImmutableState) error {
	return nil
}

// ChatConfig returns the default chat configuration with all features enabled.
// Override to restrict: e.g., DefaultChatConfig().WithoutDMs()
func (g *GameDelegate) ChatConfig() boardgame.ChatConfig {
	return boardgame.DefaultChatConfig()
}

// ChatPolicyForPlayer returns the chat policy for a specific player based on
// the ChatConfig and current game state. Auto-detects team channels from
// PlayerTeam and generates DM channels for all seated player pairs.
func (g *GameDelegate) ChatPolicyForPlayer(state boardgame.ImmutableState, player boardgame.PlayerIndex) boardgame.ChatPolicy {
	config := g.Manager().Delegate().ChatConfig()

	if !config.IsEnabled() {
		return boardgame.ChatPolicy{Enabled: false}
	}

	var sendChannels []string
	var viewChannels []string

	if config.AllChatEnabled() {
		sendChannels = append(sendChannels, "all")
		viewChannels = append(viewChannels, "all")
	}

	if config.TeamChatEnabled() && player >= 0 && int(player) < len(state.ImmutablePlayerStates()) {
		ps := state.ImmutablePlayerStates()[player]
		if th, ok := ps.(behaviors.HasPlayerTeam); ok {
			teamName := th.GetPlayerTeam().Team.String()
			if teamName != "" {
				ch := "team/" + teamName
				sendChannels = append(sendChannels, ch)
				viewChannels = append(viewChannels, ch)
			}
		}
	}

	// DM channels are generated by the server layer (which has access to
	// UserIDsForGame) and augmented onto this policy. The base delegate
	// cannot generate them because it has no user ID mapping.

	return boardgame.ChatPolicy{
		Enabled:         true,
		SendChannels:    sendChannels,
		ViewChannels:    viewChannels,
		PrebakedOnly:    config.IsPrebakedOnly(),
		AllowedMessages: config.AllowedMessages(),
	}
}

// PlayerMayBeActive returns true for all players, unless they implement
// moves/interfaces.PlayerInactiverer, in which case IsInactive is consulted, and
// if it's true then this returns false. Designed to work well with behaviors.InactivePlayer
func (g *GameDelegate) PlayerMayBeActive(player boardgame.ImmutableSubState) bool {
	return !behaviors.PlayerIsInactive(player)
}

// Variants returns a VariantConfig with no entries.
func (g *GameDelegate) Variants() boardgame.VariantConfig {
	return boardgame.VariantConfig{}
}

// ConfigureAgents by default returns nil. If you want agents in your game,
// override this.
func (g *GameDelegate) ConfigureAgents() []boardgame.Agent {
	return nil
}

// ConfigureEnums simply returns nil. In general you want to override this with
// a body of `return Enums`, if you're using `boardgame-util config` to
// generate your enum set.
func (g *GameDelegate) ConfigureEnums() *enum.Set {
	return nil
}

// ConfigureDecks returns a zero-entry map. You want to override this if you
// have any components in your game (which the vast majority of games do)
func (g *GameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return make(map[string]*boardgame.Deck)
}

// ConfigureStackConstraintConstructors returns constraints.DefaultConstructors(),
// which provides all pre-built constraint constructors (MaxNumComponents,
// Unique, Same, MaxDistinctValues) for use in struct tags. Override this only
// if you need to add custom constructors via constraints.ExtendDefaults(), or
// return nil to disable struct-tag constraints entirely.
func (g *GameDelegate) ConfigureStackConstraintConstructors() []*boardgame.StackConstraintConstructor {
	return constraints.DefaultConstructors()
}

// ConfigureConstants returns a zero-entry map. If you have any constants you
// wa8nt to use client-side or in tag-based struct auto-inflaters, you will want
// to override this.
func (g *GameDelegate) ConfigureConstants() boardgame.PropertyCollection {
	return nil
}

// ConfigureComputedProperties returns no game-specific computed values. Games
// add them with the typed GlobalComputed*/PlayerComputed* constructors.
func (g *GameDelegate) ConfigureComputedProperties() []boardgame.ComputedProperty {
	return nil
}
