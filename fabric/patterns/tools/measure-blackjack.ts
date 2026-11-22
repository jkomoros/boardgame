/** The browser script's game of blackjack, for `measure-bytes.mjs`. */
import { blackjackGame } from "../blackjack/game.ts";
import type { TableState } from "../toolkit/table.ts";
import type { Proposal } from "../toolkit/view.ts";
import { seatSees } from "./measure-adapter.ts";

interface Shown {
  Computed: { Players: { HandValue: number }[] };
}

export function scenario() {
  const game = blackjackGame();
  return {
    game,
    seats: 3,
    options: { maxrounds: "2" },
    /** Hit below fifteen, otherwise stand, as whoever has the turn. */
    next(table: TableState): { label: string; seat: number; proposal: Proposal } | undefined {
      for (let seat = 0; seat < table.seats.length; seat++) {
        const view = seatSees(game, table, seat);
        const head = view.entries[view.entries.length - 1];
        if (head === undefined || head.currentPlayerIndex !== seat) continue;
        const state: unknown = head.state;
        const value = (state as Shown).Computed.Players[seat]?.HandValue ?? 0;
        const name = value < 15 ? "Current Player Hit" : "Current Player Stand";
        return {
          label: `${seat} ${value < 15 ? "hit" : "stand"}`,
          seat,
          proposal: { kind: "play", name, arguments: {} },
        };
      }
      return undefined;
    },
  };
}
