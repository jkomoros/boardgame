# Status: the engine in a pattern, Memory on it, blackjack

> **Superseded in part.** Both patterns have since moved to labs'
> viewer-bound release: one view per viewer, the input witness restored, and
> no per-seat views. [`viewer-bound-port.md`](viewer-bound-port.md) has what
> changed and what was measured. The latency figures, the `TODO(cfc)` table
> and "Blackjack's seat release" below describe the pattern as it was.

Branch `claude/fabric-b-memory-pattern-70154b`, 2026-11-21, on top of
`fabric-factoring` at `b3c609ce`. Nothing pushed, no PR, `master` untouched,
no Go changed, nothing installed. Files touched are under
`server/static/src/engine/`, `fabric/patterns/` and `fabric/status/`.
`fabric/guest/` was built, not edited. Nothing was committed or left in the
labs checkout.

## Result

| Step | Outcome |
|---|---|
| 1. The engine in the sandbox | **Done.** The whole engine, rules layer, legal evaluator and all six games load and run in a pattern. The sandbox's refusals are deliberate design; the engine was rewritten to them in place. 1,168 of 1,168 unit tests, all 22 conformance documents |
| 2. Memory on the engine | **Done.** No rules are written in the pattern. Parity with Go on all six recordings, under Node and inside the sandbox. Two-user browser game replayed. **A move now takes a median of 3.7 s, against 1.8 s before** |
| 3. Blackjack | **Done as far as `cf test`.** Per-seat views from the engine's sanitization; parity on all seven recordings; multi-user test passes. Not playable in a browser until the guest draws blackjack and reads a view a seat |

## Which labs, which toolchains

