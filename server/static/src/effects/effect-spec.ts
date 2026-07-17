import type { AnimationTimingPolicy } from '../components/boardgame-animatable-item.js';
import type { ClientMove } from '../types/api.js';

export type EffectTone =
  | 'neutral'
  | 'reward'
  | 'confirm'
  | 'attention'
  | 'warning'
  | 'magic';

export type EffectIntensity = 'subtle' | 'small' | 'medium' | 'large';

export interface NamedEffectAnchor {
  readonly kind: 'named';
  readonly name: string;
}

export interface PointEffectAnchor {
  readonly kind: 'point';
  readonly x: number;
  readonly y: number;
}

/** Elements are ideal for immediate interaction feedback; named anchors are replay-safe. */
export type EffectAnchor = NamedEffectAnchor | PointEffectAnchor | HTMLElement;

export interface EffectBase {
  /** Stable identity within a transition or composition. */
  readonly key?: string;
  /** Semantic palette role; independent of the visual recipe. */
  readonly tone?: EffectTone;
  /** Semantic scale; low-level pixel/count tuning belongs in `advanced`. */
  readonly intensity?: EffectIntensity;
  /** Optional author salt added to the framework's game/version/descriptor seed. */
  readonly seedKey?: string;
  /** Immediate for local feedback; version aligns independently-derived companion effects. */
  readonly timing?: AnimationTimingPolicy;
  /** Raw visual overrides are intentionally isolated from the common semantic API. */
  readonly advanced?: Readonly<{
    palette?: readonly string[];
    durationMs?: number;
  }>;
}

export interface BurstEffectSpec extends EffectBase {
  readonly kind: 'burst';
  readonly at: EffectAnchor;
  readonly advanced?: EffectBase['advanced'] & Readonly<{
    count?: number;
    spreadPx?: number;
  }>;
}

export interface PulseEffectSpec extends EffectBase {
  readonly kind: 'pulse';
  readonly at: EffectAnchor;
  readonly advanced?: EffectBase['advanced'] & Readonly<{
    scale?: number;
  }>;
}

export interface TravelEffectSpec extends EffectBase {
  readonly kind: 'travel';
  readonly from: EffectAnchor;
  readonly to: EffectAnchor;
  readonly advanced?: EffectBase['advanced'] & Readonly<{
    arcPx?: number;
    sizePx?: number;
  }>;
}

export interface SequenceEffectSpec extends EffectBase {
  readonly kind: 'sequence';
  readonly effects: readonly EffectSpec[];
  readonly gapMs: number;
}

export interface ParallelEffectSpec extends EffectBase {
  readonly kind: 'parallel';
  readonly effects: readonly EffectSpec[];
}

export type EffectSpec =
  | BurstEffectSpec
  | PulseEffectSpec
  | TravelEffectSpec
  | SequenceEffectSpec
  | ParallelEffectSpec;

export interface EffectTheme {
  readonly tones?: Readonly<Partial<Record<EffectTone, readonly string[]>>>;
}

export type EffectResult =
  | Readonly<{ status: 'finished' }>
  | Readonly<{ status: 'cancelled' }>
  | Readonly<{
    status: 'skipped';
    reason: 'budget' | 'missing-anchor' | 'not-connected' | 'timing';
  }>;

export interface EffectHandle {
  readonly finished: Promise<EffectResult>;
  cancel(): void;
}

export interface EffectHostAPI {
  play(effect: EffectSpec): EffectHandle;
}

interface EffectTransitionContextBase<S, MN extends string> {
  readonly after: S;
  /** Animation-safe metadata only: move name and produced version. */
  readonly move: (Omit<ClientMove, 'Name'> & Readonly<{ Name: MN }>) | null;
  readonly version: number;
  readonly snapshotEpoch: number;
  changed<T>(select: (state: S) => T): boolean;
}

/** Initial installation is distinct so checking `kind` also narrows `before`. */
export type EffectTransitionContext<S, MN extends string = string> =
  | (EffectTransitionContextBase<S, MN> & Readonly<{
    kind: 'initial';
    before: null;
  }>)
  | (EffectTransitionContextBase<S, MN> & Readonly<{
    kind: 'transition';
    before: S;
  }>);

