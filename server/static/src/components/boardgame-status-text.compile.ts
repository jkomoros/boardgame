import './boardgame-status-text.js';

const status = document.createElement('boardgame-status-text');
status.value = 12;
status.value = 'Ready';
status.value = null;
status.autoMessage = 'new';
status.announce = false;

// @ts-expect-error status values are deliberately scalar and display-ready
status.value = { score: 12 };
// @ts-expect-error only implemented animation policies are accepted
status.autoMessage = 'explode';
