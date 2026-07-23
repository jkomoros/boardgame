import { html, css, TemplateResult, type CSSResultGroup } from 'lit';
import { property } from 'lit/decorators.js';
import { BoardgameBaseGameRenderer } from './boardgame-base-game-renderer.js';
import type { FullGameState } from '../types/boardgame-types.js';
import type { MoveChoiceProjectionTypes } from '../moves/projected-choices.js';
import type { SeatPresentation } from './boardgame-table-view-base.js';
import { glyphForSlug } from './companion-avatar-catalog.js';
import type { EffectTransitionContext } from '../effects/effect-spec.js';
import type { MotionTransferDeclaration } from '../motion/transfer.js';

/**
 * BoardgameHandViewBase is the base class for the Hand view renderer that
 * a game author ships as boardgame-render-game-<X>-hand.ts (spec §7.3).
 *
 * The Hand view connects as PlayerIndex(n) — the player whose phone this
 * is — so the existing per-player sanitization gives it that player's
 * private state (their hand of cards, secret role if `sanitize:"other:
 * hidden"`, etc.).
 *
 * It extends BoardgameBaseGameRenderer for state/legality/move-proposal/
 * animation hooks, then adds a `playerState` convenience getter that
 * returns this.state.Players[this.viewingAs]. That's intentionally
 * sparse: the Hand view has no avatar strip or host controls (those are
 * Table-view-only). It also supplies the top-edge anchor and the historical
 * incoming-card compatibility default for cross-screen presentation.
 */
export class BoardgameHandViewBase<
  S extends FullGameState<object, object, object, object, object>,
  C extends object,
  MN extends string,
  MA extends Record<string, object>,
  K extends object = object,
  E extends object = object,
  MCP extends MoveChoiceProjectionTypes = Record<never, never>,
