import '../client.js';
import type { PlacementDraftBinding } from '../client.js';

declare const placement: PlacementDraftBinding<
  'tile-a' | 'tile-b',
  'north' | 'south',
  'Place Tiles',
  { Placements: string }
>;

const board = document.createElement('boardgame-spatial-board');
board.placementDraft = placement;
board.actionGroup = 'squares';

// @ts-expect-error a plain target list cannot bypass the draft controller
board.placementDraft = ['north', 'south'];
