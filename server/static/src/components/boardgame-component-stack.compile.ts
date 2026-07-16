import './boardgame-component-stack.js';
import type { StackLayout } from './boardgame-component-stack.js';

const stack = document.createElement('boardgame-component-stack');
const layout: StackLayout = 'fan';
stack.layout = layout;
stack.stack = undefined;
stack.boardCols = 8;
stack.boardRows = 8;

// @ts-expect-error layout is a closed set of implemented algorithms
stack.layout = 'carousel';
// @ts-expect-error board geometry is numeric
stack.boardCols = '8';
// @ts-expect-error stack snapshots have the generated ExpandedStack shape
stack.stack = { Components: [] };
