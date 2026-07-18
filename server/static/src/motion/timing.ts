export type AnimationTimingPolicy =
  | 'version'
  | 'immediate'
  | { localStartAtMs: number };

export interface VersionAnimationContext {
  version: number;
  startAtMs: number;
  slotDurationMs: number;
  maxAnimationDurationMs: number;
}

export type MotionTimingResolution =
  | Readonly<{
    kind: 'play';
    timing: OptionalEffectTiming;
    activeContext: VersionAnimationContext | null;
    expectedSettleMs: number;
  }>
  | Readonly<{
    kind: 'skip';
    reason: 'timing';
  }>;

export function finiteTimingMs(value: unknown): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : 0;
}

function effectiveIterations(value: unknown): number {
  if (value === undefined) return 1;
  return Math.max(0, finiteTimingMs(value));
}

function fillWhileWaiting(fill: FillMode | undefined): FillMode {
  if (fill === 'both' || fill === 'backwards') return fill;
  if (fill === 'forwards') return 'both';
  return 'backwards';
}

export function usableAnimationContext(
  context: VersionAnimationContext,
  localNow = Date.now(),
  maxFutureWaitMs = 10_000,
): VersionAnimationContext | null {
  if (!Number.isFinite(context.startAtMs)
    || !Number.isFinite(context.maxAnimationDurationMs)
    || context.maxAnimationDurationMs <= 0
    || !Number.isFinite(localNow)
    || !Number.isFinite(maxFutureWaitMs)
    || maxFutureWaitMs < 0) return null;
  const untilStart = context.startAtMs - localNow;
  if (untilStart > maxFutureWaitMs) return null;
  const lateness = Math.max(0, -untilStart);
  const remainingDuration = context.maxAnimationDurationMs - lateness;
  if (remainingDuration <= 0) return null;
  if (remainingDuration === context.maxAnimationDurationMs) return context;
  return { ...context, maxAnimationDurationMs: remainingDuration };
}

/**
 * Resolve requested WAAPI timing against a version or local scheduling policy.
 * This function is pure: animation ownership, execution, gating, and settlement
 * remain the caller's responsibility.
 */
export function resolveMotionTiming(
  requested: OptionalEffectTiming,
  options: Readonly<{
    policy?: AnimationTimingPolicy;
    context?: VersionAnimationContext | null;
    nowMs?: number;
    defaults?: OptionalEffectTiming;
    reducedMotion?: boolean;
    postAnimationDelayMs?: number;
  }> = {},
): MotionTimingResolution {
  const defaults = { ...options.defaults };
  if (options.reducedMotion && defaults.duration !== undefined) {
    defaults.duration = 0;
  }
  const timing: OptionalEffectTiming = { ...defaults, ...requested };
  for (const field of ['delay', 'duration', 'endDelay', 'iterations'] as const) {
    const value = timing[field];
    if (typeof value === 'number' && !Number.isFinite(value)) timing[field] = 0;
  }
  if ((options.postAnimationDelayMs ?? 0) > 0 && timing.endDelay === undefined) {
    timing.endDelay = options.postAnimationDelayMs;
  }
  const policy = options.policy ?? 'version';
  const now = options.nowMs ?? Date.now();
  let activeContext: VersionAnimationContext | null = null;

  if (policy === 'version') {
    const context = options.context
      ? usableAnimationContext(options.context, now)
      : null;
    if (context) {
      activeContext = context;
      const localDelay = Math.max(0, finiteTimingMs(timing.delay));
      if (localDelay >= context.maxAnimationDurationMs) {
        return Object.freeze({ kind: 'skip', reason: 'timing' });
      }
      const untilStart = Math.max(0, context.startAtMs - now);
      const requestedDuration = Math.max(0, finiteTimingMs(timing.duration));
      const iterations = effectiveIterations(timing.iterations);
      const requestedEndDelay = Math.max(0, finiteTimingMs(timing.endDelay));
      const afterStagger = context.maxAnimationDurationMs - localDelay;
      const boundedEndDelay = Math.min(requestedEndDelay, afterStagger);
      const availableActiveDuration = Math.max(0, afterStagger - boundedEndDelay);
      const requestedActiveDuration = requestedDuration * iterations;
      const boundedActiveDuration = Math.min(requestedActiveDuration, availableActiveDuration);
      timing.delay = untilStart + localDelay;
      timing.duration = iterations > 0 ? boundedActiveDuration / iterations : 0;
      timing.endDelay = boundedEndDelay;
      if (finiteTimingMs(timing.delay) > 0) timing.fill = fillWhileWaiting(timing.fill);
    }
  } else if (policy !== 'immediate') {
    const localDelay = Math.max(0, finiteTimingMs(timing.delay));
    timing.delay = Math.max(0, finiteTimingMs(policy.localStartAtMs) - now) + localDelay;
    if (finiteTimingMs(timing.delay) > 0) timing.fill = fillWhileWaiting(timing.fill);
  }

  const activeDuration = Math.max(0, finiteTimingMs(timing.duration))
    * effectiveIterations(timing.iterations);
  // WAAPI permits negative delays and end-delays. The watchdog needs the
  // remaining nonnegative wall-clock occupancy, not a sum that can be shorter
  // than the effect's repeated active duration or negative altogether.
  const expectedSettleMs = Math.max(
    0,
    finiteTimingMs(timing.delay) + activeDuration + finiteTimingMs(timing.endDelay),
  );
  return Object.freeze({
    kind: 'play',
    timing,
    activeContext,
    expectedSettleMs,
  });
}
