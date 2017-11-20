package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
	"reflect"
	"strconv"
	"sync"
)

//go:generate autoreader

//game.Name() to set of move types that are always legal
var alwaysLegalMoveTypesByGame map[string]map[string]bool
var alwaysLegalMoveTypesMutex sync.RWMutex

func init() {
	alwaysLegalMoveTypesMutex.Lock()
	alwaysLegalMoveTypesByGame = make(map[string]map[string]bool)
	alwaysLegalMoveTypesMutex.Unlock()
}

//The interface that moves that can be handled by DefaultConfig implement.
type defaultConfigMoveType interface {
	//The name for the move type
	MoveTypeName(manager *boardgame.GameManager) string
	//The name for the HelpText
	MoveTypeHelpText(manager *boardgame.GameManager) string
	//Whether the move should be a fix up.
	MoveTypeIsFixUp(manager *boardgame.GameManager) bool
}

//MustDefaultConfig is a wrapper around DefaultConfig that if it errors will
//panic. Only suitable for being used during setup.
func MustDefaultConfig(manager *boardgame.GameManager, exampleStruct boardgame.Move) *boardgame.MoveTypeConfig {
	result, err := DefaultConfig(manager, exampleStruct)

	if err != nil {
		panic("Couldn't DefaultConfig: " + err.Error())
	}

	return result
}

//DefaultConfig is a powerful default MoveTypeConfig generator. In many cases
//you'll implement moves that are very thin embeddings of moves in this
//package. Generating a MoveTypeConfig for each is a pain. This method auto-
//generates the MoveTypeConfig based on an example nil type of your move to
//install. It consults move.MoveTypeName and move.MoveTypeHelpText to generate
//the name and helptext. Moves in this package return reasonable values for
//those methods, based on the configuration you set on the rest of your move.
//See the package doc for an example of use.
func DefaultConfig(manager *boardgame.GameManager, exampleStruct boardgame.Move) (*boardgame.MoveTypeConfig, error) {

	if exampleStruct == nil {
		return nil, errors.New("nil struct provided")
	}

	defaultConfig, ok := exampleStruct.(defaultConfigMoveType)

	if !ok {
		return nil, errors.New("Example struct didn't have MoveTypeName and MoveTypeHelpText.")
	}

	name := defaultConfig.MoveTypeName(manager)
	helpText := defaultConfig.MoveTypeHelpText(manager)
	isFixUp := defaultConfig.MoveTypeIsFixUp(manager)

	val := reflect.ValueOf(exampleStruct)

	//We can accept either pointer or struct types.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	return &boardgame.MoveTypeConfig{
		Name:     name,
		HelpText: helpText,
		MoveConstructor: func() boardgame.Move {
			return reflect.New(typ).Interface().(boardgame.Move)
		},
		IsFixUp: isFixUp,
	}, nil
}

/*
Base is an optional, convenience struct designed to be embedded
anonymously in your own Moves. It implements no-op methods for many of the
required methods on Moves. Apply is not covered, because every Move
should implement their own, and if this implemented them it would obscure
errors where for example your Apply() was incorrectly named and thus not used.
In general your MoveConstructor can always be exactly the same, modulo the
name of your underlying move type:

	MoveConstructor: func() boardgame.Move {
 		return new(myMoveStruct)
	}

Base's Legal() method does basic checking for whehter the move is legal in
this phase, so your own Legal() method should always call Base.Legal() at the
top of its own method.

It is extremely rare to not use moves.Base either directly, or implicitly
within another sub-class in your move.

Base cannot help your move implement PropertyReadSetter; use autoreader to
generate that code for you.

*/
type Base struct {
	info           *boardgame.MoveInfo
	topLevelStruct boardgame.Move
}

func (d *Base) SetInfo(m *boardgame.MoveInfo) {
	d.info = m
}

//Type simply returns BaseMove.MoveInfo
func (d *Base) Info() *boardgame.MoveInfo {
	return d.info
}

