package boardgame

import (
	"errors"
)

//GameManager is a central point of coordination for games. It serves as a
//delegate for key parts of a Game's lifecycle to define the logic for a given
//type of game.It is one of the primary ways that a specific game controls
//behavior over and beyond Moves and their Legal states.
type GameManager interface {

	//DistributeComponentToStarterStack is called during set up to establish
	//the Deck/Stack invariant that every component in the chest is placed in
	//precisely one Stack. Game will call this on each component in the Chest
	//in order. This is where the logic goes to make sure each Component goes
	//into its correct starter stack. As long as you put each component into a
	//Stack, the invariant will be met at the end of SetUp. If any errors are
	//returned SetUp fails. Unlike after the game has been SetUp, you can
	//modify payload directly.
	DistributeComponentToStarterStack(state State, c *Component) error

	//CheckGameFinished should return true if the game is finished, and who
	//the winners are. Called after every move is applied.
	CheckGameFinished(state State) (finished bool, winners []int)

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

	//StartingState should return a zero'd state object for this game type.
	//All future states for this particular game will be created by Copy()ing
	//this state. If you return nil, game.SetUp() will fail.
	StartingState(numPlayers int) State

	//StateFromBlob should deserialize a JSON string of this game's State. We
	//need it to be in a game-specific bit of logic because we don't know the
	//real type of the state stuct for this game. Be sure to inflate any
	//Stacks in the state, and set playerIndex for each UserState in order.
	//It's strongly recommended that you test a round-trip of state through
	//this method.
	StateFromBlob(blob []byte, schema int) (State, error)

	//SetGame is called during game.SetUp and passes a reference to the Game
	//that the delegate is part of.
	SetGame(game *Game)
}

//DefaultGameManager is a struct that implements stubs for all of
//GameManager's methods. This makes it easy to override just one or two
//methods by creating your own struct that anonymously embeds this one. You
//almost certainly want to override StartingState.
type DefaultGameManager struct {
	Game *Game
}

func (d *DefaultGameManager) DistributeComponentToStarterStack(state State, c *Component) error {
	//The stub returns an error, because if this is called that means there
	//was a component in the deck. And if we didn't store it in a stack, then
	//we are in violation of the invariant.
	return errors.New("DistributeComponentToStarterStack was called, but the component was not stored in a stack")
}

func (d *DefaultGameManager) CheckGameFinished(state State) (finished bool, winners []int) {
	return false, nil
}

func (d *DefaultGameManager) StateFromBlob(blob []byte, schema int) (State, error) {
	return nil, errors.New("Default delegate does not know how to deserialize state objects")
}

func (d *DefaultGameManager) StartingState(numPlayers int) State {
	return nil
}

func (d *DefaultGameManager) DefaultNumPlayers() int {
	return 2
}

//The Default ProposeFixUpMove runs through all moves in FixUpMoves, in order,
//and returns the first one that is legal at the current state. In many cases,
//this behavior should be suficient and need not be overwritten. Be extra sure
//that your FixUpMoves have a conservative Legal function, otherwise you could
//get a panic from applying too many FixUp moves.
func (d *DefaultGameManager) ProposeFixUpMove(state State) Move {
	for _, move := range d.Game.FixUpMoves() {
		if err := move.Legal(state); err == nil {
			//Found it!
			return move
		}
	}
	//No moves apply now.
	return nil
}

func (d *DefaultGameManager) SetGame(game *Game) {
	d.Game = game
}
