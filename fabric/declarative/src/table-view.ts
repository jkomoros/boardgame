/**
 * Turns the view a Common Fabric table releases to one viewer (`fabric/
 * patterns/toolkit/view.ts`, `TableView`) into the frames `<bg-table>`
 * installs. Pure; no DOM.
 *
 * It lives on the element side, not in the pattern, for one reason: a pattern
 * that reshaped the released view would produce a value the table's policy
 * did not release. Handing the released value over untouched keeps the
 * pattern's JSX a pass-through.
 *
 * The shapes are restated structurally so this file imports nothing from
 * `fabric/patterns`.
 */
import type { ChoiceSet, Frame } from './bg-table.js';
import type { WireState } from './expand.js';

export interface TableViewEntry {
  readonly version: number;
  readonly state: WireState;
  readonly move?: { readonly AnimationKey: string; readonly Version: number };
  readonly currentPlayerIndex: number;
  readonly timers?: Readonly<Record<string, { readonly TimeLeft: number }>>;
  readonly outcome?: { readonly finished: boolean; readonly winners: readonly number[] };
  readonly settled?: boolean;
  readonly viewers?: Readonly<Record<string, {
    readonly moveLegality?: Readonly<Record<string, { readonly legalForPlayer?: boolean }>>;
    readonly projectedMoveChoices?: { readonly Sets?: readonly ChoiceSet[] };
  }>>;
}

export interface TableViewJSON {
  readonly gameName?: string;
  readonly phase?: 'seating' | 'playing' | 'finished';
  readonly epoch?: number;
  readonly revision?: number;
  readonly version?: number;
  /**
   * The seat of the principal this view was computed for, or -1 for someone
   * watching. A view is computed for one viewer: every entry is sanitized
   * for this seat and carries only this seat's part under `viewers`.
   */
  readonly viewer?: number;
  /** Seats by the name each occupant chose. A view does not say who holds one. */
  readonly seats?: readonly { readonly label: string }[];
  readonly minPlayers?: number;
  readonly maxPlayers?: number;
  readonly entries?: readonly TableViewEntry[];
}

export function isTableView(value: unknown): value is TableViewJSON {
  return !!value && typeof value === 'object'
    && Array.isArray((value as TableViewJSON).seats)
    && Array.isArray((value as TableViewJSON).entries);
}

/** The seat the view was computed for, or -1 for someone watching. */
export function seatOf(view: TableViewJSON | null | undefined): number {
  return typeof view?.viewer === 'number' ? view.viewer : -1;
}

/** One entry as the frame a table installs, for the viewer in `seat`. */
export function frameOf(entry: TableViewEntry, seat: number): Frame {
  const part = entry.viewers?.[String(seat)] ?? entry.viewers?.['-1'];
  return {
    view: {
      CurrentState: entry.state,
      CurrentPlayerIndex: entry.currentPlayerIndex,
      Finished: entry.outcome?.finished ?? false,
      Winners: entry.outcome?.winners ?? [],
      Version: entry.version,
    },
    // An entry with no part for this viewer projects nothing: nothing is legal.
    choices: part?.projectedMoveChoices ?? { Sets: [] },
    legality: part?.moveLegality ?? {},
    transition: { move: entry.move?.AnimationKey ?? '' },
    timers: entry.timers ?? {},
  };
}
