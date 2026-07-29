export type ComponentMotionTarget = 'host' | 'visual';
export type ComponentMotionProperty = 'transform' | 'opacity';
export type ComponentMotionChannel = `${ComponentMotionTarget}:${ComponentMotionProperty}`;

/** One compiled point on a track's timeline. `offset` is WAAPI progress in [0,1]. */
export interface ComponentMotionSample {
  readonly offset: number;
  readonly value: string;
}

/**
 * One immutable, single-owner visual channel transition.
 *
 * `host` is reserved for structural position/opacity. `visual` is the
 * component-owned inner presentation surface (for example a card face flip).
 *
 * Every track compiles to `samples`: at least two, uniformly spaced, spanning
 * [0,1]. An `eased` timeline is a plain two-endpoint transition whose shape the
 * animation kernel is free to ease. A `sampled` timeline already encodes its own
 * timing (a simulated trajectory, say) and must be replayed linearly.
 */
export interface ComponentMotionTrack {
  readonly target: ComponentMotionTarget;
  readonly property: ComponentMotionProperty;
  readonly samples: readonly ComponentMotionSample[];
  readonly timeline: 'eased' | 'sampled';
  /** Value the channel should hold once the animation is finished. */
  readonly resting?: string;
}

/** The classic two-endpoint form: the kernel interpolates and eases between them. */
export interface ComponentMotionEndpointInput {
  readonly target: ComponentMotionTarget;
  readonly property: ComponentMotionProperty;
  readonly from: string;
  readonly to: string;
}

/**
 * A trajectory described as a pure function of progress. Sampled at compile time
 * into `resolution` keyframes, so the curve's own timing survives playback.
 */
export type MotionCurveInput = Readonly<{
  curve: (progress: number) => string;
  /** Number of samples; clamped to [2, 256]. Defaults to 64. */
  resolution?: number;
  /** Value to hold after the animation. Defaults to `curve(1)`. */
  resting?: string;
}>;

/** Curves are component-owned presentation only; the host channel stays structural. */
export type ComponentMotionCurveInput = Readonly<{
  target: 'visual';
  property: ComponentMotionProperty;
}> & MotionCurveInput;

/**
 * An already-compiled track is itself valid input: components plan their own
 * track list and the animator re-runs the whole plan through the compiler for
 * ownership and immutability checks before playing it.
 */
export type ComponentMotionTrackInput =
  | ComponentMotionEndpointInput
  | ComponentMotionCurveInput
  | ComponentMotionTrack;

/** Component subclasses may describe only their component-owned visual surface. */
export type VisualMotionTrackInput =
  | (ComponentMotionEndpointInput & Readonly<{ target: 'visual' }>)
  | ComponentMotionCurveInput;

const MIN_CURVE_RESOLUTION = 2;
const MAX_CURVE_RESOLUTION = 256;
const DEFAULT_CURVE_RESOLUTION = 64;
/**
 * Offsets are produced as `index / (count - 1)`, so a re-validated track's
 * offsets must match that division to within ordinary double rounding. One
 * part in a million is far tighter than any real spacing (1/255 at the maximum
 * resolution) and far looser than the ~1e-16 error the division introduces.
 */
const OFFSET_EPSILON = 1e-6;

/**
 * A track that never changes value claims sole ownership of a channel and then
 * holds it still, which is a producer bug rather than a no-op worth tolerating.
 * Endpoint (`eased`) tracks are exempt: `componentMotionTracks` elides those,
 * and the FLIP compiler relies on that elision.
 */
function isConstant(samples: readonly ComponentMotionSample[]): boolean {
  return samples.every(sample => sample.value === samples[0].value);
}

export function componentMotionChannel(
  track: Pick<ComponentMotionTrack, 'target' | 'property'>,
): ComponentMotionChannel {
  return `${track.target}:${track.property}`;
}

