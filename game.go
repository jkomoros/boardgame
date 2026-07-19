package boardgame

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
)

// gameIdleTimeout is how long a modifiable game's mainLoop will wait with no
// activity before freezing itself (exiting the goroutine and removing itself
// from the warm cache). A frozen game is transparently reloaded from storage
// on the next access.
const gameIdleTimeout = 10 * time.Minute

// maxResidentGames is the maximum number of modifiable games that can be
// resident in memory (warm cache) at once. When the cache is full and a new
// game needs to be loaded, the least recently active game is evicted.
const maxResidentGames = 128

// maxRecurseCount is the number of fixUp moves that can be considered normal--
// anything more than that and we'll return an error because the delegate is
// likely going to return fixup moves forever.
const maxRecurseCount = 256

const selfInitiatorSentinel = -1

// ErrTooManyFixUps is returned from game.ProposeMove if too many fix up moves
// are applied, which implies that there is a FixUp move configured to always
// be legal, and is evidence of a serious error in your game logic.
var ErrTooManyFixUps = errors.New("we recursed deeply in fixup, which implies that ProposeFixUp has a move that is always legal")

// A Game represents a specific game between a collection of Players; an
// instantiation of a game of the given type. Create a new one with
// GameManager.NewGame().
type Game struct {
	manager *GameManager

	finished bool

	winners []PlayerIndex

	agents []string

	//The current version of State.
	version int

	// publishedHeadVersion and proposalFrontierVersion are updated in an order
	// that can transiently produce only false negatives. Server projection reads
	// them atomically instead of racing with the main loop's legacy version field.
	publishedHeadVersion    atomic.Int64
	proposalFrontierVersion atomic.Int64

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

	//done is closed to signal mainLoop to exit. Used to prevent goroutine
	//leaks when a game is frozen.
	done chan struct{}
	//freezeOnce ensures done is closed exactly once, even if multiple
	//goroutines try to freeze the game concurrently.
	freezeOnce sync.Once
	//frozen is true after the game has been frozen (mainLoop exited, removed
	//from cache). A frozen game must be reloaded from storage.
	frozen bool
	//lastActivity is updated each time a move is proposed or the game is
	//accessed via ModifiableGame. Used for idle timeout and LRU eviction.
	lastActivity time.Time

	//if true, we will not wait to propose agent moves (mainly used for
	//testing.)
	instantAgentMoves bool

	//Initalized is set to True after SetUp is called.
	initalized bool

	//allowMutableConstraints bypasses the initalized check for
	//AddConstraint/ClearConstraints. Set via ManagerInternals for testing.
	allowMutableConstraints bool

	created  time.Time
	modified time.Time

	//TODO: HistoricalState(index int) and HistoryLen() int

	//TODO: an array of Player objects.

	// legalMemoMu guards every field below it: both the field-independent
	// legality memo and the move-tape memo (design spec §5, legal_memo.go).
	// Unlike cachedCurrentState/cachedHistoricalMoves above (which are only
	// ever touched from Game.mainLoop's single goroutine), these memos are
	// also written from move.Legal() evaluation reachable off mainLoop —
	// e.g. server/api's generateFormsWithLegality calling move.Legal() from
	// an HTTP-handler goroutine concurrently with mainLoop's own Legal()
	// evaluation during fixups. Two goroutines read/writing the same Go map
	// concurrently is a fatal runtime crash (not just a benign race), so
	// these maps need real synchronization, unlike the pointer-assignment
	// caches above. See legal_memo.go's lock-ordering note: this mutex must
	// never be held while evaluating a predicate or calling a caller-
	// supplied compute() closure (arbitrary user code) — only the map/field
	// reads and writes themselves are done under the lock.
	legalMemoMu sync.Mutex

	// legalFieldIndepMemo/legalFieldIndepMemoVersion back the
	// field-independent legality memo (design spec §5, legal_memo.go),
	// bounded to at most the current head version's worth of entries.
	// Guarded by legalMemoMu.
	legalFieldIndepMemo        map[legalFieldIndepMemoKey]LegalVerdict
	legalFieldIndepMemoVersion int

	// legalTapeMemo/legalTapeMemoVersion/legalTapeMemoPhase/
	// legalTapeMemoValid back the move-tape memo (design spec §5,
	// legal_memo.go's LegalTapeMemo), bounded to at most one (version,
	// phase) pair's worth of a cached tape at a time. Guarded by
	// legalMemoMu.
	legalTapeMemo        []*MoveStorageRecord
	legalTapeMemoVersion int
	legalTapeMemoPhase   enum.EnumKey
	legalTapeMemoValid   bool
}

