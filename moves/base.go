package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/moves/internal/privateconstants"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

//go:generate boardgame-util codegen

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
	FallbackName(m *boardgame.GameManager) string
	//TODO: shouldn't HelpText also take a manager? But move.HelpText() is
	//called live, unlike Name, which is fully implied at MoveConfig install
	//time.
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

//ValidConfiguration ensures that phase progression is configured in sane way.
func (d *Base) ValidConfiguration(exampleState boardgame.State) error {
	config := d.CustomConfiguration()

	if config[privateconstants.LegalMoveProgression] != nil {

		legalPhasesRaw := config[privateconstants.LegalPhases]

		if legalPhasesRaw == nil {
			return errors.New("WithLegalMoveProgression configuration provided, but without WithLegalPhases")
		}

		legalPhases, ok := legalPhasesRaw.([]int)

		if !ok {
			return errors.New("Legal Phases unexpectedly were not ints")
		}

		delegate := exampleState.Game().Manager().Delegate()

		phaseEnum := delegate.PhaseEnum()

		if phaseEnum == nil {
			return nil
		}

		treeEnum := phaseEnum.TreeEnum()

		if treeEnum == nil {
			return nil
		}

		for _, phase := range legalPhases {
			if !treeEnum.IsLeaf(phase) {
				return errors.New("PhaseEnum() returns a TreeEnum, and MoveProgression is Nil, but the LegalPhase provided")
			}
		}

	}

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
//Subclasses generally should not override this. If with.MoveNameSuffix() was
//used, it will then add " - " + suffix to the end of the move name.
func (b *Base) DeriveName(m *boardgame.GameManager) string {

	config := b.CustomConfiguration()

	suffix := ""

	if config != nil {
		rawSuffix, hasSuffix := config[privateconstants.MoveNameSuffix]
		if hasSuffix {
			strSuffix, ok := rawSuffix.(string)
			if !ok {
				return "Unexpected Error: suffix was not a string"
			}
			if strSuffix != "" {
				suffix = " - " + strSuffix
			}
		}
	}

	return b.baseDeriveName(m) + suffix
}

//baseDeriveName does most of the name logic, but not the suffix behavior.
func (b *Base) baseDeriveName(m *boardgame.GameManager) string {

	config := b.CustomConfiguration()

	if config != nil {

		overrideName, hasOverrideName := config[privateconstants.MoveName]

		if hasOverrideName {
			strOverrideName, ok := overrideName.(string)
			if !ok {
				return "Unexpected Error: overrideName was not a string"
			}
			return strOverrideName
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
		return defaultConfig.FallbackName(m)
	}

	//Nothing worked. :-/
	return ""
}

//FallbackName is the name that is returned if other higher-priority
//methods in MoveTypeName fail. For moves.Base returns "Base Move".
func (b *Base) FallbackName(m *boardgame.GameManager) string {
	return "Base Move"
}

//HelpText will return the value passed via the WithHelpText config option, if
//it was passed. Otherwise it will fall back on the move's HelpTextFallback
//method.
func (b *Base) HelpText() string {
	config := b.CustomConfiguration()

	overrideHelpText, hasOverrideHelpText := config[privateconstants.HelpText]

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
	config := b.CustomConfiguration()
	return overrideIsFixUp(config, false)
}

//overrideIsFixUp takes the config and the base fix up value and returns the override if it exists, otherwise defaultIsFixUp
func overrideIsFixUp(config boardgame.PropertyCollection, defaultIsFixUp bool) bool {
	overrideIsFixUp, hasOverrideIsFixUp := config[privateconstants.IsFixUp]

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
	val := d.CustomConfiguration()[privateconstants.LegalPhases]
	ints, ok := val.([]int)
	if !ok {
		return nil
	}
	return ints
}

func (d *Base) legalMoveProgression() interfaces.MoveProgressionGroup {
	val := d.CustomConfiguration()[privateconstants.LegalMoveProgression]
	group, ok := val.(interfaces.MoveProgressionGroup)
	if !ok {
		return nil
	}
	return group
}

//legalInPhase will return a descriptive error if this move is not legal in
//the current phase of hte game.
func (d *Base) legalInPhase(state boardgame.ImmutableState) error {

	legalPhases := d.legalPhases()

	if len(legalPhases) == 0 {
		//If PhaseEnum is a TreeEnum, this is basically equivalent to the
		//legalPhases being []int{0}.
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

	if phaseEnum != nil {
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

//legalMoveInProgression is NOT an exhaustive check. It simply confirms that
//this specific point would be legitimate to apply. Note that sometimes this
//will return nil even when there really should be another move in front of
//this that could still apply; that other move should actually be applied due
//to ordering of moves in ProposeFixUpMove. Finally, note that technically for
//AllowMultipleInProgression moves, this relies on the sub-classes Legal()
//method terminating, becuase this method won't; because as far as the
//progression is concerned, it's legal, and it's the sub-class's Legal()
//method's job to decide it's no longer legal.
func (d *Base) legalMoveInProgression(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	group := d.legalMoveProgression()

	//If there is no legal move progression then moves are legal in the phase at any time
	if group == nil {
		return nil
	}

	currentPhase := state.Game().Manager().Delegate().CurrentPhase(state)

	historicalMoves := d.historicalMovesSincePhaseTransition(state.Game(), state.Version(), currentPhase)

	//Add ourselves ot the end of the tape, since we're proposing adding ourselves.
	historicalMoves = append(historicalMoves, &boardgame.MoveStorageRecord{
		Name: d.Name(),
	})

	return matchTape(group, movesToNames(historicalMoves))

}

func makeTape(moveNames []string) *interfaces.MoveGroupHistoryItem {
	var tapeStart *interfaces.MoveGroupHistoryItem
	var tapeEnd *interfaces.MoveGroupHistoryItem

	for _, moveName := range moveNames {
		newItem := &interfaces.MoveGroupHistoryItem{
			MoveName: moveName,
		}
		if tapeStart == nil {
			tapeStart = newItem
		}

		if tapeEnd != nil {
			tapeEnd.Rest = newItem
		}

		tapeEnd = newItem
	}

	return tapeStart
}

func movesToNames(moves []*boardgame.MoveStorageRecord) []string {
	result := make([]string, len(moves))

	for i, move := range moves {
		result[i] = move.Name
	}

	return result
}

func matchTape(group interfaces.MoveProgressionGroup, historicalMoves []string) error {

	tapeStart := makeTape(historicalMoves)

	err, rest := group.Satisfied(tapeStart)

	defaultErr := errors.NewFriendly("The move was not legal at this phase in the progression")

	if err != nil {
		return defaultErr.WithError(err.Error())
	}

	if rest != nil {
		return defaultErr.WithError("The progression only matched some of the proposed move history")
	}

	return nil
}

//stackName returns the name of the stack for helpTExt, name, etc based on the
//configPropName.
func stackName(move moveInfoer, configPropName string, exampleStack boardgame.ImmutableStack, exampleState boardgame.ImmutableState) string {
	config := move.CustomConfiguration()

	val, ok := config[configPropName]

	if ok {
		strVal, ok := val.(string)
		if ok {
			return strVal
		}
	}

	if derivedName := findStackName(exampleStack, exampleState); derivedName != "" {
		return derivedName
	}

	return "a stack"
}

func findStackName(exampleStack boardgame.ImmutableStack, exampleState boardgame.ImmutableState) string {

	if exampleStack == nil || exampleState == nil {
		return ""
	}

	if result := findStackNameInReader(exampleState.ImmutableGameState().Reader(), exampleStack); result != "" {
		return result
	}

	if result := findStackNameInReader(exampleState.ImmutablePlayerStates()[0].Reader(), exampleStack); result != "" {
		return result
	}

	return ""
}

func findStackNameInReader(reader boardgame.PropertyReader, exampleStack boardgame.ImmutableStack) string {
	for propName, propType := range reader.Props() {
		if propType != boardgame.TypeStack {
			continue
		}
		stack, err := reader.ImmutableStackProp(propName)

		if err != nil {
			log.Println("Unexpected error: " + err.Error())
			return ""
		}

		if stack == exampleStack {
			return titleCaseToWords(propName)
		}
	}
	return ""
}
