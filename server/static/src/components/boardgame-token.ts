import { BoardgameComponent } from './boardgame-component.js';
import { html, css, TemplateResult } from 'lit';
import { property } from 'lit/decorators.js';

export class BoardgameToken extends BoardgameComponent {
  static override styles = [
    BoardgameComponent.styles,
    css`
      #inner {
        height: var(--component-effective-height);
        width: var(--component-effective-width);
      }

      #inner img {
        height: 100%;
        width: 100%;
      }

      #outer.pawn {
        --component-aspect-ratio: 2.0;
      }

      #outer.meeple {
        --component-aspect-ratio: 1.25;
      }

      #outer.active #inner,
      #outer.highlighted #inner {
        animation-name: throb;
        animation-duration: 1s;
        animation-timing-function: ease-in-out;
        animation-direction: alternate;
        animation-iteration-count: infinite;
      }

      #outer.active #inner {
        --throb-color-from: rgba(136, 136, 38, 1.0);
        --throb-color-to: rgba(136, 136, 38, 0.5);
      }

      #outer.highlighted #inner {
        --throb-color-from: rgba(0, 0, 0, 1.0);
        --throb-color-to: rgba(0, 0, 0, 0.5);
      }

      #outer.active.highlighted #inner {
        --throb-color-from: rgba(255, 255, 0, 1.0);
        --throb-color-to: rgba(255, 255, 0, 0.0);
      }

      @keyframes throb {
        from {
          filter: drop-shadow(0 0 0.25em var(--throb-color-to)) drop-shadow(0 0 0.25em var(--throb-color-to));
        }
        to {
          /* double the effect so it's darker */
          filter: drop-shadow(0 0 0.25em var(--throb-color-from)) drop-shadow(0 0 0.25em var(--throb-color-from));
        }
      }

      #outer.gray img {
        filter: saturate(0.0) brightness(3.0);
      }

      #outer.green img {
        filter: hue-rotate(130deg) brightness(2.0);
      }

      #outer.teal img {
        filter: hue-rotate(185deg) brightness(2.4);
      }

      #outer.purple img {
        filter: hue-rotate(300deg) brightness(1.0);
      }

      #outer.pink img {
        filter: hue-rotate(-93deg) brightness(4) saturate(0.8);
      }

      /* red is the default color, no need for shifting */

      #outer.blue img {
        filter: hue-rotate(220deg) brightness(2.0) saturate(1.5);
      }

      #outer.orange img {
        filter: hue-rotate(50deg) brightness(2.5);
      }

      #outer.yellow img {
        filter: hue-rotate(70deg) brightness(4);
      }

      #outer.black img {
        filter: saturate(0.0) brightness(1.7);
      }
    `
  ];

  @property({ type: String })
  color = 'red';

  @property({ type: Boolean })
  active = false;

  @property({ type: Boolean })
  highlighted = false;

  @property({ type: String })
  type = 'token';

  get legalTypes(): string[] {
    return ['token', 'chip', 'cube', 'pawn', 'meeple'];
  }

  get legalColors(): string[] {
    return ['gray', 'green', 'teal', 'purple', 'pink', 'red', 'blue', 'yellow', 'orange', 'black'];
  }

  override connectedCallback() {
    super.connectedCallback();
    this.altShadow = true;
  }

  protected override updated(changedProperties: Map<string, any>) {
    super.updated(changedProperties);

    if (
      changedProperties.has('color') ||
      changedProperties.has('active') ||
      changedProperties.has('highlighted') ||
      changedProperties.has('type')
    ) {
      this._updateClasses();
    }
  }

  private _computeAsset(type: string): string {
    return `src/assets/token_${type}.svg`;
  }

  // Override _computeClasses and add some more.
  protected override _computeClasses(): string {
    const result: string[] = [this.color];
    if (this.active) {
      result.push('active');
    }
    if (this.highlighted) {
      result.push('highlighted');
    }
    result.push(this.type);
    result.push(super._computeClasses());
    return result.join(' ');
  }

  override render(): TemplateResult {
    const asset = this._computeAsset(this.type);
    return html`
      <div id="outer" class="${this._computeClasses()}" @click="${this.handleTap}" style="${this._outerStyle}">
        <div id="inner">
          <img src="${asset}" />
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-token', BoardgameToken);
