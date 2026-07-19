package boardgame

import (
	stderrors "errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jkomoros/boardgame/enum"
)

/*
This file is the declarative-legality evaluation engine (design spec §4) and
its boot-time assembly. It is the campaign's highest-risk surface: it must
deliver the "prime guarantee" that a move which does NOT declare
WithLegalPreconditions behaves byte-for-byte as it does today (the frozen chain in
moves/default.go runs unchanged), while a move that DOES declare them has its
imperative chain replaced by plan evaluation — with the switch detected
behaviorally at boot by a probe, never statically.

Layering: everything here lives in core because game.go's loops (and, later,
the move-forms ledger) call it, and because it must reach into unexported
manager state (the probe flags, the per-move plan map). The universal
predicate catalog and default templates it needs are injected from package
moves' init() (see legal_registry.go); a game's own predicates/templates ride
the per-delegate optional interfaces, consumed by type-assertion below.
*/

// CustomLegaler is the escape hatch (design spec §4): a move type opted in to
// declarative legality may implement it to run imperative "residue" AFTER all
// declared preconditions pass. It is wrapped as an opaque, unserializable
// predicate that always evaluates LAST in the plan (a client sees it as
// "unknown"). A move that both implements LegalCustom AND wholesale-overrides
// Legal() is a boot error (the override would orphan everything) — that is
// caught by the same probe that catches orphaned declarations.
type CustomLegaler interface {
	// LegalCustom runs after every declarative precondition has passed. A
	// non-nil return makes the move illegal; return legal.Errorf(key, binds)
	// to carry a structured, template-rendered message, or a plain error to
	// have its text used as a one-off template key.
	LegalCustom(state ImmutableState, proposer PlayerIndex) error
}

// legalContributor is the core-side structural view of
// moves.PreconditionsProvider (design spec §3: "optional interfaces consumed
// by type-assertion"). moves.Default/CurrentPlayer implement it; core never
// names package moves. Its []LegalSpec return is identical to package moves'
// []legal.Spec (legal.Spec is a type alias for boardgame.LegalSpec).
type legalContributor interface {
	ContributedPreconditions() []LegalSpec
}

// legalDeclarer is the core-side structural view of the DeclaredPreconditions
// half of the same moves-package surface: authored WithLegalPreconditions specs
// (in declaration order) and WithoutLegalPrecondition suppression names.
// Opt-in is reported separately because an empty WithLegalPreconditions call,
// a suppression, or LegalCustom can intentionally assemble a zero-authored plan.
type legalDeclarer interface {
	DeclaredPreconditions() ([]LegalSpec, []string)
}

type legalExplicitEnabler interface {
	LegalPlanEnabled() bool
}

// legalConstructorConfigurer is the core-side structural view of
// legal.ConstructorConfigurer (design spec §1): an optional delegate
// interface supplying game-registered predicate constructors, overlaid on the
// universal defaults.
type legalConstructorConfigurer interface {
	ConfigurePredicateConstructors() []*LegalPredicateConstructor
}

// legalTemplateConfigurer is the core-side structural view of
// legal.TemplateConfigurer (design spec §6): an optional delegate interface
// supplying game template-key overrides/additions, overlaid on the universal
// default templates.
type legalTemplateConfigurer interface {
	ConfigureLegalTemplates() map[string]string
}

// legalPlan is the per-opted-in-move-type evaluation plan, built once at
// NewGameManager (design spec §4). ordered preserves contributed/authored
// declaration order. The two buckets remain as indexes for memo/probe work,
// but never determine observable evaluation or ledger order.
type legalPlan struct {
	// moveName is the owning move type's name (mType.Name()), used as part
	// of the field-independent memo's key (legal_memo.go) so a lookup can't
	// collide across move types sharing a game.
	moveName string
	// fieldIndependent indexes resolved predicates with no move.* reads for
	// memoization and boot probes; it does not determine evaluation order.
	fieldIndependent []*LegalPredicate
	// fieldDependent indexes resolved predicates that read at least one
	// move.* path (proposerIsCurrentPlayer among them — it reads
	// move.TargetPlayerIndex); it does not determine evaluation order.
	fieldDependent []*LegalPredicate
	// ordered is the authoritative evaluation order. Each independent step
	// carries its stable memo ordinal, allowing its individual verdict to be
	// cached without hoisting it ahead of earlier dependent predicates.
	ordered []legalPlanStep
	// custom is the CustomLegaler escape-hatch wrapper (opaque,
	// unserializable, CostExpensive), evaluated dead last, or nil if the
	// move type does not implement CustomLegaler.
	custom *LegalPredicate
	// allReads is the de-duplicated union of every constituent predicate's
	// Reads, for client-evaluability/dirty-tracking consumers.
	allReads []LegalRead
	// specs is the serializable subset of the plan (every non-opaque
	// predicate's originating spec shape), preserved for the ledger and
	// wire format.
	specs []LegalSpec
}

type legalPlanStep struct {
	predicate      *LegalPredicate
	fieldDependent bool
	memoIndex      int
}

