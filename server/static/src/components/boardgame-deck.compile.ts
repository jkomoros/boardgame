import './boardgame-deck.js';
import type { SelectionDraftSelectionBinding, TargetAction } from '../client.js';

const deck = document.createElement('boardgame-deck');
declare const action: TargetAction<number>;
declare const numericSelection: SelectionDraftSelectionBinding<number>;
deck.label = 'Draw pile';
deck.emptyLabel = 'No cards';
deck.action = action;
deck.selection = numericSelection;
deck.action = null;
deck.selection = null;

// @ts-expect-error labels are strings
deck.label = 4;
// @ts-expect-error deck actions target stack indexes
deck.action = {} as TargetAction<string>;
