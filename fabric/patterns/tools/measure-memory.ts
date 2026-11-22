/** A two-player game of Memory on the ten-card board, for `measure-bytes.mjs`. */
import { memoryGame } from "../memory/game.ts";
import type { TableState } from "../toolkit/table.ts";
import type { Proposal } from "../toolkit/view.ts";
import { seatSees, secretHead } from "./measure-adapter.ts";

interface Secret {
  Game: { HiddenCards: { Indexes: number[] } };
}

export function scenario() {
  const game = memoryGame();
  let turns = 0;
  let second: number | undefined;
  return {
    game,
    seats: 2,
    options: { numcards: "small" },
    /**
     * Whoever has the turn turns over two cards: a pair on every other turn
     * and two that differ on the rest, which they then hide.
     */
    next(table: TableState): { label: string; seat: number; proposal: Proposal } | undefined {
      for (let seat = 0; seat < table.seats.length; seat++) {
        const view = seatSees(game, table, seat);
        const head = view.entries[view.entries.length - 1];
        const mine = head?.viewers[String(seat)];
        if (head === undefined || mine === undefined) continue;
        if (mine.moveLegality["Hide Cards"]?.legalForPlayer) {
          return { label: `${seat} hide`, seat, proposal: { kind: "play", name: "Hide Cards", arguments: {} } };
        }
        if (!mine.moveLegality["Reveal Card"]?.legalForPlayer) continue;
        const reveal = (slot: number) => ({
          label: `${seat} reveal`,
          seat,
          proposal: { kind: "play" as const, name: "Reveal Card", arguments: { CardIndex: String(slot) } },
        });
        if (second !== undefined) {
          const slot = second;
          second = undefined;
          return reveal(slot);
        }
        const hidden = (secretHead(table) as Secret).Game.HiddenCards.Indexes;
        const present = hidden.map((_, slot) => slot).filter((slot) => hidden[slot] >= 0);
        const type = (slot: number) => Math.floor(hidden[slot] / 2);
        const first = present[0];
        if (first === undefined) return undefined;
        const pair = present.find((slot) => slot !== first && type(slot) === type(first));
        const other = present.find((slot) => type(slot) !== type(first));
        second = turns++ % 2 === 0 && other !== undefined ? other : pair;
        return reveal(first);
      }
      return undefined;
    },
  };
}