// LegalVerdictEntry is one predicate's line in a full-ledger evaluation
// (design spec §6's move-forms "Preconditions" array). It is the shape Task
// 10's server ledger and the future client evaluator consume. Serializable
// and Reads let the server decide per-viewer evaluability; FieldDependent
// marks a verdict as provisional (computed against server-chosen move field
// defaults).
type LegalVerdictEntry struct {
	// Name is the predicate's registry name ("any" for a compositor,
	// "custom" for the escape-hatch wrapper).
	Name string
	// Args are the predicate's string args (nil for compositors/custom).
	Args []string
	// Verdict is this predicate's evaluated verdict.
	Verdict LegalVerdict
	// Serializable reports whether this predicate has a wire form (false for
	// the custom wrapper and any predicate containing one).
	Serializable bool
	// ClientEvaluable reports that the generic client catalog knows this
	// predicate's semantics. It is independent of viewer-specific visibility.
	ClientEvaluable bool
	// FieldDependent reports whether this predicate reads any move.* path
	// (its verdict is provisional on the server-chosen move fields).
	FieldDependent bool
	// Reads is this predicate's declared read-set (nil for the opaque
	// custom wrapper).
	Reads []LegalRead
}

// evaluate runs the plan against ctx in contributed/authored declaration
// order, followed by the custom tail (design spec §4).
//
// In hot-path mode (fullLedger == false) it short-circuits on the FIRST
// non-Pass verdict (Fail OR Unknown — fail-closed: an Unknown means legality
// could not be confirmed, so the move is treated as not legal and its verdict
// returned, exactly as a Fail would be), preserving today's first-failure
// error precedence; the returned entry slice is nil. In this mode each
// field-independent predicate's verdict is served from (and stored into)
// ctx.State.Game()'s memo (design spec §5's honest table, legal_memo.go)
// without moving that predicate from its declared position. A ctx with no
// Game to anchor a memo to
// (e.g. a probe or an isolated test evaluating against ExampleState()) just
// evaluates fresh every time, with no memo interaction at all.
//
// In full-ledger mode (fullLedger == true) it evaluates EVERY predicate,
// returns the parallel []LegalVerdictEntry for the ledger, and returns as the
// overall verdict the first non-Pass encountered (or Pass if all passed).
// Every predicate is run through evalLegalPredicate, so a panicking or
// invalid-verdict predicate degrades to a fail-closed Unknown, never to Pass.
// Full-ledger mode never consults or populates the memo: it produces fresh
// per-predicate entries in declaration order.
func (p *legalPlan) evaluate(ctx LegalContext, fullLedger bool) (LegalVerdict, []LegalVerdictEntry) {
	if p == nil {
		// A nil plan cannot confirm legality; fail closed.
		return LegalVerdict{Outcome: LegalUnknown, Reason: "nil legal plan"}, nil
	}

	overall := LegalVerdict{Outcome: LegalPass}
	haveOverall := false

	var entries []LegalVerdictEntry

	consider := func(step legalPlanStep) bool {
		pred := step.predicate
		v := p.evaluateStep(ctx, step, fullLedger)
		if fullLedger {
			entries = append(entries, LegalVerdictEntry{
				Name:            pred.Name,
				Args:            pred.Args,
				Verdict:         v,
				Serializable:    pred.Serializable(),
				ClientEvaluable: pred.ClientEvaluable,
				FieldDependent:  step.fieldDependent,
				Reads:           pred.Reads,
			})
		}
		if v.Outcome != LegalPass && !haveOverall {
			overall = v
			haveOverall = true
			if !fullLedger {
				return true // short-circuit
			}
		}
		return false
	}

	steps := p.ordered
	if len(steps) == 0 {
		// Compatibility for low-level tests and internally hand-built plans.
		for i, pred := range p.fieldIndependent {
			steps = append(steps, legalPlanStep{predicate: pred, memoIndex: i})
		}
		for _, pred := range p.fieldDependent {
			steps = append(steps, legalPlanStep{predicate: pred, fieldDependent: true, memoIndex: -1})
		}
	}
	for _, step := range steps {
		if consider(step) {
			return overall, nil
		}
	}
	if p.custom != nil {
		if consider(legalPlanStep{predicate: p.custom, fieldDependent: true, memoIndex: -1}) {
			return overall, nil
		}
	}

	return overall, entries
}

func (p *legalPlan) evaluateStep(ctx LegalContext, step legalPlanStep, fullLedger bool) LegalVerdict {
	if fullLedger || step.fieldDependent {
		return evalLegalPredicate(step.predicate, ctx)
	}
	game := legalMemoGame(ctx)
	if game == nil {
		return evalLegalPredicate(step.predicate, ctx)
	}
	key := legalFieldIndepMemoKey{
		moveName:  p.moveName,
		version:   ctx.State.Version(),
		proposer:  ctx.ProposerPlayerIndex,
		predicate: step.memoIndex,
	}
	if verdict, ok := game.legalFieldIndepMemoGet(key); ok {
		return verdict
	}
	verdict := evalLegalPredicate(step.predicate, ctx)
	game.legalFieldIndepMemoSet(key, verdict)
	return verdict
}

// legalMemoGame returns ctx.State.Game(), or nil if ctx.State itself is nil
// (a nil ImmutableState has no Game() to call). Game() itself may also
// legitimately return nil (e.g. GameManager.ExampleState()'s state has no
// backing *Game) — callers treat either case as "nothing to memoize
// against".
func legalMemoGame(ctx LegalContext) *Game {
	if ctx.State == nil {
		return nil
	}
	return ctx.State.Game()
}

