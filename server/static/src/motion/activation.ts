import type { StructuralExecutedTiming } from './structural-plan.js';

export function primaryStructuralAnimationIndex(
  pathKind: 'stationary' | 'travel' | undefined,
  timings: readonly StructuralExecutedTiming[],
): number | null {
  if (timings.length === 0) return null;
  if (pathKind === 'travel') {
    const spatial = timings.findIndex(timing => timing.channel === 'host:transform');
    if (spatial >= 0) return spatial;
  }
  let primary = 0;
  for (let index = 1; index < timings.length; index++) {
    if (timings[index].delayMs < timings[primary].delayMs) primary = index;
  }
  return primary;
}

interface ActivationObservation {
  readonly animation: Animation;
  readonly delayMs: number;
  readonly activate: () => void;
}

/** One frame sampler for every delayed structural animation owned by an animator. */
export class MotionActivationMonitor {
  private readonly observations = new Map<string, ActivationObservation>();
  private frame: number | null = null;

  observe(
    segmentId: string,
    animation: Animation,
    delayMs: number,
    activate: () => void,
  ): void {
    this.cancel(segmentId);
    if (delayMs <= 0) {
      activate();
      return;
    }
    this.observations.set(segmentId, { animation, delayMs, activate });
    this.ensureFrame();
  }

  cancel(segmentId: string): void {
    this.observations.delete(segmentId);
    if (this.observations.size === 0 && this.frame !== null) {
      cancelAnimationFrame(this.frame);
      this.frame = null;
    }
  }

  clear(): void {
    this.observations.clear();
    if (this.frame !== null) cancelAnimationFrame(this.frame);
    this.frame = null;
  }

  private ensureFrame(): void {
    if (this.frame !== null) return;
    this.frame = requestAnimationFrame(() => this.sample());
  }

  private sample(): void {
    this.frame = null;
    for (const [segmentId, observation] of this.observations) {
      const { animation, delayMs, activate } = observation;
      const currentTime = animation.currentTime;
      if (typeof currentTime === 'number' && currentTime + 0.5 >= delayMs) {
        this.observations.delete(segmentId);
        activate();
      } else if (animation.playState === 'idle' || animation.playState === 'finished') {
        this.observations.delete(segmentId);
      }
    }
    if (this.observations.size > 0) this.ensureFrame();
  }
}
