# Guest message protocol

Version 1. Types: [`protocol.ts`](protocol.ts), which imports nothing.

The **guest** is one HTML document running in a sandboxed iframe. The
**host** is whatever embedded it: a Common Fabric pattern through
`cf-iframe`, or [`harness.html`](harness.html). They exchange plain JSON with
`postMessage`. The guest has no server, no storage and no network.

## Envelope

Every message is an object with three fixed fields and a `type`:

```json
{ "channel": "boardgame-guest", "v": 1, "type": "..." }
```

- The guest accepts a message only if it came from the window that embeds it
  (`event.source === window.parent`), carries this channel and version, and
  has one of the three host types below. Anything else is ignored silently.
- The guest posts to its parent with target origin `*`. A sandboxed guest has
  an opaque origin and cannot name its parent's. What it posts is what the
  person did, never game state.

## Host to guest

### `init`

Static facts about the table. Send it once the guest has posted
`ready` with `stage: "loaded"`.

| Field | Meaning |
|---|---|
| `gameName` | Which bundled renderer to mount. The guest reports an `error` if it has none by that name |
| `chest` | Decks, enums, constants and legal-message templates, as the Go manager marshals its chest |
| `viewer` | The seat the views will be computed for. `-1` is the observer |
| `players` | One `{ playerIndex, label, color? }` per seat, in seat order |
| `moveInputSchemaFingerprint` | Optional. The engine's fingerprint of its move inputs. Without it the guest uses the fingerprint compiled into its renderer |
| `gameId` | Optional. Identifies the table. The guest passes it to the renderer host, where it seeds effects and keys transitions so two tables do not share them. Absent is the same as empty |

A second `init` resets the guest completely.

### `state`

A **window** of state versions for one viewer.

| Field | Meaning |
|---|---|
| `viewer` | The seat the entries were computed for |
| `epoch` | Optional, default 0. Changes when history stops being comparable: a new game on the same table, a rewind |
| `entries` | One or more state entries. Order does not matter; the guest sorts by version |

Each entry:

| Field | Meaning |
|---|---|
| `version` | The state version, a non-negative integer |
| `state` | The viewer's sanitized state, unexpanded: what `Game.JSONForPlayer(...).CurrentState` marshals |
| `move` | The sanitized record of the move that produced this version, `{ AnimationKey, Version, Properties? }`, or `null` |
| `currentPlayerIndex` | Whose turn it is at this version |
| `timers` | Optional. `{ [timerId]: { TimeLeft } }`, milliseconds left when the entry was produced |
| `outcome` | `{ finished, winners }` as of this version |
| `moveLegality` | Optional. `{ [moveName]: { legalForPlayer, legalForAnyone, error? } }` for default-bound moves. Absent means nothing is legal here, which is right for a version a fix-up replaced |
| `zeroInputMoves` | Optional. Names of moves that take no argument |
| `projectedMoveChoices` | Optional. Per-candidate legality for single-argument moves, in the Go API's wire shape. An absent `Sets` is read as empty |

**How the guest plays a window.** The host is expected to deliver the latest
value of a cell, not an ordered log, so windows overlap and repeat.

1. Entries are grouped under a key, `epoch` plus `viewer`.
2. Under a key the guest remembers the highest version it has queued. An
   entry at or below it is ignored. This makes re-sending any window safe.
3. Every other entry is queued in version order and shown one at a time. The
   next one is installed only when the animation gate reports the current
   one settled.
4. **The first window under a key contributes only its newest entry**, shown
   without animation. Nothing has been shown yet, so the older entries are
   history this viewer never saw. A guest that loads mid-game, or switches
   seat, cuts to the present.
5. A different key (a new `viewer` or `epoch`) discards the queue and what is
   on screen, then applies rule 4.
6. If a version arrives that is not the successor of the last one queued, the
   guest still plays it, animating straight across the gap, and says so with
   `jumpedFrom` on that version's `animations-done`. The host's window was
   too short for how fast the game moved.

**Sizing the window.** It must cover every version produced between two
deliveries the guest receives. One player move produces one version plus its
fix-ups; Memory's longest chain is three (reveal, capture, finish turn).

### `move-result`

Optional. The host's verdict on one `propose-move`.

| Field | Meaning |
|---|---|
| `requestId` | From the `propose-move` |
| `ok` | Whether the move was applied |
| `error` | Optional text for a refusal |
| `actualVersion` | Optional. The head version, when the refusal was for staleness |

A host need not send it. A proposal is settled by whichever comes first:

- a `state` entry newer than the proposal's `snapshotVersion`, which counts as
  success;
- a `move-result`;
- ten seconds of silence, after which the control reports a failure and the
  person may try again.

Until then the renderer holds that one move as pending and proposes no other.

## Guest to host

### `ready`

