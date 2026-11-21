import { LitElement, css, html, nothing } from 'lit';
import { property } from 'lit/decorators.js';

/** A small labelled surface for game-owned stats, actions, and component zones. */
export class BoardgameMat extends LitElement {
  static override styles = css`
    :host { display: block; min-width: 0; container-type: inline-size; }
    #mat {
      display: grid; grid-template-columns: minmax(0, 1fr);
      gap: var(--boardgame-mat-gap, .75rem); min-width: 0;
      border: var(--boardgame-mat-border, 1px solid color-mix(in srgb, currentColor 25%, transparent));
      border-radius: var(--boardgame-mat-radius, .75rem);
      padding: var(--boardgame-mat-padding, .75rem);
      background: var(--boardgame-mat-background, color-mix(in srgb, currentColor 4%, transparent));
    }
    #mat.illustrated { grid-template-columns: minmax(0, var(--boardgame-mat-art-width, 9rem)) minmax(0, 1fr); }
    #art { width: 100%; aspect-ratio: var(--boardgame-mat-art-aspect-ratio, 4 / 3); object-fit: cover; border-radius: .5rem; }
    #content { min-width: 0; }
    #header { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: baseline; gap: .5rem; }
    #label { font: inherit; font-weight: 700; margin: 0; }
    ::slotted(*) { min-width: 0; max-width: 100%; box-sizing: border-box; }
    @container (max-width: 24rem) { #mat.illustrated { grid-template-columns: minmax(0, var(--boardgame-mat-compact-art-width, 5.5rem)) minmax(0, 1fr); } }
    @container (max-width: 15rem) { #mat.illustrated { grid-template-columns: minmax(0, 1fr); } #art { max-width: 9rem; } }
  `;

  @property({ type: String }) label = '';
  @property({ type: Number, attribute: 'heading-level' }) headingLevel = 3;
  @property({ type: String }) art = '';
  @property({ type: String, attribute: 'art-alt' }) artAlt = '';

  override render() {
    if (!this.label.trim()) throw new Error('boardgame-mat requires a label');
    if (!Number.isInteger(this.headingLevel) || this.headingLevel < 1 || this.headingLevel > 6) throw new Error('boardgame-mat heading-level must be 1 through 6');
    return html`<section id="mat" part="mat" class=${this.art ? 'illustrated' : ''} aria-labelledby="label">
      ${this.art ? html`<img id="art" part="art" src=${this.art} alt=${this.artAlt} loading="lazy" decoding="async">` : nothing}
      <div id="content" part="content">
        <div id="header" part="header"><span id="label" part="label" role="heading" aria-level=${this.headingLevel}>${this.label}</span><slot name="actions"></slot></div>
        <slot></slot>
      </div>
    </section>`;
  }
}
customElements.define('boardgame-mat', BoardgameMat);
declare global { interface HTMLElementTagNameMap { 'boardgame-mat': BoardgameMat; } }
