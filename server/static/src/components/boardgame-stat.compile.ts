import './boardgame-stat.js';
import type { ExpandedStack } from '../types/boardgame-types.js';

const stat = document.createElement('boardgame-stat');
stat.icon = '🍖';
stat.label = 'Food';
stat.value = 12;
stat.value = 'Busted';
stat.value = null;
stat.autoMessage = 'new';
stat.announce = false;
stat.stacked = true;
stat.hideWhenZero = true;

const hidden: boolean = stat.suppressed;
void hidden;

declare const food: ExpandedStack;
stat.stack = food;
stat.stack = null;

const derived: number | undefined = stat.displayCapacity;
void derived;

// @ts-expect-error a capacity is read off the stack; there is no number form
stat.stack = 6;
// @ts-expect-error stat values are deliberately scalar and display-ready
stat.value = { score: 12 };
// @ts-expect-error only implemented animation policies are accepted
stat.autoMessage = 'explode';
