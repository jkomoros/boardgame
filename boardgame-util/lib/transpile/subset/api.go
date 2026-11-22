package subset

import "sort"

// Entry kinds (Entry.Kind): how the translator treats a call the table knows.
const (
	// KindCall is an ordinary call: the translator emits Facade.
	KindCall = "call"
	// KindChain is valid only as the receiver of another table call
	// (state.Game() in state.Game().MoveByName(n)); on its own it is
	// unknown-call.
	KindChain = "chain"
	// KindErased has no value of its own: the game and player handles
	// (state.GameState(), a sub-state's State()) are erased like
	// concreteStates, and the property access goes through state.
	KindErased = "erased"
	// KindDelegate is a method of boardgame.GameDelegate called through
	// state.Manager().Delegate(): a direct call of the game's own method.
	KindDelegate = "delegate"
	// KindFunc is a call that takes a function argument (a less callback);
	// a function literal passed straight to it is not a closure-escape.
	KindFunc = "func"
	// KindFormat is a fmt call whose format string must use only the verbs
	// sprintf ports (%d %s %v %q).
	KindFormat = "format"
)

// Entry is one row of the API table: Go calls the translator knows, the
// facade call it emits, and whether the engine has that call today (see
// EngineGaps).
type Entry struct {
	// Go lists the qualified Go names the row covers, as QualifiedName spells
	// them: "pkg.Func" for a function, "(pkg.Type).Method" for a method,
	// where Type is the type that declares the method (an interface for an
	// interface method, without a pointer star).
	Go []string
	// Facade is the call the translator emits (go/api.ts and friends), or a
	// short description for a row with no call of its own.
	Facade string
	// Kind is one of the Kind constants.
	Kind string
	// Source is "proposal" for a row of fabric/TRANSPILE-PROPOSAL.md's API
	// table, or "added: <why>" for a row the examples needed that the
	// proposal's table does not list.
	Source string
}

// Gap reports whether the engine lacks this entry's facade call today.
func (e Entry) Gap() bool {
	_, gap := EngineGaps[e.Facade]
	return gap
}

const proposal = "proposal"

