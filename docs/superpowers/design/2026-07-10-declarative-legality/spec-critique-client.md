# Spec Critique — The Designed-for-Client Future & the Explainability Contract

**Spec:** `docs/superpowers/specs/2026-07-10-declarative-legality-design.md`
**Critic angle:** the promise that this server-first representation "supports a
later TypeScript evaluator and rich client UX." Stress-testing that promise
against how static data actually reaches the client today (chest / `data.Chest`
in `expandMoveForms`), the sanitization policies in `sanitization.go`, and the
#693 hidden-info constraint.

The design is genuinely strong on the engine and anti-tarpit fronts. But its
central client-facing promise — "localizable, greppable, re-renderable on server
or client" messages plus a "TS evaluator [that] re-runs evaluable predicates
locally" — rests on **three artifacts the spec never specifies**: (1) where the
template table lives, (2) how a TS predicate implementation stays in lockstep
with its Go twin, and (3) precise evaluability rules. Each is load-bearing;
without them the promise is not shippable, and one is not even *server*-shippable
today. Findings below, by severity.

---

## BLOCKING

### B1. The key→template string table has no home — and this breaks the *server*, not just the future client

The spec is emphatic that a `Message` is "a template key plus named bindings —
**never a pre-baked string**":

> ```go
> type Message struct {
>     Template string         // "reveal.no_cards_left"
>     Bindings map[string]any // {"left": 0}
> }
> ```

and that this is "re-renderable on server or client." But **nowhere does the
spec say where `"reveal.no_cards_left"` → `"You have no cards left to reveal
this turn"` lives.** The migrations lean on this hard — every acid test carries
inline comments like `// "You have no cards left to reveal this turn"` — yet
those strings exist only as Go source comments. There is no registry, no
per-game map, no chest field, nothing.

This is not merely a client gap. Trace the server path the spec itself
describes:

- §3: "*Nothing flattens to a rendered string until the last moment
  (`Verdict.Error()` adapts to `error`/`errors.Friendly` for the Legal()
  return)*."
- `Verdict.Error()` must therefore produce a rendered string. I read
  `errors/main.go`: `errors.Friendly` holds `msg`/`friendlyMsg` **strings** — it
  has no template-key concept whatsoever. `Friendly.Error()` returns a
  pre-baked string.

So `Verdict.Error()` cannot produce anything renderable without a table to
resolve `Template` + `Bindings` → string. **The server's own `Legal() error`
return is unrenderable as specified.** This is a hole in the shipped v1, not the
deferred TS follow-up. The spec's §8 claim "All three error strings preserved
verbatim as templates" is currently false: they are preserved as *comments*.

**Minimally-viable fix — ship the table in the chest, exactly like Enums.**

I verified the existing static-data pipeline end to end:
- `server/api/main.go:1487` ships `"Chest": game.Manager().Chest()` in the info
  payload.
- `ComponentChest.MarshalJSON` (`component_chest.go:192`) marshals a fixed
  `{Decks, Enums, Constants}` struct.
- The client's `expandMoveForms` (`actions/game.ts:377`) reads `data.Chest`,
  then `(chest as any).Enums[field.EnumName]` to resolve enum values into forms.

Templates are *structurally identical* to Enums: a game-static, boot-time-fixed,
string-keyed table the client already knows how to receive via the chest. The
MVA:

1. New optional delegate method rhyming with `ConfigureEnums`:
   ```go
   // GameDelegate
   func (g *GameDelegate) ConfigureLegalTemplates() map[string]string
   // base.GameDelegate merges legal.DefaultTemplates() (built-in predicate
   // messages: proposerIsCurrentPlayer, inPhase, stackConstraints, ...) with
   // the game's overrides. Fail-fast at NewGameManager: every Template key
   // referenced by any WithMessage / FailT / Errorf in any plan must resolve,
   // else boot error with the move name + missing key.
   ```
   This validation dovetails with the spec's existing boot-time path-validation
   story ("a typo'd path fails at boot") — a typo'd *template key* should fail
   the same way, and the same validation pass can check both.
