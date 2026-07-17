export interface RetryPolicy {
  readonly baseDelayMs: number;
  readonly maxDelayMs: number;
  readonly jitterRatio: number;
}

export const LIVE_SESSION_RETRY_POLICY: RetryPolicy = Object.freeze({
  baseDelayMs: 500,
  maxDelayMs: 30_000,
  jitterRatio: 0.2,
});

/**
 * Returns a capped exponential delay with symmetric jitter.
 *
 * `attempt` is zero-based. Supplying random makes the policy deterministic in
 * tests; production callers intentionally use Math.random so clients that
 * disconnect together do not reconnect together.
 */
export function retryDelayMs(
  attempt: number,
  random: () => number = Math.random,
  policy: RetryPolicy = LIVE_SESSION_RETRY_POLICY,
): number {
  if (!Number.isInteger(attempt) || attempt < 0) {
    throw new TypeError(`retry attempt must be a non-negative integer; got ${attempt}`);
  }
  if (!(policy.baseDelayMs > 0) || !(policy.maxDelayMs >= policy.baseDelayMs)) {
    throw new TypeError('retry policy requires 0 < baseDelayMs <= maxDelayMs');
  }
  if (!(policy.jitterRatio >= 0 && policy.jitterRatio <= 1)) {
    throw new TypeError('retry policy jitterRatio must be between 0 and 1');
  }
  const sample = random();
  if (!(sample >= 0 && sample <= 1)) {
    throw new TypeError('retry random sample must be between 0 and 1');
  }

  const exponent = Math.min(attempt, 30);
  const unjittered = Math.min(policy.maxDelayMs, policy.baseDelayMs * (2 ** exponent));
  const multiplier = 1 - policy.jitterRatio + (2 * policy.jitterRatio * sample);
  return Math.max(0, Math.round(unjittered * multiplier));
}
