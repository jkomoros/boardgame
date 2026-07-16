import './boardgame-game-outcome.js';

const outcome = document.createElement('boardgame-game-outcome');
outcome.finished = true;
outcome.animating = false;
outcome.winners = [0, 2];
outcome.winnerLabels = ['Ada', 'Grace'];
outcome.viewer = null;
outcome.viewer = 0;

// @ts-expect-error winners are player indexes
outcome.winners = ['0'];
// @ts-expect-error viewer is a player index or the explicit public null
outcome.viewer = 'observer';
// @ts-expect-error winner labels are strings
outcome.winnerLabels = [1];
