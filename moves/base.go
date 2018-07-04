package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"reflect"
	"strconv"
	"strings"
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
type autoConfigFallbackMoveType interface {
	//The last resort move-name generator that MoveName will fall back on if
	//none of the other options worked.
	FallbackName() string
	FallbackHelpText() string
}

//A func that will fail to compile if all of the moves don't have a valid fallback.
func ensureAllMovesSatisfyFallBack() {
	var m autoConfigFallbackMoveType
	m = new(ApplyUntil)
	m = new(ApplyUntilCount)
	m = new(ApplyCountTimes)
	m = new(Base)
	m = new(CollectCountComponents)
	m = new(CollectComponentsUntilGameCountReached)
	m = new(CollectComponentsUntilPlayerCountLeft)
	m = new(CurrentPlayer)
	m = new(DealCountComponents)
	m = new(DealComponentsUntilGameCountLeft)
	m = new(DealComponentsUntilPlayerCountReached)
	m = new(FinishTurn)
	m = new(MoveCountComponents)
	m = new(MoveComponentsUntilCountLeft)
	m = new(MoveComponentsUntilCountReached)
	m = new(RoundRobin)
	m = new(RoundRobinNumRounds)
	m = new(ShuffleStack)
	m = new(StartPhase)
	m = new(DefaultComponent)
	if m != nil {
		return
	}
}

