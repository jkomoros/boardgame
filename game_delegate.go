package boardgame

import (
	"github.com/Sirupsen/logrus"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"math"
	"sort"
	"strings"
)

//GameDelegate is the place that various parts of the game lifecycle can be
//modified to support this particular game. Typically you embed
//DefaultGameDelegate in your won struct, and only override methods whose
//default behavior is incorrect for your game.
type GameDelegate interface {

	//Name is a string that defines the type of game this is. The name should
	//be unique and compact, and avoid any special characters other than "-"
	//or "_", since they will sometimes be used in a URL path. Good examples
	//are "tictactoe", "blackjack". Once configured, names should never change
	//over the lifetime of the gametype, since it will be persisted in
	//storage. Subclasses should override this.
	Name() string

	//DisplayName is a string that defines the type of game this is in a way
	//appropriate for humans. The name should be unique but human readable. It
	//is purely for human consumption, and may change over time with no
	//adverse effects. Good examples are "Tic Tac Toe", "Blackjack".
	//Subclasses should override this.
	DisplayName() string

	//Description is a string that describes the game type in a descriptive
	//sentence. A reasonable value for "tictactoe" is "A classic game where
	//players compete to get three in a row"
	Description() string

	//ConfigureMoves will be called during creation of a GameManager in
	//NewGameManager. This is the time to install moves onto the manager by
	//returning a list of moves to install. Typically you use moves.Combine
	//and friends to organize your list of moves to install. If the moves you
	//add are illegal for any reason, NewGameManager will fail with an error.
	//By the time this is called. delegate.SetManager will already have been
	//called, so you'll have access to the manager via Manager().
	ConfigureMoves() []MoveConfig

	//ConfigureAgents will be called when creating a new GameManager. Emit the
	//agents you want to install.
	ConfigureAgents() []Agent

	//ConfigureDecks will be called when the GameManager is being booted up.
	//Each entry in the return value will be configured on the ComponentChest
	//that is being created.
	ConfigureDecks() map[string]*Deck

	//ConfigureEnums is called during set up of a new GameManager. Return the
	//set of enums you want to be associated with this GameManagaer's Chest.
	ConfigureEnums() *enum.Set

	//ConfigureConstants is called during set-up of a new GameManager. Return
	//the map of constants you want to create, which will be configured onto
	//the newly created chest via AddConstant. If any of the AddConstant calls
	//errors, the GameManager will fail to be set up. Constants are primarily
	//useful in two cases: first, when you want to have access to a constant
	//value client-side, and second, when you want to be able to use a
	//constant value in a tag-based struct inflater.
	ConfigureConstants() PropertyCollection

	//GameStateConstructor and PlayerStateConstructor are called to get an
	//instantiation of the concrete game/player structs that your package
	//defines. This is used both to create the initial state, but also to
	//inflate states from the database. These methods should always return the
	//underlying same type of struct when called. This means that if different
	//players have very different roles in a game, there might be many
	//properties that are not in use for any given player. The simple
	//properties (ints, bools, strings) should all be their zero-value.
	//Importantly, all Stacks, Timers, and Enums should be non- nil, because
	//an initialized struct contains information about things like MaxSize,
	//Size, and a reference to the deck they are affiliated with. GameManger
	//will automatically create and use StructInflaters for these types of
	//objects, allowing you to use tag-based configuration to automatically
	//inflate these properties. See the documentation for StructInflater for
	//more. Since these two methods are always required and always specific to
	//each game type, DefaultGameDelegate does not implement them, as an extra
	//reminder that you must implement them yourself.
	GameStateConstructor() ConfigurableSubState
	//PlayerStateConstructor is similar to GameStateConstructor, but
	//playerIndex is the value that this PlayerState must return when its
	//PlayerIndex() is called.
	PlayerStateConstructor(player PlayerIndex) ConfigurablePlayerState

	//DynamicComponentValuesConstructor returns an empty
	//DynamicComponentValues for the given deck. If nil is returned, then the
	//components in that deck don't have any dynamic component state. This
	//method must always return the same underlying type of struct for the
	//same deck. Like GameStateConstructor and PlayerStateConstructor, the
	//engine will automatically create StructInflaters for these objects,
	//allowing you to use tag-based inflation of properties. See StructInflate
	//for more. If the returned object also implements the ComponentValues
	//interface, then SetContainingComponent will be called on the
	//DynamicComponent whenever one is created, with a reference back to the
	//component it's associated with.
	DynamicComponentValuesConstructor(deck *Deck) ConfigurableSubState

	//DistributeComponentToStarterStack is called during set up to establish
	//the Deck/Stack invariant that every component in the chest is placed in
	//precisely one Stack. Game will call this on each component in the Chest
	//in order. This is where the logic goes to make sure each Component goes
	//into its correct starter stack. You must return a non-nil Stack for each
	//call, after which the given Component will be inserted into
	//NextSlotIndex of that stack. If that is not the ordering you desire, you
	//can fix it up in FinishSetUp by using SwapComponents. If any errors are
	//returned, any nil Stacks are returned, or any returned stacks don't have
	//space for another component, NewGame will fail and return an error.
	//State and Component are only provided for reference; do not modify them.
	DistributeComponentToStarterStack(state ImmutableState, c Component) (ImmutableStack, error)

	//BeginSetup is a chance to modify the initial state object *before* the
	//components are distributed to it. It is also where the variant
	//configuration for your gametype will be passed (it will already have
	//been checked for legality and had all configure defaults set), although
	//you can also retrieve that at any time via game.Variant(). This is a
	//good place to configure state that will be necessary for you to make the
	//right decisions in DistributeComponentToStarterStack, or to transcribe
	//config information you were passed into properties on your gameState as
	//appropriate. If error is non-nil, Game setup will be aborted, with the
	//reasoning including the error message provided.
	BeginSetUp(state State, variant Variant) error

	//FinishSetUp is called during NewGame, *after* components have been
	//distributed to their StarterStack. This is the last chance to modify the
	//state before the game's initial state is considered final. For example,
	//if you have a card game this is where you'd make sure the starter draw
	//stacks are shuffled. If your game has multiple rounds, or if you don't
	//want the game to start with it already set-up (e.g. you want to show
	//animations of starter cards being dealt) then it's probably best to do
	//most of the logic in a SetUp phase. See the README for more. If error is
	//non-nil, Game setup will be aborted, with the reasoning including the
	//error message provided.
	FinishSetUp(state State) error

	//CheckGameFinished should return true if the game is finished, and who
	//the winners are. Called after every move is applied.
	CheckGameFinished(state ImmutableState) (finished bool, winners []PlayerIndex)

	//ProposeFixUpMove is called after a move has been applied. It may return
	//a FixUp move, which will be applied before any other moves are applied.
	//If it returns nil, we may take the next move off of the queue. FixUp
	//moves are useful for things like shuffling a discard deck back into a
	//draw deck, or other moves that are necessary to get the GameState back
	//into reasonable shape.
	ProposeFixUpMove(state ImmutableState) Move

	//DefaultNumPlayers returns the number of users that this game defaults to.
	//For example, for tictactoe, it will be 2. If 0 is provided to
	//manager.NewGame(), we wil use this value instead.
	DefaultNumPlayers() int

	//Min/MaxNumPlayers should return the min and max number of players,
	//respectively. The engine doesn't use this directly, instead looking at
	//LegalNumPlayers. Typically your LegalNumPlayers will check the given
	//number of players is between these two extremes.
	MinNumPlayers() int
	MaxNumPlayers() int

	//LegalNumPlayers will be consulted when a new game is created. It should
	//return true if the given number of players is legal, and false
	//otherwise. If this returns false, the NewGame will fail with an error.
	//Game creation will automatically reject a numPlayers that does not
	//result in at least one player existing. Generally this is simply
	//checking to make sure the number of players is between Min and Max
	//(inclusive), although some games could only allow, for example, even
	//numbers of players.
	LegalNumPlayers(numPlayers int) bool

	//Variants returns a VariantConfig, which describes the different
	//categories of configuration values and the legal values they may take
	//on. Ultimately your LegalVariant is the final arbiter of which variants
	//are legal; this method is mainly used so that user interfaces know which
	//variants to show to the user. In general you shouldn't call this, but
	//instead call gameManager.Variants() which will ensure your VariantConfig
	//is initalized and memoize the return result.
	Variants() VariantConfig

	//CurrentPlayerIndex returns the index of the "current" player--a notion
	//that is game specific (and sometimes inapplicable). If CurrentPlayer
	//doesn't make sense (perhaps the game never has a notion of current
	//player, or the type of round that we're in has no current player), this
	//should return ObserverPlayerIndex. The result of this method is used to
	//power state.CurrentPlayer.
	CurrentPlayerIndex(state ImmutableState) PlayerIndex

	//CurrentPhase returns the phase that the game state is currently in.
	//Phase is a formalized convention used in moves.Base to make it easier to
	//write fix-up moves that only apply in certain phases, like SetUp. The
	//return result is primarily used in moves.Base to check whether it is one
	//of the phases in a give Move's LegalPhases. See moves.Base for more
	//information.
	CurrentPhase(state ImmutableState) int

	//PhaseEnum returns the enum for game phases (the return values of
	//CurrentPhase are expected to be valid enums within that enum). If this
	//returns a non-nil enums.TreeEnum, then the state will not be able to be
	//saved if CurrentPhase() returns a value that is not a leaf-node.
	PhaseEnum() enum.Enum

	//SanitizationPolicy is consulted when sanitizing states. It is called for
	//each prop in the state, including the set of groups that this player is
	//a mamber of. In practice the default behavior of DefaultGameDelegate,
	//which uses struct tags to figure out the policy, is sufficient and you
	//do not need to override this. For more on how sanitization works, see
	//the package doc. The statePropetyRef passed will always have the Index
	//properties set to -1, signifying that the returned policy applies to all
	//items in the Stack/Board.
	SanitizationPolicy(prop StatePropertyRef, groupMembership map[int]bool) Policy

	//If you have computed properties that you want to be included in your
	//JSON (for example, for use clientside), export them here by creating a
	//dictionary with their values.
	ComputedGlobalProperties(state ImmutableState) PropertyCollection
	ComputedPlayerProperties(player ImmutablePlayerState) PropertyCollection

	//Diagram should return a basic debug rendering of state in multi-line
	//ascii art. Useful for debugging. State.Diagram() will reach out to this
	//method.
	Diagram(s ImmutableState) string

	//SetManager configures which manager this delegate is in use with. A
	//given delegate can only be used by a single manager at a time.
	SetManager(manager *GameManager)

	//Manager returns the Manager that was set on this delegate.
	Manager() *GameManager
}

