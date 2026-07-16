import { defineRendererFixture } from '../../src/testing/renderer-fixture.js';
import type { GameClientContract } from './_game_renderer.js';
import { moveInputSchemaFingerprint } from './_move_args.js';
import type { State } from './_types.js';

const token = (index: number, value: 'X' | 'O') => ({
  Index: index,
  Values: { Value: value },
  Deck: 'tokens',
  GameName: 'tictactoe',
  ID: `tictactoe-token-${index}`,
});

const xAtZero = token(0, 'X');
const oAtFour = token(1, 'O');
const xAtEight = token(2, 'X');
const xUnused = token(3, 'X');
const oUnused = token(4, 'O');

const slots = {
  Deck: 'tokens',
  Indexes: [xAtZero.Index, -1, -1, -1, oAtFour.Index, -1, -1, -1, xAtEight.Index],
  IDs: [xAtZero.ID, '', '', '', oAtFour.ID, '', '', '', xAtEight.ID],
  IDsLastSeen: { [xAtZero.ID]: 4, [oAtFour.ID]: 4, [xAtEight.ID]: 4 },
  ShuffleCount: 0,
  Size: 9,
  MaxSize: 9,
  GameName: 'tictactoe',
  Components: [xAtZero, null, null, null, oAtFour, null, null, null, xAtEight],
} as const;

const unusedTokens = (component: ReturnType<typeof token>) => ({
  Deck: 'tokens',
  Indexes: [component.Index],
  IDs: [component.ID],
  IDsLastSeen: { [component.ID]: 4 },
  ShuffleCount: 0,
  GameName: 'tictactoe',
  Components: [component],
});

export const tictactoeFixtureState = {
  Game: {
    CurrentPlayer: 0,
    Phase: 'After First Move',
    Slots: slots,
  },
  Players: [
    {
      PlayerInactive: false,
      SeatClosed: false,
      SeatFilled: true,
      TokenValue: 'X',
      TokensToPlaceThisTurn: 1,
      UnusedTokens: unusedTokens(xUnused),
    },
    {
      PlayerInactive: false,
      SeatClosed: false,
      SeatFilled: true,
      TokenValue: 'O',
      TokensToPlaceThisTurn: 1,
      UnusedTokens: unusedTokens(oUnused),
    },
  ],
} as const satisfies State;

export const tictactoeRendererFixture = defineRendererFixture<GameClientContract>({
  tagName: 'boardgame-render-game-tictactoe',
  snapshot: {
    schemaVersion: 1,
    state: tictactoeFixtureState,
    viewingAsPlayer: 0,
    currentPlayerIndex: 0,
    moveLegality: {
      'Place Token': { legalForPlayer: true, legalForAnyone: true },
    },
    version: 4,
    outcome: { finished: false, winners: [] },
    surface: 'game',
    serverMoveInputSchemaFingerprint: moveInputSchemaFingerprint,
    previewDisabledSpaces: [0, 4, 8],
  },
});
