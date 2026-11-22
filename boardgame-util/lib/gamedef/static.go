package gamedef

import "errors"

// StaticInfo is the static half of the exporter: what the game package's
// source says that the live manager cannot. Static produces it from
// go/packages and go/types; Export joins it with the run-time half.
//
// It is the interface between workstream 2 (static.go) and workstream 1
// (export.go). It is not itself serialized into GameData; Export copies its
// parts into the places GameData names.
type StaticInfo struct {
	// Package is the import path that was loaded.
	Package string
	// PackageName is the package's name, the prefix of every FnRef.
	PackageName string
	// Moves is every move type the package declares or configures, keyed
	// by MoveStatic.GoType ("blackjack.moveCurrentPlayerHit"). Export finds
	// each configured move's entry from the reflected type of its move.
	Moves map[string]MoveStatic
	// Traits is what the state structs embed and which interface methods
	// the game declares.
	Traits Traits
	// Computed is computed property name to the function that computes it.
	Computed map[string]FnRef
	// Predicates is custom predicate name to its constructor.
	Predicates map[string]FnRef
	// Counters is a custom progression counter's closure, by the position
	// Export reports for it ("<move name>/<path>"), for a counter that fits
	// no stock family. Empty in v1, where such a counter is refused
	// (custom-counter).
	Counters map[string]FnRef
	// SetUp and End carry the FnRefs of the hooks the game declares; the
	// non-FnRef fields of End (LowScoreWins) come from the run-time half.
	SetUp SetUpData
	End   EndData
	// CurrentPlayerIndex is the delegate's override, if declared.
	CurrentPlayerIndex FnRef
	// Functions is every function reaching engine output, sorted by Func.
	Functions []FunctionEntry
}

// MoveStatic is one move type, as its source declares it.
type MoveStatic struct {
	// GoType is "pkg.typeName".
	GoType string
	// StockType is the first embedded type declared in package moves, by
	// its name there ("CurrentPlayer", "StartPhase", ...), found by walking
	// the embedding chain with types.LookupFieldOrMethod. "" when the only
	// ancestor is moves.Default (then MoveData.Stock is null).
	StockType string
	// Bodies are the methods whose declaring package is the game's: Apply,
	// Legal, LegalCustom, DefaultsForState. A promoted method is absent.
	Bodies MoveBodies
}

// Static loads the game package at importPath and returns its static half.
//
// Not implemented: workstream 2 owns it.
func Static(importPath string) (*StaticInfo, error) {
	return nil, errors.New("gamedef: static half not implemented")
}