//PropertyCollection is just an alias for map[string]interface{}
type PropertyCollection map[string]interface{}

//Copy returns a shallow copy of PropertyCollection
func (p PropertyCollection) Copy() PropertyCollection {
	result := make(PropertyCollection, len(p))
	for key, val := range result {
		result[key] = val
	}
	return result
}

//DefaultGameDelegate is a struct that implements stubs for all of
//GameDelegate's methods. This makes it easy to override just one or two
//methods by creating your own struct that anonymously embeds this one. Name,
//GameStateConstructor, PlayerStateConstructor, and ConfigureMoves are not
//implemented, since those almost certainly must be overridden for your
//particular game.
type DefaultGameDelegate struct {
	manager *GameManager
}

//Diagram returns the string "This should be overriden to render a reasonable state here"
func (d *DefaultGameDelegate) Diagram(state ImmutableState) string {
	return "This should be overriden to render a reasonable state here"
}

//DisplayName by default just returns the title-case of Name() that is
//returned from the delegate in use.
func (d *DefaultGameDelegate) DisplayName() string {
	return strings.Title(d.Manager().Delegate().Name())
}

//Description defaults to "" if not overriden.
func (d *DefaultGameDelegate) Description() string {
	return ""
}

//Manager returns the manager object that was provided to SetManager.
func (d *DefaultGameDelegate) Manager() *GameManager {
	return d.manager
}

