import { property } from 'lit/decorators.js';
import { BoardgameBaseGameRenderer } from './boardgame-base-game-renderer.js';
import type { SeatPresentation } from './boardgame-table-view-base.js';

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
}
