# Boardgame on Common Fabric: handoff

Written 2026-11-21 by the orchestrating session. Branch `fabric-factoring`,
local only. Nothing in this repository was pushed, no pull request was opened,
and `master` is untouched. The plan this work follows is
[PROPOSAL.md](PROPOSAL.md); the research behind it is [RESEARCH.md](RESEARCH.md).

Every number below comes from a status file in [`status/`](status/), written by
the session that ran the command. The final combined run is in
[`status/integration.md`](status/integration.md); read that for what passes on
the head commit.

## What exists now

| Piece | Where | State |
|---|---|---|
| Client layers and a typed `RendererHost` | `server/static/src`, [status/client-layers.md](status/client-layers.md) | `boardgame-render-game` imports nothing from the webapp. A layer test forbids upward imports; its allow-list is empty |
| Two packages | `server/static/packages/elements`, `.../renderer-kit`, [status/client-packages.md](status/client-packages.md) | Layers 1 and 2 with explicit `exports`. The webapp reaches them by name. Old import paths used by game clients, the generator and tests still work |
| `cf-iframe` guest | `fabric/guest`, [status/guest-harness.md](status/guest-harness.md) | One file, about 0.6 MB of the 4 MiB limit. Depends only on the two packages. Draws Memory and blackjack with the unmodified game renderers |
| TypeScript engine | `server/static/src/engine`, [status/engine-core.md](status/engine-core.md), [engine-legal.md](status/engine-legal.md), [engine-rules.md](status/engine-rules.md) | Core, legal framework, rules layer, six ported games. Matches Go on all 22 conformance documents at every version for every viewer. Written without classes so a pattern can load it |
| Go conformance exporter | `boardgame-util/lib/conformance`, `conformance/`, [status/conformance-export.md](status/conformance-export.md) | 22 documents with a staleness check. The only Go added; no existing Go behavior changed |
| Memory on Fabric | `fabric/patterns/memory`, [status/viewer-bound-port.md](status/viewer-bound-port.md) | Sealed record, one `move` handler, one `viewFor` lift released to the principal it was computed for. Two browser identities played whole games with server execution on. A move takes a median 1.5 s |
| Blackjack on Fabric | `fabric/patterns/blackjack`, [status/viewer-bound-port.md](status/viewer-bound-port.md) | The same pattern as Memory over a different game; parity with Go on seven recordings. Three browser identities and the table's owner as a watcher played whole games; the owner is shown no hole card. Deal 1.2 s, a move a median 2.6 s, a round's end 9.6 s. Needs the labs branch with `viewerBoundRelease` on |
| Board in pattern JSX, no iframe | `fabric/patterns-declarative`, [status/declarative-pattern.md](status/declarative-pattern.md) | Two identities played Memory through a pattern whose JSX is boardgame elements, on a local labs spike branch |
| Go 1.27 | commit `3a00e697` | `golang.org/x/tools` raised so the tree builds and tests under Go 1.27.4 |

## What lives in labs