const gameIDLength = 16

// DelayedError is a chan on which an error (or nil) will be sent at a later
// time. Primarily returned from game.ProposeMove(), so the method can return
// immediately even before the move is processed, which might take a long time
// if there are many moves ahead in the queue.
type DelayedError chan error

type proposedMoveItem struct {
	move            Move
	proposer        PlayerIndex
	expectedVersion *int
	//Ch is the channel we should either return an error on and then close, or
	//send nil and close.
	ch DelayedError
}

// StaleVersionError means a proposal was created from a state version that is
// no longer current. The check happens inside the game's serialized main loop,
// immediately before legality and application, so concurrent HTTP requests
// cannot both pass a racy handler-level version check.
type StaleVersionError struct {
	Expected int
	Actual   int
}

func (e *StaleVersionError) Error() string {
	return "move proposal used stale game version " + strconv.Itoa(e.Expected) + "; current version is " + strconv.Itoa(e.Actual)
}

var defaultStringRand *rand.Rand

func init() {

	defaultStringRand = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

}

const randomStringChars = "ABCDEF0123456789"

// randomString returns a random string of the given length. If rand is not
// nil, will use that source. Ohterwise will use a global source.
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

// Created returns the time stamp when this game was first created.
func (g *Game) Created() time.Time {
	return g.created
}

// Modified returns the timstamp when the last move was applied to this game.
func (g *Game) Modified() time.Time {
	return g.modified
}

// Variant returns a copy of the Variant passed to NewGame to create this
// game originally.
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

// Winners is the player indexes who were winners. Typically, this will be
// one player, but it could be multiple in the case of tie, or 0 in the
// case of a draw. Will return nil if Finished() is not yet true.
func (g *Game) Winners() []PlayerIndex {
	return g.winners
}

// Finished is whether the came has been completed. If it is over, the Winners
// will be set. A game is finished when GameDelegate.CheckGameFinished()
// returns true. Once a game is Finished it may never be un-finished, and no
// more moves may ever be applied to it.
func (g *Game) Finished() bool {
	return g.finished
}

// Manager is a reference to the GameManager that controls this game.
func (g *Game) Manager() *GameManager {
	return g.manager
}

// NumPlayers returns the number of players for this game, based on how many
// PlayerStates are in CurrentState. Note that if your game logic is complex,
// this is likely NOT what you want, instead you might want
// GameDelegate.NumSeatedActivePlayers. See the package doc of
// boardgame/behaviors for more.
func (g *Game) NumPlayers() int {
	return g.numPlayers
}

// JSONForPlayer returns an object appropriate for being json'd via
// json.Marshal. The object is the equivalent to what MarshalJSON would output,
// only as an object, and with state sanitized for the current player. State
// should be a state for this game (e.g. an old version). If state is nil, the
// game's CurrentState will be used. This is effectively equivalent to
// state.SanitizeForPlayer().
func (g *Game) JSONForPlayer(player PlayerIndex, state ImmutableState) (interface{}, error) {

	if state == nil {
		state = g.CurrentState()
	}

	state, err := state.SanitizedForPlayer(player)

	if err != nil {
		return nil, errors.New("Couldn't sanitize state: " + err.Error())
	}

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
	}, nil
}

// MarshalJSON returns a marshaled version of the output of JSONForPlayer for
// AdminPlayerIndex.
func (g *Game) MarshalJSON() ([]byte, error) {
	//We define our own MarshalJSON because if we didn't there'd be an infinite loop because of the redirects back up.
	val, err := g.JSONForPlayer(AdminPlayerIndex, nil)
	if err != nil {
		return nil, err
	}
	return json.Marshal(val)
}

