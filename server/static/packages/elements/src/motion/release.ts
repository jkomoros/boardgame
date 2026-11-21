export interface MotionReleaseDeclaration {
  /** Optional diagnostic name. It is not cross-transition identity. */
  readonly key?: string;
  /** Fraction of each selected primary animation's active interval. */
  readonly progress: number;
  /** Omit to wait for every armed structural primary in the cycle. */
  readonly subjects?: readonly string[];
}

export interface CompiledMotionReleaseDeclaration {
  readonly key?: string;
  readonly progress: number;
  readonly subjects?: readonly string[];
}

export interface MotionReleaseParticipant {
  readonly subjectId: string;
  readonly animation: Animation;
}

/** Accept one release source exactly once for the currently installed cycle. */
export function isCurrentMotionCycleRelease(
  cycleId: unknown,
  activeCycleId: number,
  releasedCycleId: number,
): cycleId is number {
  return Number.isInteger(cycleId)
    && cycleId === activeCycleId
    && cycleId !== releasedCycleId;
}

function nonempty(value: unknown, label: string): string {
  if (typeof value !== 'string' || !value.trim()) {
    throw new Error(`motion release ${label} must be a nonempty string`);
  }
  return value.trim();
}

/** Atomically validate one transition-local queue cutover policy. */
export function compileMotionRelease(
  declaration: MotionReleaseDeclaration,
): CompiledMotionReleaseDeclaration {
  if (!declaration || typeof declaration !== 'object') {
    throw new Error('motion release declaration must be an object');
  }
  if (!Number.isFinite(declaration.progress)
    || declaration.progress <= 0 || declaration.progress >= 1) {
    throw new Error('motion release progress must be finite and strictly between 0 and 1');
  }
  let subjects: readonly string[] | undefined;
  if (declaration.subjects !== undefined) {
    if (!Array.isArray(declaration.subjects) || declaration.subjects.length === 0) {
      throw new Error('motion release subjects must be a nonempty array when supplied');
    }
    const compiled = declaration.subjects.map(subject => nonempty(subject, 'subject ID'));
    if (new Set(compiled).size !== compiled.length) {
      throw new Error('motion release subjects must be unique');
    }
    subjects = Object.freeze(compiled);
  }
  return Object.freeze({
    progress: declaration.progress,
    ...(declaration.key === undefined ? {} : { key: nonempty(declaration.key, 'key') }),
    ...(subjects === undefined ? {} : { subjects }),
  });
}

/** Resolve an optional semantic selection without silently weakening it. */
export function selectMotionReleaseParticipants(
  declaration: CompiledMotionReleaseDeclaration,
  participants: readonly MotionReleaseParticipant[],
): readonly MotionReleaseParticipant[] | null {
  if (declaration.subjects === undefined) {
    return participants.length === 0 ? null : Object.freeze([...participants]);
  }
  const selected: MotionReleaseParticipant[] = [];
  for (const subjectId of declaration.subjects) {
    const matches = participants.filter(participant => participant.subjectId === subjectId);
    if (matches.length !== 1) return null;
    selected.push(matches[0]);
  }
  return Object.freeze(selected);
}

function finiteNumber(value: unknown): number | null {
  return typeof value === 'number' && Number.isFinite(value) ? value : null;
}

/** True only when real WAAPI local time has crossed the requested active fraction. */
export function animationReachedMotionProgress(animation: Animation, progress: number): boolean {
  if (animation.playState === 'finished') return true;
  if (animation.playState === 'idle') return false;
  const currentTime = finiteNumber(animation.currentTime);
  const timing = animation.effect?.getComputedTiming();
  const delay = finiteNumber(timing?.delay);
  const activeDuration = finiteNumber(timing?.activeDuration);
  if (currentTime === null || delay === null || activeDuration === null || activeDuration <= 0) {
    return false;
  }
  return currentTime + 0.5 >= delay + activeDuration * progress;
}

/** One generation-scoped sampler for an exact set of structural primaries. */
export class MotionReleaseMonitor {
  private frame: number | null = null;
  private participants: readonly MotionReleaseParticipant[] = [];
  private progress = 0;
  private release: (() => void) | null = null;

  observe(
    participants: readonly MotionReleaseParticipant[],
    progress: number,
    release: () => void,
  ): void {
    this.clear();
    this.participants = Object.freeze([...participants]);
    this.progress = progress;
    this.release = release;
    this.sample();
  }

  clear(): void {
    if (this.frame !== null) cancelAnimationFrame(this.frame);
    this.frame = null;
    this.participants = [];
    this.release = null;
  }

  private sample = (): void => {
    this.frame = null;
    if (!this.release) return;
    if (this.participants.every(({ animation }) => (
      animationReachedMotionProgress(animation, this.progress)
    ))) {
      const release = this.release;
      this.clear();
      release();
      return;
    }
    // Idle/cancelled participants deliberately fail closed; cycle settlement
    // is the fallback and clear() aborts this sampler on the next generation.
    this.frame = requestAnimationFrame(this.sample);
  };
}

export function motionRelease(
  declaration: MotionReleaseDeclaration,
): CompiledMotionReleaseDeclaration {
  return compileMotionRelease(declaration);
}
