import { LitElement } from 'lit';
import { property } from 'lit/decorators.js';
import { animHooks } from '../utils/anim-test-hooks.js';

export class BoardgameAnimatableItem extends LitElement {
  @property({ type: Boolean })
  noAnimate = false;

  private _liveAnimations = new Set<Animation>();
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

  // play is the single entry point for starting an animation on this item
  // (host element, #inner, or any shadow child). Ground truth for
  // completion is the returned Animation's settlement — there is nothing
  // to guess (spec: WAAPI rewrite).
  play(element: HTMLElement, keyframes: Keyframe[], timing?: OptionalEffectTiming,
       opts?: { gated?: boolean }): Animation | null {
    if (this.noAnimate) return null;
    const gated = (opts?.gated ?? true) && this.waitForAnimation;
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    const resolvedTiming: OptionalEffectTiming = {
      duration: reduced ? 0 : this.animationLengthMs(),
      easing: 'ease-in-out',
      fill: 'none',
      ...timing,
    };
    if (this.postAnimationDelay > 0 && resolvedTiming.endDelay === undefined) {
      resolvedTiming.endDelay = this.postAnimationDelay;
    }
    const anim = element.animate(keyframes, resolvedTiming);
    this._liveAnimations.add(anim);
    if (gated) {
      this._liveGatedCount++;
      if (this._liveGatedCount === 1) {
        // Tell the render-game watchdog how long this gated play is
        // declared to run so it can extend its deadline past a legitimately
        // long cycle (stagger delay + duration + post-animation-delay)
        // instead of force-closing mid-animation. Numbers only — coerce the
        // resolved timing fields (they may be CSSNumericValue-ish or absent).
        const num = (v: unknown): number =>
          typeof v === 'number' && isFinite(v) ? v : 0;
        const expectedSettleMs = num(resolvedTiming.delay)
          + num(resolvedTiming.duration)
          + num(resolvedTiming.endDelay);
        this.dispatchEvent(new CustomEvent('will-animate',
          { bubbles: true, composed: true, detail: { ele: this, expectedSettleMs } }));
      }
    }
    animHooks.record('play', this.tagName.toLowerCase() + (this.id ? `#${this.id}` : ''));
    // finished rejects on cancel(); both paths are settlement for us.
    anim.finished.catch(() => {}).finally(() => this._animationSettled(anim, gated));
    return anim;
  }

  private _animationSettled(anim: Animation, gated: boolean) {
    if (!this._liveAnimations.delete(anim)) return; // already accounted
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