//SetManager keeps a reference to the passed manager, and returns it when
//Manager() is called.
func (d *DefaultGameDelegate) SetManager(manager *GameManager) {
	d.manager = manager
}

//DynamicComponentValuesConstructor returns nil, as not all games have
//DynamicComponentValues. Override this if your game does require
//DynamicComponentValues.
func (d *DefaultGameDelegate) DynamicComponentValuesConstructor(deck *Deck) ConfigurableSubState {
	return nil
}

//The Default ProposeFixUpMove runs through all moves in Moves, in order, and
//returns the first one that returns true from IsFixUp and is legal at the
//current state. In many cases, this behavior should be suficient and need not
//be overwritten. Be extra sure that your FixUpMoves have a conservative Legal
//function, otherwise you could get a panic from applying too many FixUp
//moves. Wil emit debug information about why certain fixup moves didn't apply
//if the Manager's log level is Debug or higher.
func (d *DefaultGameDelegate) ProposeFixUpMove(state ImmutableState) Move {

	isDebug := d.Manager().Logger().Level >= logrus.DebugLevel

	var logEntry *logrus.Entry

	if isDebug {
		logEntry = d.Manager().Logger().WithFields(logrus.Fields{
			"game":    state.Game().Id(),
			"version": state.Version(),
		})
		logEntry.Debug("***** ProposeFixUpMove called *****")
	}

	for _, move := range state.Game().Moves() {

		var entry *logrus.Entry
		if isDebug {
			entry = logEntry.WithField("movetype", move.Info().Name())
		}

		if !move.IsFixUp() {
			//Not a fix up move
			continue
		}

		if err := move.Legal(state, AdminPlayerIndex); err == nil {
			if isDebug {
				entry.Debug(move.Info().Name() + " : MATCH")
			}
			//Found it!
			return move
		} else {
			if isDebug {
				entry.Debug(move.Info().Name() + " : " + err.Error())
			}
		}
	}
	if isDebug {
		logEntry.Debug("NO MATCH")
	}
	//No moves apply now.
	return nil
}

