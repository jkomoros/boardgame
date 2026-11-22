/**
 * What `measure-bytes.mjs` needs from the toolkit, in one place, so the
 * script measures whatever shape the toolkit has at a commit.
 */
import type { CompiledGame } from "../../../server/static/src/engine/rules/index.ts";
import {
  held,
  OBSERVER_SEAT,
  seatViewOf,
  tableOf,
  type TableState,
} from "../toolkit/table.ts";
import type { TableView } from "../toolkit/view.ts";

/** The table as `stepTable` takes it. */
export function storedOf(table: TableState | undefined): TableState {
  return tableOf(table);
}

/**
 * The viewers a view is computed for with `seats` people seated and one
 * watcher looking: one instance for each principal who looks.
 */
export function looking(seats: number, gameName: string): number[] {
  void gameName;
  const viewers = [OBSERVER_SEAT];
  for (let seat = 0; seat < seats; seat++) viewers.push(seat);
  return viewers;
}

/** The view computed for `viewer`; nothing before anyone looks at a table. */
export function viewOfTable(
  game: CompiledGame,
  table: TableState | undefined,
  viewer: number,
): TableView | undefined {
  return table === undefined
    ? undefined
    : seatViewOf(game, table, Math.min(viewer, table.seats.length - 1));
}

/** What the player in `seat` sees, for a scenario to choose a move from. */
export function seatSees(
  game: CompiledGame,
  table: TableState,
  seat: number,
): TableView {
  return seatViewOf(game, table, seat);
}

/** The unsanitized head state, which only a measuring script may look at. */
export function secretHead(table: TableState): unknown {
  const all = held(table);
  return all[all.length - 1]?.state;
}
