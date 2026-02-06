import { LitElement, html, css } from 'lit';
import { property, query } from 'lit/decorators.js';
import './boardgame-component-animator.js';
import '@polymer/paper-spinner/paper-spinner-lite.js';

/**
 * BoardgameRenderGame dynamically loads and manages game-specific renderers.
 * Handles animation coordination, state synchronization, and loading states.
 */
class BoardgameRenderGame extends LitElement {
  static override styles = css`
    #container {
      position: relative;
    }

    #loading[active] {
      visibility: visible;
      opacity: 1;
      transition: visibility var(--animation-length) step-start, opacity var(--animation-length, 0.25s) linear;
    }

    #loading {
      position: absolute;
      top: 0;
      left: 0;
      height: 100%;
      width: 100%;
      background-color: rgba(255, 255, 255, 0.7);
      z-index: 10;
      visibility: hidden;
      opacity: 0;
      transition: visibility var(--animation-length) step-end, opacity var(--animation-length, 0.25s) linear;
    }

    #loading > div {
      height: 100%;
      width: 100%;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
    }

    paper-spinner-lite {
      height: 100px;
      width: 100px;
      --paper-spinner-stroke-width: 10px;
    }
  `;

  @property({ type: Object })
  state: any = null;

  @property({ type: Object })
  chest: any = null;

  @property({ type: Boolean })
  active = false;

  @property({ type: String })
  diagram = '';

  @property({ type: String })
  gameName = '';

  @property({ type: Object, attribute: false })
  renderer: HTMLElement | null = null;

