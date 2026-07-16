import type { ClockSyncMessage, VersionTimingMessage } from '../types/socket-frame.js';
export type { ClockSyncMessage, VersionTimingMessage } from '../types/socket-frame.js';

/**
 * Cross-screen clock estimation and version-bound animation scheduling.
 *
 * The server announces a version and then sends a sibling timing frame. A
 * timing value is meaningful only for that exact (game, version) pair: HTTP
 * version fetches may return several queued bundles, so a mutable "latest"
 * timestamp cannot safely drive them. CompanionAnimationTimeline is the one
 * client-side owner of that association.
 */

export interface VersionAnimationContext {
  version: number;
  startAtMs: number;
  slotDurationMs: number;
  maxAnimationDurationMs: number;
}

export type CompanionSchedule =
  | { kind: 'scheduled'; context: VersionAnimationContext }
  | { kind: 'awaiting-timing'; waitMs: number }
  | { kind: 'immediate' };

const TIMING_GRACE_MS = 200;
// Emergency ceiling only. Normal retention is lifecycle-based: the state
// manager forgets a version after installing/skipping it and resetGame clears
// navigation leftovers, so a throttled but legitimate queue is not evicted
// merely because another surface drains at a different pace.
const MAX_RECORDED_VERSIONS = 4096;

export function usableAnimationContext(
  context: VersionAnimationContext,
  localNow = Date.now(),
  maxFutureWaitMs = 10_000,
): VersionAnimationContext | null {
  const untilStart = context.startAtMs - localNow;
  if (untilStart > maxFutureWaitMs) return null;
  const lateness = Math.max(0, -untilStart);
  const remainingDuration = context.maxAnimationDurationMs - lateness;
  // A late client may still join the shared cycle, but it must only consume
  // the part of the visible-animation budget that remains. Starting a fresh
  // full-duration effect here would spill into the next version's slot.
  if (remainingDuration <= 0) return null;
  if (remainingDuration === context.maxAnimationDurationMs) return context;
  return { ...context, maxAnimationDurationMs: remainingDuration };
}

export class CompanionSyncEstimator {
  private oneWaySamples: number[] = [];
  private clockSamples: Array<{ offset: number; roundTrip: number }> = [];
  private readonly windowSize = 30;
  private readonly minSamplesForEstimate = 3;

  ingest(msg: VersionTimingMessage, localNow = Date.now()): void {
    const oneWay = localNow - msg.serverSentAt;
    this.oneWaySamples.push(oneWay);
    if (this.oneWaySamples.length > this.windowSize) this.oneWaySamples.shift();
  }

  ingestClockSync(msg: ClockSyncMessage, localReceivedAt = Date.now()): void {
    if (!msg || !Number.isFinite(msg.clientSentAt) || !Number.isFinite(msg.serverAt)) return;
    const roundTrip = Math.max(0, localReceivedAt - msg.clientSentAt);
    const localMidpoint = msg.clientSentAt + roundTrip / 2;
    this.clockSamples.push({ offset: localMidpoint - msg.serverAt, roundTrip });
    if (this.clockSamples.length > this.windowSize) this.clockSamples.shift();
  }

  minOffset(): number | null {
    if (this.clockSamples.length >= this.minSamplesForEstimate) {
      return this.clockSamples.reduce((best, sample) =>
        sample.roundTrip < best.roundTrip ? sample : best).offset;
    }
    // One-way fallback when timing-policy frames are available but the
    // optional clock-sync exchange has not produced enough samples yet.
    if (this.oneWaySamples.length < this.minSamplesForEstimate) return null;
    return Math.min(...this.oneWaySamples);
  }

  localEquivalent(serverEpochMs: number): number {
    const offset = this.minOffset();
    return offset === null ? serverEpochMs : serverEpochMs + offset;
  }

  sampleCount(): number {
    return Math.max(this.clockSamples.length, this.oneWaySamples.length);
  }
}

interface RecordedTiming {
  serverPlayAt: number;
  slotDurationMs: number;
  maxAnimationDurationMs: number;
}

export class CompanionAnimationTimeline {
  readonly estimator = new CompanionSyncEstimator();
  private readonly timings = new Map<string, RecordedTiming>();
  private readonly announcements = new Map<string, number>();

