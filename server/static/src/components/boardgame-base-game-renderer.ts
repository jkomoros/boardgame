import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import type { MoveLegalityInfo } from '../selectors.js';
import type { MovePreviewSpec } from '../legal/previewLegality.js';
import type { FullGameState } from '../types/boardgame-types.js';
import { START_MOVE_NAMES, getReadyToStartError } from './gathering-shared.js';
import type { ComponentAnimatorAPI } from './boardgame-component-animator.js';
import {
  creatorMoveInputFromLegacyStrings,
  serializeCreatorMoveInputForServer,
  type MoveInputSchema,
} from '../moves/input.js';

export class BoardgameBaseGameRenderer<
  GS extends object = Record<string, unknown>,
  PS extends object = Record<string, unknown>,
  MN extends string = string,
  MA extends Record<string, object> = Record<string, Record<string, unknown>>
> extends LitElement {
  /** Generated safe-input contract installed by a bound/game renderer. */
  protected readonly moveInputSchema: MoveInputSchema | null = null;
  protected readonly moveInputSchemaFingerprint: string | null = null;

  /** Canonical fingerprint supplied by the current server /info response. */
  @property({ type: String, attribute: false })
  serverMoveInputSchemaFingerprint: string | null = null;
  @property({ type: Object })
  state: FullGameState<GS, PS> | null = null;

  @property({ type: Object })
  chest: Record<string, unknown> | null = null;

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
   * `BoardgameBaseGameRenderer<GameState, PlayerState, MoveName, MoveArgs>`,
   * this method provides compile-time checking that the move name is valid
   * and the arguments match the expected fields.
   *
   * Usage in a game renderer:
   * ```
   * import { MoveNames, type MoveName } from './_move_names.js';
   * import type { MoveInputs } from './_move_args.js';
   *
   * class MyRenderer extends BoardgameBaseGameRenderer<GS, PS, MoveName, MoveInputs> {
   *   handleClick() {
   *     this.proposeMove(MoveNames.RevealCard, { CardIndex: 3 });
   *   }
   * }
   * ```
   */
  proposeMove<K extends MN & string>(
    moveName: K,
    ...args: K extends keyof MA
      ? {} extends MA[K]
        ? [args?: MA[K]]
        : [args: MA[K]]
      : [args: Record<string, unknown>]
  ): void {
    this._proposeMoveNative(moveName, args[0] ?? {});
  }

  private _proposeMoveNative(moveName: string, nativeArgs: unknown): void {
    // Convert all values to strings for the server (form-encoded submission).
    // Booleans must be "1"/"0" (not "true"/"false") because the server uses
    // strconv.Atoi for boolean fields.
    let stringArgs: Readonly<Record<string, string>>;
    if (this.moveInputSchema || this.moveInputSchemaFingerprint) {
      if (!this.moveInputSchema || !this.moveInputSchemaFingerprint) {
        throw new Error('Incomplete generated move-input contract on renderer');
      }
      stringArgs = serializeCreatorMoveInputForServer(
        this.moveInputSchema,
        this.moveInputSchemaFingerprint,
        this.serverMoveInputSchemaFingerprint,
        moveName,
        nativeArgs,
      );
    } else {
      // Explicit legacy compatibility path for renderers not yet bound to a
      // generated contract.
      stringArgs = Object.fromEntries(Object.entries(nativeArgs as Record<string, unknown>).map(([key, value]) => [
        key,
        typeof value === 'boolean' ? (value ? '1' : '0') : String(value),
      ]));
    }
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

  private _boundHandleButtonTapped?: (e: Event) => void;

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this._boundHandleButtonTapped = (e: Event) => this._handleButtonTapped(e);

    // CHANGED: tap → click (Polymer event → standard event)
    this.addEventListener('click', this._boundHandleButtonTapped);
    this.addEventListener('component-tapped', this._boundHandleButtonTapped);
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    if (this._boundHandleButtonTapped) {
      this.removeEventListener('click', this._boundHandleButtonTapped);
      this.removeEventListener('component-tapped', this._boundHandleButtonTapped);
    }
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

  private _handleButtonTapped(e: Event): void {
    const composedPath = e.composedPath();
    let ele: HTMLElement | null = null;

    for (const tempEle of composedPath) {
      // Runtime type check (no unsafe casts)
      if (!(tempEle instanceof Element)) continue;
      if (!tempEle.hasAttribute) continue;

      // Only accept string-valued proposeMove properties: the renderer
      // element itself (and anything extending BoardgameBaseGameRenderer)
      // has a proposeMove METHOD, which must not be mistaken for the
      // legacy string-property/attribute convention.
      const rawProposeMove = (tempEle as Element & { proposeMove?: unknown }).proposeMove;
      const proposeMove = (typeof rawProposeMove === 'string' ? rawProposeMove : null) || tempEle.getAttribute('propose-move');
      if (proposeMove) {
        // found it!
        ele = tempEle as HTMLElement;
        break;
      }
    }

    if (!ele) {
      return;
    }

    if (ele.hasAttribute('boardgame-component') && e.type === 'click') {
      // Cards we'll fire on the component-tapped, not the click.
      return;
    }

    const rawMoveName = (ele as HTMLElement & { proposeMove?: unknown }).proposeMove;
    const moveName = (typeof rawMoveName === 'string' ? rawMoveName : null) || ele.getAttribute('propose-move');
    if (!moveName) return;

    const data = ele.dataset;
    const args: Record<string, string> = {};

    for (const key in data) {
      if (!Object.prototype.hasOwnProperty.call(data, key)) continue;
      if (!key.startsWith('arg')) continue;
      let effectiveKey = key.replace('arg', '');
      // Handle the case where the attribute was literally just data-arg
      if (!effectiveKey) continue;
      // The first character is now upperCase, which is desired as per Move field convention
      const value = data[key];
      if (value !== undefined) args[effectiveKey] = value;
    }

    const nativeArgs = this.moveInputSchema
      ? creatorMoveInputFromLegacyStrings(this.moveInputSchema, moveName, args)
      : args;
    this._proposeMoveNative(moveName, nativeArgs);
  }

  override render() {
    return html``;
  }
}

customElements.define('boardgame-base-game-renderer', BoardgameBaseGameRenderer);
