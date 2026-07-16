import { LitElement, html, css } from 'lit';
import { property, query } from 'lit/decorators.js';
import { BoardgameComponentAnimator } from './boardgame-component-animator.js';
import type { MoveForm } from '../types/api.js';
import type { MoveLegalityInfo } from '../selectors.js';
import { movePreviewBatch } from '../api.js';
import {
  disabledSpacesFromResults,
  samePreviewSpaces,
  previewOutcome,
} from '../legal/previewLegality.js';
import { surfaceForGame } from '../utils/companion-surface.js';
import { animHooks } from '../utils/anim-test-hooks.js';
import type { MovePreviewTransport, MoveSubmissionGate, MoveTransport } from '../moves/action.js';
import type { TargetPreviewTransport } from '../moves/target-action.js';
import type { PlayerPresentation } from '../status/player-presentation.js';
import { BoardgameBaseGameRenderer } from './boardgame-base-game-renderer.js';
import { BoardgameTableViewBase } from './boardgame-table-view-base.js';
import { BoardgameHandViewBase } from './boardgame-hand-view-base.js';
import type { FullGameState, GameChest } from '../types/boardgame-types.js';
import { retryDelayMs } from '../utils/retry-policy.js';

type HostedState = FullGameState<object, object, object, object, object>;
export type HostedGameRenderer = BoardgameBaseGameRenderer<
  HostedState,
  object,
  string,
  Record<string, object>
>;

/**
 * BoardgameRenderGame dynamically loads and manages game-specific renderers.
 * Handles animation coordination, state synchronization, and loading states.
 */
class BoardgameRenderGame extends LitElement {
  static override styles = css`
    #container {
      position: relative;
    }

    #connection-status {
      box-sizing: border-box;
      width: min(100% - 2rem, 48rem);
      margin: 0.75rem auto;
      padding: 0.75rem 1rem;
      border: 1px solid var(--md-sys-color-outline-variant, #CCC4B8);
      border-radius: 0.75rem;
      color: var(--md-sys-color-on-surface, #1d1b20);
      background: var(--md-sys-color-surface-container-high, #ece6f0);
      z-index: 10;
    }

    #connection-status > div {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0.75rem;
    }

    .spinner {
      flex: 0 0 auto;
      width: 1rem;
      height: 1rem;
      border: 3px solid var(--md-sys-color-outline-variant, #CCC4B8);
      border-top-color: var(--md-sys-color-primary, #2E6B4F);
      border-radius: 50%;
      animation: spin 1s linear infinite;
    }

    button {
      min-height: 2.75rem;
      padding-inline: 1rem;
      border: 1px solid currentColor;
      border-radius: 999px;
      color: var(--md-sys-color-primary, #2E6B4F);
      background: transparent;
      font: inherit;
      font-weight: 600;
      cursor: pointer;
    }

    button:disabled {
      opacity: 0.6;
      cursor: not-allowed;
    }

    @media (prefers-reduced-motion: reduce) {
      .spinner { animation: none; }
    }

    #renderer-error {
      box-sizing: border-box;
      max-width: 48rem;
      margin: 1rem auto;
      padding: 1rem;
      border: 2px solid var(--md-sys-color-error, #ba1a1a);
      border-radius: 0.75rem;
      color: var(--md-sys-color-on-error-container, #410002);
      background: var(--md-sys-color-error-container, #ffdad6);
    }

    #renderer-error h2 {
      margin-block: 0 0.5rem;
      font-size: 1.1rem;
    }

    #renderer-error p {
      margin-block: 0.5rem 0;
      overflow-wrap: anywhere;
    }

    @keyframes spin {
      0% { transform: rotate(0deg); }
      100% { transform: rotate(360deg); }
    }
  `;

  @property({ type: Object })
  state: HostedState | null = null;

  // Scoped timing policy associated with the exact installed game version.
  // The shared animator consumes it as animateBetween's default.
  @property({ type: Object })
  animationContext: import('./companion-sync.js').VersionAnimationContext | null = null;

