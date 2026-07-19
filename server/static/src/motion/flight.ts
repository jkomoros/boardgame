import {
  centeredInversionDelta,
} from './geometry.ts';
import type {
  FlipGeometry,
  ViewportGeometry,
} from './geometry.ts';
import {
  componentMotionTracks,
} from './component-track.ts';
import type { ComponentMotionTrack } from './component-track.ts';

export interface CompiledViewportFlight {
  readonly source: ViewportGeometry;
  readonly destination: ViewportGeometry;
  readonly inversion: FlipGeometry;
  readonly tracks: readonly ComponentMotionTrack[];
}

/**
 * Compile an arrival carried by an element at its natural destination.
 * `source` contributes geometry only; the destination element owns playback.
 */
export function compileViewportFlight(
  source: ViewportGeometry,
  destination: ViewportGeometry,
  restingTransform = 'none',
): CompiledViewportFlight {
  const delta = centeredInversionDelta(destination, source);
  const changed = Math.abs(delta.x) > 0.5 || Math.abs(delta.y) > 0.5;
  const inversion = Object.freeze({
    translateX: delta.x,
    translateY: delta.y,
    scale: 1,
    changed,
  });
  const resting = restingTransform.trim() || 'none';
  const invertedTransform = resting === 'none'
    ? `translate(${delta.x}px, ${delta.y}px)`
    : `translate(${delta.x}px, ${delta.y}px) ${resting}`;
  const tracks = changed ? componentMotionTracks([{
    target: 'host',
    property: 'transform',
    from: invertedTransform,
    to: resting,
  }]) : Object.freeze([]);
  return Object.freeze({ source, destination, inversion, tracks });
}
