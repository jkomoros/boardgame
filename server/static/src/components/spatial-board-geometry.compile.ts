import {
  piecesFromSizedStacks,
  rasterBoardArtwork,
  type BoardGeometry,
  type BoardGeometryFactory,
  type BoardPathOverlay,
  type ExpandedStack,
  type RasterBoardArtwork,
} from '../client.js';
import { BoardgameSpatialBoard } from './boardgame-spatial-board.js';

const keys = ['library', 'study'] as const;
declare const stack: ExpandedStack<object, object>;
declare const libraryRegion: SVGGraphicsElement;
declare const studyRegion: SVGGraphicsElement;

const pieces = piecesFromSizedStacks([stack], keys);
const geometry = {
  spaces: [
    { key: keys[0], label: 'Library', group: 'rooms', region: libraryRegion },
    { key: keys[1], label: 'Study', group: 'rooms', region: studyRegion, order: 1 },
  ],
} satisfies BoardGeometry<(typeof keys)[number]>;
const geometryFactory = ((svg) => ({
  spaces: [...svg.querySelectorAll<SVGGraphicsElement>('[data-board-space]')].map((region, index) => ({
    key: keys[index]!, label: keys[index]!, region,
  })),
})) satisfies BoardGeometryFactory<(typeof keys)[number]>;

void pieces;
void geometry;
void geometryFactory;

const rasterArtwork = rasterBoardArtwork({
  src: 'game-src/example/board.webp',
  viewportAspectRatio: 4 / 3,
  fit: 'contain',
  spaces: [
    {
      key: keys[0],
      label: 'Library',
      group: 'rooms',
      region: { shape: 'circle', center: { x: 0.25, y: 0.4 }, radius: 0.1 },
      pieceAnchor: { x: 0.3, y: 0.5 },
    },
    {
      key: keys[1],
      label: 'Study',
      region: { shape: 'polygon', points: [
        { x: 0.5, y: 0.1 }, { x: 0.9, y: 0.1 }, { x: 0.7, y: 0.8 },
      ] },
    },
  ],
});
const exactRasterArtwork: RasterBoardArtwork<(typeof keys)[number]> = rasterArtwork;
void exactRasterArtwork;
const route = {
  id: 'secret-passage',
  label: 'Secret passage from Library to Study',
  spaces: keys,
  tone: 'secondary',
  width: 5,
} satisfies BoardPathOverlay<(typeof keys)[number]>;

// @ts-expect-error authored geometry must provide the actual pointer-hit region
const missingRegion: BoardGeometry<'library'> = { spaces: [{ key: 'library', label: 'Library' }] };
void missingRegion;

const misspelledKey: BoardGeometry<(typeof keys)[number]> = {
  // @ts-expect-error space keys remain a literal union rather than widening silently
  spaces: [{ key: 'libary', label: 'Library', region: libraryRegion }],
};
void misspelledKey;

rasterBoardArtwork({
  src: '/board.png',
  spaces: [{
    key: 'library',
    label: 'Library',
    // @ts-expect-error hotspot shapes are a closed, discriminated union
    region: { shape: 'ellipse', center: { x: 0.5, y: 0.5 }, radius: 0.2 },
  }],
});

// @ts-expect-error descriptor keys preserve the game's literal space union
const wrongRasterKey: RasterBoardArtwork<(typeof keys)[number]> = rasterBoardArtwork({
  src: '/board.png',
  spaces: [{
    key: 'libary',
    label: 'Library',
    region: { shape: 'circle', center: { x: 0.5, y: 0.5 }, radius: 0.2 },
  }],
});
void wrongRasterKey;

const spatialBoard = new BoardgameSpatialBoard();
spatialBoard.panZoom = true;
spatialBoard.maxZoom = 6;
spatialBoard.actionGroup = 'rooms';
spatialBoard.pathOverlays = [route];
spatialBoard.revealSpace(keys[0]);
spatialBoard.resetViewport();
// @ts-expect-error maxZoom is numeric
spatialBoard.maxZoom = 'far';

const invalidRoute: BoardPathOverlay<(typeof keys)[number]> = {
  id: 'invalid', label: 'Invalid route', spaces: keys,
  // @ts-expect-error route tones are a closed themeable policy
  tone: 'rainbow',
};
void invalidRoute;
