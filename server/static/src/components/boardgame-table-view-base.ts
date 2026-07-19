import { html, css, TemplateResult, type CSSResultGroup } from 'lit';
import { property, state } from 'lit/decorators.js';
import { BoardgameBaseGameRenderer } from './boardgame-base-game-renderer.js';
import type { FullGameState } from '../types/boardgame-types.js';
import type { MoveChoiceProjectionTypes } from '../moves/projected-choices.js';
import { glyphForSlug } from './companion-avatar-catalog.js';
import { apiHttpPost, buildGameUrl, type ApiResponse } from '../api.js';
import { decodeHostActionResponse } from '../types/host-action-response.js';
import { apiPath } from '../util.js';
import './boardgame-game-outcome.js';
import type { EffectTransitionContext } from '../effects/effect-spec.js';
import type { MotionTransferDeclaration } from '../motion/transfer.js';

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
 * The opt-in helpers are implemented here, including avatar/host chrome,
 * outcome UI, the fake-deck row, and declarative Table deal presentation.
 */
export class BoardgameTableViewBase<
  S extends FullGameState<object, object, object, object, object>,
  C extends object,
  MN extends string,
  MA extends Record<string, object>,
  K extends object = object,
  E extends object = object,
  MCP extends MoveChoiceProjectionTypes = Record<never, never>,
