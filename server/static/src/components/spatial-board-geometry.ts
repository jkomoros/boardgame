import type { Component, ExpandedStack } from '../types/boardgame-types.js';

export type SpatialBoardKey = string | number;

export interface BoardPiece<Key extends SpatialBoardKey> {
  readonly id: string;
  readonly space: Key;
  readonly stack: ExpandedStack<object, object>;
  readonly slot: number;
  readonly component: Component<object, object>;
}

export type BoardPathTone = 'primary' | 'secondary' | 'danger' | 'muted';

/** A noninteractive, accessible route drawn through known board anchors. */
export interface BoardPathOverlay<Key extends SpatialBoardKey> {
  readonly id: string;
  /** Complete accessible description of what the path means. */
  readonly label: string;
  /** Two or more geometry keys, connected through their piece anchors. */
  readonly spaces: readonly Key[];
  readonly tone?: BoardPathTone;
  readonly width?: number;
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
  /** Optional interaction scope for multi-purpose maps (for example tiles, edges, vertices). */
  readonly group?: string;
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

/** Build custom geometry from the same sanitized SVG that the board displays. */
export type BoardGeometryFactory<Key extends SpatialBoardKey> = (
  svg: SVGSVGElement,
) => BoardGeometry<Key>;

export interface ResolvedBoardGeometrySpace<Key extends SpatialBoardKey>
  extends Omit<Required<BoardGeometrySpace<Key>>, 'group'> {
  readonly group: string | null;
}

export interface ResolvedBoardGeometry<Key extends SpatialBoardKey> {
  readonly spaces: readonly ResolvedBoardGeometrySpace<Key>[];
  readonly byKey: ReadonlyMap<Key, ResolvedBoardGeometrySpace<Key>>;
}

export type RasterArtworkFit = 'contain' | 'cover' | 'fill';

export interface NormalizedBoardPoint {
  readonly x: number;
  readonly y: number;
}

export type NormalizedBoardRegion =
  | { readonly shape: 'circle'; readonly center: NormalizedBoardPoint; readonly radius: number }
  | { readonly shape: 'rect'; readonly x: number; readonly y: number; readonly width: number; readonly height: number }
  | { readonly shape: 'polygon'; readonly points: readonly NormalizedBoardPoint[] };

export interface NormalizedBoardSpace<Key extends SpatialBoardKey> {
  readonly key: Key;
  readonly label: string;
  readonly order?: number;
  readonly group?: string;
  readonly region: NormalizedBoardRegion;
  readonly focusAnchor?: NormalizedBoardPoint;
  readonly pieceAnchor?: NormalizedBoardPoint;
}

export interface RasterBoardArtwork<Key extends SpatialBoardKey> {
  readonly kind: 'raster';
  readonly src: string;
  readonly spaces: readonly NormalizedBoardSpace<Key>[];
  readonly fit: RasterArtworkFit;
  readonly viewportAspectRatio: number | null;
}

export interface RasterBoardArtworkOptions<Key extends SpatialBoardKey> {
  readonly src: string;
  readonly spaces: readonly NormalizedBoardSpace<Key>[];
  readonly fit?: RasterArtworkFit;
  readonly viewportAspectRatio?: number;
}

const rasterFits = new Set<RasterArtworkFit>(['contain', 'cover', 'fill']);

/** Create an immutable normalized-hotspot contract for PNG/JPEG/WebP/etc. artwork. */
export function rasterBoardArtwork<Key extends SpatialBoardKey>(
  options: RasterBoardArtworkOptions<Key>,
): RasterBoardArtwork<Key> {
  if (!options || typeof options !== 'object') fail('raster artwork options must be an object');
  const src = validateRasterSource(options.src);
  const fit = options.fit ?? 'contain';
  if (!rasterFits.has(fit)) fail(`raster artwork has unknown fit ${JSON.stringify(fit)}`);
  const viewportAspectRatio = options.viewportAspectRatio ?? null;
  if (viewportAspectRatio !== null && (!Number.isFinite(viewportAspectRatio) || viewportAspectRatio <= 0)) {
    fail('raster artwork viewportAspectRatio must be a finite positive number');
  }
  if (!Array.isArray(options.spaces) || options.spaces.length === 0) {
    fail('raster artwork requires at least one normalized space');
  }
  if (options.spaces.length > MAX_BOARD_SPACES) fail(`geometry exceeds the ${MAX_BOARD_SPACES}-space limit`);
  const keys = new Set<string>();
  const orders = new Set<string>();
  const spaces = options.spaces.map((space, index) => {
    validateSpaceKey(space.key, index);
    const canonicalKey = String(space.key);
    if (keys.has(canonicalKey)) fail(`duplicate canonical space key ${JSON.stringify(canonicalKey)}`);
    keys.add(canonicalKey);
    if (typeof space.label !== 'string' || !space.label.trim()) {
      fail(`space ${JSON.stringify(space.key)} has no accessible label`);
    }
    const group = validateOptionalGroup(space.group, `space ${JSON.stringify(space.key)}`);
    const order = space.order ?? index;
    if (!Number.isSafeInteger(order) || order < 0) fail(`space ${JSON.stringify(space.key)} has invalid order ${order}`);
    const scopedOrder = `${group ?? ''}\u0000${order}`;
    if (orders.has(scopedOrder)) fail(`duplicate keyboard order ${order} in group ${JSON.stringify(group ?? 'all')}`);
    orders.add(scopedOrder);
    const region = freezeRegion(space.region, space.key);
    const focusAnchor = space.focusAnchor ? freezePoint(space.focusAnchor, `space ${JSON.stringify(space.key)} focusAnchor`) : undefined;
    const pieceAnchor = space.pieceAnchor ? freezePoint(space.pieceAnchor, `space ${JSON.stringify(space.key)} pieceAnchor`) : undefined;
    return Object.freeze({ key: space.key, label: space.label.trim(), order, group, region, focusAnchor, pieceAnchor });
  });
  return Object.freeze({
    kind: 'raster' as const,
    src,
    spaces: Object.freeze(spaces),
    fit,
    viewportAspectRatio,
  });
}

/** Build the in-memory SVG scene consumed by the existing spatial-board pipeline. */
export function rasterArtworkScene<Key extends SpatialBoardKey>(
  artwork: RasterBoardArtwork<Key>,
  intrinsicWidth: number,
  intrinsicHeight: number,
): { readonly svg: SVGSVGElement; readonly geometry: BoardGeometry<Key> } {
  if (!Number.isSafeInteger(intrinsicWidth) || intrinsicWidth <= 0
    || !Number.isSafeInteger(intrinsicHeight) || intrinsicHeight <= 0) {
    fail('raster image must decode to positive safe-integer intrinsic dimensions');
  }
  if (intrinsicWidth > MAX_RASTER_DIMENSION || intrinsicHeight > MAX_RASTER_DIMENSION
    || intrinsicWidth * intrinsicHeight > MAX_RASTER_PIXELS) {
    fail(`raster image exceeds the ${MAX_RASTER_DIMENSION}-pixel dimension or ${MAX_RASTER_PIXELS}-pixel area limit`);
  }
  // Revalidate callers that bypassed rasterBoardArtwork with an unsafe cast.
  const validated = rasterBoardArtwork({
    src: artwork.src,
    spaces: artwork.spaces,
    fit: artwork.fit,
    ...(artwork.viewportAspectRatio === null ? {} : { viewportAspectRatio: artwork.viewportAspectRatio }),
  });
  const namespace = 'http://www.w3.org/2000/svg';
  const outerHeight = intrinsicHeight;
  const outerWidth = (validated.viewportAspectRatio ?? (intrinsicWidth / intrinsicHeight)) * outerHeight;
  if (!Number.isFinite(outerWidth) || outerWidth <= 0) {
    fail('raster artwork viewportAspectRatio overflows the decoded image dimensions');
  }
  const svg = document.createElementNS(namespace, 'svg');
  svg.setAttribute('viewBox', `0 0 ${outerWidth} ${outerHeight}`);
  svg.setAttribute('preserveAspectRatio', 'xMidYMid meet');
  svg.setAttribute('aria-hidden', 'true');
  svg.setAttribute('focusable', 'false');

  const scene = document.createElementNS(namespace, 'svg');
  scene.setAttribute('x', '0');
  scene.setAttribute('y', '0');
  scene.setAttribute('width', String(outerWidth));
  scene.setAttribute('height', String(outerHeight));
  scene.setAttribute('viewBox', `0 0 ${intrinsicWidth} ${intrinsicHeight}`);
  scene.setAttribute('preserveAspectRatio', validated.fit === 'contain'
    ? 'xMidYMid meet'
    : validated.fit === 'cover' ? 'xMidYMid slice' : 'none');
  svg.append(scene);

  const image = document.createElementNS(namespace, 'image');
  image.setAttribute('href', validated.src);
  image.setAttribute('x', '0');
  image.setAttribute('y', '0');
  image.setAttribute('width', String(intrinsicWidth));
  image.setAttribute('height', String(intrinsicHeight));
  image.setAttribute('preserveAspectRatio', 'none');
  image.setAttribute('pointer-events', 'none');
  scene.append(image);

  const geometrySpaces = validated.spaces.map(space => {
    const region = createRegionElement(space.region, intrinsicWidth, intrinsicHeight);
    region.setAttribute('data-space', '');
    if (space.group) region.setAttribute('data-board-group', space.group);
    region.setAttribute('fill', 'transparent');
    region.setAttribute('stroke', 'transparent');
    region.setAttribute('pointer-events', 'all');
    scene.append(region);
    const focusAnchor = space.focusAnchor
      ? createAnchorElement(space.focusAnchor, intrinsicWidth, intrinsicHeight)
      : undefined;
    const pieceAnchor = space.pieceAnchor
      ? createAnchorElement(space.pieceAnchor, intrinsicWidth, intrinsicHeight)
      : undefined;
    if (focusAnchor) scene.append(focusAnchor);
    if (pieceAnchor) scene.append(pieceAnchor);
    return { key: space.key, label: space.label, order: space.order, group: space.group, region, focusAnchor, pieceAnchor };
  });
  return Object.freeze({ svg, geometry: Object.freeze({ spaces: Object.freeze(geometrySpaces) }) });
}

function validateRasterSource(value: unknown): string {
  if (typeof value !== 'string' || !value.trim()) fail('raster artwork src must be a non-empty URL');
  const src = value.trim();
  if (/^\s*javascript:/i.test(src)) fail('raster artwork src uses a forbidden URL protocol');
  if (/^\s*data:/i.test(src) && !/^data:image\/(?:png|jpeg|gif|webp|avif);/i.test(src)) {
    fail('raster artwork data URLs must contain a supported raster image');
  }
  return src;
}

function validateSpaceKey(key: unknown, index: number): void {
  if ((typeof key !== 'string' && typeof key !== 'number')
    || (typeof key === 'string' && !key.length)
    || (typeof key === 'number' && !Number.isFinite(key))) {
    fail(`space ${index} has an invalid key`);
  }
}

function validateOptionalGroup(group: unknown, description: string): string | undefined {
  if (group === undefined) return undefined;
  if (typeof group !== 'string' || !group.trim()) fail(`${description} group must be a non-empty string when provided`);
  if (group !== group.trim()) fail(`${description} group must not have surrounding whitespace`);
  if (group.length > 128 || /[\u0000-\u001f\u007f]/.test(group)) {
    fail(`${description} group must be at most 128 characters without control characters`);
  }
  return group;
}

function freezePoint(point: NormalizedBoardPoint, description: string): Readonly<NormalizedBoardPoint> {
  if (!point || typeof point !== 'object' || !normalized(point.x) || !normalized(point.y)) {
    fail(`${description} must have finite x/y coordinates from 0 through 1`);
  }
  return Object.freeze({ x: point.x, y: point.y });
}

function normalized(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 1;
}

function positiveNormalized(value: unknown): value is number {
  return normalized(value) && value > 0;
}

function freezeRegion(region: NormalizedBoardRegion, key: SpatialBoardKey): NormalizedBoardRegion {
  const description = `space ${JSON.stringify(key)} region`;
  if (!region || typeof region !== 'object') fail(`${description} must be a normalized shape`);
  if (region.shape === 'circle') {
    const center = freezePoint(region.center, `${description} center`);
    if (!positiveNormalized(region.radius)) fail(`${description} radius must be greater than 0 and at most 1`);
    return Object.freeze({ shape: 'circle' as const, center, radius: region.radius });
  }
  if (region.shape === 'rect') {
    if (!normalized(region.x) || !normalized(region.y)
      || !positiveNormalized(region.width) || !positiveNormalized(region.height)
      || region.x + region.width > 1 || region.y + region.height > 1) {
      fail(`${description} rectangle must fit within normalized coordinates`);
    }
    return Object.freeze({ shape: 'rect' as const, x: region.x, y: region.y, width: region.width, height: region.height });
  }
  if (region.shape === 'polygon') {
    if (!Array.isArray(region.points) || region.points.length < 3) fail(`${description} polygon requires at least three points`);
    if (region.points.length > MAX_POLYGON_POINTS) {
      fail(`${description} polygon exceeds the ${MAX_POLYGON_POINTS}-point limit`);
    }
    const points = Object.freeze(region.points.map((point, index) => freezePoint(point, `${description} point ${index}`)));
    const width = Math.max(...points.map(point => point.x)) - Math.min(...points.map(point => point.x));
    const height = Math.max(...points.map(point => point.y)) - Math.min(...points.map(point => point.y));
    if (width <= 0 || height <= 0) fail(`${description} polygon must have positive width and height`);
    const twiceArea = Math.abs(points.reduce((sum, point, index) => {
      const next = points[(index + 1) % points.length]!;
      return sum + point.x * next.y - next.x * point.y;
    }, 0));
    if (twiceArea <= Number.EPSILON) fail(`${description} polygon must enclose a positive area`);
    return Object.freeze({ shape: 'polygon' as const, points });
  }
  fail(`${description} has unknown shape ${JSON.stringify((region as { shape?: unknown }).shape)}`);
}

function createRegionElement(
  region: NormalizedBoardRegion,
  width: number,
  height: number,
): SVGGraphicsElement {
  const namespace = 'http://www.w3.org/2000/svg';
  if (region.shape === 'circle') {
    const element = document.createElementNS(namespace, 'circle');
    element.setAttribute('cx', String(region.center.x * width));
    element.setAttribute('cy', String(region.center.y * height));
    element.setAttribute('r', String(region.radius * Math.min(width, height)));
    return element;
  }
  if (region.shape === 'rect') {
    const element = document.createElementNS(namespace, 'rect');
    element.setAttribute('x', String(region.x * width));
    element.setAttribute('y', String(region.y * height));
    element.setAttribute('width', String(region.width * width));
    element.setAttribute('height', String(region.height * height));
    return element;
  }
  const element = document.createElementNS(namespace, 'polygon');
  element.setAttribute('points', region.points.map(point => `${point.x * width},${point.y * height}`).join(' '));
  return element;
}

function createAnchorElement(point: NormalizedBoardPoint, width: number, height: number): SVGCircleElement {
  const element = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
  element.setAttribute('cx', String(point.x * width));
  element.setAttribute('cy', String(point.y * height));
  element.setAttribute('r', String(Math.max(0.001, Math.min(width, height) / 10000)));
  element.setAttribute('fill', 'transparent');
  element.setAttribute('pointer-events', 'none');
  return element;
}

const MAX_SVG_BYTES = 2 * 1024 * 1024;
const MAX_BOARD_SPACES = 512;
const MAX_POLYGON_POINTS = 256;
const MAX_RASTER_DIMENSION = 16384;
const MAX_RASTER_PIXELS = 100_000_000;
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

