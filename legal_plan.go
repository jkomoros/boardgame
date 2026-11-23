package boardgame

import (
	"fmt"
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
	// readsMovePlayer reports that some step (at any depth, via the
	// unioned Reads) reads a "player.X" path, so evaluation pins the move's
	// player before running predicates. Plans without one never call the
	// delegate's CurrentPlayerIndex on their own.
	readsMovePlayer bool
	// movePlayerIsProposer reports that "player.X" in this plan names the
	// proposer rather than the current player (see
	// LegalVerdictEntry.MovePlayerIsProposer).
	movePlayerIsProposer bool
	// independentReadsMovePlayer reports that some field-independent step
	// reads a "player.X" path, so the memo key must carry the resolved
	// move player.
	independentReadsMovePlayer bool
	// specs is the serializable subset of the plan (every non-opaque
	// predicate's originating spec shape), preserved for the ledger and
	// wire format.
	specs []LegalSpec
}

type legalPlanStep struct {
	predicate      *LegalPredicate
	fieldDependent bool
	memoIndex      int
	// readsMovePlayer reports that the predicate reads a "player.X" path,
	// whose resolution may come from the move's on-behalf-of player (see
	// legalMovePlayerIndex). Such a step is reported as field-dependent in
	// the ledger exactly when that happened for this evaluation.
	readsMovePlayer bool
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
	// MovePlayerIsProposer reports that this entry's "player.X" reads name
	// the proposer rather than the game's current player: the move is made
	// for its own player (LegalOnBehalfOfer) and its plan has no
	// proposerIsCurrentPlayer check tying that player to the current one
	// (a moves.AnyPlayer move). ReadsEvaluable uses it to judge those reads
	// conservatively.
	MovePlayerIsProposer bool
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

	// Resolve the move's player once, while the whole move is in hand:
	// predicates reading "player.X" use it through ctx, the memo keys on it,
	// and the ledger reports when it came from a move field.
	// Plans with no memoized player.* step keep the historical memo key.
	movePlayer := ObserverPlayerIndex
	movePlayerFromMove := false
	if p.readsMovePlayer {
		ctx, movePlayerFromMove = ctx.withMovePlayer()
		if p.independentReadsMovePlayer {
			// An unresolved player never shares a memo entry with a resolved one.
			movePlayer = AnyPlayerIndex
			if ctx.movePlayerResolved {
				movePlayer = ctx.movePlayer
			}
		}
	}

	consider := func(step legalPlanStep) bool {
		pred := step.predicate
		var v LegalVerdict
		if fullLedger && haveOverall && pred.residue {
			// A framework residue assumes its predecessors passed; the hot
			// path would never have run it (legal_residue.go).
			v = legalNotEvaluatedVerdict()
		} else {
			v = p.evaluateStep(ctx, step, fullLedger, movePlayer)
		}
		if fullLedger {
			entries = append(entries, LegalVerdictEntry{
				Name:                 pred.Name,
				Args:                 pred.Args,
				Verdict:              v,
				Serializable:         pred.Serializable(),
				ClientEvaluable:      pred.ClientEvaluable,
				FieldDependent:       step.fieldDependent || (step.readsMovePlayer && movePlayerFromMove),
				Reads:                pred.Reads,
				MovePlayerIsProposer: p.movePlayerIsProposer,
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

func (p *legalPlan) evaluateStep(ctx LegalContext, step legalPlanStep, fullLedger bool, movePlayer PlayerIndex) LegalVerdict {
	if fullLedger || step.fieldDependent {
		return evalLegalPredicate(step.predicate, ctx)
	}
	game := legalMemoGame(ctx)
	if game == nil {
		return evalLegalPredicate(step.predicate, ctx)
	}
	key := legalFieldIndepMemoKey{
		moveName:   p.moveName,
		version:    ctx.State.Version(),
		proposer:   ctx.ProposerPlayerIndex,
		movePlayer: movePlayer,
		predicate:  step.memoIndex,
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
	return g != nil && g.legalPlans != nil && g.legalPlans[moveName] != nil && !g.legalPlansAreSuspended()
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
	if g.legalPlans == nil || g.legalPlansAreSuspended() {
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
	return fmt.Errorf("move %q implements CustomLegaler (LegalCustom), which automatically opts into declarative legality, but it embeds no framework move type from package moves, so nothing would ever evaluate its plan; embed moves.Default (or any other framework move) or move the LegalCustom logic into a Legal() override", moveName)
}

// assembleLegalPlans is called once at the end of NewGameManager (after moves
// are installed and the example state exists). For every installed move type
// that has opted in to declarative legality (declares WithLegalPreconditions or
// explicitly enables a zero-authored-spec plan), it assembles and validates
// its plan (every embedded framework base type contributes its own legality;
// see legal_residue.go), stores it, and probes that the declarations are
// actually reachable. Any failure is a boot error naming the offending move. A
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

		// Every framework base type expresses its legality as its
		// contribution (design spec 2026-11-22), so there is no base-type
		// allowlist: whatever the move embeds contributes below.

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

		if err := legalValidateResidueAtoms(mType.Name(), contributed, authored, suppressions, registry); err != nil {
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

		// Footgun-batch F3: runtime enforcement (legal_reads.go) already
		// makes an under-declared read fail closed instead of serving a
		// stale memoized verdict; this probe turns the unconditional case
		// into a boot error naming the predicate and path.
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
		sealLegalPredicate(pred, pred.Args)
		readsMovePlayer := legalReadsIncludeMovePlayerPath(pred.Reads)
		if legalPredicateOpaque(pred) || legalReadsIncludeMovePath(pred.Reads) {
			plan.fieldDependent = append(plan.fieldDependent, pred)
			plan.ordered = append(plan.ordered, legalPlanStep{predicate: pred, fieldDependent: true, memoIndex: -1, readsMovePlayer: readsMovePlayer})
		} else {
			memoIndex := len(plan.fieldIndependent)
			plan.fieldIndependent = append(plan.fieldIndependent, pred)
			plan.ordered = append(plan.ordered, legalPlanStep{predicate: pred, memoIndex: memoIndex, readsMovePlayer: readsMovePlayer})
			if readsMovePlayer {
				plan.independentReadsMovePlayer = true
			}
		}
	}
	if _, ok := move.(CustomLegaler); ok {
		plan.custom = newLegalCustomWrapper()
	}
	plan.allReads = legalPlanAllReads(plan)
	plan.readsMovePlayer = legalReadsIncludeMovePlayerPath(plan.allReads)
	if _, ok := move.(LegalOnBehalfOfer); ok {
		plan.movePlayerIsProposer = true
		for _, pred := range predicates {
			if pred.Name == "proposerIsCurrentPlayer" {
				plan.movePlayerIsProposer = false
			}
		}
	}
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
			return legalVerdictFromError(cl.LegalCustom(ctx.State, ctx.ProposerPlayerIndex))
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
	g.legalProbeReachedResidues = nil
	// The returned error is deliberately ignored: the probe observes
	// reachability via the flags, not the verdict.
	legalProbeCall(probeMove, exampleState)
	reached := g.legalProbeReached
	reachedResidues := g.legalProbeReachedResidues
	g.legalProbing = false
	g.legalProbeReached = false
	g.legalProbeReachedResidues = nil

	if !reached {
		return legalProbeUnreachableError(mType.Name())
	}
	if missing := legalProbeUnreachedResidues(g.legalPlans[mType.Name()], reachedResidues); len(missing) > 0 {
		return legalProbeUnreachedResidueError(mType.Name(), missing)
	}
	return nil
}

// legalProbeUnreachableError reports a move whose own Legal() override never
// super-calls into moves.Default.Legal, so its declarations would be dead.
func legalProbeUnreachableError(moveName string) error {
	return fmt.Errorf("move %q declares preconditions but its Legal() override never reaches moves.Default.Legal — declarations would be dead (use LegalCustom for imperative residue, or super-call the embedded chain); put the super-call FIRST in your override — one that conditionally returns before super-calling can trip this same probe even against the always-valid example state used to run it; every framework base type's own Legal() reaches it, so this probe only ever fires on a move's OWN override", moveName)
}

/*
Footgun-batch F3: the undeclared-read boot probe.

Since the enforced-reads design (legal_reads.go,
docs/superpowers/specs/2026-11-22-legality-enforced-reads-design.md) this probe
is a DIAGNOSTIC, not the safety mechanism. Correctness no longer depends on it:
at runtime a predicate that declares no move-dependent read is evaluated with
the move withheld, and ctx.ResolvePath refuses undeclared paths, so an
under-declared read degrades to a deterministic fail-closed Unknown instead of
a stale memoized verdict. What the probe adds is a better error, earlier: it
evaluates every game-registered predicate once at boot, against the example
state, under the same read scope the runtime uses, and turns an unconditional
violation into a boot error naming the move, the predicate, and the path.

A field-independent predicate is probed with a sentinel Move whose
PropertyReader panics on every property access, so a raw ctx.Move.Reader()
access is distinguishable from an unrelated panic. A field-dependent predicate
is probed with the real example move; only out-of-scope ResolvePath calls are
reported for it.

Precision properties, both load-bearing:

  - It can only fire on a genuine violation: the sentinel panic payload is an
    unexported type only the sentinel reader throws, scope violations are
    recorded by ResolvePath itself, and any OTHER recovered panic is
    swallowed (at runtime evalLegalPredicate degrades such a panic to a
    fail-closed Unknown — a predicate that panics at boot for unrelated
    reasons must not newly fail boot).
  - It cannot change evaluation semantics; it only adds boot errors.

Blind spots that remain diagnostic-only (runtime enforcement still covers
them): a read conditional on state the example state doesn't exhibit, and a
concrete type assertion on ctx.Move (the sentinel is not the game's type).
Catalog predicates are skipped: they get identical runtime enforcement, and
their read-sets are covered in-repo by the conformance corpus.
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

// legalProbeUndeclaredMoveReads walks every predicate in plan (recursing
// through compositor Sub trees, so a game-registered predicate inside an
// "any" is probed too) and evaluates each game-registered predicate once
// against exampleState under its runtime read scope. A move read from a
// predicate that declares none, or an out-of-scope ResolvePath, is a boot
// error naming the move and the predicate. See the file-section comment
// above for the full rationale.
func legalProbeUndeclaredMoveReads(moveName string, plan *legalPlan, gameRegistered map[string]bool,
	exampleState ImmutableState, exampleMove Move, chest *ComponentChest) error {

	if len(gameRegistered) == 0 || plan == nil {
		return nil
	}

	var walk func(pred *LegalPredicate) error
	walk = func(pred *LegalPredicate) error {
		if pred == nil {
			return nil
		}
		if gameRegistered[pred.Name] && pred.Evaluate != nil && !pred.opaque {
			if err := legalProbeGameRegistered(moveName, pred, exampleState, exampleMove, chest); err != nil {
				return err
			}
		}
		for _, sub := range pred.Sub {
			if err := walk(sub); err != nil {
				return err
			}
		}
		return nil
	}

	for _, step := range plan.ordered {
		if err := walk(step.predicate); err != nil {
			return err
		}
	}
	return nil
}

func legalProbeGameRegistered(moveName string, pred *LegalPredicate, exampleState ImmutableState, exampleMove Move, chest *ComponentChest) error {
	scope := pred.readScope
	if scope == nil {
		scope = newLegalReadScope(pred.Reads)
	}
	enforcement := &legalReadEnforcement{predicate: pred.Name, scope: scope, moveWithheld: !scope.readsMove}
	ctx := LegalContext{
		State:               exampleState,
		Move:                exampleMove,
		ProposerPlayerIndex: ObserverPlayerIndex,
		Chest:               chest,
		reads:               enforcement,
	}
	if enforcement.moveWithheld {
		ctx.Move = &legalProbeSentinelMove{Move: exampleMove}
	}
	touchedMove := legalProbeEvaluateTouchesMove(pred, ctx)
	if touchedMove || (enforcement.violation != "" && enforcement.moveWithheld && legalPathReadsMove(enforcement.violation)) {
		return fmt.Errorf("move %q: game-registered predicate %q reads properties off the move at evaluation time but declares no move.* (or players[move.*]) path in Reads, so the engine evaluates it without the move and it can never pass — declare every move path the predicate reads in its Reads, or leave Reads unset to derive them from the spec's path args (see legal/doc.go's game-registered predicates section)", moveName, pred.Name)
	}
	if enforcement.violation != "" {
		return fmt.Errorf("move %q: game-registered predicate %q resolves %q, which is not in its declared Reads, so the engine refuses the read and the predicate can never pass — declare it in Reads, or leave Reads unset to derive them from the spec's path args (see legal/doc.go's game-registered predicates section)", moveName, pred.Name, enforcement.violation)
	}
	return nil
}

// legalProbeEvaluateTouchesMove runs pred.Evaluate DIRECTLY (not through
// evalLegalPredicate, whose recover would silently convert the sentinel
// panic to an Unknown verdict and hide the evidence) and reports whether it
// panicked with the sentinel payload. Any other panic — or any verdict at
// all — reports false. Scope violations are recorded on ctx's enforcement
// for the caller to inspect.
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
