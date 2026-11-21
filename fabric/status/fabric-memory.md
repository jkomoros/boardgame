# Memory on Common Fabric: status

Stream B, steps 1 and 2 joined, of [`../PROPOSAL.md`](../PROPOSAL.md): Memory
as a Common Fabric pattern, drawn by the boardgame guest in a `cf-iframe`.
Branch `claude/fabric-b-memory-pattern-70154b`. Nothing pushed, no PR,
`master` untouched. Everything is under `fabric/patterns/`, `fabric/guest/`
and this file; `server/static/src`, `examples/` and all Go are unmodified,
and nothing was committed or left behind in the labs checkout.

Start with [`../patterns/memory/README.md`](../patterns/memory/README.md).
What the runtime refused is in
[`../patterns/memory/FINDINGS.md`](../patterns/memory/FINDINGS.md).

## Outcome

- **A two-user game was played in a browser.** Two Chromium contexts, two
  freshly registered identities, a complete game through the deployed piece
  on labs' local servers with server execution on: 14 reveals, two
  mismatches hidden by the timer, five pairs captured, the same winner on
  both screens. The cards flip and fly through the real animator.
  Screenshots are in `fabric/patterns/memory/screenshots/`.
- **The pattern is the idiomatic shape**: one space, the state as a sealed
  record, one `move` handler taking the proposer from the runtime, one
  blessed lift returning a structured view in the client's wire shape,
  released to the table's space in two steps.
- **The release rule is one condition short of designed**, in one marked
  place.

## Which labs, which toolchains

- Labs `6e12b01bf7`, the tip of branch `worktree-agent-ad697e53f88bab905`,
  which includes `5b1c700b4a` (served handlers attributed to the handler).
  The final game, screenshots, latency and the last test runs are from it.
  The findings were first met on labs main `8c6f09c5fc`; FINDINGS says which
  result is from which. The design depends on that fix.
- Deno 2.9.8. Node v24.22.3.
- Go: `go-1.27` (`3a00e697`) is merged into this branch. This work changed
  no Go. After the merge, `./scripts/go-local vet ./fabric/guest/tools/...`
  passed on the machine's default Go 1.27.4 with no `GOTOOLCHAIN`.

## Interim seams

One, in `fabric/patterns/memory/table.tsx` under "The policy, and the interim
seam". It holds the missing condition in a comment.

| Drops | Because | Consequently not held today |
|---|---|---|
| The input witness (`inputWitness` naming `move`) | FINDINGS F2: a record its writer reads and then sets keeps no witness. Confirmed on the fix branch: record state with the witness releases nothing; one JSON string with it plays a whole game | The release is of whatever is in the table's cell, not only of what `move` wrote |

