// Go: legal/catalog_framework.go, legal_framework.go and moves/
// catalog_framework.go. The "base check" predicates moves.Default and
// moves.CurrentPlayer contribute ahead of a move's own specs: inPhase,
// inProgression, stackConstraints. (The fourth, proposerIsCurrentPlayer, is
// in catalogPlayers.ts.)
//
// Phase and progression logic is not implemented here. inPhase reads the
// current phase from LegalContext.framework; inProgression delegates the
// whole decision to it. The engine task that owns phases and move
// progressions supplies that object.

import type { LegalPropertyReader, LegalStack, LegalState } from './access.ts';
import { specArgs, specError, specTemplate } from './catalogSupport.ts';
import { goAtoi, goQuote } from './goformat.ts';
import { type LegalPredicateConstructor, newLegalPredicate } from './predicate.ts';
import { TEMPLATE_IN_PHASE, TEMPLATE_IN_PROGRESSION, TEMPLATE_STACK_CONSTRAINTS } from './templates.ts';
import {
  LEGAL_COST_CHEAP,
  LEGAL_COST_MODERATE,
  type LegalSpec,
  failT,
  passVerdict,
  unknownVerdict,
} from './types.ts';
import type { LegalFramework } from './access.ts';

/** Go: legal.InPhase. Legal in any of `phases` (enum keys) or their descendants; no phases means every phase. */
export function inPhase(...phases: number[]): LegalSpec {
  return { name: 'inPhase', args: phases.map(String) };
}

/** Go: moves.inProgressionSpec. Contributed, never authored: the owning move's name is its only argument. */
export function inProgression(moveName: string): LegalSpec {
  return { name: 'inProgression', args: [moveName] };
}

/** Go: legal.StackConstraints. Property NAMES on the game state, not paths. */
export function stackConstraints(srcProperty: string, dstProperty: string): LegalSpec {
  return { name: 'stackConstraints', args: [srcProperty, dstProperty] };
}

/**
 * Go: LegalInPhaseCheck. null when the current phase (or one of its TreeEnum
 * ancestors) is in `legalPhases`, otherwise Go's error text.
 */
export function legalInPhaseCheck(state: LegalState, framework: LegalFramework, legalPhases: readonly number[]): string | null {
  if (legalPhases.length === 0) {
    return null;
  }
  const phase = framework.currentPhase(state);
  const currentKey = phase === null ? 0 : phase.key;
  const candidates = phase !== null && phase.ancestors !== null ? phase.ancestors : [currentKey];
  for (const legalPhase of legalPhases) {
    if (candidates.includes(legalPhase)) {
      return null;
    }
  }
  return 'Move is not legal in phase ' + (phase === null ? String(currentKey) : phase.name);
}

/** Go: reader.ImmutableStackProp(name) as LegalStackConstraintsCheck uses it: any failure is "nothing to check". */
function gameStack(reader: LegalPropertyReader, name: string): LegalStack | null {
  const props = reader.props();
  if (!Object.prototype.hasOwnProperty.call(props, name) || props[name] !== 'TypeStack') {
    return null;
  }
  try {
    return (reader.prop(name) as LegalStack | null) ?? null;
  } catch {
    return null;
  }
}

/**
 * Go: LegalStackConstraintsCheck. Whether the source stack's first component
 * would be accepted by the destination. A missing or empty stack on either
 * side is "nothing to check" (null).
 */
export function legalStackConstraintsCheck(state: LegalState, srcName: string, dstName: string): string | null {
  const reader = state.game();
  const src = gameStack(reader, srcName);
  if (src === null) {
    return null;
  }
  const dst = gameStack(reader, dstName);
  if (dst === null) {
    return null;
  }
  const first = src.first();
  if (first === null) {
    return null;
  }
  return first.mayMoveTo(dst);
}

