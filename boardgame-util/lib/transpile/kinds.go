package transpile

import (
	"go/types"
	"strings"
)

// kind is how a Go value is represented in TypeScript: the proposal's
// "Types" table, plus the handles the facade uses for framework objects.
type kind int

const (
	// kPlain is a value of a type the table maps directly: int, bool,
	// string, float64, error, slices and maps of them.
	kPlain kind = iota
	// kState is boardgame.State or ImmutableState: always the one `state`.
	kState
	// kGame is the game's *gameState: no value; its properties go through
	// state.
	kGame
	// kDelegate is the game's delegate: no value.
	kDelegate
	// kPlayer is *playerState or a player sub-state: the player's index.
	kPlayer
	// kMove is one of the game's move types: the MoveInstance.
	kMove
	// kValues is a component's static values struct: a DynamicRef
	// { deck, index } (a Comp is one).
	kValues
	// kDynamic is a component's dynamic values struct: a DynamicRef.
	kDynamic
	// kStack is any stack: a StackRef.
	kStack
	// kComp is a ComponentInstance: Comp | null.
	kComp
	// kComponent is a boardgame.Component (a deck entry): a DynamicRef.
	kComponent
	// kTimer is a timer property: the engine's TimerHandle.
	kTimer
	// kEnumVal is an enum.Val property: no value; an owner and a property
	// name that the enum facade calls take.
	kEnumVal
	// kRand is *rand.Rand from state.Rand(): the state's GoRand.
	kRand
	// kMoveIface is a boardgame.Move from MoveByName: MoveInstance | null.
	kMoveIface
	// kVariant is a boardgame.Variant: the variant record.
	kVariant
	// kGameObject is *boardgame.Game (state.Game()): only a chain receiver.
	kGameObject
	// kManager is *boardgame.GameManager or the delegate interface reached
	// through it: only a chain receiver.
	kManager
	// kEnum is an enum.Enum (or RangeEnum, TreeEnum) package variable: the
	// enum's name.
	kEnum
	// kGraph is a grid graph package variable: the facade's GraphRef.
	kGraph
)

const (
	boardgamePath = "github.com/jkomoros/boardgame"
	movesPath     = boardgamePath + "/moves"
	basePath      = boardgamePath + "/base"
	enumPath      = boardgamePath + "/enum"
)

func namedOf(t types.Type) (*types.Named, bool) {
	t = types.Unalias(t)
	pointer := false
	if p, ok := t.(*types.Pointer); ok {
		t = types.Unalias(p.Elem())
		pointer = true
	}
	named, _ := t.(*types.Named)
	return named, pointer
}

func pkgPathOf(named *types.Named) string {
	if named == nil || named.Obj().Pkg() == nil {
		return ""
	}
	return named.Obj().Pkg().Path()
}

// qualifiedType is "path.Name" of a named type.
func qualifiedType(named *types.Named) string {
	return pkgPathOf(named) + "." + named.Obj().Name()
}

// kindOf classifies a Go type.
func (p *Package) kindOf(t types.Type) kind {
	if t == nil {
		return kPlain
	}
	named, pointer := namedOf(t)
	if named == nil {
		return kPlain
	}
	if named.Obj().Pkg() == p.tpkg {
		if !pointer {
			return kPlain
		}
		switch {
		case named == p.gameState:
			return kGame
		case named == p.playerState:
			return kPlayer
		case named == p.delegate:
			return kDelegate
		case p.isMoveType(named):
			return kMove
		case p.embedsFrom(named, basePath, "ComponentValues"):
			return kValues
		case p.embedsFrom(named, basePath, "SubState"):
			return kDynamic
		}
		return kPlain
	}
	switch qualifiedType(named) {
	case boardgamePath + ".State", boardgamePath + ".ImmutableState":
		return kState
	case boardgamePath + ".ImmutableSubState", boardgamePath + ".SubState", boardgamePath + ".ConfigurableSubState":
		return kPlayer
	case boardgamePath + ".Stack", boardgamePath + ".ImmutableStack", boardgamePath + ".SizedStack", boardgamePath + ".ImmutableSizedStack",
		boardgamePath + ".MergedStack", boardgamePath + ".ImmutableMergedStack":
		return kStack
	case boardgamePath + ".ComponentInstance", boardgamePath + ".ImmutableComponentInstance":
		return kComp
	case boardgamePath + ".Component":
		return kComponent
	case boardgamePath + ".ComponentValues":
		return kValues
	case boardgamePath + ".Timer", boardgamePath + ".ImmutableTimer":
		return kTimer
	case boardgamePath + ".Move":
		return kMoveIface
	case boardgamePath + ".Variant":
		return kVariant
	case boardgamePath + ".Game":
		return kGameObject
	case boardgamePath + ".GameManager", boardgamePath + ".GameDelegate":
		return kManager
	case enumPath + ".Val", enumPath + ".ImmutableVal", enumPath + ".RangeVal", enumPath + ".ImmutableRangeVal":
		return kEnumVal
	case enumPath + ".Enum", enumPath + ".RangeEnum", enumPath + ".TreeEnum":
		return kEnum
	case enumPath + "/graph.Graph", enumPath + "/graph.EnumGraph":
		return kGraph
	case "math/rand.Rand":
		return kRand
	case boardgamePath + "/components/playingcards.Card", boardgamePath + "/components/dice.Value":
		return kValues
	case boardgamePath + "/components/dice.DynamicValue":
		return kDynamic
	}
	return kPlain
}

