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

Concurrency note: unlike Game's existing cachedCurrentState/
cachedHistoricalMoves (game.go) — which are only ever touched from
Game.mainLoop's single goroutine — these two memos are also written from
move.Legal() evaluation reachable OFF mainLoop: server/api's
generateFormsWithLegality calls move.Legal() from an HTTP-handler goroutine,
concurrently with mainLoop's own Legal() evaluation during fixups. Two
goroutines reading/writing the same Go map concurrently is a fatal runtime
crash, not a benign data race, so both memos are guarded by Game.legalMemoMu
(game.go).

Lock-ordering rule: legalMemoMu is held ONLY around the map/field reads and
writes below — it must NEVER be held while evaluating a legal predicate
bucket or invoking a caller-supplied compute() closure, since both run
arbitrary user code (state field access, custom legal predicates, etc.) that
must not execute while holding an internal lock. Both memo functions below
therefore follow a "check under lock, compute unlocked, write under lock"
(double-checked) pattern rather than bracketing the whole
check-compute-store sequence in one critical section. This means two
goroutines racing on the same key can both miss the cache and both compute —
that's fine, since recomputing a legality verdict is idempotent and the last
write simply wins; no lock is held long enough for the map itself to be
observed in a torn state.
*/

// legalFieldIndepMemoKey is the field-independent legality memo's key
// (design spec §5's honest table): one entry per (move type, state version,
// proposer, predicate ordinal).
type legalFieldIndepMemoKey struct {
	moveName string
	version  int
	proposer PlayerIndex
	// predicate is the field-independent predicate's stable ordinal within
	// the plan. Caching each verdict separately lets evaluation walk the
	// original declaration order even when independent and dependent
	// predicates are interleaved.
	predicate int
}

// legalFieldIndepMemoGet returns the memoized field-independent verdict for
// key, if g's cache currently holds an entry for key.version. A cache
// currently keyed to any OTHER version is treated as empty (it is stale —
// see legalFieldIndepMemoSet's eviction rule).
//
// Guarded by g.legalMemoMu — see this file's top-of-file concurrency note.
func (g *Game) legalFieldIndepMemoGet(key legalFieldIndepMemoKey) (LegalVerdict, bool) {
	g.legalMemoMu.Lock()
	defer g.legalMemoMu.Unlock()

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
//
// Guarded by g.legalMemoMu — see this file's top-of-file concurrency note.
func (g *Game) legalFieldIndepMemoSet(key legalFieldIndepMemoKey, verdict LegalVerdict) {
	g.legalMemoMu.Lock()
	defer g.legalMemoMu.Unlock()

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
//
// Guarded by g.legalMemoMu — see this file's top-of-file concurrency note.
// compute is deliberately invoked OUTSIDE the lock (it is caller-supplied,
// effectively user code): two goroutines can race and both compute the same
// (upToVersion, targetPhase) tape, but that recomputation is idempotent, so
// whichever write lands last simply wins.
func (g *Game) LegalTapeMemo(upToVersion int, targetPhase enum.EnumKey, compute func() []*MoveStorageRecord) []*MoveStorageRecord {
	g.legalMemoMu.Lock()
	if g.legalTapeMemoValid && g.legalTapeMemoVersion == upToVersion && g.legalTapeMemoPhase == targetPhase {
		result := g.legalTapeMemo
		g.legalMemoMu.Unlock()
		return result
	}
	g.legalMemoMu.Unlock()

	result := compute()

	g.legalMemoMu.Lock()
	g.legalTapeMemo = result
	g.legalTapeMemoVersion = upToVersion
	g.legalTapeMemoPhase = targetPhase
	g.legalTapeMemoValid = true
	g.legalMemoMu.Unlock()

	return result
}
