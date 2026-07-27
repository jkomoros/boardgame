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
  componentMotionTrackEasing,
} from '../motion/component-track.js';
import type {
  ComponentMotionChannel,
  ComponentMotionTarget,
  ComponentMotionTrack,
} from '../motion/component-track.js';
import type { AnimatableRegistry } from '../motion/animatable-registry.js';

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

  // Animation -> whether it holds the completion gate. The gated flag is
  // what lets finishGatedAnimations() force-settle a stale cycle's
  // participants (the registry sweep's job) without touching UNGATED ambient
  // loops (an infinite highlight throb), which were never cycle participants.
  private _liveAnimations = new Map<Animation, boolean>();
  private _activeHookFrames = new Map<Animation, number>();
  private _liveGatedCount = 0;
  private _settledResolvers: Array<() => void> = [];

  // The ambient AnimatableRegistry discovered by the connect-time walk (see
  // _ambientLookup / connectedCallback), cached so disconnectedCallback can
  // symmetrically unregister from the SAME registry even after this
  // element has been detached (by then the walk itself can no longer reach
  // an ancestor -- parentNode is already null).
  private _ambientRegistry: AnimatableRegistry | null = null;

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

  // Discovers the ambient AnimatableRegistry (#714's non-component
  // discovery gap, Task 9) and registers with it so an item outside the
  // component animator's own stack bookkeeping (a standalone die,
  // status-text, fading-text, a game-authored token, ...) is still reached
  // by render-game's cycle-start reset. Lit's connectedCallback runs before
  // this element's own first render, but the discovery walk only climbs
  // ANCESTORS -- by the time connectedCallback fires this element is
  // already attached to a connected tree, so the full ancestor chain
  // (crossing shadow roots and slots) is already in place regardless of
  // this element's own render state. Always call super first.
  override connectedCallback(): void {
    super.connectedCallback();
    this._ambientRegistry = this._ambientLookup<AnimatableRegistry>('animatableRegistry', () => true);
    this._ambientRegistry?.register(this);
  }

  // Unregisters from the SAME registry found at connect time (never
  // re-walks here: at disconnect this element may already be unparented,
  // so the walk could no longer reach the ancestor that registered it).
  // Always call super, and always last, mirroring other overrides in this
  // codebase that release resources before yielding to the base class.
  //
  // Orphan-settle safety net (#714 Phase 2 gate finding): a BOARD/stack
  // component gets beforeOrphaned() (force-settle) from the animator before
  // removal, but that mechanism is stack-specific -- a roster-hosted (or any
  // other ambiently-discovered) animatable removed from the DOM mid-animation
  // has no equivalent caller. Without a settle, its WAAPI animation keeps
  // running against the document timeline after detach, AND once it does
  // finish, its `animation-done` CustomEvent (bubbles + composed) dispatches
  // from a node with no parent -- nothing to bubble to -- so the gate never
  // hears it and is stuck open until the watchdog force-closes it. Deferred
  // to a microtask (not checked synchronously here) because a synchronous
  // reparent -- Lit moving this element to a new parent within the same
  // synchronous span -- also fires disconnectedCallback; checking
  // isConnected only after the microtask queue drains distinguishes a
  // genuine removal from a same-tick reparent, so a reparented element's
  // in-flight animation is never needlessly snapped.
  override disconnectedCallback(): void {
    this._ambientRegistry?.unregister(this);
    this._ambientRegistry = null;
    if (this._liveGatedCount > 0) {
      queueMicrotask(() => {
        if (!this.isConnected) this.finishAllAnimations();
      });
    }
    super.disconnectedCallback();
  }

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
    // Timing is derived PER TRACK, not shared across the batch. A sampled
    // track already encodes its own timing in its samples, so the kernel's
    // default effect-level 'ease-in-out' would time-warp it; an endpoint track
    // in the same batch still wants that default. One shared OptionalEffectTiming
    // cannot express both.
    //
    // Effect level, not per-keyframe, deliberately: WAAPI keyframe easing
    // already defaults to linear (a per-keyframe write would be a no-op), and
    // boardgame-component-animator publishes effect.getTiming().easing into
    // StructuralExecutedTiming, which would report 'linear' for every track --
    // and so stop meaning anything -- if character lived in the keyframes.
    //
    // BINDING HAPPENS INSIDE THE TRY, and that placement is load-bearing.
    // boardgame-component-animator plays a whole cycle's components in a bare
    // `for` loop, and a component's playAnimation() writes its FINAL transform
    // only after this returns -- so an exception escaping here does not fail one
    // component, it unwinds the loop and leaves every component after it in the
    // cycle frozen at its INVERTED FLIP transform, off in the position it was
    // animating out of. A producer error on one track has to stay one track's
    // problem: the catch below reports it loudly (console.error, and a
    // 'playback-error' result) and returns, so the caller still lands on its own
    // resting style and its siblings still play.
    const playbacks: MotionTrackPlayback[] = [];
    try {
      const bindings = tracks.map(track => {
        const channel = componentMotionChannel(track);
        const easing = componentMotionTrackEasing(track);
        if (easing !== undefined && timing?.easing !== undefined) {
          // Two time warps on one channel is the same class of producer error as
          // two owners on one channel: the sampled trajectory and the caller's
          // easing curve both claim the channel's timeline.
          throw new Error(
            `component motion channel ${channel} carries its own sampled timeline; `
            + `an explicit timing.easing (${String(timing.easing)}) would warp it`,
          );
        }
        return Object.freeze({
          track,
          channel,
          target: this.motionTrackTarget(track.target),
          timing: easing === undefined ? timing : { ...timing, easing },
        });
      });
      const missing = bindings.find(binding => !binding.target);
      if (missing) {
        return Object.freeze({
          status: 'skipped',
          reason: 'missing-target',
          channel: missing.channel,
        });
      }

      for (const binding of bindings) {
        const animation = this.play(
          binding.target!,
          [...componentMotionKeyframes(binding.track)],
          binding.timing,
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
        // Animations run with fill:'none', so the instant one settles -- by
        // natural completion, by finish() from the cycle sweep, or by the
        // duration-0 effect reduced motion resolves to -- the element renders
        // its RESTING style, not the last keyframe. playAnimation writes a
        // resting style for the host channel only; a component-owned visual
        // channel has no framework write, which is why the card maintains its
        // inner transform by hand. Writing the track's declared resting value
        // here makes that structural rather than per-producer discipline. Safe
        // to write while the animation is live: an active effect overrides the
        // inline style for the length of its active interval.
        if (binding.track.resting !== undefined) {
          if (binding.track.property === 'transform') {
            binding.target!.style.transform = binding.track.resting;
          } else {
            binding.target!.style.opacity = binding.track.resting;
          }
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

  // _ambientLookup climbs from this element's parent -- crossing shadow
  // roots and slots -- looking for the nearest ancestor exposing `propName`,
  // calling `isProvider` on each candidate value found to decide whether
  // that ancestor is a genuine provider or must be climbed past. Shared by
  // the two ambient discovery walks below: they cross the exact same DOM
  // shape (commit 7172dd24 made this climb past NULL contexts so nested
  // animatable wrappers like status-text don't sever deeper items from the
  // real provider above them) but disagree on what "provider" means for
  // their respective property, hence the caller-supplied predicate.
  //
  // `protected` because it is the framework's one ambient-discovery walk and a
  // subclass may need a third: <boardgame-die> seeds its physics roll from the
  // ambient `gameVersion` the same way, so that a game gets reproducible rolls
  // without wiring a property through every renderer.
  protected _ambientLookup<T>(propName: string, isProvider: (value: T) => boolean): T | null {
    let node: Node | null = this.assignedSlot ?? this.parentNode;
    while (node) {
      if (propName in node) {
        const value = (node as unknown as Record<string, T>)[propName];
        if (isProvider(value)) return value;
      }
      if (node instanceof ShadowRoot) {
        node = node.host;
      } else {
        node = (node as Element).assignedSlot ?? node.parentNode;
      }
    }
    return null;
  }

  private _ambientAnimationContext(): VersionAnimationContext | null {
    // Prefer the render-game provider over a component's cached value. This
    // crosses shadow roots and slots, so standalone dice and game-authored
    // animatable items get the same context as stack-managed components.
    //
    // A POPULATED context ends the walk; a null one does not. Every
    // BoardgameAnimatableItem inherits an `animationContext` property
    // defaulting to null, so once wrapper elements (status-text and
    // friends, #714) join the class hierarchy, a presence check would stop
    // the walk at the nearest animatable ancestor and silently sever
    // nested items (status-text's own fading-text) from the render-game
    // provider above it. Climbing past nulls preserves the legit
    // "provider currently between cycles" case too: with no populated
    // context anywhere, the result is null either way.
    const found = this._ambientLookup<VersionAnimationContext | null>(
      'animationContext',
      (ctx) => ctx != null,
    );
    return found ?? this.animationContext;
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
    // composite 'replace' is pinned EXPLICITLY because it is load-bearing
    // (Phase 3 gate regression critic): when two transform animations run
    // on one host (a layoutTransform self-play plus the same cycle's FLIP
    // host track), replace semantics mean the higher animation wins
    // outright each frame — and since both encode the same net geometry,
    // the winner renders one correct motion. Under 'add' the identical
    // setup would visibly double the motion, and no parity golden can
    // catch that (curves are displacement-normalized, so a uniform 2x
    // divides out). Do not remove or parameterize this without a test
    // that pins the same-host composite case.
    const anim = element.animate(keyframes, { ...resolution.timing, composite: 'replace' });
    this._liveAnimations.set(anim, gated);
    if (gated) {
      this._liveGatedCount++;
      // Tell the render-game watchdog how long this gated play is declared to
      // run so it can extend its deadline past a legitimately long cycle
      // (stagger delay + duration + post-animation-delay) instead of
      // force-closing mid-animation.
      //
      // EVERY gated play declares, not just the 0->1 transition. Declaring
      // once was only ever harmless because every track in a batch shared one
      // timing, so the first play's expectedSettleMs spoke for the whole
      // element. It no longer does: a sampled track carries its own (longer)
      // duration, and compileComponentMotionTracks orders visual tracks LAST,
      // so a first-play-only declaration would arm the watchdog from the short
      // host FLIP and force-close mid-tumble. AnimationGate.willAnimate is
      // idempotent (a Map set keyed by element) and monotone (it only ever
      // extends the deadline, never shrinks it), so re-declaring is safe.
      this.dispatchEvent(new CustomEvent('will-animate',
        {
          bubbles: true,
          composed: true,
          detail: { ele: this, expectedSettleMs: resolution.expectedSettleMs },
        }));
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

  // Jump one live animation to its end state (or cancel it if it cannot
  // finish -- an infinite animation throws InvalidStateError from finish()).
  // Either path drives settlement through the finished-promise .finally().
  private _forceSettle(anim: Animation): void {
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

  // finishAllAnimations jumps EVERY live animation to its end state --
  // including ungated ambient loops (an infinite highlight throb). Reserved
  // for the paths where this element is leaving the tree for good
  // (beforeOrphaned / disconnectedCallback): once detached, an ambient loop
  // running against the document timeline is pure waste, so kill it too.
  // Do NOT use this for the cycle-interruption sweep -- see
  // finishGatedAnimations.
  finishAllAnimations(): void {
    for (const anim of [...this._liveAnimations.keys()]) this._forceSettle(anim);
  }

  // finishGatedAnimations force-settles only the GATED animations -- the
  // completion-cycle participants. It backs render-game's cycle-start
  // registry sweep, whose purpose is to end a stale cycle's still-running
  // animations before the next cycle's play() overlaps them. UNGATED ambient
  // decoration (an infinite throb started with { gated: false }) was never a
  // cycle participant and must survive the sweep -- cancelling it here is the
  // ambient-animation-sweep regression (evidence pack
  // 2026-07-26-ambient-animation-sweep.md).
  finishGatedAnimations(): void {
    for (const [anim, gated] of [...this._liveAnimations]) {
      if (gated) this._forceSettle(anim);
    }
  }
}

customElements.define('boardgame-animatable-item', BoardgameAnimatableItem);
