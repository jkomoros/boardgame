import { LitElement } from 'lit';
import { property } from 'lit/decorators.js';
import { animHooks } from '../utils/anim-test-hooks.js';
import {
  finiteTimingMs,
  resolveMotionTiming,
} from '../motion/timing.js';
import type {
  AnimationTimingPolicy,
  VersionAnimationContext,
} from '../motion/timing.js';
import {
  componentMotionChannel,
  componentMotionKeyframes,
} from '../motion/component-track.js';
import type {
  ComponentMotionChannel,
  ComponentMotionTarget,
  ComponentMotionTrack,
} from '../motion/component-track.js';

export type { AnimationTimingPolicy } from '../motion/timing.js';

export interface PlayOptions {
  gated?: boolean;
  /** Defaults to the installed version slot; use immediate for a local effect. */
  timing?: AnimationTimingPolicy;
}

export interface MotionTrackPlayback {
  readonly channel: ComponentMotionChannel;
  readonly track: ComponentMotionTrack;
  readonly animation: Animation;
}

export type MotionTrackPlayResult =
  | Readonly<{
    status: 'started';
    playbacks: readonly MotionTrackPlayback[];
  }>
  | Readonly<{
    status: 'skipped';
    reason: 'missing-target' | 'not-started' | 'playback-error';
    channel?: ComponentMotionChannel;
  }>;

interface PlayInstrumentation {
  recordActive?: boolean;
}

export class BoardgameAnimatableItem extends LitElement {
  @property({ type: Boolean })
  noAnimate = false;

  // Installed by the shared animator for the state version currently being
  // rendered. Keeping this on the common play primitive makes ordinary FLIP
  // and property animations participate without game-renderer plumbing.
  animationContext: VersionAnimationContext | null = null;

  private _liveAnimations = new Set<Animation>();
  private _activeHookFrames = new Map<Animation, number>();
  private _liveGatedCount = 0;
  private _settledResolvers: Array<() => void> = [];

  @property({ type: Number, attribute: 'post-animation-delay' })
  postAnimationDelay = 0;

  // waitForAnimation controls whether this item's animations hold the
  // completion gate (#716). Boolean-ish attribute with one twist: since
  // the property DEFAULTS to true, the only useful thing an attribute can
  // express is "false" — so unlike a stock Lit Boolean, the literal
  // string "false" parses as false. Absent attribute → default (true).
  @property({
    attribute: 'wait-for-animation',
    converter: {
      fromAttribute: (value: string | null) => value !== 'false',
      toAttribute: (value: boolean) => (value ? '' : 'false'),
    },
  })
  waitForAnimation = true;

  // beforeOrphaned is called when we know we're about to be orphaned (for
  // example if we're an animating component that will be removed when done
  // animating). it's our last chance to settle so the gate never waits on a
  // detached element.
  beforeOrphaned() {
    // Last chance before removal: settle everything so the gate never
    // waits on a detached element.
    this.finishAllAnimations();
  }

  // animationLengthMs reads the effective --animation-length CSS variable
  // (games set it; render-game sets it from the renderer's animationLength()
  // hook). Accepts '0.25s' or '250ms'. Returns milliseconds.
  animationLengthMs(): number {
    const raw = getComputedStyle(this).getPropertyValue('--animation-length').trim();
    if (!raw) return 250;
    if (raw.endsWith('ms')) return parseFloat(raw) || 250;
    if (raw.endsWith('s')) return (parseFloat(raw) || 0.25) * 1000;
    const n = parseFloat(raw);
    return isNaN(n) ? 250 : n;
  }

  get isAnimating(): boolean {
    return this._liveGatedCount > 0;
  }

  /** Resolve the two deliberately finite DOM ownership surfaces. */
  protected motionTrackTarget(target: ComponentMotionTarget): HTMLElement | null {
    return target === 'host' ? this : null;
  }

  /** Execute immutable owned tracks through the shared timing/gating kernel. */
  protected playMotionTracks(
    tracks: readonly ComponentMotionTrack[],
    timing?: OptionalEffectTiming,
    opts?: PlayOptions,
  ): MotionTrackPlayResult {
    if (tracks.length === 0) {
      return Object.freeze({ status: 'skipped', reason: 'not-started' });
    }
    const bindings = tracks.map(track => Object.freeze({
      track,
      channel: componentMotionChannel(track),
      target: this.motionTrackTarget(track.target),
    }));
    const missing = bindings.find(binding => !binding.target);
    if (missing) {
      return Object.freeze({
        status: 'skipped',
        reason: 'missing-target',
        channel: missing.channel,
      });
    }

    const playbacks: MotionTrackPlayback[] = [];
    try {
      for (const binding of bindings) {
        const animation = this.play(
          binding.target!,
          [...componentMotionKeyframes(binding.track)],
          timing,
          opts,
        );
        if (!animation) {
          for (const playback of playbacks) playback.animation.cancel();
          return Object.freeze({
            status: 'skipped',
            reason: 'not-started',
            channel: binding.channel,
          });
        }
        playbacks.push(Object.freeze({
          channel: binding.channel,
          track: binding.track,
          animation,
        }));
      }
    } catch (error) {
      for (const playback of playbacks) playback.animation.cancel();
      console.error('[animation] motion track playback failed:', error);
      return Object.freeze({ status: 'skipped', reason: 'playback-error' });
    }
    return Object.freeze({
      status: 'started',
      playbacks: Object.freeze(playbacks),
    });
  }

