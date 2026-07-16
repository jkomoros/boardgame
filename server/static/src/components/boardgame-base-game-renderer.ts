import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import type { MoveLegalityInfo } from '../selectors.js';
import type { MovePreviewSpec } from '../legal/previewLegality.js';
import type { FullGameState, GameChest } from '../types/boardgame-types.js';
import { START_MOVE_NAMES, getReadyToStartError } from './gathering-shared.js';
import type { ComponentAnimatorAPI } from './boardgame-component-animator.js';
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
import {
  cancelTargetActionPreview,
  notifyTargetActionLiveStateChanged,
  type TargetAction,
  type TargetKey,
  type TargetPreviewTransport,
} from '../moves/target-action.js';
import { LegacyProposalAdapter } from '../moves/legacy-proposal-adapter.js';

type MoveInputFor<
  K extends string,
  Inputs extends Record<string, object>,
> = K extends keyof Inputs ? Inputs[K] : Record<string, unknown>;

type ExactMoveInput<Expected extends object, Actual extends Expected> = Actual &
  Record<Exclude<keyof Actual, keyof Expected>, never>;

export class BoardgameBaseGameRenderer<
  S extends FullGameState<object, object, object, object, object>,
  C extends object,
  MN extends string,
  MA extends Record<string, object>,
> extends LitElement {
  /** Generated safe-input contract installed by a bound/game renderer. */
  protected readonly moveInputSchema: MoveInputSchema | null = null;
  protected readonly moveInputSchemaFingerprint: string | null = null;

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
  readonly #legacyProposalAdapter = new LegacyProposalAdapter(
    this,
    () => this.moveInputSchema,
    (moveName, nativeArguments) => this._proposeMoveNative(moveName, nativeArguments),
  );
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
  chest: GameChest<C> | null = null;

  @property({ type: String })
  diagram = '';

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Number })
  currentPlayerIndex = 0;

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
    if (this.viewingAsPlayer === -2) return true;
    // AnyPlayerIndex (-3): any player can act (simultaneous phase)
    if (this.currentPlayerIndex === -3) return true;
    return this.currentPlayerIndex === this.viewingAsPlayer;
  }

  /**
   * The FLIP animator that wraps this renderer. Renderers live inside
   * boardgame-render-game's #container, a sibling of the #animator element
   * in the same shadow root. Use for one-off cross-screen animations:
   * `this.animator?.animateBetween(cardId, 'hand-top-edge')`. Null before
   * the renderer is attached (or in tests outside boardgame-render-game).
   */
  protected get animator(): ComponentAnimatorAPI | null {
    const root = this.getRootNode();
    if (!(root instanceof ShadowRoot)) return null;
    return root.querySelector('#animator') as any;
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
   * Type-safe move proposal. When your game renderer extends
   * `BoardgameBaseGameRenderer<State, ComponentCatalog, MoveName, MoveArgs>`,
   * this method provides compile-time checking that the move name is valid
   * and the arguments match the expected fields.
   *
   * Usage in a game renderer:
   * ```
   * import { MoveNames, type MoveName } from './_move_names.js';
   * import type { MoveInputs } from './_move_args.js';
   *
   * class MyRenderer extends BoardgameBaseGameRenderer<State, ComponentCatalog, MoveName, MoveInputs> {
   *   handleClick() {
   *     this.proposeMove(MoveNames.RevealCard, { CardIndex: 3 });
   *   }
   * }
   * ```
   */
  proposeMove<
    K extends MN & string,
    Actual extends MoveInputFor<K, MA> = MoveInputFor<K, MA>,
  >(
    moveName: K,
    ...args: {} extends MoveInputFor<K, MA>
      ? [args?: ExactMoveInput<MoveInputFor<K, MA>, Actual>]
      : [args: ExactMoveInput<MoveInputFor<K, MA>, Actual>]
  ): void {
    this._proposeMoveNative(moveName, args[0] ?? {});
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

  private _proposeMoveNative(moveName: string, nativeArgs: unknown): void {
    // Convert all values to strings for the server (form-encoded submission).
    // Booleans must be "1"/"0" (not "true"/"false") because the server uses
    // strconv.Atoi for boolean fields.
    if (!this.moveInputSchema || !this.moveInputSchemaFingerprint) {
      throw new Error(
        'Renderer has no generated move-input contract; extend the generated GameRenderer base',
      );
    }
    const stringArgs = serializeCreatorMoveInputForServer(
      this.moveInputSchema,
      this.moveInputSchemaFingerprint,
      this.serverMoveInputSchemaFingerprint,
      moveName,
      nativeArgs,
    );
    this.dispatchEvent(new CustomEvent('propose-move', {
      composed: true,
      bubbles: true,
      detail: { name: moveName, arguments: stringArgs }
    }));
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

  override connectedCallback() {
    super.connectedCallback();
    this.#legacyProposalAdapter.connect();
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    for (const action of this.#moveActionCache.values()) cancelMoveActionPreview(action);
    for (const action of this.#targetActionCache.values()) cancelTargetActionPreview(action);
    this.#moveActionCache.clear();
    this.#targetActionCache.clear();
    this.#lastMoveSnapshotKey = '';
    this.#legacyProposalAdapter.disconnect();
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
  // The default one returns 0 for all combinations. See also animationOverlap.
  animationLength(fromMove: Record<string, unknown> | null, toMove: Record<string, unknown> | null): number {
    return 0;
  }

  // animationOverlap returns a fraction (0-1) of the animation length after
  // which the next state can be installed, even if the current animation is
  // still running. 0 (default) = wait for animation to complete (no overlap).
  // 0.5 = start next animation when this one is halfway done. Values outside
  // 0-1 are clamped. See also animationLength.
  animationOverlap(fromMove: Record<string, unknown> | null, toMove: Record<string, unknown> | null): number {
    return 0;
  }

  override render() {
    return html``;
  }
}

customElements.define('boardgame-base-game-renderer', BoardgameBaseGameRenderer);
