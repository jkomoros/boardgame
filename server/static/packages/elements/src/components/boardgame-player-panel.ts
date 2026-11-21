import { LitElement, css, html } from 'lit';
import { property, state } from 'lit/decorators.js';

type OptionalPlayerPanelSlot = 'actions' | 'footer' | 'status';

/** Semantic, responsive container for arbitrary game-owned player content. */
export class BoardgamePlayerPanel extends LitElement {
  static override styles = css`
    :host {
      display: block;
      min-width: 0;
      height: 100%;
      container-type: inline-size;
    }

    #panel {
      box-sizing: border-box;
      display: grid;
      align-content: start;
      gap: var(--boardgame-player-panel-gap, 0.75rem);
      min-width: 0;
      min-height: 100%;
      padding: var(--boardgame-player-panel-padding, 1rem);
      border: var(--boardgame-player-panel-border, 1px solid color-mix(in srgb, currentColor 20%, transparent));
      border-radius: var(--boardgame-player-panel-radius, 0.875rem);
      background: var(--boardgame-player-panel-background, color-mix(in srgb, currentColor 3%, transparent));
      color: var(--boardgame-player-panel-color, inherit);
    }

    #panel.active {
      border-color: var(--boardgame-player-panel-active-border, color-mix(in srgb, #2e7d32 65%, transparent));
      background: var(--boardgame-player-panel-active-background, color-mix(in srgb, #2e7d32 9%, transparent));
      box-shadow: var(--boardgame-player-panel-active-shadow, 0 0 0 1px color-mix(in srgb, #2e7d32 25%, transparent));
    }

    #header {
      display: flex;
      flex-wrap: wrap;
      align-items: baseline;
      gap: var(--boardgame-player-panel-header-gap, 0.5rem 0.75rem);
      min-width: 0;
    }

    #heading {
      margin: 0;
      font: inherit;
      font-size: var(--boardgame-player-panel-heading-size, 1.125rem);
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

    #active {
      padding: 0.125rem 0.5rem;
      border-radius: 999px;
      background: var(--boardgame-player-panel-active-label-background, color-mix(in srgb, #2e7d32 18%, transparent));
      font-size: 0.75rem;
      font-weight: 700;
    }

    #content,
    #status,
    #actions,
    #footer {
      min-width: 0;
    }

    #actions {
      margin-block-start: auto;
    }

    #footer {
      font-size: var(--boardgame-player-panel-footer-size, 0.875rem);
      color: var(--boardgame-player-panel-footer-color, inherit);
    }

    [hidden] { display: none !important; }
    ::slotted(*) { box-sizing: border-box; min-width: 0; max-width: 100%; }

    @container (max-width: 18rem) {
      #panel { padding: var(--boardgame-player-panel-narrow-padding, 0.75rem); }
      #header { align-items: stretch; flex-direction: column; }
    }
  `;

  @property({ type: String })
  label = '';

  @property({ type: Number, attribute: 'heading-level' })
  headingLevel = 3;

  @property({ type: Boolean, attribute: 'hide-heading' })
  hideHeading = false;

  @property({ type: Boolean, reflect: true })
  active = false;

  @property({ type: String, attribute: 'active-label' })
  activeLabel = 'Current player';

  @state()
  private readonly populatedSlots: Record<OptionalPlayerPanelSlot, boolean> = {
    actions: false,
    footer: false,
    status: false,
  };

  override render() {
    this.#validateConfiguration();
    return html`
      <section
        id="panel"
        part=${this.active ? 'panel active' : 'panel'}
        class=${this.active ? 'active' : ''}
        aria-labelledby="heading"
        aria-current=${this.active ? 'true' : 'false'}>
        <header id="header" part="header">
          <div id="heading" part="heading" class=${this.hideHeading ? 'visually-hidden' : ''}
            role="heading" aria-level=${this.headingLevel}>
            ${this.label.trim()}
          </div>
          ${this.active ? html`<span id="active" part="active-label">${this.activeLabel.trim()}</span>` : null}
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
    if (typeof this.label !== 'string' || !this.label.trim()) {
      throw new Error('boardgame-player-panel: label must be a non-empty player name');
    }
    if (!Number.isSafeInteger(this.headingLevel) || this.headingLevel < 1 || this.headingLevel > 6) {
      throw new Error('boardgame-player-panel: headingLevel must be a safe integer from 1 through 6');
    }
    if (typeof this.activeLabel !== 'string' || !this.activeLabel.trim()) {
      throw new Error('boardgame-player-panel: activeLabel must be a non-empty string');
    }
  }
}

customElements.define('boardgame-player-panel', BoardgamePlayerPanel);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-player-panel': BoardgamePlayerPanel;
  }
}