// LegalProbeActive is engine-internal plumbing for the declarative-legality
// boot probe (design spec "prime guarantee" rule 4). moves.Default.Legal
// calls it as the VERY FIRST thing in its body: if it returns true, a boot
// probe is in progress, and Default.Legal must immediately return nil — the
// probe only needs to OBSERVE that execution reached Default.Legal (a
// wholesale Legal() override that never super-calls never reaches it, so its
// declarations are orphaned and boot fails; a super-calling override reaches
// it and is fully supported). This is set/read only during NewGameManager's
// single-threaded boot, before the timer goroutine starts and before any game
// runs, so the flags are stable-false at runtime and this is race-free (the
// only writes happen-before NewGameManager returns). No caller other than the
// frozen framework chain should ever call it.
func (g *GameManager) LegalProbeActive() bool {
	if g.legalProbing {
		g.legalProbeReached = true
		return true
	}
	return false
}

// LegalPlanAssembled reports whether moveName has an assembled declarative
// plan. Framework move types use it after a Default.Legal super-call to avoid
// re-running legality that the plan already contributed and evaluated.
func (g *GameManager) LegalPlanAssembled(moveName string) bool {
	return g != nil && g.legalPlans != nil && g.legalPlans[moveName] != nil
}

// LegalEvaluatePlan is engine-internal plumbing the frozen framework chain
// calls (moves.Default.Legal) to run the opt-in plan path. If the move type
// named moveName is opted in (a plan was assembled for it at boot), it
// evaluates that plan against (state, move, proposer) in hot-path
// short-circuit mode and returns (true, err) where err is the first
// failure's rendered error (nil if the plan passes), already carrying the
// game's template table. Otherwise it returns (false, nil) and the caller
// runs its frozen imperative chain unchanged — this is the seam that keeps
// non-opted-in moves byte-for-byte identical.
func (g *GameManager) LegalEvaluatePlan(moveName string, state ImmutableState, move Move, proposer PlayerIndex) (bool, error) {
	if g.legalPlans == nil {
		return false, nil
	}
	plan := g.legalPlans[moveName]
	if plan == nil {
		return false, nil
	}

	ctx := LegalContext{
		State:               state,
		Move:                move,
		ProposerPlayerIndex: proposer,
		Chest:               g.chest,
	}
	verdict, _ := plan.evaluate(ctx, false)
	err := verdict.Error()
	if le, ok := err.(*LegalError); ok {
		return true, le.AttachTable(g.legalTemplateTable)
	}
	return true, err
}

// LegalEvaluateLedger is engine-internal plumbing the server ledger (design
// spec §6, Task 10) calls to get every predicate's individual verdict for an
// opted-in move type in ONE evaluation pass, replacing the two separate
// Legal() calls (player-perspective + admin-structural) the pre-Task-10
// server made. If moveName is opted in (a plan was assembled for it at
// boot), it runs full-ledger evaluation against (state, move, proposer) and
// returns (verdict, entries, true). verdict is computed with the exact same
// first-non-Pass semantics LegalEvaluatePlan's hot-path short-circuit mode
// uses (see legalPlan.evaluate's doc comment: fullLedger mode still latches
// onto the FIRST non-Pass verdict encountered, in plan order, even though it
// keeps evaluating every remaining predicate for the ledger) — so rendering
// verdict through LegalRenderVerdict reproduces move.Legal()'s error text
// byte-for-byte for the same (state, move, proposer). Otherwise (moveName is
// opaque) it returns (LegalVerdict{}, nil, false) and the caller must fall
// back to the move's imperative Legal() path — the frozen two-call path Task
// 10 leaves untouched for non-opted-in moves.
func (g *GameManager) LegalEvaluateLedger(moveName string, state ImmutableState, move Move, proposer PlayerIndex) (LegalVerdict, []LegalVerdictEntry, bool) {
	if g.legalPlans == nil {
		return LegalVerdict{}, nil, false
	}
	plan := g.legalPlans[moveName]
	if plan == nil {
		return LegalVerdict{}, nil, false
	}

	ctx := LegalContext{
		State:               state,
		Move:                move,
		ProposerPlayerIndex: proposer,
		Chest:               g.chest,
	}
	verdict, entries := plan.evaluate(ctx, true)
	return verdict, entries, true
}

// LegalRenderVerdict renders v into the same error text move.Legal() itself
// would produce for the equivalent verdict: v.Error() (LegalError, for any
// non-Pass Outcome) attached to this manager's own merged legal template
// table, exactly as LegalEvaluatePlan attaches it before returning. Returns
// "" for a Pass verdict (v.Error() is a true nil interface — see
// LegalVerdict.Error()'s doc comment — so there is nothing to render). This
// is what lets the server ledger derive LegalForPlayerError/LegalForAnyone's
// rendered text from a LegalEvaluateLedger verdict without this package
// exposing legalTemplateTable itself.
func (g *GameManager) LegalRenderVerdict(v LegalVerdict) string {
	err := v.Error()
	if err == nil {
		return ""
	}
	if le, ok := err.(*LegalError); ok {
		return le.AttachTable(g.legalTemplateTable).Error()
	}
	return err.Error()
}

// legalCustomUnsupportedBaseError reports a LegalCustom method that cannot be
// attached to a declarative plan because the move does not embed a supported
// moves-package base. LegalCustom automatically opts supported moves in.
func legalCustomUnsupportedBaseError(moveName string) error {
	return fmt.Errorf("move %q implements CustomLegaler (LegalCustom), which automatically opts into declarative legality, but its base type does not support declarative legality (only moves.Default, moves.CurrentPlayer, moves.RecordCurrentPlayerChoice, moves.FixUp, moves.FixUpMulti, and moves.StartPhase do); switch to one of those base types or move the LegalCustom logic into a Legal() override", moveName)
}

