import './boardgame-turn-status.js';
import { ObserverPlayerIndex, type TurnStatusContext } from '../status/turn-status.js';

const context = {
  currentPlayerIndex: 0,
  viewerPlayerIndex: ObserverPlayerIndex,
  finished: false,
  animating: false,
} as const satisfies TurnStatusContext;

const status = document.createElement('boardgame-turn-status');
status.turn = context;
status.playerLabels = ['Ada', 'Grace'];
status.activeLabel = 'Your move';

// @ts-expect-error turn contexts require animation and completion state
status.turn = { currentPlayerIndex: 0, viewerPlayerIndex: 0 };
// @ts-expect-error player labels are strings
status.playerLabels = [1, 2];
// @ts-expect-error active labels are strings
status.activeLabel = false;
