import './boardgame-market.js';
import type { SelectionDraftSelectionBinding, TargetAction } from '../client.js';

const market = document.createElement('boardgame-market');
declare const action: TargetAction<number>;
declare const stableSelection: SelectionDraftSelectionBinding<string>;
market.label = 'Contracts';
market.sourceLabel = 'Contract deck';
market.displayLabel = 'Face-up contracts';
market.attachmentPosition = 'before';
market.action = action;
market.selection = stableSelection;

// @ts-expect-error attachment position is closed
market.attachmentPosition = 'over';
// @ts-expect-error market actions target visible stack indexes
market.action = {} as TargetAction<string>;