  for (const element of [svg, ...svg.querySelectorAll('*')]) {
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
  if (geometry.spaces.length > MAX_BOARD_SPACES) {
    fail(`geometry exceeds the ${MAX_BOARD_SPACES}-space limit`);
  }
  const keys = new Set<string>();
  const orders = new Set<string>();
  const resolved = geometry.spaces.map((space, index) => {
    if ((typeof space.key !== 'string' && typeof space.key !== 'number')
      || (typeof space.key === 'string' && !space.key.length)
      || (typeof space.key === 'number' && !Number.isFinite(space.key))) {
      fail(`space ${index} has an invalid key`);
    }
    const canonicalKey = String(space.key);
    if (keys.has(canonicalKey)) fail(`duplicate canonical space key ${JSON.stringify(canonicalKey)}`);
    keys.add(canonicalKey);
    const label = space.label.trim();
    if (!label) fail(`space ${JSON.stringify(space.key)} has no accessible label`);
    const group = validateOptionalGroup(space.group, `space ${JSON.stringify(space.key)}`);
    const order = space.order ?? index;
    if (!Number.isInteger(order) || order < 0) fail(`space ${JSON.stringify(space.key)} has invalid order ${order}`);
    const scopedOrder = `${group ?? ''}\u0000${order}`;
    if (orders.has(scopedOrder)) fail(`duplicate keyboard order ${order} in group ${JSON.stringify(group ?? 'all')}`);
    orders.add(scopedOrder);
    const region = graphicsElement(space.region, `space ${JSON.stringify(space.key)} region`);
    return Object.freeze({
      key: space.key,
      label,
      order,
      group: group ?? null,
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
  const regionKeys = new Set(regions.map(element => element.getAttribute('data-board-space') ?? ''));
  const anchors = (attribute: 'data-board-focus-anchor' | 'data-board-piece-anchor') => {
    const result = new Map<string, Element>();
    for (const element of svg.querySelectorAll(`[${attribute}]`)) {
      const key = element.getAttribute(attribute) ?? '';
      if (!key) fail(`${attribute} must not be empty`);
      if (!regionKeys.has(key)) fail(`${attribute} references unknown space ${JSON.stringify(key)}`);
      if (result.has(key)) fail(`duplicate ${attribute} for space ${JSON.stringify(key)}`);
      result.set(key, element);
    }
    return result;
  };
  const focusAnchors = anchors('data-board-focus-anchor');
  const pieceAnchors = anchors('data-board-piece-anchor');
  return {
    spaces: regions.map((element, index) => {
      const key = element.getAttribute('data-board-space') ?? '';
      const title = element.querySelector(':scope > title')?.textContent ?? '';
      const label = element.getAttribute('data-board-label')
        ?? element.getAttribute('aria-label')
        ?? title;
      const rawOrder = element.getAttribute('data-board-order');
      const group = element.getAttribute('data-board-group') ?? undefined;
      return {
        key,
        label,
        order: rawOrder === null ? index : Number(rawOrder),
        group,
        region: element as SVGGraphicsElement,
        focusAnchor: focusAnchors.get(key) as SVGGraphicsElement | undefined,
        pieceAnchor: pieceAnchors.get(key) as SVGGraphicsElement | undefined,
      };
    }),
  };
}
