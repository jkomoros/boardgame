// Go: legal/catalog_players.go and legal/catalog_behaviors.go. The
// player-shaped predicates: the allActivePlayers quantifier, the two
// proposer checks, and playerBoolAt with its typed player selectors.

import type { LegalPropertyReader } from './access.ts';
import { legalCompareOp, playerBoolWant } from './catalogCompare.ts';
import { readTyped, specArgs, specError, specTemplate, unknownFrom } from './catalogSupport.ts';
import { goAtoi, goQuote, goStringSlice } from './goformat.ts';
import { resolveBoolPath, resolvePlayerIndexPath } from './path.ts';
import { playerIndexEnsureValid, playerIndexEquivalent, playerIndexValid } from './playerIndex.ts';
import { LEGAL_ANY_NAME, type LegalPredicateConstructor, newLegalPredicate } from './predicate.ts';
import {
  TEMPLATE_ALL_ACTIVE_PLAYERS,
  TEMPLATE_PLAYER_ACTIVE,
  TEMPLATE_PLAYER_ALREADY_SUBMITTED,
  TEMPLATE_PLAYER_BOOL,
  TEMPLATE_PLAYER_INACTIVE,
  TEMPLATE_PLAYER_NOT_ADMIN,
  TEMPLATE_PLAYER_NOT_SUBMITTED,
  TEMPLATE_PROPOSER_NOT_MOVE_PLAYER,
  TEMPLATE_PROPOSER_NOT_YOUR_TURN,
  TEMPLATE_PROPOSER_TARGET_INVALID,
  TEMPLATE_SEAT_NOT_CLOSED,
  TEMPLATE_SEAT_NOT_FILLED,
} from './templates.ts';
import {
  LEGAL_COST_CHEAP,
  LEGAL_COST_MODERATE,
  LEGAL_COST_TRIVIAL,
  type LegalOutcome,
  type LegalRead,
  type LegalSpec,
  type PropertyType,
  failT,
  passVerdict,
  unknownVerdict,
} from './types.ts';

/** Go: legal.Any. The only compositor: passes when any sub passes. Depth 1 only. */
export function any(...subs: LegalSpec[]): LegalSpec {
  return { name: 'any', sub: subs };
}

/**
 * Go: legal.AllActivePlayers. Requires `inner` of every active player.
 * `inner` is a playerBool, a player-path propAtLeast/propCompare, or an
 * any() of those.
 */
export function allActivePlayers(inner: LegalSpec): LegalSpec {
  return { name: 'allActivePlayers', sub: [inner] };
}

/** Go: legal.ProposerIsCurrentPlayer. */
export function proposerIsCurrentPlayer(): LegalSpec {
  return { name: 'proposerIsCurrentPlayer' };
}

/** Go: legal.ProposerIsPlayerFromMove. Carries AdminBypass, as Go's builder does. */
export function proposerIsPlayerFromMove(field: string): LegalSpec {
  return { name: 'proposerIsPlayerFromMove', args: [field], adminPolicy: 'bypass' };
}

/** Go: legal.PlayerSelector. Names which player a behavior predicate reads. */
export interface PlayerSelector {
  /** The path for `prop` on the selected player. */
  path(prop: string): string;
  /** True for the proposer selector, whose behavior predicates carry AdminBypass. */
  readonly isProposer: boolean;
}

/** Go: legal.CurrentPlayer. */
export function currentPlayer(): PlayerSelector {
  return { path: prop => 'player.' + prop, isProposer: false };
}

/** Go: legal.Proposer. */
export function proposer(): PlayerSelector {
  return { path: prop => 'proposer.' + prop, isProposer: true };
}

/** Go: legal.PlayerFromMove. `field` is a PlayerIndex-typed move field. */
export function playerFromMove(field: string): PlayerSelector {
  return { path: prop => 'players[move.' + field + '].' + prop, isProposer: false };
}

/** Go: legal.PlayerBoolAt. */
export function playerBoolAt(player: PlayerSelector, prop: string, want: boolean): LegalSpec {
  return { name: 'playerBoolAt', args: [player.path(prop), String(want)] };
}

