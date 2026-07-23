export type MotionPresencePolicy = 'travel-only' | 'scale-fade';

export interface MotionPresenceFacts {
  readonly scale: number;
  readonly opacity: number;
}

/** Compile a semantic collection policy into finite numeric endpoint facts. */
export function compileMotionPresence(
  policy: MotionPresencePolicy,
): MotionPresenceFacts {
  if (policy === 'travel-only') {
    return Object.freeze({ scale: 1, opacity: 1 });
  }
  if (policy !== 'scale-fade') {
    throw new Error('unknown motion presence transition');
  }
  return Object.freeze({ scale: 0.6, opacity: 0 });
}

/** Serialize numeric facts only at the existing host-track composition seam. */
export function motionPresenceHostStyle(facts: MotionPresenceFacts): Readonly<{
  transform: string;
  opacity: string;
}> {
  return Object.freeze({
    transform: facts.scale === 1 ? '' : `scale(${facts.scale})`,
    opacity: String(facts.opacity),
  });
}