//CurrentPlayerIndex returns gameState.CurrentPlayer, if that is a PlayerIndex
//property. If not, returns ObserverPlayerIndex.≈
func (d *DefaultGameDelegate) CurrentPlayerIndex(state ImmutableState) PlayerIndex {
	index, err := state.ImmutableGameState().Reader().PlayerIndexProp("CurrentPlayer")

	if err != nil {
		//Guess that's not where they store CurrentPlayer.
		return ObserverPlayerIndex
	}

	return index
}

//CurrentPhase by default with return the value of gameState.Phase, if it is
//an enum. If it is not, it will return -1 instead, to make it more clear that
//it's an invalid CurrentPhase (phase 0 is often valid).
func (d *DefaultGameDelegate) CurrentPhase(state ImmutableState) int {

	phaseEnum, err := state.ImmutableGameState().Reader().ImmutableEnumProp("Phase")

	if err != nil {
		//Guess it wasn't there
		return -1
	}

	return phaseEnum.Value()

}

//PhaseEnum defaults to the enum named "Phase" which is the convention for the
//name of the Phase enum. moves.Base will handle cases where that isn't a
//valid enum gracefully.
func (d *DefaultGameDelegate) PhaseEnum() enum.Enum {
	return d.Manager().Chest().Enums().Enum("Phase")
}

func (d *DefaultGameDelegate) DistributeComponentToStarterStack(state ImmutableState, c Component) (ImmutableStack, error) {
	//The stub returns an error, because if this is called that means there
	//was a component in the deck. And if we didn't store it in a stack, then
	//we are in violation of the invariant.
	return nil, errors.New("DistributeComponentToStarterStack was called, but the component was not stored in a stack")
}

//SanitizatinoPolicy uses struct tags to identify the right policy to apply
//(see the package doc on SanitizationPolicy for how to configure those tags).
//It sees which policies apply given the provided group membership, and then
//returns the LEAST restrictive policy that applies. This behavior is almost
//always what you want; it is rare to need to override this method.
func (d *DefaultGameDelegate) SanitizationPolicy(prop StatePropertyRef, groupMembership map[int]bool) Policy {

	manager := d.Manager()

	inflater := manager.Internals().StructInflater(prop)

	if inflater == nil {
		return PolicyInvalid
	}

	policyMap := inflater.PropertySanitizationPolicy(prop.PropName)

	var applicablePolicies []int

	for group, isMember := range groupMembership {

		//The only ones that are in the map should be `true` but sanity check
		//just in case.
		if !isMember {
			continue
		}

		//Only if the policy is actually in the map should we use it
		if policy, ok := policyMap[group]; ok {
			applicablePolicies = append(applicablePolicies, int(policy))
		}
	}

	if len(applicablePolicies) == 0 {
		return PolicyVisible
	}

	sort.Ints(applicablePolicies)

	return Policy(applicablePolicies[0])

}

//ComputedGlobalProperties returns nil.
func (d *DefaultGameDelegate) ComputedGlobalProperties(state ImmutableState) PropertyCollection {
	return nil
}

//ComputedPlayerProperties returns nil.
func (d *DefaultGameDelegate) ComputedPlayerProperties(player ImmutablePlayerState) PropertyCollection {
	return nil
}

//BeginSetUp does not do anything and returns nil.
func (d *DefaultGameDelegate) BeginSetUp(state State, variant Variant) error {
	//Don't need to do anything by default
	return nil
}

//FinishSetUp doesn't do anything and returns nil.
func (d *DefaultGameDelegate) FinishSetUp(state State) error {
	//Don't need to do anything by default
	return nil
}

//defaultCheckGameFinishedDelegate can be private because
//DefaultGameFinished implements the methods by default.
type defaultCheckGameFinishedDelegate interface {
	GameEndConditionMet(state ImmutableState) bool
	PlayerScore(pState ImmutablePlayerState) int
	LowScoreWins() bool
}

