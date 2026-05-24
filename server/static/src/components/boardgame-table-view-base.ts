import { html, TemplateResult } from 'lit';
import { property } from 'lit/decorators.js';
import { BoardgameBaseGameRenderer } from './boardgame-base-game-renderer.js';

/**
 * SeatPresentation mirrors the server's seatpresentation.StorageRecord
 * (server/api/seatpresentation/main.go). Each entry describes one filled
 * seat in a companion-mode game; the array is delivered to the Table view
 * so it can render the avatar strip across the top edge.
 */
export interface SeatPresentation {
  playerIndex: number;
  displayName: string;
  avatarSlug: string;
}

/**
 * BoardgameTableViewBase is the base class for the Table view renderer that
 * a game author ships as boardgame-render-game-<X>-table.ts (spec §7.2).
 *
 * The Table view connects as ObserverPlayerIndex (game.go's existing
 * sanitization machinery handles privacy) and renders the public board for
 * a shared screen — laptop hooked to a TV, tablet on the kitchen counter,
 * etc.
 *
 * It extends BoardgameBaseGameRenderer so it inherits the existing typed
 * state, move-legality, move-proposal, and animation hooks. The new
 * companion-mode-only properties are added here. None of them are
 * required for the base to compile a render() — a game author can ignore
 * any property they don't need.
 *
 * Helper render methods (renderAvatarStrip, renderHostControls,
 * renderFakeDeckRow) are intentionally opt-in: the author calls them from
 * their own render() at the spot in the layout they want them. The base
 * does NOT auto-inject anything into light DOM — that would conflict with
 * Lit's reactive contract.
 *
 * V1 ships the property surface; the helper-render IMPLEMENTATIONS are
 * stubs returning empty templates. Phase 3 fills in renderAvatarStrip /
 * renderHostControls once seatPresentation + presence wiring lands on the
 * client. Phase 4 fills in renderFakeDeckRow for cross-screen animations.
 */
export class BoardgameTableViewBase<
  GS extends Record<string, unknown> = Record<string, unknown>,
  PS extends Record<string, unknown> = Record<string, unknown>,
  MN extends string = string,
  MA extends Record<string, Record<string, unknown>> = Record<string, Record<string, unknown>>
> extends BoardgameBaseGameRenderer<GS, PS, MN, MA> {

  /**
   * Per-seat avatar + name records, indexed by player index. May contain
   * gaps (e.g. seat 2 unjoined → no entry for playerIndex=2). The Table
   * view's avatar strip renders one tile per non-empty seat.
   */
  @property({ type: Array })
  seatPresentations: SeatPresentation[] = [];

  /**
   * Player indices currently flagged absent (heartbeat stale). The Table
   * view renders a "Waiting for <name>" badge over the corresponding
   * avatar tile and (if the absent player is also the current player)
   * exposes a SkipTurn button to the host.
   */
  @property({ type: Array })
  absentPlayers: number[] = [];

  /**
   * True iff this Table view is being rendered for the host (game creator
   * with the surface=table cookie). Host-only controls (Lock room toggle,
   * SkipTurn button on the current-player badge) are gated on this flag.
   */
  @property({ type: Boolean })
  isHost = false;

  /**
   * True iff CompanionLocked is set on this game (host has explicitly
   * locked the room against new joiners). Renders as a closed-padlock
   * indicator on the room code display.
   */
  @property({ type: Boolean })
  roomLocked = false;

  /**
   * The 4 (or 5) letter room code displayed during the pre-game lobby and,
   * smaller, in a corner during gameplay. Empty string before the server
   * has resolved it.
   */
  @property({ type: String })
  roomCode = '';

  /**
   * Server-clock instant (ms since epoch) at which this state's cross-
   * screen animation should begin playing. Set by boardgame-game-state-
   * manager from the "version-timing" WebSocket message (spec §8.4).
   * null = play immediately on state install (the V1 default for missing
   * timing data).
   */
  @property({ type: Number })
  serverPlayAt: number | null = null;

  /**
   * Opt-in helper: renders the avatar strip across the top edge of the
   * Table view. Pulls from seatPresentations + absentPlayers. V1 stub —
   * Phase 3 wires up the real treatment with pulse animations for the
   * current player and "Waiting…" overlays for absent players.
   */
  protected renderAvatarStrip(): TemplateResult {
    return html``;
  }

  /**
   * Opt-in helper: renders host-only controls (Lock room toggle, SkipTurn
   * button on the absent current-player badge). No-op when !isHost.
   * V1 stub — Phase 3 wires the real UI + click handlers.
   */
  protected renderHostControls(): TemplateResult {
    return html``;
  }

  /**
   * Opt-in helper: renders the fake-deck row along the bottom edge of the
   * Table view, one off-screen-positioned stub stack per seated player
   * (left-to-right in seat order). This is the cross-screen animation
   * source/destination for cards moving between the public board and a
   * player's hand. V1 stub — Phase 4 wires the synthetic-ID stubs +
   * animateBetween() integration with the FLIP animator.
   */
  protected renderFakeDeckRow(): TemplateResult {
    return html``;
  }
}