  @property({ type: Boolean })
  rendererLoaded = false;

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Number })
  currentPlayerIndex = 0;

  @property({ type: Boolean })
  socketActive = false;

  @property({ type: Number })
  defaultAnimationLength = 0;

  @property({ type: Object, attribute: false })
  private _activeAnimations: Map<HTMLElement, boolean> | null = null;

  @property({ type: Boolean, attribute: false })
  private _allAnimationsDoneFired = true;

  @query('#animator')
  private _animator?: any;

  @query('#container')
  private _container?: HTMLElement;

  private _boundComponentWillAnimate?: (e: Event) => void;
  private _boundComponentAnimationDone?: (e: Event) => void;

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);

    this._boundComponentWillAnimate = (e: Event) => this._componentWillAnimate(e as CustomEvent);
    this._boundComponentAnimationDone = (e: Event) => this._componentAnimationDone(e as CustomEvent);

    this.addEventListener('will-animate', this._boundComponentWillAnimate);
    this.addEventListener('animation-done', this._boundComponentAnimationDone);
    this._resetAnimating();
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    if (this._boundComponentWillAnimate) {
      this.removeEventListener('will-animate', this._boundComponentWillAnimate);
    }
    if (this._boundComponentAnimationDone) {
      this.removeEventListener('animation-done', this._boundComponentAnimationDone);
    }
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if (changedProperties.has('diagram')) {
      this._diagramChanged(this.diagram);
    }

    if (changedProperties.has('active')) {
      this._activeChanged(this.active);
    }

    if (changedProperties.has('viewingAsPlayer')) {
      this._viewingAsPlayerChanged(this.viewingAsPlayer);
    }

    if (changedProperties.has('currentPlayerIndex')) {
      this._currentPlayerIndexChanged(this.currentPlayerIndex);
    }

    if (changedProperties.has('chest')) {
      this._chestChanged(this.chest);
    }

    if (changedProperties.has('gameName')) {
      this._gameNameChanged(this.gameName);
    }

    if (changedProperties.has('defaultAnimationLength')) {
      this._defaultAnimationLengthChanged(this.defaultAnimationLength);
    }

    if (changedProperties.has('state')) {
      this._stateChanged(this.state, changedProperties.get('state') as any);
    }
  }

  private _diagramChanged(newValue: string) {
    if (!this.renderer) {
      return;
    }
    (this.renderer as any).diagram = newValue;
  }

  private _activeChanged(newValue: boolean) {
    if (!newValue) {
      // The game view has gone inactive
      // Clear out state now so by the time we switch back it will be null
      // and we minimize chance of trying to render state with the wrong renderer
      this.state = null;
      this.diagram = '';
      this.viewingAsPlayer = 0;
      this.currentPlayerIndex = 0;
      this._removeRenderer();
    } else {
      if (this.rendererLoaded) {
        this._instantiateRenderer();
      }
    }
  }

  private _ensureActiveAnimations() {
    if (this._activeAnimations) return;
    this._activeAnimations = new Map();
  }

  private _resetAnimating() {
    this._activeAnimations = null;
    this._ensureActiveAnimations();
    this._allAnimationsDoneFired = false;
  }

  private _componentWillAnimate(e: CustomEvent) {
    this._ensureActiveAnimations();
    this._activeAnimations!.set(e.detail.ele, true);
  }

  private _componentAnimationDone(e: CustomEvent) {
    // If we're already done, don't bother firing again
    this._ensureActiveAnimations();
    if (this._activeAnimations!.size === 0) return;
    this._activeAnimations!.delete(e.detail.ele);
    if (this._activeAnimations!.size === 0) {
      this._notifyAnimationsDone();
    }
  }

  private _nextStateIfNoAnimations() {
    if (this._activeAnimations && this._activeAnimations.size === 0) {
      this._notifyAnimationsDone();
    }
  }

  private _notifyAnimationsDone() {
    if (this._allAnimationsDoneFired) return;
    this._allAnimationsDoneFired = true;
    this.dispatchEvent(new CustomEvent('all-animations-done', { composed: true }));
  }

  private _defaultAnimationLengthChanged(newValue: number) {
    if (newValue === 0) {
      this.style.removeProperty('--animation-length');
      return;
    }
    this.style.setProperty('--animation-length', `${newValue / 1000}s`);
  }

  private _stateChanged(newState: any, oldState: any) {
    if (!this.renderer) return;
    const stateWasNull = ((this.renderer as any).state == null);
    if (newState && !stateWasNull) {
      this._resetAnimating();
      this._animator?.prepare();
    }

    // For Lit renderers, set property directly
    (this.renderer as any).state = newState;

    if (newState && !stateWasNull) {
      // Call animateFlip. When all of the things that will be animating have
      // started, check to see if no animations have been registered; if they
      // haven't, then we can advance to the next state immediately.
      this._animator?.animateFlip().then(() => this._nextStateIfNoAnimations());
      // TODO: technically it's possible that no animations fire, but
      // this._animator.animateFlip() returns immediately but schedules work in a
      // rAF callback. We used to check for this._activeAnimations.size == 0
      // and then bail, but that always triggered because animateFlip() returns
      // immediately.
    }
  }

  private _viewingAsPlayerChanged(newValue: number) {
    if (!this.renderer) return;
    (this.renderer as any).viewingAsPlayer = newValue;
  }

  private _currentPlayerIndexChanged(newValue: number) {
    if (!this.renderer) return;
    (this.renderer as any).currentPlayerIndex = newValue;
  }

  private _chestChanged(newValue: any) {
    if (!this.renderer) return;
    (this.renderer as any).chest = newValue;
  }

  private _gameNameChanged(newValue: string) {
    // If there was a state, it might be for a different game type which would
    // cause a render error
    this.state = null;
    this.rendererLoaded = false;
    this._removeRenderer();

    if (!newValue) return;

    import(`../../game-src/${newValue}/boardgame-render-game-${newValue}.js`)
      .then(() => this._instantiateRenderer(), null);
  }

  private _removeRenderer() {
    if (this.renderer && this._container) {
      this._container.removeChild(this.renderer);
    }
    this.renderer = null;
  }

  private _instantiateRenderer() {
    // The import loaded! Add it!
    this.rendererLoaded = true;

    const ele = document.createElement(`boardgame-render-game-${this.gameName}`) as any;

    ele.diagram = this.diagram;
    ele.state = this.state;
    ele.viewingAsPlayer = this.viewingAsPlayer;
    ele.currentPlayerIndex = this.currentPlayerIndex;
    ele.chest = this.chest;

    this.renderer = ele;

    if (this._container) {
      this._container.appendChild(ele);
    }

    // Only try to fire if there's a state. If it's the first time this
    // session we load the renderer, this will probably happen after the first
    // non-nil state is installed (it takes time to download the component), so
    // we'll need to ask for the next state. But if you load the same game type
    // again, the renderer will load immediately, most likely before the state
    // is installed. If we called this._notifyAnimationsDone() before there's a
    // state, it would be useless (and would prevent it from firing later).
    if (this.state) {
      // Sometimes the renderer is instantiated after the state is already
      // databound--which means that `all-animations-done` won't have fired.
      // _notifyAnimationsDone won't fire it again if it's already fired.
      window.requestAnimationFrame(() => this._notifyAnimationsDone());
    }
  }

  override render() {
    return html`
      <boardgame-component-animator
        id="animator"
        .ancestorOffsetParent="${this._container}">
      </boardgame-component-animator>

      <div ?hidden="${this.rendererLoaded}">
        <h2>Diagram of ${this.gameName}</h2>
        <pre>${this.diagram}</pre>
      </div>

      <div id="container">
        <!-- Dynamic renderer will be inserted here -->
      </div>

      <div id="loading" ?active="${!this.socketActive}">
        <div>
          <paper-spinner-lite ?active="${!this.socketActive}"></paper-spinner-lite>
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-render-game', BoardgameRenderGame);

export { BoardgameRenderGame };
