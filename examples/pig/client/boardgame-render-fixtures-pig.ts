import { defineRendererFixture } from '../../src/testing/renderer-fixture.js';
import type { GameClientContract } from './_game_renderer.js';
import { moveInputSchemaFingerprint } from './_move_args.js';
import type { State } from './_types.js';

const die = {
  Index: 0,
  Values: { Faces: [1, 2, 3, 4, 5, 6] },
  Deck: 'dice',
  GameName: 'pig',
  ID: 'pig-die-0',
  DynamicValues: { SelectedFace: 3, Value: 4 },
} as const;

const dieStack = {
  Deck: 'dice',
  Indexes: [0],
  IDs: [die.ID],
  IDsLastSeen: { [die.ID]: 3 },
  ShuffleCount: 0,
  Size: 1,
  MaxSize: 1,
  GameName: 'pig',
  Components: [die],
} as const;

const player = (score: number) => ({
  DieCounted: true,
  Done: false,
  Eliminated: false,
  PlayerInactive: false,
  RoundScore: 4,
  Score: score,
  SeatClosed: false,
  SeatFilled: true,
});

export const pigFixtureState = {
  Game: {
    CurrentPlayer: 0,
    Die: dieStack,
    TargetScore: 100,
  },
  Players: [player(24), player(18)],
  Components: {
    dice: [die.DynamicValues],
  },
} as const satisfies State;

export const pigRendererFixture = defineRendererFixture<GameClientContract>({
  tagName: 'boardgame-render-game-pig',
  snapshot: {
    schemaVersion: 1,
    state: pigFixtureState,
    viewingAsPlayer: 0,
    currentPlayerIndex: 0,
    moveLegality: {
      'Roll Dice': { legalForPlayer: true, legalForAnyone: true },
      'Done Turn': { legalForPlayer: true, legalForAnyone: true },
    },
    version: 3,
    outcome: { finished: false, winners: [] },
    surface: 'game',
    serverMoveInputSchemaFingerprint: moveInputSchemaFingerprint,
    playerPresentations: [
      { playerIndex: 0, label: 'Alice', color: '#7c3aed' },
      { playerIndex: 1, label: 'Bob', color: '#0f766e' },
    ],
  },
});
