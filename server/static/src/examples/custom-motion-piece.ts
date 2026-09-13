/** A complete advanced authoring example; imports only the supported facade. */
import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import {
  BoardgameComponent,
  componentView,
  type ExpandedStack,
  type VisualMotionTrackInput,
} from '../client.js';

function rotation(heading: unknown): string {
  return `rotate(${typeof heading === 'number' && Number.isFinite(heading) ? heading : 0}deg)`;
}

export class CompassPiece extends BoardgameComponent {
  @property({ type: Number }) heading = 0;

  static override styles = [BoardgameComponent.styles, css`
    #outer { width: 56px; height: 56px; padding: 8px; }
    #inner {
      width: 100%; height: 100%; background: #6750a4;
      clip-path: polygon(50% 0, 95% 95%, 50% 72%, 5% 95%);
    }
  `];

  override get animatingProperties(): string[] { return ['heading']; }

  protected override propertyMotionTracks(
    before: Readonly<Record<string, unknown>>,
    after: Readonly<Record<string, unknown>>,
  ): readonly VisualMotionTrackInput[] {
    return [{ target: 'visual', property: 'transform',
      from: rotation(before['heading']), to: rotation(after['heading']) }];
  }

  override render() {
    return html`<div id="outer"><div id="inner"
      style=${`transform: ${rotation(this.heading)}`} aria-hidden="true"></div></div>`;
  }
}
customElements.define('example-compass-piece', CompassPiece);

declare global {
  interface HTMLElementTagNameMap { 'example-compass-piece': CompassPiece; }
}

export const compassView = componentView<ExpandedStack<{ Heading: number }>, CompassPiece>(
  () => document.createElement('example-compass-piece'),
  { properties: context => ({ heading: context.kind === 'visible' ? context.component.Values.Heading : 0 }) },
);