2. Add `Templates map[string]string` to `ComponentChest.MarshalJSON`'s struct.
   Now every client that receives the chest receives the template table for
   free, through the exact channel it already trusts for Enums.
3. `Verdict.Error()` renders server-side by looking up the manager's template
   table and interpolating `Bindings` (a trivial `{{.left}}`-style pass, or
   ICU-lite if plurals matter — see B1a). Feed the rendered string into
   `errors.NewFriendly(rendered, Fields(bindings))` so nothing downstream
   changes.

**Why the chest and not `client_config.js`:** `client_config.js` is
deploy-global (offline mode flag, etc.), whereas templates are **per-game-type**
and must version with the game package. The chest already has exactly that
scope and lifetime. Do not invent a second static-data channel; #feedback:
prefer existing primitives.

**B1a — bindings-to-string is underspecified.** `{"left": 0}` →
`"...no cards left..."` implies the template *ignores* the binding, but
`{"left": 3}` might want "3 cards left." The interpolation grammar (named
placeholders? plural forms?) is unspecified. MVA: named `{{.key}}` interpolation
only, no logic (consistent with the anti-tarpit "no user arithmetic" ethos). If
plurals are needed later, they belong in a richer template value, not in
predicate code. State this explicitly so Go and TS renderers agree
byte-for-byte.

---

### B2. "TS evaluator re-runs evaluable predicates" requires a second implementation the spec never budgets, versions, or conformance-tests

The client contract is explicit:

> the future TS evaluator re-runs evaluable predicates locally against the
> sanitized state (live graying, zero round trips)