type CommonOptions = Omit<EffectBase, 'kind'>;
type BurstOptions = CommonOptions & Pick<BurstEffectSpec, 'at'>;
type PulseOptions = CommonOptions & Pick<PulseEffectSpec, 'at'>;
type TravelOptions = CommonOptions & Pick<TravelEffectSpec, 'from' | 'to'>;

function nonEmpty(value: string, label: string): string {
  const normalized = value.trim();
  if (!normalized) throw new Error(`${label} must not be empty`);
  return normalized;
}

function freezeAdvanced<T extends EffectBase['advanced']>(advanced: T): T {
  if (!advanced) return advanced;
  return Object.freeze({
    ...advanced,
    ...(advanced.palette ? { palette: Object.freeze([...advanced.palette]) } : {}),
  }) as T;
}

function common<T extends CommonOptions>(options: T): T {
  return {
    ...options,
    ...(options.key === undefined ? {} : { key: nonEmpty(options.key, 'effect key') }),
    ...(options.seedKey === undefined ? {} : { seedKey: nonEmpty(options.seedKey, 'seedKey') }),
    ...(options.advanced === undefined ? {} : { advanced: freezeAdvanced(options.advanced) }),
  };
}

export const fx = Object.freeze({
  anchor(name: string): NamedEffectAnchor {
    return Object.freeze({ kind: 'named', name: nonEmpty(name, 'anchor name') });
  },

  point(x: number, y: number): PointEffectAnchor {
    if (!Number.isFinite(x) || !Number.isFinite(y)) {
      throw new Error('effect point coordinates must be finite');
    }
    return Object.freeze({ kind: 'point', x, y });
  },

  burst(options: BurstOptions): BurstEffectSpec {
    return Object.freeze({ kind: 'burst', ...common(options) });
  },

  pulse(options: PulseOptions): PulseEffectSpec {
    return Object.freeze({ kind: 'pulse', ...common(options) });
  },

  travel(options: TravelOptions): TravelEffectSpec {
    return Object.freeze({ kind: 'travel', ...common(options) });
  },

  sequence(effects: readonly EffectSpec[], options: CommonOptions & { gapMs?: number } = {}): SequenceEffectSpec {
    const gapMs = options.gapMs ?? 0;
    if (!Number.isFinite(gapMs) || gapMs < 0) throw new Error('sequence gapMs must be a finite non-negative number');
    return Object.freeze({
      kind: 'sequence',
      ...common(options),
      effects: Object.freeze([...effects]),
      gapMs,
    });
  },

  parallel(effects: readonly EffectSpec[], options: CommonOptions = {}): ParallelEffectSpec {
    return Object.freeze({
      kind: 'parallel',
      ...common(options),
      effects: Object.freeze([...effects]),
    });
  },
});

export function defineEffectTheme(theme: EffectTheme): EffectTheme {
  const tones = theme.tones
    ? Object.fromEntries(Object.entries(theme.tones).map(([tone, colors]) => [
      tone,
      colors ? Object.freeze([...colors]) : colors,
    ]))
    : undefined;
  return Object.freeze(tones ? { tones: Object.freeze(tones) } : {});
}

export const DEFAULT_EFFECT_THEME: EffectTheme = Object.freeze({});

export function createEffectTransitionContext<S, MN extends string>(options: {
  before: S | null;
  after: S;
  move: ClientMove | null;
  version: number;
  snapshotEpoch: number;
}): EffectTransitionContext<S, MN> {
  const { before, after } = options;
  const common = {
    after,
    move: options.move as EffectTransitionContext<S, MN>['move'],
    version: options.version,
    snapshotEpoch: options.snapshotEpoch,
    changed<T>(select: (state: S) => T): boolean {
      if (before === null) return true;
      return !Object.is(select(before), select(after));
    },
  };
  return before === null
    ? Object.freeze({ ...common, kind: 'initial', before: null })
    : Object.freeze({ ...common, kind: 'transition', before });
}