`{ stage, games }`. Sent with `stage: "loaded"` when the guest is listening,
and with `stage: "mounted"` after an `init` has been applied and the renderer
exists. `games` lists the renderers in this bundle. `state` messages sent
before `mounted` are queued, not lost.

### `animations-done`

One per version, in the order shown.

| Field | Meaning |
|---|---|
| `version`, `viewer`, `epoch` | Which version settled |
| `how` | `played`: installed and animated. `initial`: the first version under a key, shown without animation. `skipped`: the renderer's `animationLength` hook asked for an intermediate version not to be shown |
| `remaining` | Versions still queued behind it |
| `jumpedFrom` | Present when the previous version shown was not `version - 1` |

`initial` is posted once the first render has painted. It does not wait for
entrance animations the renderer may run on first render (FINDINGS F17).

A host does not have to act on this message. It exists so a host can pace
deliveries, and so a test can see that every version was played.

### `propose-move`

The person made a move.

| Field | Meaning |
|---|---|
| `requestId` | Unique per proposal |
| `name` | The move's name |
| `arguments` | `{ [field]: string }`, serialized as the Go API's move forms take them |
| `snapshotVersion` | The version on screen when the gesture happened |
| `viewer` | The seat the guest was showing |

The guest does not send a proposal while a newer version is queued or
animating. `viewer` is for the host's bookkeeping only: **the host must take
the proposer from its own runtime, never from this field.**

### `tick`

`{ timerId, version }`. A timer the guest was shown has reached zero on the
guest's clock, counted from when its entry arrived. It is a request to check
deadlines, not a claim that one has passed. Sent once per timer.

### `error`

`{ message }`. Diagnostics: an unknown game, an uncaught exception, a
Content-Security-Policy violation, or a renderer that failed to mount.
Nothing depends on the host reading it.

## Inside a `cf-iframe`

A Common Fabric pattern does not post these messages. It hands the guest
named cells and streams, and `src/fabric-bridge.ts` speaks this protocol to
the renderer host on the pattern's behalf, in the same document.

### What the pattern hands over

The types are in `fabric/patterns/toolkit/view.ts`.

| Name | Kind | Holds |
|---|---|---|
| `view` | cell, read-only | A `TableView` computed for whoever is looking: the viewer's seat, each seat's name, the phase, and the versions the last input produced, sanitized for the viewer's seat, or for a watcher if they hold none |
| `info` | cell, read-only | A `TableInfo`: the game's name and display name, its chest, the engine's `moveInputSchemaFingerprint`, seat colours, and `setupOptions` for the lobby |
| `move` | stream | A `Proposal`. No proposer is sent |

### One view, computed for the viewer

There is one `view` whatever the game. The pattern's released lift asks the
runtime which principal it is computed for and finds that principal's seat in
the sealed table, so the runtime holds one instance of the view for each
principal who looks, and each guest is handed its own
(`fabric/patterns/blackjack/table.tsx`). A seated player's instance holds
their own hole card and nobody else's; a watcher's holds nobody's.

The guest does not choose what it is shown and has nothing to ask for but
`view`. The view's `viewer` is the seat it was computed for, or `-1`; the
guest never learns its own principal or anyone else's, and a seat is named
only by the label its occupant chose. The head entry's `viewers` holds one
part, the viewer's own; a version that is not settled holds none.

What keeps one principal's instance from another is the pattern's release
rule, at the runtime's display boundary.
`fabric/patterns/blackjack/README.md` says what that rule holds.

### How the protocol's messages are produced

| Protocol | In a `cf-iframe` |
|---|---|
| `init` | Built from `info` (game name, chest, seat colours, `moveInputSchemaFingerprint`) and the seats the view names. Sent again when the deal (`epoch`), the viewer's seat or the number of seats changes |
| `state` | Every new value of the view: its `entries` are the window, its `epoch` the epoch. `viewer` is the viewer's seat, or `-1` |
| `propose-move` | An event on the `move` stream, `{ kind: "play", name, arguments }`. No proposer is sent |
| `move-result` | The pattern sends none. After `send()` resolves the bridge reads the view again, and reports a failure if the table did not advance |
| `tick` | An event on the `move` stream, `{ kind: "tick", timerId }`, repeated a few times while that version is still the head, since the table's clock may be behind the guest's |
| `ready`, `animations-done`, `error` | Not forwarded; `error` is shown in the guest's status line |

## Not in version 1

- **Table actions** (lock room, skip turn, switch to solo). They belong to the
  companion Table surface, which this guest does not mount.
- **Companion surfaces and synchronized timing.** The guest always mounts the
  solo renderer and never supplies a version animation context.
- **Admin perspective.** The guest proposes as the viewer it was given.
- **Chat, gathering, seating.** The pattern's own runtime provides these.
- **Art and fonts.** No asset crosses the protocol yet. A game that needs art
  will need an `assets` field of `data:` URLs on `init` (FINDINGS F9).
