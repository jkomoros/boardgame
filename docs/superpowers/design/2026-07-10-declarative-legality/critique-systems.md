# Systems Critique: Declarative Move Legality (Designs A / B / C)

Adversarial read against the brief's four systems questions:

1. Does the composition model survive the `Default → CurrentPlayer → game-move`
   embedding chain without surprising override semantics?
2. Is the serialization future-proof for a TS evaluator under sanitization
   (three-valued honesty)?
3. Do the engine wins (Q7) *follow from the representation*, or are they merely
   asserted?
4. Where does each design secretly become a Turing tarpit or a config-soup?

All three converge on the same skeleton — named+args serializable predicate,
registry rhyming with `constraints`, one imperative escape hatch, phase index +
read-path dirty tracking, three-valued client honesty. That convergence is
itself a finding: **the representation is nearly forced**, so the designs should
be judged on the parts they *don't* share — override semantics, the composition
seam into the *existing* embedding chain, and whether the caching claims are
actually sound rather than plausible.

---

## The shared load-bearing lie: "accumulate down the embedding chain"

All three assert that `Default → CurrentPlayer → game-move` composition becomes
*declarative accumulation* — each layer appends named predicates to a config
bag, and the imperative `if err := m.CurrentPlayer.Legal(...)` boilerplate
"vanishes." **None of the three proves this survives contact with the actual
mechanism**, and it is the single most important systems question in the brief.

