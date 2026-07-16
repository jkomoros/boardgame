import '../client.js';
import type { PlacementDraftBinding } from '../client.js';

declare const placement: PlacementDraftBinding<
  'tile-a' | 'tile-b',
  0 | 1 | 2 | 3,
  'Place Tiles',
  { Placements: string }
>;

const board = document.createElement('boardgame-game-board');
board.rows = 2;
board.cols = 2;
board.placementDraft = placement;

// @ts-expect-error a plain target list cannot bypass the placement controller
board.placementDraft = [0, 1, 2, 3];
