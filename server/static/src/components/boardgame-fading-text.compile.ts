import './boardgame-fading-text.js';

const fading = document.createElement('boardgame-fading-text');
fading.trigger = true;
fading.trigger = 1.5;
fading.autoMessage = 'diff';
fading.suppress = 'falsey';
fading.announce = false;

// @ts-expect-error callout messages must already be display-ready strings
fading.message = 12;
// @ts-expect-error triggers must already be scalar display state
fading.trigger = { current: true };
// @ts-expect-error only implemented message policies are accepted
fading.autoMessage = 'difference';
// @ts-expect-error only implemented suppression policies are accepted
fading.suppress = 'empty';