The problem: today the chain is a Go embedding chain resolved by method dispatch
at *call* time. `moves.CurrentPlayer` *embeds* `moves.Default`; a game move
*embeds* `moves.CurrentPlayer`. The config bag (`CustomConfiguration()`) is a
**flat `PropertyCollection` keyed by string**, populated by `auto.Config`
options that *do not know what layer they came from*. When `WithLegalPhases` and
`WithPreconditions(...)` both write into the same bag, there is exactly one
`configPropPreconditions` slice — there is no "Default's slice ∪ CurrentPlayer's
slice ∪ own slice" because **the layers share one bag**. So "accumulate in
embedding order" is not free; it requires each embedded type's *constructor* (or
a codegen pass) to inject its own predicates into that shared slice at the right
position, and to do so *before* the author's `WithPreconditions` runs so ordering
is base-first. Design B is the only one that even gestures at this ("the engine
collects the full ordered list declaratively by walking the config bags that
`auto.Config` merged") — but `auto.Config` does not merge per-layer bags; it runs
a flat option list. **The seam where CurrentPlayer contributes
`proposer_is_current` "automatically" is undesigned in all three.** Whoever
implements this discovers that either (a) embedded moves must register predicates
in an `init`/constructor hook the option system then prepends to, or (b) codegen
must synthesize the accumulation — neither of which any design specifies.

This is not a nitpick: it is *the* override-semantics question. If accumulation
is "flat append into a shared bag," then `WithoutPrecondition("proposer_is_current")`
in a game move must find and remove an entry it never saw added, matched by name.
That works only if the base layers' contributions are named and injected
*deterministically* — which reintroduces exactly the ordering problem above.

Grading below weights this: the design that is most honest and concrete about
this seam wins credibility even if its answer is imperfect.

---

# Design A — Composability: One Precondition Algebra

## Best idea worth stealing

**The unification of `SerializedCond` as simultaneously leaf-record and AST-node**
(§1: *"A combinator is just an atom whose args are other `Cond`s (`Sub`). So
`SerializedCond` is simultaneously the name+args record (leaf) and the AST node
(branch). One format, no duplication."*). This is genuinely elegant and it is the
one place A earns its "composability" title: there is a single wire format, a
single registry, and `All`/`Any`/`Not` need no special serialization path. B and
C both bolt composition on as "one allowed level of compositor" and then spend a
risk bullet apologizing for it; A makes composition *the same thing as a leaf*,
which is the correct systems move if you are going to have composition at all.
Steal this regardless of who wins.

## Worst flaw

**A talks itself into the tarpit it claims to avoid, one atom at a time — and the
`ElseWhen` construct is the smoking gun.** §9a:

> `StackHasComponentAt("game.HiddenCards", "move.CardIndex").`
> `    ElseWhen("game.VisibleCards", HasComponentAt,`
> `        "that card has already been revealed").`
> `    Else("there is no card at that index"),`

A's own risk #2 confesses this: *"`ElseWhen` (the memory nil-vs-revealed branch)
is a mini conditional. It's the thin end of a wedge toward if/else in the DSL."*
This is A convicting itself. The whole thesis is "relations over paths, no
user-authored control flow" — but the *very first acid test* needs branch
selection ("if source empty AND dest full → message X, else message Y"), and A's
answer is to invent an inline conditional operator on the atom. Combined with
`ForEachActivePlayer` (a quantifier, risk #1) and `RepeatFromState` (a dynamic
count), A has, by the end of its own document, introduced: composition (`All`/
`Any`/`Not`), inline conditionals (`ElseWhen`), bounded quantification
(`ForEachActivePlayer`), and dynamic value resolution (`RepeatFromState`,
`StackValueAtKey`). That is not a bounded relation catalog — **that is a small
expression language with three-valued semantics**, which is precisely a Turing
tarpit's on-ramp. A is the *most* likely of the three to sprawl, because its
organizing principle ("everything is one algebra") actively rewards adding one
more operator rather than dropping to the escape hatch. B and C's clumsier "one
compositor level, else `LegalCustom`" is *architecturally* more tarpit-resistant
than A's elegant-but-open algebra, even though it reads worse.

Secondary: A's `Trit` ordering (`Unknown=0, Pass, Fail`) makes `Unknown` the zero
value (§1 comment: *"must be zero value: an un-evaluated Cond is Unknown"*),
which is defensible, but then §5 asserts `Fail` dominates `Unknown` dominates
`Pass` under `All` — a Kleene-strong three-valued `AND` where `Fail` is
absorbing. That is *correct* for the client-honesty goal (you can be confidently
illegal even with a hidden input), but A never states the dual for `Any`, and the
`Not` case is left entirely unspecified. Three-valued `Not(Unknown)=Unknown` is
the only sound choice, but with `All`/`Any`/`Not` all first-class and serialized,
a TS evaluator must reimplement Kleene logic *exactly* or client and server
disagree on `Unknown`-adjacent verdicts. A's "conformance corpus" (risk #5) is
hand-waved as "not built here."

## Scores

- Elegance: **9** — the leaf≡node unification is the best single idea in the field.
- Explainability: **6** — the `Eval` tree is genuinely structured and the #65 win
  is real, but `ElseWhen`/nested combinators make a failure a *tree to interpret*,
  not a sentence; the author must reason about three-valued short-circuit to
  predict which message a player sees.
- Engine-leverage: **7** — same phase-index + path-index story as B/C, correctly
  derived; but A never separates field-independent evaluation into a cached bucket
  the way B does, and `CustomLegal`'s "unknown pathIndex ⇒ always re-eval" is the
  only invalidation rule stated. The `FieldDependent()` auto-derivation (risk #4)
  is the right instinct but unbuilt.
- Migration-credibility: **5** — the acid tests *are* shown verbatim, but 9a needs
  a newly-invented conditional operator, 9b needs a new quantifier atom, and 9c
  needs a new domain atom — three new primitives to migrate three moves is the
  opposite of "the catalog already covers it." "0 lines of Go" is achieved by
  moving the complexity into DSL surface area.

---

# Design B — Engine & Performance

## Best idea worth stealing

**The write-set instrumentation on the mutable-state setter path, tied to the
*same* namespaced path vocabulary as `Refs.StatePaths`** (§7c: *"wrap the mutable
state handed to `Apply` so that each `SetXxxProp` / stack mutation records its
namespaced path into a `writeSet`... invalidate a cached plan-result only if
`plan.allStatePaths ∩ writeSet ≠ ∅`"*). This is the only design that makes the
dirty-tracking win *mechanically fall out of the representation* rather than
asserting it. Because the read-vocabulary (`Refs.StatePaths`) and the
write-vocabulary (instrumented setters) are **the same namespace by
construction**, the intersection test is sound rather than hopeful. A and C both
claim "the engine already produces state deltas / knows what changed" — B is the
only one that says *where the write-set comes from and why it lines up with the
read-set*. Steal this; it is the actual engineering.

## Worst flaw

**B's headline caching win is unsound as specified, and B's own risk bullet
quietly admits the correctness hole while the body sells it as a win.** §7c body:

> *"a different move type whose preconditions read only `Game.Deck` keeps its
> cached verdict across that transition. Instead of 'every state change
> invalidates every move's Legal,' we get 'a state change invalidates only the
> plans that read what changed.'"*

Then risk #1:

> *"Dirty-tracking (§7c) assumes we can cheaply and completely record the paths an
> `Apply` mutates... timers, dynamic component values, and behaviors-driven
> mutations must all funnel through instrumented setters or the cache goes
> stale-and-wrong... This is the riskiest load-bearing claim."*

A cache that can go **"stale-and-wrong"** is not a performance optimization, it is
a correctness bug that produces illegal moves being accepted or legal moves being
rejected. "Stale-and-wrong" in a *legality* cache means the engine applies a move
it should have rejected. B correctly identifies this as the riskiest claim but
still puts it in the summary table as a clean win ("invalidate only
path-overlapping plans"). The honest version is: this win requires an exhaustive
audit that *every* mutation in core funnels through an instrumented setter,
including component moves that touch two stacks, timer ticks, and behavior-driven
writes — and a single un-instrumented path silently corrupts legality. The
conservative fallback B offers ("if unsure, invalidate everything this version")
is correct but *erases the entire win* for exactly the games complex enough to
need it. So B's central complexity-shift table has a footgun: the row "verdict
after a state change: invalidate only path-overlapping plans" is only true if you
win a completeness argument B hasn't made.

Compounding this: **B's `Players[current]` resolution is a real soundness hole B
half-notices.** Risk #2: *"`Players[current]` in a ref-set expands to `Players[*]`
for invalidation purposes (coarser but sound)."* But the current player can change
mid-transition (turn advance), and a predicate whose `Refs` said
`Players[current].CardsLeftToReveal` was *evaluated* against player 2 but is
*invalidated* against `Players[*]` — the read-set used for evaluation and the
read-set used for invalidation are different sets. B picks the sound-but-coarse
option, which is correct, but it means the field-independent memo (§7b) keyed by
`(moveType, stateVersion, proposer)` and the path-invalidation (§7c) disagree
about what "current player" means. Two caches with two different notions of
current-player is a bug generator.

Secondary: B's field-independent memo (§7b) admits its own key is nearly useless
in the hot path — *"the fixup poll (up to 256 recursions, but each is a new
version, so cache is really within one poll's candidate set)."* If every fixup
recursion is a new state version, the `stateVersion`-keyed cache **almost never
hits across the 256-deep loop** — it only helps within a single candidate set
where the same move type is listed twice. B buries the fact that its most-hyped
cache barely fires on the exact loop (#640) the brief calls out.

## Scores

- Elegance: **6** — the plan/bucket/cache machinery is coherent but heavy;
  `PreconditionPlan`, `fieldIndepCache`, write-set instrumentation, `Cost` enum,
  `Refs` with sentinels — this is an *engine*, not an *algebra*. Powerful, not
  delightful. The brief's quality bar is "elegant and delightful"; B optimizes a
  different objective.
- Explainability: **7** — `PreconditionFailure{Predicate, Args, MessageTemplate,
  Bindings, ReadPaths}` is the most operationally complete failure record of the
  three (it carries `ReadPaths` for #65 debugging), and the "log which predicate
  blocked the fixup without re-running" is concrete. Loses a point because message
  *selection* within a multi-branch predicate (`revealableCardAt`) is hidden
  inside Go, so the explainability is only as good as the predicate author.
- Engine-leverage: **8** — highest of the three on raw mechanism, and the *only*
  one to correctly separate field-independent memoization from path-invalidation
  and to source the write-set. Docked two points because the headline win is
  gated on an un-won completeness argument and the memo barely hits the #640 loop.
- Migration-credibility: **7** — acid tests shown verbatim with real reader calls
  (`ImmutableStackProp`, `MaySwapComponentsByKey`); `revealableCardAt` honestly
  keeps the two-branch nil check *inside one predicate's Go* rather than inventing
  a DSL conditional (the correct call, and the direct rebuttal to A's `ElseWhen`).
  The `LegalCustom` interface coexistence is the cleanest of the three. Loses
  credibility only on the `Cost`-reordering-changes-messages risk (#6), a real
  behavior change to migrated games.

---

# Design C — Delight (DX & Explainability)

## Best idea worth stealing

**Message templates as `(template-key, bindings)` pairs decoupled from
predicates, rendered by one `legal.Render` function, shipped to the client as
data** (§4). C is the only design that fully separates *the failure* from *the
prose*: a predicate emits `FailT("reveal.no_cards_left", B{"left": have})` and
never formats a string. This is what actually delivers three of the brief's goals
at once — localizability, client-side re-rendering without a round trip, and #65
greppability — and it does so without A's tree-interpretation burden or B's
per-predicate `MessageTemplate` duplication. The `Message{Template, Bindings}`
type is the cleanest explainability primitive in the field. Steal it wholesale;
it is orthogonal to whose composition model wins.

## Worst flaw

**C's core-layering claim is internally contradictory and, read literally,
breaks the "purely sugar / degrades gracefully" locked decision.** §3 proposes
that core `boardgame` define a *different, lossy* interface than the `legal`
package's real one:

> `type LegalResult struct { Legal bool; Unknown bool; Error error }`

and separately, §1's `Verdict` carries `Outcome`, `Message{Template, Bindings}`,
and `Reason`. So the engine (in core) sees `LegalResult` (a bool + bool + rendered
error), while the `legal` package sees `Verdict` (structured template + bindings).
**The structured message — the entire point of Design C — is flattened to a
rendered `error` string at the core boundary.** But §4 and §7 both require the
engine (which lives in core, per B's correct observation that the fixup loop is in
`game.go`) to ship `(template, bindings)` to the client and to log structured
failures for #65. C's own client payload (§4) shows `"Bindings": {"left": 0}`
reaching the client — but the core `Evaluator.Evaluate` returns `LegalResult`
whose `Error error` has *already lost the bindings*. The two interfaces C defines
cannot both be true: either core sees the structured `Verdict` (and core is no
longer minimal — it imports the template concept), or core sees `LegalResult` (and
the bindings never reach the client because the engine that builds move-forms
lives in core). C waves at this — *"Predicate/Verdict/Context/PropPath type defs?
NO — see below"* — but "see below" resolves it by making core's interface lossy,
which silently kills the design's marquee feature at exactly the layer that has to
deliver it.

This is worse than A's or B's worst flaw because it is *self-defeating*: C's
differentiator is explainability, and C's layering discards the explainability
data at the core seam where the server assembles the client payload.

Secondary — **C is the most config-soupy of the three by its own admission.** §2
offers *three* authoring surfaces (struct tags, `With*` options, fluent
`Require`/`All`/`Any`/`Not`) all "funneling to the same slice." The brief warns
against config-soup; C answers "three ways to do it, one destination" and calls it
delight. But three surfaces with different capabilities (struct tags can't nest;
`With*` is flat; `Require` can compose) means an author must learn *which* surface
a given precondition is expressible in — e.g. §9a mixes a struct tag
(`legal:"index_into(game.HiddenCards)"`) with `With*` options for the *same move*,
and §9's actual migrations then **abandon struct tags entirely** and use only
`With*`. C shows a struct-tag example in §2 that no acid test uses. That is the
definition of surface that exists to look delightful in the intro but doesn't
carry the real cases. And §9a's
`"reveal.already_revealed_if(game.VisibleCards)"` — a template key with an
*embedded conditional predicate reference* stuffed into a string argument — is C's
own `ElseWhen`, smuggled into a template name. C did not avoid A's inline
conditional; it hid it inside a stringly-typed template key, which is *worse*
because it's unvalidated until runtime (C's risk #1 admits `@Field` typos aren't
caught until `NewGameManager` at best).

## Scores

- Elegance: **7** — the `Verdict`/`Message` types are the cleanest primitives, but
  three authoring surfaces and the two-interface core/legal split (`Verdict` vs
  `LegalResult`) are a coherence tax. Delightful to read, less coherent to build.
- Explainability: **9** — best-in-field *as a spec*: template+bindings, the
  client ledger with per-precondition `verdict`/`clientEvaluable`/`reason`, and
  the #65 log line are all concrete and player-facing. This is C's real win and it
  is a big one — docked only because the core-layering flaw threatens delivery.
- Engine-leverage: **5** — C asserts the phase-index and read-path dirty-tracking
  wins (§7) in the same words as A/B but does the *least* to ground them: it says
  *"the engine already produces state deltas for storage"* and reuses that, without
  B's write-set instrumentation or any account of why the delta granularity matches
  `Reads()` granularity. And `Unknown` explicitly does **not** short-circuit (§5:
  *"we keep evaluating so the client gets the fullest ledger"*) — a deliberate
  choice for ledger richness that *costs* engine performance on exactly the hot
  path, directly trading away Q7 for Q4. C is honest about this; it's still the
  weakest engine story.
- Migration-credibility: **6** — acid tests are verbatim and the checkers `Legal()`
  tail returning a *template* error (§9.3) is a genuinely nice touch (structured
  even in the escape hatch). But 9a's `already_revealed_if(...)` string-embedded
  conditional and the struct-tag surface that no migration actually uses undercut
  the "reads like a rulebook sentence" promise on the very cases meant to prove it.

---

## Cross-cutting verdicts

**Composition seam (the brief's Q1):** None survives cleanly. B is the most honest
("walk the config bags `auto.Config` merged") but wrong about the mechanism
(there is one flat bag, not per-layer bags). A and C assert accumulation "just
works" with less scrutiny. **Best of a bad lot: B**, because it at least names the
override operation (`WithoutPrecondition` + suppression entry) and the ordering
rule concretely.

**Serialization / TS honesty (Q2):** A's leaf≡node format is the most future-proof
*shape*; C's template+bindings is the most future-proof *message* payload; B's
flat `[]{Name,Args}` is the most future-proof *and* the easiest to reimplement in
TS (no Kleene-`Not` to mirror). A's first-class `Not` is a hidden TS-conformance
liability none of the others carry.

**Engine wins follow from representation? (Q3):** Only **B** derives them
(write-set ↔ read-set same namespace). A and C *assert* them in nearly identical
language. C actively spends engine performance (non-short-circuiting `Unknown`) to
buy explainability.

**Tarpit / soup (Q4):** A is the tarpit risk (open algebra rewards new operators;
`ElseWhen` + quantifier + dynamic count already in v1). C is the soup risk (three
authoring surfaces; stringly-typed conditional template keys). B is the most
disciplined ("exactly one compositor level, else `LegalCustom`") but pays for it
in inelegance and a caching correctness hole.

---

## Scores at a glance

| Design | Elegance | Explainability | Engine-leverage | Migration-credibility |
|---|---|---|---|---|
| A — Composability | 9 | 6 | 7 | 5 |
| B — Engine | 6 | 7 | 8 | 7 |
| C — Delight | 7 | 9 | 5 | 6 |

**Overall ranking (one sentence):** No design ships as-is — **B** has the only
sound engine mechanism and the most honest composition seam and should be the
spine, but it must steal A's leaf≡node serialization format and C's
template+bindings message model to clear the brief's "elegant, delightful,
explainable" bar, so the winning design is B's engine wearing C's error model and
A's wire format.