func (d *Base) SetTopLevelStruct(m boardgame.Move) {
	d.topLevelStruct = m
}

//TopLevelStruct returns the object that was set via SetTopLevelStruct.
func (d *Base) TopLevelStruct() boardgame.Move {
	return d.topLevelStruct
}

//DefaultsForState doesn't do anything
func (d *Base) DefaultsForState(state boardgame.State) {
	return
}

//Description defaults to returning the Type's HelpText()
func (d *Base) Description() string {
	return d.Info().Type().HelpText()
}

//ValidConfiguration always returns nil because there is no required
//configuration for moves.Base.
func (d *Base) ValidConfiguration(exampleState boardgame.MutableState) error {
	return nil
}

//MoveTypeName is used by DefaultConfig to generate the name. Subclasses
//normally override this, but you often don't have to.
func (b *Base) MoveTypeName(manager *boardgame.GameManager) string {
	return "Base Move"
}

//MoveTypeHelpText is used by DefaultConfig to generate the HelpText.
//Subclasses normally overridd this, but you often don't have to.
func (b *Base) MoveTypeHelpText(manager *boardgame.GameManager) string {
	return "A base move that does nothing on its own"
}

//MoveTypeIsFixUp is used by Defaultconfig to generate the IsFixUp value. Base
//is false, but other moves in this package will return true.
func (b *Base) MoveTypeIsFixUp(manager *boardgame.GameManager) bool {
	//TODO: once we have reasonable overridings in all other moves that
	//derive, default this to false.
	return true
}

//Legal checks whether the game's CurrentPhase (as determined by the delegate)
//is one of the LegalPhases for this moveType. A zero-length LegalPhases is
//interpreted as the move being legal in all phases. The string for the
//current phase will be based on the enum value of the PhaseEnum named by
//delegate.PhaseEnumName(), if it exists. Next, it checks to see if the give
//move is at a legal point in the move progression for this phase, if it
//exists. Each move in the move progression must show up 1 or more times. The
//method checks to see if we were to make this move, would the moves since the
//last phase change match the pattern? If your move can be made legally
//multiple times in a row in a given move progression, implement
//moveinterfaces.AllowMultipleInProgression() and return true.
func (d *Base) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := d.legalInPhase(state); err != nil {
		return err
	}

	return d.legalMoveInProgression(state, proposer)

}

//legalInPhase will return a descriptive error if this move is not legal in
//the current phase of hte game.
func (d *Base) legalInPhase(state boardgame.State) error {

	legalPhases := d.Info().Type().LegalPhases()

	if len(legalPhases) == 0 {
		return nil
	}

	currentPhase := state.Game().Manager().Delegate().CurrentPhase(state)

	for _, phase := range legalPhases {
		if phase == currentPhase {
			return nil
		}
	}

	phaseName := strconv.Itoa(currentPhase)

	if phaseEnum := state.Game().Manager().Delegate().PhaseEnum(); phaseEnum != nil {
		phaseName = phaseEnum.String(currentPhase)
	}

	return errors.New("Move is not legal in phase " + phaseName)
}

func (d *Base) historicalMovesSincePhaseTransition(game *boardgame.Game, upToVersion int, targetPhase int) []*boardgame.MoveStorageRecord {

	moves := game.MoveRecords(upToVersion)

	//TODO: ideally we'd memoize this so all base moves for this game for this
	//version could use the result. If we do that, we'll want to make sure the
	//lifetime of the cache does not extend beyond the lifetime of the game,
	//or is purged every so often.

	if len(moves) == 0 {
		return nil
	}

	alwaysLegalMoveTypesMutex.RLock()
	alwaysLegalMoveTypes, ok := alwaysLegalMoveTypesByGame[game.Name()]
	alwaysLegalMoveTypesMutex.RUnlock()

	if !ok {

		alwaysLegalMoveTypes = make(map[string]bool)

		//Create the list!
		for _, fixUpMove := range game.Manager().FixUpMoveTypes() {
			if len(fixUpMove.LegalPhases()) == 0 {
				alwaysLegalMoveTypes[fixUpMove.Name()] = true
			}
		}

		alwaysLegalMoveTypesMutex.Lock()
		alwaysLegalMoveTypesByGame[game.Name()] = alwaysLegalMoveTypes
		alwaysLegalMoveTypesMutex.Unlock()
	}

	var keptMoves []*boardgame.MoveStorageRecord

	for i := len(moves) - 1; i >= 0; i-- {
		move := moves[i]

		if alwaysLegalMoveTypes[move.Name] {
			//We skip move types that are always legal for the purposes of
			//matching.
			continue
		}

		if move.Phase != targetPhase {
			//Must have fallen off the end of the current phase's most recent run
			break
		}

		keptMoves = append(keptMoves, move)
	}

	//keptMoves is backwards, reverse it.

	moves = nil

	for i := len(keptMoves) - 1; i >= 0; i-- {
		moves = append(moves, keptMoves[i])
	}

	return moves

}

