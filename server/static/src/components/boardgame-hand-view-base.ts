import { html, css, TemplateResult } from 'lit';
import { property } from 'lit/decorators.js';
import { BoardgameBaseGameRenderer } from './boardgame-base-game-renderer.js';
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
  GS extends Record<string, unknown> = Record<string, unknown>,
  PS extends Record<string, unknown> = Record<string, unknown>,
  MN extends string = string,
  MA extends Record<string, Record<string, unknown>> = Record<string, Record<string, unknown>>
> extends BoardgameBaseGameRenderer<GS, PS, MN, MA> {

  /**
   * The player index this Hand view is bound to. Equals viewingAsPlayer
   * for symmetric games; for asymmetric games it's whichever seat the
   * phone claimed during /api/join/seat. Mirrored as `viewingAs` for the
   * convenience getter `playerState` below.
   */
  get viewingAs(): number {
    return this.viewingAsPlayer;
  }

  @property({ type: String })
  gameName = '';

  @property({ type: String })
  gameId = '';

  /**
   * Per-seat avatar + name records for everyone in the game (same data
   * the Table view receives). The Hand view typically only needs this
   * for a small banner ("Playing as BrightFox") or to show who the
   * current player is.
   */
  @property({ type: Array })
  seatPresentations: SeatPresentation[] = [];

  /**
   * Server-clock instant (ms since epoch) at which this state's cross-
   * screen animation should begin playing. Set by boardgame-game-state-
   * manager from the "version-timing" WebSocket message (spec §8.4).
   */
  @property({ type: Number })
  serverPlayAt: number | null = null;

  /**
   * Convenience shortcut to this player's own substate. Common rendering
   * pattern: `${this.playerState.Hand.Components.map(...)}`. Returns
   * undefined if state is null OR viewingAsPlayer is out of bounds (e.g.
   * the player hasn't been seated yet).
   */
  protected get playerState(): PS | undefined {
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

  protected override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated?.(changedProperties);
    if (!this.autoFlyIncoming) return;
    if (!changedProperties.has('state')) return;
    const ids = this._collectOwnCardIds();
    const prev = this._prevOwnCardIds;
    this._prevOwnCardIds = ids;
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
   * Opt-in helper: renders a one-line header for the Hand view — who this
   * phone is playing as (avatar + the display name picked in the join
   * flow) and whose turn it is ("Your turn" / "Waiting for <name>…").
   * These are the two questions every player asks their phone between
   * moves; call this at the top of render() to answer both for free.
   */
  protected renderHandHeader(): TemplateResult {
    const me = this.seatPresentations.find((s) => s.playerIndex === this.viewingAs);
    const isMyTurn = this.isCurrentPlayer;
    const current = this.seatPresentations.find((s) => s.playerIndex === this.currentPlayerIndex);
    const turnText = isMyTurn
      ? 'Your turn'
      : current
        ? `Waiting for ${current.displayName}…`
        : 'Waiting…';
    return html`
      <div class="hand-header">
        ${me ? html`
          <span class="hand-identity">
            <span class="hand-avatar">${glyphForSlug(me.avatarSlug)}</span>
            ${me.displayName}
          </span>
        ` : ''}
        <span class="hand-turn ${isMyTurn ? 'my-turn' : ''}">${turnText}</span>
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

  static styles = css`
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