//PlayerGameScorer is an optional interface that can be implemented by
//PlayerSubStates. If it is implemented, DefaultGameDelegate's default
//PlayerScore() method will return it.
type PlayerGameScorer interface {
	//Score returns the overall score for the game for the player at this
	//point in time.
	GameScore() int
}

//CheckGameFinished by default checks delegate.GameEndConditionMet(). If true,
//then it fetches delegate.PlayerScore() for each player and returns all
//players who have the highest score as winners. (If delegate.LowScoreWins()
//is true, instead of highest score, it does lowest score.) To use this
//implementation simply implement those methods. This is sufficient for many
//games, but not all, so sometimes needs to be overriden.
func (d *DefaultGameDelegate) CheckGameFinished(state ImmutableState) (finished bool, winners []PlayerIndex) {

	if d.Manager() == nil {
		return false, nil
	}

	//Have to reach up to the manager's delegate to get the thing that embeds
	//us. Don't use the comma-ok pattern because we want to panic with
	//descriptive error if not met.
	checkGameFinished := d.Manager().Delegate().(defaultCheckGameFinishedDelegate)

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
		score := checkGameFinished.PlayerScore(player)

		if score == extremeScore {
			winners = append(winners, PlayerIndex(i))
		}
	}

	return true, winners

}

//LowScoreWins is used in DefaultGameDelegate's CheckGameFinished. If false
//(default) higher scores are better. If true, however, then lower scores win
//(similar to golf), and all of the players with the lowest score win.
func (d *DefaultGameDelegate) LowScoreWins() bool {
	return false
}

//GameEndConditionMet is used in the default CheckGameFinished implementation.
//It should return true when the game is over and ready for scoring.
//CheckGameFinished uses this by default; if you override CheckGameFinished
//you don't need to override this. The default implementation of this simply
//returns false.
func (d *DefaultGameDelegate) GameEndConditionMet(state ImmutableState) bool {
	return false
}

//PlayerScore is used in the default CheckGameFinished implementation. It
//should return the score for the given player. CheckGameFinished uses this by
//default; if you override CheckGameFinished you don't need to override this.
//The default implementation returns pState.GameScore() (if pState implements
//the PlayerGameScorer interface), or 0 otherwise.
func (d *DefaultGameDelegate) PlayerScore(pState ImmutablePlayerState) int {
	if scorer, ok := pState.(PlayerGameScorer); ok {
		return scorer.GameScore()
	}
	return 0
}

//DefaultNumPlayers returns 2.
func (d *DefaultGameDelegate) DefaultNumPlayers() int {
	return 2
}

//MinNumPlayers returns 1
func (d *DefaultGameDelegate) MinNumPlayers() int {
	return 1
}

//MaxNumPlayers returns 16
func (d *DefaultGameDelegate) MaxNumPlayers() int {
	return 16
}

//LegalNumPlayers checks that the number of players is between MinNumPlayers
//and MaxNumPlayers, inclusive. You'd only want to override this if some
//player numbers in that range are not legal, for example a game where only
//even numbers of players may play.
func (d *DefaultGameDelegate) LegalNumPlayers(numPlayers int) bool {

	min := d.Manager().Delegate().MinNumPlayers()
	max := d.Manager().Delegate().MaxNumPlayers()

	return numPlayers >= min && numPlayers <= max

}

//Variants returns a VariantConfig with no entries.
func (d *DefaultGameDelegate) Variants() VariantConfig {
	return VariantConfig{}
}

//ConfigureAgents by default returns nil. If you want agents in your game,
//override this.
func (d *DefaultGameDelegate) ConfigureAgents() []Agent {
	return nil
}

//ConfigureEnums simply returns nil. In general you want to override this with
//a body of `return Enums`, if you're using `boardgame-util config` to
//generate your enum set.
func (d *DefaultGameDelegate) ConfigureEnums() *enum.Set {
	return nil
}

//ConfigureDecks returns a zero-entry map. You want to override this if you
//have any components in your game (which the vast majority of games do)
func (d *DefaultGameDelegate) ConfigureDecks() map[string]*Deck {
	return make(map[string]*Deck)
}

//ConfigureConstants returns a zero-entry map. If you have any constants you
//wa8nt to use client-side or in tag-based struct auto-inflaters, you will want
//to override this.
func (d *DefaultGameDelegate) ConfigureConstants() PropertyCollection {
	return nil
}
