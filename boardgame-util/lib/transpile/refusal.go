/*
Package transpile translates the small imperative Go functions of a game
into TypeScript over a restricted subset, and refuses the rest with a
machine-readable reason. See fabric/TRANSPILE-PROPOSAL.md, "The translator".

This file is the refusal list. The type and the reasons live in the leaf
package lib/transpile/refusal, so that lib/gamedef can carry refusals
without importing this package; they are re-exported here so the
translator's callers need only one import.
*/
package transpile

import "github.com/jkomoros/boardgame/boardgame-util/lib/transpile/refusal"

// Refusal is one reason the translator would not translate one function:
// { Func, Reason, Detail, File, Line, Column }.
type Refusal = refusal.Refusal

// The closed list of refusal reasons; see package refusal for each meaning.
const (
	ReasonUnknownCall          = refusal.UnknownCall
	ReasonEngineGap            = refusal.EngineGap
	ReasonUnknownType          = refusal.UnknownType
	ReasonPackageVarRead       = refusal.PackageVarRead
	ReasonPackageVarWrite      = refusal.PackageVarWrite
	ReasonMapRange             = refusal.MapRange
	ReasonSliceAlias           = refusal.SliceAlias
	ReasonStringBytes          = refusal.StringBytes
	ReasonStructCopy           = refusal.StructCopy
	ReasonPointerLocal         = refusal.PointerLocal
	ReasonSizedInt             = refusal.SizedInt
	ReasonFixedArray           = refusal.FixedArray
	ReasonInterfaceDispatch    = refusal.InterfaceDispatch
	ReasonTypeSwitch           = refusal.TypeSwitch
	ReasonClosureEscape        = refusal.ClosureEscape
	ReasonGeneric              = refusal.Generic
	ReasonGoroutine            = refusal.Goroutine
	ReasonDefer                = refusal.Defer
	ReasonGoto                 = refusal.Goto
	ReasonFallthrough          = refusal.Fallthrough
	ReasonReflection           = refusal.Reflection
	ReasonUnsafe               = refusal.Unsafe
	ReasonForeignRand          = refusal.ForeignRand
	ReasonIO                   = refusal.IO
	ReasonSubMsDuration        = refusal.SubMsDuration
	ReasonPredicateConstructor = refusal.PredicateConstructor
	ReasonCustomCounter        = refusal.CustomCounter
	ReasonCalleeRefused        = refusal.CalleeRefused
)

// Reasons is every valid reason, in the order of the proposal's table.
func Reasons() []string { return refusal.Reasons() }

// ValidReason reports whether reason is in the closed list.
func ValidReason(reason string) bool { return refusal.ValidReason(reason) }
