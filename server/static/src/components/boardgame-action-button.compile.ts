import './boardgame-action-button.js';

const button = document.createElement('boardgame-action-button');
button.label = 'Draw a card';
button.action = null;
button.unboundReason = 'Choose a card first';

// @ts-expect-error accessible labels are strings
button.label = 12;
// @ts-expect-error arbitrary objects are not bound move actions
button.action = {};
// @ts-expect-error unbound reasons are strings
button.unboundReason = false;