function exactTrack(input: ComponentMotionTrackInput): ComponentMotionTrack {
  if (input.target !== 'host' && input.target !== 'visual') {
    throw new Error('component motion target must be host or visual');
  }
  if (input.property !== 'transform' && input.property !== 'opacity') {
    throw new Error('component motion property must be transform or opacity');
  }
  const normalize = (value: string): string => {
    if (input.property === 'transform') return value.trim() || 'none';
    const parsed = Number(value);
    if (!Number.isFinite(parsed)) {
      throw new Error('component motion opacity values must be finite');
    }
    return String(Math.min(1, Math.max(0, parsed)));
  };

  if ('samples' in input && 'curve' in input) {
    // The input union makes this unrepresentable, but excess-property checking
    // does not catch it through a variable. Silently dropping the curve would
    // animate something the producer never asked for.
    throw new Error('component motion input cannot carry both samples and a curve');
  }

  if ('samples' in input) {
    // The animator feeds the overridable planMotionTracks hook's output straight
    // back through here, so this branch validates untrusted input. It must be at
    // least as strict as the compiler below, or an ill-formed plan reaches
    // element.animate and throws at playback time instead of at planning time.
    if (!Array.isArray(input.samples) || input.samples.length < 2) {
      throw new Error('component motion tracks need at least two samples');
    }
    if (input.timeline !== 'eased' && input.timeline !== 'sampled') {
      throw new Error('component motion timeline must be eased or sampled');
    }
    if (input.timeline === 'eased' && input.samples.length !== 2) {
      throw new Error('eased component motion tracks must have exactly two samples');
    }
    if (input.timeline === 'sampled' && input.target !== 'visual') {
      throw new Error('component motion curves are not allowed on the host channel');
    }
    // A resting value is a claim to write the channel's INLINE STYLE after the
    // animation, and on the host channel that surface is not the track's to
    // claim: `boardgame-animatable-item` owns the host element's transform
    // through its `layoutTransform` setter, which early-returns when the value
    // it is handed equals the one it last wrote. So an inline write from here
    // would not merely race the setter -- it would win permanently, because the
    // setter believes the element already carries the value it is displaying
    // and never writes it again. Rejected for the same reason a sampled
    // timeline is above: the host channel is structural, and only the framework
    // may say where a component sits.
    if (input.resting !== undefined && input.target !== 'visual') {
      throw new Error('component motion resting values are not allowed on the host channel');
    }
    const samples = input.samples.map(sample => {
      if (!Number.isFinite(sample?.offset) || sample.offset < 0 || sample.offset > 1) {
        throw new Error('component motion sample offsets must lie in [0,1]');
      }
      if (typeof sample.value !== 'string') {
        throw new Error('component motion sample values must be strings');
      }
      return Object.freeze({ offset: sample.offset, value: normalize(sample.value) });
    });
    for (let index = 1; index < samples.length; index++) {
      if (samples[index].offset <= samples[index - 1].offset) {
        throw new Error('component motion sample offsets must strictly increase');
      }
    }
    if (samples[0].offset !== 0 || samples[samples.length - 1].offset !== 1) {
      throw new Error('component motion samples must span [0,1]');
    }
    for (let index = 0; index < samples.length; index++) {
      const expected = index / (samples.length - 1);
      if (Math.abs(samples[index].offset - expected) > OFFSET_EPSILON) {
        throw new Error('component motion samples must be uniformly spaced');
      }
    }
    if (input.timeline === 'sampled' && isConstant(samples)) {
      throw new Error('component motion curve is constant and animates nothing');
    }
    return Object.freeze({
      target: input.target,
      property: input.property,
      samples: Object.freeze(samples),
      timeline: input.timeline,
      ...(input.resting === undefined ? {} : { resting: normalize(input.resting) }),
    });
  }

  if ('curve' in input) {
    if (input.target !== 'visual') {
      throw new Error('component motion curves are not allowed on the host channel');
    }
    if (typeof input.curve !== 'function') {
      throw new Error('component motion curve must be a function of progress');
    }
    const supplied = input.resolution ?? DEFAULT_CURVE_RESOLUTION;
    if (typeof supplied !== 'number') {
      throw new Error('component motion curve resolution must be a number');
    }
    // Resolution is clamped, never rejected. NaN carries no magnitude to clamp
    // toward, so it falls back to the default the same way an absent value does.
    const requested = Number.isNaN(supplied) ? DEFAULT_CURVE_RESOLUTION : supplied;
    const count = Math.min(
      MAX_CURVE_RESOLUTION,
      Math.max(MIN_CURVE_RESOLUTION, Math.round(requested)),
    );
    const samples: ComponentMotionSample[] = [];
    for (let index = 0; index < count; index++) {
      const progress = index / (count - 1);
      const value = input.curve(progress);
      if (typeof value !== 'string') {
        throw new Error('component motion curve must return strings');
      }
      samples.push(Object.freeze({ offset: progress, value: normalize(value) }));
    }
    if (isConstant(samples)) {
      throw new Error('component motion curve is constant and animates nothing');
    }
    return Object.freeze({
      target: input.target,
      property: input.property,
      samples: Object.freeze(samples),
      timeline: 'sampled' as const,
      resting: input.resting === undefined
        ? samples[samples.length - 1].value
        : normalize(input.resting),
    });
  }

  if (typeof input.from !== 'string' || typeof input.to !== 'string') {
    throw new Error('component motion endpoints must be strings');
  }
  return Object.freeze({
    target: input.target,
    property: input.property,
    samples: Object.freeze([
      Object.freeze({ offset: 0, value: normalize(input.from) }),
      Object.freeze({ offset: 1, value: normalize(input.to) }),
    ]),
    timeline: 'eased' as const,
  });
}