> extends BoardgameBaseGameRenderer<S, C, MN, MA, K, E, MCP> {

  /**
   * Per-seat avatar + name records, indexed by player index. May contain
   * gaps (e.g. seat 2 unjoined → no entry for playerIndex=2). The Table
   * view's avatar strip renders one tile per non-empty seat.
   */
  @property({ attribute: false })
  seatPresentations: readonly SeatPresentation[] = Object.freeze([]);

  /**
   * Player indices currently flagged absent (heartbeat stale). The Table
   * view renders a "Waiting for <name>" badge over the corresponding
   * avatar tile and (if the absent player is also the current player)
   * exposes a SkipTurn button to the host.
   */
  @property({ attribute: false })
  absentPlayers: readonly number[] = Object.freeze([]);

  /**
   * True iff this Table view is being rendered for the host (game creator
   * with the surface=table cookie). Host-only controls (Lock room toggle,
   * SkipTurn button on the current-player badge) are gated on this flag.
   */
  @property({ type: Boolean })
  isHost = false;

  /**
   * When true (the default), the base purely compares adjacent authoritative
   * snapshots and declares one retained-stub transfer for each player whose
   * aggregate sanitized hand count grows. The element marked id="deal-source"
   * supplies geometry for the projector half of the cross-screen deal.
   * Missing source or stub becomes an ordinary terminal skipped segment.
   * Hand size = total length of all Stack-shaped playerState properties
   * (sanitized stacks still carry placeholder indexes, so counts survive
   * hiding). Set false for bespoke animation wiring.
   */
  @property({ type: Boolean })
  autoFlyDeals = true;

  override motionTransfersForTransition(
    context: EffectTransitionContext<S, MN>,
  ): readonly MotionTransferDeclaration[] {
    const inherited = super.motionTransfersForTransition(context);
    if (context.kind === 'initial' || !this.autoFlyDeals) return inherited;
    const before = this._handSizes(context.before);
    const after = this._handSizes(context.after);
    const arrivals = after.flatMap((size, playerIndex) => (
      size > (before[playerIndex] ?? 0)
        ? [Object.freeze({
          key: `auto-table:p${playerIndex}:hand-growth`,
          subjectId: `player-${playerIndex}-hand-growth`,
          source: 'deal-source',
          carrier: `stub:p${playerIndex}:hand`,
          durationMs: 600,
        })]
        : []
    ));
    return Object.freeze([...inherited, ...arrivals]);
  }

  private _handSizes(state: S): number[] {
    const players = state.Players ?? [];
    return players.map((p) => {
      let total = 0;
      for (const value of Object.values(p as Record<string, unknown>)) {
        const indexes = (value as { Indexes?: unknown })?.Indexes;
        if (Array.isArray(indexes)) total += indexes.length;
      }
      return total;
    });
  }

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

  // gameName + gameId are populated by boardgame-render-game's
  // _instantiateRenderer (P5 polish). Used to build the URL path for
  // host-action endpoints (hostSkipTurn, setRoomLock, switchToSolo).
  // Without these the host-action handlers below silently no-op — they
  // were previously trying to dig the gameName out of the sanitized
  // state object via state.Manager.Delegate.Name(), which doesn't exist
  // on the client-side state.
  @state()
  private _hostFeedback = '';

  @state()
  private _hostActionPending: 'lock' | 'solo' | 'skip' | null = null;

  private _hostFeedbackTimer: ReturnType<typeof setTimeout> | null = null;

  private _showHostFeedback(msg: string) {
    this._hostFeedback = msg;
    if (this._hostFeedbackTimer) clearTimeout(this._hostFeedbackTimer);
    this._hostFeedbackTimer = setTimeout(() => { this._hostFeedback = ''; }, 4000);
  }

  /**
   * Opt-in helper: renders the avatar strip across the top edge of the
   * Table view. Tile per seated player; per-tile pulse for the current
   * player; "Waiting for Alice (m:ss)" overlay for absent players plus
   * (if isHost AND this absent player is also current) a Skip button.
   *
   * Authors call this from render() at the spot in the layout they want
   * the avatar strip. Skip wiring goes through the game's normal API
   * base URL — the strip POSTs hostSkipTurn directly rather than going
   * through a Redux action, keeping the dependency footprint of the
   * Table view base small.
   */
  protected renderAvatarStrip(): TemplateResult {
    const seats = this._seatedSeats();
    return html`
      <div class="avatar-strip">
        ${seats.map(s => this._renderSeatTile(s))}
      </div>
    `;
  }

  /**
   * Opt-in helper: renders host-only controls (Lock room toggle). Skip
   * button is rendered inline on the absent-current-player badge by
   * renderAvatarStrip; this helper is for chrome that doesn't fit in
   * the avatar strip. P5.5 wires the Lock-room toggle here.
   */
  protected renderHostControls(): TemplateResult {
    if (!this.isHost) return html``;
    return html`
      <div class="host-controls" aria-busy=${this._hostActionPending !== null ? 'true' : 'false'}>
        <label>
          <input type="checkbox"
            .checked=${this.roomLocked}
            ?disabled=${this._hostActionPending !== null}
            @change=${(e: Event) => this._onLockRoomToggle((e.target as HTMLInputElement).checked)}>
          Lock room (no new joins)
        </label>
        ${this._switchToSoloConfirming
          ? html`
            <div class="switch-to-solo-confirm">
              <span class="warning-text">This may reveal hidden info and cannot be undone.</span>
              <button class="switch-to-solo danger" ?disabled=${this._hostActionPending !== null}
                @click=${this._onSwitchToSolo}>
                Yes, switch to solo
              </button>
              <button ?disabled=${this._hostActionPending !== null} @click=${this._cancelSwitchToSolo}>Cancel</button>
            </div>
          `
          : html`
            <button class="switch-to-solo" ?disabled=${this._hostActionPending !== null}
              @click=${this._onSwitchToSolo}>
              Switch to solo mode
            </button>
          `
        }
        ${this._hostFeedback ? html`<div class="host-feedback">${this._hostFeedback}</div>` : ''}
      </div>
    `;
  }

  /**
   * Opt-in helper: renders the room-code banner (≥120pt code + QR) for
   * the Table view pre-game lobby (spec §12.1). The banner hides once
   * the game has started (state.Version > 0) so the playing surface
   * isn't cluttered with the join prompt. V1 uses an external QR
   * service (qrserver.com); production can swap in a self-hosted
   * generator as a P5+ polish item.
   */
  protected renderRoomCodeBanner(): TemplateResult {
    if (!this.roomCode) return html``;
    // Shrink to the corner badge once nobody else can be waited for —
    // every seat is either claimed or closed/inactivated (games with
    // WaitForEnoughPlayers legally start below capacity and permanently
    // close their unfilled seats). That's the moment the code stops
    // mattering to the room. (The old check read state.Game.Version,
    // which doesn't exist — Version lives on the game record — so the
    // giant lobby banner never shrank. And a version-based check would be
    // wrong anyway: seat claims bump the version while the lobby is still
    // gathering.)
    const players = this.state?.Players ?? [];
    const claimed = new Set(this.seatPresentations.map((s) => s.playerIndex));
    const roomSettled = players.length > 0 && players.every((p, i) =>
      claimed.has(i) || Reflect.get(p, 'SeatClosed') === true || Reflect.get(p, 'PlayerInactive') === true);
    if (roomSettled || this.gameFinished) {
      // Game has started — render a small persistent badge in the corner
      // rather than the giant pre-game banner.
      return html`
        <div class="room-code-mini">
          <small>Room <strong>${this.roomCode}</strong>${this.roomLocked ? html`🔒` : ''}</small>
        </div>
      `;
    }
    // Pre-game: full-bleed banner with QR + giant code. QR is served
    // self-hosted via /api/game/<name>/<id>/qrcode.png (P5+ polish —
    // replaces the earlier qrserver.com cross-origin call).
    const origin = encodeURIComponent(window.location.origin);
    const qrSrc = this.gameName && this.gameId
      ? apiPath(`game/${this.gameName}/${this.gameId}/qrcode.png`) + `?origin=${origin}`
      : '';
    return html`
      <div class="room-code-banner">
        <div class="room-code-instructions">
          <p>Go to <strong>${window.location.host}/join</strong></p>
          <p>and enter the code</p>
        </div>
        <div class="room-code-giant">${this.roomCode}</div>
        ${qrSrc ? html`<img class="room-code-qr" src=${qrSrc} alt="Join QR code"
          @error=${(e: Event) => { (e.target as HTMLImageElement).style.display = 'none'; }}>` : ''}
      </div>
    `;
  }

  private async _onLockRoomToggle(locked: boolean) {
    if (!this.gameName || !this.gameId) return;
    if (this._hostActionPending !== null) {
      this.requestUpdate();
      return;
    }
    this._hostActionPending = 'lock';
    try {
      const response = await apiHttpPost(buildGameUrl(this.gameName, this.gameId, 'setRoomLock'), { locked });
      if (!response.data) {
        this._showHostFeedback(this._hostActionError(response, 'Failed to update lock'));
        return;
      }
      const result = decodeHostActionResponse(response.data);
      if (result.locked !== locked) throw new Error('Server returned contradictory room lock state');
      this.roomLocked = locked;
      this._showHostFeedback(locked ? 'Room locked' : 'Room unlocked');
    } catch (error) {
      this._showHostFeedback(`Lock toggle failed: ${error instanceof Error ? error.message : String(error)}`);
    } finally {
      this._hostActionPending = null;
    }
  }

  @state()
  private _switchToSoloConfirming = false;

  private _switchToSoloTimer: ReturnType<typeof setTimeout> | null = null;

  private _onSwitchToSolo() {
    if (!this._switchToSoloConfirming) {
      this._switchToSoloConfirming = true;
      this._switchToSoloTimer = setTimeout(() => {
        this._switchToSoloConfirming = false;
      }, 5000);
      return;
    }
    this._doSwitchToSolo();
  }

  private _cancelSwitchToSolo() {
    this._switchToSoloConfirming = false;
    if (this._switchToSoloTimer) clearTimeout(this._switchToSoloTimer);
  }

  private async _doSwitchToSolo() {
    this._switchToSoloConfirming = false;
    if (this._switchToSoloTimer) clearTimeout(this._switchToSoloTimer);
    if (!this.gameName || !this.gameId) return;
    if (this._hostActionPending !== null) return;
    this._hostActionPending = 'solo';
    try {
      const response = await apiHttpPost(buildGameUrl(this.gameName, this.gameId, 'switchToSolo'), {});
      if (!response.data) {
        this._showHostFeedback(this._hostActionError(response, 'Switch to solo failed'));
        return;
      }
      decodeHostActionResponse(response.data);
      this._showHostFeedback('Switching to solo mode...');
    } catch (error) {
      this._showHostFeedback(`Switch to solo failed: ${error instanceof Error ? error.message : String(error)}`);
    } finally {
      this._hostActionPending = null;
    }
  }

  private _hostActionError(response: ApiResponse<unknown>, fallback: string): string {
    return response.error || response.friendlyError || fallback;
  }

  /**
   * The game-specific board area. Override this (instead of render()) to
   * get the standard Table chrome for free: the default render() composes
   * room-code banner → game-over banner → avatar strip → host controls →
   * YOUR BOARD → fake-deck row. Games that want a different arrangement
   * override render() itself and call the helpers à la carte — every
   * existing game does that today, so this default is purely additive.
   */
  protected renderBoard(): TemplateResult {
    return html``;
  }

  override render() {
    return html`
      ${this.renderRoomCodeBanner()}
      ${this.renderGameOverBanner()}
      ${this.renderAvatarStrip()}
      ${this.renderHostControls()}
      ${this.renderBoard()}
      ${this.renderFakeDeckRow()}
    `;
  }

  /**
   * Opt-in helper: renders the game-over celebration when the game is
   * finished — winners by avatar + display name on the shared screen.
   * Empty until gameFinished; renders a draw message when Winners is
   * empty. Call it near the top of render(); it's the projector's payoff
   * moment, so it's intentionally loud.
   */
  protected renderGameOverBanner(): TemplateResult {
    // Winners without a seat-presentation row (AI agents never have one;
    // a human's row write is deliberately non-fatal at join) still get
    // announced — by seat label — rather than being silently dropped,
    // which used to turn a real win into "It's a draw."
    const winnerLabels = this.gameWinners.map((i) => {
      const seat = this.seatPresentations.find((s) => s.playerIndex === i);
      return seat ? `${glyphForSlug(seat.avatarSlug)} ${seat.displayName}` : `Player ${i + 1}`;
    });
    return html`
      <boardgame-game-outcome
        .finished=${this.gameFinished}
        .animating=${this.animating}
        .winners=${this.gameWinners}
        .winnerLabels=${winnerLabels}>
      </boardgame-game-outcome>
    `;
  }

  /**
   * Opt-in helper: renders the fake-deck row along the bottom edge of the
   * Table view (spec §8). One stub stack per seated player, left-to-right
   * in seat order. Each stub element has id "stub:p<N>:hand" — a
   * synthetic ID distinct from any real component.id, so the FLIP
   * animator's flat _infoById map cannot collide with real cards. The id is
   * a private DOM anchor detail, not motion subject identity.
   *
   * Stubs are rendered with low opacity at the bottom of the screen, one per
   * seated player, with the seat's display name visible. `autoFlyDeals`
   * declares an arrival from #deal-source whenever adjacent sanitized
   * snapshots show that player's aggregate hand count grow. Games with more
   * precise semantics disable that lossy default and override
   * motionTransfersForTransition().
   *
   * Future polish: position stubs at the screen edge (off-viewport) so
   * cards visually "fly off" toward the player; for V1 they're visible
   * placeholders so authors can see and tune the layout.
   */
  protected renderFakeDeckRow(): TemplateResult {
    const seats = this._seatedSeats();
    return html`
      <div class="fake-deck-row">
        ${seats.map(s => html`
          <div class="fake-deck-stub" id="stub:p${s.playerIndex}:hand">
            <small>${s.displayName}</small>
          </div>
        `)}
      </div>
    `;
  }

  // ---- internal helpers below ----

  private _seatedSeats(): SeatPresentation[] {
    // Stable left-to-right order by playerIndex.
    return [...this.seatPresentations].sort((a, b) => a.playerIndex - b.playerIndex);
  }

  private _isAbsent(playerIndex: number): boolean {
    return this.absentPlayers.includes(playerIndex);
  }

  private _renderSeatTile(seat: SeatPresentation): TemplateResult {
    const absent = this._isAbsent(seat.playerIndex);
    const isCurrent = seat.playerIndex === this.currentPlayerIndex;
    const showSkip = this.isHost && absent && isCurrent;
    return html`
      <div class="seat-tile ${isCurrent ? 'current' : ''} ${absent ? 'absent' : ''}">
        <div class="seat-avatar">${glyphForSlug(seat.avatarSlug)}</div>
        <div class="seat-name">${seat.displayName}</div>
        ${absent ? html`<div class="seat-waiting">Waiting…</div>` : ''}
        ${showSkip ? html`<button class="seat-skip" ?disabled=${this._hostActionPending !== null}
          @click=${() => this._onSkipTurn(seat.playerIndex)}>Skip turn</button>` : ''}
      </div>
    `;
  }

  private async _onSkipTurn(playerIndex: number) {
    if (!this.gameName || !this.gameId) {
      this._showHostFeedback('Cannot skip — game info not loaded yet');
      return;
    }
    if (this._hostActionPending !== null) return;
    this._hostActionPending = 'skip';
    try {
      const response = await apiHttpPost(buildGameUrl(this.gameName, this.gameId, 'hostSkipTurn'), {});
      if (!response.data) {
        if (response.status === 429) {
          this._showHostFeedback('Please wait a moment before skipping again');
        } else if (response.status === 409) {
          this._showHostFeedback('Player is not absent yet — wait for them to disconnect');
        } else {
          this._showHostFeedback(this._hostActionError(response, 'Skip failed'));
        }
        return;
      }
      decodeHostActionResponse(response.data);
      this._showHostFeedback('Turn skipped');
    } catch (error) {
      this._showHostFeedback(`Skip failed: ${error instanceof Error ? error.message : String(error)}`);
    } finally {
      this._hostActionPending = null;
    }
  }

  static styles: CSSResultGroup = css`
    .avatar-strip {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
      padding: 12px;
      justify-content: center;
    }
    .seat-tile {
      min-width: 80px;
      padding: 8px 12px;
      border: 2px solid #ddd;
      border-radius: 12px;
      text-align: center;
      background: white;
      position: relative;
    }
    .seat-tile.current {
      border-color: #1a73e8;
      box-shadow: 0 0 0 4px rgba(26, 115, 232, 0.15);
    }
    .seat-tile.absent {
      opacity: 0.5;
    }
    .seat-avatar {
      font-size: 32px;
    }
    .seat-name {
      font-weight: 600;
      font-size: 14px;
    }
    .seat-waiting {
      color: #c62828;
      font-size: 12px;
      margin-top: 4px;
    }
    .seat-skip {
      margin-top: 4px;
      padding: 4px 12px;
      border: 1px solid #c62828;
      background: #fff;
      color: #c62828;
      border-radius: 6px;
      font-weight: 600;
      cursor: pointer;
    }
    .fake-deck-row {
      display: flex;
      justify-content: space-around;
      align-items: flex-end;
      padding: 8px 16px;
      gap: 8px;
      position: fixed;
      left: 0;
      right: 0;
      bottom: 0;
      pointer-events: none;
    }
    .fake-deck-stub {
      width: 64px;
      height: 88px;
      border: 1px dashed rgba(255, 255, 255, 0.5);
      border-radius: 6px;
      background: rgba(255, 255, 255, 0.18);
      text-align: center;
      padding-top: 4px;
      font-size: 10px;
      color: rgba(255, 255, 255, 0.85);
      /* Visible — these ARE the flying proxies for the cross-screen deal
         animation. opacity: 0 here silently hid the entire projector half
         of the effect (the stub flew a perfect arc nobody could see). */
      opacity: 0.9;
    }
    .room-code-banner {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 32px;
      padding: 24px;
      background: rgba(0,0,0,0.6);
      color: white;
      border-radius: 12px;
      margin: 24px auto;
      max-width: 720px;
    }
    .room-code-instructions p {
      margin: 4px 0;
      font-size: 20px;
    }
    .room-code-giant {
      font-size: 128px;
      font-weight: 900;
      letter-spacing: 16px;
      font-family: monospace;
    }
    .room-code-qr {
      width: 180px;
      height: 180px;
      background: white;
      border-radius: 8px;
      padding: 8px;
    }
    .room-code-mini {
      position: fixed;
      top: 8px;
      right: 8px;
      padding: 4px 8px;
      background: rgba(0,0,0,0.5);
      color: white;
      border-radius: 4px;
      font-family: monospace;
    }
    .host-controls {
      display: flex;
      gap: 16px;
      align-items: center;
      padding: 8px;
    }
    .host-controls .switch-to-solo {
      padding: 6px 12px;
      border: 1px solid #888;
      background: #fff;
      border-radius: 6px;
      cursor: pointer;
    }
    .host-controls .switch-to-solo.danger {
      border-color: #c62828;
      color: #c62828;
      font-weight: 600;
    }
    .switch-to-solo-confirm {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .switch-to-solo-confirm .warning-text {
      color: #c62828;
      font-size: 13px;
      font-weight: 600;
    }
    .host-feedback {
      padding: 4px 12px;
      background: rgba(0,0,0,0.7);
      color: white;
      border-radius: 6px;
      font-size: 13px;
      animation: fadeIn 0.2s ease;
    }
    @keyframes fadeIn {
      from { opacity: 0; }
      to { opacity: 1; }
    }
  `;
}
