package boardgame

import (
	"github.com/Sirupsen/logrus"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"sort"
	"strconv"
)

//GameConfig is just a map of keys to values that are passed to your game so
//it can configure different alternate rulesets, for example using a Short
//variant that uses fewer cards and should play faster, or using a different
//deck of cards than normal. The config will be considered legal if it passes
//Delegate.LegalConfig(), and will be passed to Delegate.BeginSetup so that
//you can set up your game in whatever way makes sense for a given Config.
//Your Delegate defines what valid keys and values are with its return value
//for Configs(), and how they should show to the user with ConfigDisplay.
type GameConfig map[string]string

//GameDelegate is the place that various parts of the game lifecycle can be
//modified to support this particular game.
type GameDelegate interface {

	//Name is a string that defines the type of game this is. The name should
	//be unique and compact. Good examples are "tictactoe", "blackjack". Once
	//configured, names should never change over the lifetime of the gametype,
	//since it will be persisted in storage. Subclasses should override this.
	Name() string

	//DisplayName is a string that defines the type of game this is in a way
	//appropriate for humans. The name should be unique but human readable. It
	//is purely for human consumption, and may change over time with no
	//adverse effects. Good examples are "Tic Tac Toe", "Blackjack".
	//Subclasses should override this.
	DisplayName() string

	//DistributeComponentToStarterStack is called during set up to establish
	//the Deck/Stack invariant that every component in the chest is placed in
	//precisely one Stack. Game will call this on each component in the Chest
	//in order. This is where the logic goes to make sure each Component goes
	//into its correct starter stack. You must return a non-nil Stack for each
	//call, after which the given Component will be inserted into
	//NextSlotIndex of that stack. If that is not the ordering you desire, you
	//can fix it up in FinishSetUp by using SwapComponents. If any errors are
	//returned, any nil Stacks are returned, or any returned stacks don't have
	//space for another component, game.SetUp will fail. State and Component
	//are only provided for reference; do not modify them.
	DistributeComponentToStarterStack(state State, c *Component) (Stack, error)

	//BeginSetup is a chance to modify the initial state object *before* the
	//components are distributed to it. It is also where the config for your
	//gametype will be passed (it will have already passed LegalConfig). This
	//is a good place to configure state that will be necessary for you to
	//make the right decisions in DistributeComponentToStarterStack. If error
	//is non-nil, Game setup will be aborted, with the reasoning including the
	//error message provided.
	BeginSetUp(state MutableState, config GameConfig) error

	//FinishSetUp is called during game.SetUp, *after* components have been
	//distributed to their StarterStack. This is the last chance to modify the
	//state before the game's initial state is considered final. For example,
	//if you have a card game this is where you'd make sure the starter draw
	//stacks are shuffled. If error is non-nil, Game setup will be aborted,
	//with the reasoning including the error message provided.
	FinishSetUp(state MutableState) error

	//CheckGameFinished should return true if the game is finished, and who
	//the winners are. Called after every move is applied.
	CheckGameFinished(state State) (finished bool, winners []PlayerIndex)

	//ProposeFixUpMove is called after a move has been applied. It may return
	//a FixUp move, which will be applied before any other moves are applied.
	//If it returns nil, we may take the next move off of the queue. FixUp
	//moves are useful for things like shuffling a discard deck back into a
	//draw deck, or other moves that are necessary to get the GameState back
	//into reasonable shape.
	ProposeFixUpMove(state State) Move

	//DefaultNumPlayers returns the number of users that this game defaults to.
	//For example, for tictactoe, it will be 2. If 0 is provided to
	//game.SetUp(), we wil use this value insteadp.
	DefaultNumPlayers() int

	//LegalNumPlayers will be consulted when a new game is created. It should
	//return true if the given number of players is legal, and false
	//otherwise. If this returns false, the game's SetUp will fail. Game.SetUp
	//will automatically reject a numPlayers that does not result in at least
	//one player existing.
	LegalNumPlayers(numPlayers int) bool

	//Configs returns a list of all of the various config values that are
	//valid for the given config keys. Ultimately your LegalConfig is the
	//final arbiter of which configs are legal; this method is mainly used so
	//that user interfaces know which configs to show to the user.
	Configs() map[string][]string

	//ConfigKeyDisplay will be called to figure out the user-visible name this
	//config key should have, and a description of what it does to show to a
	//user. It will be called repeatedly by each key in the map returned by
	//your Configs().
	ConfigKeyDisplay(key string) (displayName, description string)

	//ConfigValueDisplay is called to figure out the displayname and
	//description for each key/val in Config to show to users in the
	//interface. It will be called repeatedly on every key/val pair in the map
	//returned by Configs().
	ConfigValueDisplay(key, val string) (displayName, description string)

	//LegalConfig will be consulted when a new game is created. It should
	//return nil if the provided config is a reasonable configuration for your
	//gametype, and a descriptive error (that's reasonable to show to the end
	//user) otherwise. If this returns non-nil, the game's SetUp will fail.
	LegalConfig(config GameConfig) error

	//CurrentPlayerIndex returns the index of the "current" player--a notion
	//that is game specific (and sometimes inapplicable). If CurrentPlayer
	//doesn't make sense (perhaps the game never has a notion of current
	//player, or the type of round that we're in has no current player), this
	//should return ObserverPlayerIndex. The result of this method is used to
	//power state.CurrentPlayer.
	CurrentPlayerIndex(state State) PlayerIndex

	//CurrentPhase returns the phase that the game state is currently in.
	//Phase is a formalized convention used in moves.Base to make it easier to
	//write fix-up moves that only apply in certain phases, like SetUp. The
	//return result is primarily used in moves.Base to check whether it is one
	//of the phases in a give Move's LegalPhases. See moves.Base for more
	//information.
	CurrentPhase(state State) int

	//PhaseEnum returns the enum for game phases (the return values of
	//CurrentPhase are expected to be valid enums within that enum). Primarily
	//used by moves.Base to generate meaningful error messages in Legal().
	PhaseEnum() enum.Enum

	//PhaseMoveProgression returns the names of the strings of moves in the
	//given phase that must be applied in order. moves.Base's Legal() method
	//uses this to determine if a given move is allowed to apply now. A nil
	//return denotes that any move that is legal in this phase is legal at any
	//time in the phase. This functionality is useful for SetUp phases where
	//you have many steps to apply in a row and signaling of when to apply a
	//move can be error prone. See moves.Base's Legal method documentation for
	//more about how to use it.
	PhaseMoveProgression(phase int) []string

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
	//Size, and a reference to the deck they are affiliated with. It is also
	//possible to use tag-based auto-initalization for these fields; see the
	//package doc on Constructors.  Since these two methods are always
	//required and always specific to each game type, DefaultGameDelegate does
	//not implement them, as an extra reminder that you must implement them
	//yourself.
	GameStateConstructor() ConfigurableSubState
	//PlayerStateConstructor is similar to GameStateConstructor, but
	//playerIndex is the value that this PlayerState must return when its
	//PlayerIndex() is called.
	PlayerStateConstructor(player PlayerIndex) ConfigurablePlayerState

	//DynamicComponentValuesConstructor returns an empty DynamicComponentValues for
	//the given deck. If nil is returned, then the components in that deck
	//don't have any dynamic component state. This method must always return
	//the same underlying type of struct for the same deck.
	DynamicComponentValuesConstructor(deck *Deck) ConfigurableSubState

	//SanitizationPolicy is consulted when sanitizing states. It is called for
	//each prop in the state, including the set of groups that this player is
	//a mamber of. In practice the default behavior of DefaultGameDelegate,
	//which uses struct tags to figure out the policy, is sufficient and you
	//do not need to override this. For more on how sanitization works, see
	//the package doc.
	SanitizationPolicy(prop StatePropertyRef, groupMembership map[int]bool) Policy

	//If you have computed properties that you want to be included in your
	//JSON (for example, for use clientside), export them here by creating a
	//dictionary with their values.
	ComputedGlobalProperties(state State) PropertyCollection
	ComputedPlayerProperties(player PlayerState) PropertyCollection

	//Diagram should return a basic debug rendering of state in multi-line
	//ascii art. Useful for debugging. State.Diagram() will reach out to this
	//method.
	Diagram(s State) string

	//SetManager configures which manager this delegate is in use with. A
	//given delegate can only be used by a single manager at a time.
	SetManager(manager *GameManager)

	//Manager returns the Manager that was set on this delegate.
	Manager() *GameManager
}