/** Go: behaviorPlayerBool. */
function behaviorPlayerBool(player: PlayerSelector, prop: string, want: boolean, template: string): LegalSpec {
  const spec: LegalSpec = { ...playerBoolAt(player, prop, want), message: template };
  return player.isProposer ? { ...spec, adminPolicy: 'bypass' } : spec;
}

/** Go: legal.PlayerHasSubmitted. */
export function playerHasSubmitted(player: PlayerSelector): LegalSpec {
  return behaviorPlayerBool(player, 'PlayerSubmitted', true, TEMPLATE_PLAYER_NOT_SUBMITTED);
}

/** Go: legal.PlayerHasNotSubmitted. */
export function playerHasNotSubmitted(player: PlayerSelector): LegalSpec {
  return behaviorPlayerBool(player, 'PlayerSubmitted', false, TEMPLATE_PLAYER_ALREADY_SUBMITTED);
}

/** Go: legal.PlayerIsActive. */
export function playerIsActive(player: PlayerSelector): LegalSpec {
  return behaviorPlayerBool(player, 'PlayerInactive', false, TEMPLATE_PLAYER_INACTIVE);
}

/** Go: legal.PlayerIsInactive. */
export function playerIsInactive(player: PlayerSelector): LegalSpec {
  return behaviorPlayerBool(player, 'PlayerInactive', true, TEMPLATE_PLAYER_ACTIVE);
}

/** Go: legal.PlayerSeatIsFilled. */
export function playerSeatIsFilled(player: PlayerSelector): LegalSpec {
  return behaviorPlayerBool(player, 'SeatFilled', true, TEMPLATE_SEAT_NOT_FILLED);
}

/** Go: legal.PlayerSeatIsClosed. */
export function playerSeatIsClosed(player: PlayerSelector): LegalSpec {
  return behaviorPlayerBool(player, 'SeatClosed', true, TEMPLATE_SEAT_NOT_CLOSED);
}

/** Go: legal.PlayerIsAdmin. */
export function playerIsAdmin(player: PlayerSelector): LegalSpec {
  return behaviorPlayerBool(player, 'IsAdmin', true, TEMPLATE_PLAYER_NOT_ADMIN);
}

/** Go: allActivePlayersLeaf. One per-player check, compiled to read a single player's reader. */
interface AllActivePlayersLeaf {
  evaluate: (reader: LegalPropertyReader) => LegalOutcome;
  read: LegalRead;
  requiredType: PropertyType;
}

/** Go: playerPathProp. */
function playerPathProp(path: string): string {
  const prefix = 'player.';
  if (!path.startsWith(prefix)) {
    specError(
      `path ${goQuote(path)} must be a player.* path (AllActivePlayers' inner propAtLeast/propCompare only accept player-path args)`,
    );
  }
  const prop = path.slice(prefix.length);
  if (prop === '') {
    specError(`path ${goQuote(path)} is missing a property name after ${goQuote(prefix)}`);
  }
  return prop;
}

function rethrowWithPrefix(prefix: string, fn: () => string): string {
  try {
    return fn();
  } catch (err) {
    return specError(prefix + (err instanceof Error ? err.message : String(err)));
  }
}

