import './boardgame-game-surface.js';

const surface = document.createElement('boardgame-game-surface');
surface.heading = 'Memory';
surface.headingLevel = 1;
surface.hideHeading = true;

// @ts-expect-error game headings are strings
surface.heading = 4;
// @ts-expect-error heading levels are numeric
surface.headingLevel = '2';
// @ts-expect-error heading visibility is boolean
surface.hideHeading = 'false';
