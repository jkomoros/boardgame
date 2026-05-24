import { html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import { BoardgameTableViewBase } from '../../../server/static/src/components/boardgame-table-view-base.js';
import type { MoveName } from './_move_names.js';
import type { GameState, PlayerState } from './_types.js';

/**
 * Werewolf Table view (the shared projector screen). Connects as
 * ObserverPlayerIndex so roles are hidden by sanitization. Shows player
 * status, voting progress, phase info, and win/loss announcements.
 */
@customElement('boardgame-render-game-werewolf-table')
export class WerewolfTableView extends BoardgameTableViewBase<GameState, PlayerState, MoveName> {
  static override styles = [
    BoardgameTableViewBase.styles,
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

    // Check win conditions for banner
    const activePlayers = players.filter(p => !p.PlayerInactive);
    const alive = activePlayers.filter(p => !p.Eliminated);
    const aliveWerewolves = alive.filter(p => p.Role === 'Werewolf').length;
    const aliveVillagers = alive.filter(p => p.Role !== 'Werewolf').length;
    const gameOver = this.state?.Game && (aliveWerewolves === 0 || aliveWerewolves >= aliveVillagers) && phase !== 'Gathering';

    // Determine phase CSS class
    let phaseClass = 'phase-gathering';
    if (phase === 'Day') phaseClass = 'phase-day';
    if (phase === 'Night') phaseClass = 'phase-night';

    return html`
      <h1>Werewolf</h1>
      ${this.renderRoomCodeBanner()}
      ${this.renderAvatarStrip()}
      ${this.renderHostControls()}

      ${gameOver ? html`
        ${aliveWerewolves === 0
          ? html`<div class="winner-banner winner-villagers">Villagers Win!</div>`
          : html`<div class="winner-banner winner-werewolves">Werewolves Win!</div>`
        }
      ` : ''}

      <div class="phase-banner ${phaseClass}">
        ${phase === 'Gathering' ? 'Waiting for players...' : `${phase} - Round ${round}`}
      </div>

      ${phase === 'Night' && !gameOver ? html`
        <div class="night-message">Night time -- werewolves are choosing...</div>
      ` : ''}

      <div class="players-circle">
        ${activePlayers.map((player, index) => {
          const realIndex = players.indexOf(player);
          const hasVoted = player.Vote >= 0;
          return html`
            <div class="player-tile ${player.Eliminated ? 'eliminated' : ''}">
              <div class="name">Player ${realIndex}</div>
              ${player.Eliminated
                ? html`<div class="status">ELIMINATED</div>`
                : html`
                  ${phase === 'Day' && hasVoted
                    ? html`<div class="vote-info">Voted for P${player.Vote}</div>`
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
