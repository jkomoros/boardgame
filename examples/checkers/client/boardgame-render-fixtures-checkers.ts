import { defineRendererFixture } from '../../src/testing/renderer-fixture.js';
import type { GameClientContract } from './_game_renderer.js';
import { moveInputSchemaFingerprint } from './_move_args.js';
import type { State } from './_types.js';

const redToken = {
  Index: 0,
  Values: { Color: 'Red' },
  DynamicValues: { Crowned: false },
  Deck: 'tokens',
  GameName: 'checkers',
  ID: 'checkers-red-0',
} as const;

const emptyStack = {
  Deck: 'tokens', Indexes: [], IDs: [], IDsLastSeen: {}, ShuffleCount: 0,
  GameName: 'checkers', Components: [],
} as const;

const spaces = Array.from({ length: 64 }, () => null) as (typeof redToken | null)[];
spaces[17] = redToken;

export const checkersFixtureState = {
  Game: {
    CurrentPlayer: 0,
    Phase: 'Playing',
    Spaces: {
      Deck: 'tokens',
      Indexes: spaces.map(component => component?.Index ?? -1),
      IDs: spaces.map(component => component?.ID ?? ''),
      IDsLastSeen: { [redToken.ID]: 3 },
      ShuffleCount: 0,
      Size: 64,
      MaxSize: 64,
      GameName: 'checkers',
      Components: spaces,
    },
    UnusedTokens: emptyStack,
  },
  Players: [0, 1].map(index => ({
    CapturedTokens: emptyStack,
    Color: index === 0 ? 'Red' as const : 'Black' as const,
    FinishedTurn: false,
    PlayerInactive: false,
    SeatClosed: false,
    SeatFilled: true,
  })),
} as const satisfies State;

export const checkersRendererFixture = defineRendererFixture<GameClientContract>({
  tagName: 'boardgame-render-game-checkers',
  snapshot: {
    schemaVersion: 1,
    state: checkersFixtureState,
    viewingAsPlayer: 0,
    currentPlayerIndex: 0,
    moveLegality: {
      'Move Token': { legalForPlayer: true, legalForAnyone: true },
    },
    version: 3,
    outcome: { finished: false, winners: [] },
    surface: 'game',
    serverMoveInputSchemaFingerprint: moveInputSchemaFingerprint,
  },
});
