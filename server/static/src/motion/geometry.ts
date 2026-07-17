/** A minimal rectangle independent of DOMRect identity and coordinate space. */
export interface GeometryRect {
  readonly top: number;
  readonly left: number;
  readonly width: number;
  readonly height: number;
}

export interface GeometryPoint {
  readonly x: number;
  readonly y: number;
}

export interface FlipGeometry {
  readonly translateX: number;
  readonly translateY: number;
  readonly scale: number;
  readonly changed: boolean;
  readonly invertedTransform: string;
}

/** Capture an element in viewport coordinates for overlays and cross-root travel. */
export function captureViewportGeometry(
  element: Pick<HTMLElement, 'getBoundingClientRect'>,
): GeometryRect {
  const rect = element.getBoundingClientRect();
  return {
    top: rect.top,
    left: rect.left,
    width: rect.width,
    height: rect.height,
  };
}

/**
 * Capture the existing FLIP coordinate space by accumulating offset parents.
 * The ancestor itself is included, preserving the animator's established math.
 */
export function captureOffsetGeometry(
  element: HTMLElement,
  ancestorOffsetParent: HTMLElement | null,
): GeometryRect {
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
  return { top, left, width, height };
}

export function geometryCenter(rect: GeometryRect): GeometryPoint {
  return {
    x: rect.left + rect.width / 2,
    y: rect.top + rect.height / 2,
  };
}

/**
 * Return the translation that visually places an element resting at `resting`
 * over `source`, aligned center-to-center—the inversion used by a flight.
 */
export function centeredInversionDelta(
  resting: GeometryRect,
  source: GeometryRect,
): GeometryPoint {
  const restingCenter = geometryCenter(resting);
  const sourceCenter = geometryCenter(source);
  return {
    x: sourceCenter.x - restingCenter.x,
    y: sourceCenter.y - restingCenter.y,
  };
}

function finiteScale(numerator: number, denominator: number): number {
  const scale = numerator / denominator;
  return Number.isFinite(scale) && scale !== 0 ? scale : 1;
}

/** Compile before/after geometry into the structural animator's FLIP inversion. */
export function solveFlipGeometry(
  before: GeometryRect,
  after: GeometryRect,
  options: Readonly<{
    rotates?: boolean;
    beforeTransform?: string;
  }> = {},
): FlipGeometry {
  const scale = finiteScale(options.rotates ? before.height : before.width, after.width);
  const translateY = before.top - after.top - (after.height - before.height) / 2;
  const translateX = before.left - after.left - (after.width - before.width) / 2;
  const changed = Math.abs(translateY) > 0.5
    || Math.abs(translateX) > 0.5
    || Math.abs(scale - 1) > 0.01;
  const translate = `translateY(${translateY}px) translateX(${translateX}px)`;
  const scaleTransform = `scale(${scale})`;
  return Object.freeze({
    translateX,
    translateY,
    scale,
    changed,
    invertedTransform: `${translate} ${options.beforeTransform ?? ''} ${scaleTransform}`,
  });
}
