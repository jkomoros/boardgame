/**
 * companion-sync implements the client-side cross-screen animation sync
 * estimator (spec §8.4). The Table+Hand companion mode wants the same
 * card to fly across two physical devices at roughly the same wall-clock
 * instant — for V1 we don't promise frame-perfect alignment, just visible
 * coherence.
 *
 * The server stamps every "version-timing" socket message with
 * serverSentAt (ms since epoch, captured at broadcast) and serverPlayAt
 * (serverSentAt + ANIMATION_LEAD_MS, the wall-clock instant clients
 * should target).
 *
 * On the client, we maintain a minimum-wins one-way latency estimator
 * over the last 30 frames. The minimum sample is the closest estimate of
 * pure one-way delivery time (variance only ever adds; never subtracts).
 * That offset is then used to convert serverPlayAt into a local epoch ms
 * value, which game renderers can schedule animation timers against:
 *
 *   const playAtLocal = companionSync.localEquivalent(msg.serverPlayAt);
 *   const delay = Math.max(0, playAtLocal - Date.now());
 *   setTimeout(() => doAnimation(), delay);
 *
 * Limitations (spec §8.4):
 * - Asymmetric routes (phone on cell, projector on Wi-Fi) bias the
 *   minimum-wins estimator, but the bias is consistent per surface,
 *   so each side is self-consistent.
 * - JS GC pauses and background-tab throttling can shift setTimeout
 *   firing by 50-200ms — we accept this in V1.
 * - First state push beats the 30-sample window. Animations play
 *   immediately on state install when fewer than 3 samples are present.
 */

interface VersionTimingMessage {
  version: number;
  serverSentAt: number;
  serverPlayAt: number;
}

class CompanionSyncEstimator {
  private samples: number[] = [];
  private readonly windowSize = 30;
  private readonly minSamplesForEstimate = 3;

  /**
   * ingest is called for each "version-timing" socket message. Records the
   * one-way latency sample (localNow - serverSentAt) into the rolling
   * window.
   */
  ingest(msg: VersionTimingMessage): void {
    const oneWay = Date.now() - msg.serverSentAt;
    this.samples.push(oneWay);
    if (this.samples.length > this.windowSize) {
      this.samples.shift();
    }
  }

  /**
   * minOffset returns the rolling-window minimum one-way latency, or
   * null if not enough samples have been collected to commit to an
   * estimate. Game code that gets null should play animations
   * immediately on state install (V1 graceful degradation).
   */
  minOffset(): number | null {
    if (this.samples.length < this.minSamplesForEstimate) return null;
    return Math.min(...this.samples);
  }

  /**
   * localEquivalent converts a server-side epoch ms timestamp into the
   * local-clock epoch ms instant at which the same wall-clock moment
   * occurs. Falls back to the server timestamp itself if the estimator
   * doesn't have enough samples — which usually means the animation
   * fires immediately on receive (the safe default).
   */
  localEquivalent(serverEpochMs: number): number {
    const offset = this.minOffset();
    if (offset === null) {
      // Not enough samples — return the server timestamp directly.
      // Callers should then use max(0, ts - Date.now()) which will be
      // negative for past timestamps (animations play immediately).
      return serverEpochMs;
    }
    return serverEpochMs + offset;
  }

  /** sampleCount is exposed for diagnostics / tests. */
  sampleCount(): number {
    return this.samples.length;
  }
}

export const companionSync = new CompanionSyncEstimator();

// Test seam (mirrors window.__bgAnimTestHooks): the estimator + play-at
// singletons are otherwise module-private, so the cross-screen sync spec
// (tests/animations/waapi-companion.spec.ts) has no way to drive them
// deterministically across two browser contexts. Cost is two property
// writes at module load; harmless in production.
declare global {
  interface Window {
    __bgCompanionSync?: {
      estimator: CompanionSyncEstimator;
      latestServerPlayAt: () => number | null;
      ingestVersionTiming: (data: VersionTimingMessage) => void;
    };
  }
}
if (typeof window !== 'undefined') {
  window.__bgCompanionSync = {
    estimator: companionSync,
    latestServerPlayAt,
    ingestVersionTiming,
  };
}

// _latestServerPlayAt is the server-anchored play-at instant from the
// most recent version-timing message. Game renderers can read it via
// latestServerPlayAt() and feed it to companionSync.localEquivalent()
// to schedule cross-screen animation playback at a clock-aligned
// instant. Updated by ingestVersionTiming on every state push.
let _latestServerPlayAt: number | null = null;

export function latestServerPlayAt(): number | null {
  return _latestServerPlayAt;
}

/**
 * ingestVersionTiming is the boardgame-game-state-manager hook that
 * forwards inbound "version-timing" socket messages into the estimator.
 * Kept as a free function so the state manager doesn't need to import
 * the singleton directly.
 */
export function ingestVersionTiming(data: VersionTimingMessage): void {
  if (!data || typeof data.serverSentAt !== 'number') {
    return;
  }
  companionSync.ingest(data);
  _latestServerPlayAt = typeof data.serverPlayAt === 'number' ? data.serverPlayAt : null;
}
