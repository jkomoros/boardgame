package boardgame

//Agent represents an Artificial Intelligence agent that plays as a specific
//player in a specific game. Agents are created and affiliated with
//GameManagers, and then when new games are set up they may be asked to play
//on behalf of a player. After moves are made on a game they are asked if they
//have a move to propose. In many cases the state of the game is sufficient,
//but in some cases Agents may need to store additional information; this is
//handled by the agent marshaling and unmarshaling byte sequences themselves.
type Agent interface {
	//Name is the unique, static name for this type of agent. Many games will
	//have only one type of agent, but the reason for this field is that some
	//games will have different agents with radically different play styles.
	//The name is how agents will be looked up within a manager.
	Name() string

	//DisplayName is a name for the agent that is human-friendly and need not
	//be unique. "Artificial Intelligence", "Computer", "Robby the Robot" are
	//all reasonable examples.
	DisplayName() string

	//SetUpForGame is called when SetUp is called on a Game and Agents are
	//configured for some of the players. This is the chance of the Agent to
	//initialize its state. Whatever state is returned will be stored in the
	//storage layer and passed back to the Agent later.
	SetUpForGame(game *Game, player PlayerIndex) (agentState []byte)

	//ProposeMove is where the meat of Agents happen. It is called once after
	//every MoveChain is made on game (that is, after every player Move and
	//its attendant chain of FixUp moves have all been applied). It is passed
	//the index of the player it is playing at, the game, and the last-stored
	//state for this agent. The game may be interrogated for CurrentState,
	//PlayerMoves, etc, but should NOT have ProposeMove called directly. This
	//method should return a non-nil move ≈ if it wants to propose one.
	//newState is the new state for this agent; if it is nil, a new state will
	//not be saved and the next time ProposeMove is called the previously used
	//state will be provided again.
	ProposeMove(game *Game, player PlayerIndex, agentState []byte) (move Move, newState []byte)
}
