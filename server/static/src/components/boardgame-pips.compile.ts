import './boardgame-pips.js';

const pips = document.createElement('boardgame-pips');
pips.glyph = '☘';
pips.count = 3;
pips.max = 5;
pips.max = null;
pips.emptyGlyph = '○';
pips.label = 'Luck';
pips.hideWhenZero = true;

const filled: number = pips.filled;
const unfilled: number = pips.unfilled;
const suppressed: boolean = pips.suppressed;
void filled;
void unfilled;
void suppressed;

// @ts-expect-error a count is a number of symbols, never the symbols themselves
pips.count = '☘☘☘';
// @ts-expect-error a capacity is a number or absent, never a stack
pips.max = { Size: 5 };
