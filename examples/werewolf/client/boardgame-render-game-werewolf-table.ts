import { html, css, type ReadinessParticipant } from '../../src/client.js';
import { TableRenderer, registerTableRenderer } from './_game_renderer.js';

/**
 * Werewolf Table view (the shared projector screen). Connects as
 * ObserverPlayerIndex so roles are hidden by sanitization. Shows player
 * status, voting progress, phase info, and win/loss announcements.
 */
@registerTableRenderer
export class WerewolfTableView extends TableRenderer {
  static override styles = [
    TableRenderer.styles,
    css`
      :host {
        display: block;
        min-height: 100vh;
        padding: 24px;
        background: #1a1a2e;
        color: #e0e0e0;
        font-family: system-ui, sans-serif;
      }
      h1 {
        text-align: center;
        margin: 0 0 8px 0;
        font-size: 28px;
      }
      .phase-banner {
        text-align: center;
        font-size: 22px;
        font-weight: 700;
        padding: 12px;
        border-radius: 8px;
        margin: 16px auto;
        max-width: 480px;
      }
      .phase-day {
        background: #f9a825;
        color: #1a1a2e;
      }
      .phase-night {
        background: #283593;
        color: #e0e0e0;
      }
      .phase-gathering {
        background: #455a64;
        color: #e0e0e0;
      }
      .players-circle {
        display: flex;
        flex-wrap: wrap;
        justify-content: center;
        gap: 16px;
        margin: 24px auto;
        max-width: 720px;
      }
      .player-tile {
        width: 120px;
        padding: 12px;
        border: 2px solid #555;
        border-radius: 12px;
        text-align: center;
        background: #2a2a3e;
        position: relative;
      }
      .player-tile.eliminated {
        opacity: 0.4;
        text-decoration: line-through;
      }
      .player-tile .name {
        font-weight: 600;
        font-size: 14px;
        margin-bottom: 4px;
      }
      .player-tile .vote-info {
        font-size: 12px;
        color: #90caf9;
        margin-top: 4px;
      }
      .player-tile .status {
        font-size: 11px;
        color: #ef5350;
        margin-top: 4px;
        font-weight: 700;
      }
      .night-message {
        text-align: center;
        font-size: 18px;
        font-style: italic;
        color: #7986cb;
        margin: 16px 0;
      }
      boardgame-readiness {
        max-width: 30rem;
        margin: 1rem auto;
        --boardgame-readiness-background: #2a2a3e;
        --boardgame-readiness-border: 1px solid #555;
      }
      .winner-banner {
        text-align: center;
        font-size: 32px;
        font-weight: 900;
        padding: 24px;
        border-radius: 12px;
        margin: 24px auto;
        max-width: 600px;
      }
      .winner-villagers {
        background: #2e7d32;
        color: white;
      }
      .winner-werewolves {
        background: #b71c1c;
        color: white;
      }
    `,
  ];

  override render() {
    const game = this.state?.Game;
    const players = this.state?.Players ?? [];
    const phase = game?.Phase ?? 'Gathering';
    const round = (game?.RoundNumber ?? 0) + 1;

    // NOTE: win conditions CANNOT be computed here. The Table view sees
    // observer-sanitized state, where every hidden Role reads as the
    // enum's zero value ("Villager") — so "count the alive werewolves"
    // always returns 0 and the old banner declared "Villagers Win!" from
    // the very first Day. Game-over comes from the server via the plumbed
    // gameFinished/gameWinners from the delegate's CheckGameFinished hook
    // (renderGameOverBanner below).
    const activePlayers = players.flatMap((player, playerIndex) => player.PlayerInactive
      ? []
      : [{ player, playerIndex }]);
    const nameFor = (i: number): string => {
      const seat = this.seatPresentations.find((s) => s.playerIndex === i);
      return seat ? seat.displayName : `Player ${i}`;
    };

    // Determine phase CSS class
    let phaseClass = 'phase-gathering';
    if (phase === 'Day') phaseClass = 'phase-day';
    if (phase === 'Night') phaseClass = 'phase-night';
    const dayReadiness: readonly ReadinessParticipant<number>[] = activePlayers.map(({ player, playerIndex }) => ({
      key: playerIndex,
      label: nameFor(playerIndex),
      state: player.Eliminated ? 'not-required' : player.DayVote >= 0 ? 'ready' : 'waiting',
    }));

    return html`
      <h1>Werewolf</h1>
      ${this.renderRoomCodeBanner()}
      ${this.renderGameOverBanner()}
      ${this.renderAvatarStrip()}
      ${this.renderHostControls()}

      <div class="phase-banner ${phaseClass}">
        ${phase === 'Gathering' ? 'Waiting for players...' : `${phase} - Round ${round}`}
      </div>

      ${phase === 'Night' && !this.gameFinished ? html`
        <div class="night-message">Night time -- werewolves are choosing...</div>
      ` : ''}

      ${phase === 'Day' && !this.gameFinished ? html`
        <boardgame-readiness
          label="Day votes"
          complete-label="All votes cast"
          progress-label="votes cast"
          ready-label="Voted"
          waiting-label="Thinking"
          not-required-label="Eliminated"
          .participants=${dayReadiness}>
        </boardgame-readiness>
      ` : ''}

      <div class="players-circle">
        ${activePlayers.map(({ player, playerIndex }) => {
          const hasVoted = player.DayVote >= 0;
          return html`
            <div class="player-tile ${player.Eliminated ? 'eliminated' : ''}">
              <div class="name">${nameFor(playerIndex)}</div>
              ${player.Eliminated
                ? html`<div class="status">ELIMINATED</div>`
                : html`
                  ${phase === 'Day' && hasVoted
                    ? html`<div class="vote-info">Voted for ${nameFor(player.DayVote)}</div>`
                    : ''}
                  ${phase === 'Day' && !hasVoted && !player.Eliminated
                    ? html`<div class="vote-info">Thinking...</div>`
                    : ''}
                `
              }
            </div>
          `;
        })}
      </div>

      ${this.renderFakeDeckRow()}
    `;
  }
}
