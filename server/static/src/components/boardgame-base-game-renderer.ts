import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import type { MoveLegalityInfo } from '../selectors.js';
import type { MovePreviewSpec } from '../legal/previewLegality.js';
import type { FullGameState, GameChest } from '../types/boardgame-types.js';
import type { ClientMove } from '../types/api.js';
import { START_MOVE_NAMES, getReadyToStartError } from './gathering-shared.js';
import type { ComponentAnimatorAPI } from './boardgame-component-animator.js';
import { DEFAULT_EFFECT_THEME } from '../effects/effect-spec.js';
import type {
  EffectHostAPI,
  EffectSpec,
  EffectTheme,
  EffectTransitionContext,
} from '../effects/effect-spec.js';
import type { MotionStaggerCohortSpec } from '../motion/cohort.js';
import type { MotionTransferDeclaration } from '../motion/transfer.js';
import type { MotionReleaseDeclaration } from '../motion/release.js';
import {
  serializeCreatorMoveInputForServer,
  validateCreatorMoveInput,
  type MoveInputSchema,
} from '../moves/input.js';
import {
  MoveSubmissionGate,
  cancelMoveActionPreview,
  createMoveAction,
  moveSnapshotKey,
  notifyMoveActionLiveStateChanged,
  type MoveActionFor,
  type MoveActionService,
  type MovePreviewTransport,
  type MoveTransport,
} from '../moves/action.js';
import type { ProjectedMoveChoicesWire } from '../types/api.js';
import {
  buildProjectedMoveChoices,
  ProjectedMoveChoices,
  type MessageDescriptor,
  type MoveChoiceProjectionSchemaEntry,
  type MoveChoiceProjectionTypes,
} from '../moves/projected-choices.js';
import {
  cancelTargetActionPreview,
  notifyTargetActionLiveStateChanged,
  type TargetAction,
  type TargetKey,
  type TargetPreviewTransport,
} from '../moves/target-action.js';
import {
  AdminPlayerIndex,
  AnyPlayerIndex,
  type TurnStatusContext,
} from '../status/turn-status.js';
import {
  fallbackPlayerPresentation,
  type PlayerPresentation,
} from '../status/player-presentation.js';

type MoveInputFor<
  K extends string,
  Inputs extends Record<string, object>,
> = K extends keyof Inputs ? Inputs[K] : Record<string, unknown>;

export class BoardgameBaseGameRenderer<
  S extends FullGameState<object, object, object, object, object>,
  C extends object,
  MN extends string,
  MA extends Record<string, object>,
  K extends object = object,
  E extends object = object,
  MCP extends MoveChoiceProjectionTypes = Record<never, never>,
