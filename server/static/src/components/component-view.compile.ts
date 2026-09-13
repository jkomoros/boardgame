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
    // The art and per-card colour slots, through the typed view -- which is the
    // whole point of them: they are facts about the COMPONENT, so a stylesheet
    // that only sees the deck cannot express them.
    art: context.kind === 'visible' ? `/art/${context.component.Values.rank}.png` : '',
    artFit: 'contain',
    backArt: '/art/back.png',
    frontColor: '#fff5dc',
    inkColor: '#18324a',
    tall: true,
  }),
});

const badFit = cardView<Cards>({
  // @ts-expect-error a card's art fits by covering or by containing, nothing else
  properties: () => ({ artFit: 'fill' }),
});
void badFit;

const artedTokens = tokenView<Cards>({
  properties: () => ({ art: '/art/chip.png', recolorArt: true }),
});
void artedTokens;

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

// Dice satisfy the same host contract without a factory cast.
import { dieView } from '../client.js';
const dice = dieView<ExpandedStack<{ Faces: number[] }, { SelectedFace: number; RollCount: number }>>({
  rollBudget: { durationMs: 900, maxSolidDice: 5 },
  properties: context => ({ symbols: context.kind === 'visible' ? { '6': '★' } : null }),
});
dice.withProperties({ faceNames: { 6: 'Star' } });
// @ts-expect-error identity remains stack-owned for dice
dice.withProperties({ id: 'borrowed' });
// @ts-expect-error stack mode cannot be disabled through a die recipe
dice.withProperties({ stackManaged: false });
// @ts-expect-error stack-managed dice derive their faces from the bound item
dice.withProperties({ faces: [6, 6, 6] });
// @ts-expect-error stack-managed dice derive their selected face from the bound item
dice.withProperties({ selectedFaceIndex: 2 });
// @ts-expect-error the enclosing stack owns activation for stack-managed dice
dice.withProperties({ action: null });
// @ts-expect-error dice do not expose card flips
dice.withProperties({ faceUp: true });

dieView<ExpandedStack<{ Faces: number[] }, { SelectedFace: number; RollCount: number }>>({
  // @ts-expect-error dynamic recipe properties cannot take stack ownership either
  properties: () => ({
    stackManaged: false,
  }),
});
