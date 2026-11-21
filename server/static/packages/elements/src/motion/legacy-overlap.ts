export interface LegacyAnimationOverlapPolicy<M> {
  animationOverlap(fromMove: M | null, toMove: M | null): number;
}

export interface LegacyAnimationOverlapDecision {
  /** True when the renderer overrides the framework's compatibility hook. */
  readonly configured: boolean;
  /** Null means ordinary cycle settlement; otherwise arm this exact delay. */
  readonly delayMs: number | null;
}

export function hasLegacyAnimationOverlap<M>(
  renderer: LegacyAnimationOverlapPolicy<M>,
  defaultHook: LegacyAnimationOverlapPolicy<M>['animationOverlap'],
): boolean {
  return renderer.animationOverlap !== defaultHook;
}

/**
 * Compile the historical state-clock overlap contract without interpreting it
 * as structural-motion progress. The successor is deliberately an input: old
 * renderers commonly select deal cadence from `toMove`.
 */
export function compileLegacyAnimationOverlap<M>(
  renderer: LegacyAnimationOverlapPolicy<M>,
  defaultHook: LegacyAnimationOverlapPolicy<M>['animationOverlap'],
  fromMove: M | null,
  toMove: M | null,
  effectiveAnimationLengthMs: number,
): LegacyAnimationOverlapDecision {
  if (!hasLegacyAnimationOverlap(renderer, defaultHook)) {
    return Object.freeze({ configured: false, delayMs: null });
  }
  const fraction = Math.max(0, Math.min(1, renderer.animationOverlap(fromMove, toMove)));
  return Object.freeze({
    configured: true,
    delayMs: fraction > 0 && fraction < 1
      ? fraction * effectiveAnimationLengthMs
      : null,
  });
}