// StorageRecord returns a GameStorageRecord representing the aspects of this
// game that should be serialized to storage.
func (g *Game) StorageRecord() *GameStorageRecord {

	return &GameStorageRecord{
		Name:                    g.Manager().Delegate().Name(),
		Version:                 g.Version(),
		ProposalFrontierVersion: g.ProposalFrontierVersion(),
		ProposalFrontierKnown:   g.AtProposalFrontier(),
		Winners:                 g.Winners(),
		Finished:                g.Finished(),
		Created:                 g.Created(),
		Modified:                g.Modified(),
		ID:                      g.ID(),
		SecretSalt:              g.secretSalt,
		NumPlayers:              g.NumPlayers(),
		Agents:                  g.Agents(),
		Variant:                 g.Variant(),
	}
}

// Name returns the name of this game type. Convenience method for
// game.Manager().Delegate().Name().
func (g *Game) Name() string {
	return g.manager.Delegate().Name()
}

// ID returns the unique id string that corresponds to this particular game.
// The ID is used in URLs and to retrieve this particular game from storage.
func (g *Game) ID() string {
	return g.id
}

// Agents returns the agent configuration for the game.
func (g *Game) Agents() []string {
	return g.agents
}

// Version returns the version number of the highest State that is stored for
// this game. This number will increase by one every time a move is applied.
func (g *Game) Version() int {
	return g.version
}

// ProposalFrontierVersion returns the last state version at which the game had
// completed a serialized proposal and its full recursive fix-up chain. A
// newer durable state is intermediate history, not yet a safe source of
// actions for the next player proposal.
func (g *Game) ProposalFrontierVersion() int {
	return int(g.proposalFrontierVersion.Load())
}

// AtProposalFrontier reports whether the current state is ready to advertise
// and accept the next player proposal.
func (g *Game) AtProposalFrontier() bool {
	_, ok := g.proposalFrontierSnapshot()
	return ok
}

func (g *Game) proposalFrontierSnapshot() (int, bool) {
	head := int(g.publishedHeadVersion.Load())
	frontier := g.ProposalFrontierVersion()
	return frontier, head >= 0 && head == frontier
}

func (g *Game) publishHeadVersion(version int) {
	g.publishedHeadVersion.Store(int64(version))
}

func (g *Game) invalidateProposalFrontier(persist bool) error {
	head := int(g.publishedHeadVersion.Load())
	g.proposalFrontierVersion.Store(-1)
	if !persist {
		return nil
	}
	if storage, ok := g.manager.Storage().(ProposalFrontierStorage); ok {
		return storage.SaveProposalFrontier(g.ID(), head, -1)
	}
	return nil
}

func (g *Game) markProposalFrontier() error {
	head := int(g.publishedHeadVersion.Load())
	if storage, ok := g.manager.Storage().(ProposalFrontierStorage); ok {
		if err := storage.SaveProposalFrontier(g.ID(), head, head); err != nil {
			g.proposalFrontierVersion.Store(-1)
			return err
		}
	}
	g.proposalFrontierVersion.Store(int64(head))
	return nil
}

// CurrentState returns the state object for the current state. Equivalent,
// semantically, to game.State(game.Version())
func (g *Game) CurrentState() ImmutableState {
	if g.cachedCurrentState == nil {
		g.cachedCurrentState = g.State(g.Version())
	}
	return g.cachedCurrentState
}

// State returns the state of the game at the given version. Because states can
// only be modffied in moves, the state returned is immutable.
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

// Move returns the Move that was applied to get the Game to the given version;
// an inflated version of the MoveStorageRecord. Not to be confused with
// Moves(), which returns examples of moves that haven't yet been applied, but
// have their defaults set based on the current state.
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

