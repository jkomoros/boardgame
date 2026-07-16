import '../client.js';
import type { SelectionDraftBinding } from '../client.js';

declare const draft: SelectionDraftBinding<'clay' | 'ore', 'Trade', { Cards: string }>;
const option = document.createElement('boardgame-selection-option');
option.option = draft.option('clay');
option.label = 'Clay card';
option.disabled = false;

// @ts-expect-error the draft's exact candidate union rejects foreign choices
draft.option('wood');
// @ts-expect-error a controller-produced option, not an arbitrary key, owns interaction
option.option = 'clay';