// isMoveType reports whether named (a game type) is a move: its pointer has
// Apply and HelpText methods.
func (p *Package) isMoveType(named *types.Named) bool {
	for _, name := range []string{"Apply", "HelpText", "Legal"} {
		obj, _, _ := types.LookupFieldOrMethod(types.NewPointer(named), true, p.tpkg, name)
		if _, ok := obj.(*types.Func); !ok {
			return false
		}
	}
	return true
}

// embedsFrom reports whether named embeds (at any depth) the type name of
// package path.
func (p *Package) embedsFrom(named *types.Named, path, name string) bool {
	obj, _, _ := types.LookupFieldOrMethod(named, true, p.tpkg, name)
	field, ok := obj.(*types.Var)
	if !ok || !field.Embedded() {
		return false
	}
	fieldNamed, _ := namedOf(field.Type())
	return fieldNamed != nil && pkgPathOf(fieldNamed) == path
}

// erasedKind reports whether values of k have no TypeScript value.
func erasedKind(k kind) bool {
	switch k {
	case kGame, kDelegate, kState, kEnumVal, kGameObject, kManager:
		return true
	}
	return false
}

// tsType is the TypeScript type of a Go type, or "" when the emitter has no
// mapping (a refusal).
func (p *Package) tsType(t types.Type) string {
	t = types.Unalias(t)
	switch p.kindOf(t) {
	case kPlayer:
		return "number"
	case kMove:
		return "MoveInstance"
	case kValues, kDynamic, kComponent:
		return "DynamicRef"
	case kStack:
		return "StackRef"
	case kComp:
		return "Comp | null"
	case kTimer:
		return "TimerHandle"
	case kRand:
		return "GoRand"
	case kMoveIface:
		return "MoveInstance | null"
	case kVariant:
		return "Readonly<Record<string, string>>"
	case kState:
		return "RulesState"
	case kEnum:
		return "string"
	case kGraph:
		return "GraphRef"
	case kGame, kDelegate, kEnumVal, kGameObject, kManager:
		return ""
	}
	if named, ok := t.(*types.Named); ok {
		if named.Obj().Pkg() == nil && named.Obj().Name() == "error" {
			return "string | null"
		}
	}
	switch u := t.Underlying().(type) {
	case *types.Basic:
		switch {
		case u.Info()&types.IsBoolean != 0:
			return "boolean"
		case u.Info()&types.IsString != 0:
			return "string"
		case u.Info()&(types.IsInteger|types.IsFloat) != 0:
			return "number"
		case u.Kind() == types.UntypedNil:
			return "null"
		}
	case *types.Slice:
		elem := p.tsType(u.Elem())
		if elem == "" {
			return ""
		}
		if strings.Contains(elem, " ") {
			elem = "(" + elem + ")"
		}
		return elem + "[] | null"
	case *types.Map:
		key, value := p.tsType(u.Key()), p.tsType(u.Elem())
		if key == "" || value == "" {
			return ""
		}
		return "Map<" + key + ", " + value + ">"
	case *types.Tuple:
		var parts []string
		for i := 0; i < u.Len(); i++ {
			part := p.tsType(u.At(i).Type())
			if part == "" {
				return ""
			}
			parts = append(parts, part)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *types.Interface:
		if u.NumMethods() == 1 && u.Method(0).Name() == "Error" {
			return "string | null"
		}
	}
	return ""
}

// zeroValue is the TypeScript zero value of a Go type, or "".
func (p *Package) zeroValue(t types.Type) string {
	switch p.kindOf(t) {
	case kComp, kMoveIface:
		return "null"
	case kPlain:
	default:
		return ""
	}
	if named, ok := types.Unalias(t).(*types.Named); ok && named.Obj().Pkg() == nil && named.Obj().Name() == "error" {
		return "null"
	}
	switch u := t.Underlying().(type) {
	case *types.Basic:
		switch {
		case u.Info()&types.IsBoolean != 0:
			return "false"
		case u.Info()&types.IsString != 0:
			return "''"
		case u.Info()&(types.IsInteger|types.IsFloat) != 0:
			return "0"
		}
	case *types.Slice:
		return "null"
	case *types.Interface:
		return "null"
	}
	return ""
}

func isError(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)
	return ok && named.Obj().Pkg() == nil && named.Obj().Name() == "error"
}

func isString(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Info()&types.IsString != 0
}

func isInteger(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Info()&types.IsInteger != 0
}

func isFloat(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Info()&types.IsFloat != 0
}

func isBool(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Info()&types.IsBoolean != 0
}

func isDuration(t types.Type) bool {
	named, _ := namedOf(t)
	return named != nil && pkgPathOf(named) == "time" && named.Obj().Name() == "Duration"
}