> extends LitElement {
  /** Generated safe-input contract installed by a bound/game renderer. */
  protected readonly moveInputSchema: MoveInputSchema | null = null;
  protected readonly moveInputSchemaFingerprint: string | null = null;
  /** Generated finite projection contract installed by a game renderer base. */
  protected readonly moveChoiceProjectionSchema: readonly MoveChoiceProjectionSchemaEntry[] = [];
  protected readonly moveChoiceProjectionSchemaFingerprint: string = '';
  /** Client-owned localizable prompts; no presentation copy crosses the wire. */
  protected readonly moveChoiceMessages: Readonly<Partial<Record<keyof MCP & string, MessageDescriptor>>> =
    {} as Readonly<Partial<Record<keyof MCP & string, MessageDescriptor>>>;

  @property({ type: Object, attribute: false })
  projectedMoveChoicesWire: ProjectedMoveChoicesWire | null = null;

  private _projectedMoveChoices: ProjectedMoveChoices<MCP> | null = null;

  /** Exact typed candidate sets for this snapshot. */
  get choices(): ProjectedMoveChoices<MCP> | null {
    return this._projectedMoveChoices;
  }

  /** Canonical fingerprint supplied by the current server /info response. */
  @property({ type: String, attribute: false })
  serverMoveInputSchemaFingerprint: string | null = null;
  @property({ type: Object })
  state: S | null = null;

  @property({ type: String, attribute: false })
  gameName = '';

  @property({ type: String, attribute: false })
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
  moveSubmissionGate: MoveSubmissionGate = new MoveSubmissionGate();

  readonly #moveActionCache = new Map<string, import('../moves/action.js').BoundMoveAction<string, object>>();
  readonly #targetActionCache = new Map<string, TargetAction<TargetKey, string, object>>();
  readonly #moveActionService: MoveActionService = {
    currentClientSchemaFingerprint: () => this.moveInputSchemaFingerprint ?? '',
    currentServerSchemaFingerprint: () => this.serverMoveInputSchemaFingerprint,
    currentTransport: () => this.moveTransport,
    currentPreviewTransport: () => this.movePreviewTransport,
    currentTargetPreviewTransport: () => this.targetPreviewTransport,
    currentGate: () => this.moveSubmissionGate,
    nextRequestID: () => `${this.gameId || 'game'}-v${this.gameVersion}-move-${++this.#moveRequestSequence}`,
    validate: (moveName, input) => this.moveInputSchema
      ? validateCreatorMoveInput(this.moveInputSchema, moveName, input).errors
      : [],
    serialize: (moveName, input) => {
      if (!this.moveInputSchema || !this.moveInputSchemaFingerprint) {
        throw new Error('Renderer has no generated move-input contract; extend the generated GameRenderer base');
      }
      return serializeCreatorMoveInputForServer(
        this.moveInputSchema,
        this.moveInputSchemaFingerprint,
        this.serverMoveInputSchemaFingerprint,
        moveName,
        input,
      );
    },
    changed: () => this.requestUpdate(),
    telemetry: event => this.dispatchEvent(new CustomEvent('move-action-result', {
      bubbles: true,
      composed: true,
      detail: event,
    })),
    actionCache: this.#moveActionCache,
    targetActionCache: this.#targetActionCache,
  };
  #lastMoveSnapshotKey = '';
  #moveRequestSequence = 0;

  @property({ type: Object })
  chest: GameChest<C, K, E> | null = null;

  @property({ type: String })
  diagram = '';

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Number })
  currentPlayerIndex = 0;

  /** Public, sanitized player labels/colors supplied by the renderer host. */
  @property({ attribute: false })
  playerPresentations: readonly PlayerPresentation[] = Object.freeze([]);

  /** Stable creator-facing presentation with a useful fallback for fixtures. */
  playerPresentation(playerIndex: number): PlayerPresentation {
    if (!Number.isSafeInteger(playerIndex) || playerIndex < 0) {
      return fallbackPlayerPresentation(playerIndex);
    }
    return this.playerPresentations.find(player => player.playerIndex === playerIndex)
      ?? fallbackPlayerPresentation(playerIndex);
  }

  /**
   * Map of move name → legality info, set by boardgame-render-game from the
   * Redux store. Renderers should use the convenience helpers
   * isMoveCurrentlyLegal() and isMovePossible() instead of reading this
   * directly.
   */
  @property({ type: Object })
  moveLegality: Record<string, MoveLegalityInfo> = {};

  /**
   * Mirrors the game record's Finished/Winners (plumbed through
   * boardgame-render-game). Winners are player indexes; empty array with
   * gameFinished=true means a draw / no-winner ending.
   */
  @property({ type: Boolean })
  gameFinished = false;

  @property({ type: Array })
  gameWinners: number[] = [];

  /**
   * Mirrors boardgame-render-game's isAnimating (set at both gate flips and
   * at instantiation — see _applyAnimatingToRenderer). Table and Hand view
   * subclasses gate their verdict UI (renderGameOverBanner /
   * renderHandHeader) on this so the outcome never appears while the last
   * animation cycle (e.g. the winning card landing) is still in flight —
   * it should only announce itself once the board has visually settled
   * (#798).
   */
  @property({ type: Boolean })
  animating = false;

  get isCurrentPlayer(): boolean {
    // AdminPlayerIndex (-2): admin can always act
    if (this.viewingAsPlayer === AdminPlayerIndex) return true;
    // AnyPlayerIndex (-3): any player can act (simultaneous phase)
    if (this.currentPlayerIndex === AnyPlayerIndex) return true;
    return this.currentPlayerIndex === this.viewingAsPlayer;
  }

  /** Complete, sentinel-aware input for the standard turn-status primitive. */
  get turnStatus(): TurnStatusContext {
    return {
      currentPlayerIndex: this.currentPlayerIndex,
      viewerPlayerIndex: this.viewingAsPlayer,
      finished: this.gameFinished,
      animating: this.animating,
    };
  }

  /**
   * The FLIP animator that wraps this renderer. Renderers live inside
   * boardgame-render-game's #container, a sibling of the #animator element
   * in the same shadow root. Use for one-off cross-screen animations:
   * `this.animator?.fly({ subjectId: cardId, source: 'hand-top-edge',
   * carrier: cardId })`. Null before
   * the renderer is attached (or in tests outside boardgame-render-game).
   */
  protected get animator(): ComponentAnimatorAPI | null {
    const root = this.getRootNode();
    if (!(root instanceof ShadowRoot)) return null;
    return root.querySelector('#animator') as any;
  }

  /**
   * Plays immutable visual-effect descriptors. Effects are decorative and
   * never own game truth or hold the state queue.
   */
  protected get effects(): EffectHostAPI | null {
    const root = this.getRootNode();
    if (!(root instanceof ShadowRoot)) return null;
    return root.querySelector('#effects') as EffectHostAPI | null;
  }

  /** Semantic palette overrides for this game renderer's effects. */
  effectTheme(): EffectTheme {
    return DEFAULT_EFFECT_THEME;
  }

  /**
   * Pure, exactly-once planning hook for an installed authoritative snapshot.
   * Return immutable descriptors; never start effects from Lit lifecycle hooks.
   * Initial loads are explicit (`context.kind === 'initial'`) so reconnecting
   * to an already-finished game does not accidentally replay a celebration.
   */
  effectsForTransition(_context: EffectTransitionContext<S, MN>): readonly EffectSpec[] {
    return [];
  }

  /**
   * Pure structural start scheduling for one authoritative transition.
   * Subject array order is the cadence order. Effects remain observational
   * and cannot retime this motion.
   */
  motionCohortsForTransition(
    _context: EffectTransitionContext<S, MN>,
  ): readonly MotionStaggerCohortSpec[] {
    return [];
  }

  /**
   * Pure, transition-local retained-carrier presentation intent. `key` is
   * unique only within this installed transition; it is not cross-device
   * identity unless the server exposes an identical safe token to each view.
   */
  motionTransfersForTransition(
    _context: EffectTransitionContext<S, MN>,
  ): readonly MotionTransferDeclaration[] {
    return [];
  }

  /**
   * Pure policy for admitting an already-buffered successor before this
   * structural cycle settles. This is a destructive cutover: the next install
   * terminalizes this generation; it is not concurrent multi-generation motion.
   */
  motionReleaseForTransition(
    _context: EffectTransitionContext<S, MN>,
  ): MotionReleaseDeclaration | null {
    return null;
  }

  /**
   * Returns true if the named move is legal for the viewing player right now.
   * Use this to disable buttons when a move can't be made (e.g. not your turn).
   */
  isMoveCurrentlyLegal(moveName: MN): boolean {
    return this.moveLegality[moveName]?.legalForPlayer ?? false;
  }

  /**
   * previewDisabledSpaces is set by boardgame-render-game after a batch legality
   * preview (movePreviewBatch) returns: the board spaces whose candidate move is
   * currently illegal for the viewing player. Bind it to your
   * <boardgame-game-board .disabledSpaces="${this.previewDisabledSpaces}"> — the
   * board already grays and blocks clicks on disabled spaces. Empty by default
   * (no preview, or every candidate legal).
   */
  @property({ type: Array })
  previewDisabledSpaces: number[] = [];

  /**
   * previewSpec opts a renderer into per-target legality preview. Return the
   * move type to check plus one PreviewCandidate per board space you want tested
   * (its space index + the move args for targeting there); boardgame-render-game
   * batches them to the server (movePreviewBatch) and feeds the illegal ones
   * back via previewDisabledSpaces so the board can gray them BEFORE the player
   * clicks. Return null (the default) to leave preview off. It is called
   * whenever state / moveForms / the viewing player change, so keep it cheap and
   * pure (no side effects, just read this.state).
   */
  previewSpec(): MovePreviewSpec | null {
    return null;
  }

  /**
   * requestPreviewRefresh asks the host (boardgame-render-game) to re-run this
   * renderer's previewSpec() and refresh the board's legality graying. Call it
   * after LOCAL interaction state that previewSpec() reads changes WITHOUT a
   * server round-trip — e.g. a multi-step move where the player just selected a
   * source piece, so the legal destinations are now different. State / move-form
   * / viewing-player changes already trigger a refresh automatically; this hook
   * is only for renderer-local state the host can't observe.
   */
  protected requestPreviewRefresh(): void {
    this.dispatchEvent(new CustomEvent('preview-refresh-requested', { composed: true, bubbles: true }));
  }

  /**
   * Returns true if the named move is structurally possible right now (legal
   * for any player / admin). Use this to hide buttons entirely when a move
   * isn't applicable in the current game phase.
   */
  isMovePossible(moveName: MN): boolean {
    return this.moveLegality[moveName]?.legalForAnyone ?? false;
  }

  /**
   * Creates the canonical typed action for creator-authored controls. Zero-input
   * moves can propose immediately; required-input moves expose only with(args)
   * until their exact generated native input is bound.
   */
  move<K extends MN & string>(name: K): MoveActionFor<K, MoveInputFor<K, MA>> {
    const snapshotKey = this._moveSnapshotKey();
    return createMoveAction<K, MN, MA>(name, this.#moveActionService, {
      snapshotKey,
      currentSnapshotKey: () => this._moveSnapshotKey(),
      snapshotVersion: this.gameVersion,
      currentSnapshotVersion: () => this.gameVersion,
      viewingAsPlayer: this.viewingAsPlayer,
      proposingAsPlayer: this.proposingAsPlayer,
      proposingAsAdmin: this.proposingAsAdmin,
      currentLegality: () => this.moveLegality[name],
      currentAnimating: () => this.animating,
      baselineLegalityApplies: this.proposingAsPlayer === this.viewingAsPlayer,
      actionCacheKey: snapshotKey,
    });
  }

  /**
   * Returns true if any gathering-related moves (team/role/color selection,
   * or start-game moves) are currently legal. Game renderers can use
   * `if (this.gatheringActive) { ... }` to conditionally render
   * gathering-specific UI.
   */
  get gatheringActive(): boolean {
    for (const name of Object.keys(this.moveLegality)) {
      const info = this.moveLegality[name];
      if (!info.legalForAnyone) continue;
      if (START_MOVE_NAMES.has(name)) {
        return true;
      }
    }
    if (getReadyToStartError(this.state)) {
      return true;
    }
    return false;
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    for (const action of this.#moveActionCache.values()) cancelMoveActionPreview(action);
    for (const action of this.#targetActionCache.values()) cancelTargetActionPreview(action);
    this.#moveActionCache.clear();
    this.#targetActionCache.clear();
    this.#lastMoveSnapshotKey = '';
  }

  protected override willUpdate(changedProperties: Map<PropertyKey, unknown>): void {
    super.willUpdate(changedProperties);
    const snapshotKey = this._moveSnapshotKey();
    if (snapshotKey !== this.#lastMoveSnapshotKey) {
      for (const action of this.#moveActionCache.values()) cancelMoveActionPreview(action);
      for (const action of this.#targetActionCache.values()) cancelTargetActionPreview(action);
      this.#moveActionCache.clear();
      this.#targetActionCache.clear();
      this.#lastMoveSnapshotKey = snapshotKey;
    } else if (changedProperties.has('animating')) {
      // Animation state is deliberately not part of snapshot identity: a
      // button should keep the same action (and preview) while the visual gate
      // opens and closes. Notify subscribed controls because passing that same
      // cached object through Lit does not trigger their property update.
      for (const action of this.#moveActionCache.values()) {
        notifyMoveActionLiveStateChanged(action);
      }
      for (const action of this.#targetActionCache.values()) {
        notifyTargetActionLiveStateChanged(action);
      }
    }
    if (changedProperties.has('projectedMoveChoicesWire')
      || changedProperties.has('gameVersion')
      || changedProperties.has('playerPresentations')) {
      this._installProjectedMoveChoices();
    }
  }

  private _installProjectedMoveChoices(): void {
    const wire = this.projectedMoveChoicesWire;
    if (!wire) {
      this._projectedMoveChoices = null;
      this._notifyProjectedMoveChoicesChanged();
      return;
    }
    try {
      this._projectedMoveChoices = buildProjectedMoveChoices<MCP>({
        wire,
        stateVersion: this.gameVersion,
        schema: this.moveChoiceProjectionSchema,
        schemaFingerprint: this.moveChoiceProjectionSchemaFingerprint,
        playerPresentations: this.playerPresentations,
        action: (move, input) => {
          const builder = this.move(move as unknown as MN & string) as unknown as import('../moves/action.js').MoveActionBuilder<
            typeof move,
            MCP[typeof move]['input']
          >;
          return builder.with(input);
        },
        messages: this.moveChoiceMessages,
      });
    } catch (error) {
      console.error('[projected-choices] rejected invalid snapshot:', error);
      this._projectedMoveChoices = ProjectedMoveChoices.failed<MCP>();
    }
    this._notifyProjectedMoveChoicesChanged();
  }

  private _notifyProjectedMoveChoicesChanged(): void {
    this.dispatchEvent(new CustomEvent('projected-choices-changed', {
      bubbles: true,
      composed: true,
    }));
  }

  private _moveSnapshotKey(): string {
    return moveSnapshotKey({
      gameName: this.gameName,
      gameID: this.gameId,
      epoch: this.snapshotEpoch,
      version: this.gameVersion,
      viewingAsPlayer: this.viewingAsPlayer,
      proposingAsPlayer: this.proposingAsPlayer,
      proposingAsAdmin: this.proposingAsAdmin,
      serverSchemaFingerprint: this.serverMoveInputSchemaFingerprint,
    });
  }

  // animationLength is consulted when applying an animation to configure the
  // animation length (in milliseconds) by setting `--animation-length` on the
  // renderer. Zero will specify default animation length (that is, unset an
  // override style). A negative return value will skip the animation entirely.
  // The default one returns 0 for all combinations.
  animationLength(_fromMove: ClientMove | null, _toMove: ClientMove | null): number {
    return 0;
  }

  /**
   * Compatibility state-clock cutover. Unlike motionReleaseForTransition(),
   * this can inspect the already-buffered successor and retains its historical
   * fraction-of-animationLength semantics. An override takes precedence over
   * motionReleaseForTransition() for that renderer.
   */
  animationOverlap(_fromMove: ClientMove | null, _toMove: ClientMove | null): number {
    return 0;
  }

  override render() {
    return html``;
  }
}

customElements.define('boardgame-base-game-renderer', BoardgameBaseGameRenderer);