Re-running `playerPropAtLeast` in TS means a **TypeScript implementation of the
predicate's semantics** — `Spec{name, args, sub}` + the ledger is the *identity*
and the *inputs*, but not the *behavior*. There must be exactly one TS function
per catalog `name` whose Kleene three-valued output (Pass/Fail/**Unknown**)
matches the Go `Evaluate` on every input. The spec is completely silent on:

1. **Where the TS catalog lives** and how it's kept in existence (it doesn't
   exist yet — fine, it's deferred — but the *representation* must make it
   buildable, which is the thing being reviewed now).
2. **Versioning/conformance across Go and TS.** The spec's own anti-tarpit rules
   show the authors *know* this is the sharp edge — they cut first-class `not`
   precisely because "*client and server must agree exactly on `Unknown`
   semantics*." Good instinct, but the same divergence risk applies to **every
   leaf predicate and the `any` compositor**, not just `not`. Cutting `not`
   narrows the surface; it does not close it. A `playerPropAtLeast` whose TS twin
   handles a sanitized `PolicyLen` stack differently from Go will gray out a
   legal move — a *worse* UX than no local evaluation, because it's confidently
   wrong.

**Fix — make conformance a first-class deliverable of the representation, via a
shared JSON test corpus.**

The `Spec` is already serializable and `Context` is already "no I/O, no
mutation, four values." That is *exactly* the precondition for a
language-neutral golden corpus. Propose:

- A repo-committed `legal/conformance/*.json` corpus, one file per catalog name,
  each case:
  ```jsonc
  {
    "predicate": {"name": "playerPropAtLeast", "args": ["player.Cards", "1"]},
    "state": { /* sanitized SubState fixture */ },
    "move":  { /* or null for field-independent */ },
    "proposer": 0,
    "expect": {"outcome": "fail", "message": {"template": "...", "bindings": {"left": 0}}}
  }
  ```
- The Go unit tests the spec already promises (§9: "every catalog predicate —
  Pass/Fail/Unknown cases … registry round-trip") **generate and consume** this
  corpus rather than hand-rolling assertions. The corpus is the artifact; the Go
  test is one consumer.
- The future TS evaluator ships with a runner that loads the same corpus and
  asserts identical `Verdict`s. Conformance is then a CI gate, not a hope. A new
  catalog predicate that lacks corpus coverage is un-shippable; a TS twin that
  fails a case fails CI.
- **Catalog versioning:** stamp the chest's template/catalog payload with a
  `catalogVersion` integer. If the client's bundled TS catalog is older than the
  server's, the client sets `evaluable:false` wholesale and falls back to
  server verdicts. This is the honest degradation the spec already espouses,
  applied to the version-skew case it forgot.

This is a **representation-level** requirement: without the corpus contract baked
in now, the "designed-for" claim is aspirational. It's cheap to add now
(fixtures + a JSON schema) and prohibitively expensive to retrofit after the
catalog has 30 predicates and two divergent implementations.

---

## IMPORTANT

### I1. `evaluable` is defined as "every Reads() path is Visible" — this is *too strict* and silently kills the very common case the spec brags about

The evaluability rule:

> ```
> evaluable = has a serialized form (not the escape hatch)
>           ∧ every Reads() path is Visible under the sanitization
>             transformation applied for this viewer
> ```

But the spec then *contradicts its own rule* in §6:

> memory's stacks are `sanitize:"order"` — slot occupancy is visible — so
> memory's entire plan is client-evaluable.

Under a literal "every Reads() path is **Visible**" rule, a `PolicyOrder` stack
is **not** `PolicyVisible`, so `RevealableCardAt` (which reads
`game.HiddenCards`, a `PolicyOrder` stack) would compute `evaluable:false` — and
memory's plan would *not* be client-evaluable. The spec's headline example
fails its own definition. The rule as written is wrong; the prose is right.

I read `sanitization.go` to get the policy lattice precisely
(`PolicyVisible > PolicyOrder > PolicyLen > PolicyNonEmpty > PolicyHidden`):

| Policy | What survives for a stack |
|---|---|
| `PolicyVisible` | all component values |
| `PolicyOrder` | slot occupancy + order; values are generic |
| `PolicyLen` | length only (sorted generic components) |
| `PolicyNonEmpty` | zero vs. non-zero |
| `PolicyHidden` | empty |

Different predicates read different *facets*. `RevealableCardAt` only asks "is a
component present at index N?" — that's **occupancy**, satisfied by
`PolicyOrder`. A hypothetical `stackSizeAtLeast` only needs **count**, satisfied
by `PolicyLen`. A `propCompare` reading a component's `Color` genuinely needs
**values**, i.e. `PolicyVisible`. Collapsing all three into "needs Visible"
makes the client blind exactly where the spec claims it sees.

**Fix — per-predicate visibility requirement, expressed as a minimum policy per
read path.** Extend the `Reads()` contract from bare paths to
`(path, visibility-need)`:

```go
type ReadNeed int
const (
    NeedsOccupancy ReadNeed = iota // satisfied by PolicyOrder or stronger
    NeedsCount                     // satisfied by PolicyLen or stronger
    NeedsValues                    // requires PolicyVisible
)
type PropRead struct { Path PropPath; Need ReadNeed }
```

Then:
```
evaluable = has serialized form
          ∧ ∀ read ∈ Reads(): effectivePolicy(read.Path, viewer) ⊒ read.Need
