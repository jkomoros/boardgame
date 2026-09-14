import { html, css } from 'lit';
import { glyphForSlug, projectedPlayerChoices, viewPlayerProp } from '../../src/client.js';
import { HandRenderer, registerHandRenderer } from './_game_renderer.js';
import { MoveNames } from './_move_names.js';

/**
 * Werewolf Hand view (each player's phone). Connects as PlayerIndex(n)
 * so the player can see their own role (via sanitize:"other:hidden" on
 * behaviors.PlayerRole). Shows role, voting buttons, and game status.
 */
@registerHandRenderer
export class WerewolfHandView extends HandRenderer {
  static override styles = [
    HandRenderer.styles,
    css`
      :host {
        display: block;
        min-height: 100vh;
        padding: 16px;
        background: #1a1a2e;
        color: #e0e0e0;
        font-family: system-ui, sans-serif;
      }
      h1 {
        text-align: center;
        margin: 0 0 8px 0;
        font-size: 20px;
      }
      .role-banner {
        text-align: center;
        font-size: 28px;
        font-weight: 900;
        padding: 16px;
        border-radius: 12px;
        margin: 16px auto;
        max-width: 320px;
      }
      .role-villager {
        background: #2e7d32;
        color: white;
      }
      .role-werewolf {
        background: #b71c1c;
        color: white;
      }
      .role-unknown {
        background: #455a64;
        color: white;
      }
      .fellow-wolves {
        text-align: center;
        font-size: 14px;
        color: #ef9a9a;
        margin: 8px 0;
      }
      .phase-info {
        text-align: center;
        font-size: 16px;
        margin: 12px 0;
        color: #90caf9;
      }
      boardgame-timer {
        display: block;
        max-width: 20rem;
        margin: 0.75rem auto;
      }
      .vote-section {
        margin: 16px auto;
        max-width: 320px;
      }
      .vote-section boardgame-target-list {
        --boardgame-target-list-gap: 8px;
      }
      .voted-message {
        text-align: center;
        font-size: 18px;
        color: #81c784;
        margin: 16px 0;
        font-weight: 600;
      }
      .sleep-message {
        text-align: center;
        font-size: 24px;
        margin: 32px 0;
        color: #7986cb;
        font-style: italic;
      }
      .eliminated-banner {
        text-align: center;
        font-size: 24px;
        font-weight: 700;
        color: #ef5350;
        margin: 32px 0;
        padding: 24px;
        border: 2px solid #ef5350;
        border-radius: 12px;
      }
    `,
  ];

  override render() {
    const game = this.state?.Game;
    const player = this.playerState;
    const phase = game?.Phase ?? 'Gathering';

    if (!player || phase === 'Gathering') {
      return html`
        ${this.renderTopEdgeAnchor()}
        <h1>Werewolf</h1>
        <div class="phase-info">Waiting for the game to start...</div>
      `;
    }

    const role = viewPlayerProp(this.state, this.viewingAsPlayer, 'Role');
    const isWerewolf = role.known && role.value === 'Werewolf';
    const isEliminated = player.Eliminated;
    const vote = phase === 'Night'
      ? viewPlayerProp(this.state, this.viewingAsPlayer, 'NightVote')
      : viewPlayerProp(this.state, this.viewingAsPlayer, 'DayVote');
    const hasVoted = vote.known && vote.value >= 0;

    // FellowWolves is computed server-side and visible only to this player.
    // Reading other players' sanitized Role values cannot reveal teammates.
    const fellowWolves = isWerewolf ? player.FellowWolves : [];

    // Keep game-owned avatar labels while using the server-projected candidate
    // universe and its already-bound actions.
    const nameFor = (i: number): string => {
      const seat = this.seatPresentations.find((s) => s.playerIndex === i);
      return seat ? `${glyphForSlug(seat.avatarSlug)} ${seat.displayName}` : `Player ${i + 1}`;
    };
    const voteSet = phase === 'Night'
      ? this.choices?.get(MoveNames.CastNightVote)
      : this.choices?.get(MoveNames.CastVote);
    const votes = voteSet ? projectedPlayerChoices(voteSet, nameFor) : null;

    return html`
      ${this.renderTopEdgeAnchor()}
      <h1>Werewolf</h1>

      <!-- Role banner -->
      <div class="role-banner ${!role.known ? 'role-unknown' : isWerewolf ? 'role-werewolf' : 'role-villager'}">
        ${!role.known ? 'ROLE UNKNOWN' : isWerewolf ? 'WEREWOLF' : 'VILLAGER'}
      </div>

      ${isWerewolf && fellowWolves.length > 0 ? html`
        <div class="fellow-wolves">
          Fellow wolves: ${fellowWolves.map(nameFor).join(', ')}
        </div>
      ` : ''}

      ${isEliminated ? html`
        <div class="eliminated-banner">You have been eliminated</div>
      ` : html`
        <div class="phase-info">
          ${phase === 'Day' ? `Day - Round ${(game?.RoundNumber ?? 0) + 1}` : ''}
          ${phase === 'Night' ? `Night - Round ${(game?.RoundNumber ?? 0) + 1}` : ''}
        </div>
        <boardgame-timer
          label="Voting ends in"
          .timer=${game?.VoteTimer ?? null}>
        </boardgame-timer>

        ${phase === 'Night' && !isWerewolf ? html`
          <div class="sleep-message">Sleep tight...</div>
        ` : html`
          ${hasVoted ? html`
            <div class="voted-message">Vote cast. Waiting for others...</div>
          ` : html`
            ${votes ? html`<div class="vote-section">
              <boardgame-target-list
                .label=${phase === 'Day' ? 'Vote to eliminate' : 'Choose a target'}
                .choices=${votes}>
              </boardgame-target-list>
            </div>` : ''}
          `}
        `}
      `}
    `;
  }
}
