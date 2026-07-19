export type ComponentMotionTarget = 'host' | 'visual';
export type ComponentMotionProperty = 'transform' | 'opacity';

/**
 * One immutable, single-owner visual channel transition.
 *
 * `host` is reserved for structural position/opacity. `visual` is the
 * component-owned inner presentation surface (for example a card face flip).
 */
export interface ComponentMotionTrack {
  readonly target: ComponentMotionTarget;
  readonly property: ComponentMotionProperty;
  readonly from: string;
  readonly to: string;
}

export interface ComponentMotionTrackInput {
  readonly target: ComponentMotionTarget;
  readonly property: ComponentMotionProperty;
  readonly from: string;
  readonly to: string;
}

function exactTrack(input: ComponentMotionTrackInput): ComponentMotionTrack {
  if (input.target !== 'host' && input.target !== 'visual') {
    throw new Error('component motion target must be host or visual');
  }
  if (input.property !== 'transform' && input.property !== 'opacity') {
    throw new Error('component motion property must be transform or opacity');
  }
  if (typeof input.from !== 'string' || typeof input.to !== 'string') {
    throw new Error('component motion endpoints must be strings');
  }
  const normalize = (value: string): string => {
    if (input.property === 'transform') return value.trim() || 'none';
    const parsed = Number(value);
    if (!Number.isFinite(parsed)) {
      throw new Error('component motion opacity endpoints must be finite');
    }
    return String(Math.min(1, Math.max(0, parsed)));
  };
  return Object.freeze({
    target: input.target,
    property: input.property,
    from: normalize(input.from),
    to: normalize(input.to),
  });
}

/** Copy, deduplicate by owned channel, and discard visual no-ops. */
export function componentMotionTracks(
  inputs: readonly ComponentMotionTrackInput[],
): readonly ComponentMotionTrack[] {
  const channels = new Set<string>();
  const result: ComponentMotionTrack[] = [];
  for (const input of inputs) {
    const track = exactTrack(input);
    if (track.from === track.to) continue;
    const channel = `${track.target}:${track.property}`;
    if (channels.has(channel)) {
      throw new Error(`component motion channel ${channel} has multiple owners`);
    }
    channels.add(channel);
    result.push(track);
  }
  return Object.freeze(result);
}

export interface BaseComponentMotionInput {
  readonly needsHostTransition: boolean;
  readonly invertedTransform: string;
  readonly finalTransform: string;
  readonly beforeOpacity: string;
  readonly finalOpacity: string;
  readonly visualTracks?: readonly ComponentMotionTrackInput[];
}

/** Compile structural host channels and component-owned visual channels once. */
export function compileComponentMotionTracks(
  input: BaseComponentMotionInput,
): readonly ComponentMotionTrack[] {
  const tracks: ComponentMotionTrackInput[] = [];
  if (input.needsHostTransition) {
    tracks.push({
      target: 'host',
      property: 'transform',
      from: input.invertedTransform,
      to: input.finalTransform,
    });
  }
  const beforeOpacity = Number.parseFloat(input.beforeOpacity || '1');
  const afterOpacity = Number.parseFloat(input.finalOpacity || '1');
  if (Number.isFinite(beforeOpacity)
    && Number.isFinite(afterOpacity)
    && Math.abs(beforeOpacity - afterOpacity) > 0.01) {
    tracks.push({
      target: 'host',
      property: 'opacity',
      from: String(beforeOpacity),
      to: String(afterOpacity),
    });
  }
  tracks.push(...(input.visualTracks ?? []));
  return componentMotionTracks(tracks);
}
