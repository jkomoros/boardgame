import {
  defineRendererFixture,
  type RendererFixtureSnapshot,
} from '../../src/testing/renderer-fixture.js';
import type { GameClientContract } from './_game_renderer.js';
import { pigRendererFixture } from './boardgame-render-fixtures-pig.js';

const misspelledMoveLegality = {
  'Roll Dice': { legalForPlayer: true, legalForAnyone: true },
  'Done Trun': { legalForPlayer: true, legalForAnyone: true },
};

const invalidMoveSnapshot: RendererFixtureSnapshot<GameClientContract> = {
  ...pigRendererFixture.snapshot,
  // @ts-expect-error Misspelled and missing generated move names must fail fixture compilation.
  moveLegality: misspelledMoveLegality,
};

defineRendererFixture<GameClientContract>({
  // @ts-expect-error A contract-bound fixture cannot name another game's renderer tag.
  tagName: 'boardgame-render-game-tictactoe',
  snapshot: pigRendererFixture.snapshot,
});

defineRendererFixture<GameClientContract>({
  tagName: 'boardgame-render-game-pig-table',
  snapshot: pigRendererFixture.snapshot,
});

defineRendererFixture<GameClientContract>({
  tagName: 'boardgame-render-game-pig-hand',
  snapshot: pigRendererFixture.snapshot,
});

defineRendererFixture<GameClientContract>({
  // @ts-expect-error Only the generated game, table, and hand renderer tags are accepted.
  tagName: 'boardgame-render-game-pig-sideboard',
  snapshot: pigRendererFixture.snapshot,
});

void invalidMoveSnapshot;