| Branch (in `commonfabric/labs`) | State | What |
|---|---|---|
| `fix/served-handler-write-identity` | Pushed, draft PR commonfabric/labs#10111, CI green | Served handlers are attributed to the handler. Every pattern here depends on it for a witnessed release |
| `worktree-agent-ad697e53f88bab905` | Local | First spike: a sealed state machine. Findings copied to [labs-spikes/SEALED-STATE-MACHINE-FINDINGS.md](labs-spikes/SEALED-STATE-MACHINE-FINDINGS.md) |
| `worktree-agent-ac9c0255af3654f11` | Local | Port of the list-release draft (labs#8608) and a card table with per-seat views. Findings copied to [labs-spikes/SEAT-VIEWS-FINDINGS.md](labs-spikes/SEAT-VIEWS-FINDINGS.md), which ends with an unposted draft comment for specs#59 |
| `worktree-agent-a53d3b2fae9b79d6b` | Local, head `70f5c390ec` | **Viewer-bound release**, the CFC extension this use case asked for, behind the `viewerBoundRelease` option (off by default). A lift asks `currentPrincipal()`; the runtime stamps that viewer on the `TransformedBy` evidence; a module rule releases to that viewer. Also fixes the witness on a record its writer reads and sets, and stale stamps on unchanged outputs. Worked example `patterns/examples/cfc-exchange-rules/card-table.tsx`; draft specification change and an unposted specs#59 comment in `MODULE-NAMED-AUDIENCE-PROPOSAL.md` beside it. See "Viewer-bound release" below |
| `spike/boardgame-trusted-components` | Local | The shell loads a script of boardgame elements when an environment variable is set. Also here as a patch under `fabric/patterns-declarative/labs-spike/` |

## The combined branch

From [status/integration.md](status/integration.md), on commit `1d7668c4`:

- Pass: webapp type check, 1,168 unit tests, the layer test (11), the build,
  the guest (28), the declarative experiment (34), the pattern Node tests
  (36), both multi-user `cf test` suites (202 and 92), both in-sandbox parity
  tests, `go build ./...`, the Go conformance and build-tooling tests, and
  the client check of every game.
- Fail: 1 of 187 renderer tests and 15 of 292 e2e tests. Twelve of the e2e
  failures and the renderer failure were failing before any change of this
  effort. The other three are load-dependent; one of them
  (`waapi-gate.spec.ts:120`) was not run alone, so it is not shown to be a
  flake.
- The JSX-board pattern was rerun afterwards on the integrated branch
  ([status/declarative-pattern.md](status/declarative-pattern.md), "Rerun on
  the integrated branch"): three import breaks from the merge fixed, the
  two-identity game replayed, and a card recorded mid-flight.
- The browser games were played one commit before `1d7668c4`; that commit
  adds two opt-in members to the stack element which the guest does not set.

## What is not held today

Written to the design, with each runtime gap marked `TODO(cfc)` in one place,
as instructed. These are the properties a reader should not assume.

1. **Anyone who can read the table's space can read the sealed state.** The
   memory server delivers a space's documents to every reader whatever their
   label (ruled a non-conformance in specs#95). Observed with a `READ` key.
2. **A view is released over whatever is in the table's cell, not only over
   what `move` wrote.** The release rule ships without its input witness,
   because a record that its writer reads and then sets keeps no witness
   (Memory FINDINGS F2). With the state stored as one string the witnessed
   rule works end to end; we did not ship that shape.
3. **In blackjack every reader of the space can read every seat's view,**
   hole cards included. An unmodified client asks only for its own seat.
   Release to a seat's occupants needs list release (specs#59, labs#8608).
4. **Even with list release, the table's owner can read any hand.** The seat
   spike shows it with a test: the owner binds the seat list's writer from
   another module and lists themselves. `writeAuthorizedBy` names a handler,
   not a run. This needs a ruling; no pattern can close it.
5. **No modified client was tried** against any of this.

Items 2, 3 and 4 no longer describe this branch: the patterns were ported to
the viewer-bound release ([status/viewer-bound-port.md](status/viewer-bound-port.md))
and those three are closed at the runtime's display boundary. Item 1 stands,
and it is now the whole of what is not held: a key with only `READ` printed
the game's secret salt with `cf cell get`. The port also found that a space's
owner has a documented command, `cf cell release-by-intent`, that lists this
module's policy clause as one it would release. That was listed and not run.

## Viewer-bound release

Alex's direction on seeing items 3 and 4: fix it in CFC and the runtime, not
in the pattern. A labs agent built that on `worktree-agent-a53d3b2fae9b79d6b`.

- **Shape.** One sealed record, one `move` handler, one blessed lift that
  takes the record, asks the runtime who it is computed for, and returns that
  principal's view. One rule releases the output to that principal. No list,
  no per-seat cell, no second handler; the seat count is a number in the
  state.
- **Why the owner gains nothing.** The audience is not an input. Binding
  `move` from another module, or running the module again over the live
  state, computes the caller's own view. Both are pinned by tests there.
- **Tested there:** the card table's pattern tests (26 and 28, server
  execution on and off), a three-principal display test (6 tests, 22 steps;
  the one suite the orchestrator reran), and the runner's CFC slice (510
  files, 0 failed).
- **Not done there:** no browser, no Lean, the runner's whole suite (849 of
  about 1,400 files ran), and `check-cfc-correspondence` fails on 16
  citations from the list-release port the branch sits on. A landable pull
  request should be cut without that port.
- **Needs a reviewer who knows the specs pin:** the two fixes are on by
  default, classed as conforming against specs `main`, not the pinned
  snapshot.
- **Still open under it:** a value a pattern derives from the released view
  with code no rule names is sealed again (decision 3 below); the state must
  default to `null`; a modified client is trusted for its own user's view
  when server execution is off.

### The port, measured

From [status/viewer-bound-port.md](status/viewer-bound-port.md), by the
pattern session; the orchestrator read the report and the merged code and
reran none of it.

- **The 18.6 s move was the runtime, and a labs fix removed it.** The
  unchanged pattern takes a median 2.6 s a move and 5.5 s to deal on the
  labs branch. A bisect puts the change on `2033f78fae`, the fix that stamps
  a computation's result whole. Why it is faster is inferred from the commit
  message.
