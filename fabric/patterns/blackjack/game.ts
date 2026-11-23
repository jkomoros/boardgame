/**
 * Blackjack, as this table plays it: the TypeScript engine's port of
 * `examples/blackjack`, generated from the Go by `boardgame-util transpile`
 * (`server/static/src/engine/games/blackjack/`), compiled. Nothing about the
 * rules is written here.
 */
import {
  compileGame,
  type CompiledGame,
} from "../../../server/static/src/engine/rules/index.ts";
import { blackjack } from "../../../server/static/src/engine/games/blackjack/index.ts";

import { infoOf } from "../toolkit/table.ts";
import type { TableInfo } from "../toolkit/view.ts";

/**
 * The compiled game. A function, not a constant: a pattern module may hold
 * only functions and plain data at module scope.
 */
export function blackjackGame(): CompiledGame {
  return compileGame(blackjack());
}

/** Colours for seats, in seat order. */
const SEAT_COLORS = [
  "#D32F2F",
  "#1976D2",
  "#388E3C",
  "#F57C00",
  "#7B1FA2",
  "#00796B",
  "#5D4037",
];

/** The public facts a renderer needs once: the deck and message texts. */
export function blackjackInfo(): TableInfo {
  return infoOf(blackjackGame(), SEAT_COLORS);
}
