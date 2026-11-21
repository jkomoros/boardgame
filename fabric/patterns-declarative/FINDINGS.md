# Boardgame elements in pattern JSX: findings

Stream D, step 3. Labs paths are relative to a labs checkout at
`fix/served-handler-write-identity` (`4815052651`). "Ran" means run and read
here, in headless Chromium against labs' local servers on port offset 460,
Node v24.22.3, Deno 2.9.8. "Read" means read in source and not exercised.

## Verdict

**It works, and the reconciler does not defeat it.** A complete two-identity
game of Memory was played, twice, through a pattern whose JSX is boardgame
elements in the shell's own document. The table, `move` and `seatView` are
imported unchanged from `fabric/patterns/memory`.

It works for one shape only: **the pattern hands the released view to one
element, whole, and everything under that element is markup whose props never
change.** The two other shapes an author would reach for both fail, for
different reasons, and both failures were reproduced (R2, R3 below).

It is not yet a trusted component set. The labs change is 55 lines that load
a script; nothing in it is a policy (see "How far from a set a user's policy
admits").

## 1. How labs admits an element into pattern JSX today

| Question | Answer | Where |
|---|---|---|
| Where is the `cf-*` set registered? | Each component calls `customElements.define` in its own `index.ts`; the shell's entry imports the whole set | `runtime/ui/src/v2/components/cf-button/index.ts:5`; `runtime/shell/src/index.ts:3` (`import "@commonfabric/ui"`); built by felt from `runtime/shell/felt.config.ts` |
| How is a tag typed? | By hand, in one closed interface: `JSX.IntrinsicElements`, whose index signature is commented out. The file reaches a pattern's compile as a vendored lib | `runtime/html/src/jsx.d.ts:5174`-`5175`; `base/static/assets/types/jsx.d.ts` (a symlink to it); `base/js-compiler/typescript/options.ts:100` |
| Is an unknown tag rejected? | By the type checker only: `Property 'bg-table' does not exist on type 'JSX.IntrinsicElements'`. **A pattern can add the tag itself** by declaration merging, in its own file or in a module it imports, and then it compiles and deploys (Ran) | `elements.ts` here |
| Is there a worker-to-DOM allow-list of tags? | **No.** The reconciler drops `script` and nothing else; the applicator calls `document.createElement(tagName)` for any name | `runtime/html/src/worker/reconciler.ts:4055`; `runtime/html/src/main/applicator.ts:357` |
| What does gate a tag, then? | Whether the page defined it. An undefined tag is an inert element | Ran: the elements upgrade and draw once the shell loads the bundle |
| How do props cross? | As **properties**, not attributes (`target[key] = value`), except `data-*`, `aria-*` and a string `style`. A value is a `FabricValue`: plain data, structured-cloned; a cell bound to a prop is resolved in the worker and its value sent. Only a `$prop` binding hands the element a live `CellHandle` | `runtime/html/src/render-utils.ts:118`-`152`; `runtime/html/src/vdom-ops.ts` (`SetPropOp`); `applicator.ts` `#setBinding` |
| How do events cross? | Any prop starting `on` is a listener for the lower-cased rest, so `onpropose-move` listens for `propose-move`. No allow-list of event names. The event is serialized to a fixed set of scalar fields plus `detail`, which must be a `FabricValue` or the event is lost with a thrown error | `runtime/html/src/render-utils.ts:26`-`29`; `runtime/html/src/main/events.ts:153`-`166`, `toSerializableValue` |
| Does a gesture's trust cross? | `provenance: { origin: "dom", trusted: true }` is attached only when `event.isTrusted`. A custom element's own `CustomEvent` is never trusted, so `propose-move` arrives with no provenance. The handler runs all the same | `runtime/html/src/main/event-provenance.ts:36`-`53`; Ran |
| Are children keyed? | By a hash of the child's content, with a cell keyed by its link. The JSX `key` prop is declared and ignored | `runtime/html/src/worker/keying.ts:86`-`131`; `reconciler.ts` `#updateChildren` |
| Is behavior keyed to tag names anywhere? | Yes: tables in the reconciler say which props of which `cf-*` tag load a remote resource, show text under an integrity policy, or take a reference. A tag not in them gets the default for every prop | `reconciler.ts:124`-`208` (`TEXT_INTEGRITY_PROP_SINKS`, `REFERENCE_BINDING_SINKS`, `ACCESS_PLACEHOLDER_BINDINGS`, `REMOTE_LOAD_PROPS`) |
| What do the docs say about third-party or user-trusted components? | Nothing that describes a mechanism. Three statements of intent: narrow the surface "to trusted components"; future work should define "opaque trusted UI islands"; a component that needs isolation should be hosted out of process | `docs/plans/security-privacy-roadmap.md:52`; `docs/plans/runner_cfc_implementation.md:1065`; `docs/specs/sandboxing/cross-origin-isolation.md:205` |
| Is there a CSP on the shell page? | None was found on the page or the toolshed's shell routes (Read, by search) | `runtime/shell/public/index.html` |

So today the `cf-*` set is "trusted" by being compiled into the shell.
Nothing else distinguishes it at run time.

### What "a user's policy admits trusted components" needs that does not exist

1. **A record of admission.** Somewhere a user (or a space) says "this set,
   at this digest". Nothing holds one. The spike uses an environment variable
   read when the shell is built.
2. **Loading by that record, after sign-in, checked against the digest.** The
   shell has one page-level bundle and an immutable `/builds/<sha>/`
   namespace for its own scripts; no path loads code a user named.
3. **A confined namespace.** `customElements` is one registry for the page.
   An admitted set could define `cf-button` first, or any tag another set
   wants. Admission has to bind a set to a prefix and refuse the rest.
4. **An enforced tag check.** Today an unadmitted tag is merely inert. Under
   a policy, the reconciler or applicator must refuse a tag that is neither
   built in nor admitted for this viewer, since what the tag does now depends
   on who is looking.
5. **Per-element entries in the policy tables.** `REMOTE_LOAD_PROPS` and its
   neighbors are literals in the reconciler. An admitted set must declare
   which of its props can cause a fetch or show text, and the reconciler must
   read that. `bg-stack`'s `face` can name card art; the element restricts it
   to relative and `data:image/` URLs itself, and labs does not know.
6. **Types that travel with the set.** Already possible: `elements.ts` is in
   this repository and labs needed no change for it. But nothing ties the
   types a pattern compiled against to the set a viewer loaded.
7. **Isolation, or an explicit decision to do without it.** An admitted
   element runs in the shell's document with the shell's authority: it can
   read the DOM of every other piece on the page and reach
   `globalThis.commonfabric`. `cf-iframe` exists because of that. A trusted
   set is trusted with everything, which is a far stronger grant than
   "may draw cards".
8. **Trusted-gesture provenance for a set's own events.** An element's
   synthetic event carries no `provenance`. `bg-stack` emits `propose-move`
   only for a trusted click, but labs has only the element's word for it. If
   a handler is ever to require a real gesture, an admitted element needs a
   way to vouch that the runtime can check.

## 2. The labs spike

`labs-spike/0001-spike-shell-trusted-components.patch`, on local branch
`spike/boardgame-trusted-components` (`1ef3d7ce99`), never pushed. Three
files, 55 lines:

- `runtime/shell/felt.config.ts`: with `CF_SPIKE_TRUSTED_COMPONENTS` naming a
  built ES module, add it as the entry `scripts/trusted-components` and
  define `$SPIKE_TRUSTED_COMPONENTS`.
- `runtime/shell/src/lib/spike-trusted-components.ts`: append a module
  `<script>` for it at boot.
- `runtime/shell/src/index.ts`: import that module.

**Built bundle, by path; no source copied.** The boardgame repository builds
`fabric/declarative/dist/bg-elements.js` (409 kB, 119 kB gzipped, Lit
included, no imports). Labs is given its path. A path import of the
TypeScript was not used: labs' esbuild would have to resolve `lit` from
`server/static/src`, which has no `node_modules` of its own.

**Typed for JSX with no labs change**, by `elements.ts`.

Nothing in the runtime, the reconciler, the applicator or `jsx.d.ts` was
changed. That is itself the finding of section 1.

## 3. The pattern, and what the real reconciler did

`memory/main.tsx` is 60 lines of JSX and four small handlers:

```tsx
<bg-table name="memory" gameName="memory" chest={info.chest}
          tableView={view} principal={me}
          onpropose-move={play({ move: propose })}
          ontimer-elapsed={tick({ move: propose })}>
  <bg-stack path="Game.Cards" layout="grid" messy face="Type"
            pick={{ move: "Reveal Card", arg: "CardIndex", slots: "hidden" }} />
  <bg-stack path="Players.0.WonCards" layout="stack" messy face="Type" />
  ...
</bg-table>
```

The first experiment's two findings are applied: `bg-table` is the scoped
animation root, and components are paired by id. Three things were added to
`fabric/declarative` for this:

- `tableView` and `principal` on `bg-table` (`src/table-view.ts`). The table
  takes the released `TableView` untouched, finds the viewer's seat, and
  queues every version in the window it has not queued.
- `timer-elapsed`, a second event out of `bg-table`, so the hide-cards timer
  needs no script.
- `adoptEarlyProperties`, for a host that assigns a property before the
  element is defined.

### R1. The recommended shape: nothing is recreated (Ran, two games)

| | Game 1 | Game 2 |
|---|---|---|
| Reveals played | 16 | 16 |
| Mismatches hidden by the timer | 3 (5.0 to 6.1 s) | 3 (5.2 to 6.1 s) |
| Winner, on both screens | seat 1, 6 cards to 4 | seat 0, 6 cards to 4 |
| Times each `bg-table` and `bg-stack` was connected | not traced | **1 each, both players** |
| `tableView` assignments a player | not traced | 24 |
| Versions installed by the table | 35 | 35 |
| Motion cycles; with an animation | 35; 24 | 35; 24 |
| Animator's `plays` counter (`__bgAnimTestHooks`) | 58 | 58 |
| Watchdog firings; interrupted cycles | 0; 0 | 0; 0 |
| "Content hidden by policy" on screen | 0 | 0 |

24 assignments carried 35 versions: the window delivered versions the
reconciler never showed as "the newest value". A table fed only the head
would have skipped 11 of them, and with them the reveal before each capture.

The proposer was the runtime's throughout: each table found its own seat by
comparing `me` (written from `currentPrincipal()` in the click's handler)
with the seat `move` recorded, and `move` reads `currentPrincipal()` itself,
one stream send away from the handler the DOM event reached. A click out of
turn was never needed, since the elements disable what the view says is not
legal.

### R2. A value the pattern derives from the released view is withheld (Ran)

`memory/probe-release.tsx`, two identities, after the deal:

| Read of the released view | On screen |
|---|---|
| `{view?.revision}` | `3` |
| `computed(() => view?.revision)` | "Content hidden by policy" |
| `view?.seats.map(seat => <span>{seat.label}</span>)` | "Content hidden by policy" |
| `<bg-stack stack={view?.entries[0]?.state.Game.Cards}>` | 10 cards |

The table's rule releases what `seatView` computed. A `computed` or a `.map`
in the pattern is another computation, and its result is not that. So a
pattern may read the released view by a fixed path and may not reshape it.

Consequences, each of which shaped the design:

- "The newest version" is `entries[entries.length - 1]`, which is not a fixed
  path. `memory/per-stack.tsx` tries it and draws nothing (Ran: two
  placeholders, an empty board). **Choosing the version to show must happen
  in the element**, which is why `bg-table` takes the whole view.
- A variable number of seats cannot be a `.map` in JSX. `main.tsx` has two
  piles written out. A game for two to six needs a repeater element.
- The same goes for any text computed from the view: "Alice won", a score.
- This matches the Memory pattern's F3 and F4 and extends them: there the
  guest read the view whole; here JSX reads it, and only path reads survive.

This is a property of the Memory table's policy, not of the elements. A
table whose rule also blessed a second lift could hand out reshaped values.

### R3. With nothing confidential, per-stack props still do not animate (Ran)

`reconciler/main.tsx`: six cards, public state, a grid bound to a computed
stack and two piles from a `.map`. Six presses of "Next".

| Observation | Result |
|---|---|
| Times each element was connected | 1 each: the `.map` did **not** recreate its elements |
| Stack assignments | 15, in 13 separate groups: 3 together at first render, then **every one alone** |
| Gap between the grid's assignment and the pile's, for one press | about 3 ms |
| Motion cycles with an animation | **0 of 9** |
| Cycles interrupted by the next assignment | 3 |

Each cell bound to a prop is its own subscription and its own batch. The
card leaves the grid in one batch and arrives in the pile in the next, so no
single change ever contains both ends of the move. The element sees a card
vanish, then a card appear.

### Did the real reconciler behave as the simulation did?

| First experiment's simulation | Real reconciler |
|---|---|
| Recreating every child on each state kills animation unless the table remembers geometry by component id | **Not what happens.** Elements bound to cells, including those from a `.map`, are kept and their props updated. Remembered geometry was never exercised in a game |
| Recreating only changed children: same | Same answer: no recreation observed |
| Recreating the table itself survives through `retainedByName` | Not exercised: the table was connected once |
| Separate properties can pair one version's view with another's legality, hence one atomic `frame` | **Confirmed, and worse than argued.** Props do arrive separately, a few milliseconds apart, and that alone removes every flight (R3) |
| A table scoped to its subtree | Held: one table a page, no interference seen. Two tables on one labs page were not tried |
| (not simulated) | Intermediate versions are lost unless the input is a window (R1: 24 assignments, 35 versions) |
| (not simulated) | Policy forbids the pattern reshaping the view at all (R2) |

The simulation tested the wrong hazard. The mechanism it produced (geometry
by id) is harmless and unused here; the design decision it argued for on
other grounds (one atomic input owned by the table) is what made this work.

## 4. Element changes requested of the client session

Nothing new in `server/static/src` was needed to run this. The wrapper still
reaches into two privates, as in the first experiment. These two diffs would
remove that; neither changes behavior for a caller that does not opt in.

`server/static/src/components/boardgame-component-stack.ts`, a stable
collection id (the wrapper redefines `id` on the instance today):

```diff
+  /** A stable id for this collection. Random when the host gives none. */
+  @property({ type: String, attribute: 'collection-id' })
+  collectionId = '';
+
   get id(): string {
-    return this._id;
+    return this.collectionId || this._id;
   }
```

Same file, a scope for the animator's stack list (the wrapper redefines
`_sharedStackList` on the animator's inner stack today):

```diff
+  /** When set, the stacks an animator built on this element sees. */
+  stackScope: (() => BoardgameComponentStack[]) | null = null;
+
   get _sharedStackList(): BoardgameComponentStack[] {
-    return sharedStackList;
+    return this.stackScope?.() ?? sharedStackList;
   }
```

**Withdrawn:** the first experiment's third request, a public way to hand
the animator a remembered "before". The real reconciler does not recreate
elements, so it is not needed for this path. It stays needed for a host that
does recreate (none is known).

## 5. How far from a set a user's policy admits

| Piece | State |
|---|---|
| Elements a pattern's JSX uses directly, with types | Done, here |
| A full game through them, two identities, policy-sealed table | Done, here |
| Loading the set into the shell | Spike: an environment variable at shell build time |
| A user or space saying which sets it admits | Does not exist |
| The runtime refusing an unadmitted tag | Does not exist; today such a tag is inert |
| Namespace confinement, digest pinning | Does not exist |
| The set declaring its props to the remote-load and integrity tables | Does not exist |
| Isolation of an admitted set from the rest of the page | Does not exist, and is the hard part: an admitted element has the shell's authority |
| A repeater, text from state, a lobby, as elements | Not built; needed because of R2 |

The elements are the small part and are mostly done. Admission is a labs
design that has not started: items 1 to 5 and 7 of section 1.

## 6. Smaller things

- Every JSX prop is a property, so the names are `gameName`, `holdBefore`,
  `fauxComponents`, not the attribute spellings. `class` is `className`.
- Lit re-applies its own reactive properties after an upgrade, but a plain
  accessor assigned before the element is defined stays shadowed by the
  instance value. `adoptEarlyProperties` fixes that for `tableView`,
  `principal`, `frame`, `view`, `chest` and a stack's `stack`. It was not
  observed to matter: the shell defines the elements long before a piece
  renders.
- The lobby is two `cf-button`s that are always there, because nothing of the
  table is released before its first move (the Memory pattern's F5) and a
  condition on the view would be a derived value (R2).
- The first move of a table took about 15 s to show; later ones a median of
  2.0 s from click to face, the animation included.
- Two copies of Lit are on the page (labs' and the bundle's). Nothing broke.
- The script's sampler of running Web Animations read 0 on almost every move
  and was not diagnosed, and no screenshot caught a card in flight. The
  evidence that animations ran is the animator's own `plays` counter and the
  table's count of cycles that started one.
