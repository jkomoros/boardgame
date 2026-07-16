import { LitElement, css, html } from 'lit';
import { property } from 'lit/decorators.js';

export type ActionBarOrientation = 'horizontal' | 'responsive' | 'vertical';
export type ActionBarAlignment = 'center' | 'end' | 'space-between' | 'start';

const orientations = new Set<ActionBarOrientation>(['horizontal', 'responsive', 'vertical']);
const alignments = new Set<ActionBarAlignment>(['center', 'end', 'space-between', 'start']);

export class BoardgameActionBar extends LitElement {
  static override styles = css`
    :host {
      display: block;
      container-type: inline-size;
    }

    #bar {
      display: flex;
      flex-flow: row wrap;
      align-items: center;
      justify-content: center;
      gap: var(--boardgame-action-gap, 0.75rem);
    }

    #bar.start { justify-content: flex-start; }
    #bar.end { justify-content: flex-end; }
    #bar.space-between { justify-content: space-between; }

    #bar.vertical {
      flex-direction: column;
      align-items: stretch;
    }

    #bar.vertical.start { align-items: flex-start; }
    #bar.vertical.center { align-items: center; }
    #bar.vertical.end { align-items: flex-end; }

    #bar.vertical ::slotted(boardgame-action-button) {
      --boardgame-action-width: 100%;
    }

    @container (max-width: 30rem) {
      #bar.responsive {
        flex-direction: column;
        align-items: stretch;
      }

      #bar.responsive ::slotted(boardgame-action-button) {
        --boardgame-action-width: 100%;
      }
    }
  `;

  /** Accessible name for the action group. The useful default keeps the common case trivial. */
  @property({ type: String })
  label = 'Game actions';

  @property({ type: String })
  orientation: ActionBarOrientation = 'responsive';

  @property({ type: String })
  alignment: ActionBarAlignment = 'center';

  override render() {
    this.#validateConfiguration();
    return html`
      <div
        id="bar"
        part="bar"
        class=${`${this.orientation} ${this.alignment}`}
        role="group"
        aria-label=${this.label.trim()}>
        <slot></slot>
      </div>
    `;
  }

  #validateConfiguration(): void {
    if (!this.label.trim()) {
      throw new Error('boardgame-action-bar: label must be a non-empty accessible group name');
    }
    if (!orientations.has(this.orientation)) {
      throw new Error(`boardgame-action-bar: unknown orientation "${this.orientation}"`);
    }
    if (!alignments.has(this.alignment)) {
      throw new Error(`boardgame-action-bar: unknown alignment "${this.alignment}"`);
    }
  }
}

customElements.define('boardgame-action-bar', BoardgameActionBar);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-action-bar': BoardgameActionBar;
  }
}
