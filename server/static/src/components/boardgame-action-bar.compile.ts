import './boardgame-action-bar.js';

const bar = document.createElement('boardgame-action-bar');
bar.label = 'Turn actions';
bar.orientation = 'vertical';
bar.alignment = 'start';

// @ts-expect-error orientation is a closed implemented policy
bar.orientation = 'diagonal';
// @ts-expect-error alignment is a closed implemented policy
bar.alignment = 'around';
// @ts-expect-error accessible group labels are strings
bar.label = 12;