// MoveRecords returns all of the move storage records up to upToVersion, in
// ascending order. If upToVersion is 0 or less, game.Version() will be used
// for upToVersion. It is cached so repeated calls should be fast. This is a
// wrapper around game.Manager().Storage().Moves(), cached for performance.
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

// NumAgentPlayers returns the number of players who have agents configured on
// them.
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

// starterState returns a starting, not-yet-saved State that is configured with all moving parts.
func (g *Game) starterState(numPlayers int) (State, error) {
	state, err := g.Manager().emptyState(numPlayers)

	if err != nil {
		return nil, err
	}

	state.game = g

	return state, nil
}

// SetUp initializes a specific game object and gets it ready for the first
// move to apply. SetUp must be called before ProposeMove can be called. Even
// if an error is returned, the game should be in a consistent state. If
// numPlayers is 0, we will use delegate.DefaultNumPlayers(). Variant may be
// nil; the values will be passed to NewVariant if agentNames is not nil, it
// should have len(numPlayers). The strings in each index represent the agent
// to install for that player (empty strings mean a human player).
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
			if err := stateCopy.(*state).validateStackAttachment(mutableStack); err != nil {
				return baseErr.WithError("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ": returned stack is not attached: " + err.Error())
			}
			if mutableStack.Deck() != component.Deck() {
				return baseErr.WithError("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ": returned stack belongs to a different deck")
			}

			mutableStack.insertComponentAt(mutableStack.nextSlot(), component.ImmutableInstance(stateCopy))
		}
	}

	if err := g.manager.delegate.FinishSetUp(stateCopy); err != nil {
		return errors.New("FinishSetUp errored: " + err.Error())
	}
	if err := stateCopy.(*state).validateComponentConservation(); err != nil {
		return baseErr.WithError("Initial state violated component conservation: " + err.Error())
	}

	g.created = time.Now()
	g.modified = time.Now()
	g.lastActivity = time.Now()

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

	// Setup is one serialized proposal boundary too: only advertise the
	// resulting state after every setup fix-up has completed successfully.
	if err := g.markProposalFrontier(); err != nil {
		return baseErr.WithError("Couldn't persist the setup proposal frontier: " + err.Error())
	}

	//TODO: start up agents.

	if g.Modifiable() {

		//Can't start this until now, otherwise we could have a race.
		go g.mainLoop()
	}

	return nil
}

// triggerFixUp signals that we want to ensure that a fixUp loop runs even if no
// moves have been made, because some state that a move relies on outside of game
// state has changed.
func (g *Game) triggerFixUp() DelayedError {
	//If we aren't a modifiable copy then we need to dispatch to the one that is

	delayed := make(DelayedError)

	if !g.modifiable {
		game := g.manager.ModifiableGame(g.ID())
		if game == nil {
			delayed <- errors.New("could not find modifiable game")
			return delayed
		}
		if err := game.invalidateProposalFrontier(true); err != nil {
			go func() { delayed <- err }()
			return delayed
		}
		select {
		case game.fixUpTriggered <- delayed:
		case <-game.done:
			delayed <- errors.New("game has been frozen")
		}
	} else {
		if err := g.invalidateProposalFrontier(true); err != nil {
			go func() { delayed <- err }()
			return delayed
		}
		select {
		case g.fixUpTriggered <- delayed:
		case <-g.done:
			delayed <- errors.New("game has been frozen")
		}
	}
	return delayed
}

