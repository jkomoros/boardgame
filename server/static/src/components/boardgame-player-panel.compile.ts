import './boardgame-player-panel.js';

const panel = document.createElement('boardgame-player-panel');
panel.label = 'Ada';
panel.headingLevel = 2;
panel.hideHeading = true;
panel.active = true;
panel.activeLabel = 'Acting';

// @ts-expect-error player labels are strings
panel.label = 2;
// @ts-expect-error heading levels are numeric
panel.headingLevel = '3';
// @ts-expect-error active state is boolean
panel.active = 'true';