/*
Base is an optional, convenience struct designed to be embedded
anonymously in your own Moves. It implements no-op methods for many of the
required methods on Moves. Apply is not covered, because every Move
should implement their own, and if this implemented them it would obscure
errors where for example your Apply() was incorrectly named and thus not used.

Base's Legal() method does basic checking for whehter the move is legal in
this phase, so your own Legal() method should always call Base.Legal() (or the
Legal method of whichever struct you embedded that in turn calls Base.Legal())
at the top of its own method.

Base contains a fair bit of logic for generating the values that auto.Config
will use for the move configuration; see MoveType* methods on Base for more
information.

It is extremely rare to not use moves.Base either directly, or implicitly
within another sub-class in your move.

Base cannot help your move implement PropertyReadSetter; use autoreader to
generate that code for you.

+autoreader
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
func (d *Base) DefaultsForState(state boardgame.ImmutableState) {
	return
}

//Description defaults to returning the Type's HelpText()
func (d *Base) Description() string {
	return d.TopLevelStruct().HelpText()
}

//ValidConfiguration always returns nil because there is no required
//configuration for moves.Base.
func (d *Base) ValidConfiguration(exampleState boardgame.State) error {
	return nil
}

var titleCaseReplacer *strings.Replacer

//titleCaseToWords writes "ATitleCaseString" to "A Title Case String"
func titleCaseToWords(in string) string {

	//substantially recreated in autoreader/enums.go

	if titleCaseReplacer == nil {

		var replacements []string

		for r := 'A'; r <= 'Z'; r++ {
			str := string(r)
			replacements = append(replacements, str)
			replacements = append(replacements, " "+str)
		}

		titleCaseReplacer = strings.NewReplacer(replacements...)

	}

	return strings.TrimSpace(titleCaseReplacer.Replace(in))

}

//DeriveName is used by auto.Config to generate the name for the move. This
//implementation is where the majority of MoveName magic logic comes from.
//First, it will use the configuration passed to auto.Config via WithMoveName,
//if provided. Next, it checks the name of the topLevelStruct via reflection.
//If the struct does not come from the moves package, it will create a name
//like `MoveMyMove` --> `My Move`. Finally, if it's a struct from this
//package, it will fall back on whatever the FallbackName() method returns.
//Subclasses generally should not override this.
func (b *Base) DeriveName() string {

	moveInfo := b.Info()

	if moveInfo != nil {

		moveType := b.Info().Type()

		if moveType != nil {
			config := moveType.CustomConfiguration()

			overrideName, hasOverrideName := config[configNameMoveName]

			if hasOverrideName {
				strOverrideName, ok := overrideName.(string)
				if !ok {
					return "Unexpected Error: overrideName was not a string"
				}
				return strOverrideName
			}
		}
	}

	move := b.TopLevelStruct()

	val := reflect.ValueOf(move)

	//We can accept either pointer or struct types.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	if !strings.HasSuffix(typ.PkgPath(), "boardgame/moves") {
		//For any move struct where the top level isn't in this package, just
		//title case its name and be done with it!
		name := typ.Name()
		name = strings.TrimPrefix(name, "Move")
		name = strings.TrimPrefix(name, "move")
		name = titleCaseToWords(name)
		return name
	}

	defaultConfig, ok := move.(autoConfigFallbackMoveType)

	if ok {
		return defaultConfig.FallbackName()
	}

	//Nothing worked. :-/
	return ""

}

//FallbackName is the name that is returned if other higher-priority
//methods in MoveTypeName fail. For moves.Base returns "Base Move".
func (b *Base) FallbackName() string {
	return "Base Move"
}

//HelpText will return the value passed via the WithHelpText config option, if
//it was passed. Otherwise it will fall back on the move's HelpTextFallback
//method.
func (b *Base) HelpText() string {
	config := b.Info().Type().CustomConfiguration()

	overrideHelpText, hasOverrideHelpText := config[configNameHelpText]

	if hasOverrideHelpText {
		strOverrideHelpText, ok := overrideHelpText.(string)
		if !ok {
			return "Unexpected Error: overrideHelpText was not a string"
		}
		return strOverrideHelpText
	}

	move := b.TopLevelStruct()

	defaultConfig, ok := move.(autoConfigFallbackMoveType)

	if ok {
		return defaultConfig.FallbackHelpText()
	}

	//Nothing worked. :-/
	return ""

}

//FallbackHelpText is the help text that will be used by HelpText if nothing
//was passed via WithHelpText to auto.Config. By default it returns "A base
//move that does nothing on its own"
func (b *Base) FallbackHelpText() string {
	return "A base move that does nothing on its own"
}

//IsFixUp will return the value passed with WithFixUp, falling back on
//returning false.
func (b *Base) IsFixUp() bool {
	config := b.Info().Type().CustomConfiguration()
	return overrideIsFixUp(config, false)
}

//overrideIsFixUp takes the config and the base fix up value and returns the override if it exists, otherwise defaultIsFixUp
func overrideIsFixUp(config boardgame.PropertyCollection, defaultIsFixUp bool) bool {
	overrideIsFixUp, hasOverrideIsFixUp := config[configNameIsFixUp]

	if hasOverrideIsFixUp {
		boolOverrideIsFixUp, ok := overrideIsFixUp.(bool)
		if !ok {
			return false
		}
		return boolOverrideIsFixUp
	}

	return defaultIsFixUp
}

//Legal checks whether the game's CurrentPhase (as determined by the delegate)
//is one of the LegalPhases for this moveType. If the delegate's PhaseEnum is
//a TreeEnum, it will also pass this test if delegate.CurrentPhase() value's
//ancestors match the legal move type. A zero-length LegalPhases is
//interpreted as the move being legal in all phases. The string for the
//current phase will be based on the enum value of the PhaseEnum named by
//delegate.PhaseEnumName(), if it exists. Next, it checks to see if the give
//move is at a legal point in the move progression for this phase, if it
//exists. Each move in the move progression must show up 1 or more times. The
//method checks to see if we were to make this move, would the moves since the
//last phase change match the pattern? If your move can be made legally
//multiple times in a row in a given move progression, implement
//interfaces.AllowMultipleInProgression() and return true.
func (d *Base) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := d.legalInPhase(state); err != nil {
		return err
	}

	return d.legalMoveInProgression(state, proposer)

}

//Name returns the name of this move according to MoveInfo.Name(). A simple
//convenience wrapper that allows you to avoid a nil check.
func (d *Base) Name() string {
	if d.info == nil {
		return ""
	}
	return d.info.Name()
}

//CustomConfiguration returns the custom configuration associated with this
//move, according to MoveInfo.CustomConfiguration(). A simple convenience
//wrapper that allows you to avoid a nil check.
func (d *Base) CustomConfiguration() boardgame.PropertyCollection {

	if d.info == nil {
		return nil
	}

	return d.info.CustomConfiguration()

}

func (d *Base) legalPhases() []int {
	val := d.Info().Type().CustomConfiguration()[configNameLegalPhases]
	ints, ok := val.([]int)
	if !ok {
		return nil
	}
	return ints
}

func (d *Base) legalMoveProgression() []string {
	val := d.Info().Type().CustomConfiguration()[configNameLegalMoveProgression]
	strs, ok := val.([]string)
	if !ok {
		return nil
	}
	return strs
}

//legalInPhase will return a descriptive error if this move is not legal in
//the current phase of hte game.
func (d *Base) legalInPhase(state boardgame.ImmutableState) error {

	legalPhases := d.legalPhases()

	if len(legalPhases) == 0 {
		return nil
	}

	delegate := state.Game().Manager().Delegate()

	phaseEnum := delegate.PhaseEnum()

	var treeEnum enum.TreeEnum
	if phaseEnum != nil {
		treeEnum = phaseEnum.TreeEnum()
	}

	currentPhase := delegate.CurrentPhase(state)

	//totalCurrentPhases is all of the current phases we could be considered
	//to be in. Deafaults to an []int with just the current phase.
	totalCurrentPhases := []int{currentPhase}

	if treeEnum != nil {
		//If PhaseEnum is a tree, then the phase we're in for this purpose is
		//all ancestor phases.
		totalCurrentPhases = treeEnum.Ancestors(currentPhase)
	}

	for _, phase := range legalPhases {
		for _, candidateCurrentPhase := range totalCurrentPhases {
			if phase == candidateCurrentPhase {
				return nil
			}
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
		for _, move := range game.Moves() {

			if !move.IsFixUp() {
				continue
			}

			if len(d.legalPhases()) == 0 {
				alwaysLegalMoveTypes[move.Info().Name()] = true
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

func (d *Base) legalMoveInProgression(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	currentPhase := state.Game().Manager().Delegate().CurrentPhase(state)

	pattern := d.legalMoveProgression()

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

		allowMultiple, ok := testMove.(interfaces.AllowMultipleInProgression)

		if !ok || !allowMultiple.AllowMultipleInProgression() {
			return errors.New("This move was just applied and is not configured to allow multiple in a row in this phase.")
		}

		return nil
	}

	lastMoveType := state.Game().Manager().MoveTypeByName(lastMoveRecord.Name)

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

//stackName returns the name of the stack for helpTExt, name, etc based on the
//configPropName.
func stackName(move moveInfoer, configPropName string) string {
	config := move.Info().Type().CustomConfiguration()

	val, ok := config[configPropName]

	if ok {
		strVal, ok := val.(string)
		if ok {
			return strVal
		}
	}

	return "a stack"
}