// Table is the API table, data only. The translator (lib/transpile) reads
// the same rows. Keep it sorted by row meaning, not alphabetically; Lookup
// indexes it.
var Table = []Entry{
	// State and players.
	{Go: []string{"(boardgame.ImmutableState).ImmutablePlayerStates", "(boardgame.State).PlayerStates"}, Facade: "playerStates", Kind: KindCall, Source: proposal + ": len(state.PlayerStates()) is state.numPlayers(); ranging over it is a snapshot of player indexes"},
	{Go: []string{"(boardgame.ImmutableState).CurrentPlayerIndex", "(boardgame.State).CurrentPlayer", "(boardgame.ImmutableState).ImmutableCurrentPlayer"}, Facade: "currentPlayerIndex", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.PlayerIndex).EnsureValid"}, Facade: "ensureValid", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.PlayerIndex).Next"}, Facade: "next", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.PlayerIndex).Previous"}, Facade: "previous", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.PlayerIndex).Valid"}, Facade: "valid", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.PlayerIndex).WithinBounds"}, Facade: "withinBounds", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ImmutableState).Rand", "(boardgame.State).Rand"}, Facade: "rand", Kind: KindCall, Source: proposal + ": state.Rand() is the only legal source"},
	{Go: []string{"(*rand.Rand).Intn"}, Facade: "rand().intn", Kind: KindCall, Source: proposal},
	{Go: []string{"(*rand.Rand).Perm"}, Facade: "rand().perm", Kind: KindCall, Source: proposal},
	{Go: []string{"(*rand.Rand).Shuffle"}, Facade: "rand().shuffle", Kind: KindFunc, Source: proposal},
	{Go: []string{"(*rand.Rand).Int63n"}, Facade: "rand().int63n", Kind: KindCall, Source: proposal},
	{Go: []string{"(*rand.Rand).Float64"}, Facade: "rand().float64", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ImmutableState).Game"}, Facade: "state.Game()", Kind: KindChain, Source: proposal + ": only as state.Game().Variant(), MoveByName or MoveByNameForState"},
	{Go: []string{"(*boardgame.Game).Variant"}, Facade: "variantValue", Kind: KindCall, Source: proposal},
	{Go: []string{"(*boardgame.Game).MoveByName"}, Facade: "moveByName", Kind: KindCall, Source: proposal},
	{Go: []string{"(*boardgame.Game).MoveByNameForState"}, Facade: "newMove", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ImmutableState).Manager"}, Facade: "state.Manager()", Kind: KindChain, Source: proposal + ": only as state.Manager().Delegate().M(state)"},
	{Go: []string{"(*boardgame.GameManager).Delegate"}, Facade: "Manager().Delegate()", Kind: KindChain, Source: proposal},
	{Go: []string{"(boardgame.GameDelegate).*"}, Facade: "delegate dispatch", Kind: KindDelegate, Source: proposal + ": a direct call of the game's own M"},
	{Go: []string{"(boardgame.ImmutableState).ImmutableGameState", "(boardgame.State).GameState"}, Facade: "game handle", Kind: KindErased, Source: "added: the Types table gives *gameState no value (property access goes through state); this is concreteStates inlined"},
	{Go: []string{"(boardgame.ImmutableSubState).ImmutableState", "(boardgame.SubState).State", "(*base.SubState).State", "(*base.SubState).ImmutableState"}, Facade: "state", Kind: KindErased, Source: "added: a sub-state's State() is the one state the function already has (memory's TurnDone and computed helper call it before concreteStates)"},

	// Behaviors.
	{Go: []string{"behaviors.PlayerIsInactive"}, Facade: "playerIsInactive", Kind: KindCall, Source: proposal},
	{Go: []string{"behaviors.PlayerIsEliminated"}, Facade: "playerIsEliminated", Kind: KindCall, Source: proposal},
	{Go: []string{"(*behaviors.PlayerElimination).SetEliminated"}, Facade: "setEliminated", Kind: KindCall, Source: proposal},
	{Go: []string{"(*behaviors.InactivePlayer).SetPlayerInactive"}, Facade: "setPlayerInactive", Kind: KindCall, Source: proposal},
	{Go: []string{"(*behaviors.Seat).SetSeatFilled"}, Facade: "setSeatFilled", Kind: KindCall, Source: proposal},
	{Go: []string{"(*behaviors.PhaseBehavior).SetCurrentPhase"}, Facade: "setPhase", Kind: KindCall, Source: proposal},

	// Enums.
	{Go: []string{"(enum.ImmutableVal).Value"}, Facade: "enumKey", Kind: KindCall, Source: proposal},
	{Go: []string{"(enum.Val).SetValue"}, Facade: "setEnumKey", Kind: KindCall, Source: proposal},
	{Go: []string{"(enum.ImmutableVal).Equals"}, Facade: "enumEquals", Kind: KindCall, Source: proposal},
	{Go: []string{"(enum.EnumKey).Int"}, Facade: "enum key as int", Kind: KindCall, Source: "added: EnumKey and int are both number; Int() is the identity conversion the Expressions section allows"},
	{Go: []string{"(enum.ImmutableRangeVal).RangeValue", "(enum.RangeVal).RangeValue"}, Facade: "rangeValue", Kind: KindCall, Source: proposal},
	{Go: []string{"(enum.RangeEnum).RangeToValue"}, Facade: "rangeToValue", Kind: KindCall, Source: proposal},
	{Go: []string{"(enum.RangeEnum).ValueToRange"}, Facade: "valueToRange", Kind: KindCall, Source: proposal},
	{Go: []string{"(graph.Graph).NeighborsByKey", "(*graph.EnumGraph).NeighborsByKey"}, Facade: "neighborsByKey", Kind: KindCall, Source: proposal},

	// Stacks.
	{Go: []string{"(boardgame.ImmutableStack).Len", "(boardgame.ImmutableStack).NumComponents", "(boardgame.ImmutableStack).SlotsRemaining", "(boardgame.ImmutableStack).MaxSize"}, Facade: "stackView", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).ComponentAt", "(boardgame.ImmutableStack).ImmutableComponentAt"}, Facade: "componentAt", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).First", "(boardgame.ImmutableStack).ImmutableFirst"}, Facade: "first", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).Last", "(boardgame.ImmutableStack).ImmutableLast"}, Facade: "last", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).Components", "(boardgame.ImmutableStack).ImmutableComponents"}, Facade: "components", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).ComponentAtKey", "(boardgame.ImmutableStack).ImmutableComponentAtKey"}, Facade: "componentAtKey", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).SwapComponentsByKey"}, Facade: "swapComponentsByKey", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).MoveAllTo"}, Facade: "moveAllTo", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).MoveCountTo"}, Facade: "moveCountTo", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).MayMoveAllTo", "(boardgame.ImmutableStack).MayMoveAllTo"}, Facade: "mayMoveAllTo", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).MayMoveCountTo", "(boardgame.ImmutableStack).MayMoveCountTo"}, Facade: "mayMoveCountTo", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).Shuffle"}, Facade: "shuffle", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).PublicShuffle"}, Facade: "publicShuffle", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).SwapComponents"}, Facade: "swapComponents", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Stack).SortComponents"}, Facade: "sortComponents", Kind: KindFunc, Source: proposal},
	{Go: []string{"(boardgame.SizedStack).SetSize", "(boardgame.Stack).SetSize"}, Facade: "setSize", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.SizedStack).ExpandSize", "(boardgame.Stack).ExpandSize"}, Facade: "expandSize", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.SizedStack).ContractSize", "(boardgame.Stack).ContractSize"}, Facade: "contractSize", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.SizedStack).SizeToFit", "(boardgame.Stack).SizeToFit"}, Facade: "sizeToFit", Kind: KindCall, Source: proposal},

	// Components.
	{Go: []string{"(boardgame.Component).Values"}, Facade: "values", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ComponentInstance).DynamicValues", "(boardgame.ImmutableComponentInstance).ImmutableDynamicValues"}, Facade: "dynamic", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ComponentInstance).MoveTo"}, Facade: "moveTo", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ComponentInstance).SecretMoveTo"}, Facade: "secretMoveTo", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ComponentInstance).MoveToFirstSlot"}, Facade: "moveToFirstSlot", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ComponentInstance).MoveToLastSlot"}, Facade: "moveToLastSlot", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ComponentInstance).MoveToNextSlot"}, Facade: "moveToNextSlot", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ComponentInstance).SlideToFirstSlot"}, Facade: "slideToFirstSlot", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ComponentInstance).SlideToLastSlot"}, Facade: "slideToLastSlot", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ImmutableComponentInstance).MayMoveTo"}, Facade: "mayMoveTo", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ImmutableComponentInstance).MayMoveToSlot"}, Facade: "mayMoveToSlot", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.ImmutableComponentInstance).ContainingImmutableStack", "(boardgame.ComponentInstance).ContainingStack"}, Facade: "containingStack", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Component).Deck"}, Facade: "c.deck", Kind: KindChain, Source: "added: only as c.Deck().Name(); a Comp handle carries its deck"},
	{Go: []string{"(*boardgame.Deck).Name"}, Facade: "deck name", Kind: KindCall, Source: "added: the deck field of the Comp handle"},
	{Go: []string{"(boardgame.ComponentValues).ContainingComponent", "(*base.ComponentValues).ContainingComponent"}, Facade: "component of values", Kind: KindChain, Source: "added: a component method's receiver is the component's handle; only as ContainingComponent().ImmutableInstance(state) or Instance(state)"},
	{Go: []string{"(boardgame.Component).ImmutableInstance", "(boardgame.Component).Instance"}, Facade: "componentInstance", Kind: KindCall, Source: "added: the handle itself, componentInstance(state, deck, index), located where the component is now"},

	// Timers.
	{Go: []string{"(boardgame.ImmutableTimer).Active"}, Facade: "timer.active", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Timer).Start"}, Facade: "timer.start", Kind: KindCall, Source: proposal},
	{Go: []string{"(boardgame.Timer).Cancel"}, Facade: "timer.cancel", Kind: KindCall, Source: proposal},

	// Errors, strings and formatting.
	{Go: []string{"errors.New"}, Facade: "the string", Kind: KindCall, Source: proposal},
	{Go: []string{"(error).Error"}, Facade: "the string", Kind: KindCall, Source: proposal},
	{Go: []string{"fmt.Errorf", "fmt.Sprintf"}, Facade: "sprintf", Kind: KindFormat, Source: proposal},
	{Go: []string{"legal.Errorf"}, Facade: "legalError", Kind: KindCall, Source: proposal},
	{Go: []string{"strconv.Itoa"}, Facade: "itoa", Kind: KindCall, Source: proposal},
	{Go: []string{"strconv.Atoi"}, Facade: "atoi", Kind: KindCall, Source: proposal},
	{Go: []string{"strings.Join"}, Facade: "join", Kind: KindCall, Source: proposal},
	{Go: []string{"strings.Repeat"}, Facade: "repeat", Kind: KindCall, Source: proposal},
	{Go: []string{"strings.Split"}, Facade: "split", Kind: KindCall, Source: proposal},
	{Go: []string{"strings.ToLower"}, Facade: "toLower", Kind: KindCall, Source: proposal},
	{Go: []string{"strings.TrimSpace"}, Facade: "trimSpace", Kind: KindCall, Source: proposal},
	{Go: []string{"strings.HasPrefix"}, Facade: "hasPrefix", Kind: KindCall, Source: proposal},
	{Go: []string{"strings.HasSuffix"}, Facade: "hasSuffix", Kind: KindCall, Source: proposal},
	{Go: []string{"strings.Contains"}, Facade: "contains", Kind: KindCall, Source: proposal},
	{Go: []string{"sort.Sort"}, Facade: "gosort", Kind: KindCall, Source: proposal + " (Semantics): gosort.ts ports sort.Sort"},
	{Go: []string{"sort.Stable", "sort.SliceStable"}, Facade: "stableSort", Kind: KindFunc, Source: proposal + " (Semantics): Array.prototype.sort is stable"},

	// Libraries.
	{Go: []string{"(*dice.DynamicValue).Roll"}, Facade: "lib/dice.ts roll", Kind: KindCall, Source: proposal},
}