> extends BoardgameBaseGameRenderer<S, C, MN, MA, K, E, MCP> {

  /**
   * The player index this Hand view is bound to. Equals viewingAsPlayer
   * for symmetric games; for asymmetric games it's whichever seat the
   * phone claimed during /api/join/seat. Mirrored as `viewingAs` for the
   * convenience getter `playerState` below.
   */
  get viewingAs(): number {
    return this.viewingAsPlayer;
  }

  /**
   * Per-seat avatar + name records for everyone in the game (same data
   * the Table view receives). The Hand view typically only needs this
   * for a small banner ("Playing as BrightFox") or to show who the
   * current player is.
   */
  @property({ attribute: false })
  seatPresentations: readonly SeatPresentation[] = Object.freeze([]);

  /**
   * Convenience shortcut to this player's own substate. Common rendering
   * pattern: `${this.playerState.Hand.Components.map(...)}`. Returns
   * undefined if state is null OR viewingAsPlayer is out of bounds (e.g.
   * the player hasn't been seated yet).
   */
  protected get playerState(): S['Players'][number] | undefined {
    if (!this.state || !this.state.Players) return undefined;
    const idx = this.viewingAs;
    if (idx < 0 || idx >= this.state.Players.length) return undefined;
    return this.state.Players[idx];
  }

  /**
   * When true (the default), the base preserves the existing automatic Hand
   * choreography: new own-card IDs launch simultaneous animateBetween flights
   * from the top edge in their already-final visual pose. Moving among the
   * player's stacks does not retrigger. New authored choreography belongs in
   * motionTransfersForTransition(); set this false when opting into it.
   */
  @property({ type: Boolean })
  autoFlyIncoming = true;

  // Compatibility baseline. A first render or identity change records the
  // visible hand without replaying cards that were already present.
  private _prevOwnCardIds: Set<string> | null = null;

  // Buzz the phone when it becomes this player's turn — the player's eyes
  // are usually on the projector, so a local haptic is the natural cue.
  // navigator.vibrate is a no-op-safe progressive enhancement (undefined on
  // iOS Safari and desktop; short-circuits silently).
  private _wasMyTurn = false;

  protected override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated?.(changedProperties);
    if (changedProperties.has('viewingAsPlayer')) this._prevOwnCardIds = null;
    const myTurn = this.isCurrentPlayer && !this.gameFinished;
    if (myTurn && !this._wasMyTurn) {
      // Browsers block vibration before the first user gesture (and log a
      // console error) — only buzz once the user has interacted.
      if (navigator.userActivation?.hasBeenActive) {
        navigator.vibrate?.(200);
      }
    }
    this._wasMyTurn = myTurn;
    if (!changedProperties.has('state')) return;
    const ids = this._collectCardIds(this.playerState);
    const previous = this._prevOwnCardIds;
    this._prevOwnCardIds = ids;
    if (!this.autoFlyIncoming || previous === null) return;
    const incoming = [...ids].filter(id => !previous.has(id));
    if (incoming.length === 0) return;
    const anchor = this.shadowRoot?.getElementById('hand-top-edge') ?? 'hand-top-edge';
    requestAnimationFrame(() => requestAnimationFrame(() => {
      for (const id of incoming) void this.animator?.animateBetween(id, anchor, 600);
    }));
  }

  override motionTransfersForTransition(
    context: EffectTransitionContext<S, MN>,
  ): readonly MotionTransferDeclaration[] {
    return super.motionTransfersForTransition(context);
  }

  private _collectCardIds(player: S['Players'][number] | undefined): Set<string> {
    const out = new Set<string>();
    const ps = player as Record<string, unknown> | undefined;
    if (!ps) return out;
    for (const value of Object.values(ps)) {
      const ids = (value as { IDs?: unknown })?.IDs;
      if (!Array.isArray(ids)) continue;
      for (const id of ids) {
        if (typeof id === 'string' && id) out.add(id);
      }
    }
    return out;
  }

  /**
   * The game-specific hand area. Override this (instead of render()) to
   * get the standard Hand chrome for free: the default render() composes
   * top-edge anchor → hand header (identity + turn status) → YOUR HAND.
   * Games that want a different arrangement override render() itself and
   * call the helpers à la carte.
   */
  protected renderHand(): TemplateResult {
    return html``;
  }

  override render() {
    return html`
      ${this.renderTopEdgeAnchor()}
      ${this.renderHandHeader()}
      ${this.renderHand()}
    `;
  }

  /**
   * Opt-in helper: renders a one-line header for the Hand view — who this
   * phone is playing as (avatar + the display name picked in the join
   * flow) and whose turn it is ("Your turn" / "Waiting for <name>…").
   * These are the two questions every player asks their phone between
   * moves; call this at the top of render() to answer both for free.
   */
  protected renderHandHeader(): TemplateResult {
    const me = this.seatPresentations.find((s) => s.playerIndex === this.viewingAs);
    let statusText: string;
    let statusClass = '';
    if (this.gameFinished && this.animating) {
      // gameFinished can arrive while the final animation cycle is still
      // playing (the winning move's card is still in flight) — hold off on
      // the verdict text until it lands (#798). The watchdog force-closes
      // the gate within 4s, so this can never wedge permanently.
      statusText = 'Game over…';
    } else if (this.gameFinished) {
      // The verdict replaces the turn status once the game ends.
      if (this.gameWinners.includes(this.viewingAs)) {
        statusText = '🎉 You won!';
        statusClass = 'my-turn';
      } else if (this.gameWinners.length > 0) {
        statusText = 'Game over — you lost.';
      } else {
        statusText = "Game over — it's a draw.";
      }
    } else if (this.isCurrentPlayer) {
      statusText = 'Your turn';
      statusClass = 'my-turn';
    } else {
      const current = this.seatPresentations.find((s) => s.playerIndex === this.currentPlayerIndex);
      statusText = current ? `Waiting for ${current.displayName}…` : 'Waiting…';
    }
    return html`
      <div class="hand-header">
        ${me ? html`
          <span class="hand-identity">
            <span class="hand-avatar">${glyphForSlug(me.avatarSlug)}</span>
            ${me.displayName}
          </span>
        ` : ''}
        <span class="hand-turn ${statusClass}">${statusText}</span>
      </div>
    `;
  }

  /**
   * Opt-in helper: renders a small invisible anchor element at the top
   * edge of the Hand view, representing "from/to the Table". Cards dealt
   * to this player should be animated from this anchor; cards played
   * should exit through it. The element has a stable id ("hand-top-edge")
   * so both compatibility flights and transfer declarations can name it as
   * arrival geometry. Games with more precise deal semantics disable the
   * compatibility default and override motionTransfersForTransition().
   */
  protected renderTopEdgeAnchor(): TemplateResult {
    return html`<div class="hand-top-edge-anchor" id="hand-top-edge"></div>`;
  }

  static styles: CSSResultGroup = css`
    .hand-top-edge-anchor {
      position: fixed;
      top: 0;
      left: 50%;
      width: 1px;
      height: 1px;
      opacity: 0;
      pointer-events: none;
    }
    .hand-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 8px;
      font-size: 14px;
      margin-bottom: 12px;
    }
    .hand-identity {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      font-weight: 600;
    }
    .hand-avatar {
      font-size: 20px;
    }
    .hand-turn {
      opacity: 0.75;
    }
    .hand-turn.my-turn {
      opacity: 1;
      font-weight: 700;
    }
  `;
}
