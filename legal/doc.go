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
A move opts in per-type by passing `moves.WithLegalPreconditions(...)` to
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
	    moves.WithLegalPreconditions(
	        legal.PropAtLeast("player.CardsLeftToReveal", 1).
	            WithMessage("reveal.no_cards_left"),
	        legal.RevealableCardAt("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
	        legal.MayMoveToSameSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
	    ),
	)

Each argument to WithLegalPreconditions is a Spec built by one of this package's
catalog functions (PropAtLeast, PropCompare, PlayerBool, PlayerBoolIs,
PlayerBoolAt, PlayerHasSubmitted, PlayerHasNotSubmitted, PlayerIsActive,
PlayerIsInactive, PlayerSeatIsFilled, PlayerSeatIsClosed, PlayerIsAdmin,
StackCount, StackEmpty, StackNotEmpty, PropEquals, PropNotEquals,
ComponentPresentAt, ComponentAbsentAt, ComponentPresentAtKey, MayMoveTo,
MayMoveToSlot, MayMoveToSameSlot, MayMoveCountTo, MayMoveAllTo, MaySwapComponents,
MaySwapComponentsByKey, Any, AllActivePlayers, RevealableCardAt,
ComponentPropEqualsCurrentPlayer, ProposerIsCurrentPlayer, InPhase,
StackConstraints — the full, current list is DefaultConstructors()).
At NewGameManager, every declared Spec is resolved through the registry,
every path it references is validated (a typo is a boot error naming the
move and the path — never a mid-game surprise), and one ordered plan is
assembled per opted-in move type: that move type's own *contributed* specs
first (derived automatically from moves.Default/CurrentPlayer's existing
configuration — inPhase, inProgression, stackConstraints, and, for
CurrentPlayer, proposerIsCurrentPlayer), then the authored WithLegalPreconditions
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
    PlayerBool are the general-purpose relations; ComponentPresentAt and the
    MayMove and MaySwap predicates are the stack-shaped ones. Prefer these first.

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

`any` (Any, in this package) is the only compositor, registry-enforced to
depth 1 — no nested `any`, no first-class `not` wrapper around an arbitrary
sub-spec. There is deliberately no `all`: the plan's ordered list of
top-level specs already is the conjunction. The completeness round (design
spec 2026-07-12) added explicit NEGATION LEAVES instead of a general `not`
compositor — PlayerBoolIs(prop, false), PropNotEquals, and ComponentAbsentAt
each invert one specific relation, covering the common single-property
negation without a general wrapper. A disjunction-of-conjunctions shape
((A∧B)∨(C∧D)) still has no expression (no nested `any`, no `all`) and stays
LegalCustom. Framework move types whose semantics are inherently negated or
conditional (ApplyUntil and its subclasses, RoundRobin) stay opaque rather
than bending this rule.

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
template. Implementing LegalCustom automatically opts a move on a supported
base into declarative legality; a custom-only move needs no constructor marker.
LegalCustom combined with a wholesale Legal() override on the
same type is a boot error too: the override would orphan both the declared
plan and the residue.

