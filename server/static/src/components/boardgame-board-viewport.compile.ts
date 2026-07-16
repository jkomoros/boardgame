import type { BoardViewportChange } from '../client.js';
import { BoardgameBoardViewport } from './boardgame-board-viewport.js';

const viewport = new BoardgameBoardViewport();
viewport.label = 'City map navigation';
viewport.maxScale = 6;
viewport.zoomStep = 0.5;
viewport.hideControls = false;
const view: BoardViewportChange = viewport.view;
viewport.zoomIn();
viewport.zoomOut();
viewport.resetView();
viewport.setView(view);
declare const marker: SVGGraphicsElement;
viewport.reveal(marker, 16);
void view;

// @ts-expect-error maximum scale is numeric
viewport.maxScale = 'large';
// @ts-expect-error reveal targets are DOM elements
viewport.reveal('harbor');
// @ts-expect-error view offsets cannot be strings
viewport.setView({ scale: 2, x: 'left', y: 0 });
