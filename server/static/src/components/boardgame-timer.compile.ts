import './boardgame-timer.js';

const timer = document.createElement('boardgame-timer');
timer.timer = { ID: 'hide-cards', IsTimer: true };
timer.label = 'Cards hide in';
timer.format = 'clock';
timer.hideProgress = true;

// @ts-expect-error timer display formats are a closed policy
timer.format = 'minutes';
// @ts-expect-error generated timer IDs are strings
timer.timer = { ID: 1, IsTimer: true };
// @ts-expect-error timer marker is the true literal
timer.timer = { ID: 'timer', IsTimer: false };
