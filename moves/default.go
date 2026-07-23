package moves

import (
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"github.com/jkomoros/boardgame/legal"
)

//go:generate boardgame-util codegen

// game.Name() to set of move types that have no move progression logic
var noProgressionMoveTypesByGame map[string]map[string]bool
var noProgressionMoveTypesMutex sync.RWMutex

func init() {
	noProgressionMoveTypesMutex.Lock()
	noProgressionMoveTypesByGame = make(map[string]map[string]bool)
	noProgressionMoveTypesMutex.Unlock()
}

// The interface that moves that can be handled by DefaultConfig implement.
type autoConfigFallbackMoveType interface {
	//The last resort move-name generator that MoveName will fall back on if
	//none of the other options worked.
	FallbackName(m *boardgame.GameManager) string
	//TODO: shouldn't HelpText also take a manager? But move.HelpText() is
	//called live, unlike Name, which is fully implied at MoveConfig install
	//time.
	FallbackHelpText() string
}

// A func that will fail to compile if all of the moves don't have a valid fallback.
func ensureAllMovesSatisfyFallBack() {
	var m autoConfigFallbackMoveType
	m = new(ApplyUntil)
	m = new(ApplyUntilCount)
	m = new(ApplyCountTimes)
	m = new(Default)
	m = new(CollectCountComponents)
	m = new(CollectComponentsUntilGameCountReached)
	m = new(CollectComponentsUntilPlayerCountLeft)
	m = new(CollectAllComponents)
	m = new(CurrentPlayer)
	m = new(DealCountComponents)
	m = new(DealComponentsUntilGameCountLeft)
	m = new(DealComponentsUntilPlayerCountReached)
	m = new(DealAllComponents)
	m = new(FinishTurn)
	m = new(MoveCountComponents)
	m = new(MoveComponentsUntilCountLeft)
	m = new(MoveComponentsUntilCountReached)
	m = new(MoveAllComponents)
	m = new(RoundRobin)
	m = new(RoundRobinNumRounds)
	m = new(ShuffleStack)
	m = new(StartPhase)
	m = new(DefaultComponent)
	m = new(Increment)
	m = new(SeatPlayer)
	m = new(ActivateInactivePlayer)
	m = new(CloseEmptySeat)
	m = new(CloseAllSeats)
	m = new(InactivateEmptySeat)
	m = new(WaitForEnoughPlayers)
	m = new(MoveOnGraph)
	m = new(HopAlongPath)
	m = new(AdvanceToken)
	m = new(AllPlayersSubmitted)
	m = new(ResetAllPlayerSubmissions)
	m = new(AnyPlayer)
	m = new(AdminPlayer)
	m = new(SelectTeam)
	m = new(SelectRole)
	m = new(SelectColor)
	if m != nil {
		return
	}
}

/*
Default is an optional, convenience struct designed to be embedded
anonymously in your own Moves. It builds on [base.Move] to add to a layer of
default Legal logic and configuraability. Apply is not covered, because every
Move should implement their own, and if this implemented them it would obscure
errors where for example your Apply() was incorrectly named and thus not used.

Default's Legal() method does basic checking for whehter the move is legal in
this phase, so your own Legal() method should always call Default.Legal() (or the
Legal method of whichever struct you embedded that in turn calls Default.Legal())
at the top of its own method.

Default contains a fair bit of logic for generating the values that [AutoConfigurer.Config]
will use for the move configuration; see MoveType* methods on Default for more
information.

It is extremely rare to not use Default either directly, or implicitly
within another sub-class in your move.

boardgame:codegen
*/
type Default struct {
	base.Move
}

