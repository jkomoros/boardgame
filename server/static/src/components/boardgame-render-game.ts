import { LitElement, html, css } from 'lit';
import { property, query } from 'lit/decorators.js';
import './boardgame-component-animator.js';
import type { BoardgameComponentStack } from './boardgame-component-stack.js';
import type { MoveForm } from '../types/api.js';
import type { MoveLegalityInfo } from '../selectors.js';
import { surfaceForGame } from '../utils/companion-surface.js';
import { animHooks } from '../utils/anim-test-hooks.js';

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
      background-color: rgba(250, 246, 240, 0.7);
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

    .spinner {
      width: 100px;
      height: 100px;
      border: 10px solid var(--md-sys-color-outline-variant, #CCC4B8);
      border-top: 10px solid var(--md-sys-color-primary, #2E6B4F);
      border-radius: 50%;
      animation: spin 1s linear infinite;
    }

    @keyframes spin {
      0% { transform: rotate(0deg); }
      100% { transform: rotate(360deg); }
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

  // gameId is needed to read the per-game surface cookie (surface_<gameId>)
  // for companion-mode routing. Empty string means we haven't been told the
  // gameID yet (the loader will operate as solo until it's set).
  @property({ type: String })
  gameId = '';

  @property({ type: Object })
  companionInfo: import('../types/store').CompanionInfo | null = null;

  // isOwner is the doGameInfo IsOwner bool — true if the authenticated
  // user is the game's Owner. Pass-through; the surface renderer combines
  // this with its own surface-cookie check to compute isHost.
  @property({ type: Boolean })
  isOwner = false;

  // gameFinished/gameWinners mirror the game record's Finished/Winners so
  // renderers can show an ending (winner banner, you-won/lost) without
  // bespoke plumbing. Winners are player indexes.
  @property({ type: Boolean })
  gameFinished = false;

  @property({ type: Array })
  gameWinners: number[] = [];


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

  @property({ type: Array, attribute: false })
  moveForms: MoveForm[] | null = null;

  @property({ type: Number })
  defaultAnimationLength = 0;

  @property({ type: Object, attribute: false })
  private _activeAnimations: Map<HTMLElement, boolean> | null = null;

  @property({ type: Boolean, attribute: false })
  private _allAnimationsDoneFired = true;

  // isAnimating reflects whether the animation gate is currently open (an
  // animation cycle is in flight). Reflected to the `is-animating` attribute
  // so tests and ancestor CSS can observe it without reaching into internals,
  // and mirrored via `animating-changed` so ancestors (boardgame-game-view)
  // can wire it into move-disabling UI without polling. See #721.
  @property({ type: Boolean, reflect: true, attribute: 'is-animating' })
  isAnimating = false;

  @query('#animator')
  private _animator?: any;

  @query('#container')
  private _container?: HTMLElement;

  private _boundComponentWillAnimate?: (e: Event) => void;
  private _boundComponentAnimationDone?: (e: Event) => void;
  private _animationWatchdogTimer: ReturnType<typeof setTimeout> | null = null;

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);

    this._boundComponentWillAnimate = (e: Event) => this._componentWillAnimate(e as CustomEvent);
    this._boundComponentAnimationDone = (e: Event) => this._componentAnimationDone(e as CustomEvent);

    this.addEventListener('will-animate', this._boundComponentWillAnimate);
    this.addEventListener('animation-done', this._boundComponentAnimationDone);
    this._activeAnimations = null;
    this._ensureActiveAnimations();
    this._allAnimationsDoneFired = false;
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    if (this._boundComponentWillAnimate) {
      this.removeEventListener('will-animate', this._boundComponentWillAnimate);
    }
    if (this._boundComponentAnimationDone) {
      this.removeEventListener('animation-done', this._boundComponentAnimationDone);
    }
    // Clean up watchdog timer to prevent firing after element is removed.
    if (this._animationWatchdogTimer !== null) {
      clearTimeout(this._animationWatchdogTimer);
      this._animationWatchdogTimer = null;
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

    if (changedProperties.has('moveForms')) {
      this._moveFormsChanged(this.moveForms);
    }

    if (changedProperties.has('state')) {
      this._stateChanged(this.state, changedProperties.get('state') as any);
    }

    if (changedProperties.has('companionInfo')) {
      this._companionInfoChanged(this.companionInfo);
    }

    if (changedProperties.has('isOwner')) {
      this._isOwnerChanged(this.isOwner);
    }

    if (changedProperties.has('gameFinished') || changedProperties.has('gameWinners')) {
      this._applyGameOutcomeToRenderer();
    }

    if (changedProperties.has('renderer')) {
      this.dispatchEvent(new CustomEvent('renderer-changed', {
        composed: true, bubbles: true, detail: { value: this.renderer }
      }));
    }
  }

  private _diagramChanged(newValue: string) {
    if (!this.renderer) {
      return;
    }
    (this.renderer as any).diagram = newValue;
  }

  // _companionInfoChanged propagates the companionInfo bundle to the
  // inner surface renderer whenever the gameInfo response refreshes. The
  // base view classes (BoardgameTableViewBase / BoardgameHandViewBase)
  // expose typed properties for the unpacked fields; we set them all so
  // the renderer can react to e.g. an absent player coming back without
  // a full re-mount.
  private _companionInfoChanged(_newValue: import('../types/store').CompanionInfo | null) {
    if (!this.renderer) return;
    this._applyCompanionPropsToRenderer(this.renderer);
  }

  private _isOwnerChanged(newValue: boolean) {
    if (!this.renderer) return;
    (this.renderer as any).isOwner = newValue;
    this._recomputeIsHost();
  }

  private _applyCompanionPropsToRenderer(ele: HTMLElement) {
    const r = ele as any;
    const info = this.companionInfo;
    r.seatPresentations = info?.SeatPresentations || [];
    r.absentPlayers = info?.Absent || [];
    r.roomCode = info?.RoomCode || '';
    r.roomLocked = info?.RoomLocked || false;
    r.companionMode = info?.CompanionMode || false;
    this._recomputeIsHost();
    this._applyGameOutcomeToRenderer();
  }

  private _applyGameOutcomeToRenderer() {
    if (!this.renderer) return;
    const r = this.renderer as any;
    r.gameFinished = this.gameFinished;
    r.gameWinners = this.gameWinners;
  }

  // _applyAnimatingToRenderer mirrors isAnimating onto the renderer so the
  // Table/Hand view bases can gate outcome/verdict rendering on it (#798
  // final piece): the outcome must never appear while the last animation
  // cycle (e.g. the winning card landing) is still in flight. Called at
  // both gate flips (_resetAnimating / _notifyAnimationsDone) and at
  // renderer instantiation so a renderer created mid-cycle (or finished
  // and re-instantiated on a surface switch) starts with the correct value
  // rather than defaulting to false.
  private _applyAnimatingToRenderer() {
    if (!this.renderer) return;
    (this.renderer as any).animating = this.isAnimating;
  }

  private _recomputeIsHost() {
    if (!this.renderer) return;
    // Prefer the server's own verdict (CompanionInfo.IsHost, computed with
    // the same Owner-or-override + surface-cookie rule the host-action
    // endpoints enforce) so a host promoted via /claimHost sees controls
    // even though they aren't the Owner. Fall back to the local derivation
    // for older payloads that lack the field.
    const info = this.companionInfo as any;
    if (info && typeof info.IsHost === 'boolean') {
      (this.renderer as any).isHost = info.IsHost;
      return;
    }
    const surface = surfaceForGame(this.gameId);
    (this.renderer as any).isHost = this.isOwner && surface === 'table';
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
        // Re-instantiate with the CURRENT surface's suffix — the plain
        // solo element was never registered on companion surfaces (only
        // the suffixed module was imported), so instantiating '' here
        // would create an un-upgraded element that renders nothing.
        this._instantiateRenderer(this._surfaceSuffix(this.gameId));
      }
    }
  }

  private _ensureActiveAnimations() {
    if (this._activeAnimations) return;
    this._activeAnimations = new Map();
  }

  private _clearAllAnimatingComponents() {
    if (!this._animator) return;
    const stacks: BoardgameComponentStack[] = this._animator.stackElement?._sharedStackList ?? [];
    for (const stack of stacks) {
      stack.clearAnimatingComponents();
    }
  }

  private _resetAnimating() {
    animHooks.record('gate-open');
    // Clear any existing watchdog timer from a previous animation cycle.
    if (this._animationWatchdogTimer !== null) {
      clearTimeout(this._animationWatchdogTimer);
      this._animationWatchdogTimer = null;
    }
    this._activeAnimations = null;
    this._ensureActiveAnimations();
    this._allAnimationsDoneFired = false;
    this.isAnimating = true;
    this._applyAnimatingToRenderer();
    this.dispatchEvent(new CustomEvent('animating-changed', {
      bubbles: true, composed: true, detail: { value: this.isAnimating }
    }));
    // Start a watchdog timer. If animations complete normally,
    // _notifyAnimationsDone() will clear it before it fires.
    this._animationWatchdogTimer = setTimeout(() => {
      this._animationWatchdogTimer = null;
      if (this._allAnimationsDoneFired) return;
      const pendingComponents: string[] = [];
      if (this._activeAnimations) {
        for (const [ele] of this._activeAnimations) {
          const tag = ele?.tagName?.toLowerCase() ?? 'unknown';
          const id = ele?.id ? `#${ele.id}` : '';
          pendingComponents.push(`${tag}${id}`);
        }
      }
      animHooks.record('watchdog', pendingComponents.join(','));
      console.error(
        `[boardgame-render-game] Animation watchdog timeout: animations did not complete within 4s. ` +
        `Force-firing all-animations-done. Pending components (${pendingComponents.length}): ${pendingComponents.join(', ') || 'none'}`
      );
      this._notifyAnimationsDone();
    }, 4000);
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
    // Animations completed normally — cancel the watchdog timer.
    if (this._animationWatchdogTimer !== null) {
      clearTimeout(this._animationWatchdogTimer);
      this._animationWatchdogTimer = null;
    }
    this._allAnimationsDoneFired = true;
    animHooks.record('gate-close');
    this.isAnimating = false;
    this._applyAnimatingToRenderer();
    this.dispatchEvent(new CustomEvent('animating-changed', {
      bubbles: true, composed: true, detail: { value: this.isAnimating }
    }));
    this.dispatchEvent(new CustomEvent('all-animations-done', { composed: true, bubbles: true }));
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
      // Clear stale faux animating components from any interrupted animation
      // cycle before prepare() captures positions. This prevents old faux
      // components' transitionend from interfering with the new cycle.
      this._clearAllAnimatingComponents();
      this._animator?.prepare();
    }

    // For Lit renderers, set property directly
    (this.renderer as any).state = newState;

    if (newState && !stateWasNull) {
      // Call animateFlip. When all of the things that will be animating have
      // started, check to see if no animations have been registered; if they
      // haven't, then we can advance to the next state immediately.
      this._animator?.animateFlip().then(() => this._nextStateIfNoAnimations());
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

  private _moveFormsChanged(moveForms: MoveForm[] | null) {
    if (!this.renderer) return;
    (this.renderer as any).moveLegality = BoardgameRenderGame._deriveLegality(moveForms);
  }

  private static _deriveLegality(moveForms: MoveForm[] | null): Record<string, MoveLegalityInfo> {
    const result: Record<string, MoveLegalityInfo> = {};
    if (!moveForms) return result;
    for (const form of moveForms) {
      result[form.Name] = {
        legalForPlayer: form.LegalForPlayer ?? false,
        legalForAnyone: form.LegalForAnyone ?? false,
        error: form.LegalForPlayerError,
      };
    }
    return result;
  }

  private async _gameNameChanged(newValue: string) {
    // If there was a state, it might be for a different game type which would
    // cause a render error
    this.state = null;
    this.rendererLoaded = false;
    this._removeRenderer();

    if (!newValue) return;

    const suffix = this._surfaceSuffix(this.gameId);

    try {
      // Use /* @vite-ignore */ to allow fully dynamic imports in dev mode.
      // If a companion-mode suffix is in play (-table / -hand), try the
      // suffixed import first; on failure fall back to the solo renderer
      // with a console warning. This makes solo-mode games safe to load even
      // when a stale surface cookie is present, and surfaces deployment
      // errors (missing -table.ts / -hand.ts on a supporting game) loudly.
      if (suffix) {
        try {
          await import(/* @vite-ignore */ `../../game-src/${newValue}/boardgame-render-game-${newValue}${suffix}.ts`);
          this._instantiateRenderer(suffix);
          return;
        } catch (innerError) {
          console.warn(
            `[boardgame-render-game] surface renderer ${newValue}${suffix} failed to load; falling back to solo:`,
            innerError,
          );
        }
      }
      await import(/* @vite-ignore */ `../../game-src/${newValue}/boardgame-render-game-${newValue}.ts`);
      this._instantiateRenderer('');
    } catch (error) {
      console.error(`Failed to load game renderer for ${newValue}:`, error);
    }
  }

  // _surfaceSuffix returns the filename suffix to add to the renderer
  // module / custom-element name for the current surface, or empty for
  // solo. Pure function of the cookie + query string (see
  // utils/companion-surface.ts, shared with boardgame-game-view).
  private _surfaceSuffix(gameId: string): string {
    const s = surfaceForGame(gameId);
    if (s === 'table') return '-table';
    if (s === 'hand') return '-hand';
    return '';
  }

  private _removeRenderer() {
    if (this.renderer && this._container) {
      this._container.removeChild(this.renderer);
    }
    this.renderer = null;
  }

  private _instantiateRenderer(surfaceSuffix: string = '') {
    // The import loaded! Add it!
    this.rendererLoaded = true;

    const ele = document.createElement(`boardgame-render-game-${this.gameName}${surfaceSuffix}`) as any;

    ele.diagram = this.diagram;
    ele.state = this.state;
    ele.viewingAsPlayer = this.viewingAsPlayer;
    ele.currentPlayerIndex = this.currentPlayerIndex;
    ele.chest = this.chest;
    ele.moveLegality = BoardgameRenderGame._deriveLegality(this.moveForms);
    // Pass game name + ID + companion props through so the Table/Hand
    // view bases can call host endpoints (which require these in the URL
    // path) and render the avatar strip, room code banner, etc.
    ele.gameName = this.gameName;
    ele.gameId = this.gameId;
    ele.isOwner = this.isOwner;
    ele.animating = this.isAnimating;

    // Assign this.renderer BEFORE applying companion props:
    // _applyCompanionPropsToRenderer calls _recomputeIsHost which guards
    // on this.renderer — applying first would silently no-op the isHost
    // computation, leaving the host's Table view without host controls
    // until the next companionInfo change (which may never come in a
    // quiet lobby).
    this.renderer = ele;
    this._applyCompanionPropsToRenderer(ele);

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

      <!-- Suppress the connection-lost dim once the game is finished: the
           socket closing after game end is expected, not an outage, and
           dimming the final scoreboard reads as a broken page. -->
      <div id="loading" ?active="${!this.socketActive && !this.gameFinished}">
        <div>
          <div class="spinner"></div>
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-render-game', BoardgameRenderGame);

export { BoardgameRenderGame };