/** Go: inPhaseConstructor. */
export function inPhaseConstructor(): LegalPredicateConstructor {
  return {
    name: 'inPhase',
    construct(spec) {
      const args = specArgs(spec);
      const phases = args.map((arg, i) => {
        const parsed = goAtoi(arg);
        if (!parsed.ok) {
          return specError(`legal: inPhase: arg ${i} (${goQuote(arg)}) is not an integer enum key: ${parsed.error}`);
        }
        return parsed.value;
      });
      const template = specTemplate(spec, TEMPLATE_IN_PHASE);
      return newLegalPredicate({
        name: 'inPhase',
        clientEvaluable: true,
        args: [...args],
        reads: [{ path: 'game.Phase', facet: 'values' }],
        requiredReadTypes: { 'game.Phase': 'TypeEnum' },
        cost: LEGAL_COST_CHEAP,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['detail'] },
        evaluate(ctx) {
          if (ctx.state === null) {
            return unknownVerdict('legal: inPhase: state was nil');
          }
          if (phases.length === 0) {
            return passVerdict();
          }
          if (!ctx.framework) {
            // No Go counterpart: Go always has a delegate to ask.
            return unknownVerdict('legal: inPhase: no phase provider was supplied (LegalContext.framework)');
          }
          const err = legalInPhaseCheck(ctx.state, ctx.framework, phases);
          if (err !== null) {
            return failT(template, { detail: err });
          }
          return passVerdict();
        },
      });
    },
  };
}

/**
 * Go: moves.inProgressionConstructor. The verdict is the host's
 * moveInProgression answer; this predicate only shapes it.
 */
export function inProgressionConstructor(): LegalPredicateConstructor {
  return {
    name: 'inProgression',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 1) {
        specError(`moves: inProgression requires 1 arg (moveName), got ${args.length}`);
      }
      const moveName = args[0];
      const template = specTemplate(spec, TEMPLATE_IN_PROGRESSION);
      return newLegalPredicate({
        name: 'inProgression',
        args: [...args],
        reads: [{ path: 'game.Phase', facet: 'values' }],
        requiredReadTypes: { 'game.Phase': 'TypeEnum' },
        cost: LEGAL_COST_MODERATE,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['detail'] },
        evaluate(ctx) {
          if (ctx.state === null) {
            return unknownVerdict('moves: inProgression: state was nil');
          }
          if (!ctx.framework || !ctx.framework.moveInProgression) {
            // No Go counterpart: Go always has the move's own progression.
            return unknownVerdict('moves: inProgression: no progression provider was supplied (LegalContext.framework)');
          }
          const result = ctx.framework.moveInProgression(moveName, ctx.state, ctx.proposer);
          if (result === null) {
            return passVerdict();
          }
          if (typeof result === 'string') {
            return failT(template, { detail: result });
          }
          return unknownVerdict(result.unknown);
        },
      });
    },
  };
}

/** Go: stackConstraintsConstructor. */
export function stackConstraintsConstructor(): LegalPredicateConstructor {
  return {
    name: 'stackConstraints',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 2) {
        specError(`legal: stackConstraints requires 2 args (srcProperty, dstProperty), got ${args.length}`);
      }
      const [srcName, dstName] = args;
      const template = specTemplate(spec, TEMPLATE_STACK_CONSTRAINTS);
      return newLegalPredicate({
        name: 'stackConstraints',
        clientEvaluable: true,
        args: [...args],
        reads: [
          { path: 'game.' + srcName, facet: 'values' },
          { path: 'game.' + dstName, facet: 'values' },
        ],
        requiredReadTypes: { ['game.' + srcName]: 'TypeStack', ['game.' + dstName]: 'TypeStack' },
        cost: LEGAL_COST_MODERATE,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['detail'] },
        evaluate(ctx) {
          if (ctx.state === null) {
            return unknownVerdict('legal: stackConstraints: state was nil');
          }
          const err = legalStackConstraintsCheck(ctx.state, srcName, dstName);
          if (err !== null) {
            return failT(template, { detail: err });
          }
          return passVerdict();
        },
      });
    },
  };
}
