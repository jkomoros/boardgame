import { LitElement, css, html } from 'lit';
import { property, state } from 'lit/decorators.js';

type OptionalSurfaceSlot = 'actions' | 'footer' | 'status';

/** Semantic, responsive root composition for a game renderer. */
export class BoardgameGameSurface extends LitElement {
  static override styles = css`
    :host {
      display: block;
      min-width: 0;
      container-type: inline-size;
    }

    #surface {
      box-sizing: border-box;
      display: grid;
      gap: var(--boardgame-game-surface-gap, 1rem);
      width: min(100%, var(--boardgame-game-surface-max-width, 72rem));
      min-width: 0;
      margin-inline: auto;
      padding: var(--boardgame-game-surface-padding, clamp(0.75rem, 3cqi, 1.5rem));
    }

    #header {
      display: flex;
      flex-wrap: wrap;
      align-items: baseline;
      justify-content: space-between;
      gap: var(--boardgame-game-surface-header-gap, 0.75rem);
      min-width: 0;
    }

    #heading {
      margin: 0;
      font: inherit;
      font-size: var(--boardgame-game-surface-heading-size, clamp(1.5rem, 6cqi, 2rem));
      font-weight: 700;
      line-height: 1.2;
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

    #content,
    #status,
    #actions,
    #footer {
      min-width: 0;
    }

    #status,
    #actions {
      display: grid;
      gap: var(--boardgame-game-surface-section-gap, 0.75rem);
    }

    #footer {
      color: var(--boardgame-game-surface-footer-color, inherit);
      font-size: var(--boardgame-game-surface-footer-size, 0.875rem);
    }

    [hidden] {
      display: none !important;
    }

    ::slotted(*) {
      box-sizing: border-box;
      min-width: 0;
      max-width: 100%;
    }

    @container (max-width: 30rem) {
      #surface {
        gap: var(--boardgame-game-surface-narrow-gap, 0.75rem);
      }

      #header {
        align-items: stretch;
        flex-direction: column;
      }
    }
  `;

  @property({ type: String })
  heading = '';

  @property({ type: Number, attribute: 'heading-level' })
  headingLevel = 2;

  @property({ type: Boolean, attribute: 'hide-heading' })
  hideHeading = false;

  @state()
  private readonly populatedSlots: Record<OptionalSurfaceSlot, boolean> = {
    actions: false,
    footer: false,
    status: false,
  };

  override render() {
    this.#validateConfiguration();
    return html`
      <section id="surface" part="surface" aria-labelledby="heading">
        <header id="header" part="header">
          <div
            id="heading"
            part="heading"
            class=${this.hideHeading ? 'visually-hidden' : ''}
            role="heading"
            aria-level=${this.headingLevel}>
            ${this.heading.trim()}
          </div>
          <slot name="header"></slot>
        </header>
        <div id="status" part="status" ?hidden=${!this.populatedSlots.status}>
          <slot name="status" @slotchange=${this.#slotChanged}></slot>
        </div>
        <div id="content" part="content"><slot></slot></div>
        <div id="actions" part="actions" ?hidden=${!this.populatedSlots.actions}>
          <slot name="actions" @slotchange=${this.#slotChanged}></slot>
        </div>
        <footer id="footer" part="footer" ?hidden=${!this.populatedSlots.footer}>
          <slot name="footer" @slotchange=${this.#slotChanged}></slot>
        </footer>
      </section>
    `;
  }

  #slotChanged(event: Event): void {
    const slot = event.target;
    if (!(slot instanceof HTMLSlotElement)) return;
    const name = slot.name;
    if (name !== 'actions' && name !== 'footer' && name !== 'status') return;
    const populated = slot.assignedElements({ flatten: true }).length > 0;
    if (this.populatedSlots[name] === populated) return;
    this.populatedSlots[name] = populated;
    this.requestUpdate();
  }

  #validateConfiguration(): void {
    if (!this.heading.trim()) {
      throw new Error('boardgame-game-surface: heading must be a non-empty game name');
    }
    if (!Number.isSafeInteger(this.headingLevel) || this.headingLevel < 1 || this.headingLevel > 6) {
      throw new Error('boardgame-game-surface: headingLevel must be a safe integer from 1 through 6');
    }
  }
}

customElements.define('boardgame-game-surface', BoardgameGameSurface);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-game-surface': BoardgameGameSurface;
  }
}
