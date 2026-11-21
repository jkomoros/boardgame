/** Facets come from the server's existing sanitization/legality truth table. */
export type StateFacet = 'values' | 'count' | 'occupancy' | 'order' | 'nonempty';

export interface StateVisibility {
  readonly Game: Readonly<Record<string, readonly StateFacet[]>>;
  readonly Players: readonly Readonly<Record<string, readonly StateFacet[]>>[];
}

export type ViewedValue<T> =
  | { readonly known: true; readonly value: T }
  | { readonly known: false };

const UNKNOWN = Object.freeze({ known: false } as const);

/** Missing metadata (including older servers/fixtures) cannot prove knowledge. */
export function facetAvailable(facets: readonly StateFacet[] | undefined, facet: StateFacet): boolean {
  return Array.isArray(facets) && facets.includes(facet);
}

/** Read a game property without mistaking a sanitized default for a real value. */
export function viewGameProp<
  S extends { readonly Game: object; readonly Visibility?: StateVisibility },
  K extends keyof S['Game'] & string,
>(state: S | null | undefined, key: K): ViewedValue<S['Game'][K]> {
  if (!state || !Object.prototype.hasOwnProperty.call(state.Game, key) ||
      !facetAvailable(state.Visibility?.Game[key], 'values')) return UNKNOWN;
  return { known: true, value: (state.Game as S['Game'])[key] };
}

/** Player index is explicit: the viewer and the property owner may differ. */
export function viewPlayerProp<
  S extends { readonly Players: readonly object[]; readonly Visibility?: StateVisibility },
  K extends keyof S['Players'][number] & string,
>(state: S | null | undefined, playerIndex: number, key: K): ViewedValue<S['Players'][number][K]> {
  if (!state || !Number.isInteger(playerIndex) || playerIndex < 0) return UNKNOWN;
  const player = state.Players[playerIndex];
  if (!player || !Object.prototype.hasOwnProperty.call(player, key) ||
      !facetAvailable(state.Visibility?.Players[playerIndex]?.[key], 'values')) return UNKNOWN;
  return { known: true, value: (player as S['Players'][number])[key] };
}
