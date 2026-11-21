// Client evaluability: could a given viewer have computed a predicate's
// verdict themselves? The TypeScript mirror of Go's facetSurvives
// (legal_path.go), LegalReadEvaluable (legal_evaluable.go), the Visibility
// block a viewer's state carries (state_visibility.go) and the server's
// per-viewer ledger entry (server/api's buildPreconditionEntry).
//
// A predicate declares, per property it reads, the FACET it needs: the
// values, only a count, only which slots are occupied, only the order, or
// only whether it is empty. Sanitization policies hide some facets and keep
// others, so a count check stays evaluable for a viewer who cannot see the
// cards.

import { parseLegalPath } from './path.ts';
import type { LegalVerdictEntry } from './plan.ts';
import { ADMIN_PLAYER_INDEX, LEGAL_FACETS, type LegalFacet, type LegalMessage, type LegalOutcome, type LegalRead } from './types.ts';

/** Go: Policy, by its sanitization tag name. */
export type LegalPolicy = 'visible' | 'order' | 'len' | 'nonempty' | 'hidden';

/**
 * Go: facetSurvives. Whether reading `facet` of a property sanitized with
 * `policy` yields trustworthy data.
 *
 *              values  count  occupancy  order  nonempty
 *   visible      yes    yes      yes      yes     yes
 *   order        no     yes      yes      yes     yes
 *   len          no     yes      no       no      yes
 *   nonempty     no     no       no       no      yes
 *   hidden       no     no       no       no      no
 */
export function facetSurvives(policy: LegalPolicy, facet: LegalFacet): boolean {
  switch (facet) {
    case 'values':
      return policy === 'visible';
    case 'count':
      return policy === 'visible' || policy === 'order' || policy === 'len';
    case 'occupancy':
    case 'order':
      return policy === 'visible' || policy === 'order';
    case 'nonempty':
      return policy === 'visible' || policy === 'order' || policy === 'len' || policy === 'nonempty';
    default:
      return false;
  }
}

/** Go: propertyVisibility for one property. The facets a policy leaves trustworthy, in Go's order. */
export function visibleFacets(policy: LegalPolicy): LegalFacet[] {
  return LEGAL_FACETS.filter(facet => facetSurvives(policy, facet));
}

/**
 * Go: StateVisibility, exactly as a viewer's state JSON carries it under
 * "Visibility": per property, the facets that viewer may trust.
 */
export interface LegalVisibility {
  Game: Readonly<Record<string, readonly string[]>>;
  Players: readonly Readonly<Record<string, readonly string[]>>[];
}

/**
 * Go: state.visibleFacets(transformation) for a fresh (not already
 * sanitized) state. Builds the Visibility block from per-property policies;
 * a property with no policy is visible.
 */
export function legalVisibilityFromPolicies(
  gameProps: readonly string[],
  gamePolicies: Readonly<Record<string, LegalPolicy>>,
  players: readonly { props: readonly string[]; policies: Readonly<Record<string, LegalPolicy>> }[],
): LegalVisibility {
  const facetsFor = (props: readonly string[], policies: Readonly<Record<string, LegalPolicy>>) =>
    Object.fromEntries(
      props.map(prop => [prop, visibleFacets(Object.prototype.hasOwnProperty.call(policies, prop) ? policies[prop] : 'visible')]),
    );
  return {
    Game: facetsFor(gameProps, gamePolicies),
    Players: players.map(player => facetsFor(player.props, player.policies)),
  };
}

/** A viewer's view of a state, as far as evaluability is concerned. */
export interface LegalViewerView {
  /** The player index the state was sanitized for. */
  viewer: number;
  /** The view's Visibility block, or null when it carries none (nothing is then evaluable). */
  visibility: LegalVisibility | null;
  /** The state's current player index, which player.X reads resolve against. */
  currentPlayerIndex: number;
}

