import {
  isVisibleComponent,
  type Component,
  type GameChest,
  type ExpandedStack,
  type FullGameState,
} from './boardgame-types.js';

type CardValues = { Suit: 'Hearts' | 'Spades' };
type CardDynamic = { FaceUp: boolean };
declare const component: Component<CardValues, CardDynamic> | null;
declare const stack: ExpandedStack<CardValues, CardDynamic>;
declare const state: FullGameState<{ Cards: typeof stack }, { Score: number }>;
declare const stateWithDynamicComponents: FullGameState<
  { Cards: typeof stack },
  { Score: number },
  Record<string, never>,
  Record<string, never>,
  { cards: readonly ({ FaceUp: boolean } | null)[] }
>;
declare const chest: GameChest<
  { cards: readonly { readonly Index: number; readonly Values: CardValues }[] },
  { readonly numCards: 9; readonly friendly: true }
>;
declare const unboundChest: GameChest;

if (isVisibleComponent(component)) {
  component.Values.Suit;
  component.DynamicValues?.FaceUp;
}

// @ts-expect-error Opaque or null components must be narrowed before values are read.
component.Values;
// @ts-expect-error Renderer snapshots are deeply readonly.
state.Game.Cards.Components.push(null);
// @ts-expect-error Nested player state is readonly.
state.Players[0]!.Score = 3;
stateWithDynamicComponents.Components?.cards[0]?.FaceUp;
chest.Decks?.cards[0]?.Values.Suit;
chest.Constants?.numCards;
// @ts-expect-error Static chest entries do not have expanded instance IDs.
chest.Decks?.cards[0]?.ID;
// @ts-expect-error Dynamic component snapshots are deeply readonly.
stateWithDynamicComponents.Components?.cards.push(null);
// @ts-expect-error Unbound framework chest types do not advertise creator deck names.
unboundChest.Decks?.cards;
// @ts-expect-error Generated game constants reject nonexistent names.
chest.Constants?.numCard;
// @ts-expect-error Unbound framework chest types do not advertise creator constant names.
unboundChest.Constants?.numCards;
