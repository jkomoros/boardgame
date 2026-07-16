import '../client.js';
import type { SelectionDraftBinding } from '../client.js';

declare const draft: SelectionDraftBinding<'clay' | 'ore', 'Trade', { Cards: string }>;
const option = document.createElement('boardgame-selection-option');
option.draft = draft;
option.choice = 'clay';
option.label = 'Clay card';
option.disabled = false;

// @ts-expect-error choices use stable string/number keys
option.choice = { card: 'clay' };
// @ts-expect-error a controller binding, not an arbitrary list, owns interaction
option.draft = ['clay'];
