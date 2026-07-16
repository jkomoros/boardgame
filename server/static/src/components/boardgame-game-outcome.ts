import { LitElement, css, html } from 'lit';
import { property } from 'lit/decorators.js';

export type GameOutcomeViewer = number | null;

/**
 * Server-authoritative verdict presentation. The verdict remains absent while
 * the final animation is in flight and is announced only after the board settles.
 */
export class BoardgameGameOutcome extends LitElement {
  static override styles = css`
    :host {
      display: block;
      container-type: inline-size;
    }

    #outcome {
      box-sizing: border-box;
      max-width: var(--boardgame-outcome-max-width, 40rem);
      margin: var(--boardgame-outcome-margin, 1.5rem auto);
      padding: var(--boardgame-outcome-padding, 1.5rem);
      border: var(--boardgame-outcome-border, 2px solid #d6a700);
      border-radius: var(--boardgame-outcome-radius, 1rem);
      text-align: center;
      background: var(--boardgame-outcome-background, color-mix(in srgb, #ffd700 15%, transparent));
      animation: outcome-arrive 220ms ease-out both;
    }

    #title {
      margin: 0 0 0.5rem;
      font-size: var(--boardgame-outcome-title-size, clamp(1.75rem, 6cqi, 2.5rem));
      font-weight: 800;
    }

    #message {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      justify-content: center;
      gap: 0.5rem;
      font-size: var(--boardgame-outcome-message-size, clamp(1.125rem, 4cqi, 1.75rem));
    }

    .winner {
      font-weight: 700;
    }

    @keyframes outcome-arrive {
      from { opacity: 0; transform: scale(0.96); }
      to { opacity: 1; transform: scale(1); }
    }

    @media (prefers-reduced-motion: reduce) {
      #outcome { animation: none; }
    }
  `;

  @property({ type: Boolean })
  finished = false;

  @property({ type: Boolean })
  animating = false;

  @property({ type: Array })
  winners: readonly number[] = [];

  /** Optional display names in exactly the same order as winners. */
  @property({ type: Array, attribute: false })
  winnerLabels: readonly string[] = [];

  /** Null renders a public verdict; a player index renders win/loss language. */
  @property({ attribute: false })
  viewer: GameOutcomeViewer = null;

  @property({ type: String })
  title = 'Game over!';

  override render() {
    this._validateConfiguration();
    if (!this.finished || this.animating) return null;
    const labels = this.winners.map((winner, index) => this.winnerLabels[index] ?? `Player ${winner + 1}`);
    const message = this.viewer === null
      ? this._publicMessage(labels)
      : this._personalMessage(this.viewer);
    return html`
      <section id="outcome" part="outcome" role="status" aria-live="polite" aria-atomic="true">
        <h2 id="title" part="title">${this.title.trim()}</h2>
        <div id="message" part="message">${message}</div>
      </section>
    `;
  }

  private _publicMessage(labels: readonly string[]) {
    if (labels.length === 0) return html`It's a draw.`;
    return html`
      ${labels.map(label => html`<span class="winner" part="winner">${label}</span>`)}
      <span>${labels.length === 1 ? 'wins!' : 'win!'}</span>
    `;
  }

  private _personalMessage(viewer: number) {
    if (this.winners.length === 0) return html`It's a draw.`;
    return this.winners.includes(viewer) ? html`You won!` : html`You lost.`;
  }

  private _validateConfiguration(): void {
    if (!this.title.trim()) throw new Error('boardgame-game-outcome: title must be non-empty');
    if (!Array.isArray(this.winners)) throw new Error('boardgame-game-outcome: winners must be an array of player indexes');
    const seen = new Set<number>();
    for (const winner of this.winners) {
      if (!Number.isSafeInteger(winner) || winner < 0) {
        throw new Error('boardgame-game-outcome: every winner must be a nonnegative safe player index');
      }
      if (seen.has(winner)) throw new Error(`boardgame-game-outcome: duplicate winner index ${winner}`);
      seen.add(winner);
    }
    if (this.viewer !== null && (!Number.isSafeInteger(this.viewer) || this.viewer < 0)) {
      throw new Error('boardgame-game-outcome: viewer must be null or a nonnegative safe player index');
    }
    if (!Array.isArray(this.winnerLabels)) throw new Error('boardgame-game-outcome: winnerLabels must be an array');
    if (this.winnerLabels.length !== 0 && this.winnerLabels.length !== this.winners.length) {
      throw new Error('boardgame-game-outcome: winnerLabels must be empty or have exactly one label per winner');
    }
    for (const label of this.winnerLabels) {
      if (typeof label !== 'string' || !label.trim()) {
        throw new Error('boardgame-game-outcome: winner labels must be non-empty strings');
      }
    }
    if (!this.finished && this.winners.length > 0) {
      throw new Error('boardgame-game-outcome: winners cannot be present before finished is true');
    }
  }
}

customElements.define('boardgame-game-outcome', BoardgameGameOutcome);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-game-outcome': BoardgameGameOutcome;
  }
}
