import {
  piecesFromSizedStacks,
  type BoardGeometry,
  type BoardGeometryFactory,
  type ExpandedStack,
} from '../client.js';

const keys = ['library', 'study'] as const;
declare const stack: ExpandedStack<object, object>;
declare const libraryRegion: SVGGraphicsElement;
declare const studyRegion: SVGGraphicsElement;

const pieces = piecesFromSizedStacks([stack], keys);
const geometry = {
  spaces: [
    { key: keys[0], label: 'Library', region: libraryRegion },
    { key: keys[1], label: 'Study', region: studyRegion, order: 1 },
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

// @ts-expect-error authored geometry must provide the actual pointer-hit region
const missingRegion: BoardGeometry<'library'> = { spaces: [{ key: 'library', label: 'Library' }] };
void missingRegion;

const misspelledKey: BoardGeometry<(typeof keys)[number]> = {
  // @ts-expect-error space keys remain a literal union rather than widening silently
  spaces: [{ key: 'libary', label: 'Library', region: libraryRegion }],
};
void misspelledKey;
