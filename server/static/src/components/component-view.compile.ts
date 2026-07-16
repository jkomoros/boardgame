import { cardView, componentView, html, tokenView, type ExpandedStack } from '../client.js';
import { BoardgameComponent } from './boardgame-component.js';

interface CardValues {
  readonly rank: string;
}

type Cards = ExpandedStack<CardValues, Readonly<{ marked: boolean }>>;

const cards = cardView<Cards>({
  render: context => {
    if (context.kind === 'visible') {
      const rank: string = context.component.Values.rank;
      const marked: boolean | undefined = context.component.DynamicValues?.marked;
      return html`${rank}${marked ? '!' : ''}`;
    }
    return null;
  },
  properties: context => ({
    faceUp: context.kind === 'visible',
    rotated: true,
  }),
});

const tokens = tokenView<Cards>({
  properties: context => ({
    color: context.kind === 'visible' ? context.component.Values.rank : '',
  }),
});

const rotatedCards = cards.withProperties({ rotated: true, faceUp: false });
const blueTokens = tokens.withProperties({ color: 'blue', type: 'meeple' });

class CustomPiece extends BoardgameComponent {
  label = '';
}

const custom = componentView<Cards, CustomPiece>(
  () => new CustomPiece(),
  {
    properties: context => ({
      label: context.kind === 'visible' ? context.component.Values.rank : 'Hidden',
    }),
  },
);

void cards;
void tokens;
void custom;
void rotatedCards;
void blueTokens;

cards.withProperties({
  // @ts-expect-error stack-specific properties remain checked against the card host
  faceUpp: true,
});

cards.withProperties({
  // @ts-expect-error stable identity is framework-owned for bound views too
  id: 'creator-owned-id',
});

cardView<Cards>({
  // @ts-expect-error misspelled component properties must fail at author time
  properties: () => ({
    faceUpp: true,
  }),
});

cardView<Cards>({
  // @ts-expect-error stable identity is owned by the stack, not a view recipe
  properties: () => ({ id: 'creator-owned-id' }),
});

cardView<Cards>({
  render: context => {
    if (context.kind !== 'visible') return null;
    // @ts-expect-error generated component values remain exact inside the view
    return context.component.Values.missing;
  },
});
