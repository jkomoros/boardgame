import './boardgame-component-zone.js';

const zone = document.createElement('boardgame-component-zone');
zone.label = 'Draw pile';
zone.layout = 'pile';
zone.headingLevel = 3;
zone.stack = undefined;
zone.hideCount = true;

// @ts-expect-error component zones deliberately exclude board geometry layouts
zone.layout = 'board';
// @ts-expect-error component-zone layouts are a closed implemented set
zone.layout = 'carousel';
// @ts-expect-error heading levels are numeric
zone.headingLevel = '2';
// @ts-expect-error labels are strings
zone.label = 12;