//PropertyCollection is just an alias for map[string]interface{}
type PropertyCollection map[string]interface{}

//DefaultGameDelegate is a struct that implements stubs for all of
//GameDelegate's methods. This makes it easy to override just one or two
//methods by creating your own struct that anonymously embeds this one.
//GameStateConstructor and PlayerStateConstructor are not implemented, since
//those almost certainly must be overridden for your particular game.
type DefaultGameDelegate struct {
	manager          *GameManager
	moveProgressions map[int][]string
}

//AddMovesForPhaseProgression is a convenience wrapper designed to configure
//moves on the game manager during initialization for moves that are only
//useful as part of a progression. Adds each MoveType to the game manager, and
//also configures the value that will be returned from the default
//PhaseMoveProgression() for the phase in question. If the move types aren't
//part of a move progression, then adding them with manager.BulkAddMoveTypes
//is sufficient.
func (d *DefaultGameDelegate) AddMovesForPhaseProgression(phase int, config ...*MoveTypeConfig) error {

	if d.Manager() == nil {
		return errors.New("Delegate has not had its manager set yet")
	}

	if d.moveProgressions == nil {
		d.moveProgressions = make(map[int][]string)
	}

	var moveProgression []string

	for i, moveConfig := range config {

		moveProgression = append(moveProgression, moveConfig.Name)

		if err := d.Manager().AddMoveType(moveConfig); err != nil {
			return errors.New("Couldn't add " + strconv.Itoa(i) + " move config: " + err.Error())
		}
	}

	d.moveProgressions[phase] = moveProgression

	return nil

}

