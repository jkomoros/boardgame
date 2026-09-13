import { LitElement, css, html } from 'lit';
import { property, query } from 'lit/decorators.js';
import type { SelectionDraftSelectionBinding } from '../moves/selection-draft.js';
import type { TargetAction } from '../moves/target-action.js';
import type { ExpandedStack } from '../types/boardgame-types.js';
import type { ComponentView } from './component-view.js';
import './boardgame-component-zone.js';
import type { BoardgameComponentStack } from './boardgame-component-stack.js';

/** A labelled draw pile built from the ordinary stack and zone contracts. */
export class BoardgameDeck extends LitElement {
  static override styles = css`
    :host {
      display: block;
      min-width: 0;
      width: max-content;
      max-width: 100%;
    }

    boardgame-component-zone {
      min-width: 0;
      --boardgame-zone-min-block-size: 0;
    }
  `;

  @property({ type: String }) label = '';
  @property({ type: String, attribute: 'empty-label' }) emptyLabel = 'Exhausted';
  @property({ type: Number, attribute: 'heading-level' }) headingLevel = 2;
  @property({ attribute: false }) stack: ExpandedStack | null | undefined = null;
  @property({ attribute: false }) componentView: ComponentView | null = null;
  @property({ attribute: false }) action: TargetAction<number> | null = null;
  @property({ attribute: false })
  selection: SelectionDraftSelectionBinding<number> | SelectionDraftSelectionBinding<string> | null = null;
  @property({ type: Number, attribute: 'faux-components' }) fauxComponents = 0;
  @property({ type: Boolean, attribute: 'hide-count' }) hideCount = false;

  @query('boardgame-component-zone') private _zone!: HTMLElement & { stackElement?: BoardgameComponentStack | null };

  /** Direct access for motion/presence customization that the deck does not model. */
  get stackElement(): BoardgameComponentStack | null {
    return this._zone?.stackElement ?? null;
  }

  override render() {
    return html`
      <boardgame-component-zone
        exportparts="zone,heading,label,count,content,stack,empty"
        .label=${this.label}
        .emptyLabel=${this.emptyLabel}
        .headingLevel=${this.headingLevel}
        .stack=${this.stack}
        .componentView=${this.componentView}
        .action=${this.action}
        .selection=${this.selection}
        .fauxComponents=${this.fauxComponents}
        .hideCount=${this.hideCount}>
      </boardgame-component-zone>
    `;
  }
}

customElements.define('boardgame-deck', BoardgameDeck);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-deck': BoardgameDeck;
  }
}