// assembleLegalPlans is called once at the end of NewGameManager (after moves
// are installed and the example state exists). For every installed move type
// that has opted in to declarative legality (declares WithLegalPreconditions or
// explicitly enables a zero-authored-spec plan), it
// verifies the move is on a supported base (design spec §5's seam:
// legalSupportedMovesBaseTypes — Default, CurrentPlayer,
// RecordCurrentPlayerChoice, FixUp, FixUpMulti, StartPhase), assembles and
// validates its plan, stores it, and probes that
// the declarations are actually reachable. Any failure is a boot error naming
// the offending move (and, for the seam check, the unsupported base type). A
// move with neither form of opt-in is left entirely alone: no plan, no probe
// — its frozen chain runs at runtime exactly as today.
func (g *GameManager) assembleLegalPlans(exampleState ImmutableState) error {
	registry, templateTable, gameRegistered := g.buildLegalRegistryAndTemplates()
	g.legalTemplateTable = templateTable

	for _, mType := range g.moves {
		move := mType.NewMove(exampleState)
		if move == nil {
			// A half-functioning move type; other boot checks report it.
			continue
		}

		declarer, ok := move.(legalDeclarer)
		if !ok {
			// Not opt-in-capable: this move embeds no moves-package base that
			// provides the DeclaredPreconditions surface, so it has no plan and
			// its frozen chain runs at runtime. But if it implements
			// CustomLegaler, that residue can never be wrapped into a plan
			// (footgun-batch F5) — the identical silent fail-open as the
			// no-authored-specs path below — so reject it here too, before this
			// guard would skip the move.
			if _, isCustom := move.(CustomLegaler); isCustom {
				return legalCustomUnsupportedBaseError(mType.Name())
			}
			continue
		}
		authored, suppressions := declarer.DeclaredPreconditions()
		_, hasCustom := move.(CustomLegaler)
		explicitlyEnabled := false
		if enabler, ok := move.(legalExplicitEnabler); ok {
			explicitlyEnabled = enabler.LegalPlanEnabled()
		}
		if len(authored) == 0 && !explicitlyEnabled && !hasCustom {
			// Not opted in. The frozen imperative chain runs unchanged.
			continue
		}

		// Seam (design spec §5): a move embedding any moves-package framework
		// type outside legalSupportedMovesBaseTypes cannot opt in — its
		// imperative Legal() would interleave wrongly with plan evaluation.
		if base := legalUnsupportedMovesBaseType(move); base != "" {
			return fmt.Errorf("move %q declares preconditions but embeds unsupported framework move type %q: only moves.Default, moves.CurrentPlayer, moves.RecordCurrentPlayerChoice, moves.FixUp, moves.FixUpMulti, and moves.StartPhase support declarative legality (the seam allowlist is legalSupportedMovesBaseTypes in legal_plan.go, enforced structurally by moves/seam_source_test.go — widening it requires that type to declare no Legal() override of its own)", mType.Name(), base)
		}

		var contributed []LegalSpec
		if contributor, ok := move.(legalContributor); ok {
			contributed = contributor.ContributedPreconditions()
		}

		// Footgun-batch F2 flavor 1: every suppression must match at least
		// one contributed spec name, or it is either a typo ("inphase") or
		// an opt-out of a check this move never had ("inProgression" on a
		// move with no progression) — both silently no-ops before this
		// guard. The CurrentPlayer-specific proposer guard above stays the
		// more specific error for its case (on a CurrentPlayer-embedding
		// move the proposer atom IS contributed, so it would pass this
		// check).
		if err := validateLegalSuppressions(mType.Name(), contributed, suppressions); err != nil {
			return err
		}

		specs := assembleLegalSpecList(contributed, authored, suppressions)

		predicates, err := resolveLegalSpecs(specs, registry, g.chest, exampleState, move.Reader())
		if err != nil {
			return fmt.Errorf("move %q: %w", mType.Name(), err)
		}

		if err := validateLegalTemplates(specs, predicates, templateTable); err != nil {
			return fmt.Errorf("move %q: %w", mType.Name(), err)
		}
		if err := validateLegalEmittedTemplates(predicates, templateTable); err != nil {
			return fmt.Errorf("move %q: %w", mType.Name(), err)
		}
		// Footgun-batch F4: with every emitted key known to resolve, also
		// check that each resolved template BODY's {placeholders} are a
		// subset of the bindings its predicate declares it emits with that
		// key — covering WithMessage retargets and ConfigureLegalTemplates
		// body overrides alike. Predicates without EmittedBindings metadata
		// (game-registered ones predating the field) are skipped; see
		// validateLegalEmittedBindings.
		if err := validateLegalEmittedBindings(predicates, templateTable); err != nil {
			return fmt.Errorf("move %q: %w", mType.Name(), err)
		}

		plan := buildLegalPlanFromPredicates(mType.Name(), predicates, specs, move)
		if g.legalPlans == nil {
			g.legalPlans = make(map[string]*legalPlan)
		}
		g.legalPlans[mType.Name()] = plan

		// Footgun-batch F3: game-registered predicates' declared Reads are
		// an honor system, and an UNDER-declared move.* read is worse than a
		// client-side inaccuracy — it lands the predicate in the
		// field-independent bucket, whose verdict is memoized without the
		// move's fields in the key, so the SERVER itself serves stale
		// verdicts. Smoke-probe each field-independent game-registered
		// predicate once, against a sentinel move whose reader panics on any
		// property access.
		if err := legalProbeUndeclaredMoveReads(mType.Name(), plan, gameRegistered, exampleState, move, g.chest); err != nil {
			return err
		}

		if err := g.probeLegalReachable(mType, exampleState); err != nil {
			return err
		}
	}

	return nil
}

