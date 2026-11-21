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
| Memory on Fabric | `fabric/patterns/memory`, [status/fabric-memory.md](status/fabric-memory.md), [status/engine-fabric.md](status/engine-fabric.md) | Sealed table, one `move` handler, per-viewer view from the engine. Two browser identities played whole games with server execution on |
| Blackjack on Fabric | `fabric/patterns/blackjack`, [status/engine-fabric.md](status/engine-fabric.md) | Per-seat views from the engine's sanitization; parity with Go on seven recordings; multi-user `cf test` passes. Browser play: see `status/integration.md` |
| Board in pattern JSX, no iframe | `fabric/patterns-declarative`, [status/declarative-pattern.md](status/declarative-pattern.md) | Two identities played Memory through a pattern whose JSX is boardgame elements, on a local labs spike branch |
| Go 1.27 | commit `3a00e697` | `golang.org/x/tools` raised so the tree builds and tests under Go 1.27.4 |

## What lives in labs

| Branch (in `commonfabric/labs`) | State | What |
|---|---|---|
| `fix/served-handler-write-identity` | Pushed, draft PR commonfabric/labs#10111, CI green | Served handlers are attributed to the handler. Every pattern here depends on it for a witnessed release |
| `worktree-agent-ad697e53f88bab905` | Local | First spike: a sealed state machine. Findings copied to [labs-spikes/SEALED-STATE-MACHINE-FINDINGS.md](labs-spikes/SEALED-STATE-MACHINE-FINDINGS.md) |
| `worktree-agent-ac9c0255af3654f11` | Local | Port of the list-release draft (labs#8608) and a card table with per-seat views. Findings copied to [labs-spikes/SEAT-VIEWS-FINDINGS.md](labs-spikes/SEAT-VIEWS-FINDINGS.md), which ends with an unposted draft comment for specs#59 |
| `spike/boardgame-trusted-components` | Local | The shell loads a script of boardgame elements when an environment variable is set. Also here as a patch under `fabric/patterns-declarative/labs-spike/` |

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

## Decisions for you

Ordered by how much later work they shape.

1. **Is a table owner who can read hands acceptable for now?** It decides
   whether blackjack waits for a specs ruling or ships as a demo. Item 4 above.
2. **Does "trusted component" mean full trust?** An admitted element runs with
   the shell's authority over the whole page. If that is acceptable for a set
   a user chose, admission is bookkeeping; if not, it is an isolation project.
   Labs has no admission mechanism today: a tag works when the page defines it.
3. **May a pattern reshape a table's released view?** Today a `computed` or
   `.map` over the view is withheld by policy, so the JSX can only pass the
   view whole to one element. Blessing more lifts in the table would let
   pattern authors write seats and labels in JSX.
4. **Speed or Go's exact state.** Memory on the engine takes a median 3.7 s a
   move against 1.8 s with hand-written rules, because Go's Memory carries the
   110 unused cards in every version and the runtime's cost follows stored
   size. Options are in `status/engine-fabric.md`.
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

Nothing below has been sent. The specs#59 follow-up is drafted at the end of
`labs-spikes/SEAT-VIEWS-FINDINGS.md`.

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
- **Stream B:** blackjack's designed seat release once list release lands;
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
