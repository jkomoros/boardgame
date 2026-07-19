import { html, css, TemplateResult, type CSSResultGroup } from 'lit';
import { property } from 'lit/decorators.js';
import { BoardgameBaseGameRenderer } from './boardgame-base-game-renderer.js';
import type { FullGameState } from '../types/boardgame-types.js';
import type { MoveChoiceProjectionTypes } from '../moves/projected-choices.js';
import type { SeatPresentation } from './boardgame-table-view-base.js';
import { glyphForSlug } from './companion-avatar-catalog.js';

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
 * Table-view-only). Phase 4's top-edge off-screen anchor for cross-screen
 * animations is also wired here; V1 ships the prop surface only.
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
   * When true (the default), the base watches this player's own state for
   * newly-arrived card ids and flies them in from the top-edge anchor
   * automatically — the phone half of the cross-screen deal animation,
   * with zero author wiring. "Newly arrived" means an id that appears in
   * any Stack-shaped property of playerState and was not present in ANY
   * of them on the previous state (so cards shuffling between the
   * player's own stacks don't retrigger). Games whose incoming-card
   * semantics don't fit (or that wire bespoke animations) set this false
   * and call this.animator.animateBetween themselves.
   */
  @property({ type: Boolean })
  autoFlyIncoming = true;

  // null = no baseline yet (first render / reload mid-game): we record
  // what's already there without animating it.
  private _prevOwnCardIds: Set<string> | null = null;

  // Buzz the phone when it becomes this player's turn — the player's eyes
  // are usually on the projector, so a local haptic is the natural cue.
  // navigator.vibrate is a no-op-safe progressive enhancement (undefined on
  // iOS Safari and desktop; short-circuits silently).
  private _wasMyTurn = false;

  protected override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated?.(changedProperties);
    if (changedProperties.has('viewingAsPlayer')) {
      // The first state can install while we're still resolving as an
      // observer — playerState is undefined then, and a baseline recorded
      // at that moment is empty-but-non-null. When the seat identity
      // resolves a beat later, every long-held card would diff as
      // "incoming" and the whole hand would replay from the top edge.
      // Identity changed ⇒ start the baseline over.
      this._prevOwnCardIds = null;
    }
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
    // Keep the baseline current even when auto-fly is off, so toggling
    // the flag back on doesn't diff against a stale snapshot and fly in
    // every card at once.
    const ids = this._collectOwnCardIds();
    const prev = this._prevOwnCardIds;
    this._prevOwnCardIds = ids;
    if (!this.autoFlyIncoming) return;
    if (prev === null) return;
    const incoming = [...ids].filter((id) => !prev.has(id));
    if (incoming.length === 0) return;
    const anchor = this.shadowRoot?.getElementById('hand-top-edge') ?? 'hand-top-edge';
    // The card elements are rendered by child <boardgame-component-stack>
    // elements that re-render asynchronously after receiving the new
    // state; wait two frames so the new cards exist before we measure.
    requestAnimationFrame(() => requestAnimationFrame(() => {
      for (const id of incoming) {
        this.animator?.animateBetween(id, anchor, 600);
      }
    }));
  }

  private _collectOwnCardIds(): Set<string> {
    const out = new Set<string>();
    const ps = this.playerState as Record<string, unknown> | undefined;
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
   * so authors can call this.animator.animateBetween(realCardId,
   * "hand-top-edge", durationMs) to wire deal/play animations.
   *
   * V1 ships the anchor element only — game authors wire the actual
   * animation calls from their own renderer's state-change reactions.
   * The base doesn't auto-detect deals because deal-ness is game-
   * specific (which moves count as "incoming card from Table"?).
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
