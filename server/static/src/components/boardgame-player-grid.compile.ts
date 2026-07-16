import './boardgame-player-grid.js';

const grid = document.createElement('boardgame-player-grid');
grid.label = 'Opponents';
grid.headingLevel = 3;
grid.hideHeading = true;
grid.emptyLabel = 'Waiting for players';

// @ts-expect-error headings use numeric levels
grid.headingLevel = '2';
// @ts-expect-error player collection labels are strings
grid.label = 4;
// @ts-expect-error empty state labels are strings
grid.emptyLabel = false;