  private _ambientAnimationContext(): VersionAnimationContext | null {
    // Prefer the render-game provider over a component's cached value. This
    // crosses shadow roots and slots, so standalone dice and game-authored
    // animatable items get the same context as stack-managed components.
    let node: Node | null = this.assignedSlot ?? this.parentNode;
    while (node) {
      if ('animationContext' in node) {
        return (node as Node & { animationContext: VersionAnimationContext | null }).animationContext;
      }
      if (node instanceof ShadowRoot) {
        node = node.host;
      } else {
        node = (node as Element).assignedSlot ?? node.parentNode;
      }
    }
    return this.animationContext;
  }

  // play is the single entry point for starting an animation on this item
  // (host element, #inner, or any shadow child). Ground truth for
  // completion is the returned Animation's settlement — there is nothing
  // to guess (spec: WAAPI rewrite).
  play(element: HTMLElement, keyframes: Keyframe[], timing?: OptionalEffectTiming,
       opts?: PlayOptions, instrumentation?: PlayInstrumentation): Animation | null {
    if (this.noAnimate) return null;
    const gated = (opts?.gated ?? true) && this.waitForAnimation;
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    const timingPolicy = opts?.timing ?? 'version';
    const resolution = resolveMotionTiming(timing ?? {}, {
      policy: timingPolicy,
      context: timingPolicy === 'version' ? this._ambientAnimationContext() : null,
      defaults: {
        duration: this.animationLengthMs(),
        easing: 'ease-in-out',
        fill: 'none',
      },
      reducedMotion: reduced,
      // End delay is part of synchronized occupancy, not an after-the-fact
      // addition that may push the queue into a later version slot.
      postAnimationDelayMs: this.postAnimationDelay,
    });
    if (resolution.kind === 'skip') return null;
    const anim = element.animate(keyframes, resolution.timing);
    this._liveAnimations.add(anim);
    if (gated) {
      this._liveGatedCount++;
      if (this._liveGatedCount === 1) {
        // Tell the render-game watchdog how long this gated play is
        // declared to run so it can extend its deadline past a legitimately
        // long cycle (stagger delay + duration + post-animation-delay)
        // instead of force-closing mid-animation. Numbers only — coerce the
        this.dispatchEvent(new CustomEvent('will-animate',
          {
            bubbles: true,
            composed: true,
            detail: { ele: this, expectedSettleMs: resolution.expectedSettleMs },
          }));
      }
    }
    animHooks.record('play', this.tagName.toLowerCase() + (this.id ? `#${this.id}` : ''));
    if (instrumentation?.recordActive !== false) {
      const detail = this.tagName.toLowerCase() + (this.id ? `#${this.id}` : '');
      const delay = finiteTimingMs(resolution.timing.delay);
      const observeActive = () => {
        if (!this._liveAnimations.has(anim)) return;
        const currentTime = anim.currentTime;
        if (typeof currentTime === 'number' && currentTime + 0.5 >= delay) {
          this._activeHookFrames.delete(anim);
          animHooks.record('active', detail, resolution.activeContext ? {
            version: resolution.activeContext.version,
            targetAtMs: resolution.activeContext.startAtMs,
          } : undefined);
          return;
        }
        this._activeHookFrames.set(anim, requestAnimationFrame(observeActive));
      };
      this._activeHookFrames.set(anim, requestAnimationFrame(observeActive));
    }
    // finished rejects on cancel(); both paths are settlement for us.
    anim.finished.catch(() => {}).finally(() => this._animationSettled(anim, gated));
    return anim;
  }

  private _animationSettled(anim: Animation, gated: boolean) {
    if (!this._liveAnimations.delete(anim)) return; // already accounted
    const activeFrame = this._activeHookFrames.get(anim);
    if (activeFrame !== undefined) {
      cancelAnimationFrame(activeFrame);
      this._activeHookFrames.delete(anim);
    }
    animHooks.record('settle', this.tagName.toLowerCase() + (this.id ? `#${this.id}` : ''));
    if (!gated) return;
    this._liveGatedCount--;
    if (this._liveGatedCount <= 0) {
      this._liveGatedCount = 0;
      const resolvers = this._settledResolvers;
      this._settledResolvers = [];
      for (const r of resolvers) r();
      this.dispatchEvent(new CustomEvent('animation-done',
        { bubbles: true, composed: true, detail: { ele: this } }));
    }
  }

  settled(): Promise<void> {
    if (this._liveGatedCount === 0) return Promise.resolve();
    return new Promise((resolve) => this._settledResolvers.push(resolve));
  }

  // finishAllAnimations jumps every live animation to its end state and
  // resolves settlement. Called when a new animation cycle must start
  // while a previous one is in flight (spec: Interruption semantics).
  finishAllAnimations(): void {
    for (const anim of [...this._liveAnimations]) {
      try {
        if (anim.playState === 'running' || anim.playState === 'finished') {
          anim.finish();
        } else {
          anim.cancel();
        }
      } catch {
        // finish() throws InvalidStateError for infinite animations; cancel instead.
        try { anim.cancel(); } catch { /* already dead */ }
      }
    }
  }
}

customElements.define('boardgame-animatable-item', BoardgameAnimatableItem);