Not seams: the reader's role is back, as the two-step `Space` release
(FINDINGS F3). And anyone who can read the space can read the sealed table
today (specs#95, observed on main).

## Commits

| Commit | What |
|---|---|
| `1a5bc657` Add Memory as a Common Fabric pattern drawn by the boardgame guest | `fabric/patterns/toolkit` (table, wire types), `fabric/patterns/memory` (rules, sealed table, piece, both tests, the two-user script), the guest's Fabric entry point and bridge, the fix to the guest's mount |
| `54e6b59e` Document the Memory pattern | `FINDINGS.md`, both READMEs, screenshots, guest `README.md` and `PROTOCOL.md`, the script's measurements, this note |
| (merge) `go-1.27` into this branch | The Go 1.27 fix, from the orchestrator |
| (the commit after it) Release to the table's space; narrow the seam to the witness | The two-step rule in `table.tsx`, rerun on the labs fix branch, new screenshots, FINDINGS and this note revised |

## What was verified, and how

"Ran" means the command was run and its result read. "Browser" means
Chromium against the local toolshed.

| Claim | How | Evidence |
|---|---|---|
| A full two-player game through `move` | Ran: `cf test multi-user.test.tsx` | 202 passed, 0 failed |
| No view shows a hidden card's face | Ran: same test, every step, every entry of the window; `rules.test.ts` | |
| Wrong-player, forged-proposer, unseated and illegal moves are refused | Ran: same test (revision unchanged after each); `rules.test.ts` | |
| Two principals get consistent views | Ran: same test. Each player predicts the other's next card from their own view and notes, then checks it | Both players' `agrees` assertions, at least five comparisons each |
| The timer: a tick before the deadline does nothing, one after hides the cards and passes the turn | Ran: `node --test rules.test.ts` with a moved clock | 9 of 9 |
| The same in a browser, against the server's clock | Browser: the script waits out each mismatch without clicking | Hidden 4.4 to 6.9 s after the second card showed |
| The pattern deploys from this repository with `cf --root` | Ran | `cf piece new`, about 2 s |
| The guest loads in `cf-iframe` and mounts the real `boardgame-render-game` | Browser | Screenshots; the script finds `boardgame-render-game` and its cards in the guest frame |
| Two identities play a whole game | Browser: `tools/play-two-users.mjs`, run to completion three times on main and three on the fix branch | Exit 0; "… won." on both |
| Cards animate | Browser: the client's own counters in each guest | 29 cycles, 238 animations, 0 watchdog firings in the final game; up to 10 running animations on a move |
| The proposer is the runtime's | Browser and cf test | The view's seats hold each browser's DID, which the script never sends; a payload naming someone else is ignored |
| F1 to F5 of FINDINGS | Browser, with variants of the pattern that are not kept | Described there |
| The standalone guest still works | Ran: `npm run type-check`, `npm test` in `fabric/guest` | 21 of 21 |
| What the sandbox accepts of the engine | Ran: `cf check` on probes | FINDINGS, "The engine in the sandbox" |
| A `READ` key reads the sealed table; a `WRITE` key cannot forge it through `cf` | Ran: `cf cell get --input`, `cf cell set --input` | FINDINGS, "What a table protects" |

**Only reasoned, or not done:**

- Whether the client ran `move` speculatively. Nothing visible suggested it;
  which of two explanations holds was not determined.
- A modified client. Every "refused" above is by a stock client or the
  server.
- More than two players, the 20- and 40-card boards, a second game on a
  table (`restart`) in a browser. `restart` is covered in `rules.test.ts`.
- Leaving a seat and a watcher in a browser. Both are in the tests.
- Any browser but headless Chromium. The in-app browser pane was not used.
- The README's commands were each run, but not once through from a clean
  checkout in one go.
- Go parity. The rules mirror the Go game's shape, move names and fix-up
  order, and use the engine's `componentId`, but not its shuffle or its
  unused-card stack; `rules.ts` marks both `TODO(engine)`.

## The security table

From `fabric/RESEARCH.md`, "What a table protects, and from whom", with what
was observed. The CLI rows were run on labs main `8c6f09c5fc`; the rest on
both.

| Property | RESEARCH.md says of labs today | Observed |
|---|---|---|
| A player cannot see another seat's cards | Holds only against an unmodified client | As said. The view never holds a hidden face (tested). A key with only `READ` read the sealed table's salt and layout with `cf cell get --input` |
| A player cannot make a move as someone else | Same as designed, per the protocol document | Holds in every case tried: forged payload fields ignored, seats bound to each browser's own DID, an unseated key's proposals changed nothing |
| A player cannot make an illegal move | Same, for the authoritative run | Holds in every case tried |
| A player cannot forge the state or a view | A modified client with write access can author into those documents | Through the stock `cf`, a write to the table is refused (`writeAuthorizedBy requires a trusted verified binding identity at /table`) and a write to the view failed and changed nothing. A modified client was not tried |
| The table owner cannot see hands | Open | Not examined; the owner is a reader, so the first row applies |
| (added) The view is released only over a state `move` wrote | | Not held: the interim seam |
| (added) The view is released only to the space's readers | | The view carries `Space(table)` and the display admits that space's readers. Nobody outside the space could load the piece to test the label alone |

## Open questions

1. **Is F2 a ruling to ask for or a constraint to design around?** The
   specification says a destination its writer read is not stamped. A sealed
   state machine reads its state by definition. Either the witness rule
   grows a case for "read only what this same writer wrote", or every such
   pattern stores one opaque scalar.
2. **When does the served-handler fix land on labs main?** `5b1c700b4a` works
   here, at the display boundary in a browser. A second labs branch,
   `fix/served-handler-write-identity`, was seen and not run.
3. **What happens to tables in flight when F2 closes?** Restoring the witness in
   `table.tsx` changes the policy's digest, which strands any table in
   flight. Fine for a demo; the migration question in RESEARCH.md stands.
4. **Latency.** Over two seconds a move, growing with the view. Splitting the
   view by version is untried, and F4 says each piece of a split view must
   be newly built.
5. **Classes in the engine.** The sandbox refuses a class at module scope, so
   none of the engine's stacks, state, sanitization or generator can be
   imported. Either the engine grows a class-free surface or labs allows
   classes. This blocks step 3 of stream B.
6. **How does a guest learn who is looking?** `identify` works but is a
   second handler and an event sent on load. A read of the viewer's own
   principal from the guest's bridge would remove both.
7. **Where should the built guest live?** `fabric/patterns/guest/guest.js` is
   1.08 MB and ignored by Git, so a checkout cannot deploy without Node and a
   labs checkout to bundle the client from. Labs checks its guests' bundles
   in.

## Blockers

None that needs the owner. No account, login or system setting was needed:
the identities are local passphrase identities on a local server.