WithoutLegalPrecondition(name moves.PreconditionName) suppresses one *contributed* check by its
stable name (moves.Default/CurrentPlayer's own, never a game-authored one):
pass the exported constants moves.PreconditionInPhase,
moves.PreconditionInProgression, moves.PreconditionStackConstraints, or
moves.PreconditionProposerIsCurrentPlayer rather than raw strings. Use it
when a move needs to opt out of something it would otherwise inherit — the
moves.ForceFinishTurn "inherit nothing" pattern, now expressible without
its own bespoke base type. It does not remove an authored WithLegalPreconditions
spec; those are simply not passed in the first place if you don't want
them.

WithoutLegalPrecondition itself opts the move in, so a suppression can never
be dead configuration merely because there are no authored specs. The boot
gauntlet still enforces two important errors,
each naming the offending move:

  - An unmatched name: the suppression names no spec the move actually
    contributes. This covers both a typo ("inphase" — pass the constants
    and typos become compile errors instead) and suppressing a check the
    move never had (moves.PreconditionInProgression on a move with no
    configured move progression, or
    moves.PreconditionProposerIsCurrentPlayer on a moves.Default-embedding
    move, which never contributes that atom). The error lists the move's
    real contributed spec names.

  - moves.PreconditionProposerIsCurrentPlayer on a
    moves.CurrentPlayer-embedding move: CurrentPlayer.Legal() runs its
    proposer-equivalence check imperatively, unconditionally, after its
    super-call into Default.Legal, so suppressing the contributed atom
    would silently desync the plan/ledger (which would then say "legal")
    from Legal() itself (which still rejects a wrong proposer). If a
    CurrentPlayer-based move genuinely needs no proposer check, embed
    moves.Default instead.

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

validated at NewGameManager: every template key referenced by a declared
Spec (a WithMessage override) or listed in a predicate's EmittedTemplates
metadata (which every catalog constructor populates with the keys its
Evaluate can FailT with) must resolve to registered text — an unregistered
key there is a boot error naming the move, never a runtime surprise. That
guarantee is scoped to declared Specs and EmittedTemplates: a FailT or
Errorf key born INSIDE arbitrary Go — a LegalCustom body, or an emission
path a game-registered predicate didn't list in EmittedTemplates — cannot be
boot-validated (closures are not introspectable), and an unregistered key
there degrades at render time to RenderMessage's bare-key fallback instead.
Use Spec.WithMessage("your.key") on any catalog builder's return value to
point that one predicate's failure at a specific key instead of its
predicate-level default; see memory's
`PropAtLeast(...).WithMessage("reveal.no_cards_left")` for a real example.
RenderMessage never panics on a missing binding — it renders the raw
placeholder name instead, so a template/binding mismatch is visibly wrong in
output rather than crashing a request. For every catalog predicate (and any
game-registered predicate declaring EmittedBindings metadata — see below),
that mismatch cannot survive to runtime at all: NewGameManager also
validates each resolved template BODY, requiring its {placeholders} to be a
subset of the bindings the owning predicate declares it emits with that key.
This covers a WithMessage retarget at a game template AND a
ConfigureLegalTemplates body override of a catalog default — e.g. pointing
PropAtLeast at a body referencing {frobs} is a boot error naming the move,
the template key, and the unemitted placeholder (each catalog predicate's
bindings per key are documented on its Template* constant). One subtlety for
predicates that fail on MORE THAN ONE branch (e.g. mayMoveTo/mayMoveToSlot:
{index} when no component sits at the source, {detail} when the move itself
is rejected): the per-branch bindings documented on the Template* constants
apply to the DEFAULT per-branch keys, but a single WithMessage override
retargets EVERY branch at your one key, so the bindings your template body
may reference shrink to the branches' INTERSECTION — only a binding attached
on every failure path is guaranteed renderable. For mayMoveTo/mayMoveToSlot
that intersection is empty, so an overriding body may not reference any
placeholder at all; for proposerIsCurrentPlayer both branches emit {detail},
so {detail} survives the collapse.

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

Return only your additions or intentional overrides; the framework overlays
them on the full built-in catalog automatically. A delegate that does not
implement the interface keeps the defaults unchanged. checkers'
`checkers.spaceIsBlack` is the reference
example (examples/checkers/main.go's ConfigurePredicateConstructors): a
board-geometry check with no natural fit in a game-agnostic catalog, but
which is still a genuine relation over one path (a space index), so it
becomes a small, named, serializable game-registered predicate rather than
LegalCustom residue.

Two hard requirements on a game-registered Evaluate, and the consequences of
breaking them:

Declared Reads must cover every path Evaluate touches — over-approximate,
never under-approximate. Reads is an honor system (a Go closure cannot be
introspected), and the failure mode of UNDER-declaring a move.* (or
players[move.*]) read is much worse than a client-side evaluability skew:
plan assembly sorts a predicate with no declared move reads into the
field-independent bucket, whose verdict is memoized keyed on (move type,
state version, proposer) — WITHOUT the move's field values — so the SERVER
ITSELF serves a stale verdict when only the move's fields change. The boot
gauntlet smoke-probes the common shape of this bug: each field-independent
game-registered predicate is evaluated once against the example state with a
sentinel move whose property reader panics on any access, and touching the
move's properties is a boot error naming the move and the predicate. Do not
lean on the probe, though — it cannot see a move read that is conditional on
state the example state doesn't exhibit, or one made through a concrete type
assertion on ctx.Move rather than ctx.ResolvePath/ctx.Move.Reader(), so an
honest Reads declaration remains YOUR responsibility.

For every declared read whose evaluator expects a particular property shape,
also declare RequiredReadTypes. NewGameManager then rejects a property that
exists under the right name but has the wrong PropertyType, instead of letting
the predicate degrade to Unknown whenever it runs. If (and only if) the
predicate genuinely accepts several property types, use AllowedReadTypes with
that explicit set; PropEquals is the catalog example. Type metadata may only
name paths present in Reads, and a path cannot appear in both maps. The built-in
catalog declares a type contract for every read, so literal path typos and type
mismatches are caught during manager construction. `boardgame-util lint`
preserves that authoritative check and, when a built-in constructor's path is
a string literal with one unambiguous source occurrence, reports the failure at
the literal's file and line. Computed or ambiguous paths retain the full boot
error without a potentially misleading source guess.

Evaluate must be a pure function of its LegalContext: no time, no
randomness, no I/O, no mutation, nothing outside ctx. Verdicts are memoized
and replayed (the same field-independent memo above) and are assumed
reproducible everywhere a verdict crosses a boundary (ledger, logs, a future
client evaluator); an impure Evaluate makes the cached verdict silently
wrong with no error anywhere.

Alongside
EmittedTemplates (the keys your Evaluate can FailT with, boot-validated
against the game's template table), a game-registered predicate SHOULD also
populate EmittedBindings: a map from each emitted template key to the
binding names Evaluate is guaranteed to attach with it (nil/empty for a
bindings-free emission). With that metadata declared, NewGameManager
validates each key's resolved template body's {placeholders} against it —
the same placeholder/binding boot check every catalog predicate gets (see
"Template tables" above), and checkers' spaceIsBlack demonstrates the
pattern. The metadata is optional for backward compatibility: a predicate
with a nil EmittedBindings map skips this validation entirely, and a
template/binding mismatch then only shows up as a bare placeholder name in
rendered output. A game-registered predicate defaults to evaluable:false even
when serializable; serialization does not imply a generic client knows its
semantics. The server remains the source of truth, exactly like CustomLegaler.

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

# v4 limits (read honestly, not as marketing)

v1 shipped counts as an unused facet (FacetCount existed, no predicate read
it), no typed equality, no move-field-indexed player paths, and only the
bare playerBool/ComponentPresentAt negation-free leaves. The completeness
round (design spec 2026-07-12) closed those four gaps — StackCount/
StackEmpty/StackNotEmpty, PropEquals/PropNotEquals, the
players[move.<Field>].<Prop> path kind, and PlayerBoolIs/ComponentAbsentAt,
respectively — and widened the composition seam. What's left, honestly:

  - **Current player and proposer are deliberately distinct.** `player.X`
    resolves against CurrentPlayerIndex; `proposer.X` resolves against the
    concrete proposing player and therefore works during simultaneous play.
    Prefer CurrentPlayer(), Proposer(), or PlayerFromMove("Field") rather
    than assembling behavior paths by hand. Proposer-read client
    evaluability is conservative across every player because the public
    evaluability API receives a viewer, not a trusted proposer. These helpers
    return a typed PlayerSelector backed by one internal player-index-source
    resolver. Semantic behavior conditions on Proposer carry explicit admin
    bypass policy; Observer and Any are never valid actors.
  - **AllActivePlayers' inner leaf only accepts int/bool-typed properties.**
    PlayerBool/PlayerBoolIs, player-path PropAtLeast/PropCompare, or an Any
    of those — never an enum- or PlayerIndex-typed property, even though
    PropEquals now supports both at the top level. A per-player quantifier
    over "has everyone voted" (PlayerIndex-typed) or "everyone selected
    class X" (enum-typed) has no expression yet. Found migrating a
    werewolf-shaped game; not yet closed.
  - **Negation is explicit-leaf-only, not general.** PlayerBoolIs(prop,
    false), PropNotEquals, and ComponentAbsentAt cover single-property
    negation; there is still no general `not` wrapper and no `all`
    compositor, so a disjunction-of-conjunctions shape ((A∧B)∨(C∧D)) has no
    expression and stays LegalCustom.
  - **DynamicComponentValues have no path grammar equivalent.** A check
    reading a component's per-game dynamic values (as opposed to its static
    chest-defined Values()) — a card's current face-up type, a species'
    population counter — always needs LegalCustom. Across every game
    surveyed so far, this is the single most common reason a real move
    stays partially opaque.
  - **The composition seam is moves.Default, moves.CurrentPlayer,
    moves.FixUp, moves.FixUpMulti, and moves.StartPhase.** All five declare
    no Legal() override of their own (a source-parse test enforces this
    invariant, so a future override is a boot-red test forcing a conscious
    seam decision). Every other framework move type in package moves —
    DealCountComponents, FinishTurn, RoundRobin, and the rest — is opaque:
    it does not implement the contribution interface, so a move embedding
    one of them and declaring WithLegalPreconditions fails at boot, naming the
    unsupported base type. FinishTurn/DealCountComponents specifically stay
    blocked because a partial contribution could desync the ledger;
    round-robin/progression-aware predicates are future work.
  - **MayMoveTo takes one source index; MayMoveToSlot takes distinct source
    and destination fields.** Use MayMoveToSameSlot for mirrored layouts.
  - **An unknown enum value name (PropEquals/PropNotEquals) is a
    LegalUnknown at evaluate time, not a boot-time construction error.** A
    constructor-time typo guard catches the common case when a chest is
    available at construction, but the wider "unknown enum name = boot
    error" design aspiration isn't fully delivered — closing it for real
    needs either a wider constructor signature (example state reachable at
    construction) or a dedicated boot-validation hook, neither of which
    exists yet.
  - **Memoization does not reorder checks.** Field-independent verdicts are
    cached per predicate, but the evaluator and ledger always traverse the
    contributed/authored declaration order. A later state-only check can
    never jump ahead of an earlier move-field check merely because it is
    cacheable.

None of these are dead ends — LegalCustom always works, and each gap above
is exactly the shape of thing rule 1-3 above is designed to grow the catalog
into, one purpose-built predicate at a time, as real games need it.
*/
package legal