// buildLegalRegistryAndTemplates builds the constructor registry and template
// table for this game: the process-wide universal defaults (registered by
// package moves' init, see legal_registry.go) overlaid with the delegate's
// own game-registered constructors / template overrides, if it implements the
// optional legal.ConstructorConfigurer / legal.TemplateConfigurer interfaces.
//
// The third return is the set of predicate names supplied by the delegate.
// It feeds legalProbeUndeclaredMoveReads (footgun-batch F3). Universal
// constructors are audited in-repo, but every delegate-supplied constructor
// is probed, including one that deliberately overrides a universal name.
func (g *GameManager) buildLegalRegistryAndTemplates() (map[string]*LegalPredicateConstructor, map[string]string, map[string]bool) {
	registry, templates := legalRegistrySnapshot()

	var gameRegistered map[string]bool
	if cc, ok := g.delegate.(legalConstructorConfigurer); ok {
		for _, ctor := range cc.ConfigurePredicateConstructors() {
			if ctor == nil || ctor.Name == "" {
				continue
			}
			if gameRegistered == nil {
				gameRegistered = make(map[string]bool)
			}
			gameRegistered[ctor.Name] = true
			registry[ctor.Name] = ctor
		}
	}

	if tc, ok := g.delegate.(legalTemplateConfigurer); ok {
		for k, v := range tc.ConfigureLegalTemplates() {
			templates[k] = v
		}
	}

	return registry, templates, gameRegistered
}

// validateLegalSuppressions is the footgun-batch F2 flavor-1 boot check for
// one opted-in move type: every WithoutLegalPrecondition name must match at least
// one CONTRIBUTED spec name (suppression removes contributed atoms only —
// see assembleLegalSpecList below — so an unmatched name could never have
// any effect). An unmatched name is a boot error naming the move, the
// unmatched name, and the move's actual contributed names, so the author can
// tell a typo ("inphase") apart from suppressing a check the move never
// contributes ("inProgression" on a move with no configured progression).
func validateLegalSuppressions(moveName string, contributed []LegalSpec, suppressions []string) error {
	if len(suppressions) == 0 {
		return nil
	}

	names := make(map[string]bool, len(contributed))
	quoted := make([]string, 0, len(contributed))
	for _, spec := range contributed {
		if !names[spec.Name] {
			quoted = append(quoted, fmt.Sprintf("%q", spec.Name))
		}
		names[spec.Name] = true
	}

	contributedDesc := "none at all — this move contributes no specs, so no suppression could ever match"
	if len(quoted) > 0 {
		contributedDesc = strings.Join(quoted, ", ")
	}

	for _, name := range suppressions {
		if !names[name] {
			return fmt.Errorf("move %q suppresses %q via WithoutLegalPrecondition but contributes no spec with that name: suppression only removes a CONTRIBUTED atom from the plan, so an unmatched name is either a typo or an opt-out of a check this move never had (this move's contributed spec names: %s)", moveName, name, contributedDesc)
		}
	}
	return nil
}

// assembleLegalSpecList produces the final ordered spec list for a plan
// (design spec §2's "Plan assembly"): contributed atoms first (base-first,
// deterministic), then authored atoms in declaration order, MINUS any
// contributed atom whose Name is in suppressions (WithoutLegalPrecondition
// suppresses inherited contributions only — authored atoms are never
// suppressed, they are the opt-in itself).
func assembleLegalSpecList(contributed, authored []LegalSpec, suppressions []string) []LegalSpec {
	suppressed := make(map[string]bool, len(suppressions))
	for _, name := range suppressions {
		suppressed[name] = true
	}

	specs := make([]LegalSpec, 0, len(contributed)+len(authored))
	for _, spec := range contributed {
		if suppressed[spec.Name] {
			continue
		}
		specs = append(specs, spec)
	}
	specs = append(specs, authored...)
	return specs
}

// buildLegalPlanFromPredicates records resolved predicates in authoritative
// declaration order and also indexes them into field-independent/dependent
// buckets (by whether they read a move.* path) for memoization and probes. It
// appends the CustomLegaler wrapper (if move implements CustomLegaler) as the
// plan's custom tail. moveName is stored on
// the plan for the field-independent memo's key (legal_memo.go); pass "" if
// the caller doesn't care about memoization (e.g. an isolated test that
// never evaluates against a real *Game — see legalPlan.evaluate, which skips
// the memo whenever ctx.State.Game() is nil regardless of moveName). specs is
// the parallel serializable spec list.
//
// Invariant: the super-call into Default.Legal evaluates this plan exactly
// once per Legal call. Supported framework types return after that super-call;
// they do not re-run frozen imperative copies of contributed predicates.
func buildLegalPlanFromPredicates(moveName string, predicates []*LegalPredicate, specs []LegalSpec, move Move) *legalPlan {
	plan := &legalPlan{specs: specs, moveName: moveName}
	for _, pred := range predicates {
		if legalReadsIncludeMovePath(pred.Reads) {
			plan.fieldDependent = append(plan.fieldDependent, pred)
			plan.ordered = append(plan.ordered, legalPlanStep{predicate: pred, fieldDependent: true, memoIndex: -1})
		} else {
			memoIndex := len(plan.fieldIndependent)
			plan.fieldIndependent = append(plan.fieldIndependent, pred)
			plan.ordered = append(plan.ordered, legalPlanStep{predicate: pred, memoIndex: memoIndex})
		}
	}
	if _, ok := move.(CustomLegaler); ok {
		plan.custom = newLegalCustomWrapper()
	}
	plan.allReads = legalPlanAllReads(plan)
	return plan
}

