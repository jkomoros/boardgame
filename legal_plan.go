package boardgame

import (
	stderrors "errors"
	"fmt"
	"reflect"
	"strings"
)

/*
This file is the declarative-legality evaluation engine (design spec §4) and
its boot-time assembly. It is the campaign's highest-risk surface: it must
deliver the "prime guarantee" that a move which does NOT declare
WithPreconditions behaves byte-for-byte as it does today (the frozen chain in
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
// half of the same moves-package surface: authored WithPreconditions specs
// (in declaration order) and WithoutPrecondition suppression names. A nil/
// empty authored-specs return means the move type has NOT opted in (design
// spec §2: "declaring is implementing" — authored specs are the opt-in
// signal; contributions or suppressions alone do not opt in).
type legalDeclarer interface {
	DeclaredPreconditions() ([]LegalSpec, []string)
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
// NewGameManager (design spec §4). Predicates are split into buckets by
// whether they read any move.* path; the split exists so a later task can
// memoize the field-independent half across the move-forms player+admin
// double pass (keyed without the move). Evaluation order is plan order —
// contributed atoms first (base-first), then authored atoms in declaration
// order, then custom — with NO Cost sort (design spec §4: what you declare is
// what runs, in the order you wrote it).
type legalPlan struct {
	// moveName is the owning move type's name (mType.Name()), used as part
	// of the field-independent memo's key (legal_memo.go) so a lookup can't
	// collide across move types sharing a game.
	moveName string
	// fieldIndependent holds resolved predicates with no move.* reads, in
	// plan order.
	fieldIndependent []*LegalPredicate
	// fieldDependent holds resolved predicates that read at least one move.*
	// path (proposerIsCurrentPlayer among them — it reads
	// move.TargetPlayerIndex), in plan order, evaluated after every
	// fieldIndependent predicate.
	fieldDependent []*LegalPredicate
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
	// FieldDependent reports whether this predicate reads any move.* path
	// (its verdict is provisional on the server-chosen move fields).
	FieldDependent bool
	// Reads is this predicate's declared read-set (nil for the opaque
	// custom wrapper).
	Reads []LegalRead
}

// evaluate runs the plan against ctx. Evaluation order is fieldIndependent →
// fieldDependent → custom, each in plan order (design spec §4).
//
// In hot-path mode (fullLedger == false) it short-circuits on the FIRST
// non-Pass verdict (Fail OR Unknown — fail-closed: an Unknown means legality
// could not be confirmed, so the move is treated as not legal and its verdict
// returned, exactly as a Fail would be), preserving today's first-failure
// error precedence; the returned entry slice is nil. In this mode the
// fieldIndependent bucket's overall verdict is served from (and stored into)
// ctx.State.Game()'s field-independent memo (design spec §5's honest table,
// legal_memo.go) rather than always re-evaluated — see
// evaluateFieldIndependentMemoized. A ctx with no Game to anchor a memo to
// (e.g. a probe or an isolated test evaluating against ExampleState()) just
// evaluates fresh every time, with no memo interaction at all.
//
// In full-ledger mode (fullLedger == true) it evaluates EVERY predicate,
// returns the parallel []LegalVerdictEntry for the ledger, and returns as the
// overall verdict the first non-Pass encountered (or Pass if all passed).
// Every predicate is run through evalLegalPredicate, so a panicking or
// invalid-verdict predicate degrades to a fail-closed Unknown, never to Pass.
// Full-ledger mode never consults or populates the memo: it needs a fresh
// per-predicate entry for every bucket member, which a cached bucket-level
// verdict can't supply.
func (p *legalPlan) evaluate(ctx LegalContext, fullLedger bool) (LegalVerdict, []LegalVerdictEntry) {
	if p == nil {
		// A nil plan cannot confirm legality; fail closed.
		return LegalVerdict{Outcome: LegalUnknown, Reason: "nil legal plan"}, nil
	}

	overall := LegalVerdict{Outcome: LegalPass}
	haveOverall := false

	var entries []LegalVerdictEntry

	consider := func(pred *LegalPredicate, fieldDependent bool) bool {
		v := evalLegalPredicate(pred, ctx)
		if fullLedger {
			entries = append(entries, LegalVerdictEntry{
				Name:           pred.Name,
				Args:           pred.Args,
				Verdict:        v,
				Serializable:   pred.Serializable(),
				FieldDependent: fieldDependent,
				Reads:          pred.Reads,
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

	if fullLedger {
		for _, pred := range p.fieldIndependent {
			if consider(pred, false) {
				return overall, nil
			}
		}
	} else {
		if verdict := p.evaluateFieldIndependentMemoized(ctx); verdict.Outcome != LegalPass {
			return verdict, nil
		}
		// Passed (or the bucket was empty): fall through to fieldDependent /
		// custom below with overall/haveOverall still at their Pass/false
		// zero values, exactly as if the fieldIndependent loop above had run
		// and found nothing to complain about.
	}
	for _, pred := range p.fieldDependent {
		if consider(pred, true) {
			return overall, nil
		}
	}
	if p.custom != nil {
		if consider(p.custom, true) {
			return overall, nil
		}
	}

	return overall, entries
}

// evaluateFieldIndependentMemoized returns p's fieldIndependent bucket's
// overall verdict (Pass if every member passed, else the FIRST non-pass
// verdict encountered, in plan order — exactly what the fieldIndependent
// loop in evaluate's fullLedger branch would compute), consulting and
// populating ctx.State.Game()'s field-independent memo (design spec §5,
// legal_memo.go) keyed on (p.moveName, ctx.State.Version(), ctx.Proposer).
//
// If ctx.State (or its Game()) is nil there is nowhere to anchor a memo, so
// this just evaluates the bucket fresh with no caching at all — the
// behavior every existing isolated test (evaluating against a bare
// LegalContext{} or ExampleState(), whose Game() is nil) already exercises.
func (p *legalPlan) evaluateFieldIndependentMemoized(ctx LegalContext) LegalVerdict {
	game := legalMemoGame(ctx)
	if game == nil {
		return evaluateLegalPredicateBucket(p.fieldIndependent, ctx)
	}

	key := legalFieldIndepMemoKey{moveName: p.moveName, version: ctx.State.Version(), proposer: ctx.Proposer}

	if verdict, ok := game.legalFieldIndepMemoGet(key); ok {
		return verdict
	}

	verdict := evaluateLegalPredicateBucket(p.fieldIndependent, ctx)
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

// evaluateLegalPredicateBucket evaluates preds in order via
// evalLegalPredicate, short-circuiting on and returning the FIRST non-Pass
// verdict, or a Pass verdict if every predicate passed (or preds is empty).
// This is the same short-circuit semantics evaluate's fieldIndependent loop
// uses, factored out so evaluateFieldIndependentMemoized can compute a
// cache-miss's value without duplicating it.
func evaluateLegalPredicateBucket(preds []*LegalPredicate, ctx LegalContext) LegalVerdict {
	for _, pred := range preds {
		if v := evalLegalPredicate(pred, ctx); v.Outcome != LegalPass {
			return v
		}
	}
	return LegalVerdict{Outcome: LegalPass}
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
		State:    state,
		Move:     move,
		Proposer: proposer,
		Chest:    g.chest,
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
		State:    state,
		Move:     move,
		Proposer: proposer,
		Chest:    g.chest,
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

// assembleLegalPlans is called once at the end of NewGameManager (after moves
// are installed and the example state exists). For every installed move type
// that has opted in to declarative legality (declares WithPreconditions), it
// verifies the move is on a supported base (design spec §2's v1 seam:
// Default/CurrentPlayer only), assembles and validates its plan, stores it,
// and probes that the declarations are actually reachable. Any failure is a
// boot error naming the offending move (and, for the seam check, the
// unsupported base type). A move with no authored preconditions is left
// entirely alone: no plan, no probe — its frozen chain runs at runtime
// exactly as today.
func (g *GameManager) assembleLegalPlans(exampleState ImmutableState) error {
	registry, templateTable := g.buildLegalRegistryAndTemplates()
	g.legalTemplateTable = templateTable

	for _, mType := range g.moves {
		move := mType.NewMove(exampleState)
		if move == nil {
			// A half-functioning move type; other boot checks report it.
			continue
		}

		declarer, ok := move.(legalDeclarer)
		if !ok {
			continue
		}
		authored, suppressions := declarer.DeclaredPreconditions()
		if len(authored) == 0 {
			// Not opted in (design spec §2). Frozen chain runs at runtime.
			continue
		}

		// v1 seam (design spec §2): a move embedding any moves-package
		// framework type beyond Default/CurrentPlayer cannot opt in — its
		// imperative Legal() would interleave wrongly with plan evaluation.
		if base := legalUnsupportedMovesBaseType(move); base != "" {
			return fmt.Errorf("move %q declares preconditions but embeds unsupported framework move type %q: only moves.Default and moves.CurrentPlayer support declarative legality in v1", mType.Name(), base)
		}

		var contributed []LegalSpec
		if contributor, ok := move.(legalContributor); ok {
			contributed = contributor.ContributedPreconditions()
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

		plan := buildLegalPlanFromPredicates(mType.Name(), predicates, specs, move)
		if g.legalPlans == nil {
			g.legalPlans = make(map[string]*legalPlan)
		}
		g.legalPlans[mType.Name()] = plan

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
func (g *GameManager) buildLegalRegistryAndTemplates() (map[string]*LegalPredicateConstructor, map[string]string) {
	registry, templates := legalRegistrySnapshot()

	if cc, ok := g.delegate.(legalConstructorConfigurer); ok {
		for _, ctor := range cc.ConfigurePredicateConstructors() {
			if ctor == nil || ctor.Name == "" {
				continue
			}
			registry[ctor.Name] = ctor
		}
	}

	if tc, ok := g.delegate.(legalTemplateConfigurer); ok {
		for k, v := range tc.ConfigureLegalTemplates() {
			templates[k] = v
		}
	}

	return registry, templates
}

// assembleLegalSpecList produces the final ordered spec list for a plan
// (design spec §2's "Plan assembly"): contributed atoms first (base-first,
// deterministic), then authored atoms in declaration order, MINUS any
// contributed atom whose Name is in suppressions (WithoutPrecondition
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

// buildLegalPlanFromPredicates splits resolved predicates into the plan's
// field-independent / field-dependent buckets (by whether a predicate reads
// any move.* path — see legalReadsIncludeMovePath), preserving plan order
// within each bucket, and appends the CustomLegaler wrapper (if move
// implements CustomLegaler) as the plan's custom tail. moveName is stored on
// the plan for the field-independent memo's key (legal_memo.go); pass "" if
// the caller doesn't care about memoization (e.g. an isolated test that
// never evaluates against a real *Game — see legalPlan.evaluate, which skips
// the memo whenever ctx.State.Game() is nil regardless of moveName). specs is
// the parallel serializable spec list.
func buildLegalPlanFromPredicates(moveName string, predicates []*LegalPredicate, specs []LegalSpec, move Move) *legalPlan {
	plan := &legalPlan{specs: specs, moveName: moveName}
	for _, pred := range predicates {
		if legalReadsIncludeMovePath(pred.Reads) {
			plan.fieldDependent = append(plan.fieldDependent, pred)
		} else {
			plan.fieldIndependent = append(plan.fieldIndependent, pred)
		}
	}
	if _, ok := move.(CustomLegaler); ok {
		plan.custom = newLegalCustomWrapper()
	}
	plan.allReads = legalPlanAllReads(plan)
	return plan
}

// legalPlanAllReads returns the de-duplicated union of every plan predicate's
// Reads, in first-seen order (field-independent, then field-dependent).
func legalPlanAllReads(plan *legalPlan) []LegalRead {
	seen := make(map[LegalRead]bool)
	var out []LegalRead
	add := func(preds []*LegalPredicate) {
		for _, pred := range preds {
			for _, r := range pred.Reads {
				if !seen[r] {
					seen[r] = true
					out = append(out, r)
				}
			}
		}
	}
	add(plan.fieldIndependent)
	add(plan.fieldDependent)
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
			err := cl.LegalCustom(ctx.State, ctx.Proposer)
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
		return fmt.Errorf("move %q declares preconditions but its Legal() override never reaches moves.Default.Legal — declarations would be dead (use LegalCustom for imperative residue, or super-call the embedded chain)", mType.Name())
	}
	return nil
}

// legalMovesPackagePathSuffix is the import-path suffix identifying the
// framework moves package. Matches moves/default.go's own convention for
// recognizing moves-package types by reflection.
const legalMovesPackagePathSuffix = "boardgame/moves"

// legalUnsupportedMovesBaseType walks move's anonymous-embed graph looking
// for an embedded struct type from the framework moves package that is NOT
// Default or CurrentPlayer (design spec §2's v1 seam). It returns the name of
// the first such type found, or "" if the move embeds only supported
// framework types. Detection is by reflection over the embed graph (the
// spec-offered alternative to a marker-method chain) because it is robust to
// deep embedding and needs no cooperation from the framework types
// themselves: CurrentPlayer embeds Default, and both are allowed, so a move
// built on either walks clean; a move built on DealCountComponents,
// StartPhase, FinishTurn, etc. surfaces that type's name.
func legalUnsupportedMovesBaseType(move Move) string {
	seen := make(map[reflect.Type]bool)

	var walk func(t reflect.Type) string
	walk = func(t reflect.Type) string {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct || seen[t] {
			return ""
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
				name := ft.Name()
				if name != "Default" && name != "CurrentPlayer" {
					return name
				}
			}
			if found := walk(ft); found != "" {
				return found
			}
		}
		return ""
	}

	return walk(reflect.TypeOf(move))
}
