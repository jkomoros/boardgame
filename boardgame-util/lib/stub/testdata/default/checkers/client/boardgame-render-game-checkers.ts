import { css, html } from '../../src/client.js';
import { GameRenderer, registerGameRenderer } from './_game_renderer.js';

@registerGameRenderer
export class BoardgameRenderGameCheckers extends GameRenderer {
  //Spreading GameRenderer.styles first is what brings in the shared
  //layout/state vocabulary (.horizontal, .vertical, .center, .flex, .active,
  //.responding, .selected, .targetable, .disabled, .eliminated) -- assigning
  //css`` directly to static styles would REPLACE it.
  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
    css`
      :host { display: block; }
    `,
  ];

  override render() {
    return html`
      <p>Build your game renderer here. State and move names are strictly typed.</p>
    `;
  }
}