// legalPlanAllReads returns the de-duplicated union of every plan predicate's
// Reads, in first-seen declaration order.
func legalPlanAllReads(plan *legalPlan) []LegalRead {
	seen := make(map[LegalRead]bool)
	var out []LegalRead
	for _, step := range plan.ordered {
		for _, r := range step.predicate.Reads {
			if !seen[r] {
				seen[r] = true
				out = append(out, r)
			}
		}
	}
	return out
}

// newLegalCustomWrapper builds the opaque escape-hatch predicate wrapping a
// move's LegalCustom (design spec §4). It is unserializable (opaque), the
// most expensive cost tier, declares no reads (unknown read-set), and its
// Evaluate calls ctx.Move.LegalCustom, translating a returned *LegalError
// back into its structured Verdict and any other error into a one-off Fail
// whose template key is the error's own text (rendered as itself when absent
// from the table).
func newLegalCustomWrapper() *LegalPredicate {
	return &LegalPredicate{
		Name:   "custom",
		Cost:   LegalCostExpensive,
		opaque: true,
		Evaluate: func(ctx LegalContext) LegalVerdict {
			cl, ok := ctx.Move.(CustomLegaler)
			if !ok {
				return LegalVerdict{Outcome: LegalUnknown, Reason: "custom: move does not implement CustomLegaler"}
			}
			err := cl.LegalCustom(ctx.State, ctx.ProposerPlayerIndex)
			if err == nil {
				return LegalVerdict{Outcome: LegalPass}
			}
			var le *LegalError
			if stderrors.As(err, &le) {
				return le.Verdict
			}
			return LegalVerdict{
				Outcome: LegalFail,
				Message: &LegalMessage{Template: err.Error()},
			}
		},
	}
}

// probeLegalReachable runs the boot probe for one opted-in move type (design
// spec "prime guarantee" rule 4). It flags the manager as probing, calls the
// example move's Legal() against the real example state (a valid state, so
// any imperative residue a super-calling override runs after the super-call
// executes harmlessly — the probe ignores the returned error), then checks
// whether execution reached moves.Default.Legal (which sets the reached flag
// via LegalProbeActive). If it did not, the move's declarations can never
// execute: boot fails naming the move.
func (g *GameManager) probeLegalReachable(mType *moveType, exampleState ImmutableState) error {
	probeMove := mType.NewMove(exampleState)
	if probeMove == nil {
		return nil
	}

	g.legalProbing = true
	g.legalProbeReached = false
	// The returned error is deliberately ignored: the probe observes
	// reachability via the flag, not the verdict.
	_ = probeMove.Legal(exampleState, ObserverPlayerIndex)
	reached := g.legalProbeReached
	g.legalProbing = false
	g.legalProbeReached = false

	if !reached {
		return fmt.Errorf("move %q declares preconditions but its Legal() override never reaches moves.Default.Legal — declarations would be dead (use LegalCustom for imperative residue, or super-call the embedded chain); put the super-call FIRST in your override — one that conditionally returns before super-calling can trip this same probe even against the always-valid example state used to run it; only moves embedding a base type from the seam allowlist (legalSupportedMovesBaseTypes in legal_plan.go: Default, CurrentPlayer, RecordCurrentPlayerChoice, FixUp, FixUpMulti, StartPhase) can opt in at all; the bases beyond the original Default/CurrentPlayer seam declare no Legal() override, so this probe should only ever fire on a move's OWN override, never on those embedded bases", mType.Name())
	}
	return nil
}

/*
Footgun-batch F3: the undeclared-move-read boot probe.

A game-registered predicate's declared Reads is an honor system (there is no
way to introspect a Go closure), and the failure mode of UNDER-declaring a
move.* read is uniquely bad: buildLegalPlanFromPredicates sorts the predicate
into the field-independent bucket, whose overall verdict is memoized keyed on
(moveName, stateVersion, proposer) — WITHOUT the move's field values
(legal_memo.go) — so the server itself serves a stale verdict when only the
move's fields change. The probe below catches the common shape of that bug
mechanically at boot: it evaluates each field-independent game-registered
predicate once, against the real example state, with a sentinel Move whose
PropertyReader panics on every property access. Reaching the move's
properties at all proves the predicate depends on the move, so the recovered
sentinel panic becomes a boot error naming the move and the predicate.

Precision properties, both load-bearing:

  - It can only fire on a genuine move-property access: the sentinel panic
    payload is an unexported type only the sentinel reader throws, and any
    OTHER recovered panic is swallowed (at runtime evalLegalPredicate
    degrades such a panic to a fail-closed Unknown, exactly as before this
    probe existed — a predicate that panics at boot for unrelated reasons
    must not newly fail boot).
  - It cannot false-negative into a behavior change: the probe only ADDS a
    boot error; evaluation semantics are untouched.

Known blind spots, accepted deliberately (see legal/doc.go's game-registered
predicates section): a predicate whose move read is conditional on state the
example state doesn't exhibit; a predicate that reaches the move via
ctx.Move.Info().ConcreteMove() or a concrete type assertion rather than
ctx.Move.Reader()/ctx.ResolvePath; and a delegate overriding a UNIVERSAL
catalog name (probing is scoped to names outside the default registry — see
buildLegalRegistryAndTemplates). Catalog predicates are skipped: their Reads
are audited in-repo, and probing them for every game would add boot cost for
no new information.
*/