function survives(props: Readonly<Record<string, readonly string[]>>, prop: string, facet: LegalFacet): boolean {
  // Go: legalPolicyForProp treats a property with no policy as visible. The
  // Visibility block lists every property a state has, so "absent" means the
  // property does not exist, and the predicate reading it is unknown anyway.
  if (!Object.prototype.hasOwnProperty.call(props, prop)) {
    return true;
  }
  return props[prop].includes(facet);
}

/**
 * Go: LegalReadEvaluable. Whether `read`'s facet survives the viewer's
 * sanitization.
 *
 *   move.X                     always: the proposer supplies move fields
 *   an admin viewer            always: admin is omniscient
 *   game.X                     the game property's facets
 *   player.X                   the CURRENT player's property
 *   proposer.X, players[*].X,
 *   players[move.F].X          must survive for EVERY player
 *
 * A malformed path or a view with no Visibility fails closed.
 */
export function legalReadEvaluable(view: LegalViewerView, read: LegalRead): boolean {
  const parsed = parseLegalPath(read.path);
  if (!parsed.ok) {
    return false;
  }
  if (parsed.value.kind === 'move') {
    return true;
  }
  if (view.viewer === ADMIN_PLAYER_INDEX) {
    return true;
  }
  const visibility = view.visibility;
  if (visibility === null) {
    return false;
  }
  switch (parsed.value.kind) {
    case 'game':
      return survives(visibility.Game, parsed.value.prop, read.facet);
    case 'player': {
      const current = view.currentPlayerIndex;
      if (current < 0 || current >= visibility.Players.length) {
        return false;
      }
      return survives(visibility.Players[current], parsed.value.prop, read.facet);
    }
    default:
      if (visibility.Players.length === 0) {
        return false;
      }
      return visibility.Players.every(player => survives(player, parsed.value.prop, read.facet));
  }
}

/** Go: LegalReadsEvaluable. Every read must be evaluable. */
export function legalReadsEvaluable(view: LegalViewerView, reads: readonly LegalRead[]): boolean {
  return reads.every(read => legalReadEvaluable(view, read));
}

/**
 * Whether a client could reproduce one ledger entry's verdict: the generic
 * client catalog knows the predicate, it has a serialized form, and every
 * read survives the viewer's sanitization. (Go: buildPreconditionEntry's
 * `evaluable`.)
 */
export function legalEntryEvaluable(view: LegalViewerView, entry: LegalVerdictEntry): boolean {
  return entry.clientEvaluable && entry.serializable && legalReadsEvaluable(view, entry.reads);
}

/**
 * The wire shape of one ledger line. Structurally the PreconditionEntry of
 * src/types/api.d.ts (Go: server/api's preconditionEntry).
 */
export interface LegalPreconditionEntry {
  name: string;
  args?: string[];
  verdict: LegalOutcome;
  message?: LegalMessage;
  evaluable: boolean;
  provisional?: boolean;
}

/**
 * Go: server/api's buildPreconditionEntry. One ledger entry as the server
 * would ship it to `view.viewer`. Bindings are derived from state, so they
 * ship only when the viewer could have derived the verdict themselves; an
 * entry that is not evaluable carries its template key alone.
 */
export function buildPreconditionEntry(view: LegalViewerView, entry: LegalVerdictEntry): LegalPreconditionEntry {
  const evaluable = legalEntryEvaluable(view, entry);
  const out: LegalPreconditionEntry = { name: entry.name, verdict: entry.verdict.outcome, evaluable };
  if (entry.args.length > 0) {
    out.args = [...entry.args];
  }
  if (entry.fieldDependent) {
    out.provisional = true;
  }
  const message = entry.verdict.message;
  if (message) {
    out.message =
      evaluable && message.bindings && Object.keys(message.bindings).length > 0
        ? { template: message.template, bindings: message.bindings }
        : { template: message.template };
  }
  return out;
}

/** Go: legalFormFromLedger's Preconditions array. The whole ledger for one viewer. */
export function buildPreconditionLedger(view: LegalViewerView, entries: readonly LegalVerdictEntry[]): LegalPreconditionEntry[] {
  return entries.map(entry => buildPreconditionEntry(view, entry));
}