// ValidConfiguration ensures that phase progression is configured in sane way.
func (d *Default) ValidConfiguration(exampleState boardgame.State) error {
	config := d.CustomConfiguration()

	if config[configPropLegalMoveProgression] != nil {

		legalPhasesRaw := config[configPropLegalPhases]

		if legalPhasesRaw == nil {
			return errors.New("WithLegalMoveProgression configuration provided, but without WithLegalPhases")
		}

		legalPhases, ok := legalPhasesRaw.([]enum.EnumKey)

		if !ok {
			return errors.New("Legal Phases unexpectedly were not EnumKeys")
		}

		delegate := exampleState.Manager().Delegate()

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

// titleCaseToWords writes "ATitleCaseString" to "A Title Case String"
func titleCaseToWords(in string) string {

	//substantially recreated in boardgame-util/lib/codegen/enums.go

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

// DeriveName is used by auto.Config to generate the name for the move. This
// implementation is where the majority of MoveName magic logic comes from.
// First, it will use the configuration passed to auto.Config via WithMoveName,
// if provided. Next, it checks the name of the concreteMove via reflection.
// If the struct does not come from the moves package, it will create a name
// like `MoveMyMove` --> `My Move`. Finally, if it's a struct from this
// package, it will fall back on whatever the FallbackName() method returns.
// Subclasses generally should not override this. If WithMoveNameSuffix() was
// used, it will then add " - " + suffix to the end of the move name.
func (d *Default) DeriveName(m *boardgame.GameManager) string {

	config := d.CustomConfiguration()

	suffix := ""

	if config != nil {
		rawSuffix, hasSuffix := config[configPropMoveNameSuffix]
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

	return d.baseDeriveName(m) + suffix
}

// baseDeriveName does most of the name logic, but not the suffix behavior.
func (d *Default) baseDeriveName(m *boardgame.GameManager) string {

	config := d.CustomConfiguration()

	if config != nil {

		overrideName, hasOverrideName := config[configPropMoveName]

		if hasOverrideName {
			strOverrideName, ok := overrideName.(string)
			if !ok {
				return "Unexpected Error: overrideName was not a string"
			}
			return strOverrideName
		}
	}

	move := d.Info().ConcreteMove()

	val := reflect.ValueOf(move)

	//We can accept either pointer or struct types.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	if !strings.HasSuffix(typ.PkgPath(), "boardgame/moves") {
		//For any concrete move whose type isn't in this package, just
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

// FallbackName is the name that is returned if other higher-priority
// methods in MoveTypeName fail. For moves.Default returns "Base Move".
func (d *Default) FallbackName(m *boardgame.GameManager) string {
	return "Default Move"
}

// HelpText will return the value passed via the WithHelpText config option, if
// it was passed. Otherwise it will fall back on the move's HelpTextFallback
// method.
func (d *Default) HelpText() string {
	config := d.CustomConfiguration()

	overrideHelpText, hasOverrideHelpText := config[configPropHelpText]

	if hasOverrideHelpText {
		strOverrideHelpText, ok := overrideHelpText.(string)
		if !ok {
			return "Unexpected Error: overrideHelpText was not a string"
		}
		return strOverrideHelpText
	}

	move := d.Info().ConcreteMove()

	defaultConfig, ok := move.(autoConfigFallbackMoveType)

	if ok {
		return defaultConfig.FallbackHelpText()
	}

	//Nothing worked. :-/
	return ""

}

// FallbackHelpText is the help text that will be used by HelpText if nothing
// was passed via WithHelpText to auto.Config. By default it returns "A default
// move that does nothing on its own"
func (d *Default) FallbackHelpText() string {
	return "A default move that does nothing on its own"
}

// IsFixUp will return the value passed with WithFixUp, falling back on
// returning false.
func (d *Default) IsFixUp() bool {
	config := d.CustomConfiguration()
	return overrideIsFixUp(config, false)
}

// overrideIsFixUp takes the config and the base fix up value and returns the override if it exists, otherwise defaultIsFixUp
func overrideIsFixUp(config boardgame.PropertyCollection, defaultIsFixUp bool) bool {
	overrideIsFixUp, hasOverrideIsFixUp := config[configPropIsFixUp]

	if hasOverrideIsFixUp {
		boolOverrideIsFixUp, ok := overrideIsFixUp.(bool)
		if !ok {
			return false
		}
		return boolOverrideIsFixUp
	}

	return defaultIsFixUp
}

// Legal checks whether the game's CurrentPhase (as determined by the delegate)
// is one of the LegalPhases for this moveType. If the delegate's PhaseEnum is
// a TreeEnum, it will also pass this test if delegate.CurrentPhase() value's
// ancestors match the legal move type. A zero-length LegalPhases is
// interpreted as the move being legal in all phases. The string for the
// current phase will be based on the enum value of the PhaseEnum named by
// delegate.PhaseEnumName(), if it exists. Next, it checks to see if the give
// move is at a legal point in the move progression for this phase, if it
// exists. Each move in the move progression must show up 1 or more times.
// (Moves that don't define a progression group are ignored, since they may
// show up at any time in the phase.) The method checks to see if we were to
// make this move, would the moves since the last phase change match the
// pattern? If your move can be made legally multiple times in a row in a given
// move progression, implement interfaces.AllowMultipleInProgression() and
// return true.
func (d *Default) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	return d.legalWithStackConstraints(state, proposer, true)
}

// legalWithoutStackConstraints runs Default's ordinary legality chain except
// for its generic first-component stack check. It is intentionally private:
// framework move bases may use it only when their own Legal method immediately
// performs a stronger stack preflight.
func (d *Default) legalWithoutStackConstraints(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	return d.legalWithStackConstraints(state, proposer, false)
}

func (d *Default) legalWithStackConstraints(state boardgame.ImmutableState, proposer boardgame.PlayerIndex, checkStackConstraints bool) error {

	// Declarative-legality seam (design spec "prime guarantee"). These two
	// checks run FIRST THING and are pure sugar for un-opted-in moves:
	//   1. The boot probe (rule 4): during NewGameManager, the engine calls
	//      each opted-in move's Legal() against a probe. Reaching here proves
	//      the move's declarations are reachable; we record that and return
	//      nil immediately (the probe observes reachability, not a verdict).
	//   2. The opt-in plan path: if this move type opted in via
	//      WithLegalPreconditions, its assembled plan REPLACES this frozen chain —
	//      evaluated exactly once here, so a super-calling Legal() override
	//      composes its own imperative residue around the plan without
	//      double-evaluating the contributed atoms.
	// For a move that did NOT opt in, both calls are cheap no-ops and control
	// falls through to the byte-for-byte-unchanged frozen chain below.
	if manager := state.Manager(); manager != nil {
		if manager.LegalProbeActive() {
			return nil
		}
		if handled, err := manager.LegalEvaluatePlan(d.Name(), state, d.Info().ConcreteMove(), proposer); handled {
			return err
		}
	}

	if err := d.legalInPhase(state); err != nil {
		return err
	}

	if err := d.legalMoveInProgression(state, proposer); err != nil {
		return err
	}

	if checkStackConstraints {
		return d.legalStackConstraints(state)
	}
	return nil
}

// legalStackConstraints checks stack constraints for moves that declare
// source and destination stacks via WithSourceProperty/WithDestinationProperty.
// It reads the stacks by property name from ImmutableGameState, gets the first
// component from the source, and checks whether the destination's constraints
// would accept it. This gives early feedback at Legal() time, complementing
// the safety-net check in moveComponentImpl during Apply().
func (d *Default) legalStackConstraints(state boardgame.ImmutableState) error {
	config := d.CustomConfiguration()

	srcName, ok := config[configPropSourceProperty].(string)
	if !ok {
		return nil
	}
	dstName, ok := config[configPropDestinationProperty].(string)
	if !ok {
		return nil
	}

	// The actual check is extracted to core (boardgame.LegalStackConstraintsCheck,
	// legal_framework.go) so that this frozen chain and legal's
	// "stackConstraints" wrapper predicate (legal/catalog_framework.go) call
	// exactly one implementation. See that file's doc comment.
	return boardgame.LegalStackConstraintsCheck(state, srcName, dstName)
}

func (d *Default) legalPhases() []enum.EnumKey {
	val := d.CustomConfiguration()[configPropLegalPhases]
	keys, ok := val.([]enum.EnumKey)
	if !ok {
		return nil
	}
	return keys
}

type legalMoveProgressioner interface {
	legalMoveProgression() MoveProgressionGroup
}

func (d *Default) legalMoveProgression() MoveProgressionGroup {
	val := d.CustomConfiguration()[configPropLegalMoveProgression]
	group, ok := val.(MoveProgressionGroup)
	if !ok {
		return nil
	}
	return group
}

// currentPhaseInfo extracts both the ImmutableVal and EnumKey from the
// delegate's CurrentPhase. If CurrentPhase returns nil (no phase configured),
// returns (nil, 0).
func currentPhaseInfo(state boardgame.ImmutableState) (enum.ImmutableVal, enum.EnumKey) {
	val := state.Manager().Delegate().CurrentPhase(state)
	if val == nil {
		return nil, 0
	}
	return val, val.Value()
}

// legalInPhase will return a descriptive error if this move is not legal in
// the current phase of the game.
func (d *Default) legalInPhase(state boardgame.ImmutableState) error {

	// The actual check is extracted to core (boardgame.LegalInPhaseCheck,
	// legal_framework.go) so that this frozen chain and legal's "inPhase"
	// wrapper predicate (legal/catalog_framework.go) call exactly one
	// implementation. See that file's doc comment.
	return boardgame.LegalInPhaseCheck(state, d.legalPhases())
}

// historicalMovesSincePhaseTransition is memoized per (game, upToVersion,
// targetPhase) via game.LegalTapeMemo (boardgame/legal_memo.go, design spec
// §5's "Tape memoization" engine win) — this retires the TODO that used to
// live right here ("ideally we'd memoize this so all base moves for this
// game for this version could use the result... make sure the lifetime of
// the cache does not extend beyond the lifetime of the game, or is purged
// every so often"): the memo lives ON the *boardgame.Game, so it is
// garbage-collected with the game and is bounded to at most one
// (version, phase) pair resident at a time — see LegalTapeMemo's doc
// comment. Both this frozen chain (legalMoveInProgression, below) and the
// "inProgression" declarative predicate (catalog_framework.go) call this
// same method, so they share one tape walk per (game, version) by
// construction.
func (d *Default) historicalMovesSincePhaseTransition(game *boardgame.Game, upToVersion int, targetPhase enum.EnumKey) []*boardgame.MoveStorageRecord {
	return game.LegalTapeMemo(upToVersion, targetPhase, func() []*boardgame.MoveStorageRecord {
		return computeHistoricalMovesSincePhaseTransition(game, upToVersion, targetPhase)
	})
}

// computeHistoricalMovesSincePhaseTransition is the actual (uncached) tape
// walk; historicalMovesSincePhaseTransition above wraps it in the memo. See
// that method's doc comment.
func computeHistoricalMovesSincePhaseTransition(game *boardgame.Game, upToVersion int, targetPhase enum.EnumKey) []*boardgame.MoveStorageRecord {

	moves := game.MoveRecords(upToVersion)

	if len(moves) == 0 {
		return nil
	}

	//When generating this list of moves to compare against to test if the
	//move progression matches, we want to skip any moves who couldn't be part
	//of the progression. At this point we've already bailed early if the move
	//we're considering isn't legal in this phase. And since moves can only be
	//registered in one phase, and they must be this phase, that means that if
	//they don't have a move progression then we can skip them because they're
	//allowed to match at any point in the move progression.

	noProgressionMoveTypesMutex.RLock()
	noProgressionMoveTypes, ok := noProgressionMoveTypesByGame[game.Name()]
	noProgressionMoveTypesMutex.RUnlock()

	if !ok {

		noProgressionMoveTypes = make(map[string]bool)

		//Create the list!
		for _, move := range game.Moves() {

			progressionMove, ok := move.(legalMoveProgressioner)

			if ok {
				//If it has a legalMoveProgression method then we have to see
				//if it returns a move progression.
				if progression := progressionMove.legalMoveProgression(); progression != nil {
					continue
				}

			}
			noProgressionMoveTypes[move.Info().Name()] = true
		}

		noProgressionMoveTypesMutex.Lock()
		noProgressionMoveTypesByGame[game.Name()] = noProgressionMoveTypes
		noProgressionMoveTypesMutex.Unlock()
	}

	var keptMoves []*boardgame.MoveStorageRecord

	for i := len(moves) - 1; i >= 0; i-- {
		move := moves[i]

		if noProgressionMoveTypes[move.Name] {
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

// legalMoveInProgression is NOT an exhaustive check. It simply confirms that
// this specific point would be legitimate to apply. Note that sometimes this
// will return nil even when there really should be another move in front of
// this that could still apply; that other move should actually be applied due
// to ordering of moves in ProposeFixUpMove. Finally, note that technically for
// AllowMultipleInProgression moves, this relies on the sub-classes Legal()
// method terminating, becuase this method won't; because as far as the
// progression is concerned, it's legal, and it's the sub-class's Legal()
// method's job to decide it's no longer legal.
func (d *Default) legalMoveInProgression(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	group := d.legalMoveProgression()

	//If there is no legal move progression then moves are legal in the phase at any time
	if group == nil {
		return nil
	}

	_, currentPhase := currentPhaseInfo(state)

	historicalMoves := d.historicalMovesSincePhaseTransition(state.Game(), state.Version(), currentPhase)

	//Add ourselves ot the end of the tape, since we're proposing adding ourselves.
	historicalMoves = append(historicalMoves, &boardgame.MoveStorageRecord{
		Name: d.Name(),
	})

	// ctx carries state/proposer to matchTape so a StatefulMoveProgressionGroup
	// (e.g. RepeatFromProp, #644) anywhere in group's tree can resolve a
	// state-driven count — see groups.go's StatefulMoveProgressionGroup and
	// design spec §7's "named plumbing change". Move/Chest are left zero:
	// move-tape matching has never needed them, and the legal.Read this
	// package declares for "inProgression" (catalog_framework.go) doesn't
	// include any move.* path.
	ctx := legal.Context{State: state, ProposerPlayerIndex: proposer}

	return matchTape(group, movesToNames(historicalMoves), ctx)

}

func makeTape(moveNames []string) *MoveGroupHistoryItem {
	var tapeStart *MoveGroupHistoryItem
	var tapeEnd *MoveGroupHistoryItem

	for _, moveName := range moveNames {
		newItem := &MoveGroupHistoryItem{
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

// matchTape walks group against the tape built from historicalMoves, via
// satisfiedDispatch(group, tapeStart, ctx) — so a StatefulMoveProgressionGroup
// (e.g. RepeatFromProp, #644) anywhere in group's tree receives ctx, while a
// plain MoveProgressionGroup (everything before this task, and any
// third-party group) is evaluated exactly as before. See groups.go's
// StatefulMoveProgressionGroup doc comment for the full rationale.
func matchTape(group MoveProgressionGroup, historicalMoves []string, ctx legal.Context) error {

	tapeStart := makeTape(historicalMoves)

	rest, err := satisfiedDispatch(group, tapeStart, ctx)

	defaultErr := errors.NewFriendly("The move was not legal at this phase in the progression")

	if err != nil {
		return defaultErr.WithError(err.Error())
	}

	if rest != nil {
		return defaultErr.WithError("The progression only matched some of the proposed move history")
	}

	return nil
}

// stackName returns the name of the stack for helpTExt, name, etc based on the
// configPropName.
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
