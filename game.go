package boardgame

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"github.com/jkomoros/boardgame/errors"
)

//maxRecurseCount is the number of fixUp moves that can be considered normal--
//anything more than that and we'll return an error because the delegate is
//likely going to return fixup moves forever.
const maxRecurseCount = 256

const selfInitiatorSentinel = -1

//ErrTooManyFixUps is returned from game.ProposeMove if too many fix up moves
//are applied, which implies that there is a FixUp move configured to always
//be legal, and is evidence of a serious error in your game logic.
var ErrTooManyFixUps = errors.New("we recursed deeply in fixup, which implies that ProposeFixUp has a move that is always legal")

//A Game represents a specific game between a collection of Players; an
//instantiation of a game of the given type. Create a new one with
//GameManager.NewGame().
type Game struct {
	manager *GameManager

	finished bool

	winners []PlayerIndex

	agents []string

	//The current version of State.
	version int

	numPlayers int

	variant Variant

	//Memozied answer to CurrentState. Invalidated whenever ApplyMove is
	//called.
	cachedCurrentState    ImmutableState
	cachedHistoricalMoves []*MoveStorageRecord

	//Modifiable controls whether moves can be made on this game.
	modifiable bool

	//A unique ID provided to this game when it is created.
	id string

	//A secret salt that is used to generate semi-stable Ids for components.
	//Never transmitted to client.
	secretSalt string

	//Proposed moves is where moves that have been proposed but have not yet been applied go.
	proposedMoves chan *proposedMoveItem
	//How a game can be signaled to trigger a pass of fixups
	fixUpTriggered chan DelayedError

	//if true, we will not wait to propose agent moves (mainly used for
	//testing.)
	instantAgentMoves bool

	//Initalized is set to True after SetUp is called.
	initalized bool

	created  time.Time
	modified time.Time

	//TODO: HistoricalState(index int) and HistoryLen() int

	//TODO: an array of Player objects.
}

const gameIDLength = 16

//DelayedError is a chan on which an error (or nil) will be sent at a later
//time. Primarily returned from game.ProposeMove(), so the method can return
//immediately even before the move is processed, which might take a long time
//if there are many moves ahead in the queue.
type DelayedError chan error

type proposedMoveItem struct {
	move     Move
	proposer PlayerIndex
	//Ch is the channel we should either return an error on and then close, or
	//send nil and close.
	ch DelayedError
}

var defaultStringRand *rand.Rand

func init() {

	defaultStringRand = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

}

const randomStringChars = "ABCDEF0123456789"

//randomString returns a random string of the given length. If rand is not
//nil, will use that source. Ohterwise will use a global source.
func randomString(length int, rnd *rand.Rand) string {
	var result = ""

	if rnd == nil {
		rnd = defaultStringRand
	}

	for len(result) < length {
		result += string(randomStringChars[rnd.Intn(len(randomStringChars))])
	}

	return result
}

//Created returns the time stamp when this game was first created.
func (g *Game) Created() time.Time {
	return g.created
}

//Modified returns the timstamp when the last move was applied to this game.
func (g *Game) Modified() time.Time {
	return g.modified
}

//Variant returns a copy of the Variant passed to NewGame to create this
//game originally.
func (g *Game) Variant() Variant {

	if g.variant == nil {
		return nil
	}

	result := make(Variant, len(g.variant))

	for key, val := range g.variant {
		result[key] = val
	}

	return result
}

//Winners is the player indexes who were winners. Typically, this will be
//one player, but it could be multiple in the case of tie, or 0 in the
//case of a draw. Will return nil if Finished() is not yet true.
func (g *Game) Winners() []PlayerIndex {
	return g.winners
}

//Finished is whether the came has been completed. If it is over, the Winners
//will be set. A game is finished when GameDelegate.CheckGameFinished()
//returns true. Once a game is Finished it may never be un-finished, and no
//more moves may ever be applied to it.
func (g *Game) Finished() bool {
	return g.finished
}

