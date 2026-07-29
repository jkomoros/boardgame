import { LitElement, html, css } from 'lit';
import { property, query, state } from 'lit/decorators.js';
import './boardgame-component-animator.js';
import type { BoardgameComponentAnimator } from './boardgame-component-animator.js';
import './boardgame-effect-layer.js';
import type { BoardgameEffectLayer } from './boardgame-effect-layer.js';
import type { ClientMove, MoveForm } from '../types/api.js';
import { createEffectTransitionContext } from '../effects/effect-spec.js';
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
import type { ProjectedMoveChoicesWire } from '../types/api.js';
import {
  defaultMessageResolver,
  type MessageResolver,
  type MoveChoiceProjectionTypes,
  type ProjectedMoveChoices,
} from '../moves/projected-choices.js';
import './boardgame-projected-choices.js';
import { BoardgameBaseGameRenderer } from './boardgame-base-game-renderer.js';
import { BoardgameTableViewBase } from './boardgame-table-view-base.js';
import { BoardgameHandViewBase } from './boardgame-hand-view-base.js';
import type { FullGameState, GameChest } from '../types/boardgame-types.js';
import { retryDelayMs } from '../utils/retry-policy.js';
import { compileMotionTransferDeclarations } from '../motion/transfer.js';
import { compileMotionRelease } from '../motion/release.js';
import { AnimationGate, type AnimationGateCallbacks } from '../motion/animation-gate.js';
import { AnimatableRegistry } from '../motion/animatable-registry.js';

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

    #container.with-projected-choice-tray {
      /* Keep ordinary-flow board content scrollable above the fixed action
         tray. Full-viewport game renderers still treat the tray as a modal-ish
         overlay, but can never place the only controls below the fold. */
      padding-bottom: var(--projected-choice-tray-height, 0px);
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
  // The shared animator consumes it as fly()'s default.
  @property({ type: Object, attribute: 'animation-context' })
  animationContext: import('./companion-sync.js').VersionAnimationContext | null = null;

  @property({ type: Object, attribute: false })
  transitionMove: ClientMove | null = null;

  @property({ type: Object })
  chest: GameChest<object> | null = null;

  @property({ type: Boolean })
  active = false;

  @property({ type: String })
  diagram = '';

  @property({ type: String, attribute: 'game-name' })
  gameName = '';

  // gameId is needed to read the per-game surface cookie (surface_<gameId>)
  // for companion-mode routing. Empty string means we haven't been told the
  // gameID yet (the loader will operate as solo until it's set).
  @property({ type: String, attribute: 'game-id' })
  gameId = '';

  @property({ type: Number, attribute: false })
  gameVersion = 0;

  @property({ type: Number, attribute: false })
  snapshotEpoch = 0;

  @property({ type: Object, attribute: false })
  projectedMoveChoicesWire: ProjectedMoveChoicesWire | null = null;

  @property({ attribute: false })
  messageResolver: MessageResolver = defaultMessageResolver;

  /** Opaque manager-issued identity for one installed animation cycle. */
  @property({ type: Number, attribute: false })
  motionCycleId = 0;

  /** True when the renderer's successor-aware compatibility hook owns cutover. */
  @property({ type: Boolean, attribute: false })
  legacyAnimationOverlapConfigured = false;

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

  @property({ type: Object, attribute: 'companion-info' })
  companionInfo: import('../types/store').CompanionInfo | null = null;

  // isOwner is the doGameInfo IsOwner bool — true if the authenticated
  // user is the game's Owner. Pass-through; the surface renderer combines
  // this with its own surface-cookie check to compute isHost.
  @property({ type: Boolean, attribute: 'is-owner' })
  isOwner = false;

  // gameFinished/gameWinners mirror the game record's Finished/Winners so
  // renderers can show an ending (winner banner, you-won/lost) without
  // bespoke plumbing. Winners are player indexes.
  @property({ type: Boolean, attribute: 'game-finished' })
  gameFinished = false;

  @property({ type: Array, attribute: 'game-winners' })
  gameWinners: number[] = [];

  @property({ attribute: false })
  playerPresentations: readonly PlayerPresentation[] = Object.freeze([]);


  @property({ type: Object, attribute: false })
  renderer: HostedGameRenderer | null = null;

  @property({ type: Boolean, attribute: 'renderer-loaded' })
  rendererLoaded = false;

  @property({ type: String, attribute: false })
  rendererError = '';

  @state()
  private projectedChoiceTrayHeight = 0;

  // Imports cannot be aborted, so invalidate their completion whenever
  // navigation selects a different renderer identity or removes this host.
  private _rendererLoadGeneration = 0;
  private _rendererSurfaceSuffix: '' | '-table' | '-hand' = '';

  @property({ type: Number, attribute: 'viewing-as-player' })
  viewingAsPlayer = 0;

  @property({ type: Number, attribute: 'current-player-index' })
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
  @property({ type: Number, attribute: 'preview-as-player' })
  previewAsPlayer = 0;

  @property({ type: Boolean, attribute: 'preview-as-admin' })
  previewAsAdmin = false;

  @property({ type: Boolean, attribute: 'socket-active' })
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

  @property({ type: Number, attribute: 'default-animation-length' })
  defaultAnimationLength = 0;

  private _activeMotionCycleId = 0;

  // isAnimating reflects whether the animation gate is currently open (an
  // animation cycle is in flight). Reflected to the `is-animating` attribute
  // so tests and ancestor CSS can observe it without reaching into internals,
  // and mirrored via `animating-changed` so ancestors (boardgame-game-view)
  // can wire it into move-disabling UI without polling. See #721.
  @property({ type: Boolean, reflect: true, attribute: 'is-animating' })
  isAnimating = false;

  // Ambient discovery surface for animatable items NOT tracked by the
  // component animator's own stack-component bookkeeping (standalone dice,
  // status-text wrappers, fading-text, tokens, ... -- #714's "non-component"
  // gap, Task 9). BoardgameAnimatableItem's connected/disconnectedCallback
  // walk up to find this property (same walk shape as the ambient
  // animationContext lookup, factored into _ambientLookup -- see
  // boardgame-animatable-item.ts) and self-register/unregister. Reset at
  // cycle start (_resetAnimating) -- this REPLACES nothing: the animator's
  // own component iteration (_clearAllAnimatingComponents) remains
  // authoritative for stack-tracked components.
  readonly animatableRegistry = new AnimatableRegistry();

  @query('#animator')
  private _animator?: BoardgameComponentAnimator;

  @query('#effects')
  private _effects?: BoardgameEffectLayer;

  @query('#container')
  private _container?: HTMLElement;

  private _boundComponentWillAnimate?: (e: Event) => void;
  private _boundComponentAnimationDone?: (e: Event) => void;
  private _lastPresentationTransitionKey = '';
  // Fired (composed) by the inner renderer via requestPreviewRefresh() when its
  // LOCAL interaction state changes (e.g. a multi-step move selected a source
  // piece) so previewSpec() must be re-evaluated without a state/turn change.
  private _boundPreviewRefreshRequested?: (e: Event) => void;

  // The animation-completion gate (see src/motion/animation-gate.ts). The
  // callbacks below preserve, verbatim, the side effects that used to live
  // inline in _resetAnimating/_notifyAnimationsDone/the watchdog timeout.
  private readonly _gate = new AnimationGate({
    onOpen: () => this._onGateOpen(),
    onAllDone: () => this._onGateAllDone(),
    onWatchdog: (pending, budgetMs) => this._onGateWatchdog(pending, budgetMs),
    setTimer: (cb, ms) => setTimeout(cb, ms),
    clearTimer: (handle) => clearTimeout(handle as ReturnType<typeof setTimeout>),
    now: () => Date.now(),
  } satisfies AnimationGateCallbacks);

  private _onGateOpen(): void {
    animHooks.record('gate-open');
    this.isAnimating = true;
    this._applyAnimatingToRenderer();
    this.dispatchEvent(new CustomEvent('animating-changed', {
      bubbles: true, composed: true, detail: { value: this.isAnimating }
    }));
  }

  private _onGateAllDone(): void {
    animHooks.record('gate-close');
    this.isAnimating = false;
    this._applyAnimatingToRenderer();
    this.dispatchEvent(new CustomEvent('animating-changed', {
      bubbles: true, composed: true, detail: { value: this.isAnimating }
    }));
    this.dispatchEvent(new CustomEvent('all-animations-done', {
      composed: true,
      bubbles: true,
      detail: Object.freeze({ cycleId: this._activeMotionCycleId }),
    }));
  }

  private _onGateWatchdog(pending: readonly string[], budgetMs: number): void {
    animHooks.record('watchdog', pending.join(','));
    console.error(
      `[boardgame-render-game] Animation watchdog timeout: animations did not complete ` +
      `within their declared budget (${budgetMs}ms). Force-firing all-animations-done. ` +
      `Pending components (${pending.length}): ${pending.join(', ') || 'none'}`
    );
  }

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
    this.addEventListener('projected-choices-changed', this._projectedChoicesChanged);
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
    this.removeEventListener('projected-choices-changed', this._projectedChoicesChanged);
    // Clean up watchdog timer to prevent firing after element is removed.
    this._gate.dispose();
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
      this._configureEffectLayer();
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

    if (changedProperties.has('projectedMoveChoicesWire') && this.renderer) {
      this.renderer.projectedMoveChoicesWire = this.projectedMoveChoicesWire;
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
    const desiredSurface = this._surfaceSuffix(this.gameId);
    if (this.gameName && desiredSurface !== this._rendererSurfaceSuffix) {
      void this._rendererIdentityChanged(this.gameName, this.gameId);
      return;
    }
    if (!this.renderer) return;
    this._applyCompanionPropsToRenderer(this.renderer);
  }

  /** Installs one authoritative game-info snapshot at the dynamic-renderer
   * boundary. The property assignment preserves ordinary Lit composition;
   * the synchronous propagation prevents a mounted renderer from retaining
   * stale room/Table authority metadata until some unrelated render. */
  installCompanionInfo(info: import('../types/store').CompanionInfo | null): void {
    this.companionInfo = info;
    this._companionInfoChanged(info);
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
    this._configureEffectLayer();
  }

  // _applyAnimatingToRenderer mirrors isAnimating onto the renderer so the
  // Table/Hand view bases can gate outcome/verdict rendering on it (#798
  // final piece): the outcome must never appear while the last animation
  // cycle (e.g. the winning card landing) is still in flight. Called at
  // both gate flips (_onGateOpen / _onGateAllDone) and at
  // renderer instantiation so a renderer created mid-cycle (or finished
  // and re-instantiated on a surface switch) starts with the correct value
  // rather than defaulting to false.
  private _applyAnimatingToRenderer() {
    if (!this.renderer) return;
    this.renderer.animating = this.isAnimating;
  }

  private _recomputeIsHost() {
    if (!(this.renderer instanceof BoardgameTableViewBase)) return;
    // Host authority is a server verdict backed by the active Table lease.
    // Never reconstruct it from presentation state: a stale local `table`
    // surface is not a capability and must not expose controls optimistically.
    this.renderer.isHost = this.companionInfo?.IsHost === true;
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

  private _clearAllAnimatingComponents() {
    if (!this._animator) return;
    this._animator.clearAnimatingComponents();
  }

  // Opens a fresh gate cycle under the current motionCycleId. Kept as a
  // named method (not inlined) because animation tests deliberately reach
  // it directly (it's TS-private, not JS-private) to open the completion
  // gate in isolation, without incidental FLIP animations.
  //
  // Before opening the gate, settle every ambiently-registered animatable
  // item (Task 9, #714): force-finish any animation left over from a prior
  // cycle and install this cycle's animationContext directly, mirroring
  // what the animator already does for its own tracked stack components.
  // This is the piece that was missing for standalone items (a die, a
  // status-text) -- without it, a same-cycle interruption (two state
  // installs landing before the first one's animations naturally settle,
  // see _stateChanged's cycle-id-change branch) force-closes the GATE but
  // leaves an untracked item's own WAAPI animation physically running,
  // so its next play() overlaps the stale one. finishGatedAnimations() is a
  // no-op for an already-settled item, so this is invisible in steady
  // state.
  //
  // finishGatedAnimations (not finishAllAnimations): this sweep interrupts a
  // stale CYCLE, so it must only force-settle that cycle's GATED
  // participants. An UNGATED ambient loop -- an infinite highlight throb on
  // an active/highlighted token, started with { gated: false } -- was never a
  // cycle participant and must keep running across the cycle. Sweeping it with
  // finishAllAnimations cancelled it every state change, and since the token's
  // active/highlighted did not change nothing re-armed it, so a highlighted
  // token stopped glowing after the first move (the retired CSS @keyframes
  // throb was class-driven and immune). See evidence pack
  // docs/superpowers/specs/evidence/2026-07-26-ambient-animation-sweep.md.
  private _resetAnimating() {
    this._activeMotionCycleId = this.motionCycleId;
    for (const item of this.animatableRegistry.items()) {
      item.finishGatedAnimations();
      item.animationContext = this.animationContext;
    }
    this._gate.open(this._activeMotionCycleId);
  }

  private _componentWillAnimate(e: CustomEvent) {
    const ele = e.detail.ele as HTMLElement;
    const tag = ele?.tagName?.toLowerCase() ?? 'unknown';
    const id = ele?.id ? `#${ele.id}` : '';
    this._gate.willAnimate(ele, `${tag}${id}`, e.detail?.expectedSettleMs);
  }

  private _componentAnimationDone(e: CustomEvent) {
    this._gate.animationDone(e.detail.ele);
  }

  // gateWillAnimate/gateAnimationDone (Task 10, #714's second Phase 2 gap):
  // thin public delegates into this._gate, for boardgame-game-view to pipe
  // will-animate/animation-done bubbling out of boardgame-player-roster --
  // a DOM SIBLING of this element, so its own composed bubble path never
  // reaches the will-animate/animation-done listeners this element installs
  // on itself in firstUpdated(). game-view guards the will-animate
  // direction on this.isAnimating (only forward while a board cycle is
  // already open, so a roster animation outside any cycle -- e.g.
  // hover-triggered -- can never open/wedge the gate) and always forwards
  // animation-done (a participant admitted at open must be able to settle).
  // Kept identical in shape to _componentWillAnimate/_componentAnimationDone
  // above rather than merged with them: those remain this element's OWN
  // will-animate/animation-done listeners for its own subtree, unaffected
  // by roster piping.
  gateWillAnimate(e: CustomEvent): void {
    this._componentWillAnimate(e);
  }

  gateAnimationDone(e: CustomEvent): void {
    this._componentAnimationDone(e);
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
    if (newState) {
      // A NEW-cycle install that lands while the previous cycle's gate is
      // still open is a designed destructive cutover (motion release /
      // legacy overlap admit the successor before the current cycle
      // settles). The interrupted cycle must still complete its lifecycle:
      // close it under its OWN id before adopting the new cycle id below.
      // game-view ignores all-animations-done for a stale cycleId, so no
      // successor bundle is released early; without this close the
      // interrupted gate-open is never matched and the completion
      // accounting wedges permanently.
      //
      // The cycle-id-change condition is load-bearing: state also installs
      // WITHOUT a new motion cycle (doGameInfo refreshes via refresh-data /
      // requested-player / admin-mode changes). There motionCycleId is
      // unchanged, so this close would carry the STILL-CURRENT id --
      // game-view's _forwardCycleRelease would forward it and release a
      // queued successor bundle early, cutting an in-flight animation
      // short. Only a genuine cycle handoff closes the previous gate here.
      if (this.isAnimating && this.motionCycleId !== this._activeMotionCycleId) {
        this._gate.close(this._activeMotionCycleId);
      }
      this._activeMotionCycleId = this.motionCycleId;
    }
    if (this._animator) {
      this._animator.animationContext = this.animationContext;
    }
    // Refresh snapshot-derived seed/timing configuration even when the
    // animation context itself remains null across consecutive versions.
    this._configureEffectLayer();
    const previousState = this.renderer.state;
    const stateWasNull = previousState == null;
    const beforeAnchors = this._effects?.captureNamedAnchors() ?? new Map();
    if (newState && !stateWasNull) {
      this._effects?.cancelTransitionEffects();
      // Open a fresh gate cycle only if one is not already open. After a
      // genuine cycle handoff the block above just closed the old gate, so
      // this reopens for the new cycle as always. But a SAME-cycle
      // reinstall (doGameInfo refresh) landing while the gate is open must
      // join the open cycle, not open a second one: resetting here would
      // record an unmatched gate-open (permanently skewing the cumulative
      // open/close accounting) and clear the live participant map, so the
      // in-flight animations' completions could no longer close the gate
      // and only the watchdog would end the cycle.
      if (!this.isAnimating) {
        this._resetAnimating();
      }
      // Clear stale faux animating components from any interrupted animation
      // cycle before prepare() captures positions. This prevents old faux
      // components' transitionend from interfering with the new cycle.
      this._clearAllAnimatingComponents();
      this._animator?.prepare();
      // prepare() publishes cancellation for the old FLIP generation. Open
      // the new effect epoch only afterwards so stale same-subject outcomes
      // cannot satisfy or poison the next transition's motion anchors.
      this._effects?.beginMotionTransition(true);
    } else if (newState) {
      this._effects?.beginMotionTransition(false);
    }

    // For Lit renderers, set property directly
    this.renderer.state = newState;

    const presentationPlanning = newState
      ? this._planTransitionPresentation(this.renderer, previousState, newState, beforeAnchors)
      : Promise.resolve();

    if (newState && !stateWasNull) {
      // Register motion-point descriptors before the animator publishes its
      // synchronous planned/started events. Effect planning remains isolated:
      // either fulfillment or rejection releases structural playback.
      const renderer = this.renderer;
      const cycleId = this.motionCycleId;
      const startStructuralMotion = () => {
        if (this.renderer !== renderer || renderer.state !== newState
          || this.motionCycleId !== cycleId) return;
        this._animator?.animateFlip().then(() => this._gate.settleIfEmpty(cycleId));
      };
      void presentationPlanning.then(startStructuralMotion, startStructuralMotion);
    }
  }

  private async _planTransitionPresentation(
    renderer: HostedGameRenderer,
    before: HostedState | null,
    after: HostedState,
    beforeAnchors: import('./boardgame-effect-layer.js').EffectAnchorSnapshot,
  ): Promise<void> {
    const key = `${this.gameId}:${this.snapshotEpoch}:${this.gameVersion}`;
    if (this._lastPresentationTransitionKey === key) return;
    await renderer.updateComplete;
    if (this.renderer !== renderer || renderer.state !== after) return;
    if (this._lastPresentationTransitionKey === key) return;
    this._lastPresentationTransitionKey = key;
    this._effects?.installBeforeAnchors(beforeAnchors);
    const context = createEffectTransitionContext<HostedState, string>({
      before,
      after,
      move: this.transitionMove,
      version: this.gameVersion,
      snapshotEpoch: this.snapshotEpoch,
    });
    if (before !== null) {
      try {
        const cohorts = renderer.motionCohortsForTransition(context);
        if (!Array.isArray(cohorts)) {
          throw new Error('motionCohortsForTransition() must return a readonly array');
        }
        this._animator?.installMotionCohorts(cohorts);
      } catch (error) {
        // prepare() already cleared any prior declaration, so doing nothing is
        // an atomic fallback to compatibility stack timing.
        console.error('[motion] transition cohort planning failed:', error);
      }
      try {
        const transfers = compileMotionTransferDeclarations(
          renderer.motionTransfersForTransition(context),
        );
        this._animator?.installMotionTransfers(transfers);
      } catch (error) {
        // The compiler is atomic: malformed intent starts no partial batch.
        console.error('[motion] transition transfer planning failed:', error);
      }
      if (this.animationContext === null && !this.legacyAnimationOverlapConfigured) {
        try {
          const release = renderer.motionReleaseForTransition(context);
          this._animator?.installMotionRelease(
            release === null ? null : compileMotionRelease(release),
            this.motionCycleId,
          );
        } catch (error) {
          // Malformed policy fails closed to ordinary cycle settlement.
          this._animator?.installMotionRelease(null, this.motionCycleId);
          console.error('[motion] transition release planning failed:', error);
        }
      }
    }
    if (!this._effects) return;
    try {
      const effects = renderer.effectsForTransition(context);
      if (!Array.isArray(effects)) {
        throw new Error('effectsForTransition() must return a readonly array');
      }
      for (const effect of effects) this._effects.playTransition(effect);
    } catch (error) {
      console.error('[effects] transition planning failed:', error);
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
    this._rendererSurfaceSuffix = suffix;
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
  private _surfaceSuffix(gameId: string): '' | '-table' | '-hand' {
    const s = surfaceForGame(gameId, this.companionInfo?.CompanionMode);
    if (s === 'table') return '-table';
    if (s === 'hand') return '-hand';
    return '';
  }

  private _removeRenderer() {
    this._effects?.cancelAll();
    this._lastPresentationTransitionKey = '';
    if (this.renderer && this._container) {
      this._container.removeChild(this.renderer);
    }
    this.renderer = null;
  }

  private _configureEffectLayer(): void {
    if (!this._effects) return;
    this._effects.configure({
      anchorRoot: this.renderer?.renderRoot ?? null,
      seedScope: `${this.gameId}:${this.snapshotEpoch}:${this.gameVersion}`,
      theme: this.renderer?.effectTheme() ?? {},
      animationContext: this.animationContext,
      motionSource: this._animator ?? null,
    });
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
    ele.projectedMoveChoicesWire = this.projectedMoveChoicesWire;
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
    this._configureEffectLayer();
    if (this.state) {
      void this._planTransitionPresentation(ele, null, this.state, new Map());
    }

    // Only try to fire if there's a state. If it's the first time this
    // session we load the renderer, this will probably happen after the first
    // non-nil state is installed (it takes time to download the component), so
    // we'll need to ask for the next state. But if you load the same game type
    // again, the renderer will load immediately, most likely before the state
    // is installed. If we called this._gate.close() before there's a
    // state, it would be useless (and would prevent it from firing later).
    if (this.state) {
      // Sometimes the renderer is instantiated after the state is already
      // databound--which means that `all-animations-done` won't have fired.
      // gate.close() is a no-op if it's already fired.
      window.requestAnimationFrame(() => this._gate.close(this._activeMotionCycleId));
    }

    // Kick off the initial target-legality preview for the freshly-mounted
    // renderer (updated()'s change-driven refresh won't fire for props that were
    // already set before this renderer existed).
    this._scheduleRefreshPreview();
  }

  override render() {
    const projectedChoices = (this.renderer?.choices ?? null) as
      ProjectedMoveChoices<MoveChoiceProjectionTypes> | null;
    const showProjectedChoiceTray = projectedChoices !== null
      && (projectedChoices.status === 'failed' || projectedChoices.all().length > 0);
    return html`
      <boardgame-component-animator
        id="animator"
        .ancestorOffsetParent="${this._container ?? null}">
      </boardgame-component-animator>
      <boardgame-effect-layer id="effects"></boardgame-effect-layer>

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

      <div
        id="container"
        class=${showProjectedChoiceTray ? 'with-projected-choice-tray' : ''}
        style=${showProjectedChoiceTray
          ? `--projected-choice-tray-height: ${this.projectedChoiceTrayHeight}px`
          : ''}>
        <!-- Dynamic renderer will be inserted here -->
      </div>

      <boardgame-projected-choices
        .choices=${projectedChoices}
        .messageResolver=${this.messageResolver}
        @projected-choice-tray-resize=${this.projectedChoiceTrayResized}>
      </boardgame-projected-choices>

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

  private readonly _projectedChoicesChanged = (): void => {
    this.requestUpdate();
  };

  private readonly projectedChoiceTrayResized = (event: CustomEvent<{ height: number }>): void => {
    const height = event.detail?.height;
    if (!Number.isFinite(height) || height < 0) return;
    const normalized = Math.ceil(height);
    if (normalized !== this.projectedChoiceTrayHeight) {
      this.projectedChoiceTrayHeight = normalized;
    }
  };
}

customElements.define('boardgame-render-game', BoardgameRenderGame);

export { BoardgameRenderGame };