// legalProbeSentinelPanic is the distinctive panic payload
// legalProbeSentinelReader throws. Unexported and thrown by nothing else, so
// recovering a value of this type is PROOF the probed predicate touched the
// sentinel move's properties.
type legalProbeSentinelPanic struct{}

// legalProbeSentinelReader implements the full PropertyReader interface;
// every method panics with legalProbeSentinelPanic. It is the Reader() of
// legalProbeSentinelMove, below.
type legalProbeSentinelReader struct{}

func (legalProbeSentinelReader) Props() map[string]PropertyType {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) IntProp(name string) (int, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) BoolProp(name string) (bool, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) StringProp(name string) (string, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) IntSliceProp(name string) ([]int, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) BoolSliceProp(name string) ([]bool, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) StringSliceProp(name string) ([]string, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) PlayerIndexSliceProp(name string) ([]PlayerIndex, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) PlayerIndexProp(name string) (PlayerIndex, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) ImmutableStackProp(name string) (ImmutableStack, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) ImmutableBoardProp(name string) (ImmutableBoard, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) ImmutableTimerProp(name string) (ImmutableTimer, error) {
	panic(legalProbeSentinelPanic{})
}
func (legalProbeSentinelReader) Prop(name string) (interface{}, error) {
	panic(legalProbeSentinelPanic{})
}

// legalProbeSentinelMove wraps a real example move (so the sentinel is
// non-nil and every Move method a defensive predicate might call behaves
// sanely), overriding only Reader() to return the panicking sentinel reader.
// ctx.ResolvePath and ctx.Move.Reader() — the sanctioned surfaces a
// predicate reads move properties through — both hit the sentinel.
type legalProbeSentinelMove struct {
	Move
}

func (m *legalProbeSentinelMove) Reader() PropertyReader {
	return legalProbeSentinelReader{}
}

// legalProbeUndeclaredMoveReads walks plan's field-independent bucket
// (recursing through compositor Sub trees, so a game-registered predicate
// inside an "any" is probed too) and evaluates each game-registered
// predicate once against exampleState with the sentinel move. A recovered
// sentinel panic is a boot error naming the move and the predicate. See the
// file-section comment above for the full rationale and known blind spots.
//
// The field-DEPENDENT bucket is deliberately not probed: those predicates
// declared a move.* read, which is exactly the honest declaration this probe
// exists to encourage; evaluating them against a panicking reader would
// reject every one of them.
func legalProbeUndeclaredMoveReads(moveName string, plan *legalPlan, gameRegistered map[string]bool,
	exampleState ImmutableState, exampleMove Move, chest *ComponentChest) error {

	if len(gameRegistered) == 0 || plan == nil {
		return nil
	}

	ctx := LegalContext{
		State:               exampleState,
		Move:                &legalProbeSentinelMove{Move: exampleMove},
		ProposerPlayerIndex: ObserverPlayerIndex,
		Chest:               chest,
	}

	var walk func(pred *LegalPredicate) error
	walk = func(pred *LegalPredicate) error {
		if pred == nil {
			return nil
		}
		if gameRegistered[pred.Name] && pred.Evaluate != nil && legalProbeEvaluateTouchesMove(pred, ctx) {
			return fmt.Errorf("move %q: game-registered predicate %q reads properties off the move at evaluation time but declares no move.* (or players[move.*]) path in Reads: plan assembly sorted it into the field-independent bucket, whose verdict is memoized WITHOUT the move's field values in its key, so the server itself would serve stale verdicts when only the move's fields change — declare every move path the predicate reads in its Reads (over-approximating is fine; see legal/doc.go's game-registered predicates section)", moveName, pred.Name)
		}
		for _, sub := range pred.Sub {
			if err := walk(sub); err != nil {
				return err
			}
		}
		return nil
	}

	for _, pred := range plan.fieldIndependent {
		if err := walk(pred); err != nil {
			return err
		}
	}
	return nil
}

// legalProbeEvaluateTouchesMove runs pred.Evaluate DIRECTLY (not through
// evalLegalPredicate, whose recover would silently convert the sentinel
// panic to an Unknown verdict and hide the evidence) and reports whether it
// panicked with the sentinel payload. Any other panic — or any verdict at
// all — reports false: the probe's only question is "did evaluation reach
// the move's properties?", never "did evaluation succeed?".
func legalProbeEvaluateTouchesMove(pred *LegalPredicate, ctx LegalContext) (touched bool) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(legalProbeSentinelPanic); ok {
				touched = true
				return
			}
			// Unrelated panic: swallowed. At runtime evalLegalPredicate
			// degrades it to a fail-closed Unknown; a predicate that panics
			// against the example state must not NEWLY fail boot just
			// because this probe ran it.
		}
	}()
	pred.Evaluate(ctx)
	return false
}