func (d *Base) legalMoveInProgression(state boardgame.State, proposer boardgame.PlayerIndex) error {
	currentPhase := state.Game().Manager().Delegate().CurrentPhase(state)

	pattern := state.Game().Manager().Delegate().PhaseMoveProgression(currentPhase)

	//If there is no legal move progression then moves are legal in the phase at any time
	if pattern == nil {
		return nil
	}

	historicalMoves := d.historicalMovesSincePhaseTransition(state.Game(), state.Version(), currentPhase)

	progression := make([]string, len(historicalMoves))

	for i, move := range historicalMoves {
		progression[i] = move.Name
	}

	//If we were to add our target move to the historical progression, would it match the pattern?
	if !progressionMatches(append(progression, d.Info().Type().Name()), pattern) {
		return errors.New("This move is not legal at this point in the current phase.")
	}

	//Are we a new type of move in the progression? if so, is the move before
	//us still legal?

	if len(historicalMoves) == 0 {
		//We're the first move, it's fine.
		return nil
	}

	lastMoveRecord := historicalMoves[len(historicalMoves)-1]

	if lastMoveRecord.Name == d.Info().Type().Name() {

		//We're applying multiple in a row. Is that legal?

		//We can't check ourselves because we're embedded in the real move type.
		testMove := d.TopLevelStruct()

		allowMultiple, ok := testMove.(moveinterfaces.AllowMultipleInProgression)

		if !ok || !allowMultiple.AllowMultipleInProgression() {
			return errors.New("This move was just applied and is not configured to allow multiple in a row in this phase.")
		}

		return nil
	}

	lastMoveType := state.Game().Manager().FixUpMoveTypeByName(lastMoveRecord.Name)

	if lastMoveType == nil {
		lastMoveType = state.Game().Manager().PlayerMoveTypeByName(lastMoveRecord.Name)
	}

	if lastMoveType == nil {
		return errors.New("Unexpected error: couldn't find a historical move type")
	}

	//LastMove will have all of the defaults set.
	lastMove := lastMoveType.NewMove(state)

	if lastMove.Legal(state, proposer) == nil {
		return errors.New("A move that needs to happen earlier in the phase is still legal to apply.")
	}

	return nil

}

//progressionMatches returns true if the given history matches the pattern.
func progressionMatches(input []string, pattern []string) bool {

	inputPosition := 0
	patternPosition := 0

	for inputPosition < len(input) {

		inputItem := input[inputPosition]
		patternItem := pattern[patternPosition]

		if inputItem != patternItem {
			//Perhaps we just passed to the next part of the pattern?

			//that's not legal at the very front of input
			if inputPosition == 0 {
				return false
			}

			patternPosition++

			if patternPosition >= len(pattern) {
				//No more pattern, I guess we didn't match.
				return false
			}

			patternItem = pattern[patternPosition]

			if inputItem != patternItem {
				//Nope, we didn't match the next part of the pattern, we just don't match
				return false
			}

		}

		inputPosition++

	}

	//If we got to the end of the input without invalidating then it passes.
	return true

}
