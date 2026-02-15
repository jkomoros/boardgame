import { LitElement } from 'lit';
import { property } from 'lit/decorators.js';

export class BoardgameAnimatableItem extends LitElement {
  @property({ type: Boolean })
  noAnimate = false;

  // Private properties - not decorated because they don't need reactivity
  private _expectedTransitionEnds = new Map<HTMLElement, Map<string, boolean>>();
  private _outstandingTransitonEnds = 0;
  private _boundTransitionEnded?: (e: Event) => void;

  override connectedCallback() {
    super.connectedCallback();
    this.resetAnimating();
    // CRITICAL: Event listeners must be added in connectedCallback to match Polymer timing
    // Must listen on both light DOM and shadow DOM for transitionend events
    this._boundTransitionEnded = (e: Event) => this._transitionEnded(e as TransitionEvent);
    this.addEventListener('transitionend', this._boundTransitionEnded);
    // Defensive check: shadowRoot may not exist yet
    if (this.shadowRoot) {
      this.shadowRoot.addEventListener('transitionend', this._boundTransitionEnded);
    }
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    if (this._boundTransitionEnded) {
      this.removeEventListener('transitionend', this._boundTransitionEnded);
      // Defensive check: shadowRoot may not exist
      if (this.shadowRoot) {
        this.shadowRoot.removeEventListener('transitionend', this._boundTransitionEnded);
      }
    }
  }

  // willNotAnimate says whether based on our current settings we expect this ele
  // and propName to fire a transitionend. Subclasses can override, but should
  // call this. Default answer is false, but this will return true if noAnimate
  // is true.
  willNotAnimate(ele: HTMLElement, propName: string): boolean {
    if (this.noAnimate) {
      return true;
    }
    return false;
  }

  // resetAnimating should be called when we expect animating count to be zero
  resetAnimating() {
    // if (this._animatingCount != 0) console.warn(this, this._animatingCount, "Was not zero when expected");
    this._expectedTransitionEnds = new Map();
    this._outstandingTransitonEnds = 0;
  }

  // beforeOrphaned is called when we know we're about to be orphaned (for
  // example if we're an animating component that will be removed when done
  // animating). it's our last chance to fire 'animation-done' if we were going
  // to fire that.
  beforeOrphaned() {
    if (!this._expectedTransitionEnds) return;
    if (!this._expectedTransitionEnds.size) return;
    this._notifyAnimationDone();
  }

  // _expectTransitionEnd is called whenever we have just changed a property
  // that will later fire a transitionend, with the specific ele (this,
  // #inner, #outer), and propertyName we expect to fire. We only care about
  // transform and opacity changes; ignore everything else. We also will
  // ignore things that this.willNotAnimate() tell us won't animate.
  protected _expectTransitionEnd(ele: HTMLElement, propName: string) {
    if (!this._expectedTransitionEnds) {
      // This happens the first time state is installed. No biggie, just skip
      // it.
      return;
    }

    if (propName !== 'transform' && propName !== 'opacity') return;
    if (this.willNotAnimate(ele, propName)) {
      // Sometimes we will have already told us to expect one, but later we
      // realize that we actually won't. This can happen for example the first
      // time a non-spacer card is set to a spacer--we update the inner
      // transform, then set spacer, then later update again. In those cases,
      // we should forget the one we previously told ourselves to expect.
      this._removeExpectedTransition(ele, propName);
      return;
    }

    let expectedPropsMap = this._expectedTransitionEnds.get(ele);
    if (!expectedPropsMap) {
      expectedPropsMap = new Map();
      this._expectedTransitionEnds.set(ele, expectedPropsMap);
    }

    // Already set!
    if (expectedPropsMap.get(propName)) return;

    expectedPropsMap.set(propName, true);
    this._outstandingTransitonEnds++;

    if (this._outstandingTransitonEnds === 1) {
      // This was the first one, fire a will-animate.
      this.dispatchEvent(new CustomEvent('will-animate', { bubbles: true, composed: true, detail: { ele: this } }));
    }
  }

  // removes the ele and propName from the map, and returns whether it was in there.
  private _removeExpectedTransition(ele: HTMLElement, propName: string): boolean {
    if (!this._expectedTransitionEnds) return false;
    const expectedPropsMap = this._expectedTransitionEnds.get(ele);
    if (!expectedPropsMap) return false;
    if (!expectedPropsMap.get(propName)) return false;
    expectedPropsMap.delete(propName);
    this._outstandingTransitonEnds--;
    if (this._outstandingTransitonEnds < 0) {
      console.warn('Got to less than 0 transition ends somehow');
      this._outstandingTransitonEnds = 0;
    }
    if (expectedPropsMap.size === 0) {
      this._expectedTransitionEnds.delete(ele);
    }
    return true;
  }

  private _notifyAnimationDone() {
    this.dispatchEvent(new CustomEvent('animation-done', { bubbles: true, composed: true, detail: { ele: this } }));
  }

  // _transitionEnded is the handler for transitionend. It will fire for _any_
  // transition that ended on ourselves or our shadow root. We only care about
  // transform and opacity changes; ignore everything else, because we'll
  // heard about every property that changes, including box-shadow and others
  // that are non-semantic.
  private _transitionEnded(e: TransitionEvent) {
    if (e.propertyName !== 'transform' && e.propertyName !== 'opacity') return;

    // CRITICAL: Use composedPath() instead of deprecated e.path
    // Defensive check: composedPath() may not be supported in older browsers
    if (typeof e.composedPath !== 'function') {
      console.warn('composedPath not supported');
      return;
    }

    const path = e.composedPath();
    if (!path || path.length < 1) return;

    const target = path[0];
    if (!(target instanceof HTMLElement)) return;
    const ele = target;

    const changeMade = this._removeExpectedTransition(ele, e.propertyName);

    if (changeMade && this._outstandingTransitonEnds === 0) {
      // all of the animations we were expecting to finish are finished.
      this._notifyAnimationDone();
    }
  }
}

customElements.define('boardgame-animatable-item', BoardgameAnimatableItem);
