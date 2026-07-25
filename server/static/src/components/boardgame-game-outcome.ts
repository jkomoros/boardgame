import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';

export type GameOutcomeViewer = number | null;

/**
 * Server-authoritative verdict presentation. The verdict remains absent while
 * the final animation is in flight and is announced only after the board settles.
 */
export class BoardgameGameOutcome extends BoardgameAnimatableItem {
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

  // Latches so the arrival plays exactly once per reveal. render() returns
  // null while `!finished || animating` (#outcome only exists in the DOM
  // once revealed), so this must run in updated() -- which fires AFTER
  // render() has applied the new DOM -- rather than gating on some earlier
  // hook where #outcome wouldn't exist yet. Unlike fading-text's
  // animateFade() (which defers the play() call through
  // updateComplete.then(...) and therefore needs a generation token to
  // survive a mid-flight retrigger racing that continuation), this reveal
  // gate calls play() synchronously inside updated() with no async gap for
  // a stale continuation to land in, so a simple boolean latch is
  // sufficient here.
  private _arrivalPlayed = false;

  override updated(changed: Map<PropertyKey, unknown>) {
    super.updated(changed);
    const revealed = this.finished && !this.animating;
    if (revealed && !this._arrivalPlayed) {
      this._arrivalPlayed = true;
      const outcome = this.renderRoot.querySelector('#outcome') as HTMLElement | null;
      if (outcome) {
        this.play(outcome, [
          { opacity: 0, transform: 'scale(0.96)' },
          { opacity: 1, transform: 'scale(1)' },
        ], { duration: 220, easing: 'ease-out', fill: 'backwards' });
      }
    }
    if (!revealed) this._arrivalPlayed = false;
  }

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
