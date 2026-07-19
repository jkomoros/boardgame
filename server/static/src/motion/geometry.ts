export type GeometrySpace = 'viewport' | 'offset';

/** A DOMRect-independent snapshot whose coordinate space cannot be erased. */
export interface GeometryRect<Space extends GeometrySpace> {
  readonly space: Space;
  readonly top: number;
  readonly left: number;
  readonly width: number;
  readonly height: number;
}

export type ViewportGeometry = GeometryRect<'viewport'>;
export type OffsetGeometry = GeometryRect<'offset'>;

export interface GeometryPoint<Space extends GeometrySpace> {
  readonly space: Space;
  readonly x: number;
  readonly y: number;
}

export type ViewportPoint = GeometryPoint<'viewport'>;

export interface FlipGeometry {
  readonly translateX: number;
  readonly translateY: number;
  readonly scale: number;
  readonly changed: boolean;
}

/** Capture an element in viewport coordinates for overlays and cross-root travel. */
export function captureViewportGeometry(
  element: Pick<HTMLElement, 'getBoundingClientRect'>,
): ViewportGeometry {
  const rect = element.getBoundingClientRect();
  return Object.freeze({
    space: 'viewport',
    top: rect.top,
    left: rect.left,
    width: rect.width,
    height: rect.height,
  });
}

/**
 * Capture the existing FLIP coordinate space by accumulating offset parents.
 * The ancestor itself is included, preserving the animator's established math.
 */
export function captureOffsetGeometry(
  element: HTMLElement,
  ancestorOffsetParent: HTMLElement | null,
): OffsetGeometry {
  let top = 0;
  let left = 0;
  const width = element.offsetWidth;
  const height = element.offsetHeight;
  let current: HTMLElement | null = element;
  while (current) {
    top += current.offsetTop;
    left += current.offsetLeft;
    current = current === ancestorOffsetParent
      ? null
      : current.offsetParent as HTMLElement | null;
  }
  return Object.freeze({ space: 'offset', top, left, width, height });
}

export function geometryCenter(rect: ViewportGeometry): ViewportPoint {
  return Object.freeze({
    space: 'viewport',
    x: rect.left + rect.width / 2,
    y: rect.top + rect.height / 2,
  });
}

/**
 * Return the translation that visually places an element resting at `resting`
 * over `source`, aligned center-to-center—the inversion used by a flight.
 */
export function centeredInversionDelta(
  resting: ViewportGeometry,
  source: ViewportGeometry,
): ViewportPoint {
  const restingCenter = geometryCenter(resting);
  const sourceCenter = geometryCenter(source);
  return Object.freeze({
    space: 'viewport',
    x: sourceCenter.x - restingCenter.x,
    y: sourceCenter.y - restingCenter.y,
  });
}

function finiteScale(numerator: number, denominator: number): number {
  const scale = numerator / denominator;
  return Number.isFinite(scale) && scale !== 0 ? scale : 1;
}

function finiteDelta(value: number): number {
  return Number.isFinite(value) ? value : 0;
}

/** Compile before/after geometry into the structural animator's FLIP inversion. */
export function solveFlipGeometry(
  before: OffsetGeometry,
  after: OffsetGeometry,
  options: Readonly<{
    beforeOrientation: MotionEndpointOrientation;
    afterOrientation: MotionEndpointOrientation;
  }>,
): FlipGeometry {
  const scale = finiteScale(motionAxesDiffer(
    options.beforeOrientation,
    options.afterOrientation,
  ) ? before.height : before.width, after.width);
  const translateY = finiteDelta(
    before.top - after.top - (after.height - before.height) / 2,
  );
  const translateX = finiteDelta(
    before.left - after.left - (after.width - before.width) / 2,
  );
  const changed = Math.abs(translateY) > 0.5
    || Math.abs(translateX) > 0.5
    || Math.abs(scale - 1) > 0.01;
  return Object.freeze({
    translateX,
    translateY,
    scale,
    changed,
  });
}

/** Compose numeric FLIP inversion with component-owned authored transform. */
export function composeFlipTransform(
  geometry: FlipGeometry,
  authoredTransform = '',
): string {
  const translate = `translateY(${geometry.translateY}px) translateX(${geometry.translateX}px)`;
  return `${translate} ${authoredTransform} scale(${geometry.scale})`;
}
import { motionAxesDiffer } from './endpoint-pose.ts';
import type { MotionEndpointOrientation } from './endpoint-pose.ts';
