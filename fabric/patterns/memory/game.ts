/**
 * Memory, as this table plays it: the TypeScript engine's port of
 * `examples/memory`, generated from the Go by `boardgame-util transpile`
 * (`server/static/src/engine/games/memory/`), compiled.
 *
 * Nothing about the rules is written here. The shuffle, the card ids, the
 * fix-up moves, the hide-cards timer and what each viewer may see are the
 * engine's, which is held to the Go game by recorded games
 * (`parity.test.ts`).
 */
import {
  compileGame,
  type CompiledGame,
} from "../../../server/static/src/engine/rules/index.ts";
import { memory } from "../../../server/static/src/engine/games/memory/index.ts";

import { infoOf } from "../toolkit/table.ts";
import type { TableInfo } from "../toolkit/view.ts";

/**
 * The compiled game. A function, not a constant: a pattern module may hold
 * only functions and plain data at module scope, and a compiled game holds
 * functions. So the game is compiled once in each run that needs it.
 */
export function memoryGame(): CompiledGame {
  return compileGame(memory());
}

/** Colours for seats, in seat order, as the Go server's default palette has them. */
const SEAT_COLORS = [
  "#D32F2F",
  "#1976D2",
  "#388E3C",
  "#F57C00",
  "#7B1FA2",
  "#00796B",
];

/** The public facts the renderer needs once: the deck and message texts. */
export function memoryInfo(): TableInfo {
  return infoOf(memoryGame(), SEAT_COLORS);
}