/** Go: buildAllActivePlayersLeaf. */
function buildLeaf(s: LegalSpec): AllActivePlayersLeaf {
  const args = specArgs(s);
  switch (s.name) {
    case 'playerBool': {
      if (args.length !== 1 && args.length !== 2) {
        specError(`legal: allActivePlayers: inner playerBool requires 1 or 2 args (prop, optional want), got ${args.length}`);
      }
      const prop = args[0];
      let want = true;
      if (args.length === 2) {
        const parsed = playerBoolWant(args[1]);
        if (!parsed.ok) {
          specError(`legal: allActivePlayers: inner playerBool: arg 2 (want) ${parsed.error}`);
        }
        want = parsed.value;
      }
      return {
        read: { path: 'players[*].' + prop, facet: 'values' },
        requiredType: 'TypeBool',
        evaluate(reader) {
          const v = readTyped(reader, prop, 'TypeBool');
          if (v === undefined) {
            return 'unknown';
          }
          return v === want ? 'pass' : 'fail';
        },
      };
    }
    case 'propAtLeast': {
      if (args.length !== 2) {
        specError(`legal: allActivePlayers: inner propAtLeast requires 2 args (path, n), got ${args.length}`);
      }
      const prop = rethrowWithPrefix('legal: allActivePlayers: inner propAtLeast: ', () => playerPathProp(args[0]));
      const parsed = goAtoi(args[1]);
      if (!parsed.ok) {
        specError(`legal: allActivePlayers: inner propAtLeast: arg 2 (n) must be an integer: ${parsed.error}`);
      }
      const n = parsed.value;
      return {
        read: { path: 'players[*].' + prop, facet: 'values' },
        requiredType: 'TypeInt',
        evaluate(reader) {
          const v = readTyped(reader, prop, 'TypeInt');
          if (v === undefined) {
            return 'unknown';
          }
          return (v as number) >= n ? 'pass' : 'fail';
        },
      };
    }
    case 'propCompare': {
      if (args.length !== 3) {
        specError(`legal: allActivePlayers: inner propCompare requires 3 args (path, op, n), got ${args.length}`);
      }
      const prop = rethrowWithPrefix('legal: allActivePlayers: inner propCompare: ', () => playerPathProp(args[0]));
      const op = args[1];
      const cmp = legalCompareOp(op);
      if (cmp === undefined) {
        specError(`legal: allActivePlayers: inner propCompare: unknown op ${goQuote(op)} (expected one of ==, !=, <, <=, >, >=)`);
      }
      const parsed = goAtoi(args[2]);
      if (!parsed.ok) {
        specError(`legal: allActivePlayers: inner propCompare: arg 3 (n) must be an integer: ${parsed.error}`);
      }
      const n = parsed.value;
      return {
        read: { path: 'players[*].' + prop, facet: 'values' },
        requiredType: 'TypeInt',
        evaluate(reader) {
          const v = readTyped(reader, prop, 'TypeInt');
          if (v === undefined) {
            return 'unknown';
          }
          return cmp(v as number, n) ? 'pass' : 'fail';
        },
      };
    }
    default:
      return specError(
        `legal: allActivePlayers: unsupported inner predicate ${goQuote(s.name)} (v1 supports playerBool, propAtLeast, and propCompare on player.* paths, optionally composed with any())`,
      );
  }
}

/** Go: buildAllActivePlayersInner. */
function buildInner(inner: LegalSpec): {
  evaluate: (reader: LegalPropertyReader) => LegalOutcome;
  reads: LegalRead[];
  requiredTypes: Record<string, PropertyType>;
} {
  if (inner.name !== LEGAL_ANY_NAME) {
    const leaf = buildLeaf(inner);
    return { evaluate: leaf.evaluate, reads: [leaf.read], requiredTypes: { [leaf.read.path]: leaf.requiredType } };
  }
  const subSpecs = inner.sub ?? [];
  if (subSpecs.length < 2) {
    specError(
      `legal: allActivePlayers: inner ${goQuote(LEGAL_ANY_NAME)} compositor requires at least 2 sub-specs, got ${subSpecs.length}`,
    );
  }
  const leaves: AllActivePlayersLeaf[] = [];
  const reads: LegalRead[] = [];
  const requiredTypes: Record<string, PropertyType> = {};
  for (const sub of subSpecs) {
    const leaf = buildLeaf(sub);
    leaves.push(leaf);
    if (!reads.some(r => r.path === leaf.read.path && r.facet === leaf.read.facet)) {
      reads.push(leaf.read);
    }
    const prior = Object.prototype.hasOwnProperty.call(requiredTypes, leaf.read.path) ? requiredTypes[leaf.read.path] : undefined;
    if (prior !== undefined && prior !== leaf.requiredType) {
      specError(
        `legal: allActivePlayers: inner predicates require conflicting types for ${goQuote(leaf.read.path)} (${prior} and ${leaf.requiredType})`,
      );
    }
    requiredTypes[leaf.read.path] = leaf.requiredType;
  }
  return {
    reads,
    requiredTypes,
    // Kleene, per player: any pass -> pass; else any unknown -> unknown; else fail.
    evaluate(reader) {
      let sawUnknown = false;
      for (const leaf of leaves) {
        const outcome = leaf.evaluate(reader);
        if (outcome === 'pass') {
          return 'pass';
        }
        if (outcome === 'unknown') {
          sawUnknown = true;
        }
      }
      return sawUnknown ? 'unknown' : 'fail';
    },
  };
}

