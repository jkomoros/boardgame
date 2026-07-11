package legal

import (
	"fmt"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

/*
This file resolves the catalog's own "kind.Property" paths (game.X,
player.X, move.X) against a Context, using typed PropertyReader accessors.

It is a SEPARATE, smaller implementation from core's path resolver
(resolveLegalPath in legal_path.go, package boardgame): that resolver, and
the boot-time validator it's paired with (validateLegalPath), are both
unexported and therefore unreachable from this package. This is not an
oversight: the design spec's own illustrative catalog code (§8,
RevealableCardAt's intField/stackAt helpers) shows purpose-built catalog
predicates resolving paths themselves against the exported Context surface,
not sharing core's internal resolver. Every built-in and future
game-registered predicate in this package follows that same pattern.

"players[*].X" is quantifier-only per the design spec and is not resolved
here; it is out of scope for the leaf catalog predicates built in this
package so far.
*/

// legalCatalogPathKind identifies which reader a "kind.Property" path
// resolves against.
type legalCatalogPathKind int

const (
	catalogPathGame legalCatalogPathKind = iota
	catalogPathPlayer
	catalogPathMove
)

// parseCatalogPath splits path into its kind and property name. It only
// checks the path's shape, matching the KIND segment literally against
// "game", "player", or "move" (case-sensitive) — the same three (of the
// design spec's four) kinds this package's leaf predicates support.
func parseCatalogPath(path string) (legalCatalogPathKind, string, error) {
	kindStr, prop, ok := strings.Cut(path, ".")
	if !ok || prop == "" {
		return 0, "", fmt.Errorf("legal: invalid path %q: expected KIND.Property (KIND is one of game, player, move)", path)
	}
	switch kindStr {
	case "game":
		return catalogPathGame, prop, nil
	case "player":
		return catalogPathPlayer, prop, nil
	case "move":
		return catalogPathMove, prop, nil
	default:
		return 0, "", fmt.Errorf("legal: invalid path %q: unsupported path kind %q (catalog predicates support game, player, move)", path, kindStr)
	}
}

// readerForCatalogPath returns the PropertyReader that kind resolves
// against within ctx. It returns an error (never panics) whenever the
// context is insufficient to resolve — nil ctx.State, no valid current
// player for a player.* path, or nil ctx.Move for a move.* path. Every
// caller in this package turns that error into an UnknownVerdict rather
// than propagating it, which is how catalog predicates satisfy the design
// spec's runtime guard (§1: an undeclared/unsatisfiable move.* read
// degrades to Unknown, never a panic) without relying on any recover()
// machinery of their own.
func readerForCatalogPath(kind legalCatalogPathKind, ctx Context) (boardgame.PropertyReader, error) {
	switch kind {
	case catalogPathGame:
		if ctx.State == nil {
			return nil, fmt.Errorf("legal: state is nil")
		}
		return ctx.State.ImmutableGameState().Reader(), nil
	case catalogPathPlayer:
		if ctx.State == nil {
			return nil, fmt.Errorf("legal: state is nil")
		}
		current := ctx.State.ImmutableCurrentPlayer()
		if current == nil {
			return nil, fmt.Errorf("legal: no valid current player for player.* path")
		}
		return current.Reader(), nil
	case catalogPathMove:
		if ctx.Move == nil {
			return nil, fmt.Errorf("legal: move.* path but no move was provided")
		}
		return ctx.Move.Reader(), nil
	}
	return nil, fmt.Errorf("legal: unknown path kind")
}

// resolveIntPath resolves path (a "kind.Property" string) as an int against
// ctx.
func resolveIntPath(path string, ctx Context) (int, error) {
	kind, prop, err := parseCatalogPath(path)
	if err != nil {
		return 0, err
	}
	reader, err := readerForCatalogPath(kind, ctx)
	if err != nil {
		return 0, err
	}
	return reader.IntProp(prop)
}

// resolveBoolPath resolves path as a bool against ctx.
func resolveBoolPath(path string, ctx Context) (bool, error) {
	kind, prop, err := parseCatalogPath(path)
	if err != nil {
		return false, err
	}
	reader, err := readerForCatalogPath(kind, ctx)
	if err != nil {
		return false, err
	}
	return reader.BoolProp(prop)
}

// resolveStackPath resolves path as an ImmutableStack against ctx.
func resolveStackPath(path string, ctx Context) (boardgame.ImmutableStack, error) {
	kind, prop, err := parseCatalogPath(path)
	if err != nil {
		return nil, err
	}
	reader, err := readerForCatalogPath(kind, ctx)
	if err != nil {
		return nil, err
	}
	return reader.ImmutableStackProp(prop)
}

// resolveEnumPath resolves path as an enum.ImmutableVal against ctx.
func resolveEnumPath(path string, ctx Context) (enum.ImmutableVal, error) {
	kind, prop, err := parseCatalogPath(path)
	if err != nil {
		return nil, err
	}
	reader, err := readerForCatalogPath(kind, ctx)
	if err != nil {
		return nil, err
	}
	return reader.ImmutableEnumProp(prop)
}