// MainLoop should be run in a goroutine. It is what takes moves off of
// proposedMoves and applies them. It is the only method that may call
// applyMove. It will exit after gameIdleTimeout of inactivity, freezing the
// game and removing it from the warm cache.
func (g *Game) mainLoop() {
	idleTimer := time.NewTimer(gameIdleTimeout)
	defer idleTimer.Stop()
	resetTimer := func() {
		if !idleTimer.Stop() {
			select {
			case <-idleTimer.C:
			default:
			}
		}
		idleTimer.Reset(gameIdleTimeout)
	}
	for {
		select {
		case item := <-g.proposedMoves:
			if item == nil {
				return
			}
			resetTimer()
			item.ch <- g.applyProposedMove(item)
			close(item.ch)
		case delayed := <-g.fixUpTriggered:
			resetTimer()
			// applyMove deliberately ends a proposal as soon as the game becomes
			// finished, without asking the delegate for another fix-up. Recovery of
			// an unknown marker on that terminal head must follow the same rule.
			if g.Finished() {
				delayed <- g.markProposalFrontier()
				continue
			}
			move := g.manager.delegate.ProposeFixUpMove(g.CurrentState())
			if move == nil {
				delayed <- g.markProposalFrontier()
			} else {
				proposedDelayed := g.ProposeMove(move, AdminPlayerIndex)
				//We can't wait for the error here, because the mainLoop needs
				//to keep chugging to process the move we just put in the queue
				go func() {
					delayed <- (<-proposedDelayed)
				}()
			}
		case <-idleTimer.C:
			// Double-check no pending work before freezing
			select {
			case item := <-g.proposedMoves:
				if item == nil {
					return
				}
				idleTimer.Reset(gameIdleTimeout)
				item.ch <- g.applyProposedMove(item)
				close(item.ch)
			default:
				g.manager.freezeGame(g)
				return
			}
		case <-g.done:
			return
		}
	}
}

// Modifiable returns true if this instantiation of the game can be modified.
// Games that are created via GameManager.NewGame() or retrieved from
// GameManager.Game() can be modified directly via ProposeMove, and the game
// object will be updated as those changes are made. Games that return
// Modifiable() false can still have ProposeMove called on them; they will
// simply forward the move to a game for this Id that is modifiable.
func (g *Game) Modifiable() bool {
	return g.modifiable
}

// Frozen returns true if this game has been frozen (its mainLoop goroutine
// has exited and it has been removed from the warm cache). A frozen game
// will be transparently reloaded from storage on the next access via
// ModifiableGame.
func (g *Game) Frozen() bool {
	return g.frozen
}

// markFrozen safely marks the game as frozen and closes its done channel.
// It is safe to call from multiple goroutines concurrently; the done
// channel will only be closed once.
func (g *Game) markFrozen() {
	g.freezeOnce.Do(func() {
		g.frozen = true
		close(g.done)
	})
}

// Moves returns an array of all Moves with their defaults set for this current
// state. This method is useful for getting a list of all moves that could
// possibly be applied to the game at its current state.
// base.GameDelegate.ProposeFixUpMove uses this. Not to be confused with
// Move(), which returns an inflated version of a move that has already been
// succdefully applied to this game in the past.
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

// MoveByName returns a move of the given name set to reasonable defaults for
// the game at its current state. Moves() is similar to this, but returns all
// moves.
func (g *Game) MoveByName(name string) Move {
	return g.MoveByNameForState(name, g.CurrentState())
}

// MoveByNameForState returns a fresh move of the given name with defaults set
// against the supplied immutable state. Unlike MoveByName, it never consults
// CurrentState, which makes it suitable for version-pinned projections and
// other snapshot-pure reads. The state is expected to belong to this game.
func (g *Game) MoveByNameForState(name string, state ImmutableState) Move {
	if !g.initalized {
		return nil
	}
	if state == nil || state.Game() != g {
		return nil
	}

	moveType := g.manager.moveTypeByName(name)

	if moveType == nil {
		return nil
	}

	return moveType.NewMove(state)
}

// Refresh goes and sets this game object to reflect the current state of the
// underlying game in Storage. Basically, when you call manager.Game() you get
// a snapshot of the game in storage at that moment. If you believe that the
// underlying game in storage has been modified, calling Refresh() will re-load
// the snapshot, effectively. You only have to do this if you suspect that a
// modifiable version of this game somewhere in another application binary
// that's currently running may have changed since this game object was
// created. You don't need to call this after calling ProposeMove, even on non-
// modifiable games; it will have been called for you already. If you only have
// one instance of your application binary running at a time, you never need to
// call this.
func (g *Game) Refresh() {

	freshGame := g.manager.Game(g.ID())

	g.cachedCurrentState = nil
	g.cachedHistoricalMoves = nil
	g.version = freshGame.Version()
	g.publishHeadVersion(freshGame.Version())
	g.proposalFrontierVersion.Store(int64(freshGame.ProposalFrontierVersion()))
	g.finished = freshGame.Finished()
	g.winners = freshGame.Winners()

}

