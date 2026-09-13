import { defineRendererFixture } from '../../src/testing/renderer-fixture.js';
import type { GameClientContract } from './_game_renderer.js';
import { moveInputSchemaFingerprint } from './_move_args.js';
import type { State } from './_types.js';

const card = (index: number, type: string) => ({
  Index: index,
  Values: { CardSet: 'fixture', Type: type },
  Deck: 'cards',
  GameName: 'memory',
  ID: `memory-card-${index}`,
});

const first = card(0, 'sun');
const match = card(1, 'sun');
const mismatch = card(2, 'moon');

const stack = (components: readonly (ReturnType<typeof card> | null)[]) => ({
  Deck: 'cards',
  Indexes: components.map(component => component?.Index ?? -1),
  IDs: components.map(component => component?.ID ?? ''),
  IDsLastSeen: Object.fromEntries(components.flatMap(component => component
    ? [[component.ID, 4] as const]
    : [])),
  ShuffleCount: 0,
  Size: components.length,
  MaxSize: components.length,
  GameName: 'memory',
  Components: components,
});

const emptyStack = stack([]);
const hiddenCards = stack([null, null]);
function stateWithVisible(second: ReturnType<typeof card> | null): State {
  return {
    Game: {
      CardSet: 'fixture',
      Cards: stack([null, second ? null : match]),
      CurrentPlayer: 0,
      HiddenCards: hiddenCards,
      HideCardsTimer: { ID: '', IsTimer: true },
      NumCards: 2,
      UnusedCards: emptyStack,
      VisibleCards: stack([first, second]),
      Computed: { CurrentPlayerHasCardsToReveal: true },
    },
    Players: [0, 1].map(() => ({
      CardsLeftToReveal: 1,
      PlayerInactive: false,
      SeatClosed: false,
      SeatFilled: true,
      WonCards: emptyStack,
      Computed: { Color: '', MayBeActive: true },
    })),
  };
}

export const memoryOneRevealedFixtureState = stateWithVisible(null);
export const memoryMatchedFixtureState = stateWithVisible(match);
export const memoryMismatchedFixtureState = stateWithVisible(mismatch);

export const memoryRendererFixture = defineRendererFixture<GameClientContract>({
  tagName: 'boardgame-render-game-memory',
  snapshot: {
    schemaVersion: 1,
    state: memoryOneRevealedFixtureState,
    viewingAsPlayer: 0,
    currentPlayerIndex: 0,
    moveLegality: {
      'Reveal Card': { legalForPlayer: true, legalForAnyone: true },
      'Hide Cards': { legalForPlayer: false, legalForAnyone: false },
    },
    version: 3,
    outcome: { finished: false, winners: [] },
    surface: 'game',
    serverMoveInputSchemaFingerprint: moveInputSchemaFingerprint,
  },
});