//Manager is a reference to the GameManager that controls this game.
func (g *Game) Manager() *GameManager {
	return g.manager
}

//NumPlayers returns the number of players for this game, based on how many
//PlayerStates are in CurrentState. Note that if your game logic is complex,
//this is likely NOT what you want, instead you might want
//GameDelegate.NumSeatedActivePlayers. See the package doc of
//boardgame/behaviors for more.
func (g *Game) NumPlayers() int {
	return g.numPlayers
}

//JSONForPlayer returns an object appropriate for being json'd via
//json.Marshal. The object is the equivalent to what MarshalJSON would output,
//only as an object, and with state sanitized for the current player. State
//should be a state for this game (e.g. an old version). If state is nil, the
//game's CurrentState will be used. This is effectively equivalent to
//state.SanitizeForPlayer().
func (g *Game) JSONForPlayer(player PlayerIndex, state ImmutableState) interface{} {

	if state == nil {
		state = g.CurrentState()
	}

	state = state.SanitizedForPlayer(player)

	//We deliberately never include SecretSalt in the JSON blobs we create.

	return map[string]interface{}{
		"Name":               g.Name(),
		"Finished":           g.Finished(),
		"Winners":            g.Winners(),
		"CurrentState":       state,
		"CurrentPlayerIndex": g.manager.delegate.CurrentPlayerIndex(state),
		"Diagram":            state.Diagram(),
		"ID":                 g.ID(),
		"NumPlayers":         g.NumPlayers(),
		"Agents":             g.Agents(),
		"Variant":            g.Variant(),
		"Version":            g.Version(),
		"ActiveTimers":       g.manager.timers.ActiveTimersForGame(g.ID()),
	}
}

//MarshalJSON returns a marshaled version of the output of JSONForPlayer for
//AdminPlayerIndex.
func (g *Game) MarshalJSON() ([]byte, error) {
	//We define our own MarshalJSON because if we didn't there'd be an infinite loop because of the redirects back up.
	return json.Marshal(g.JSONForPlayer(AdminPlayerIndex, nil))
}

//StorageRecord returns a GameStorageRecord representing the aspects of this
//game that should be serialized to storage.
func (g *Game) StorageRecord() *GameStorageRecord {

	return &GameStorageRecord{
		Name:       g.Manager().Delegate().Name(),
		Version:    g.Version(),
		Winners:    g.Winners(),
		Finished:   g.Finished(),
		Created:    g.Created(),
		Modified:   g.Modified(),
		ID:         g.ID(),
		SecretSalt: g.secretSalt,
		NumPlayers: g.NumPlayers(),
		Agents:     g.Agents(),
		Variant:    g.Variant(),
	}
}

//Name returns the name of this game type. Convenience method for
//game.Manager().Delegate().Name().
func (g *Game) Name() string {
	return g.manager.Delegate().Name()
}

//ID returns the unique id string that corresponds to this particular game.
//The ID is used in URLs and to retrieve this particular game from storage.
func (g *Game) ID() string {
	return g.id
}

//Agents returns the agent configuration for the game.
func (g *Game) Agents() []string {
	return g.agents
}

//Version returns the version number of the highest State that is stored for
//this game. This number will increase by one every time a move is applied.
func (g *Game) Version() int {
	return g.version
}

//CurrentState returns the state object for the current state. Equivalent,
//semantically, to game.State(game.Version())
func (g *Game) CurrentState() ImmutableState {
	if g.cachedCurrentState == nil {
		g.cachedCurrentState = g.State(g.Version())
	}
	return g.cachedCurrentState
}

