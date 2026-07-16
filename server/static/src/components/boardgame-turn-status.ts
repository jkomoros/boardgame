import { LitElement, css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { turnStatusPresentation, type TurnStatusContext } from '../status/turn-status.js';

/** Persistent, sentinel-aware status for ordinary and simultaneous turns. */
export class BoardgameTurnStatus extends LitElement {
  static override styles = css`
    :host { display: block; container-type: inline-size; }
    #status {
      box-sizing: border-box;
      width: fit-content;
      max-width: 100%;
      margin: var(--boardgame-turn-status-margin, 0 auto);
      padding: var(--boardgame-turn-status-padding, 0.5rem 0.875rem);
      border: var(--boardgame-turn-status-border, 1px solid color-mix(in srgb, currentColor 24%, transparent));
      border-radius: var(--boardgame-turn-status-radius, 999px);
      background: var(--boardgame-turn-status-background, color-mix(in srgb, currentColor 7%, transparent));
      color: var(--boardgame-turn-status-color, inherit);
      font-weight: 650;
      text-align: center;
    }
    #status.active {
      background: var(--boardgame-turn-status-active-background, color-mix(in srgb, #2e7d32 16%, transparent));
      border-color: var(--boardgame-turn-status-active-border, color-mix(in srgb, #2e7d32 55%, transparent));
    }
    #status.simultaneous {
      background: var(--boardgame-turn-status-simultaneous-background, color-mix(in srgb, #1565c0 14%, transparent));
      border-color: var(--boardgame-turn-status-simultaneous-border, color-mix(in srgb, #1565c0 50%, transparent));
    }
  `;

  @property({ attribute: false })
  turn: TurnStatusContext | null = null;

  @property({ type: Array, attribute: false })
  playerLabels: readonly string[] = [];

  @property({ type: String, attribute: 'active-label' })
  activeLabel = 'Your turn';

  @property({ type: String, attribute: 'simultaneous-label' })
  simultaneousLabel = 'All players may act';

  override render() {
    if (this.turn === null) return null;
    const presentation = turnStatusPresentation(
      this.turn,
      this.playerLabels,
      this.activeLabel,
      this.simultaneousLabel,
    );
    if (!presentation) return null;
    return html`
      <div id="status" part="status ${presentation.kind}" class=${presentation.kind}
        role="status" aria-live="polite" aria-atomic="true">
        ${presentation.message}
      </div>
    `;
  }
}

customElements.define('boardgame-turn-status', BoardgameTurnStatus);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-turn-status': BoardgameTurnStatus;
  }
}