/**
 * Go: allActivePlayersConstructor. Skips players the game reports inactive,
 * fails on the first active player whose inner check fails, and is unknown
 * if none failed but at least one was unknown.
 */
export function allActivePlayersConstructor(): LegalPredicateConstructor {
  return {
    name: 'allActivePlayers',
    construct(spec) {
      const subSpecs = spec.sub ?? [];
      if (subSpecs.length !== 1) {
        specError(`legal: allActivePlayers requires exactly 1 inner spec (in Sub), got ${subSpecs.length}`);
      }
      const inner = buildInner(subSpecs[0]);
      const template = specTemplate(spec, TEMPLATE_ALL_ACTIVE_PLAYERS);
      return newLegalPredicate({
        name: 'allActivePlayers',
        // The inner spec is not shipped with the ledger, so a client cannot
        // re-evaluate this one.
        clientEvaluable: false,
        reads: inner.reads,
        requiredReadTypes: inner.requiredTypes,
        cost: LEGAL_COST_MODERATE,
        emittedTemplates: [template],
        emittedBindings: { [template]: [] },
        evaluate(ctx) {
          if (ctx.state === null) {
            return unknownVerdict('legal: allActivePlayers: state was nil');
          }
          let sawUnknown = false;
          const players = ctx.state.players();
          for (let i = 0; i < players.length; i++) {
            if (ctx.state.playerIsInactive(i)) {
              continue;
            }
            const outcome = inner.evaluate(players[i]);
            if (outcome === 'fail') {
              return failT(template);
            }
            if (outcome === 'unknown') {
              sawUnknown = true;
            }
          }
          if (sawUnknown) {
            return unknownVerdict("legal: allActivePlayers: at least one active player's inner predicate was unknown");
          }
          return passVerdict();
        },
      });
    },
  };
}

/**
 * Go: proposerIsCurrentPlayerConstructor. The move's TargetPlayerIndex must
 * be a real seat, must be the current player, and must be equivalent to the
 * proposer (admin is a wildcard for the last check only).
 */
export function proposerIsCurrentPlayerConstructor(): LegalPredicateConstructor {
  return {
    name: 'proposerIsCurrentPlayer',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 0) {
        specError(`legal: proposerIsCurrentPlayer takes no args, got ${args.length}`);
      }
      const targetInvalidTemplate = specTemplate(spec, TEMPLATE_PROPOSER_TARGET_INVALID);
      const notYourTurnTemplate = specTemplate(spec, TEMPLATE_PROPOSER_NOT_YOUR_TURN);
      return newLegalPredicate({
        name: 'proposerIsCurrentPlayer',
        clientEvaluable: true,
        reads: [
          { path: 'move.TargetPlayerIndex', facet: 'values' },
          { path: 'game.CurrentPlayer', facet: 'values' },
        ],
        requiredReadTypes: { 'move.TargetPlayerIndex': 'TypePlayerIndex', 'game.CurrentPlayer': 'TypePlayerIndex' },
        cost: LEGAL_COST_CHEAP,
        emittedTemplates: [targetInvalidTemplate, notYourTurnTemplate],
        emittedBindings: { [targetInvalidTemplate]: ['detail'], [notYourTurnTemplate]: ['detail'] },
        evaluate(ctx) {
          if (ctx.state === null) {
            return unknownVerdict('legal: proposerIsCurrentPlayer: state was nil');
          }
          const rawTarget = resolvePlayerIndexPath('move.TargetPlayerIndex', ctx);
          if (!rawTarget.ok) {
            return unknownFrom(rawTarget);
          }
          const current = ctx.state.currentPlayerIndex();
          const target = playerIndexEnsureValid(rawTarget.value, ctx.state);
          if (!playerIndexValid(target, ctx.state) || target < 0) {
            return failT(targetInvalidTemplate, { detail: 'The specified target player is not valid' });
          }
          if (!playerIndexEquivalent(target, current) || !playerIndexEquivalent(target, ctx.proposer)) {
            return failT(notYourTurnTemplate, { detail: "it's not your turn" });
          }
          return passVerdict();
        },
      });
    },
  };
}

