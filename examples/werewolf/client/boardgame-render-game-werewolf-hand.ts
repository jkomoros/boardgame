import { html, css } from 'lit';
import { glyphForSlug, targetList } from '../../src/client.js';
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
    const allPlayers = this.state?.Players ?? [];
    const phase = game?.Phase ?? 'Gathering';

    if (!player || phase === 'Gathering') {
      return html`
        ${this.renderTopEdgeAnchor()}
        <h1>Werewolf</h1>
        <div class="phase-info">Waiting for the game to start...</div>
      `;
    }

    const isWerewolf = player.Role === 'Werewolf';
    const isEliminated = player.Eliminated;
    const hasVoted = player.Vote >= 0;
    const myIndex = this.viewingAs;

    // Find fellow werewolves (if this player is a werewolf)
    const fellowWolves: number[] = [];
    if (isWerewolf) {
      allPlayers.forEach((p, i) => {
        if (i !== myIndex && p.Role === 'Werewolf' && !p.PlayerInactive) {
          fellowWolves.push(i);
        }
      });
    }

    // Build list of alive, non-inactive players for voting. Label with the
    // avatar + display name people picked in the join flow (falling back
    // to "Player N" if the seat has no presentation) — voters know each
    // other as "🦊 BrightFox", not as seat indexes.
    const nameFor = (i: number): string => {
      const seat = this.seatPresentations.find((s) => s.playerIndex === i);
      return seat ? `${glyphForSlug(seat.avatarSlug)} ${seat.displayName}` : `Player ${i}`;
    };
    const voteIndexes: number[] = [];
    allPlayers.forEach((p, i) => {
      if (p.PlayerInactive || p.Eliminated) return;
      // The server rejects self-votes in EVERY phase (moves.go: "you
      // cannot vote for yourself") — offering yourself at night just
      // produces a silently-failing tap.
      if (i === myIndex) return;
      voteIndexes.push(i);
    });

    // Determine the correct move name for this phase
    const moveName = phase === 'Night' ? MoveNames.CastNightVote : MoveNames.CastVote;
    const votes = moveName === MoveNames.CastNightVote
      ? this.move(MoveNames.CastNightVote).targets(
        voteIndexes, VoteTarget => ({ VoteTarget }),
      )
      : this.move(MoveNames.CastVote).targets(
        voteIndexes, VoteTarget => ({ VoteTarget }),
      );

    return html`
      ${this.renderTopEdgeAnchor()}
      <h1>Werewolf</h1>

      <!-- Role banner -->
      <div class="role-banner ${isWerewolf ? 'role-werewolf' : 'role-villager'}">
        ${isWerewolf ? 'WEREWOLF' : 'VILLAGER'}
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

        ${phase === 'Night' && !isWerewolf ? html`
          <div class="sleep-message">Sleep tight...</div>
        ` : html`
          ${hasVoted ? html`
            <div class="voted-message">Vote cast. Waiting for others...</div>
          ` : html`
            <div class="vote-section">
              <boardgame-target-list
                .label=${phase === 'Day' ? 'Vote to eliminate' : 'Choose a target'}
                .choices=${targetList(votes, nameFor)}>
              </boardgame-target-list>
            </div>
          `}
        `}
      `}
    `;
  }
}
