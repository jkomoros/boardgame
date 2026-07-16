import './boardgame-action-button.js';

const button = document.createElement('boardgame-action-button');
button.label = 'Draw a card';
button.action = null;

// @ts-expect-error accessible labels are strings
button.label = 12;
// @ts-expect-error arbitrary objects are not bound move actions
button.action = {};