// StockParts are the stock move methods a game move may call on its
// embedded stock ancestor, and what each becomes (the API table's
// m.<Stock>.X rows).
var StockParts = map[string]string{
	"Apply":            "stockApply",
	"Legal":            "base()",
	"DefaultsForState": "dropped when it is the first statement",
}

var tableIndex = func() map[string]int {
	index := make(map[string]int)
	for i, entry := range Table {
		for _, name := range entry.Go {
			index[name] = i
		}
	}
	return index
}()

// Lookup returns the table row for a qualified Go name (see QualifiedName).
// A method of boardgame.GameDelegate matches the "(boardgame.GameDelegate).*" row.
func Lookup(qualified string) (Entry, bool) {
	if i, ok := tableIndex[qualified]; ok {
		return Table[i], true
	}
	if len(qualified) > len(delegatePrefix) && qualified[:len(delegatePrefix)] == delegatePrefix {
		return Table[tableIndex[delegatePrefix+"*"]], true
	}
	return Entry{}, false
}

const delegatePrefix = "(boardgame.GameDelegate)."

// GoNames is every qualified Go name the table covers, sorted.
func GoNames() []string {
	names := make([]string, 0, len(tableIndex))
	for name := range tableIndex {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// FrameworkConstructors are the functions whose result a package variable
// may hold and still be read (the package-variable rule): each builds a
// framework object the translator can look up by name or rebuild from the
// exported definition. Arguments must be constants, other recognized
// framework variables, or FrameworkFuncValues.
var FrameworkConstructors = map[string]bool{
	"enum.NewSet":                                true,
	"(*enum.Set).MustAdd":                        true,
	"(*enum.Set).MustAddRange":                   true,
	"(*enum.Set).MustAddTree":                    true,
	"(*enum.Set).MustCombine":                    true,
	"graph.MustNewEnumGridConnectedness":         true,
	"graph.MustNewGridConnectedness":             true,
	"(enum.Enum).MustNewVal":                     true,
	"(enum.Enum).MustNewImmutableVal":            true,
	"behaviors.ColorEnum":                        true,
	"(*behaviors.ColorPalette).MustAddToEnumSet": true,
}

// FrameworkFuncValues are framework functions that may be passed as values
// to a FrameworkConstructor (grid directions).
var FrameworkFuncValues = map[string]bool{
	"graph.DirectionUp":            true,
	"graph.DirectionDown":          true,
	"graph.DirectionLeft":          true,
	"graph.DirectionRight":         true,
	"graph.DirectionPerpendicular": true,
	"graph.DirectionDiagonal":      true,
}