// ProposeMove is the way to propose a move to the game. DelayedError will return
// an error in the future if the move was unable to be applied, or nil if the
// move was applied successfully. Proposer is the PlayerIndex of the player who
// is notionally proposing the move. If you don't know which player is moving it,
// AdminPlayerIndex is a reasonable default that will generally allow any move to
// be made. Note that AdminPlayerIndex should be used for engine-initiated actions
// (fix-up moves, timers, debug mode); for simultaneous phases where any player
// can act, the proposer should still be the actual player index (0, 1, 2, ...),
// and your CurrentPlayerIndex() should return AnyPlayerIndex. After the move is
// applied, your GameDelegate's ProposeFixUpMove will be called; if any move is
// returned it will be applied, repeating the cycle until no moves are returned
// from ProposeFixUpMove. DelayedError will only resolve once any applicable
// FixUp moves have been applied already. This is legal to call on a
// non-modifiable game--the change will be dispatched to a modifiable version of
// the game with this ID, and afterwards this Game object's state will be updated
// in place with the new values after the change (by automatically calling
// Refresh()).
func (g *Game) ProposeMove(move Move, proposer PlayerIndex) DelayedError {
	return g.proposeMove(move, proposer, nil)
}

// ProposeMoveAtVersion proposes a move only if expectedVersion is still the
// current version when the serialized game loop is ready to apply it.
func (g *Game) ProposeMoveAtVersion(move Move, proposer PlayerIndex, expectedVersion int) DelayedError {
	return g.proposeMove(move, proposer, &expectedVersion)
}

func (g *Game) proposeMove(move Move, proposer PlayerIndex, expectedVersion *int) DelayedError {

	if !g.Modifiable() {
		return g.manager.proposeMoveOnGame(g, move, proposer, expectedVersion)
	}

	errChan := make(DelayedError, 1)

	workItem := &proposedMoveItem{
		move:            move,
		proposer:        proposer,
		expectedVersion: expectedVersion,
		ch:              errChan,
	}

	if !g.initalized {
		//The channel isn't even ready to send one.
		errChan <- errors.New("[roposed a move before the game had been successfully set-up")
		return errChan
	}

	if g.frozen {
		errChan <- errors.New("game has been frozen")
		return errChan
	}

	g.lastActivity = time.Now()

	select {
	case g.proposedMoves <- workItem:
		// queued successfully
	case <-g.done:
		errChan <- errors.New("game has been frozen")
	}

	return errChan

}

func (g *Game) applyProposedMove(item *proposedMoveItem) error {
	if item.expectedVersion != nil && g.Version() != *item.expectedVersion {
		return &StaleVersionError{Expected: *item.expectedVersion, Actual: g.Version()}
	}
	return g.applyMove(item.move, item.proposer, false, 0, selfInitiatorSentinel)
}

// triggerAgents is called after a PlayerMove (and its chain of fixUp moves) is called.
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