//State returns the state of the game at the given version. Because states can
//only be modffied in moves, the state returned is immutable.
func (g *Game) State(version int) ImmutableState {

	if version < 0 || version > g.Version() {
		return nil
	}

	record, err := g.manager.Storage().State(g.ID(), version)

	if err != nil {
		g.manager.Logger().WithField("version", version).Error("State retrieval failed" + err.Error())
		return nil
	}

	result, err := g.manager.stateFromRecord(record, version)

	if err != nil {
		g.manager.Logger().Error("StateFromBlob failed: " + err.Error())
		return nil
	}

	result.game = g

	return result

}

//Move returns the Move that was applied to get the Game to the given version;
//an inflated version of the MoveStorageRecord. Not to be confused with
//Moves(), which returns examples of moves that haven't yet been applied, but
//have their defaults set based on the current state.
func (g *Game) Move(version int) (Move, error) {

	if version < 0 || version > g.Version() {
		return nil, errors.New("Invalid version")
	}

	record, err := g.manager.Storage().Move(g.ID(), version)

	if err != nil {
		return nil, errors.New("State retrieval failed" + err.Error() + strconv.Itoa(version))
	}

	if record == nil {
		return nil, errors.New("No such record")
	}

	if record.Version != version {
		return nil, errors.New("the version of the returned move was not what was expected")
	}

	return record.inflate(g)

}

//MoveRecords returns all of the move storage records up to upToVersion, in
//ascending order. If upToVersion is 0 or less, game.Version() will be used
//for upToVersion. It is cached so repeated calls should be fast. This is a
//wrapper around game.Manager().Storage().Moves(), cached for performance.
func (g *Game) MoveRecords(upToVersion int) []*MoveStorageRecord {

	if upToVersion < 1 {
		upToVersion = g.Version()
	}

	if upToVersion == 0 {
		return nil
	}

	//g.cachedHistoricalMoves is of ALL moves. If it doesn't exist, fetch it.
	if g.cachedHistoricalMoves == nil {

		//Our cache is of ALL moves.
		moves, err := g.manager.Storage().Moves(g.ID(), 0, g.Version())

		if err != nil {
			g.Manager().Logger().Errorln("Fetching moves failed: " + err.Error())
			return nil
		}

		g.cachedHistoricalMoves = moves
	}

	//g.cacheHistoricalMoves is 1-indexed, since there are no moves for
	//version 1. Because go slice indexing is up to but not including upper
	//bound, we can leave it as is to get the desired behavior.
	return g.cachedHistoricalMoves[:upToVersion]

}

//NumAgentPlayers returns the number of players who have agents configured on
//them.
func (g *Game) NumAgentPlayers() int {

	if !g.initalized {
		return 0
	}

	result := 0

	for _, agent := range g.agents {
		if agent != "" {
			result++
		}
	}

	return result

}

//starterState returns a starting, not-yet-saved State that is configured with all moving parts.
func (g *Game) starterState(numPlayers int) (State, error) {
	state, err := g.Manager().emptyState(numPlayers)

	if err != nil {
		return nil, err
	}

	state.game = g

	return state, nil
}