  @property({ type: Object })
  chest: GameChest<object> | null = null;

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

  @property({ type: Number, attribute: false })
  gameVersion = 0;

  @property({ type: Number, attribute: false })
  snapshotEpoch = 0;

  @property({ type: Number, attribute: false })
  proposingAsPlayer = 0;

  @property({ type: Boolean, attribute: false })
  proposingAsAdmin = false;

  @property({ attribute: false })
  moveTransport: MoveTransport | null = null;

  @property({ attribute: false })
  movePreviewTransport: MovePreviewTransport | null = null;

  @property({ attribute: false })
  targetPreviewTransport: TargetPreviewTransport | null = null;

  @property({ attribute: false })
  moveSubmissionGate: MoveSubmissionGate | null = null;

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

  @property({ attribute: false })
  playerPresentations: readonly PlayerPresentation[] = Object.freeze([]);


  @property({ type: Object, attribute: false })
  renderer: HostedGameRenderer | null = null;

  @property({ type: Boolean })
  rendererLoaded = false;

  @property({ type: String, attribute: false })
  rendererError = '';

  // Imports cannot be aborted, so invalidate their completion whenever
  // navigation selects a different renderer identity or removes this host.
  private _rendererLoadGeneration = 0;

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Number })
  currentPlayerIndex = 0;

  // previewAsPlayer / previewAsAdmin are the SAME (player, admin) perspective the
  // /info fetch uses (game-view passes requestedPlayer + _admin), so the board's
  // legality preview agrees with the moveLegality already displayed. For a
  // non-admin the server ignores the player param (admin=0 -> it uses the seat),
  // so this is a no-op there; in admin "make moves as player N" it makes the
  // preview evaluate as N instead of graying the whole board. (In the edge case
  // where an admin unchecks "make moves as viewing player", a tap submits as the
  // admin index while preview grays per the viewed player — self-consistent with
  // the displayed /info legality, and strictly better than the old wholesale
  // graying.)
  @property({ type: Number })
  previewAsPlayer = 0;

  @property({ type: Boolean })
  previewAsAdmin = false;

  @property({ type: Boolean })
  socketActive = false;

  @property({ type: Number, attribute: false })
  connectionAttempts = 0;

  @property({ type: String, attribute: false })
  connectionError: string | null = null;

  @property({ type: Boolean, attribute: false })
  private _online = true;

  private readonly _networkStateChanged = () => {
    this._online = navigator.onLine;
  };

  @property({ type: Array, attribute: false })
  moveForms: MoveForm[] | null = null;

	@property({ type: String, attribute: false })
	moveInputSchemaFingerprint: string | null = null;

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
  private _animator?: BoardgameComponentAnimator;

  @query('#container')
  private _container?: HTMLElement;

  private _boundComponentWillAnimate?: (e: Event) => void;
  private _boundComponentAnimationDone?: (e: Event) => void;
  // Fired (composed) by the inner renderer via requestPreviewRefresh() when its
  // LOCAL interaction state changes (e.g. a multi-step move selected a source
  // piece) so previewSpec() must be re-evaluated without a state/turn change.
  private _boundPreviewRefreshRequested?: (e: Event) => void;
  private _animationWatchdogTimer: ReturnType<typeof setTimeout> | null = null;
  // Largest declared settle time (delay + duration + endDelay) reported by
  // any gated play() in the current cycle, via the will-animate event. Used
  // to extend the watchdog past a legitimately long cycle (stagger +
  // post-animation-delay + long --animation-length) rather than force-close
  // mid-animation. Reset to 0 at each gate open.
  private _maxExpectedSettleMs = 0;
  // Absolute local-clock (Date.now()-comparable) instant the current
  // watchdog is armed to fire at. Tracked so an incoming will-animate can
  // tell whether a longer play would outlast the deadline and re-arm.
  private _watchdogDeadlineEpoch = 0;
  // The watchdog floor: the gate never gets less than this before firing,
  // even for trivially short cycles. Longer declared cycles push it out.
  private static readonly _WATCHDOG_FLOOR_MS = 4000;
  // Slack added past a declared long cycle's expected settle instant, so
  // normal per-animation jitter/scheduling never trips the watchdog.
  private static readonly _WATCHDOG_MARGIN_MS = 1500;

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);

    this._boundComponentWillAnimate = (e: Event) => this._componentWillAnimate(e as CustomEvent);
    this._boundComponentAnimationDone = (e: Event) => this._componentAnimationDone(e as CustomEvent);

    this.addEventListener('will-animate', this._boundComponentWillAnimate);
    this.addEventListener('animation-done', this._boundComponentAnimationDone);

    // A renderer whose previewSpec() depends on local interaction state
    // (multi-step moves) fires this to force a debounced re-preview.
    this._boundPreviewRefreshRequested = () => this._scheduleRefreshPreview();
    this.addEventListener('preview-refresh-requested', this._boundPreviewRefreshRequested);
    this._activeAnimations = null;
    this._ensureActiveAnimations();
    this._allAnimationsDoneFired = false;
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    window.removeEventListener('online', this._networkStateChanged);
    window.removeEventListener('offline', this._networkStateChanged);
    if (this._boundComponentWillAnimate) {
      this.removeEventListener('will-animate', this._boundComponentWillAnimate);
    }
    if (this._boundComponentAnimationDone) {
      this.removeEventListener('animation-done', this._boundComponentAnimationDone);
    }
    if (this._boundPreviewRefreshRequested) {
      this.removeEventListener('preview-refresh-requested', this._boundPreviewRefreshRequested);
    }
    // Clean up watchdog timer to prevent firing after element is removed.
    if (this._animationWatchdogTimer !== null) {
      clearTimeout(this._animationWatchdogTimer);
      this._animationWatchdogTimer = null;
    }
    // Same for the debounced legality-preview timer: if we're torn down mid-
    // debounce, don't fire a stray movePreviewBatch / write to a dead renderer.
    // Bumping the seq also invalidates any batch already in flight.
    if (this._previewTimer !== null) {
      clearTimeout(this._previewTimer);
      this._previewTimer = null;
    }
    this._clearPreviewRetry();
    this._previewSeq++;
    this._rendererLoadGeneration++;
  }

  override connectedCallback(): void {
    super.connectedCallback();
    this._online = navigator.onLine;
    window.addEventListener('online', this._networkStateChanged);
    window.addEventListener('offline', this._networkStateChanged);
    // A module request invalidated by temporary detachment must be restarted
    // even though Lit sees no renderer-identity property change on reinsertion.
    if (this.hasUpdated && this.gameName && !this.rendererLoaded) {
      void this._rendererIdentityChanged(this.gameName, this.gameId);
    }
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if (changedProperties.has('animationContext') && this._animator) {
      this._animator.animationContext = this.animationContext;
    }

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

    if (changedProperties.has('gameName') || changedProperties.has('gameId')) {
      this._rendererIdentityChanged(this.gameName, this.gameId);
    }

    if (changedProperties.has('defaultAnimationLength')) {
      this._defaultAnimationLengthChanged(this.defaultAnimationLength);
    }

    if (changedProperties.has('moveForms')) {
      this._moveFormsChanged(this.moveForms);
    }

    if (changedProperties.has('moveInputSchemaFingerprint')) {
      this._moveInputSchemaFingerprintChanged(this.moveInputSchemaFingerprint);
    }

    if (changedProperties.has('state')) {
      this._stateChanged(this.state);
    }

    // Refresh the board's target-legality preview when anything it depends on
    // moves: the game state, the move forms (which move types exist / are
    // legal), or who we're viewing as. Debounced + coalesced in the scheduler.
    if (
      changedProperties.has('state') ||
      changedProperties.has('moveForms') ||
      changedProperties.has('viewingAsPlayer') ||
      changedProperties.has('previewAsPlayer') ||
      changedProperties.has('previewAsAdmin') ||
      changedProperties.has('socketActive')
    ) {
      this._scheduleRefreshPreview();
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

    if (changedProperties.has('playerPresentations') && this.renderer) {
      this.renderer.playerPresentations = this.playerPresentations;
    }

    if (changedProperties.has('gameVersion')
      || changedProperties.has('snapshotEpoch')
      || changedProperties.has('proposingAsPlayer')
      || changedProperties.has('proposingAsAdmin')
      || changedProperties.has('moveTransport')
      || changedProperties.has('movePreviewTransport')
      || changedProperties.has('targetPreviewTransport')
      || changedProperties.has('moveSubmissionGate')
      || changedProperties.has('socketActive')) {
      this._applyMoveActionPropsToRenderer();
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
    this.renderer.diagram = newValue;
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

  private _isOwnerChanged(_newValue: boolean) {
    if (!this.renderer) return;
    this._recomputeIsHost();
  }

  private _applyCompanionPropsToRenderer(renderer: HostedGameRenderer) {
    const info = this.companionInfo;
    if (renderer instanceof BoardgameTableViewBase) {
      renderer.seatPresentations = info?.SeatPresentations ?? [];
      renderer.absentPlayers = info?.Absent ?? [];
      renderer.roomCode = info?.RoomCode ?? '';
      renderer.roomLocked = info?.RoomLocked ?? false;
    } else if (renderer instanceof BoardgameHandViewBase) {
      renderer.seatPresentations = info?.SeatPresentations ?? [];
    }
    this._recomputeIsHost();
    this._applyGameOutcomeToRenderer();
  }

  private _applyGameOutcomeToRenderer() {
    if (!this.renderer) return;
    this.renderer.gameFinished = this.gameFinished;
    this.renderer.gameWinners = this.gameWinners;
  }

  private _applyMoveActionPropsToRenderer() {
    if (!this.renderer) return;
    const renderer = this.renderer;
    renderer.gameVersion = this.gameVersion;
    renderer.snapshotEpoch = this.snapshotEpoch;
    renderer.proposingAsPlayer = this.proposingAsPlayer;
    renderer.proposingAsAdmin = this.proposingAsAdmin;
    // A disconnected renderer may be showing a stale snapshot. Fail closed
    // until live state resumes; ExpectedVersion remains the server backstop.
    renderer.moveTransport = this.socketActive ? this.moveTransport : null;
    renderer.movePreviewTransport = this.movePreviewTransport;
    renderer.targetPreviewTransport = this.targetPreviewTransport;
    if (this.moveSubmissionGate) renderer.moveSubmissionGate = this.moveSubmissionGate;
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
    this.renderer.animating = this.isAnimating;
  }

  private _recomputeIsHost() {
    if (!(this.renderer instanceof BoardgameTableViewBase)) return;
    // Prefer the server's own verdict (CompanionInfo.IsHost, computed with
    // the same Owner-or-override + surface-cookie rule the host-action
    // endpoints enforce) so a host promoted via /claimHost sees controls
    // even though they aren't the Owner. Fall back to the local derivation
    // for older payloads that lack the field.
    const info = this.companionInfo;
    if (info && typeof info.IsHost === 'boolean') {
      this.renderer.isHost = info.IsHost;
      return;
    }
    const surface = surfaceForGame(this.gameId);
    this.renderer.isHost = this.isOwner && surface === 'table';
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
    this._animator.clearAnimatingComponents();
  }

  private _resetAnimating() {
    animHooks.record('gate-open');
    // Clear any existing watchdog timer from a previous animation cycle.
    if (this._animationWatchdogTimer !== null) {
      clearTimeout(this._animationWatchdogTimer);
      this._animationWatchdogTimer = null;
    }
    this._maxExpectedSettleMs = 0;
    this._activeAnimations = null;
    this._ensureActiveAnimations();
    this._allAnimationsDoneFired = false;
    this.isAnimating = true;
    this._applyAnimatingToRenderer();
    this.dispatchEvent(new CustomEvent('animating-changed', {
      bubbles: true, composed: true, detail: { value: this.isAnimating }
    }));
    // Arm the watchdog at the floor. Long declared cycles push it out via
    // _componentWillAnimate. If animations complete normally,
    // _notifyAnimationsDone() clears it before it fires.
    this._armWatchdog(BoardgameRenderGame._WATCHDOG_FLOOR_MS);
  }

  // _armWatchdog (re)arms the watchdog to fire `fromNowMs` from now, tracking
  // the absolute deadline so a later will-animate can decide whether to
  // extend it. The watchdog is the invariant backstop: if it ever fires, an
  // awaited animation didn't settle within its own declared budget — a bug.
  private _armWatchdog(fromNowMs: number) {
    if (this._animationWatchdogTimer !== null) {
      clearTimeout(this._animationWatchdogTimer);
      this._animationWatchdogTimer = null;
    }
    this._watchdogDeadlineEpoch = Date.now() + fromNowMs;
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
        `[boardgame-render-game] Animation watchdog timeout: animations did not complete ` +
        `within their declared budget (${fromNowMs}ms). Force-firing all-animations-done. ` +
        `Pending components (${pendingComponents.length}): ${pendingComponents.join(', ') || 'none'}`
      );
      this._notifyAnimationsDone();
    }, fromNowMs);
  }

  private _componentWillAnimate(e: CustomEvent) {
    this._ensureActiveAnimations();
    this._activeAnimations!.set(e.detail.ele, true);
    // Extend the watchdog if this play declares a settle time that would
    // outlast the current deadline. The deadline is the declared expected
    // settle instant plus a fixed margin, floored at _WATCHDOG_FLOOR_MS —
    // so e.g. 15 staggered cards each 2s long (last starts ~5.6s in) no
    // longer trip a flat 4s watchdog mid-flight.
    const expected = e.detail?.expectedSettleMs;
    if (typeof expected === 'number' && expected > this._maxExpectedSettleMs) {
      this._maxExpectedSettleMs = expected;
      const targetEpoch = Date.now() + expected + BoardgameRenderGame._WATCHDOG_MARGIN_MS;
      if (targetEpoch > this._watchdogDeadlineEpoch) {
        this._armWatchdog(targetEpoch - Date.now());
      }
    }
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

  private _stateChanged(newState: HostedState | null) {
    if (!this.renderer) return;
    if (this._animator) {
      this._animator.animationContext = this.animationContext;
    }
    const stateWasNull = this.renderer.state == null;
    if (newState && !stateWasNull) {
      this._resetAnimating();
      // Clear stale faux animating components from any interrupted animation
      // cycle before prepare() captures positions. This prevents old faux
      // components' transitionend from interfering with the new cycle.
      this._clearAllAnimatingComponents();
      this._animator?.prepare();
    }

    // For Lit renderers, set property directly
    this.renderer.state = newState;

    if (newState && !stateWasNull) {
      // Call animateFlip. When all of the things that will be animating have
      // started, check to see if no animations have been registered; if they
      // haven't, then we can advance to the next state immediately.
      this._animator?.animateFlip().then(() => this._nextStateIfNoAnimations());
    }
  }

  private _viewingAsPlayerChanged(newValue: number) {
    if (!this.renderer) return;
    this.renderer.viewingAsPlayer = newValue;
  }

  private _currentPlayerIndexChanged(newValue: number) {
    if (!this.renderer) return;
    this.renderer.currentPlayerIndex = newValue;
  }

  private _chestChanged(newValue: GameChest<object> | null) {
    if (!this.renderer) return;
    this.renderer.chest = newValue;
  }

  private _moveFormsChanged(moveForms: MoveForm[] | null) {
    if (!this.renderer) return;
    this.renderer.moveLegality = BoardgameRenderGame._deriveLegality(moveForms);
  }

  private _moveInputSchemaFingerprintChanged(fingerprint: string | null) {
    if (!this.renderer) return;
    this.renderer.serverMoveInputSchemaFingerprint = fingerprint;
  }

  // Board legality preview (movePreviewBatch). Kept view-local rather than in
  // Redux: preview results are ephemeral, tied to the exact candidate set of the
  // current head state, and consumed only by the renderer — putting them in the
  // store would add a slice plus staleness/keying concerns for no shared
  // consumer. _previewTimer debounces bursts of state/turn changes into one
  // request; _previewSeq drops a response the game has already moved past.
  private _previewTimer: ReturnType<typeof setTimeout> | null = null;
  private _previewRetryTimer: ReturnType<typeof setTimeout> | null = null;
  private _previewRetryAttempt = 0;
  private _previewSeq = 0;

  private _scheduleRefreshPreview(resetRetry = true) {
    if (resetRetry) {
      this._clearPreviewRetry();
      this._previewRetryAttempt = 0;
    }
    if (this._previewTimer !== null) clearTimeout(this._previewTimer);
    this._previewTimer = setTimeout(() => {
      this._previewTimer = null;
      void this._refreshPreview();
    }, 150);
  }

  // _refreshPreview asks the game renderer for its preview candidates
  // (previewSpec(), null unless the game opted in), batch-checks their legality
  // on the server WITHOUT applying anything, and pushes the illegal spaces back
  // onto the renderer for graying. It previews from the (previewAsPlayer,
  // previewAsAdmin) perspective — the same one /info uses — so graying agrees
  // with the displayed legality in both normal and admin "make moves as player
  // N" modes.
  private async _refreshPreview() {
    const renderer = this.renderer;
    if (!renderer) return;
    if (!this.gameName || !this.gameId) return;

    const spec = renderer.previewSpec();
    if (!spec || spec.candidates.length === 0) {
      // Opted out (or nothing to check right now) — clear any stale graying.
      // Bump the seq FIRST so a batch still in flight from a prior refresh can't
      // resolve later and re-gray the board we just cleared.
      this._previewSeq++;
      if (renderer.previewDisabledSpaces.length) {
        renderer.previewDisabledSpaces = [];
      }
      return;
    }

    if (!this.socketActive) {
      this._previewSeq++;
      this._failPreviewClosed(renderer, spec.candidates);
      return;
    }

    const seq = ++this._previewSeq;
    // Preview from the same (player, admin) perspective /info uses, so graying
    // agrees with the displayed legality. For a non-admin the server ignores the
    // player param (admin=0), so this resolves to the seat exactly as before.
    let response;
    try {
      response = await movePreviewBatch(
        this.gameName,
        this.gameId,
        spec.moveName,
        spec.candidates.map((c) => ({ Args: c.args })),
        { player: this.previewAsPlayer, admin: this.previewAsAdmin ? 1 : 0 },
      );
    } catch (error) {
      console.warn('Target legality preview failed:', error);
      response = { data: undefined };
    }
    const outcome = previewOutcome({
      startedSeq: seq,
      currentSeq: this._previewSeq,
      rendererStillMounted: this.renderer === renderer,
      hasData: !!response.data,
    });
    if (outcome === 'keep-on-error') {
      this._failPreviewClosed(renderer, spec.candidates);
      this._schedulePreviewRetry(renderer);
      return;
    }
    if (outcome !== 'apply' || !response.data) return;
    this._clearPreviewRetry();
    this._previewRetryAttempt = 0;

    // Only reassign (and thus re-render the board) when the grayed set actually
    // changed.
    const next = disabledSpacesFromResults(spec.candidates, response.data.Results);
    if (!samePreviewSpaces(next, renderer.previewDisabledSpaces)) {
      renderer.previewDisabledSpaces = next;
    }
  }

  private _failPreviewClosed(
    renderer: HostedGameRenderer,
    candidates: readonly { readonly space: number }[],
  ): void {
    const disabled = candidates.map(candidate => candidate.space);
    if (!samePreviewSpaces(disabled, renderer.previewDisabledSpaces)) {
      renderer.previewDisabledSpaces = disabled;
    }
  }

  private _schedulePreviewRetry(renderer: HostedGameRenderer): void {
    if (this._previewRetryTimer !== null || !this.socketActive) return;
    const delay = retryDelayMs(this._previewRetryAttempt++);
    this._previewRetryTimer = setTimeout(() => {
      this._previewRetryTimer = null;
      if (this.renderer !== renderer || !this.socketActive) return;
      this._scheduleRefreshPreview(false);
    }, delay);
  }

  private _clearPreviewRetry(): void {
    if (this._previewRetryTimer === null) return;
    clearTimeout(this._previewRetryTimer);
    this._previewRetryTimer = null;
  }

  private static _deriveLegality(moveForms: MoveForm[] | null): Record<string, MoveLegalityInfo> {
    const result: Record<string, MoveLegalityInfo> = {};
    if (!moveForms) return result;
    for (const form of moveForms) {
      result[form.Name] = {
        legalForPlayer: form.LegalForPlayer ?? false,
        legalForAnyone: form.LegalForAnyone ?? false,
        error: form.LegalForPlayerError,
        preconditions: form.Preconditions,
      };
    }
    return result;
  }

  private async _rendererIdentityChanged(gameName: string, gameId: string, retry = false) {
    const generation = ++this._rendererLoadGeneration;
    // Route-owned state is reset atomically by Redux. Do not mutate the host's
    // input here: retaining it is what makes a renderer retry useful.
    this.rendererLoaded = false;
    this.rendererError = '';
    this._removeRenderer();

    if (!gameName) return;
    if (!/^[a-z][a-z0-9]*$/.test(gameName)) {
      this.rendererError = `Invalid renderer game name ${JSON.stringify(gameName)}; expected lowercase letters and digits`;
      console.error(this.rendererError);
      return;
    }

    const suffix = this._surfaceSuffix(gameId);
    const baseModulePath = `../../game-src/${gameName}/boardgame-render-game-${gameName}${suffix}.ts`;
    const modulePath = retry ? `${baseModulePath}?retry=${generation}` : baseModulePath;

    try {
      // Use /* @vite-ignore */ to allow fully dynamic imports in dev mode.
      // A requested companion surface is a behavioral contract. Silently
      // substituting the solo renderer hides deployment mistakes and leaves
      // the surrounding companion chrome in a contradictory mode.
      await this._loadRendererModule(modulePath);
      if (!this._rendererLoadIsCurrent(generation, gameName, gameId)) return;
      this._instantiateRenderer(suffix);
    } catch (error) {
      if (!this._rendererLoadIsCurrent(generation, gameName, gameId)) return;
      if (!this.rendererError) {
        this.rendererError = `Failed to load renderer module ${modulePath}: ${this._errorMessage(error)}`;
      }
      console.error(`Failed to load game renderer for ${gameName}:`, error);
    }
  }

  private retryRenderer(): void {
    const suffix = this._surfaceSuffix(this.gameId);
    const tagName = `boardgame-render-game-${this.gameName}${suffix}`;
    if (customElements.get(tagName)) {
      try {
        this._instantiateRenderer(suffix);
        return;
      } catch (error) {
        this.rendererError = this._errorMessage(error);
      }
    }
    void this._rendererIdentityChanged(this.gameName, this.gameId, true);
  }

  private retryConnection(): void {
    this.dispatchEvent(new CustomEvent('retry-connection', {
      bubbles: true,
      composed: true,
    }));
  }

  private _rendererLoadIsCurrent(generation: number, gameName: string, gameId: string): boolean {
    return generation === this._rendererLoadGeneration
      && gameName === this.gameName
      && gameId === this.gameId
      && this.isConnected;
  }

  private _loadRendererModule(modulePath: string): Promise<unknown> {
    return import(/* @vite-ignore */ modulePath) as Promise<unknown>;
  }

  private _errorMessage(error: unknown): string {
    const message = error instanceof Error ? error.message : String(error);
    const normalized = message.trim() || 'unknown renderer error';
    return normalized.length <= 500 ? normalized : `${normalized.slice(0, 497)}...`;
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
    const tagName = `boardgame-render-game-${this.gameName}${surfaceSuffix}`;
    const constructor = customElements.get(tagName);
    if (!constructor) {
      this.rendererError =
        `Renderer module loaded but did not register <${tagName}>; ` +
        `use the generated registration decorator for this exact surface`;
      throw new Error(this.rendererError);
    }
    if (!(constructor.prototype instanceof BoardgameBaseGameRenderer)) {
      this.rendererError =
        `Renderer <${tagName}> must extend the generated renderer base; ` +
        `use GameRenderer, TableRenderer, or HandRenderer with its generated registration decorator`;
      throw new Error(this.rendererError);
    }
    this.rendererError = '';
    this.rendererLoaded = true;
    if (!this.active) return;

    const ele = document.createElement(tagName) as HostedGameRenderer;

    ele.diagram = this.diagram;
    ele.state = this.state;
    ele.viewingAsPlayer = this.viewingAsPlayer;
    ele.currentPlayerIndex = this.currentPlayerIndex;
    ele.playerPresentations = this.playerPresentations;
    ele.chest = this.chest;
    ele.moveLegality = BoardgameRenderGame._deriveLegality(this.moveForms);
    ele.serverMoveInputSchemaFingerprint = this.moveInputSchemaFingerprint;
    // Pass game name + ID + companion props through so the Table/Hand
    // view bases can call host endpoints (which require these in the URL
    // path) and render the avatar strip, room code banner, etc.
    ele.gameName = this.gameName;
    ele.gameId = this.gameId;
    ele.gameVersion = this.gameVersion;
    ele.snapshotEpoch = this.snapshotEpoch;
    ele.proposingAsPlayer = this.proposingAsPlayer;
    ele.proposingAsAdmin = this.proposingAsAdmin;
    ele.moveTransport = this.socketActive ? this.moveTransport : null;
    ele.movePreviewTransport = this.movePreviewTransport;
    ele.targetPreviewTransport = this.targetPreviewTransport;
    if (this.moveSubmissionGate) ele.moveSubmissionGate = this.moveSubmissionGate;
    ele.animating = this.isAnimating;
    if (this._animator) {
      this._animator.animationContext = this.animationContext;
    }

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

    // Kick off the initial target-legality preview for the freshly-mounted
    // renderer (updated()'s change-driven refresh won't fire for props that were
    // already set before this renderer existed).
    this._scheduleRefreshPreview();
  }

  override render() {
    return html`
      <boardgame-component-animator
        id="animator"
        .ancestorOffsetParent="${this._container ?? null}">
      </boardgame-component-animator>

      ${this.rendererLoaded ? null : this.rendererError ? html`
        <section id="renderer-error" role="alert" aria-live="assertive">
          <h2>Game renderer unavailable</h2>
          <p>${this.rendererError}</p>
          <p>Run <code>boardgame-util check-client</code> and fix every reported diagnostic.</p>
          <button type="button" @click=${this.retryRenderer}>Retry renderer</button>
        </section>
      ` : html`
        <div>
          <h2>Diagram of ${this.gameName}</h2>
          <pre>${this.diagram}</pre>
        </div>
      `}

      <div id="container">
        <!-- Dynamic renderer will be inserted here -->
      </div>

      <!-- Suppress the connection-lost dim once the game is finished: the
           socket closing after game end is expected, not an outage, and
           dimming the final scoreboard reads as a broken page. -->
      ${!this.socketActive && !this.gameFinished ? html`
      <section id="connection-status" role="status" aria-live="polite" aria-atomic="true">
        <div>
          <div class="spinner" aria-hidden="true"></div>
          <span>${!this._online
            ? 'You are offline. The game will reconnect automatically when the network returns.'
            : this.connectionAttempts > 0
            ? `Connection lost. Reconnecting (attempt ${this.connectionAttempts})…`
            : 'Connecting to the live game…'}</span>
          <button type="button" ?disabled=${!this._online} @click=${this.retryConnection}>Retry now</button>
        </div>
      </section>
      ` : null}
    `;
  }
}

customElements.define('boardgame-render-game', BoardgameRenderGame);

export { BoardgameRenderGame };
