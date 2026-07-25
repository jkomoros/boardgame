# Evidence pack: error dispatch recursion and post-dismissal error loss

Two defects at the same choke point — `boardgame-app`'s handling of the
`show-error` event and of the error dialog's lifecycle. Both pre-existing.

## Defect 1: re-entrant dispatch storm ("Maximum call stack size exceeded")

**Mechanism.** Every component that watches `state.game.error` dispatches a
`show-error` DOM event from inside its `stateChanged`. There are many:
measured 13-23 live `boardgame-configure-game-properties` instances on a game
page (one per listed game), plus `boardgame-player-roster` and
`boardgame-move-form`. `_handleShowError` answered each one with a
SYNCHRONOUS `store.dispatch` — a dispatch from inside a Redux subscriber
notification — and `updateAndShowError` is a thunk firing two more actions,
so each watcher pushed two more full notification passes onto the stack while
the first was still unwinding.

**Measured, one SUBMIT_MOVE_FAILURE:**

| | before | after |
|---|---|---|
| max dispatch depth | 16-17 | 1 |
| subscriber notifications | 31-33 | 9 |

Depth grew linearly with watcher count, which is why busy pages overflowed
the stack. Each watcher's own `_lastError` guard bounds it to one dispatch
per distinct error — that is why it was deep-but-finite, not truly infinite —
but it does nothing about the cross-instance amplification.

**Fix.** Defer the dispatch one microtask (the current notification finishes
first, so depth stays flat), plus dedupe identical payloads while a dialog
with the same content is on screen. The dedupe compares the RAW payload, not
stored state: `updateError` normalizes (blanks `message` when it equals
`friendlyMessage`), so state never compares equal to what the watcher sent —
a first attempt comparing against state silently never matched.

## Defect 2: after any dismissal, the next error never displayed

**Mechanism.** State was synced from md-dialog's `closed` event, which fires
only after the close ANIMATION. An error arriving in that window set
`showing=true` and was then clobbered back to false by the late handler.
Measured: the second error's `show-error` events all fired (35 of them) and
`state.game.error` updated, but `state.error.showing` landed false.

Event taxonomy, measured directly on the live dialog:
- dismissal intent → `cancel` (esc path) → `close` → ... → `closed` (post-animation, observed firing twice)
- `close` carries `showing=true` (state not yet synced); `closed` arrives long after.

**Fix.** Bind to `close` instead of `closed` — sync at intent time, leaving no
window in which a new error can be lost. Secondary, also measured: on the esc
path state previously caught up only at `closed`, so for ~100ms the `?open`
binding still read true and the dialog popped back open mid-dismissal.

## Verification

`tests/basic/error-dispatch.spec.ts` — both tests watched fail-first against
the unfixed component (depth 19 vs the asserted ≤3; the different-error case
red with `@closed`, green with `@close`) and pass after. A third candidate
test asserting escape-dismissal behaviour was DROPPED rather than committed:
it passed with and without the fix, so it pinned nothing.
