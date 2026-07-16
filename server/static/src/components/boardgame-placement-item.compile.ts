import '../client.js';
import type { PlacementDraftBinding } from '../client.js';

declare const draft: PlacementDraftBinding<'tile-a' | 'tile-b', 0 | 1, 'Place', { Placements: string }>;
const item = document.createElement('boardgame-placement-item');
item.item = draft.item('tile-a');
item.label = 'Letter A';
item.disabled = false;

// @ts-expect-error the draft's exact item union rejects foreign items
draft.item('tile-c');
// @ts-expect-error a controller-produced item binding owns interaction
item.item = 'tile-a';