func (d *DefaultGameDelegate) Diagram(state State) string {
	return "This should be overriden to render a reasonable state here"
}

func (d *DefaultGameDelegate) Name() string {
	return "default"
}

//DisplayName by default just returns the Name() that is returned from the
//delegate in use.
func (d *DefaultGameDelegate) DisplayName() string {
	return d.Manager().Delegate().Name()
}

func (d *DefaultGameDelegate) Manager() *GameManager {
	return d.manager
}

func (d *DefaultGameDelegate) SetManager(manager *GameManager) {
	d.manager = manager
}

func (d *DefaultGameDelegate) DynamicComponentValuesConstructor(deck *Deck) ConfigurableSubState {
	return nil
}

//The Default ProposeFixUpMove runs through all moves in FixUpMoves, in order,
//and returns the first one that is legal at the current state. In many cases,
//this behavior should be suficient and need not be overwritten. Be extra sure
//that your FixUpMoves have a conservative Legal function, otherwise you could
//get a panic from applying too many FixUp moves. Wil emit debug information
//about why certain fixup moves didn't apply if the Manager's log level is
//Debug or higher.
func (d *DefaultGameDelegate) ProposeFixUpMove(state State) Move {

	isDebug := d.Manager().Logger().Level >= logrus.DebugLevel

	var logEntry *logrus.Entry

	if isDebug {
		logEntry = d.Manager().Logger().WithFields(logrus.Fields{
			"game":    state.Game().Id(),
			"version": state.Version(),
		})
		logEntry.Debug("***** ProposeFixUpMove called *****")
	}

	for _, moveType := range d.Manager().FixUpMoveTypes() {
		var entry *logrus.Entry
		if isDebug {
			entry = logEntry.WithField("movetype", moveType.Name())
		}
		move := moveType.NewMove(state)
		if err := move.Legal(state, AdminPlayerIndex); err == nil {
			if isDebug {
				entry.Debug(moveType.Name() + " : MATCH")
			}
			//Found it!
			return move
		} else {
			if isDebug {
				entry.Debug(moveType.Name() + " : " + err.Error())
			}
		}
	}
	if isDebug {
		logEntry.Debug("NO MATCH")
	}
	//No moves apply now.
	return nil
}