/**
 * Sampled tracks carry their own timing, so the kernel's default effect-level
 * easing would time-warp them. `undefined` leaves the kernel's choice alone.
 */
export function componentMotionTrackEasing(
  track: ComponentMotionTrack,
): 'linear' | undefined {
  return track.timeline === 'sampled' ? 'linear' : undefined;
}

/** Copy, deduplicate by owned channel, and discard visual no-ops. */
export function componentMotionTracks(
  inputs: readonly ComponentMotionTrackInput[],
): readonly ComponentMotionTrack[] {
  const channels = new Set<string>();
  const result: ComponentMotionTrack[] = [];
  for (const input of inputs) {
    const track = exactTrack(input);
    // Endpoint no-ops vacate the channel silently; the FLIP compiler relies on
    // this to drop unchanged structural transforms. Sampled tracks never reach
    // here constant — exactTrack already threw, on both the curve and the
    // already-compiled path.
    if (track.timeline === 'eased'
      && track.samples[0].value === track.samples[1].value) continue;
    const channel = componentMotionChannel(track);
    if (channels.has(channel)) {
      throw new Error(`component motion channel ${channel} has multiple owners`);
    }
    channels.add(channel);
    result.push(track);
  }
  return Object.freeze(result);
}

/** Compile the exact WAAPI keyframes for one already-validated track. */
export function componentMotionKeyframes(
  track: ComponentMotionTrack,
): readonly Readonly<Keyframe>[] {
  const frames = track.samples.map(sample => (track.property === 'transform'
    ? { offset: sample.offset, transform: sample.value }
    : { offset: sample.offset, opacity: sample.value }));
  return Object.freeze(frames.map(frame => Object.freeze(frame)));
}

export interface BaseComponentMotionInput {
  readonly needsHostTransition: boolean;
  readonly invertedTransform: string;
  readonly finalTransform: string;
  readonly beforeOpacity: string;
  readonly finalOpacity: string;
  readonly visualTracks?: readonly VisualMotionTrackInput[];
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
