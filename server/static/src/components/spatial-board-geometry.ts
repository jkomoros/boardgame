import type { Component, ExpandedStack } from '../types/boardgame-types.js';

export type SpatialBoardKey = string | number;

export interface BoardPiece<Key extends SpatialBoardKey> {
  readonly id: string;
  readonly space: Key;
  readonly stack: ExpandedStack<object, object>;
  readonly slot: number;
  readonly component: Component<object, object>;
}

/** Explicit adapter for the common sized-stack representation used by Go state. */
export function piecesFromSizedStacks<Key extends SpatialBoardKey>(
  stacks: readonly ExpandedStack<object, object>[],
  spaceForSlot: readonly (Key | null)[],
): readonly BoardPiece<Key>[] {
  if (!spaceForSlot.length) fail('piecesFromSizedStacks requires at least one space key');
  const pieces: BoardPiece<Key>[] = [];
  stacks.forEach((stack, stackIndex) => {
    if (!Array.isArray(stack.Components)) fail(`stack ${stackIndex} Components must be an array`);
    if (stack.Components.length !== spaceForSlot.length) {
      fail(`stack ${stackIndex} has ${stack.Components.length} slots but ${spaceForSlot.length} space keys were provided`);
    }
    if (stack.IDs.length !== stack.Components.length) {
      fail(`stack ${stackIndex} has ${stack.IDs.length} IDs but ${stack.Components.length} component slots`);
    }
    stack.Components.forEach((component, slot) => {
      if (!component) return;
      const space = spaceForSlot[slot];
      if (space === null) return;
      const id = stack.IDs[slot];
      if (!id) fail(`stack ${stackIndex} occupied slot ${slot} has no stable ID`);
      pieces.push(Object.freeze({
        id,
        space: space!,
        stack,
        slot,
        component,
      }));
    });
  });
  return Object.freeze(pieces);
}

export interface BoardGeometrySpace<Key extends SpatialBoardKey> {
  readonly key: Key;
  readonly label: string;
  readonly order?: number;
  /** Element carrying the real pointer-hit geometry. Defaults to the keyed region. */
  readonly region: SVGGraphicsElement;
  /** Element whose center receives keyboard focus and focus treatment. */
  readonly focusAnchor?: SVGGraphicsElement;
  /** Element whose center receives pieces and supplies the animation anchor. */
  readonly pieceAnchor?: SVGGraphicsElement;
}

export interface BoardGeometry<Key extends SpatialBoardKey> {
  readonly spaces: readonly BoardGeometrySpace<Key>[];
}

export interface ResolvedBoardGeometrySpace<Key extends SpatialBoardKey>
  extends Required<BoardGeometrySpace<Key>> {}

export interface ResolvedBoardGeometry<Key extends SpatialBoardKey> {
  readonly spaces: readonly ResolvedBoardGeometrySpace<Key>[];
  readonly byKey: ReadonlyMap<Key, ResolvedBoardGeometrySpace<Key>>;
}

const MAX_SVG_BYTES = 2 * 1024 * 1024;
const blockedElements = new Set([
  'script', 'foreignobject', 'iframe', 'object', 'embed', 'audio', 'video', 'style',
  'animate', 'animatemotion', 'animatetransform', 'set', 'mpath', 'link',
]);
const urlAttributes = new Set(['href', 'src']);

function fail(message: string): never {
  throw new Error(`boardgame-spatial-board: ${message}`);
}

function isSafeUrl(value: string): boolean {
  const normalized = value.trim().toLowerCase();
  return normalized.startsWith('#')
    || /^data:image\/(?:png|jpeg|gif|webp);base64,/.test(normalized);
}