// Game applies the move to the state if it is currently legal. May only be
// called by mainLoop. Propose moves with game.ProposeMove instead.
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

	if err := validateMoveChoiceInputDomain(move); err != nil {
		return errors.NewFriendly(err.Error())
	}

	if err := move.Legal(currentState, proposer); err != nil {
		//It's not legal, reject.
		if isFixUp {
			// #65: fixup rejections were silently discarded before this;
			// log them at debug (design spec §6, "Server": "no exceptions").
			g.logFixupRejection(move, currentState, proposer, err)
		}
		return errors.NewFriendly(err.Error())
	}

	currentPhaseVal := g.manager.delegate.CurrentPhase(currentState)
	var currentPhase enum.EnumKey
	if currentPhaseVal != nil {
		currentPhase = currentPhaseVal.Value()
	}

	newState, err := currentState.copy(false)

	if err != nil {
		return baseErr.WithError("There was an internal error copying the state: " + err.Error())
	}

	newState.version = versionToSet

	// Some complete framework move behaviors contribute a configured state
	// effect in addition to the concrete move's Apply method. The method is
	// promoted through embedding, so a game may add its own Apply without
	// silently disabling the configured behavior.
	if applier, ok := move.(configuredMoveStateApplier); ok {
		if err := applier.ApplyConfiguredMoveState(newState); err != nil {
			return baseErr.WithError("The move's configured state effect returned an error:" + err.Error())
		}
	}

	if err := move.Apply(newState); err != nil {
		return baseErr.WithError("The move's apply function returned an error:" + err.Error())
	}

	if err := newState.validateBeforeSave(); err != nil {
		return baseErr.WithError("The modified state had an invalidity, so the move was not applied. " + err.Error())
	}

	//Check to see if that move made the game finished.

	finished, winners := g.manager.Delegate().CheckGameFinished(newState)

	// Everything above this point is a non-durable rejection path. Keep the
	// settled frontier advertised while a queued proposal is merely being
	// checked, and invalidate it only once the candidate is ready to commit.
	// SaveGameAndCurrentState records the unknown marker atomically with the new
	// head; if that save rejects, restore the unchanged in-memory head exactly.
	previousVersion := g.version
	previousFrontier := g.ProposalFrontierVersion()
	previousFinished := g.finished
	previousWinners := append([]PlayerIndex(nil), g.winners...)
	previousModified := g.modified
	_ = g.invalidateProposalFrontier(false)
	if finished {
		g.finished = true
		g.winners = winners
		//TODO: persist to database here.
	}
	g.version = versionToSet
	g.publishHeadVersion(versionToSet)

	//Note that we want the phase that we were in BEFORE this move was applied.
	moveStorageRecord := StorageRecordForMove(move, currentPhase, proposer)

	//use the precise time we'll set for the move.
	g.modified = move.Info().Timestamp()

	//TODO: test that if we fail to save state to storage everything's fine.
	if err := g.manager.Storage().SaveGameAndCurrentState(g.StorageRecord(), newState.StorageRecord(), moveStorageRecord); err != nil {
		g.version = previousVersion
		g.publishHeadVersion(previousVersion)
		g.proposalFrontierVersion.Store(int64(previousFrontier))
		g.finished = previousFinished
		g.winners = previousWinners
		g.modified = previousModified
		return baseErr.WithError("Storage returned an error:" + err.Error())
	}

	//Ok, the state stuck and is now canonical--trigger the actions it was
	//supposed to do.
	//Expire the currentState cache only after the durable head advances.
	g.cachedCurrentState = nil
	newState.committed()

	if recurseCount > maxRecurseCount {
		return ErrTooManyFixUps
	}

	if g.finished {

		if !isFixUp {
			g.markProposalFrontierAfterCommit()
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

	// Recursive fix-up frames only unwind. The initiating frame owns the single
	// durable boundary write, agent pass, and notification after the entire chain
	// has reached a terminal fix-up check.
	if !isFixUp {
		g.markProposalFrontierAfterCommit()

		// Agent scheduling is downstream of the durable proposal boundary. Failure
		// here must not erase evidence that the initiating move and fix-ups settled.
		if err := g.triggerAgents(); err != nil {
			return baseErr.WithError("Failed to trigger agent: " + err.Error())
		}

		g.manager.Storage().PlayerMoveApplied(g.StorageRecord())
	}

	return nil

}

// markProposalFrontierAfterCommit records proposal-boundary metadata without
// changing the result of an already durable move. A marker failure leaves the
// frontier unknown so an explicit recovery pass can safely certify it later.
func (g *Game) markProposalFrontierAfterCommit() {
	if err := g.markProposalFrontier(); err != nil {
		g.manager.Logger().WithError(err).WithFields(map[string]interface{}{
			"gameID":  g.ID(),
			"version": g.Version(),
		}).Error("Could not persist proposal frontier after committed move")
	}
}