- **Labs `4815052651`**, branch `fix/served-handler-write-identity` (draft PR
  commonfabric/labs#10111), through a detached worktree created and removed
  for this work. Every `cf check`, `cf test` and browser run below is on it.
- Deno 2.9.8. Node v24.22.3. Go was not run: no Go changed.

## Commits

| Commit | What |
|---|---|
| `65b8331a` Express the TypeScript engine without classes so a pattern can load it | Step 1: classes to factories, Error subclasses to named Errors, game specs to functions, the renderer into the engine, `sandbox.test.ts`, `tsconfig.sandbox.json`, the engine README |
| `b2534161` Engine: no statements or arithmetic at module scope; all six games load in a pattern | Step 1, found while loading blackjack and tictactoe |
| `1802afaf` Put Memory's table on the TypeScript engine | Step 2: the toolkit on the engine, Memory's `game.ts`, parity tests, table tests, screenshots |
| (the commit after it) Add blackjack on the same toolkit; document | Step 3, the READMEs, Memory's findings, this note |

## What ran

Baseline, before any change, on the merged branch: `npm ci`, `npm run
type-check` (clean), `npm run test:unit` **1,165 of 1,165** (one more than
the 1,164 the brief expected; nothing failed).

| What | Command | Result |
|---|---|---|
| Engine unit tests, after | `npm run test:unit` in `server/static` | 1,168 of 1,168 (three added: the sandbox shape, the scan's coverage, the renderer's equivalence) |
| Conformance corpus | part of the above | 22 of 22 documents |
| Type-check | `npm run type-check`; `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` | both clean |
| Each sandbox rule | `cf check` on 17 one-construct patterns | table in Memory's FINDINGS, "The engine in the sandbox" |
| Every game loads in a pattern | `cf check` on a pattern compiling each of the six | clean |
| The engine runs in a pattern | `cf test` on a lift that creates a game and computes a view | ran |
| Memory parity, Node | `node --test fabric/patterns/memory/parity.test.ts` | 15 of 15: 6 documents, 172 versions, 688 viewer outputs, 3,268 comparisons; each also through a key-reordering copy |
| Memory parity, in the sandbox | `cf test memory/parity.test.tsx` | 1 of 1: 6 versions, 24 viewer outputs, 0 problems |
| Memory's table | `node --test fabric/patterns/memory/table.test.ts` | 9 of 9 |
| Memory, multi-user | `cf test memory/multi-user.test.tsx` | 202 of 202, 166 s |
| Memory in a browser | `tools/play-two-users.mjs`, two identities | Three whole games on the engine (16 reveals each, timers, captures, one winner on both screens) |
| Blackjack parity, Node | `node --test fabric/patterns/blackjack/parity.test.ts` | 12 of 12: 7 documents, 419 versions, 3,118 viewer outputs, 13,729 comparisons |
| Blackjack secrecy, Node | same file | Over four whole games, every version, every viewer: more than 500 hole cards checked against every other viewer's view |
| Blackjack parity, in the sandbox | `cf test blackjack/parity.test.tsx` | 1 of 1: 16 versions, 64 viewer outputs, 0 problems |
| Blackjack, multi-user | `cf test blackjack/multi-user.test.tsx` | 92 of 92, 233 s: three players and a watcher, a whole game |
| Blackjack deploys | `cf piece new blackjack/main.tsx`; two CLI identities sat and dealt; a browser identity opened it | The watcher's view of the dealt game (181 KB) reached the guest, which showed "Watching" and reported it has no blackjack renderer |

**Only reasoned, or not done:**

- A seat's own view at the display boundary. The multi-user test reads each
  seat's view, but `cf test` has no display boundary; in the browser only
  the watcher's view was read. Seat views are released by the same rule.
- Blackjack in a browser, for the reasons under "What the guest needs" in
  its README.
- The rule as shipped on labs main without the fix. It has no witness, so
  it should not need the fix; not rerun.
- A missed delivery. With one input's versions held, a guest that misses a
  value of the view jumps instead of animating. Not provoked.
- The standalone guest's Playwright suite. `fabric/guest` was not edited.
- The whole README sequence from a clean checkout in one go.
- Pig, tictactoe, werewolf and debuganimations in a pattern beyond loading
  and compiling.

## Latency against the 1.8 s median

A proposal leaving the mover's guest to the new view reaching it, one game
of 16 reveals, nothing else running:

| | Median | Min to max |
|---|---|---|
| Before: hand-written rules, 3 KB a version | 1,760 ms | 1,083 to 2,173 |
| Engine, two inputs' versions held | 5,930 ms | 3,131 to 7,666 |
| **Engine, one input's versions held (shipped)** | **3,684 ms** | 2,595 to 6,144 |

The engine's own work is not the cost: under Node a reveal is 0.7 ms and a
view 16 ms. The state is: Go's Memory lists the 110 cards not in play in
every version, 19 KB against the 3 KB the hand-written rules kept, and the
runtime's time per move follows the size of what is stored and released.

## Runtime findings

New here; F1 to F6 are in `fabric/patterns/memory/FINDINGS.md` from before.

1. **The sandbox's refusal of classes is deliberate**, with a written reason
   (invocation isolation; SES spec section 4.2). The full list of what it
   refuses at module scope, each probed, is in Memory's FINDINGS. Not a gap.
2. **A cell position typed `unknown` reads back as `{}`, silently** (F7).
   The engine's JSON types say `unknown` throughout, so a table stores them
   under a recursive JSON type. A gap: the nested case should not be silent.
3. **A value read back from a cell has its object keys reordered.** Nothing
   in the engine depends on key order; the parity tests replay through such
   a copy to keep it so.
4. **The pattern compiler type-checks with `exactOptionalPropertyTypes`**,
   which the repository does not. Five places in the engine needed it.
5. **Cost follows size**, steeply: see "Latency".
6. **Release to a seat's occupants is not expressible** (specs#59 unlanded).
   Blackjack ships every seat's view to the table's space behind a marked
   `TODO(cfc)`.
7. **The served-handler fix on main works** for everything here that it
   touches. It answers open question 2 of `fabric-memory.md`. The witness
   itself stays out for the other reason (F2: a record its writer reads and
   sets keeps no stamp).

## The `TODO(cfc)` marks

| Where | Gap | Not held |
|---|---|---|
| `memory/table.tsx`, `blackjack/table.tsx` | Input witness lost on a record the writer reads and sets (F2; labs gap A) | A view is released over whatever is in the cell, not only what `move` wrote |
| `blackjack/table.tsx` | A release cannot name a list held in a cell (specs#59) | Every reader of the space can read every seat's view, hole cards included |

### Blackjack's seat release: what each version holds

A separate labs spike ran the designed per-seat release on a port of
labs#8608 (`SEAT-VIEWS-FINDINGS.md`, labs branch
`worktree-agent-ac9c0255af3654f11`). I read it and did not run it or port
it. `blackjack/table.tsx` now records that shape as the target in a comment.

| A seat's view is kept from | Shipped (runs on `fix/served-handler-write-identity`) | Designed (ran in the spike, on the port) |
|---|---|---|
| An unmodified client | yes: its guest asks only for its own seat | yes |
| A modified client at the table | no | yes: the display shows a seat's view only to that seat's list |
| A watcher with access to the space | no | yes |
| The table's owner | no | no: the owner can bind the list's writer from another module and list themselves |

Moving to the designed shape changes more than the rule: seating becomes a
second handler that owns the occupant lists (an entry `move` wrote lists
nobody), each seat gets a sealed key that `move` writes at setup, the number
of seats is fixed in the types, and the state has to be one scalar for the
witness to hold (F2). None of that was done here. No pattern-level guard
against the owner was added.

One `TODO(client)` in `engine/legal/message.ts`: the client's
`src/legal/renderLegalMessage.ts` should re-export the engine's renderer,
so there is one again.

## Decisions for the owner

1. **Speed or Go's exact state.** Memory is twice as slow on the engine
   because Go's Memory carries the unused deck in every version. Options:
   accept it; change the Go game (and its recordings) not to; or let a
   pattern's view omit a stack the renderer never draws, which breaks the
   "view is what Go sends" rule. Nothing was changed.
2. **One input's versions, or two.** One is 2.2 s a move faster. Two never
   skips an animation after a missed delivery.
3. **`private` is gone from the engine.** Every member of what was a class
   is now visible on its interface. If that matters, the members can be
   closure variables instead, at the cost of rewriting each method body.
4. **Who owns the legal-message renderer.** The engine now has its own copy,
   because it cannot import the client's. The client session should make
   theirs a re-export.
5. **Blackjack's release.** It ships with every seat's view readable by the
   whole table. The designed release exists as a spike on a port of
   labs#8608 and protects a seat from everyone but the table's owner. If
   neither is acceptable even for a demo, blackjack waits on specs#59 and on
   a ruling that binds a declared writer to its run.
6. **What the guest protocol should be for a view a seat.** Blackjack's
   README proposes one (`view` plus `seat0` to `seat6`); the client session
   owns the guest and has not seen it.

## For the client session

- `fabric/patterns/toolkit/view.ts` kept its type parameters, so the guest's
  bridge still type-checks against it. `TableInfo.chest` now carries
  `Constants` and the engine's real `Enums`, and `TableInfo` gained
  `moveInputSchemaFingerprint`, which the bridge could pass in `init`.
- The script's `motion` counters read `{}` since the animation test hook
  became opt-in.
- Blackjack's README, "What the guest needs".
