/*
Package refusal is the closed list of reasons the translator refuses a Go
function, and the record of one refusal.

It is a leaf package so that both the translator (lib/transpile) and the
exported definition (lib/gamedef, whose FunctionEntry carries refusals) can
import it without importing each other. lib/transpile re-exports everything
here under the same names.

The list is fabric/TRANSPILE-PROPOSAL.md, "The translator", "Refusals". A
new reason is a change to that table and to this file together.
*/
package refusal

// Refusal is one reason the translator would not translate one function.
// The JSON keys are the Go field names, capitalized like the rest of the
// exported definition.
type Refusal struct {
	// Func is the refused function's FnRef: "pkg.(*recv).Method" or "pkg.fn".
	Func string
	// Reason is one of the Reason constants below.
	Reason string
	// Detail is the specific name the reason is about (a qualified callee, a
	// facade name, a type, a package variable). The ratchet counts by it.
	Detail string
	// File is the source file, relative to the package directory.
	File string
	// Line and Column are 1-based, as go/token reports them.
	Line   int
	Column int
}

// The closed list of refusal reasons.
const (
	// UnknownCall is a call outside the package with no API-table entry.
	// Detail: the qualified name.
	UnknownCall = "unknown-call"
	// EngineGap is a call the table knows whose engine counterpart does not
	// exist yet. Detail: the facade name.
	EngineGap = "engine-gap"
	// UnknownType is a type outside the type table.
	UnknownType = "unknown-type"
	// PackageVarRead is a read of a package variable that is neither
	// constant-initialized nor a recognized framework object.
	PackageVarRead = "package-var-read"
	// PackageVarWrite is any write to a package variable.
	PackageVarWrite = "package-var-write"
	// MapRange is a range over a map.
	MapRange = "map-range"
	// SliceAlias is a slice use whose aliasing the subset does not model.
	SliceAlias = "slice-alias"
	// StringBytes is indexing or slicing a string.
	StringBytes = "string-bytes"
	// StructCopy is a struct value that is copied or mutated.
	StructCopy = "struct-copy"
	// PointerLocal is a pointer to a local.
	PointerLocal = "pointer-local"
	// SizedInt is a sized integer type (int8, uint32, ...).
	SizedInt = "sized-int"
	// FixedArray is a fixed-length array type.
	FixedArray = "fixed-array"
	// InterfaceDispatch is a method call through an interface outside the table.
	InterfaceDispatch = "interface-dispatch"
	// TypeSwitch is a type switch.
	TypeSwitch = "type-switch"
	// ClosureEscape is a function value other than a less callback or a
	// literal passed straight to a known call.
	ClosureEscape = "closure-escape"
	// Generic is a type parameter.
	Generic = "generic"
	// Goroutine is a go statement, a select, or a channel.
	Goroutine = "goroutine"
	// Defer is a defer statement.
	Defer = "defer"
	// Goto is a goto statement.
	Goto = "goto"
	// Fallthrough is a fallthrough statement.
	Fallthrough = "fallthrough"
	// Reflection is a use of package reflect.
	Reflection = "reflection"
	// Unsafe is a use of package unsafe.
	Unsafe = "unsafe"
	// ForeignRand is randomness from anything but state.Rand().
	ForeignRand = "foreign-rand"
	// IO is wall clock, I/O, logging or encoding/json.
	IO = "io"
	// SubMsDuration is a time.Duration constant that is not a whole number
	// of milliseconds.
	SubMsDuration = "sub-ms-duration"
	// PredicateConstructor is a custom legal predicate constructor: a kind
	// of body with no v1 rule.
	PredicateConstructor = "predicate-constructor"
	// CustomCounter is a progression counter that fits no stock family: a
	// kind of body with no v1 rule.
	CustomCounter = "custom-counter"
	// CalleeRefused is an in-package callee that was itself refused.
	// Detail: its FnRef.
	CalleeRefused = "callee-refused"
)

// Reasons is every valid reason, in the order of the proposal's table.
func Reasons() []string {
	return []string{
		UnknownCall,
		EngineGap,
		UnknownType,
		PackageVarRead,
		PackageVarWrite,
		MapRange,
		SliceAlias,
		StringBytes,
		StructCopy,
		PointerLocal,
		SizedInt,
		FixedArray,
		InterfaceDispatch,
		TypeSwitch,
		ClosureEscape,
		Generic,
		Goroutine,
		Defer,
		Goto,
		Fallthrough,
		Reflection,
		Unsafe,
		ForeignRand,
		IO,
		SubMsDuration,
		PredicateConstructor,
		CustomCounter,
		CalleeRefused,
	}
}

// ValidReason reports whether reason is in the closed list.
func ValidReason(reason string) bool {
	switch reason {
	case UnknownCall, EngineGap, UnknownType, PackageVarRead, PackageVarWrite,
		MapRange, SliceAlias, StringBytes, StructCopy, PointerLocal, SizedInt,
		FixedArray, InterfaceDispatch, TypeSwitch, ClosureEscape, Generic,
		Goroutine, Defer, Goto, Fallthrough, Reflection, Unsafe, ForeignRand,
		IO, SubMsDuration, PredicateConstructor, CustomCounter, CalleeRefused:
		return true
	}
	return false
}