function hasUnsafeCssUrl(value: string): boolean {
  if (/[@]import|expression\s*\(/i.test(value)) return true;
  const urls = [...value.matchAll(/url\s*\(\s*(['"]?)(.*?)\1\s*\)/gi)];
  return urls.some(match => !match[2]?.trim().startsWith('#'));
}

/** Parse and sanitize creator-controlled SVG before it enters the shadow DOM. */
export function parseTrustedBoardSvg(source: string): SVGSVGElement {
  if (!source.trim()) fail('SVG response was empty');
  if (new Blob([source]).size > MAX_SVG_BYTES) fail(`SVG exceeds the ${MAX_SVG_BYTES}-byte limit`);
  if (/<!doctype/i.test(source)) fail('SVG document types are not allowed');
  const document = new DOMParser().parseFromString(source, 'image/svg+xml');
  if (document.querySelector('parsererror')) fail('SVG could not be parsed');
  const root = document.documentElement;
  if (root.namespaceURI !== 'http://www.w3.org/2000/svg' || root.localName !== 'svg') {
    fail('loaded document is not an SVG root');
  }
  const svg = root as unknown as SVGSVGElement;
  const values = (svg.getAttribute('viewBox') ?? '').trim().split(/[\s,]+/).map(Number);
  if (values.length !== 4 || values.some(value => !Number.isFinite(value)) || values[2]! <= 0 || values[3]! <= 0) {
    fail('SVG must have a finite viewBox with positive width and height');
  }

  for (const element of [...svg.querySelectorAll('*')]) {
    if (blockedElements.has(element.localName.toLowerCase())) {
      element.remove();
      continue;
    }
    for (const attribute of [...element.attributes]) {
      const name = attribute.localName.toLowerCase();
      const value = attribute.value;
      if (name.startsWith('on')
        || hasUnsafeCssUrl(value)
        || (urlAttributes.has(name) && !isSafeUrl(value))) {
        element.removeAttributeNode(attribute);
      }
    }
  }
  return svg;
}

function graphicsElement(element: Element | undefined, description: string): SVGGraphicsElement {
  if (!(element instanceof SVGGraphicsElement)) fail(`${description} must identify an SVG graphics element`);
  try {
    const box = element.getBBox();
    if (![box.x, box.y, box.width, box.height].every(Number.isFinite) || (box.width <= 0 && box.height <= 0)) {
      fail(`${description} has no finite visible geometry`);
    }
  } catch (error) {
    const detail = error instanceof Error && error.message ? `: ${error.message}` : '';
    fail(`${description} geometry could not be measured${detail}`);
  }
  return element;
}

/** Validate an explicit geometry sidecar without conflating region and anchor geometry. */
export function resolveBoardGeometry<Key extends SpatialBoardKey>(
  geometry: BoardGeometry<Key>,
): ResolvedBoardGeometry<Key> {
  if (!geometry.spaces.length) fail('geometry must contain at least one space');
  const keys = new Set<Key>();
  const orders = new Set<number>();
  const resolved = geometry.spaces.map((space, index) => {
    if ((typeof space.key !== 'string' && typeof space.key !== 'number')
      || (typeof space.key === 'string' && !space.key.length)
      || (typeof space.key === 'number' && !Number.isFinite(space.key))) {
      fail(`space ${index} has an invalid key`);
    }
    if (keys.has(space.key)) fail(`duplicate space key ${JSON.stringify(space.key)}`);
    keys.add(space.key);
    const label = space.label.trim();
    if (!label) fail(`space ${JSON.stringify(space.key)} has no accessible label`);
    const order = space.order ?? index;
    if (!Number.isInteger(order) || order < 0) fail(`space ${JSON.stringify(space.key)} has invalid order ${order}`);
    if (orders.has(order)) fail(`duplicate keyboard order ${order}`);
    orders.add(order);
    const region = graphicsElement(space.region, `space ${JSON.stringify(space.key)} region`);
    return Object.freeze({
      key: space.key,
      label,
      order,
      region,
      focusAnchor: graphicsElement(space.focusAnchor ?? region, `space ${JSON.stringify(space.key)} focus anchor`),
      pieceAnchor: graphicsElement(space.pieceAnchor ?? region, `space ${JSON.stringify(space.key)} piece anchor`),
    });
  }).sort((left, right) => left.order - right.order);
  return Object.freeze({ spaces: Object.freeze(resolved), byKey: new Map(resolved.map(space => [space.key, space])) });
}

/** Extract the intended zero-config geometry from data attributes in authored artwork. */
export function geometryFromSvg(svg: SVGSVGElement): BoardGeometry<string> {
  const regions = [...svg.querySelectorAll('[data-board-space]')];
  if (!regions.length) fail('SVG contains no data-board-space regions');
  const focusAnchors = new Map([...svg.querySelectorAll('[data-board-focus-anchor]')]
    .map(element => [element.getAttribute('data-board-focus-anchor') ?? '', element]));
  const pieceAnchors = new Map([...svg.querySelectorAll('[data-board-piece-anchor]')]
    .map(element => [element.getAttribute('data-board-piece-anchor') ?? '', element]));
  return {
    spaces: regions.map((element, index) => {
      const key = element.getAttribute('data-board-space') ?? '';
      const title = element.querySelector(':scope > title')?.textContent ?? '';
      const label = element.getAttribute('data-board-label')
        ?? element.getAttribute('aria-label')
        ?? title;
      const rawOrder = element.getAttribute('data-board-order');
      return {
        key,
        label,
        order: rawOrder === null ? index : Number(rawOrder),
        region: element as SVGGraphicsElement,
        focusAnchor: focusAnchors.get(key) as SVGGraphicsElement | undefined,
        pieceAnchor: pieceAnchors.get(key) as SVGGraphicsElement | undefined,
      };
    }),
  };
}
