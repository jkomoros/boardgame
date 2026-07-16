import { LitElement, css, html } from 'lit';
import { property, state } from 'lit/decorators.js';

/** Responsive, semantic layout for arbitrary game-owned player panels. */
export class BoardgamePlayerGrid extends LitElement {
  static override styles = css`
    :host {
      display: block;
      min-width: 0;
      container-type: inline-size;
    }

    #region {
      min-width: 0;
    }

    #heading {
      margin: 0 0 var(--boardgame-player-grid-gap, 1rem);
      font: inherit;
      font-size: var(--boardgame-player-grid-heading-size, 1.25rem);
      font-weight: 700;
    }

    #heading.visually-hidden {
      position: absolute;
      width: 1px;
      height: 1px;
      padding: 0;
      margin: -1px;
      overflow: hidden;
      clip: rect(0, 0, 0, 0);
      white-space: nowrap;
      border: 0;
    }

    #grid {
      display: grid;
      grid-template-columns: repeat(
        auto-fit,
        minmax(min(100%, var(--boardgame-player-grid-min-width, 12rem)), 1fr)
      );
      align-items: stretch;
      gap: var(--boardgame-player-grid-gap, 1rem);
      min-width: 0;
    }

    slot {
      display: contents;
    }

    ::slotted(*) {
      box-sizing: border-box;
      min-width: 0;
    }

    #empty {
      box-sizing: border-box;
      display: grid;
      min-block-size: var(--boardgame-player-grid-empty-size, 6rem);
      place-items: center;
      padding: 1rem;
      border: 1px dashed color-mix(in srgb, currentColor 28%, transparent);
      border-radius: 0.75rem;
      color: color-mix(in srgb, currentColor 62%, transparent);
      font-style: italic;
    }
  `;

  @property({ type: String })
  label = 'Players';

  @property({ type: Number, attribute: 'heading-level' })
  headingLevel = 2;

  @property({ type: Boolean, attribute: 'hide-heading' })
  hideHeading = false;

  @property({ type: Boolean, attribute: 'hide-empty-state' })
  hideEmptyState = false;

  @property({ type: String, attribute: 'empty-label' })
  emptyLabel = 'No players';

  @state()
  private itemCount = 0;

  override render() {
    this._validateConfiguration();
    return html`
      <section id="region" part="region" aria-labelledby="heading">
        <div
          id="heading"
          part="heading"
          class=${this.hideHeading ? 'visually-hidden' : ''}
          role="heading"
          aria-level=${this.headingLevel}>
          ${this.label.trim()}
        </div>
        <div id="grid" part="grid">
          <slot @slotchange=${this._slotChanged}></slot>
          ${this.itemCount === 0 && !this.hideEmptyState
            ? html`<div id="empty" part="empty">${this.emptyLabel.trim()}</div>`
            : null}
        </div>
      </section>
    `;
  }

  private _slotChanged(event: Event): void {
    const slot = event.target;
    if (!(slot instanceof HTMLSlotElement)) return;
    this.itemCount = slot.assignedElements({ flatten: true }).length;
  }

  private _validateConfiguration(): void {
    if (!this.label.trim()) throw new Error('boardgame-player-grid: label must be a non-empty player collection name');
    if (!Number.isSafeInteger(this.headingLevel) || this.headingLevel < 1 || this.headingLevel > 6) {
      throw new Error('boardgame-player-grid: headingLevel must be a safe integer from 1 through 6');
    }
    if (!this.hideEmptyState && !this.emptyLabel.trim()) {
      throw new Error('boardgame-player-grid: emptyLabel must be non-empty unless hideEmptyState is enabled');
    }
  }
}

customElements.define('boardgame-player-grid', BoardgamePlayerGrid);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-player-grid': BoardgamePlayerGrid;
  }
}
