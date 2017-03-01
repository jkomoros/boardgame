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
	//into its correct starter stack. As long as you put each component into a
	//Stack, the invariant will be met at the end of SetUp. If any errors are
	//returned SetUp fails. Unlike after the game has been SetUp, you can
	//modify payload directly.
	DistributeComponentToStarterStack(state *State, c *Component) error

	//FinishSetUp is called during game.SetUp, after components have been
	//distributed to their StarterStack. This is the last chance to modify the
	//state before the game's initial state is considered final. For example,
	//if you have a card game this is where you'd make sure the starter draw
	//stacks are shuffled.
	FinishSetUp(state *State)

	//CheckGameFinished should return true if the game is finished, and who
	//the winners are. Called after every move is applied.
	CheckGameFinished(state *State) (finished bool, winners []int)

	//ProposeFixUpMove is called after a move has been applied. It may return
	//a FixUp move, which will be applied before any other moves are applied.
	//If it returns nil, we may take the next move off of the queue. FixUp
	//moves are useful for things like shuffling a discard deck back into a
	//draw deck, or other moves that are necessary to get the GameState back
	//into reasonable shape.
	ProposeFixUpMove(state *State) Move

	//DefaultNumPlayers returns the number of users that this game defaults to.
	//For example, for tictactoe, it will be 2. If 0 is provided to
	//game.SetUp(), we wil use this value insteadp.
	DefaultNumPlayers() int

	//StartingState should return a zero'd state object for this game type.
	//All future states for this particular game will be created by Copy()ing
	//this state. If you return nil, game.SetUp() will fail.
	StartingStateProps(numPlayers int) *StateProps

	//StateFromBlob should deserialize a JSON string of this game's State. We
	//need it to be in a game-specific bit of logic because we don't know the
	//real type of the state stuct for this game. Be sure to inflate any
	//Stacks in the state, and set playerIndex for each UserState in order.
	//It's strongly recommended that you test a round-trip of state through
	//this method.

	//GameStateFromBlob and PlayerStateFromBlob are passed blobs representing
	//those JSON'd state and should back back the given objects. This is
	//necessary because we don't know the underlying concrete type and need to
	//be told. Do not worry about re-inflating stacks, that will be done for
	//you in GameManger's StateFromBlob, as long as your stacks are visible
	//via the PropertyReader your states return.
	GameStateFromBlob(blob []byte) (GameState, error)
	PlayerStateFromBlob(blob []byte, playerIndex int) (PlayerState, error)

	//StateSanitizationPolicy returns the policy for sanitizing states in this
	//game. The policy should not change over time. See StatePolicy for more
	//on how sanitization policies are calculated and applied.
	StateSanitizationPolicy() *StatePolicy

	//Diagram should return a basic debug rendering of state in multi-line
	//ascii art. Useful for debugging. State.Diagram() will reach out to this
	//method.
	Diagram(s *State) string

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

func (d *DefaultGameDelegate) Diagram(state *State) string {
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

//The Default ProposeFixUpMove runs through all moves in FixUpMoves, in order,
//and returns the first one that is legal at the current state. In many cases,
//this behavior should be suficient and need not be overwritten. Be extra sure
//that your FixUpMoves have a conservative Legal function, otherwise you could
//get a panic from applying too many FixUp moves.
func (d *DefaultGameDelegate) ProposeFixUpMove(state *State) Move {
	for _, move := range d.Manager().FixUpMoves() {
		move.DefaultsForState(state)
		if err := move.Legal(state); err == nil {
			//Found it!
			return move
		}
	}
	//No moves apply now.
	return nil
}

func (d *DefaultGameDelegate) DistributeComponentToStarterStack(state *State, c *Component) error {
	//The stub returns an error, because if this is called that means there
	//was a component in the deck. And if we didn't store it in a stack, then
	//we are in violation of the invariant.
	return errors.New("DistributeComponentToStarterStack was called, but the component was not stored in a stack")
}

func (d *DefaultGameDelegate) StateSanitizationPolicy() *StatePolicy {
	return nil
}

func (d *DefaultGameDelegate) FinishSetUp(state *State) {
	//Don't need to do anything by default
}

func (d *DefaultGameDelegate) CheckGameFinished(state *State) (finished bool, winners []int) {
	return false, nil
}

func (d *DefaultGameDelegate) GameStateFromBlob(blob []byte) (GameState, error) {
	return nil, errors.New("Default delegate does not know how to deserialize game state objects")
}

func (d *DefaultGameDelegate) PlayerStateFromBlob(blob []byte, index int) (PlayerState, error) {
	return nil, errors.New("Default delegate does not know how to deserialize player state objects")
}

func (d *DefaultGameDelegate) StartingStateProps(numPlayers int) *StateProps {
	return nil
}

func (d *DefaultGameDelegate) DefaultNumPlayers() int {
	return 2
}
