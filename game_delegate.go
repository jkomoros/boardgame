package boardgame

import (
	"errors"
)

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
	//components are distributed to it. This is a good place to configure
	//state that will be necessary for you to make the right decisions in
	//DistributeComponentToStarterStack.
	BeginSetUp(state MutableState)

	//FinishSetUp is called during game.SetUp, *after* components have been
	//distributed to their StarterStack. This is the last chance to modify the
	//state before the game's initial state is considered final. For example,
	//if you have a card game this is where you'd make sure the starter draw
	//stacks are shuffled.
	FinishSetUp(state MutableState)

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

	//EmptyGameState and EmptyPlayerState are called to get an instantiation
	//of the concrete game/player structs that your package defines. This is
	//used both to create the initial state, but also to inflate states from
	//the database. These methods should always return the underlying same
	//type of struct when called. This means that if different players have
	//very different roles in a game, there might be many properties that are
	//not in use for any given player. The simple properties (ints, bools,
	//strings) should all be their zero-value. Importantly, all Stacks should
	//be non- nil, because an initialized struct contains information about
	//things like MaxSize, Size, and a reference to the deck they are
	//affiliated with. Game methods that use these will fail if the State
	//objects return have uninitialized stacks. Since these two methods are
	//always required and always specific to each game type,
	//DefaultGameDelegate does not implement them, as an extra reminder that
	//you must implement them yourself.
	EmptyGameState() MutableGameState
	//EmptyPlayerState is similar to EmptyGameState, but playerIndex is the
	//value that this PlayerState must return when its PlayerIndex() is
	//called.
	EmptyPlayerState(player PlayerIndex) MutablePlayerState

	//EmptyDynamicComponentValues returns an empty DynamicComponentValues for
	//the given deck. If nil is returned, then the components in that deck
	//don't have any dynamic component state. This method must always return
	//the same underlying type of struct for the same deck.
	EmptyDynamicComponentValues(deck *Deck) MutableDynamicComponentValues

	//EmptyComputed{Global,Player}PropertyCollection should return a struct
	//that implements PropertyReadSetter. Computed properties will be stored
	//in the objects that are returned. This allows users of the framework to
	//do a single cast of the underlying object and then access the properties
	//in a type-checked way after that. If you return nil, we will use a
	//generic, flexible PropertyReadSetter instead.
	EmptyComputedGlobalPropertyCollection() ComputedPropertyCollection
	EmptyComputedPlayerPropertyCollection() ComputedPropertyCollection

	//StateSanitizationPolicy returns the policy for sanitizing states in this
	//game. The policy should not change over time. See StatePolicy for more
	//on how sanitization policies are calculated and applied.
	StateSanitizationPolicy() *StatePolicy

	//ComputedPropertiesConfig returns a pointer to the config for how
	//computed properties for this game should be constructed.
	ComputedPropertiesConfig() *ComputedPropertiesConfig

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

//DefaultGameDelegate is a struct that implements stubs for all of
//GameDelegate's methods. This makes it easy to override just one or two
//methods by creating your own struct that anonymously embeds this one. You
//almost certainly want to override StartingState.
type DefaultGameDelegate struct {
	manager *GameManager
}

func (d *DefaultGameDelegate) Diagram(state State) string {
	return "This should be overriden to render a reasonable state here"
}

func (d *DefaultGameDelegate) Name() string {
	return "default"
}

func (d *DefaultGameDelegate) DisplayName() string {
	return "Default Game"
}

func (d *DefaultGameDelegate) Manager() *GameManager {
	return d.manager
}

func (d *DefaultGameDelegate) SetManager(manager *GameManager) {
	d.manager = manager
}

func (d *DefaultGameDelegate) EmptyDynamicComponentValues(deck *Deck) MutableDynamicComponentValues {
	return nil
}

func (d *DefaultGameDelegate) EmptyComputedGlobalPropertyCollection() ComputedPropertyCollection {
	return nil
}

func (d *DefaultGameDelegate) EmptyComputedPlayerPropertyCollection() ComputedPropertyCollection {
	return nil
}

//The Default ProposeFixUpMove runs through all moves in FixUpMoves, in order,
//and returns the first one that is legal at the current state. In many cases,
//this behavior should be suficient and need not be overwritten. Be extra sure
//that your FixUpMoves have a conservative Legal function, otherwise you could
//get a panic from applying too many FixUp moves.
func (d *DefaultGameDelegate) ProposeFixUpMove(state State) Move {
	for _, move := range d.Manager().FixUpMoves() {
		move.DefaultsForState(state)
		if err := move.Legal(state, AdminPlayerIndex); err == nil {
			//Found it!
			return move
		}
	}
	//No moves apply now.
	return nil
}

func (d *DefaultGameDelegate) DistributeComponentToStarterStack(state State, c *Component) (Stack, error) {
	//The stub returns an error, because if this is called that means there
	//was a component in the deck. And if we didn't store it in a stack, then
	//we are in violation of the invariant.
	return nil, errors.New("DistributeComponentToStarterStack was called, but the component was not stored in a stack")
}

func (d *DefaultGameDelegate) StateSanitizationPolicy() *StatePolicy {
	return nil
}

func (d *DefaultGameDelegate) ComputedPropertiesConfig() *ComputedPropertiesConfig {
	return nil
}

func (d *DefaultGameDelegate) BeginSetUp(state MutableState) {
	//Don't need to do anything by default
}

func (d *DefaultGameDelegate) FinishSetUp(state MutableState) {
	//Don't need to do anything by default
}

func (d *DefaultGameDelegate) CheckGameFinished(state State) (finished bool, winners []PlayerIndex) {
	return false, nil
}

func (d *DefaultGameDelegate) DefaultNumPlayers() int {
	return 2
}

func (d *DefaultGameDelegate) LegalNumPlayers(numPlayers int) bool {
	if numPlayers > 0 && numPlayers <= 10 {
		return true
	}
	return false
}