```
where `⊒` is the policy lattice. `RevealableCardAt` declares
`NeedsOccupancy` → evaluable under `PolicyOrder` → memory's plan is
client-evaluable, matching the §6 prose. This also makes `Reads()`
conservativeness (a §10 risk) *more* auditable, because the lint helper can
check the declared `Need` against what the `Evaluate` body actually touches.

Note this facet-need also must match what the **TS evaluator physically
receives**: the sanitized state on the wire for a `PolicyOrder` stack has
generic components, so the TS predicate must only ever consult occupancy for
that path. The `Need` declaration *is* the contract that keeps the TS twin from
reaching for a value that isn't there — tying I1 directly to B2's conformance
corpus (corpus fixtures must include sanitized states at each policy level).

### I2. Evaluability edge cases the spec leaves implicit — define them

The spec's `Context` has four fields and a note that `Move` is "nil during
field-independent evaluation," but the evaluability rules for several cases are
undefined. Each needs a normative answer:

- **`player.X` when the viewer isn't the current player.** `player.X` resolves
  to "the *current* player" (path grammar table). But sanitization is
  *per-viewer*: viewer 0 looking at a move whose `player.Cards` refers to the
  current player 1 sees player 1's `Cards` under player-1's-private policy
  (typically `PolicyHidden` for another player's hand). **Rule:** evaluability
  is computed against *the viewer's* sanitized projection, not the current
  player's. So `player.CardsLeftToReveal` may be `evaluable:true` for the
  current player (viewing themselves) and `evaluable:false` for an observer. The
  ledger is **per-viewer**, which the spec already implies ("computed
  server-side, per predicate, **per viewer**") but never connects to the
  `player.X`-indirection. Make explicit: *"a `player.X` read's visibility is
  evaluated as the viewer seeing the current player's property, with all the
  cross-player hiding that implies."*

- **`move.X` fields the user is editing — always client-known?** The spec
  assumes field-dependent predicates become evaluable once the client binds
  fields locally. Mostly true: the client authored the field value. **But** a
  `move.X` whose type is a component reference into a hidden stack (e.g. a card
  ID the client picked from its *own* hand but which is `PolicyHidden` to the
  proposer-as-observer) can be locally known yet reference state the predicate
  then reads. The correct rule: `move.X` **values** are always client-known
  (the client typed them), so a predicate reading *only* `move.X` is always
  evaluable; but a predicate that reads `move.X` *and* dereferences it into a
  state path (`RevealableCardAt` reads `move.CardIndex` **and**
  `game.HiddenCards`) is gated by the *state* path's policy, not the move
  field's. State this: **move-field reads are always evaluable; the
  evaluability gate is the state paths a predicate also reads.**

- **Computed properties in paths.** §1 anti-tarpit rule 3 explicitly routes
  branchy logic through "a computed state property (which a `propCompare`
  predicate can then read)." But a computed property's *sanitization policy* is
  whatever the delegate assigns — and critically, a computed prop may derive
  from hidden inputs while being *itself* `PolicyVisible`. That's fine for
  evaluability of the *reader* (the computed value is visible), **but it can
  leak** (see I4). For evaluability: **the `Reads()` of a predicate over a
  computed property is the computed property's path, and its visibility is the
  computed property's own policy** — the engine must not try to expand into the
  computed property's hidden inputs (it can't; computed props are opaque
  functions). Make explicit that `Reads()` stops at the computed-property
  boundary.

- **`any` with mixed evaluable/inevaluable children — is the compositor
  evaluable?** This is the sharpest omission. `any` is Kleene-OR:
  `Pass ∨ Unknown = Pass`, `Fail ∨ Unknown = Unknown`, `Fail ∨ Fail = Fail`.
  So the compositor's *outcome* can be decidable even when a child is Unknown
  (if another child Passes). But `evaluable` as defined is a static per-predicate
  boolean, computed from `Reads()`, *before* evaluation. **Rule needed:** an
  `any` node is `evaluable` iff **at least one child is evaluable** (because a
  single evaluable Pass decides it), BUT its *reported verdict* must honor Kleene
  — if the evaluable children all Fail and an inevaluable child exists, the
  compositor's client-side verdict is **Unknown**, not Fail. The TS evaluator
  must therefore evaluate the evaluable children *and know an inevaluable child
  exists* to correctly return Unknown. This means the ledger entry for an `any`
  compositor must serialize its children's `evaluable` flags (the nested-node
  wire format already supports this via `sub`, but the spec's ledger example
  only shows flat leaves — extend the example to show a nested `any` with
  per-child `evaluable`). Without this, the client will report a hard Fail on a
  move that might actually be legal.

### I3. The ledger doesn't distinguish "fails with server defaults" from "would pass with different fields" — the LegalForAnyone analog is lost

The spec preserves `LegalForPlayer`/`LegalForAnyone`/`LegalForPlayerError` and
adds the per-predicate `Preconditions` ledger. But it never says **what field
values the server used** when it evaluated the field-dependent bucket for the
ledger. Today's `LegalForAnyone` (main.go:1622) exists precisely to answer
"could *anyone* legally make this move" vs "can *you* right now" — a move can be
illegal-for-you-now but legal-in-principle.

The declarative ledger has the same fork but sharper: a field-dependent
predicate (`RevealableCardAt` on `move.CardIndex`) can only be evaluated by the
server against **default field values** (`DefaultsForState`). So the ledger's
`verdict:"fail"` for such a predicate means "fails **with the default
CardIndex**" — which is nearly meaningless to show the user, because the user
will pick a *different* CardIndex. The client must not gray out the move on the
basis of a field-dependent predicate that failed against server defaults; it
should show "depends on your input" and re-evaluate locally once the user binds
the field.

**Fix — the ledger must tag each precondition with its bucket and, for
field-dependent ones, mark the verdict as provisional:**

```jsonc
{"name": "revealableCardAt", "verdict": "fail", "fieldDependent": true,
 "provisional": true,   // evaluated against DefaultsForState, not user input
 "evaluable": true,     // client CAN re-run once fields bound
 "message": {...}}