// legalMovesPackagePathSuffix is the import-path suffix identifying the
// framework moves package. Matches moves/default.go's own convention for
// recognizing moves-package types by reflection.
const legalMovesPackagePathSuffix = "boardgame/moves"

// legalSupportedMovesBaseTypes is the design spec §5 seam allowlist: the
// complete set of framework moves-package base types a move may embed and
// still opt in to declarative legality. Default and CurrentPlayer are the
// original v1 seam (design spec §2) — they declare their own Legal()
// overrides, verified equivalent to plan evaluation by
// TestLegalChainStringFreeze and the CurrentPlayer opt-in tests
// (moves/legal_plan_test.go). RecordCurrentPlayerChoice inherits that exact
// CurrentPlayer legality and declares no Legal method of its own. FixUp,
// FixUpMulti, and StartPhase were added by
// Task 6 (design spec §5): none of the three declares its own Legal()
// override — their legality IS Default.Legal, verbatim, so plan evaluation
// composes exactly as it does for a bare Default-embedding move. This is
// enforced structurally, not just asserted here: moves/seam_source_test.go
// parses every moves/*.go with go/parser and fails if any type in this set
// beyond Default/CurrentPlayer is ever given its own Legal() method — a
// future override on one of these types must flip that test red, forcing a
// conscious seam decision (widen deliberately, or don't) rather than silently
// starting to interleave imperative and declarative evaluation. FixUpMulti
// specifically required its own proof before joining this set: its
// AllowMultipleInProgression() override changes move-progression matching,
// and moves/preconditions_test.go's TestFixUpMultiProgressionAtomEquivalence
// proves — empirically, against a real repeated-move tape, not just by
// reading the source — that the "inProgression" declarative atom
// (moves/catalog_framework.go) and the frozen chain agree on every
// occurrence, including repeats, because they call the exact same
// moves.Default.legalMoveInProgression method.
var legalSupportedMovesBaseTypes = map[string]bool{
	"Default":                   true,
	"CurrentPlayer":             true,
	"RecordCurrentPlayerChoice": true,
	"FixUp":                     true,
	"FixUpMulti":                true,
	"StartPhase":                true,
}

// LegalSupportedMovesBaseTypeNames is engine-internal plumbing exposing the
// design spec §5 seam allowlist (legalSupportedMovesBaseTypes, above) BEYOND
// the original v1 seam (Default, CurrentPlayer): the framework moves-package
// base types that (a) are safe to embed and still opt in to declarative
// legality, and (b) must NOT declare their own Legal() method, because their
// legality IS moves.Default.Legal, verbatim (see legalSupportedMovesBaseTypes'
// own doc comment). It exists so moves/seam_source_test.go's structural
// "these types declare no Legal()" check has a single source of truth to
// consume instead of a hand-copied duplicate of this set that could
// silently drift out of sync with the real allowlist when it widens. No
// caller other than that test should need this. The names are returned in
// no particular order.
func LegalSupportedMovesBaseTypeNames() []string {
	names := make([]string, 0, len(legalSupportedMovesBaseTypes))
	for name := range legalSupportedMovesBaseTypes {
		if name == "Default" || name == "CurrentPlayer" {
			continue
		}
		names = append(names, name)
	}
	return names
}

// legalUnsupportedMovesBaseType walks move's anonymous-embed graph looking
// for an embedded struct type from the framework moves package that is not
// in legalSupportedMovesBaseTypes (design spec §5's seam). It returns the
// name of the first such type found, or "" if the move embeds only supported
// framework types. Detection is by reflection over the embed graph (the
// spec-offered alternative to a marker-method chain) because it is robust to
// deep embedding and needs no cooperation from the framework types
// themselves: a move built on any allowlisted type (or nested combination of
// them — FixUpMulti embeds FixUp embeds Default, and all three are allowed,
// so a move built on any of them walks clean) walks clean; a move built on
// DealCountComponents, FinishTurn, etc. surfaces that type's name.
func legalUnsupportedMovesBaseType(move Move) string {
	var found string
	legalWalkMovesEmbeds(move, func(name string) bool {
		if !legalSupportedMovesBaseTypes[name] {
			found = name
			return true
		}
		return false
	})
	return found
}

// legalWalkMovesEmbeds walks move's anonymous-embed graph depth-first,
// calling visit with the name of every embedded framework moves-package type
// found (e.g. "Default", "CurrentPlayer", "StartPhase", ...), stopping and
// returning true as soon as visit returns true. legalUnsupportedMovesBaseType
// uses it to find the first framework embed outside the supported seam.
func legalWalkMovesEmbeds(move Move, visit func(name string) bool) bool {
	seen := make(map[reflect.Type]bool)

	var walk func(t reflect.Type) bool
	walk = func(t reflect.Type) bool {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct || seen[t] {
			return false
		}
		seen[t] = true

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.Anonymous {
				continue
			}
			ft := field.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() != reflect.Struct {
				continue
			}
			if strings.HasSuffix(ft.PkgPath(), legalMovesPackagePathSuffix) {
				if visit(ft.Name()) {
					return true
				}
			}
			if walk(ft) {
				return true
			}
		}
		return false
	}

	return walk(reflect.TypeOf(move))
}