/** Go: proposerIsPlayerFromMoveConstructor. */
export function proposerIsPlayerFromMoveConstructor(): LegalPredicateConstructor {
  return {
    name: 'proposerIsPlayerFromMove',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 1 || args[0] === '') {
        specError(`legal: proposerIsPlayerFromMove requires 1 non-empty arg (move PlayerIndex field), got ${goStringSlice(args)}`);
      }
      const path = 'move.' + args[0];
      const template = specTemplate(spec, TEMPLATE_PROPOSER_NOT_MOVE_PLAYER);
      return newLegalPredicate({
        name: 'proposerIsPlayerFromMove',
        args: [...args],
        reads: [{ path, facet: 'values' }],
        requiredReadTypes: { [path]: 'TypePlayerIndex' },
        cost: LEGAL_COST_TRIVIAL,
        clientEvaluable: true,
        usesProposer: true,
        emittedTemplates: [template],
        emittedBindings: { [template]: [] },
        evaluate(ctx) {
          if (ctx.state === null) {
            return unknownVerdict('legal: proposerIsPlayerFromMove: state was nil');
          }
          const raw = resolvePlayerIndexPath(path, ctx);
          if (!raw.ok) {
            return unknownFrom(raw);
          }
          const target = playerIndexEnsureValid(raw.value, ctx.state);
          if (
            target < 0 ||
            !playerIndexValid(target, ctx.state) ||
            ctx.proposer < 0 ||
            !playerIndexEquivalent(target, ctx.proposer)
          ) {
            return failT(template);
          }
          return passVerdict();
        },
      });
    },
  };
}

/** Go: playerBoolAtConstructor. A bool on the current player, the proposer, or a move-named player. */
export function playerBoolAtConstructor(): LegalPredicateConstructor {
  return {
    name: 'playerBoolAt',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 2) {
        specError(`legal: playerBoolAt requires 2 args (player path, want), got ${args.length}`);
      }
      const path = args[0];
      if (!path.startsWith('player.') && !path.startsWith('proposer.') && !path.startsWith('players[move.')) {
        specError(`legal: playerBoolAt path ${goQuote(path)} must select a player via CurrentPlayer, Proposer, or PlayerFromMove`);
      }
      const parsed = playerBoolWant(args[1]);
      if (!parsed.ok) {
        specError(`legal: playerBoolAt: arg 2 (want) ${parsed.error}`);
      }
      const want = parsed.value;
      const template = specTemplate(spec, TEMPLATE_PLAYER_BOOL);
      return newLegalPredicate({
        name: 'playerBoolAt',
        clientEvaluable: true,
        usesProposer: path.startsWith('proposer.'),
        args: [...args],
        reads: [{ path, facet: 'values' }],
        requiredReadTypes: { [path]: 'TypeBool' },
        cost: LEGAL_COST_TRIVIAL,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['prop', 'want'] },
        evaluate(ctx) {
          const value = resolveBoolPath(path, ctx);
          if (!value.ok) {
            return unknownFrom(value);
          }
          if (value.value === want) {
            return passVerdict();
          }
          // Unlike playerBool, the binding is the whole path.
          return failT(template, { prop: path, want: String(want) });
        },
      });
    },
  };
}