  announce(gameID: string, version: number, localNow = Date.now()): void {
    if (!gameID || !Number.isInteger(version)) return;
    this.recordBounded(this.announcements, this.key(gameID, version), localNow);
  }

  ingest(gameID: string, msg: VersionTimingMessage, localNow = Date.now()): void {
    if (!gameID || !this.validTiming(msg)) return;
    this.estimator.ingest(msg, localNow);
    const key = this.key(gameID, msg.version);
    this.recordBounded(this.timings, key, {
      serverPlayAt: msg.serverPlayAt,
      slotDurationMs: msg.slotDurationMs,
      maxAnimationDurationMs: msg.maxAnimationDurationMs,
    });
  }

  ingestClockSync(msg: ClockSyncMessage, localReceivedAt = Date.now()): void {
    this.estimator.ingestClockSync(msg, localReceivedAt);
  }

  /**
   * Resolve the exact version's schedule. A just-announced version waits up to
   * 200ms for its sibling timing frame; missing timing and a cold estimator
   * both deliberately degrade to immediate playback.
   */
  schedule(gameID: string, version: number, localNow = Date.now()): CompanionSchedule {
    const key = this.key(gameID, version);
    const timing = this.timings.get(key);
    if (timing) {
      if (this.estimator.minOffset() === null) return { kind: 'immediate' };
      return {
        kind: 'scheduled',
        context: {
          version,
          startAtMs: this.estimator.localEquivalent(timing.serverPlayAt),
          slotDurationMs: timing.slotDurationMs,
          maxAnimationDurationMs: timing.maxAnimationDurationMs,
        },
      };
    }

    const announcedAt = this.announcements.get(key);
    if (announcedAt !== undefined) {
      const waitMs = TIMING_GRACE_MS - (localNow - announcedAt);
      if (waitMs > 0) return { kind: 'awaiting-timing', waitMs };
    }
    return { kind: 'immediate' };
  }

  resetGame(gameID: string): void {
    const prefix = `${gameID}\u0000`;
    for (const key of this.timings.keys()) {
      if (key.startsWith(prefix)) this.timings.delete(key);
    }
    for (const key of this.announcements.keys()) {
      if (key.startsWith(prefix)) this.announcements.delete(key);
    }
  }

  forgetVersion(gameID: string, version: number): void {
    const key = this.key(gameID, version);
    this.timings.delete(key);
    this.announcements.delete(key);
  }

  private key(gameID: string, version: number): string {
    return `${gameID}\u0000${version}`;
  }

  private validTiming(msg: VersionTimingMessage): boolean {
    return !!msg && Number.isInteger(msg.version) &&
      Number.isFinite(msg.serverSentAt) && Number.isFinite(msg.serverPlayAt) &&
      Number.isFinite(msg.slotDurationMs) && msg.slotDurationMs > 0 &&
      Number.isFinite(msg.maxAnimationDurationMs) &&
      msg.maxAnimationDurationMs >= 0 && msg.maxAnimationDurationMs <= msg.slotDurationMs;
  }

  private recordBounded<T>(map: Map<string, T>, key: string, value: T): void {
    // Refresh insertion order when a duplicate frame is received.
    map.delete(key);
    map.set(key, value);
    while (map.size > MAX_RECORDED_VERSIONS) {
      const oldest = map.keys().next().value as string | undefined;
      if (oldest === undefined) break;
      map.delete(oldest);
    }
  }
}

export const companionTimeline = new CompanionAnimationTimeline();
// Kept as a focused clock-estimator export for diagnostics and callers that
// need timestamp conversion without schedule lookup.
export const companionSync = companionTimeline.estimator;

export function ingestVersionTiming(gameID: string, data: VersionTimingMessage): void {
  companionTimeline.ingest(gameID, data);
}

declare global {
  interface Window {
    __bgCompanionSync?: {
      estimator: CompanionSyncEstimator;
      timeline: CompanionAnimationTimeline;
      ingestVersionTiming: (gameID: string, data: VersionTimingMessage) => void;
    };
  }
}

if (typeof window !== 'undefined') {
  window.__bgCompanionSync = {
    estimator: companionSync,
    timeline: companionTimeline,
    ingestVersionTiming,
  };
}