```

- `fieldIndependent` predicates: verdict is authoritative for this viewer/state.
  A Fail here ⇒ gray out the move unconditionally (it's your-turn/phase/
  resource gating; no field can rescue it). **This is the true
  `LegalForAnyone:false` analog** — a field-independent Fail means "no field
  binding can make this legal."
- `fieldDependent` predicates: verdict is `provisional`; client re-runs the
  evaluable ones locally as the user edits. The move is grayed **only** when a
  field-independent predicate fails OR all field-dependent predicates are
  inevaluable-and-failing.

This gives the client the exact "fail-now vs. could-pass" distinction that
`LegalForAnyone` gave, at per-predicate granularity, and prevents the nasty UX
of a move greyed out because of a default `CardIndex` the user never chose.

### I4. #693 leak — shipping predicate args is fine, but shipping the *ledger* can leak, and computed-property routing is the sharp corner

The #693 constraint: move structs can carry hidden info. The question posed —
"does shipping predicate args (which name move *fields*) ever leak anything?" —
splits into two:

- **Args naming move fields: safe.** `args:["move.CardIndex"]` names a *field*,
  not a *value*. Field names are already public (the move form ships
  `Fields[].Name` — main.go:91). No leak from the `Spec` itself.

- **The ledger's `bindings` and `verdict` CAN leak.** This is the real hole. A
  `Message.Bindings` like `{"left": 0}` embeds a *state-derived value*. If that
  value is hidden from the viewer, shipping it in the ledger leaks it. Concretely:
  a `playerPropAtLeast("player.SecretCount", 1)` that fails against another
  player's hidden `SecretCount` would ship `bindings:{"actual": 3}` — leaking a
  hidden count. The evaluability rule (I1) *gates local re-evaluation* but the
  spec **still ships the server's verdict and message for inevaluable
  predicates** ("displays the server's last verdict for everything else"). That
  displayed message's bindings are the leak vector.

  **Fix:** for any predicate that is **not** evaluable for a viewer (its reads
  aren't sufficiently visible), the server must ship the `verdict` as
  `"unknown"` and **omit `message.bindings`** (or omit the message entirely).
  You cannot show a viewer a failure message computed from state they're not
  allowed to see. Tie this to I1: `evaluable:false` ⇒ ledger entry carries no
  state-derived bindings. The spec's own honesty principle ("It never guesses
  whether a zero is a real zero or a hidden seven") demands this — but the
  current ledger design would happily *ship the seven* in the bindings.

- **Computed-property routing (anti-tarpit rule 3) is the subtlest leak.** The
  spec encourages pushing branchy logic into a computed `PolicyVisible` property
  that a `propCompare` then reads. If that computed property is derived from
  hidden inputs, it's a leg(ergo intended) **declassification channel** — the
  author chose to expose the derived bit. That's acceptable *by author intent*,
  but the spec should name it: **computed properties are a declassification
  surface; a predicate reading one is evaluable/shippable exactly to the extent
  the computed property's own sanitization policy allows, and the author is
  responsible for that policy not leaking.** Add a §10 risk bullet: "computed
  properties used to sidestep the catalog can declassify hidden state into the
  ledger; their sanitization policy is the guard, and it is by-convention."

---

## NIT

### N1. Ledger wire example shows only flat leaves
The §6 `"Preconditions"` array shows three flat entries. Given I2's `any`
analysis and the `sub` wire format, show at least one nested compositor entry
with per-child `evaluable`, so consumers don't assume the ledger is flat.

### N2. `Reason` on Unknown is a raw string ("reads hidden property HiddenCards")
Everything else in the design is a template key + bindings for
localizability/greppability. `Verdict.Reason` breaks that discipline — it's an
ad-hoc English string, the exact anti-pattern §Problem point 3 rails against.
Make it a `*Message` too (or drop it from the client ledger; it's a
server-debug affordance and can stay server-side only). At minimum it should
never reach the client (it can name a hidden property — a mild #693 whisper).

### N3. "per viewer" ledger cost
The ledger is per-viewer (I2). The move-forms double pass (player + admin) is
mentioned, but the ledger multiplies by viewer count for spectators/companion
mode. The field-independent memo keyed `(moveType, stateVersion, proposer)` does
*not* key on viewer, yet evaluability is per-viewer. Note that the **verdict**
memo is viewer-independent (evaluation ignores sanitization; it runs on full
state server-side) but the **evaluable flag** is viewer-dependent — so cache the
verdict once, compute the cheap `evaluable` overlay per viewer. Worth one
sentence so an implementer doesn't accidentally key the whole ledger by viewer.

---

## Verdict

**Approve the architecture; block on the client contract.** The engine spine,
anti-tarpit discipline, and "purely sugar" guarantee are sound and well
defended. But the spec's marquee client-and-explainability promise is not
shippable as written: it has no template table (which breaks even the *server's*
`Verdict.Error()`), no Go↔TS conformance mechanism, and an evaluability rule
that contradicts its own headline example. These are representation-level holes —
cheap to close now, expensive after the catalog grows. None require abandoning
the design; each has a concrete, existing-primitive-shaped fix above.

### Top 3 changes

1. **Ship the template table in the chest** (B1). New
   `ConfigureLegalTemplates() map[string]string`, merged with
   `legal.DefaultTemplates()`, added to `ComponentChest.MarshalJSON` alongside
   Enums, validated at `NewGameManager`. Without this, `Verdict.Error()` renders
   nothing — a v1 server bug, not a future-client gap.

2. **Make a shared JSON conformance corpus a representation deliverable** (B2).
   `legal/conformance/*.json` consumed by Go tests now and the TS evaluator
   later, plus a `catalogVersion` stamp for graceful skew degradation. This is
   the only thing that makes "the client re-runs it identically" true rather
   than hoped.

3. **Replace "needs Visible" with per-read facet-needs
   (occupancy/count/values)** (I1), and make the ledger honest under
   sanitization: `evaluable:false` ⇒ verdict `unknown` with **no state-derived
   bindings** (I4), and tag field-dependent verdicts `provisional` so the client
   distinguishes "fails now" from "depends on your input" (I3).
