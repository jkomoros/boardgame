package legal

import (
	"fmt"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

/*
This file gives the catalog's predicates typed accessors (int, bool,
ImmutableStack, enum.ImmutableVal) for a "kind.Property" LegalPropPath. Path
GRAMMAR — what "game.X"/"player.X"/"players[*].X"/"move.X" mean, and how each
resolves against a Context — lives entirely in core now
(boardgame.LegalContext.ResolvePath, backed by resolveLegalPath and
parseLegalPath in legal_path.go). This file used to duplicate that grammar
with its own parser and reader lookup; it no longer does. Every helper below
is a thin type-switch over what ctx.ResolvePath returns, nothing more. This
is deliberate: boot validation (validateLegalPath, run once at
NewGameManager) and evaluation (ctx.ResolvePath, run every time a predicate's
Evaluate fires) now share ONE implementation of the grammar, so a path that
validates at boot can never be parsed differently at evaluation time — see
legal_path.go's ResolvePath doc comment for the full rationale. One concrete
consequence: a "players[*].X" path resolved through a leaf catalog predicate
now gets core's specific "quantifier-only, cannot be resolved to a single
value" error for free, rather than this package's old three-kind parser
rejecting it as simply "unsupported path kind".
*/

// resolveIntPath resolves path (a "kind.Property" LegalPropPath string) as
// an int against ctx, via ctx.ResolvePath.
func resolveIntPath(path string, ctx Context) (int, error) {
	val, propType, err := ctx.ResolvePath(PropPath(path))
	if err != nil {
		return 0, err
	}
	if propType != boardgame.TypeInt {
		return 0, fmt.Errorf("legal: path %q is not an int property", path)
	}
	i, ok := val.(int)
	if !ok {
		return 0, fmt.Errorf("legal: path %q resolved to an int-typed property but its value was not an int (%T)", path, val)
	}
	return i, nil
}

// resolveBoolPath resolves path as a bool against ctx, via ctx.ResolvePath.
func resolveBoolPath(path string, ctx Context) (bool, error) {
	val, propType, err := ctx.ResolvePath(PropPath(path))
	if err != nil {
		return false, err
	}
	if propType != boardgame.TypeBool {
		return false, fmt.Errorf("legal: path %q is not a bool property", path)
	}
	b, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("legal: path %q resolved to a bool-typed property but its value was not a bool (%T)", path, val)
	}
	return b, nil
}

// resolveStackPath resolves path as an ImmutableStack against ctx, via
// ctx.ResolvePath.
func resolveStackPath(path string, ctx Context) (boardgame.ImmutableStack, error) {
	val, propType, err := ctx.ResolvePath(PropPath(path))
	if err != nil {
		return nil, err
	}
	if propType != boardgame.TypeStack {
		return nil, fmt.Errorf("legal: path %q is not a stack property", path)
	}
	if val == nil {
		return nil, nil
	}
	stack, ok := val.(boardgame.ImmutableStack)
	if !ok {
		return nil, fmt.Errorf("legal: path %q resolved to a stack-typed property but its value was not an ImmutableStack (%T)", path, val)
	}
	return stack, nil
}

// resolvePlayerIndexPath resolves path as a boardgame.PlayerIndex against
// ctx, via ctx.ResolvePath. Used by proposerIsCurrentPlayer to read
// move.TargetPlayerIndex (see catalog_players.go).
func resolvePlayerIndexPath(path string, ctx Context) (boardgame.PlayerIndex, error) {
	val, propType, err := ctx.ResolvePath(PropPath(path))
	if err != nil {
		return 0, err
	}
	if propType != boardgame.TypePlayerIndex {
		return 0, fmt.Errorf("legal: path %q is not a PlayerIndex property", path)
	}
	p, ok := val.(boardgame.PlayerIndex)
	if !ok {
		return 0, fmt.Errorf("legal: path %q resolved to a PlayerIndex-typed property but its value was not a PlayerIndex (%T)", path, val)
	}
	return p, nil
}

// resolveEnumPath resolves path as an enum.ImmutableVal against ctx, via
// ctx.ResolvePath.
func resolveEnumPath(path string, ctx Context) (enum.ImmutableVal, error) {
	val, propType, err := ctx.ResolvePath(PropPath(path))
	if err != nil {
		return nil, err
	}
	if propType != boardgame.TypeEnum {
		return nil, fmt.Errorf("legal: path %q is not an enum property", path)
	}
	if val == nil {
		return nil, nil
	}
	v, ok := val.(enum.ImmutableVal)
	if !ok {
		return nil, fmt.Errorf("legal: path %q resolved to an enum-typed property but its value was not an enum.ImmutableVal (%T)", path, val)
	}
	return v, nil
}

// ResolveInt, ResolveBool, ResolvePlayerIndex, ResolveEnum and ResolveStack
// are the typed forms of ctx.ResolvePath for a predicate's Evaluate. Each
// goes through ctx.ResolvePath, so the engine's read enforcement applies
// unchanged: the path must be one of the predicate's declared (or derived)
// Reads. Each returns an error if the property has a different type, and
// ResolveEnum/ResolveStack also return an error if the value is nil, so a
// successful call always yields a usable value. A typical Evaluate turns any
// error into legal.UnknownVerdict(err.Error()).

// ResolveInt resolves path as an int.
func ResolveInt(ctx Context, path PropPath) (int, error) {
	return resolveIntPath(string(path), ctx)
}

// ResolveBool resolves path as a bool.
func ResolveBool(ctx Context, path PropPath) (bool, error) {
	return resolveBoolPath(string(path), ctx)
}

// ResolvePlayerIndex resolves path as a boardgame.PlayerIndex.
func ResolvePlayerIndex(ctx Context, path PropPath) (boardgame.PlayerIndex, error) {
	return resolvePlayerIndexPath(string(path), ctx)
}

// ResolveEnum resolves path as a non-nil enum.ImmutableVal.
func ResolveEnum(ctx Context, path PropPath) (enum.ImmutableVal, error) {
	v, err := resolveEnumPath(string(path), ctx)
	if err == nil && v == nil {
		return nil, fmt.Errorf("legal: path %q resolved to a nil enum value", path)
	}
	return v, err
}

// ResolveStack resolves path as a non-nil boardgame.ImmutableStack.
func ResolveStack(ctx Context, path PropPath) (boardgame.ImmutableStack, error) {
	stack, err := resolveStackPath(string(path), ctx)
	if err == nil && stack == nil {
		return nil, fmt.Errorf("legal: path %q resolved to a nil stack", path)
	}
	return stack, err
}