//SetUp initializes a specific game object and gets it ready for the first
//move to apply. SetUp must be called before ProposeMove can be called. Even
//if an error is returned, the game should be in a consistent state. If
//numPlayers is 0, we will use delegate.DefaultNumPlayers(). Variant may be
//nil; the values will be passed to NewVariant if agentNames is not nil, it
//should have len(numPlayers). The strings in each index represent the agent
//to install for that player (empty strings mean a human player).
func (g *Game) setUp(numPlayers int, variantValues map[string]string, agentNames []string) error {

	baseErr := errors.NewFriendly("Game couldn't be set up")

	if g.initalized {
		return baseErr.WithError("Game already initalized")
	}

	//TODO: we don't need this anymore because managers can't be created without chests.
	if g.manager.Chest() == nil {
		return baseErr.WithError("No component chest set on manager")
	}

	if numPlayers == 0 {
		numPlayers = g.manager.Delegate().DefaultNumPlayers()
	}

	if numPlayers < 1 {
		return errors.NewFriendly("The number of players, " + strconv.Itoa(numPlayers) + " is not legal. There must be one or more players.")
	}

	if !g.manager.Delegate().LegalNumPlayers(numPlayers) {
		return errors.NewFriendly("The number of players, " + strconv.Itoa(numPlayers) + " was not legal.")
	}

	variant, err := g.manager.Variants().NewVariant(variantValues)

	if err != nil {
		return errors.NewFriendly("That variation is not legal for this game: " + err.Error())
	}

	g.variant = variant

	if agentNames != nil && len(agentNames) != numPlayers {
		return baseErr.WithError("If agentNames is not nil, it must have length equivalent to numPlayers.")
	}

	if agentNames == nil {
		agentNames = make([]string, numPlayers)
	}

	g.agents = agentNames

	g.numPlayers = numPlayers

	stateCopy, err := g.starterState(numPlayers)

	if err != nil {
		return errors.Extend(err, "Couldn't get starter state")
	}

	//Make a starter one so that buildComponentIndex doesn't get called.
	stateCopy.(*state).componentIndex = make(map[Component]componentIndexItem)

	if err := g.manager.delegate.BeginSetUp(stateCopy, variant); err != nil {
		return errors.New("BeginSetUp errored: " + err.Error())
	}

	//Distribute all components to their starter locations

	for _, name := range g.Manager().Chest().DeckNames() {
		deck := g.Manager().Chest().Deck(name)
		for i, component := range deck.Components() {
			stack, err := g.manager.Delegate().DistributeComponentToStarterStack(stateCopy, component)
			if err != nil {
				return baseErr.WithError("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ":" + err.Error())
			}
			if stack == nil {
				return baseErr.WithError("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ": the delegate returned no stack.")
			}
			if stack.SlotsRemaining() < 1 {
				return baseErr.WithError("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ": the stack the delegate returned had no more slots.")
			}

			mutableStack, ok := stack.(Stack)

			if !ok {
				return baseErr.WithError("Couldn't get a mutable version of stack")
			}

			mutableStack.insertComponentAt(mutableStack.nextSlot(), component.ImmutableInstance(stateCopy))
		}
	}

	if err := g.manager.delegate.FinishSetUp(stateCopy); err != nil {
		return errors.New("FinishSetUp errored: " + err.Error())
	}

	g.created = time.Now()
	g.modified = time.Now()

	if g.Modifiable() {

		//Save the initial state to DB.
		if err := g.manager.Storage().SaveGameAndCurrentState(g.StorageRecord(), stateCopy.StorageRecord(), nil); err != nil {
			return baseErr.WithError("Storage failed: " + err.Error())
		}
	}

	g.initalized = true

	for i, name := range g.agents {
		if name == "" {
			continue
		}
		agent := g.Manager().AgentByName(name)

		if agent == nil {
			return baseErr.WithError("Couldn't find the agent for the " + strconv.Itoa(i) + " player: " + name)
		}

		agentState := agent.SetUpForGame(g, PlayerIndex(i))

		if agentState == nil {
			continue
		}

		if err := g.Manager().storage.SaveAgentState(g.ID(), PlayerIndex(i), agentState); err != nil {
			return baseErr.WithError("Couldn't save state for agent " + strconv.Itoa(i) + ": " + err.Error())
		}
	}

	//See if any fixup moves apply

	//TODO: test that fixup moves are applied at the beginning.

	move := g.manager.Delegate().ProposeFixUpMove(stateCopy)

	if move != nil {
		//We apply the move immediately. This ensures that when
		//DelayedError resolves, all of the fix up moves have been
		//applied.
		if err := g.applyMove(move, AdminPlayerIndex, true, 0, selfInitiatorSentinel); err != nil {

			if err == ErrTooManyFixUps {
				return err
			}

			//TODO: if we bail here, we haven't left Game in a consistent
			//state because we haven't rolled back what we did.
			return baseErr.WithError("Applying the first fix up move failed: " + err.Error())
		}
	}

	//TODO: start up agents.

	if g.Modifiable() {

		//Can't start this until now, otherwise we could have a race.
		go g.mainLoop()
	}

	return nil
}

//triggerFixUp signals that we want to ensure that a fixUp loop runs even if no
//moves have been made, because some state that a move relies on outside of game
//state has changed.
func (g *Game) triggerFixUp() DelayedError {
	//If we aren't a modifiable copy then we need to dispatch to the one that is

	delayed := make(DelayedError)

	if !g.modifiable {
		game := g.manager.ModifiableGame(g.ID())
		game.fixUpTriggered <- delayed
	} else {
		g.fixUpTriggered <- delayed
	}
	return delayed
}

//MainLoop should be run in a goroutine. It is what takes moves off of
//proposedMoves and applies them. It is the only method that may call
//applyMove.
func (g *Game) mainLoop() {
	for {
		select {
		case item := <-g.proposedMoves:
			if item == nil {
				return
			}
			item.ch <- g.applyMove(item.move, item.proposer, false, 0, selfInitiatorSentinel)
			close(item.ch)
		case delayed := <-g.fixUpTriggered:
			move := g.manager.delegate.ProposeFixUpMove(g.CurrentState())
			if move == nil {
				delayed <- nil
			} else {
				proposedDelayed := g.ProposeMove(move, AdminPlayerIndex)
				//We can't wait for the error here, because the mainLoop needs
				//to keep chugging to process the move we just put in the queue
				go func() {
					delayed <- (<-proposedDelayed)
				}()
			}
		}
	}
}

//Modifiable returns true if this instantiation of the game can be modified.
//Games that are created via GameManager.NewGame() or retrieved from
//GameManager.Game() can be modified directly via ProposeMove, and the game
//object will be updated as those changes are made. Games that return
//Modifiable() false can still have ProposeMove called on them; they will
//simply forward the move to a game for this Id that is modifiable.
func (g *Game) Modifiable() bool {
	return g.modifiable
}

//Moves returns an array of all Moves with their defaults set for this current
//state. This method is useful for getting a list of all moves that could
//possibly be applied to the game at its current state.
//base.GameDelegate.ProposeFixUpMove uses this. Not to be confused with
//Move(), which returns an inflated version of a move that has already been
//succdefully applied to this game in the past.
func (g *Game) Moves() []Move {

	if !g.initalized {
		return nil
	}

	types := g.manager.moveTypes()

	result := make([]Move, len(types))

	for i, moveType := range types {
		result[i] = moveType.NewMove(g.CurrentState())
	}
	return result
}

//MoveByName returns a move of the given name set to reasonable defaults for
//the game at its current state. Moves() is similar to this, but returns all
//moves.
func (g *Game) MoveByName(name string) Move {
	if !g.initalized {
		return nil
	}

	moveType := g.manager.moveTypeByName(name)

	if moveType == nil {
		return nil
	}

	return moveType.NewMove(g.CurrentState())
}

//Refresh goes and sets this game object to reflect the current state of the
//underlying game in Storage. Basically, when you call manager.Game() you get
//a snapshot of the game in storage at that moment. If you believe that the
//underlying game in storage has been modified, calling Refresh() will re-load
//the snapshot, effectively. You only have to do this if you suspect that a
//modifiable version of this game somewhere in another application binary
//that's currently running may have changed since this game object was
//created. You don't need to call this after calling ProposeMove, even on non-
//modifiable games; it will have been called for you already. If you only have
//one instance of your application binary running at a time, you never need to
//call this.
func (g *Game) Refresh() {

	freshGame := g.manager.Game(g.ID())

	g.cachedCurrentState = nil
	g.cachedHistoricalMoves = nil
	g.version = freshGame.Version()
	g.finished = freshGame.Finished()
	g.winners = freshGame.Winners()

}

//ProposeMove is the way to propose a move to the game. DelayedError will return
//an error in the future if the move was unable to be applied, or nil if the
//move was applied successfully. Proposer is the PlayerIndex of the player who
//is notionally proposing the move. If you don't know which player is moving it,
//AdminPlayerIndex is a reasonable default that will generally allow any move to
//be made. After the move is applied, your GameDelegate's ProposeFixUpMove will
//be called; if any move is returned it will be applied, repeating the cycle
//until no moves are returned from ProposeFixUpMove. DelayedError will only
//resolve once any applicable FixUp moves have been applied already. This is
//legal to call on a non-modifiable game--the change will be dispatched to a
//modifiable version of the game with this ID, and afterwards this Game object's
//state will be updated in place with the new values after the change (by
//automatically calling Refresh()).
func (g *Game) ProposeMove(move Move, proposer PlayerIndex) DelayedError {

	if !g.Modifiable() {
		return g.manager.proposeMoveOnGame(g, move, proposer)
	}

	errChan := make(DelayedError, 1)

	workItem := &proposedMoveItem{
		move:     move,
		proposer: proposer,
		ch:       errChan,
	}

	if !g.initalized {
		//The channel isn't even ready to send one.
		errChan <- errors.New("[roposed a move before the game had been successfully set-up")
		return errChan
	}

	g.proposedMoves <- workItem

	return errChan

}

//triggerAgents is called after a PlayerMove (and its chain of fixUp moves) is called.
func (g *Game) triggerAgents() error {

	if g.Finished() {
		return nil
	}

	for i, name := range g.agents {

		if name == "" {
			continue
		}

		agent := g.Manager().AgentByName(name)

		if agent == nil {
			return errors.New("Couldn't find agent for #" + strconv.Itoa(i) + ": " + name)
		}

		agentState, err := g.Manager().Storage().AgentState(g.ID(), PlayerIndex(i))

		if err != nil {
			return errors.New("Couldn't load state for agent #" + strconv.Itoa(i) + ": " + err.Error())
		}

		move, newState := agent.ProposeMove(g, PlayerIndex(i), agentState)

		if newState != nil {
			if err := g.Manager().Storage().SaveAgentState(g.ID(), PlayerIndex(i), newState); err != nil {
				return errors.New("Failed to store new state for agent #" + strconv.Itoa(i) + ": " + err.Error())
			}
		}

		if move != nil {

			//Slow down the playback of moves to more accurately emulate a human.

			//TODO: if it's already been awhile since the last move was made
			//(e.g. the agent was thinking for awhile), then apply
			//immediately.

			if g.instantAgentMoves {
				g.ProposeMove(move, PlayerIndex(i))
			} else {
				g.delayedProposeMove(move, PlayerIndex(i), 500*time.Millisecond, 2*time.Second)
			}
		}
	}
	return nil
}

func (g *Game) delayedProposeMove(move Move, proposer PlayerIndex, low time.Duration, high time.Duration) {

	diff := high - low

	timeToWait := time.Duration(rand.Intn(int(diff))) + low
	go func() {
		<-time.After(timeToWait)
		g.ProposeMove(move, proposer)
	}()
}

//Game applies the move to the state if it is currently legal. May only be
//called by mainLoop. Propose moves with game.ProposeMove instead.
func (g *Game) applyMove(move Move, proposer PlayerIndex, isFixUp bool, recurseCount int, initiator int) error {

	baseErr := errors.NewFriendly("The move could not be made")

	versionToSet := g.version + 1

	if !g.initalized {
		return baseErr.WithError("The game has not been initalized.")
	}

	if g.finished {
		return errors.NewFriendly("Game was already finished")
	}

	if g.MoveByName(move.Info().Name()) == nil {
		return baseErr.WithError("That move is not configured for this game.")
	}

	if initiator == selfInitiatorSentinel {
		//If we were passed the selfInitiatorSentinel that means that it's the
		//start of a causal chain and our initiator should be what our version
		//will be.
		initiator = versionToSet
	}

	currentState := g.CurrentState().(*state)

	if !proposer.Valid(currentState) {
		return baseErr.WithError("The proposer was not valid.")
	}

	if proposer == ObserverPlayerIndex {
		return baseErr.WithError("The proposer was the ObserverPlayerIndex, but observers may never make moves.")
	}

	move.Info().initiator = initiator
	move.Info().timestamp = time.Now()
	move.Info().version = versionToSet

	if err := move.Legal(currentState, proposer); err != nil {
		//It's not legal, reject.
		return errors.NewFriendly(err.Error())
	}

	currentPhase := g.manager.delegate.CurrentPhase(currentState)

	newState, err := currentState.copy(false)

	if err != nil {
		return baseErr.WithError("There was an internal error copying the state: " + err.Error())
	}

	newState.version = versionToSet

	if err := move.Apply(newState); err != nil {
		return baseErr.WithError("The move's apply function returned an error:" + err.Error())
	}

	if err := newState.validateBeforeSave(); err != nil {
		return baseErr.WithError("The modified state had an invalidity, so the move was not applied. " + err.Error())
	}

	//Check to see if that move made the game finished.

	finished, winners := g.manager.Delegate().CheckGameFinished(newState)

	if finished {
		g.finished = true
		g.winners = winners
		//TODO: persist to database here.
	}

	g.version = versionToSet

	//Expire the currentState cache; it's no longer valid.
	g.cachedCurrentState = nil

	//Note that we want the phase that we were in BEFORE this move was applied.
	moveStorageRecord := StorageRecordForMove(move, currentPhase, proposer)

	//use the precise time we'll set for the move.
	g.modified = move.Info().Timestamp()

	//TODO: test that if we fail to save state to storage everything's fine.
	if err := g.manager.Storage().SaveGameAndCurrentState(g.StorageRecord(), newState.StorageRecord(), moveStorageRecord); err != nil {
		//TODO: we need to undo the temporary changes we made directly to ourselves (vesrion, finished, winners)
		return baseErr.WithError("Storage returned an error:" + err.Error())
	}

	//Ok, the state stuck and is now canonical--trigger the actions it was
	//supposed to do.
	newState.committed()

	if recurseCount > maxRecurseCount {
		return ErrTooManyFixUps
	}

	if g.finished {

		if !isFixUp {
			g.manager.Storage().PlayerMoveApplied(g.StorageRecord())
		}

		return nil
	}

	//if the cache is not nil OR it's the first move, we can just append the
	//move storage record to the cache.
	if g.cachedHistoricalMoves != nil || versionToSet == 1 {
		g.cachedHistoricalMoves = append(g.cachedHistoricalMoves, moveStorageRecord)
	}

	move = g.manager.Delegate().ProposeFixUpMove(newState)

	if move != nil {
		//We apply the move immediately. This ensures that when
		//DelayedError resolves, all of the fix up moves have been
		//applied.
		if err := g.applyMove(move, AdminPlayerIndex, true, recurseCount+1, initiator); err != nil {

			if err == ErrTooManyFixUps {
				return err
			}

			//TODO: if we bail here, we haven't left Game in a consistent
			//state because we haven't rolled back what we did.
			return baseErr.WithError("Applying the fix up move failed: " + strconv.Itoa(recurseCount) + ": " + err.Error())
		}
	}

	if err := g.triggerAgents(); err != nil {
		return baseErr.WithError("Failed to trigger agent: " + err.Error())
	}

	//We only want to alert that the run is done if it was a player move that
	//was applied.
	if !isFixUp {
		g.manager.Storage().PlayerMoveApplied(g.StorageRecord())
	}

	return nil

}
