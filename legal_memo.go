package boardgame

import "github.com/jkomoros/boardgame/enum"

/*
This file implements design spec §5's two per-GAME-INSTANCE engine-win
memos: the field-independent legality memo (keyed move type / state version
/ proposer) and the move-tape memo (keyed state version / phase) that
retires moves/default.go's historicalMovesSincePhaseTransition TODO
("ideally we'd memoize this so all base moves for this game for this
version could use the result... we'll want to make sure the lifetime of the
cache does not extend beyond the lifetime of the game, or is purged every so
often").

Both memos live on *Game, not *GameManager: a GameManager is shared by every
game of a type, but a Game is one running instance with one version
sequence, so anchoring the cache to the Game means it is garbage-collected
along with the game automatically, and "bound memory: keep only the current
head version per game" (the honest table, design spec §5) is enforced by a
simple rule both memos share — a lookup/store for a DIFFERENT version than
whatever is currently cached discards the entire prior cache first, so at
most one version's worth of entries is ever resident. No separate purge
timer is needed.

Concurrency note: like Game's existing cachedCurrentState/
cachedHistoricalMoves (game.go), these caches are populated without a mutex,
matching that existing convention. Game.mainLoop's single goroutine is the
only steady-state writer; a concurrent reader from outside mainLoop (e.g.
server/api's per-request player+admin Legal() double pass hitting a
modifiable game's shared *Game) is a pre-existing risk profile shared by
every other Game-instance cache, not one this file introduces.
*/

// legalFieldIndepMemoKey is the field-independent legality memo's key
// (design spec §5's honest table): one entry per (move type, state version,
// proposer).
type legalFieldIndepMemoKey struct {
	moveName string
	version  int
	proposer PlayerIndex
}

// legalFieldIndepMemoGet returns the memoized field-independent verdict for
// key, if g's cache currently holds an entry for key.version. A cache
// currently keyed to any OTHER version is treated as empty (it is stale —
// see legalFieldIndepMemoSet's eviction rule).
func (g *Game) legalFieldIndepMemoGet(key legalFieldIndepMemoKey) (LegalVerdict, bool) {
	if g.legalFieldIndepMemo == nil || g.legalFieldIndepMemoVersion != key.version {
		return LegalVerdict{}, false
	}
	v, ok := g.legalFieldIndepMemo[key]
	return v, ok
}

// legalFieldIndepMemoSet stores verdict for key. If g's cache is currently
// keyed to a different version than key.version, the entire cache is
// discarded first and rekeyed to key.version — the "bound memory: keep only
// the current head version per game" eviction rule (design spec §5).
func (g *Game) legalFieldIndepMemoSet(key legalFieldIndepMemoKey, verdict LegalVerdict) {
	if g.legalFieldIndepMemo == nil || g.legalFieldIndepMemoVersion != key.version {
		g.legalFieldIndepMemo = make(map[legalFieldIndepMemoKey]LegalVerdict)
		g.legalFieldIndepMemoVersion = key.version
	}
	g.legalFieldIndepMemo[key] = verdict
}

// LegalTapeMemo returns the memoized move-tape for (g, upToVersion,
// targetPhase) — design spec §5's "Tape memoization" engine win, retiring
// moves/default.go's historicalMovesSincePhaseTransition TODO. If g's cache
// doesn't already hold exactly this (upToVersion, targetPhase) pair, compute
// is invoked once and its result is cached — evicting whatever was cached
// before it, so at most one (version, phase) pair is ever resident, the
// same bounded-memory rule as the field-independent memo above — before
// being returned.
//
// Both moves.Default's frozen legalMoveInProgression chain and the
// "inProgression" declarative predicate (moves/catalog_framework.go) reach
// this through the exact same legalMoveInProgression method call (see that
// method's doc comment), so they observably share one tape walk per (game,
// version) by construction, not by convention — see the Task 9 report.
func (g *Game) LegalTapeMemo(upToVersion int, targetPhase enum.EnumKey, compute func() []*MoveStorageRecord) []*MoveStorageRecord {
	if g.legalTapeMemoValid && g.legalTapeMemoVersion == upToVersion && g.legalTapeMemoPhase == targetPhase {
		return g.legalTapeMemo
	}

	result := compute()

	g.legalTapeMemo = result
	g.legalTapeMemoVersion = upToVersion
	g.legalTapeMemoPhase = targetPhase
	g.legalTapeMemoValid = true

	return result
}
