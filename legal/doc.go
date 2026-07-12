/*
Package legal is the predicate catalog and registry for declarative move
legality — a peer package to constraints, following the same shape (name +
string args, constructor registry). It aliases boardgame's Legal*-prefixed
value types (Outcome, BindingValue, Message, Verdict, Cost, Facet, PropPath,
Read, Spec, Context, Predicate, PredicateConstructor) so game authors write
legal.Spec rather than boardgame.LegalSpec. Core (package boardgame) owns the
underlying types and the evaluation engine — plan assembly, the boot-time
probe, phase bucketing, memoization, the move-form ledger; this package owns
only the catalog: predicate builders, their Evaluate implementations, and
default template text. See
docs/superpowers/specs/2026-07-10-declarative-legality-design.md for the full
design; this doc comment is the day-to-day authoring guide.

# The purely-sugar guarantee

Read this before anything else: declarative legality changes nothing for a
move that doesn't opt in. `Legal(state, proposer) error` remains the
ground-truth contract. `moves.Default.Legal()`, `moves.CurrentPlayer.Legal()`,
and every other framework move type's `Legal()` keep their existing
implementations, byte-for-byte: same checks, same order, same error strings.
A move opts in per-type by passing `moves.WithPreconditions(...)` to
`auto.Config`/`auto.MustConfig` AND not wholesale-overriding `Legal()` (it may
still implement CustomLegaler, below). For an opted-in move,
`moves.Default.Legal()` detects the declared plan at call time and evaluates
it instead of running the old imperative chain. Everything else — moves
without declarations — is opaque: the engine calls their Legal() exactly as
today, with zero behavior change and zero migration required. There is no
capability in this package that requires a game to adopt it.

# Quick start

	auto.MustConfig(
	    new(moveRevealCard),
	    moves.WithPreconditions(
	        legal.PropAtLeast("player.CardsLeftToReveal", 1).
	            WithMessage("reveal.no_cards_left"),
	        legal.RevealableCardAt("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
	        legal.MayMoveToSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
	    ),
	)

Each argument to WithPreconditions is a Spec built by one of this package's
catalog functions (PropAtLeast, PropCompare, PlayerBool, ComponentPresentAt,
ComponentPresentAtKey, MayMoveTo, MayMoveToSlot, Any, AllActivePlayers,
RevealableCardAt, ComponentPropEqualsCurrentPlayer, ProposerIsCurrentPlayer,
InPhase, StackConstraints — the full, current list is DefaultConstructors()).
At NewGameManager, every declared Spec is resolved through the registry,
every path it references is validated (a typo is a boot error naming the
move and the path — never a mid-game surprise), and one ordered plan is
assembled per opted-in move type: that move type's own *contributed* specs
first (derived automatically from moves.Default/CurrentPlayer's existing
configuration — inPhase, inProgression, stackConstraints, and, for
CurrentPlayer, proposerIsCurrentPlayer), then the authored WithPreconditions
specs in declaration order. Evaluation short-circuits on the first Fail, in
plan order — no cost-based reordering in v1, so migrated moves keep their
historical first-failure message (see "Bucket reordering" below for the one
documented exception).

See the tutorial's "Declarative Move Legality" section (TUTORIAL.md) for a
complete worked before/after using memory's real moveRevealCard migration,
and examples/checkers, examples/blackjack, examples/pig for further real
migrations at varying levels of catalog coverage.

# The catalog's three rules of growth

This catalog is deliberately small, and staying small is a design goal, not
a temporary state. Before adding a new catalog predicate, or reaching for one
that doesn't exist:

 1. **If you can express the check as a relation over one or more property
    paths, it can be a catalog predicate.** PropAtLeast, PropCompare, and
    PlayerBool are the general-purpose relations; ComponentPresentAt and
    MayMoveTo/MayMoveToSlot are the stack-shaped ones. Prefer these first.

 2. **Branchy logic becomes a purpose-built named predicate with a
    hand-written Evaluate, not new DSL surface.** RevealableCardAt is the
    model: a 12-line Go function disambiguating "no card here" from
    "already revealed" by occupancy alone. If your check genuinely has two
    or three distinct failure branches over a small, fixed set of paths,
    write a small predicate like this one (see catalog_purpose.go) rather
    than composing primitives cleverly or petitioning for a new compositor.

 3. **No user arithmetic, loops, or lambdas in serialized form, ever.** If a
    check needs computation beyond a relation or a short hand-written branch
    — summing a hand's value, walking a graph, comparing two runtime-chosen
    component values — it does not belong in the catalog. Push the
    computation into a computed state property that a relation predicate can
    then read, or use the escape hatch (LegalCustom, below). This rule is
    what keeps the catalog conformance-corpus-checkable and, eventually,
    portable to a client-side TypeScript evaluator without ever needing to
    port a general-purpose expression language.

`any` (Any, in this package) is the only compositor in v1, registry-enforced
to depth 1 — no nested `any`, no first-class `not`. There is deliberately no
`all`: the plan's ordered list of top-level specs already is the conjunction.
Framework move types whose semantics are inherently negated or conditional
(ApplyUntil and its subclasses, RoundRobin) stay opaque in v1 rather than
bending this rule.

# The escape hatch

Two knobs handle everything the catalog rules above say does NOT belong in
serialized form:

CustomLegaler is an optional interface a move type may implement:

	type CustomLegaler interface {
	    // Runs after every declared precondition has passed.
	    LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error
	}

It runs last in the plan, always, after every declarative check has passed.
It is never cached, has no serialized form, and a client sees its verdict as
"unknown" — the honest answer, since only the server can evaluate arbitrary
Go. Return legal.Errorf(templateKey, bindings) to keep the failure
structured (template key + bindings, renderable and greppable exactly like a
declarative Fail); a plain error still works and is wrapped as a one-off
template. A move implementing LegalCustom must have opted in via
WithPreconditions with at least one real spec — "declaring is implementing";
an empty WithPreconditions() call does not count, so LegalCustom alone,
without any declarative gates, is never consulted. LegalCustom combined with
a wholesale Legal() override on the same type is a boot error: the override
would orphan both the declared plan and the residue.

WithoutPrecondition(name string) suppresses one *contributed* check by its
stable name (moves.Default/CurrentPlayer's own, never a game-authored one):
"inPhase", "inProgression", "stackConstraints", "proposerIsCurrentPlayer".
Use it when a move needs to opt out of something it would otherwise inherit
— the moves.ForceFinishTurn "inherit nothing" pattern, now expressible
without its own bespoke base type. It does not remove an authored
WithPreconditions spec; those are simply not passed in the first place if
you don't want them.

"proposerIsCurrentPlayer" is suppressible ONLY on a moves.Default-embedding
move (where it does nothing anyway, since Default never contributes it in
the first place — the entry is symmetry, not function). On a
moves.CurrentPlayer-embedding move it is a boot error: CurrentPlayer.Legal()
runs its proposer-equivalence check imperatively, unconditionally, after its
super-call into Default.Legal, so suppressing the contributed atom would
silently desync the plan/ledger (which would then say "legal") from Legal()
itself (which still rejects a wrong proposer). If a CurrentPlayer-based move
genuinely needs no proposer check, embed moves.Default instead.

A `Legal()` override that super-calls into the frozen chain (the existing,
universal embedding pattern — `if err := m.CurrentPlayer.Legal(...); err !=
nil { ... }`) is fully compatible: the super-call evaluates the plan for an
opted-in move, and the override's own additional imperative checks compose
around it exactly as they always have. Wholesale overriding Legal() without
super-calling, on a move type with declared preconditions, is a boot-time
error (caught by a one-time probe at NewGameManager) — the declarations
would otherwise be silently dead code, and this design never allows a
declaration to be silently ignored.

# Template tables

Every declarative failure carries a Message{Template, Bindings} — a template
KEY plus named bindings, never a pre-baked string — so failures stay
localizable, greppable, and re-renderable anywhere a Verdict crosses a
boundary (server logs, the fixup rejection log, the move-form ledger, and
eventually a client renderer). DefaultTemplates() ships default English text
for every key the built-in catalog predicates default to
(DefaultTemplateKeys() is the authoritative list, cross-checked by this
package's own tests against DefaultTemplates()' coverage). A game extends or
overrides the table via an optional delegate method:

	// package legal — implemented optionally by delegates; base.GameDelegate
	// does NOT need changes. Absence = DefaultTemplates() only.
	type TemplateConfigurer interface {
	    ConfigureLegalTemplates() map[string]string
	}

validated at NewGameManager: every template key referenced by any declared
Spec, FailT call, or Errorf call must resolve to registered text — an
unregistered key is a boot error naming the move, never a runtime surprise.
Use Spec.WithMessage("your.key") on any catalog builder's return value to
point that one predicate's failure at a specific key instead of its
predicate-level default; see memory's
`PropAtLeast(...).WithMessage("reveal.no_cards_left")` for a real example.
RenderMessage never panics on a missing binding — it renders the raw
placeholder name instead, so a template/binding mismatch is visibly wrong in
output rather than crashing a request.

# Game-registered predicates

The registry games consume by default (DefaultConstructors()) is open to
extension, exactly like constraints.StackConstraintConstructor. A game
registers its own predicates via an optional delegate interface, consumed by
type-assertion at NewGameManager (never a new required GameDelegate method —
that would be a compile break for every existing delegate that doesn't
implement it):

	// package legal — implemented optionally by delegates.
	type ConstructorConfigurer interface {
	    ConfigurePredicateConstructors() []*PredicateConstructor
	}

Return ExtendDefaults(yourConstructor, ...) to keep the full built-in catalog
plus your addition; return DefaultConstructors() (or the built-ins directly)
to decline extension. checkers' `checkers.spaceIsBlack` is the reference
example (examples/checkers/main.go's ConfigurePredicateConstructors): a
board-geometry check with no natural fit in a game-agnostic catalog, but
which is still a genuine relation over one path (a space index), so it
becomes a small, named, serializable game-registered predicate rather than
LegalCustom residue. A game-registered predicate's declared Reads is honored
by convention — there is no framework-level verification that a
game-supplied Evaluate func only touches what it declares; keep Reads
conservative (over-approximate rather than under-approximate) so client
evaluability calculations (see below) stay honest. A game-registered
predicate degrades gracefully on a client with no TypeScript implementation
of it: its ledger entry reports evaluable:false and the server remains the
source of truth, exactly like CustomLegaler.

# What client evaluation gets, today

Every declarative failure's Reads (a conservative read-set, keyed to the
minimal facet actually touched — FacetValues, FacetCount, FacetOccupancy, or
FacetOrder) is what makes the move-form ledger's per-predicate `evaluable`
flag precise under sanitization: a stack-occupancy check needs only
FacetOccupancy, which survives more sanitization policies than FacetValues
would. This is designed for, and currently consumed by, only the server: the
server evaluates every predicate and ships the resulting ledger (verdict +
evaluable + provisional + message, per predicate) to the client alongside
the unchanged LegalForPlayer/LegalForPlayerError/LegalForAnyone fields.
There is no client-side (TypeScript) evaluator yet; that's a designed-for
follow-up the wire format and Reads/Facet machinery already anticipate.

# v1 limits (read honestly, not as marketing)

  - **`player.X` paths resolve against the game's CurrentPlayerIndex, not the
    proposing player.** In a simultaneous-move phase — every player
    proposing at once, the game's own notion of "current player" being
    Admin or none — `player.X` cannot express "the player who proposed this
    move"; it returns an error or Unknown instead. This blocked a real
    migration in a downstream game with a simultaneous-move phase, which was
    reverted specifically for this reason rather than shipped incorrect.
  - **No count/stack-size predicate.** There is no catalog predicate for
    "this stack has at least N components" (Stack.NumComponents() >= n),
    even though the Read machinery already has a FacetCount facet designed
    for exactly this purpose — no predicate uses it yet. Several real
    migrations across example and downstream games were blocked on this and
    stayed LegalCustom.
  - **No negation.** `any` is the only compositor; there is no `not`. A
    negated condition needs a purpose-built predicate (rule 2 above) or
    LegalCustom.
  - **The composition seam is moves.Default and moves.CurrentPlayer only.**
    Every other framework move type in package moves — FixUp, StartPhase,
    DealCountComponents, FinishTurn, RoundRobin, and the rest — is opaque in
    v1: it does not implement the contribution interface, so a move
    embedding one of them and declaring WithPreconditions fails at boot,
    naming the unsupported base type. Extending the seam is one-at-a-time
    follow-up work, each requiring its own golden-equivalence tests.
  - **MayMoveTo/MayMoveToSlot take a single index field**, used for both the
    source lookup and (for MayMoveToSlot) the destination slot. There is no
    variant that names two different indices.
  - **Bucket reordering narrows the "historical first-failure message"
    claim.** The plan evaluator splits a move's plan into a field-independent
    bucket (no move.* reads) and a field-dependent bucket (>=1 move.* read),
    evaluating the entire field-independent bucket before any
    field-dependent predicate — regardless of declaration order. Since
    proposerIsCurrentPlayer reads move.TargetPlayerIndex (field-dependent),
    a field-independent authored check that would have run AFTER the
    proposer check in the old linear chain can now report its failure
    FIRST, for inputs that fail both simultaneously. Real migrations
    documented and test-asserted this divergence rather than hiding it (see
    each migrated game's legal_golden_test.go and this repo's
    knownMessageOrderingDivergence-style maps).

None of these are dead ends — LegalCustom always works, and each gap above
is exactly the shape of thing rule 1-3 above is designed to grow the catalog
into, one purpose-built predicate at a time, as real games need it.
*/
package legal