func (d *DefaultGameDelegate) CurrentPlayerIndex(state State) PlayerIndex {
	return ObserverPlayerIndex
}

//CurrentPhase returns -1 by default, which amkes it more obvious that you
//haven't overridden this if you choose to use Phases. (As Phase 0 is normally
//a valid  phase).
func (d *DefaultGameDelegate) CurrentPhase(state State) int {
	return -1
}

//PhaseEnum defaults to the enum named "Phase" which is the convention for the
//name of the Phase enum. moves.Base will handle cases where that isn't a
//valid enum gracefully.
func (d *DefaultGameDelegate) PhaseEnum() enum.Enum {
	return d.Manager().Chest().Enums().Enum("Phase")
}

//PhaseMoveProgression will return the move progression if it was added with
//AddMovesForPhaseProgression. If not, will return nil, which means that any
//moves that are legal in this phase are allowed in any order. If you used
//AddMovesForPhaseProgression during setup (or have no phases with a specific
//progression of moves) then you likely have no reason to override this
//method.
func (d *DefaultGameDelegate) PhaseMoveProgression(phase int) []string {
	if d.moveProgressions == nil {
		return nil
	}
	return d.moveProgressions[phase]
}

func (d *DefaultGameDelegate) DistributeComponentToStarterStack(state State, c *Component) (Stack, error) {
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

	var validator *readerValidator
	switch prop.Group {
	case StateGroupGame:
		validator = manager.gameValidator
	case StateGroupPlayer:
		validator = manager.playerValidator
	case StateGroupDynamicComponentValues:
		validator = manager.dynamicComponentValidator[prop.DeckName]
	}

	if validator == nil {
		return PolicyInvalid
	}

	policyMap := validator.sanitizationPolicy[prop.PropName]

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

func (d *DefaultGameDelegate) ComputedGlobalProperties(state State) PropertyCollection {
	return nil
}

func (d *DefaultGameDelegate) ComputedPlayerProperties(player PlayerState) PropertyCollection {
	return nil
}

func (d *DefaultGameDelegate) BeginSetUp(state MutableState, config GameConfig) error {
	//Don't need to do anything by default
	return nil
}

func (d *DefaultGameDelegate) FinishSetUp(state MutableState) error {
	//Don't need to do anything by default
	return nil
}

func (d *DefaultGameDelegate) CheckGameFinished(state State) (finished bool, winners []PlayerIndex) {
	return false, nil
}

func (d *DefaultGameDelegate) DefaultNumPlayers() int {
	return 2
}

func (d *DefaultGameDelegate) LegalNumPlayers(numPlayers int) bool {
	if numPlayers > 0 && numPlayers <= 16 {
		return true
	}
	return false
}

func (d *DefaultGameDelegate) Configs() map[string][]string {
	return make(map[string][]string)
}

//ConfigKeyDisplay by default just returns the key and no description.
func (d *DefaultGameDelegate) ConfigKeyDisplay(key string) (displayName, description string) {
	return key, ""
}

//ConfigValueDisplay by default just returns the value and no description.
func (d *DefaultGameDelegate) ConfigValueDisplay(key, val string) (displayName, description string) {
	return val, ""
}

//LegalConfig on DefaultGameDelegate by default verifies that each of the keys
//and values in the Config are legal keys and values in the map returned by
//Configs().
func (d *DefaultGameDelegate) LegalConfig(config GameConfig) error {
	//We can't call Configs on self because that might not be the right one,
	//it might be overridden.
	del := d.Manager().Delegate()

	validConfigs := del.Configs()
	for key, val := range config {
		if _, ok := validConfigs[key]; !ok {
			return errors.New("configuration had a property called " + key + " that isn't expected")
		}
		foundAllowedVal := false
		for _, allowedVal := range validConfigs[key] {
			if val == allowedVal {
				foundAllowedVal = true
			}
		}
		if !foundAllowedVal {
			keyDisplayName, _ := del.ConfigKeyDisplay(key)
			return errors.New("configuration's " + keyDisplayName + " property had a value that wasn't allowed: " + val)
		}
	}

	return nil
}