- **The port alone was slower** (move 3.6 s, deal 8.3 s): each viewer's view
  is computed after the commit, about 0.7 s behind it.
- **Final:** deal 1.2 s, move 2.6 s median, round's end 9.6 s; Memory 1.5 s
  a move (2.2 s before). The gains came from storing the last input's
  states as a list the store can diff, keeping only the head of a deal, and
  reading the table once in the lift.
- **What is left follows the versions held:** about 0.6 s for each version a
  move produces. Holding only the head makes every blackjack move 1.2 to
  2.1 s and loses a card at a time; measured, not shipped (decision 4).
- **Tests on the final code:** pattern Node tests 39 of 39; both multi-user
  suites (92 and 202) with server execution off and on; both sandbox parity
  tests; guest 28 of 28; a display-boundary attack test in labs' harness
  (4 tests, 13 steps); three browser games of blackjack with an owner
  watching (424 of 424 secrecy checks) and two of Memory.
- **The JSX-board pattern is on it too**
  ([status/declarative-pattern.md](status/declarative-pattern.md), "On the
  viewer-bound release"): two two-identity Memory games played, 34 of 34
  element tests, on a new local labs branch
  `spike/boardgame-trusted-components-viewer-bound` (`87788f82c8`).
- **A derived value is still withheld from its own viewer.** A `computed`, a
  module-scope `lift` and a `.map` over the released view each showed
  "Content hidden by policy" to the viewer the view was released to. The
  draft proposal says the opposite ("a viewer's own derivations of their
  view stay theirs", citing §5.3). One of the two is wrong, and it should be
  settled before the proposal is filed.

## Userland machinery the runtime should own

Each row is something a pattern author should not have to write. "State"
says what the port found; the full table with the runtime feature each row
waits on is in `status/viewer-bound-port.md`, "What is still in userland".

| What the pattern does today | What the runtime should offer | State |
|---|---|---|
| Stores one view per seat, recomputed on every move whoever is looking | A computation per viewer, released to that viewer | Built in labs; patterns ported |
| Computes each viewer's redacted copy itself and releases the copy whole | Labels on fields of one record, filtered at delivery; needs a join finer than one per transaction (gap F) | Not built |
| Keeps every version of the last input in every view so the client can animate it | A log a reader consumes in order, released per viewer, each entry written once | Not built. This is the largest remaining cost |
| Keeps the move history inside the sealed state | Nothing: our engine reads the whole history. An engine API that takes a version and a tail would shrink it | Left; 5% of the table |
| Reads the view again after `send()` to learn a proposal was refused | A handler's reply to its sender | Not built |
| Copies the table before stepping it, so the stored value holds no reference | A `set` of a value read from the same cell stored as a value | Not built |
| A `tick` proposal a client sends on a timer | A wake at an instant the state names. The labs plan for scheduled work is in minutes, too coarse for a four-second timer | Not built; one `TODO(cfc)` |
| An `identify` handler so a guest learns who it is | `currentPrincipal()` | Built in labs; `identify` removed |
| State held as one string to keep its witness | A record keeps its writer's witness | Fixed in labs; state is a record |

What a pattern cannot remove, measured in labs' harness with no browser:
about half a second of label preparation and commit for a handler that
rewrites a 50 KB record, a view computation per viewer that is mostly the
cost of reading that record, and a second per-viewer round when a serving
runtime computes the views.

## Decisions for you

Ordered by how much later work they shape.

1. **Is the viewer-bound release the mechanism to propose?** It replaces the
   earlier question (is an owner who can read hands acceptable). If yes, the
   next steps are a labs pull request cut without the list-release port, a
   specs pull request from `MODULE-NAMED-AUDIENCE-PROPOSAL.md`, and the
   specs#59 follow-up. None has been opened or sent.
2. **Does "trusted component" mean full trust?** An admitted element runs with
   the shell's authority over the whole page. If that is acceptable for a set
   a user chose, admission is bookkeeping; if not, it is an isolation project.
   Labs has no admission mechanism today: a tag works when the page defines it.
3. **May a pattern reshape a table's released view?** Today a `computed` or
   `.map` over the view is withheld by policy, so the JSX can only pass the
   view whole to one element. Blessing more lifts in the table would let
   pattern authors write seats and labels in JSX.
4. **A card at a time, or speed.** A round's end at blackjack is sixteen
   versions and takes 9.6 s; the median move takes 2.6 s. Holding only the
   newest version makes those 1.5 s and 1.6 s, and cards that the Go client
   moves one at a time then move together. It is a four-line change, written
   out in `status/viewer-bound-port.md`, and not shipped. The other way out
   is the runtime's log (see the userland table).
5. **Classes are gone from the engine,** and `private` with them, because the
   pattern sandbox refuses classes at module scope by design. Members could
   become closure variables if the visibility matters.
6. **Aliases, not npm links, inside the webapp.** The Go build shares one
   `node_modules` across checkouts, so the webapp resolves its own packages by
   alias. A per-checkout build cache would remove three resolution mechanisms.
7. **`check_client.go` was edited** so the checker stages `packages/` with
   `src/`. It is build tooling, not the core, but it is existing Go.
8. **Where the built guest lives.** The Fabric bundle is 1.09 MB and ignored
   by Git, so a checkout cannot deploy without building it.
9. **Go findings to triage.** The engine sessions reproduced Go's behavior in
   every case and listed what looks wrong. The ones players would notice:
   pig's chosen starting player never gets the first turn; `ForceFinishTurn`
   is legal outside its phase; blackjack scores ten plus ace as 11;
   `AnyPlayer.Legal` mutates the move it is asked about. Sanitization of
   dynamic component values depends on map order and can panic. Full lists:
   `status/engine-core.md` ("Divergences and ambiguities"),
   `status/engine-rules.md` ("Things in the Go that look like bugs"),
   `status/engine-legal.md` ("Findings in the Go"),
   `status/conformance-export.md` ("Engine findings").
10. **`ARCHITECTURE.md` still names Go 1.25.6.**

## For the Fabric team

Nothing below has been sent. The specs#59 follow-up drafted at the end of
`labs-spikes/SEAT-VIEWS-FINDINGS.md` is superseded by the one at the end of
`MODULE-NAMED-AUDIENCE-PROPOSAL.md` on the labs branch.

| Finding | Kind | Where |
|---|---|---|
| Served handlers attributed to nobody | Bug, fixed in labs#10111 | PR description |
| A record its writer reads and sets keeps no witness | Specification question | Memory FINDINGS F2; sealed-state findings, gap A |
| A declared writer and a witness are not bound to the run that owns the document | Needs a ruling | Seat findings, (c) |
| The list-release draft refuses every compiled pattern at capture | Two changes made on the port, one of them a model | Seat findings, "The port" |
| `cfcPattern.members()` is in the draft's API and the transformer refuses it | Bug in the draft | Seat findings |
| A cell position typed `unknown` reads back as `{}` silently | Gap | Memory FINDINGS F7 |
| `pattern-tests` runs in CI only with server execution off | CI gap | labs#10111 |
| No way for a guest to read the viewer's own principal | Missing feature | `status/fabric-memory.md`, question 6 |
| A module cannot release to a principal its sealed state names | Semantic gap; proposal drafted, option D built behind a flag | `MODULE-NAMED-AUDIENCE-PROPOSAL.md` on the labs branch |
| Machinery patterns build that the runtime should own | Requests | "Userland machinery the runtime should own" above |

A question to ask them, drafted and not sent:

> We have a set of Lit elements (cards, stacks, an animator) that we want
> patterns to use directly in JSX, not only inside `cf-iframe`. On a local
> spike the shell loads them from a script and a pattern types the tags
> itself; nothing else gates a tag. What is the intended path for a component
> set a user chooses to trust: a record of what the user admits, loading by
> digest, a namespace, and should an admitted element run with the shell's
> full authority or isolated? We would build to whatever shape you intend.

## What remains

- **Stream A:** compiled packages with an explicit `define()` entry and no
  registration on import (S6); game clients and generator templates moved to
  package specifiers; token art inside a package (S5); the host's required
  custom properties written down. Table in `status/client-packages.md`.
- **Stream B:** the declarative pattern onto the per-viewer view;
  the remaining four games as patterns (they load and compile in a pattern
  and are otherwise untried there); latency.
- **Stream C:** what the corpus cannot yet check is listed in
  `status/conformance-export.md` and `status/engine-rules.md` ("Unproven").
- **Stream D:** admission of a trusted component set in labs; restoring the
  input witness when F2 is settled, which changes the policy digest and
  strands tables in flight.

## Before landing

- Read `status/integration.md` for the suites on the head commit. The webapp's
  e2e suite had failures before this work and is flaky under load; no two full
  runs in this effort failed exactly the same tests.
- The branch is 3 waves of merges from session branches (`claude/fabric-*`).
  Those branches and their worktrees under `.claude/worktrees/` can be removed
  once this one is reviewed.
- `go test ./...` and `npm run type-check:strict` were not run by anyone; both
  were documented as not green before this work.
